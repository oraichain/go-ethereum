pragma solidity ^0.8.0;

import "hardhat/console.sol";

address constant PRECOMPILED_SUM3_CONTRACT_ADDRESS = address(0x0300000000000000000000000000000000000000);

interface ISum3 {
    function calcSum3(uint256 a, uint256 b, uint256 c) external;
    function getSum3() external view returns (uint256 result);
}

contract ExampleSum3 {
    function calcSum3(uint256 a, uint256 b, uint256 c) public {
        ISum3(PRECOMPILED_SUM3_CONTRACT_ADDRESS).calcSum3(a,b,c);
    }

    function getSum3() public view returns (uint256) {
        uint256 result = ISum3(PRECOMPILED_SUM3_CONTRACT_ADDRESS).getSum3();
        return result;
    }
}
