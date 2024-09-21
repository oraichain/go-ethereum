// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/ethereum/go-ethereum/precompile/modules"
	"github.com/stretchr/testify/require"
)

func TestDelegateCallPreCompile(t *testing.T) {
	delegateCallChainConfig := params.TestChainConfig
	delegateCallChainConfig.EIP150Block = big.NewInt(-1)

	// init contract addresses
	eoaAddr := AccountRef(common.HexToAddress("0xB0ac9d216b303a32907632731a93356228CAEE87"))
	contractAddressA := AccountRef(common.HexToAddress("0x58028b5F974955dfe9E3d7CefdE31f0b1c1227d0"))
	// precompile contract
	precompileContractAddress := AccountRef(common.HexToAddress("0x9000000000000000000000000000000000000001"))
	initialBalance := big.NewInt(10)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	statedb.CreateAccount(eoaAddr.Address())
	statedb.CreateAccount(contractAddressA.Address())
	statedb.CreateAccount(precompileContractAddress.Address())
	statedb.SetBalance(contractAddressA.Address(), initialBalance)
	statedb.SetBalance(eoaAddr.Address(), initialBalance)
	statedb.Finalise(true)
	var (
		env = NewEVM(BlockContext{BlockNumber: big.NewInt(1)}, TxContext{}, statedb, params.TestChainConfig, Config{})
		gas = uint64(1000000000)
	)
	require.Equal(t, true, env.chainRules.IsEIP150)
	// EOA
	contractA := NewContract(eoaAddr, contractAddressA, initialBalance, gas)
	precompileContract := NewContract(contractA, precompileContractAddress, big.NewInt(0), gas)

	// setup precompile module
	var functions []*contract.StatefulPrecompileFunction

	contractSelector := contract.MustCalculateFunctionSelector("test()")
	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contractSelector,
		test,
	))

	// Construct the contract with functions.
	precompile, err := contract.NewStatefulPrecompileContract(functions)
	require.NoError(t, err)

	module := modules.Module{
		Address:  precompileContractAddress.Address(),
		Contract: precompile,
	}

	modules.RegisterModule(module)

	caller, _, err := env.interpreter.evm.DelegateCall(contractA, precompileContract.Address(), contractSelector, gas)
	require.NoError(t, err)

	// validation
	// should subtract the EOA's balance
	require.NotEqual(t, statedb.GetBalance(eoaAddr.Address()), initialBalance)
	// should not subtract contractA's balance because we are delegating call to precompileContract
	require.Equal(t, statedb.GetBalance(contractAddressA.Address()), initialBalance)
	// must be equal to EOA address because this is a DelegateCall, the context should be: caller of contractA, not contractA itself
	require.NotEqual(t, contractAddressA.Address().Bytes(), caller)
	require.Equal(t, eoaAddr.Address().Bytes(), caller)
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
	stateDb := accessibleState.GetStateDB()
	// do something really bad to drain / steal caller's funds without spending anything on EOA address
	stateDb.SubBalance(caller, big.NewInt(5))
	return caller[:], 0, nil
}
