// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

contract ICACallback {
    uint64 lastAckSeq;
    bytes lastAck;
    mapping (uint64 => bytes) public acknowledgement;
    event OnPacketResult(uint64 seq, bytes ack);

    function getLastAckSeq() public view returns (uint256) {
        return lastAckSeq;
    }

    function setLastAckSeq(uint64 seq, bytes calldata ack) public returns (uint64) {
        lastAckSeq = seq;
        lastAck = ack;
        acknowledgement[seq] = ack;
        emit OnPacketResult(seq, ack);
        return lastAckSeq;
    }

    function onPacketResult(uint64 seq, address packetSenderAddress, bytes calldata ack) external payable returns (bool) {
        require(packetSenderAddress == address(this), "different sender");
        acknowledgement[seq] = ack;
        lastAck = ack;
        emit OnPacketResult(seq, ack);
        return true;
    }
}
