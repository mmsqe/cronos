import json
import time
from pathlib import Path

from pystarport import cluster as c

from .utils import wait_for_block


def test_join_validator(cronos):
    data = Path(cronos.base_dir).parent
    chain_id = cronos.config["chain_id"]
    cmd = "cronosd"
    cluster = c.ClusterCLI(data, cmd=cmd, chain_id=chain_id)
    i = cluster.create_node(moniker="new joined")
    addr = cluster.address("validator", i)
    denom = "basetcro"
    stake_denom = "stake"
    amt = 100000000000000000
    stake_amt = 1000000000000000000
    coins = f"{amt}{denom},{stake_amt}{stake_denom}"
    kwargs = {"gas": 300000, "gas_prices": "100000000000basetcro"}
    sender = cluster.address("validator")
    res = cluster.transfer(sender, addr, coins, **kwargs)
    assert res["code"] == 0
    assert cluster.balance(addr) == amt
    assert cluster.balance(addr, denom=stake_denom) == stake_amt

    # start the node
    cluster.supervisor.startProcess(f"{cluster.chain_id}-node{i}")
    # wait for the new node to sync
    wait_for_block(cluster.cosmos_cli(i), cluster.block_height())
    # wait for the new node to sync
    wait_for_block(cluster.cosmos_cli(i), cluster.block_height(0))

    count1 = len(cluster.validators())
    # create validator tx
    stake_amt = 4000000000000000
    stake_coins = f"{stake_amt}{stake_denom}"
    assert cluster.create_validator(stake_coins, i, **kwargs)["code"] == 0
    time.sleep(2)

    count2 = len(cluster.validators())
    assert count2 == count1 + 1, "new validator should joined successfully"

    val_addr = cluster.address("validator", i, bech="val")
    val = cluster.validator(val_addr)
    assert val["status"] == "BOND_STATUS_BONDED"
    assert val["tokens"] == str(stake_amt)
    assert val["description"]["moniker"] == "new joined"
    assert (
        cluster.edit_validator(i, commission_rate="0.2", **kwargs)["code"] == 12
    ), "commission cannot be changed more than once in 24h"
    consensus_pubkey = val["consensus_pubkey"]
    details = json.dumps(consensus_pubkey)
    res = cluster.edit_validator(i, moniker="awesome node", details=details, **kwargs)
    assert res["code"] == 0
    val = cluster.validator(val_addr)
    desc = val["description"]
    assert desc["moniker"] == "awesome node"
    assert desc["details"] == details
