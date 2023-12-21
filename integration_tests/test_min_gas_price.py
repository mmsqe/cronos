from pathlib import Path

import pytest

from .network import setup_custom_cronos
from .utils import ADDRS, KEYS, send_transaction, w3_wait_for_block


@pytest.fixture(scope="module")
def custom_cronos(tmp_path_factory):
    path = tmp_path_factory.mktemp("min-gas-price")
    yield from setup_custom_cronos(
        path, 26500, Path(__file__).parent / "configs/min_gas_price.jsonnet"
    )


def adjust_base_fee(parent_fee, gas_limit, gas_used, params):
    "spec: https://eips.ethereum.org/EIPS/eip-1559#specification"
    change_denominator = params["base_fee_change_denominator"]
    elasticity_multiplier = params["elasticity_multiplier"]
    gas_target = gas_limit // elasticity_multiplier

    delta = parent_fee * (gas_target - gas_used) // gas_target // change_denominator
    # https://github.com/crypto-org-chain/ethermint/blob/develop/x/feemarket/keeper/eip1559.go#L104
    return max(parent_fee - delta, params["min_gas_price"])


def get_params(cli):
    params = cli.query_params("feemarket")["params"]
    return {k: int(float(v)) for k, v in params.items()}


def test_dynamic_fee_tx(custom_cronos):
    w3 = custom_cronos.w3
    amount = 10000
    before = w3.eth.get_balance(ADDRS["community"])
    tip_price = 1
    max_price = 10000000000000 + tip_price
    tx = {
        "to": "0x0000000000000000000000000000000000000000",
        "value": amount,
        "gas": 21000,
        "maxFeePerGas": max_price,
        "maxPriorityFeePerGas": tip_price,
    }
    txreceipt = send_transaction(w3, tx, KEYS["community"])
    assert txreceipt.status == 1
    blk = w3.eth.get_block(txreceipt.blockNumber)
    assert txreceipt.effectiveGasPrice == blk.baseFeePerGas + tip_price

    fee_expected = txreceipt.gasUsed * txreceipt.effectiveGasPrice
    after = w3.eth.get_balance(ADDRS["community"])
    fee_deducted = before - after - amount
    assert fee_deducted == fee_expected

    assert blk.gasUsed == txreceipt.gasUsed  # we are the only tx in the block

    # check the next block's base fee is adjusted accordingly
    w3_wait_for_block(w3, txreceipt.blockNumber + 1)
    fee = w3.eth.get_block(txreceipt.blockNumber + 1).baseFeePerGas
    params = get_params(custom_cronos.cosmos_cli())
    assert fee == adjust_base_fee(
        blk.baseFeePerGas, blk.gasLimit, blk.gasUsed, params
    ), fee


def test_base_fee_adjustment(custom_cronos):
    """
    verify base fee adjustment of three continuous empty blocks
    """
    w3 = custom_cronos.w3
    begin = w3.eth.block_number
    w3_wait_for_block(w3, begin + 3)

    blk = w3.eth.get_block(begin)
    parent_fee = blk.baseFeePerGas
    params = get_params(custom_cronos.cosmos_cli())

    for i in range(3):
        fee = w3.eth.get_block(begin + 1 + i).baseFeePerGas
        assert fee == adjust_base_fee(parent_fee, blk.gasLimit, 0, params)
        parent_fee = fee
