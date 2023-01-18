import pytest
import web3

from .utils import (
    ADDRS,
    CONTRACTS,
    deploy_contract,
)


def test_estimate_gas(cronos):
    w3 = cronos.w3
    validator = ADDRS["validator"]
    erc20 = deploy_contract(
        w3,
        CONTRACTS["TestERC20A"],
    )
    with pytest.raises(web3.exceptions.ContractLogicError) as exc:
        w3.eth.estimate_gas({
            "from": validator,
            "to": erc20.address,
            "data": "0x3246485d" # revert method
        }, block_identifier="latest")
    assert "execution reverted" in str(exc)
