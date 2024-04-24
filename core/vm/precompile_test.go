package vm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/ethereum/go-ethereum/precompile/modules"
)

type mockStatefulPrecompiledContract struct{}

func (c *mockStatefulPrecompiledContract) Run(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	return []byte{}, 0, nil
}

func TestEvmIsPrecompileMethod(t *testing.T) {
	var address = common.HexToAddress("0x0300000000000000000000000000000000000000")

	t.Run("not registered and not enabled", func(t *testing.T) {
		modules.ClearRegisteredModules()
		evm := NewEVMWithEnabledPrecompiles(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{}, nil)

		precompile, ok := evm.precompile(address)
		require.False(t, ok)
		require.Nil(t, precompile)
	})

	t.Run("registered but not enabled", func(t *testing.T) {
		modules.ClearRegisteredModules()
		evm := NewEVMWithEnabledPrecompiles(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{}, nil)

		module := modules.Module{
			Address:  address,
			Contract: new(mockStatefulPrecompiledContract),
		}
		err := modules.RegisterModule(module)
		require.NoError(t, err)

		precompile, ok := evm.precompile(address)
		require.False(t, ok)
		require.Nil(t, precompile)
	})

	t.Run("not registered but enabled", func(t *testing.T) {
		modules.ClearRegisteredModules()
		enabledPrecompiles := []common.Address{address}
		evm := NewEVMWithEnabledPrecompiles(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{}, enabledPrecompiles)

		precompile, ok := evm.precompile(address)
		require.False(t, ok)
		require.Nil(t, precompile)
	})

	t.Run("registered and enabled", func(t *testing.T) {
		modules.ClearRegisteredModules()
		enabledPrecompiles := []common.Address{address}
		evm := NewEVMWithEnabledPrecompiles(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{}, enabledPrecompiles)

		module := modules.Module{
			Address:  address,
			Contract: new(mockStatefulPrecompiledContract),
		}
		err := modules.RegisterModule(module)
		require.NoError(t, err)

		precompile, ok := evm.precompile(address)
		require.True(t, ok)
		require.NotNil(t, precompile)
	})

	// TODO(yevhenii): add more complex scenarios?
}
