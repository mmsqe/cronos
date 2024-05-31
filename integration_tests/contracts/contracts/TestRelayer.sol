// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

contract TestRelayer {
    address constant relayer = 0x0000000000000000000000000000000000000065;

    function batchCall(bytes[] memory payloads) public {
        bytes memory payload = abi.encode(payloads);
        (bool success, ) = relayer.call(payload);
        require(success, "Relayer call failed");
    }
}
