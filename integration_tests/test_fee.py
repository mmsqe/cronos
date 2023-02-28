import requests
from pystarport import ports

from .utils import ADDRS, KEYS, sign_transaction, wait_for_new_blocks


def test_dynamic(cronos):
    w3 = cronos.w3
    cli = cronos.cosmos_cli()
    sender = ADDRS["community"]
    nonce = w3.eth.get_transaction_count(sender)
    blk = wait_for_new_blocks(cli, 1, sleep=0.1)
    print(f"block number start: {blk}")
    txhashes = []
    for n in range(3):
        tx = {
            "to": "0x2956c404227Cc544Ea6c3f4a36702D0FD73d20A2",
            "value": 25000000000000000000,
            "gas": 21000,
            "maxFeePerGas": 6556868066901,
            "maxPriorityFeePerGas": 1500000000,
            "nonce": nonce + n,
        }
        signed = sign_transaction(w3, tx, KEYS["community"])
        txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
        print("txhash", txhash.hex())
        txhashes.append(txhash)
    for txhash in txhashes[0:2]:
        res = w3.eth.wait_for_transaction_receipt(txhash)
        print(res)

    url = f"http://127.0.0.1:{ports.evmrpc_port(cronos.base_port(0))}"
    params = {
        "method": "debug_traceBlockByNumber",
        "params": [hex(blk + 1)],
        "id": 1,
        "jsonrpc": "2.0",
    }
    rsp = requests.post(url, json=params)
    assert rsp.status_code == 200
    print(rsp.json())
