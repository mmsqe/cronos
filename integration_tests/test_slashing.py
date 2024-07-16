import datetime
from pathlib import Path

import pytest
from dateutil.parser import isoparse
from pystarport import ports

from .network import Cronos, setup_custom_cronos
from .utils import wait_for_block_time, wait_for_new_blocks, wait_for_port

pytestmark = pytest.mark.slow


@pytest.fixture(scope="module")
def cronos(request, tmp_path_factory):
    """start-cronos
    params: enable_auto_deployment
    """
    yield from setup_custom_cronos(
        tmp_path_factory.mktemp("slashing"),
        27100,
        Path(__file__).parent / "configs/slashing.jsonnet",
    )


def test_slashing(cronos: Cronos):
    "stop node2, wait for non-live slashing"
    cli = cronos.cosmos_cli()
    cli_2 = cronos.cosmos_cli(i=2)
    addr = cli_2.address("validator")
    val_addr = cli_2.address("validator", bech="val")
    tokens1 = int((cli.validator(val_addr))["tokens"])

    print("tokens before slashing", tokens1)
    print("stop and wait for 10 blocks")
    print(cronos.supervisorctl("stop", "cronos_777-1-node2"))
    wait_for_new_blocks(cli, 10)
    print(cronos.supervisorctl("start", "cronos_777-1-node2"))
    wait_for_port(ports.evmrpc_port(cronos.base_port(2)))

    val = cli.validator(val_addr)
    tokens2 = int(val["tokens"])
    print("tokens after slashing", tokens2)
    assert tokens2 == int(tokens1 * 0.99), "slash amount is not correct"

    assert val["jailed"], "validator is jailed"

    # try to unjail
    rsp = cli_2.unjail(addr)
    assert rsp["code"] == 4, "still jailed, can't be unjailed"

    # wait for 60s and unjail again
    wait_for_block_time(
        cli, isoparse(val["unbonding_time"]) + datetime.timedelta(seconds=60)
    )
    rsp = cli_2.unjail(addr)
    assert rsp["code"] == 0, f"unjail should success {rsp}"

    wait_for_new_blocks(cli, 3)
    assert len(cli.validators()) == 4
