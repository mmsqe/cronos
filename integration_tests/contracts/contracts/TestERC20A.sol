pragma solidity 0.8.10;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract TestERC20A is ERC20 {
	event __CronosSendToAccount(address recipient, uint256 amount);

	constructor() public ERC20("Bitcoin MAX", "MAX") {
		_mint(msg.sender, 100000000000000000000000000);
	}

	function test_native_transfer(uint amount) public {
		emit __CronosSendToAccount(msg.sender, amount);
	}

	// 0x9ffb86a5
    function revertWithMsg() public pure {
        revert("Function has been reverted");
    }

    // 0x3246485d
    function revertWithoutMsg() public pure {
        revert();
    }
}
