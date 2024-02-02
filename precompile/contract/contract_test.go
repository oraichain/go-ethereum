package contract

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func test(
	accessibleState AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	return nil, 0, nil
}

func TestPrecompileWithDuplicatedFunctionSelector(t *testing.T) {
	var functions []*StatefulPrecompileFunction

	functions = append(functions, NewStatefulPrecompileFunction(
		CalculateFunctionSelector("test(uint256,uint256)"),
		test,
	))

	functions = append(functions, NewStatefulPrecompileFunction(
		CalculateFunctionSelector("test(uint256,uint256)"),
		test,
	))

	// Construct the contract with no fallback function.
	_, err := NewStatefulPrecompileContract(functions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot create stateful precompile with duplicated function selector")
}
