// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package modules

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestInsertSortedByAddress(t *testing.T) {
	clearRegisteredModules()

	data := make([]Module, 0)
	// test that the module is registered in sorted order
	module1 := Module{
		Address: common.BigToAddress(big.NewInt(1)),
	}
	data = insertSortedByAddress(data, module1)
	require.Equal(t, []Module{module1}, data)

	module0 := Module{
		Address: common.BigToAddress(big.NewInt(0)),
	}
	data = insertSortedByAddress(data, module0)
	require.Equal(t, []Module{module0, module1}, data)

	module3 := Module{
		Address: common.BigToAddress(big.NewInt(3)),
	}
	data = insertSortedByAddress(data, module3)
	require.Equal(t, []Module{module0, module1, module3}, data)

	module2 := Module{
		Address: common.BigToAddress(big.NewInt(2)),
	}
	data = insertSortedByAddress(data, module2)
	require.Equal(t, []Module{module0, module1, module2, module3}, data)
}

func TestRegisterModule(t *testing.T) {
	clearRegisteredModules()

	const moduleNum = 4
	// create modules
	modules := make([]Module, moduleNum)
	for i := 0; i < moduleNum; i++ {
		modules[i] = Module{
			Address: common.BigToAddress(big.NewInt(int64(i))),
		}
	}

	// register modules
	for i := 0; i < moduleNum; i++ {
		err := RegisterModule(modules[i])
		require.NoError(t, err)
	}

	// get modules by address
	for i := 0; i < moduleNum; i++ {
		address := common.BigToAddress(big.NewInt(int64(i)))
		module, exists := GetPrecompileModuleByAddress(address)
		require.True(t, exists)
		require.Equal(t, module, Module{
			Address: address,
		})
	}

	// get unexisting module by address
	address := common.BigToAddress(big.NewInt(moduleNum))
	_, exists := GetPrecompileModuleByAddress(address)
	require.False(t, exists)

	// get all modules
	registeredModules := RegisteredModules()
	require.Equal(t, modules, registeredModules)
}

func TestRegisterModuleWithDuplicateAddress(t *testing.T) {
	clearRegisteredModules()

	modules := []Module{
		{
			Address: common.BigToAddress(big.NewInt(0)),
		},
	}

	err := RegisterModule(modules[0])
	require.NoError(t, err)

	err = RegisterModule(modules[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "address 0x0000000000000000000000000000000000000000 already used by a stateful precompile")

	// get all modules
	registeredModules := RegisteredModules()
	require.Equal(t, modules, registeredModules)
}

func clearRegisteredModules() {
	registeredModules = make([]Module, 0)
}
