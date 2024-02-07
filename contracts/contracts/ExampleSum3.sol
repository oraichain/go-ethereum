pragma solidity ^0.8.0;

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

    function calcSum3Call(uint256 a, uint256 b, uint256 c) public returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(ISum3.calcSum3.selector, a, b, c);

        (bool ok, bytes memory data) = address(PRECOMPILED_SUM3_CONTRACT_ADDRESS).call(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }

    function calcSum3DelegateCall(uint256 a, uint256 b, uint256 c) public returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(ISum3.calcSum3.selector, a, b, c);

        (bool ok, bytes memory data) = address(PRECOMPILED_SUM3_CONTRACT_ADDRESS).delegatecall(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }

    function getSum3StaticCall() public view returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(ISum3.getSum3.selector);

        (bool ok, bytes memory data) = address(PRECOMPILED_SUM3_CONTRACT_ADDRESS).staticcall(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }
}
