import { expect } from "chai";
import { ethers } from "hardhat";
import {BaseContract, Contract} from "ethers";
import {HardhatEthersSigner} from "@nomicfoundation/hardhat-ethers/src/signers";

const PRECOMPILED_SUM3_CONTRACT_ADDRESS: string = "0x0300000000000000000000000000000000000000";

const sum3ContractABI: string[] = [
    "function calcSum3(uint256 a, uint256 b, uint256 c) external",
    "function getSum3() external view returns (uint256)",
];

describe("Testing precompiled sum3 contract", function() {
    it("Should properly calculate sum of 3 numbers (direct call to precompiled contract)", async function () {
        const sum3Contract: Contract = new ethers.Contract(PRECOMPILED_SUM3_CONTRACT_ADDRESS, sum3ContractABI, ethers.provider);
        const signers: HardhatEthersSigner[] = await ethers.getSigners();
        const sum3ContractWithSigner: BaseContract = sum3Contract.connect(signers[0]);

        await sum3ContractWithSigner.calcSum3(2, 3, 4);
        let actual: string = await sum3Contract.getSum3();
        let expected: string = "0x0000000000000000000000000000000000000000000000000000000000000009";
        expect(actual).to.equal(expected);

        await sum3ContractWithSigner.calcSum3(3, 5, 7);
        actual = await sum3Contract.getSum3();
        expected = "0x000000000000000000000000000000000000000000000000000000000000000f"
        expect(actual).to.equal(expected);
    })

    it("Should properly calculate sum of 3 numbers (call to example contract)", async function () {
        const ExampleSum3 = await ethers.getContractFactory("ExampleSum3");
        const exampleSum3 = await ExampleSum3.deploy();

        await exampleSum3.calcSum3(2, 3, 4);
        let actual: string = await exampleSum3.getSum3();
        let expected: string = "0x0000000000000000000000000000000000000000000000000000000000000009";
        expect(actual).to.equal(expected);

        await exampleSum3.calcSum3(3, 5, 7);
        actual = await exampleSum3.getSum3();
        expected = "0x000000000000000000000000000000000000000000000000000000000000000f";
        expect(actual).to.equal(expected);
    })

    it("Should properly calculate sum of 3 numbers (Call & StaticCall OpCodes)", async function () {
        const ExampleSum3 = await ethers.getContractFactory("ExampleSum3");
        const exampleSum3 = await ExampleSum3.deploy();

        await exampleSum3.calcSum3Call(2, 3, 4);
        let actual: string = await exampleSum3.getSum3StaticCall();
        let expected: string = "0x0000000000000000000000000000000000000000000000000000000000000009";
        expect(actual).to.equal(expected);

        await exampleSum3.calcSum3Call(3, 5, 7);
        actual = await exampleSum3.getSum3StaticCall();
        expected = "0x000000000000000000000000000000000000000000000000000000000000000f";
        expect(actual).to.equal(expected);
    })

    it("Should properly calculate sum of 3 numbers (DelegateCall & StaticCall OpCodes)", async function () {
        const ExampleSum3 = await ethers.getContractFactory("ExampleSum3");
        const exampleSum3 = await ExampleSum3.deploy();

        await exampleSum3.calcSum3DelegateCall(2, 3, 4);
        let actual: string = await exampleSum3.getSum3StaticCall();
        let expected: string = "0x0000000000000000000000000000000000000000000000000000000000000009";
        expect(actual).to.equal(expected);

        await exampleSum3.calcSum3DelegateCall(3, 5, 7);
        actual = await exampleSum3.getSum3StaticCall();
        expected = "0x000000000000000000000000000000000000000000000000000000000000000f";
        expect(actual).to.equal(expected);
    })

    it("Should fail because you can't call state-changing method with staticcall opcode", async function () {
        const ErrorExampleSum3 = await ethers.getContractFactory("ErrorExampleSum3");
        const errorExampleSum3 = await ErrorExampleSum3.deploy();

        await expect(errorExampleSum3.calcSum3StaticCall(2, 3, 4)).to.be.revertedWith("call to precompiled contract failed");
    })
})
