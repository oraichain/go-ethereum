package contract_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	callerAddr   = common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979")
	contractAddr = common.HexToAddress("0x999999cf1046e68e36E1aA2E0E07105eDDD1f08E")
)

type accessibleState struct {
	stateDB contract.StateDB
}

func newAccessibleState(stateDB contract.StateDB) *accessibleState {
	return &accessibleState{
		stateDB: stateDB,
	}
}

func (s *accessibleState) GetStateDB() contract.StateDB {
	return s.stateDB
}

func test(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
	value *big.Int,
) (ret []byte, remainingGas uint64, err error) {
	return nil, 0, nil
}

func TestPrecompileWithDuplicatedFunctionSelector(t *testing.T) {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("test(uint256,uint256)"),
		test,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("test(uint256,uint256)"),
		test,
	))

	// Construct the contract with no fallback function.
	_, err := contract.NewStatefulPrecompileContract(functions)
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
			functions := []*contract.StatefulPrecompileFunction{
				contract.NewStatefulPrecompileFunction(
					tc.fnSelector,
					test,
				),
			}

			// Construct the contract with no fallback function.
			_, err := contract.NewStatefulPrecompileContract(functions)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid length of function selector")
		})
	}
}

func TestPrecompileInvalidCalls(t *testing.T) {
	var (
		stateDB                = state.NewTestStateDB(t)
		accessibleState        = newAccessibleState(stateDB)
		suppliedGas     uint64 = 1
		readOnly        bool
	)

	functions := []*contract.StatefulPrecompileFunction{
		contract.NewStatefulPrecompileFunction(
			contract.MustCalculateFunctionSelector("test()"),
			test,
		),
	}
	// Construct the contract with no fallback function.
	precompiledContract, err := contract.NewStatefulPrecompileContract(functions)
	require.NoError(t, err)

	unexistingFuncInput := contract.MustCalculateFunctionSelector("unexistingFunc()")
	invalidArgNumInput := contract.MustCalculateFunctionSelector("getSum3(uint256)")
	shortFuncSelectorInput := []byte("abc")

	for _, tc := range []struct {
		desc  string
		input []byte
		err   string
	}{
		{
			desc:  "test case #1",
			input: unexistingFuncInput,
			err:   "invalid function selector",
		},
		{
			desc:  "test case #2",
			input: invalidArgNumInput,
			err:   "invalid function selector",
		},
		{
			desc:  "test case #3",
			input: shortFuncSelectorInput,
			err:   "missing function selector to precompile",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			{
				_, _, err := precompiledContract.Run(accessibleState, callerAddr, contractAddr, tc.input, suppliedGas, readOnly, common.Big0)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
			}
		})
	}
}

type stateDBWithExt struct {
	contract.StateDB

	customVal []byte
}

func newStateDBWithExt(t *testing.T, customVal []byte) *stateDBWithExt {
	return &stateDBWithExt{
		StateDB:   state.NewTestStateDB(t),
		customVal: customVal,
	}
}

func (s *stateDBWithExt) CustomValueExt() []byte {
	return s.customVal
}

func TestPrecompileCustomStateDBExtension(t *testing.T) {
	unsupportedErr := errors.New("unsupported statedb")

	// test function that asserts a concrete type on the stateDB,
	// and returns the return value of the extended type
	testFunc := func(
		accessibleState contract.AccessibleState,
		caller common.Address,
		addr common.Address,
		input []byte,
		suppliedGas uint64,
		readOnly bool,
		value *big.Int,
	) ([]byte, uint64, error) {
		customStateDB, ok := accessibleState.GetStateDB().(*stateDBWithExt)
		if !ok {
			return nil, suppliedGas, unsupportedErr
		}

		return customStateDB.CustomValueExt(), 0, nil
	}

	// create the precompiled contract with the testFunc
	funcSelector := contract.MustCalculateFunctionSelector("getStateDBCustomValue()")
	functions := []*contract.StatefulPrecompileFunction{
		contract.NewStatefulPrecompileFunction(funcSelector, testFunc),
	}
	precompiledContract, err := contract.NewStatefulPrecompileContract(functions)
	require.NoError(t, err)

	// build the accessible state with the extended statedb
	customVal := []byte("value from statedb extension")
	accessibleState := newAccessibleState(newStateDBWithExt(t, customVal))

	// run and see the returned value from calling CustomValueExt() on the injected statedb
	retVal, remainingGas, err := precompiledContract.Run(accessibleState, callerAddr, contractAddr, funcSelector, 1, true, common.Big0)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), remainingGas)
	assert.Equal(t, customVal, retVal, "expected contract to access and return customVal from extended statedb")

	// test contract with unsupported statedb
	accessibleState = newAccessibleState(state.NewTestStateDB(t))
	retVal, remainingGas, err = precompiledContract.Run(accessibleState, callerAddr, contractAddr, funcSelector, 1, true, common.Big0)
	require.Error(t, err)
	assert.Equal(t, unsupportedErr, err)
	assert.Equal(t, uint64(1), remainingGas)
	assert.Nil(t, retVal)
}
