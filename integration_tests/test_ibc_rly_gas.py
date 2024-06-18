import pytest
from pystarport import cluster

from .ibc_utils import (
    ibc_incentivized_transfer,
    ibc_multi_transfer,
    ibc_transfer,
    log_gas_records,
    prepare_network,
    rly_transfer,
)
from .utils import wait_for_new_blocks

pytestmark = pytest.mark.ibc_rly_gas


@pytest.fixture(scope="module", params=["ibc_rly_evm", "ibc_rly"])
def ibc(request, tmp_path_factory):
    "prepare-network"
    name = request.param
    path = tmp_path_factory.mktemp(name)
    yield from prepare_network(path, name, relayer=cluster.Relayer.RLY.value)


records = []


def test_ibc(ibc):
    # chainmain-1 relayer -> cronos_777-1 signer2
    cli = ibc.cronos.cosmos_cli()
    wait_for_new_blocks(cli, 1)
    ibc_transfer(ibc, transfer_fn=rly_transfer)
    # ibc_transfer(ibc)
    ibc_incentivized_transfer(ibc)
    ibc_multi_transfer(ibc)
    diff = 0.1
    record = log_gas_records(cli)
    if record:
        print("mm-record", record)
        records.append(record)
    if len(records) == 2:
        res = float(sum(records[0]) / sum(records[1]))
        print("mm-res", res)
        assert 1 - diff <= res <= 1 + diff, res
