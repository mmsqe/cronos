// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

contract TestEvent {
    uint64 lastSeq;
    enum Status {
        NONE,
        SUCCESS,
        TIMEOUT
    }
    mapping (uint64 => Status) public statusMap;
    event OnPacketResult(uint64 seq, Status status);

    function getLastSeq() public view returns (uint256) {
        return lastSeq;
    }

    function onPacketResultCallback(uint64 seq, bool ack) external payable returns (bool) {
        // To prevent called by arbitrary user
        Status status = Status.TIMEOUT;
        if (ack) {
            status = Status.SUCCESS;
        }
        statusMap[seq] = status;
        emit OnPacketResult(seq, status);
        return true;
    }
}
