import json
import shutil
import stat
import subprocess
from contextlib import contextmanager
from pathlib import Path

import pytest
import requests
from pystarport import cluster, ports
from pystarport.cluster import SUPERVISOR_CONFIG_FILE

from .network import Cronos, setup_custom_cronos
from .utils import (
    ADDRS,
    approve_proposal,
    assert_gov_params,
    edit_ini_sections,
    send_transaction,
    wait_for_block,
    wait_for_new_blocks,
    wait_for_port,
)

pytestmark = pytest.mark.upgrade
DEFAULT_GAS_PRICE = "5000000000000basetcro"


@pytest.fixture(scope="module")
def custom_cronos(tmp_path_factory):
    yield from setup_cronos_test(tmp_path_factory)


def get_txs(base_port, end):
    port = ports.rpc_port(base_port)
    res = []
    for h in range(1, end):
        url = f"http://127.0.0.1:{port}/block_results?height={h}"
        res.append(requests.get(url).json().get("result")["txs_results"])
    return res


def init_cosmovisor(home):
    """
    build and setup cosmovisor directory structure in each node's home directory
    """
    cosmovisor = home / "cosmovisor"
    cosmovisor.mkdir()
    (cosmovisor / "upgrades").symlink_to("../../../upgrades")
    (cosmovisor / "genesis").symlink_to("./upgrades/genesis")


def post_init(path, base_port, config):
    """
    prepare cosmovisor for each node
    """
    chain_id = "cronos_777-1"
    data = path / chain_id
    cfg = json.loads((data / "config.json").read_text())
    for i, _ in enumerate(cfg["validators"]):
        home = data / f"node{i}"
        init_cosmovisor(home)

    edit_ini_sections(
        chain_id,
        data / SUPERVISOR_CONFIG_FILE,
        lambda i, _: {
            "command": f"cosmovisor run start --home %(here)s/node{i}",
            "environment": (
                "DAEMON_NAME=cronosd,"
                "DAEMON_SHUTDOWN_GRACE=1m,"
                "UNSAFE_SKIP_BACKUP=true,"
                f"DAEMON_HOME=%(here)s/node{i}"
            ),
        },
    )


def setup_cronos_test(tmp_path_factory):
    path = tmp_path_factory.mktemp("upgrade")
    port = 26200
    nix_name = "upgrade-test-package"
    cfg_name = "cosmovisor"
    configdir = Path(__file__).parent
    cmd = [
        "nix-build",
        configdir / f"configs/{nix_name}.nix",
    ]
    print(*cmd)
    subprocess.run(cmd, check=True)

    # copy the content so the new directory is writable.
    upgrades = path / "upgrades"
    shutil.copytree("./result", upgrades)
    mod = stat.S_IRWXU
    upgrades.chmod(mod)
    for d in upgrades.iterdir():
        d.chmod(mod)

    # init with genesis binary
    with contextmanager(setup_custom_cronos)(
        path,
        port,
        configdir / f"configs/{cfg_name}.jsonnet",
        post_init=post_init,
        chain_binary=str(upgrades / "genesis/bin/cronosd"),
    ) as cronos:
        yield cronos


def assert_evm_params(cli, expected, height):
    params = cli.query_params("evm", height=height)
    del params["header_hash_num"]
    assert expected == params


def check_basic_tx(c):
    # check basic tx works
    wait_for_port(ports.evmrpc_port(c.base_port(0)))
    receipt = send_transaction(
        c.w3,
        {
            "to": ADDRS["community"],
            "value": 1000,
            "maxFeePerGas": 10000000000000,
            "maxPriorityFeePerGas": 10000,
        },
    )
    assert receipt.status == 1


def exec(c, tmp_path_factory):
    """
    - propose an upgrade and pass it
    - wait for it to happen
    - it should work transparently
    """
    cli = c.cosmos_cli()
    base_port = c.base_port(0)

    wait_for_port(ports.evmrpc_port(base_port))
    wait_for_new_blocks(cli, 1)
    receipt = send_transaction(
        c.w3,
        {
            "to": ADDRS["community"],
            "value": 1000,
            "maxFeePerGas": 10000000000000,
            "maxPriorityFeePerGas": 10000,
        },
    )
    assert receipt.status == 1

    def do_upgrade(
        plan_name,
        target,
        mode=None,
        method="submit-legacy-proposal",
        gas_prices=DEFAULT_GAS_PRICE,
    ):
        rsp = cli.gov_propose_legacy(
            "community",
            "software-upgrade",
            {
                "name": plan_name,
                "title": "upgrade test",
                "description": "ditto",
                "upgrade-height": target,
                "deposit": "10000basetcro",
            },
            mode=mode,
            method=method,
            gas_prices=gas_prices,
        )
        assert rsp["code"] == 0, rsp["raw_log"]
        approve_proposal(c, rsp["logs"][0]["events"], event_query_tx=True)

        # update cli chain binary
        c.chain_binary = (
            Path(c.chain_binary).parent.parent.parent / f"{plan_name}/bin/cronosd"
        )
        # block should pass the target height
        wait_for_block(c.cosmos_cli(), target + 2, timeout=480)
        wait_for_port(ports.rpc_port(base_port))
        return c.cosmos_cli()

    height = cli.block_height()
    target_height3 = height + 15
    print("upgrade v1.4 height", target_height3)
    gov_param = cli.query_params("gov")

    cli = do_upgrade("v1.4", target_height3)

    wait_for_block(cli, 41, timeout=30)

    data = Path(c.base_dir).parent  # Same data dir as cronos fixture
    chain_id = c.config["chain_id"]  # Same chain_id as cronos fixture
    cmd = "cronosd"
    # create a clustercli object from ClusterCLI class
    clustercli = cluster.ClusterCLI(data, cmd=cmd, chain_id=chain_id)
    for i in [0]:
        clustercli.supervisor.stopProcess(f"{clustercli.chain_id}-node{i}")
        cluster.edit_app_cfg(
            clustercli.home(i) / "config/app.toml",
            c.base_port(i),
            {
                "pruning": "everything",
                "state-sync": {
                    "snapshot-interval": 0,
                },
            },
        )
        clustercli.supervisor.startProcess(f"{clustercli.chain_id}-node{i}")
        wait_for_port(ports.evmrpc_port(c.base_port(i)))

    wait_for_block(cli, 55, timeout=30)

    def check_log_for_string(log_file_path, search_string):
        with open(log_file_path, "r") as log_file:
            for line in log_file:
                if search_string in line:
                    print(f"Found: {line.strip()}")
                    return True
        return False

    log_file_path = f"{clustercli.home(0)}.log"
    search_string = "Value missing for key"
    assert not check_log_for_string(log_file_path, search_string), "found"

    with pytest.raises(AssertionError):
        cli.query_params("icaauth")
    assert_gov_params(cli, gov_param)
    receipt = send_transaction(
        c.w3,
        {
            "to": ADDRS["community"],
            "value": 1000,
            "maxFeePerGas": 10000000000000,
            "maxPriorityFeePerGas": 10000,
        },
    )
    assert receipt.status == 1


def test_cosmovisor_upgrade(custom_cronos: Cronos, tmp_path_factory):
    exec(custom_cronos, tmp_path_factory)
