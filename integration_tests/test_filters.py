from web3 import Web3

from .utils import (
    ADDRS,
    CONTRACTS,
    deploy_contract,
    send_transaction,
    wait_for_new_blocks,
)


def test_pending_transaction_filter(cluster):
    w3: Web3 = cluster.w3
    flt = w3.eth.filter("pending")
    assert flt.get_new_entries() == []
    receipt = send_transaction(w3, {"to": ADDRS["community"], "value": 1000})
    assert receipt.status == 1
    assert receipt["transactionHash"] in flt.get_new_entries()
