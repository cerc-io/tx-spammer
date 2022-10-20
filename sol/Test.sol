// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.0;

contract Test {
  address payable owner;

  modifier onlyOwner {
    require(
      msg.sender == owner,
      "Only owner can call this function."
    );
    _;
  }

  mapping(address => uint256) public data;

  constructor() {
    owner = payable(msg.sender);
  }

  function Put(address addr, uint256 value) public {
    data[addr] = value;
  }

  function close() public onlyOwner {
    selfdestruct(owner);
  }
}
