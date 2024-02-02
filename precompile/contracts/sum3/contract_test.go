package sum3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/precompile/contract"
)

var callerAddr = common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979")

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

func TestCalcSum3GasCalculations(t *testing.T) {
	var (
		stateDB         = state.NewTestStateDB(t)
		accessibleState = newAccessibleState(stateDB)
		contractAddr    = Module.Address
		readOnly        bool
	)

	for _, tc := range []struct {
		desc          string
		calcSum3Input CalcSum3Input
		suppliedGas   uint64
		remainingGas  uint64
		err           string
	}{
		{
			desc: "not enough gas",
			calcSum3Input: CalcSum3Input{
				A: big.NewInt(2),
				B: big.NewInt(3),
				C: big.NewInt(4),
			},
			suppliedGas:  calcSum3GasCost - 1,
			remainingGas: 0,
			err:          "out of gas",
		},
		{
			desc: "enough gas",
			calcSum3Input: CalcSum3Input{
				A: big.NewInt(2),
				B: big.NewInt(3),
				C: big.NewInt(4),
			},
			suppliedGas:  calcSum3GasCost,
			remainingGas: 0,
			err:          "",
		},
		{
			desc: "more than enough gas",
			calcSum3Input: CalcSum3Input{
				A: big.NewInt(2),
				B: big.NewInt(3),
				C: big.NewInt(4),
			},
			suppliedGas:  calcSum3GasCost + 1,
			remainingGas: 1,
			err:          "",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			input, err := PackCalcSum3(tc.calcSum3Input)
			require.NoError(t, err)

			ret, remainingGas, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, input, tc.suppliedGas, readOnly)
			if tc.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)
			require.Empty(t, ret)
			require.Equal(t, tc.remainingGas, remainingGas)
		})
	}
}

func TestGetSum3GasCalculations(t *testing.T) {
	var (
		stateDB         = state.NewTestStateDB(t)
		accessibleState = newAccessibleState(stateDB)
		contractAddr    = Module.Address
		readOnly        bool
	)

	for _, tc := range []struct {
		desc         string
		suppliedGas  uint64
		remainingGas uint64
		err          string
	}{
		{
			desc:         "not enough gas",
			suppliedGas:  getSum3GasCost - 1,
			remainingGas: 0,

			err: "out of gas",
		},
		{
			desc:         "enough gas",
			suppliedGas:  getSum3GasCost,
			remainingGas: 0,
			err:          "",
		},
		{
			desc:         "more than enough gas",
			suppliedGas:  getSum3GasCost + 1,
			remainingGas: 1,
			err:          "",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			input, err := PackGetSum3()
			require.NoError(t, err)

			ret, remainingGas, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, input, tc.suppliedGas, readOnly)
			if tc.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)

			sum, err := UnpackGetSum3Output(ret)
			require.NoError(t, err)
			require.Equal(t, int64(0), sum.Int64())
			require.Equal(t, tc.remainingGas, remainingGas)
		})
	}
}

func TestSum3Precompile(t *testing.T) {
	var (
		stateDB         = state.NewTestStateDB(t)
		accessibleState = newAccessibleState(stateDB)
		contractAddr    = Module.Address
		readOnly        bool
	)

	for _, tc := range []struct {
		desc          string
		calcSum3Input CalcSum3Input
		sum           int64
	}{
		{
			desc: "test case #1",
			calcSum3Input: CalcSum3Input{
				A: big.NewInt(2),
				B: big.NewInt(3),
				C: big.NewInt(4),
			},
			sum: 9,
		},
		{
			desc: "test case #2",
			calcSum3Input: CalcSum3Input{
				A: big.NewInt(3),
				B: big.NewInt(5),
				C: big.NewInt(7),
			},
			sum: 15,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// CalcSum3
			{
				input, err := PackCalcSum3(tc.calcSum3Input)
				require.NoError(t, err)

				ret, remainingGas, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, input, calcSum3GasCost, readOnly)
				require.NoError(t, err)
				require.Empty(t, ret)
				require.Equal(t, uint64(0), remainingGas)
			}

			// GetSum3
			{
				input, err := PackGetSum3()
				require.NoError(t, err)

				ret, remainingGas, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, input, getSum3GasCost, readOnly)
				require.NoError(t, err)

				sum, err := UnpackGetSum3Output(ret)
				require.NoError(t, err)
				require.Equal(t, tc.sum, sum.Int64())
				require.Equal(t, uint64(0), remainingGas)
			}
		})
	}
}

func TestSum3PrecompileInvalidCalls(t *testing.T) {
	var (
		stateDB         = state.NewTestStateDB(t)
		accessibleState = newAccessibleState(stateDB)
		contractAddr    = Module.Address
		readOnly        bool
	)

	unexistingFuncInput := contract.CalculateFunctionSelector("unexistingFunc()")
	invalidArgNumInput := contract.CalculateFunctionSelector("getSum3(uint256)")
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
				_, _, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, tc.input, calcSum3GasCost, readOnly)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
			}
		})
	}
}
