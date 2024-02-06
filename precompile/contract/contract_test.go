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
		MustCalculateFunctionSelector("test(uint256,uint256)"),
		test,
	))

	functions = append(functions, NewStatefulPrecompileFunction(
		MustCalculateFunctionSelector("test(uint256,uint256)"),
		test,
	))

	// Construct the contract with no fallback function.
	_, err := NewStatefulPrecompileContract(functions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot create stateful precompile with duplicated function selector")
}

func TestPrecompileWithInvalidFunctionSelector(t *testing.T) {
	for _, tc := range []struct {
		desc       string
		fnSelector []byte
	}{
		{
			desc:       "empty selector",
			fnSelector: []byte{},
		},
		{
			desc:       "short selector",
			fnSelector: []byte("abc"),
		},
		{
			desc:       "long selector",
			fnSelector: []byte("acbde"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			functions := []*StatefulPrecompileFunction{
				NewStatefulPrecompileFunction(
					tc.fnSelector,
					test,
				),
			}

			// Construct the contract with no fallback function.
			_, err := NewStatefulPrecompileContract(functions)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid length of function selector")
		})
	}
}
