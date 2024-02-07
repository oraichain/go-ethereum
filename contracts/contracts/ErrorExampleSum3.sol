pragma solidity ^0.8.0;

address constant PRECOMPILED_SUM3_CONTRACT_ADDRESS = address(0x0300000000000000000000000000000000000000);

// ErrorISum3 interface will be used in test case scenario, where we try to call state-changing method
// with staticcall opcode and make sure it will fail.
// That's why we declare calcSum3 method with view modifier, despite it's state-changing method.
interface ErrorISum3 {
    function calcSum3(uint256 a, uint256 b, uint256 c) external view;
}

// ErrorExampleSum3 contract will be used in test case scenario, where we try to call state-changing method
// with staticcall opcode and make sure it will fail.
// That's why we declare calcSum3StaticCall method with view modifier, despite it's state-changing method.
contract ErrorExampleSum3 {
    function calcSum3StaticCall(uint256 a, uint256 b, uint256 c) public view returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(ErrorISum3.calcSum3.selector, a, b, c);

        (bool ok, bytes memory data) = address(PRECOMPILED_SUM3_CONTRACT_ADDRESS).staticcall(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }
}
