// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import {IICAModule} from "./ICA.sol";
import {CosmosBankV1beta1MsgSend} from "./generated/cosmos/bank/v1beta1/tx.sol";
import {CosmosBaseV1beta1Coin} from "./generated/cosmos/base/v1beta1/coin.sol";
import {
    IbcApplicationsInterchain_accountsV1InterchainAccountPacketData,
    PACKET_PROTO_GLOBAL_ENUMS
} from "./generated/ibc/applications/interchain_accounts/v1/packet.sol";

contract TestICA {
    address constant icaContract = 0x0000000000000000000000000000000000000066;
    IICAModule ica = IICAModule(icaContract);

    function encodeRegister(string memory connectionID) internal view returns (bytes memory) {
        return abi.encodeWithSignature(
            "registerAccount(string,string)",
            connectionID, msg.sender, ""
        );
    }

    function callRegister(string memory connectionID) public returns (bool) {
        return ica.registerAccount(connectionID, "");
    }

    function delegateRegister(string memory connectionID) public returns (bool) {
        (bool result,) = icaContract.delegatecall(encodeRegister(connectionID));
        require(result, "call failed");
        return true;
    }

    function staticRegister(string memory connectionID) public returns (bool) {
        (bool result,) = icaContract.staticcall(encodeRegister(connectionID));
        require(result, "call failed");
        return true;
    }

    function encodeQueryAccount(string memory connectionID, address addr) internal view returns (bytes memory) {
        return abi.encodeWithSignature(
            "queryAccount(string,address)",
            connectionID, addr
        );
    }

    function callQueryAccount(string memory connectionID, address addr) public returns (string memory) {
        return ica.queryAccount(connectionID, addr);
    }

    function delegateQueryAccount(string memory connectionID, address addr) public returns (string memory) {
        (bool result, bytes memory data) = icaContract.delegatecall(encodeQueryAccount(connectionID, addr));
        require(result, "call failed");
        return abi.decode(data, (string));
    }

    function staticQueryAccount(string memory connectionID, address addr) public returns (string memory) {
        (bool result, bytes memory data) = icaContract.staticcall(encodeQueryAccount(connectionID, addr));
        require(result, "call failed");
        return abi.decode(data, (string));
    }

    function encodeSubmitMsgs(string memory connectionID, string memory data) internal view returns (bytes memory) {
        return abi.encodeWithSignature(
            "submitMsgs(string,string,uint256)",
            connectionID, msg.sender, data, 300000000000
        );
    }

    function callSubmitMsgs(string memory connectionID, string memory data) public returns (uint64) {
        return ica.submitMsgs(connectionID, data, 300000000000);
    }

    function delegateSubmitMsgs(string memory connectionID, string memory data) public returns (uint64) {
        (bool result, bytes memory data) = icaContract.delegatecall(encodeSubmitMsgs(connectionID, data));
        require(result, "call failed");
        return abi.decode(data, (uint64));
    }

    function nativeSubmitsMsgSend(
        string memory connectionID,
        string memory sender,
        string memory receiver,
        string memory denom,
        string memory amt
    ) public returns (uint64) {
        CosmosBaseV1beta1Coin.Data[] memory coins = new CosmosBaseV1beta1Coin.Data[](1);
        coins[0].denom = denom;
        coins[0].amount = amt;
        bytes memory data = abi.encodeWithSignature(
            "submitMsgs(string,string,uint256)",
            connectionID,
            msg.sender,
            IbcApplicationsInterchain_accountsV1InterchainAccountPacketData.encode(
                IbcApplicationsInterchain_accountsV1InterchainAccountPacketData.Data({
                    f_type: PACKET_PROTO_GLOBAL_ENUMS.Type.TYPE_EXECUTE_TX,
                    data: CosmosBankV1beta1MsgSend.encode(
                        CosmosBankV1beta1MsgSend.Data({
                            from_address: sender,
                            to_address: receiver,
                            amount: coins
                        })
                    ),
                    memo: ""
                })
            ),
            300000000000
        );
        return ica.submitMsgs(connectionID, bytesToString(data), 300000000000);
    }

    function bytesToString(bytes memory byteCode) public pure returns(string memory stringData) {
        uint256 blank = 0; //blank 32 byte value
        uint256 length = byteCode.length;

        uint cycles = byteCode.length / 0x20;
        uint requiredAlloc = length;

        if (length % 0x20 > 0) //optimise copying the final part of the bytes - to avoid looping with single byte writes
        {
            cycles++;
            requiredAlloc += 0x20; //expand memory to allow end blank, so we don't smack the next stack entry
        }

        stringData = new string(requiredAlloc);

        //copy data in 32 byte blocks
        assembly {
            let cycle := 0

            for
            {
                let mc := add(stringData, 0x20) //pointer into bytes we're writing to
                let cc := add(byteCode, 0x20)   //pointer to where we're reading from
            } lt(cycle, cycles) {
                mc := add(mc, 0x20)
                cc := add(cc, 0x20)
                cycle := add(cycle, 0x01)
            } {
                mstore(mc, mload(cc))
            }
        }

        //finally blank final bytes and shrink size (part of the optimisation to avoid looping adding blank bytes1)
        if (length % 0x20 > 0)
        {
            uint offsetStart = 0x20 + length;
            assembly
            {
                let mc := add(stringData, offsetStart)
                mstore(mc, mload(add(blank, 0x20)))
                //now shrink the memory back so the returned object is the correct size
                mstore(stringData, length)
            }
        }
    }
}