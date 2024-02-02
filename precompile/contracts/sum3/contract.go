package sum3

import (
	_ "embed"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/precompile/contract"
)

const (
	calcSum3GasCost uint64 = contract.WriteGasCostPerSlot
	getSum3GasCost  uint64 = contract.ReadGasCostPerSlot
)

// Singleton StatefulPrecompiledContract.
var (
	// Sum3RawABI contains the raw ABI of Sum3 contract.
	//go:embed ISum3.abi
	Sum3RawABI string

	Sum3ABI = contract.ParseABI(Sum3RawABI)

	Sum3Precompile = createSum3Precompile()
)

var (
	sumKey = common.BytesToHash([]byte("sumKey"))
)

func StoreSum(stateDB vm.StateDB, sum *big.Int) {
	valuePadded := common.LeftPadBytes(sum.Bytes(), common.HashLength)
	valueHash := common.BytesToHash(valuePadded)

	stateDB.SetState(ContractAddress, sumKey, valueHash)
}

func GetSum(stateDB vm.StateDB) (*big.Int, error) {
	value := stateDB.GetState(ContractAddress, sumKey)
	if len(value.Bytes()) == 0 {
		return big.NewInt(0), nil
	}

	var sum big.Int
	sum.SetBytes(value.Bytes())

	return &sum, nil
}

func calcSum3(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, calcSum3GasCost); err != nil {
		return nil, 0, err
	}

	// Set the nonce of the precompile's address (as is done when a contract is created) to ensure
	// that it is marked as non-empty and will not be cleaned up when the statedb is finalized.
	accessibleState.GetStateDB().SetNonce(ContractAddress, 1)
	// Set the code of the precompile's address to a non-zero length byte slice to ensure that the precompile
	// can be called from within Solidity contracts. Solidity adds a check before invoking a contract to ensure
	// that it does not attempt to invoke a non-existent contract.
	accessibleState.GetStateDB().SetCode(ContractAddress, []byte{0x1})

	if len(input) != 96 {
		return nil, remainingGas, fmt.Errorf("unexpected input length, want: 96, got: %v", len(input))
	}

	var a, b, c, rez big.Int
	a.SetBytes(input[:32])
	b.SetBytes(input[32:64])
	c.SetBytes(input[64:96])
	rez.Add(&a, &b)
	rez.Add(&rez, &c)

	StoreSum(accessibleState.GetStateDB(), &rez)

	packedOutput := make([]byte, 0)
	return packedOutput, remainingGas, nil
}

func getSum3(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, getSum3GasCost); err != nil {
		return nil, 0, err
	}

	sum, err := GetSum(accessibleState.GetStateDB())
	if err != nil {
		return nil, remainingGas, err
	}

	packedOutput := common.LeftPadBytes(sum.Bytes(), 32)
	return packedOutput, remainingGas, nil
}

// createSum3Precompile returns a StatefulPrecompiledContract with getters and setters for the precompile.
func createSum3Precompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("calcSum3(uint256,uint256,uint256)"),
		calcSum3,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("getSum3()"),
		getSum3,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}

type CalcSum3Input struct {
	A *big.Int
	B *big.Int
	C *big.Int
}

// PackCalcSum3 packs [inputStruct] of type CalcSum3Input into the appropriate arguments for calcSum3.
func PackCalcSum3(inputStruct CalcSum3Input) ([]byte, error) {
	return Sum3ABI.Pack("calcSum3", inputStruct.A, inputStruct.B, inputStruct.C)
}

// PackGetSum3 packs the include selector (first 4 func signature bytes).
func PackGetSum3() ([]byte, error) {
	return Sum3ABI.Pack("getSum3")
}

// UnpackGetSum3Output attempts to unpack given [output] into the *big.Int type output
// assumes that [output] does not include selector (omits first 4 func signature bytes)
func UnpackGetSum3Output(output []byte) (*big.Int, error) {
	res, err := Sum3ABI.Unpack("getSum3", output)
	if err != nil {
		return new(big.Int), err
	}
	unpacked := *abi.ConvertType(res[0], new(*big.Int)).(**big.Int)
	return unpacked, nil
}
