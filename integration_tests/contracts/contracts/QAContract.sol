// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract QAContract {
    function revertWithMsg() public pure {
        revert("Function has been reverted");
    }

    function revertWithoutMsg() public pure {
        revert();
    }
}