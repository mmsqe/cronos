import pytest
import web3

from .utils import ADDRS, CONTRACTS, deploy_contract


def test_estimate_gas(cronos):
    w3 = cronos.w3
    validator = ADDRS["validator"]
    erc20 = deploy_contract(
        w3,
        CONTRACTS["TestERC20A"],
    )
    # revert methods
    for data in ["0x9ffb86a5", "0x3246485d"]:
        with pytest.raises(web3.exceptions.ContractLogicError) as exc:
            params = {"from": validator, "to": erc20.address, "data": data}
            w3.eth.estimate_gas(params, block_identifier="latest")
        assert "execution reverted" in str(exc)
