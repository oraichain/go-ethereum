// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/precompile/contract"
)

func TestFunctionSignatureRegex(t *testing.T) {
	type test struct {
		str  string
		pass bool
	}

	for _, test := range []test{
		{
			str:  "getBalance()",
			pass: true,
		},
		{
			str:  "getBalance(address)",
			pass: true,
		},
		{
			str:  "getBalance(address,address)",
			pass: true,
		},
		{
			str:  "getBalance(address,address,address)",
			pass: true,
		},
		{
			str:  "getBalance(address,address,address,uint256)",
			pass: true,
		},
		{
			str:  "getBalance(address,)",
			pass: false,
		},
		{
			str:  "getBalance(address,address,)",
			pass: false,
		},
		{
			str:  "getBalance(,)",
			pass: false,
		},
		{
			str:  "(address,)",
			pass: false,
		},
		{
			str:  "()",
			pass: false,
		},
		{
			str:  "dummy",
			pass: false,
		},
	} {
		require.Equal(t, test.pass, contract.FunctionSignatureRegex.MatchString(test.str), "unexpected result for %q", test.str)
	}
}

func TestDeductGas(t *testing.T) {
	for _, tc := range []struct {
		desc         string
		suppliedGas  uint64
		requiredGas  uint64
		remainingGas uint64
		err          string
	}{
		{
			desc:        "not enough gas",
			suppliedGas: 0,
			requiredGas: 100,
			err:         "out of gas",
		},
		{
			desc:         "enough gas",
			suppliedGas:  100,
			requiredGas:  100,
			remainingGas: 0,
		},
		{
			desc:         "more than enough gas",
			suppliedGas:  200,
			requiredGas:  100,
			remainingGas: 100,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			remainingGas, err := contract.DeductGas(tc.suppliedGas, tc.requiredGas)
			if tc.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), "out of gas")
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.remainingGas, remainingGas)
			}
		})
	}
}
