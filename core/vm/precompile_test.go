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
	evm := NewEVM(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{})

	var (
		existingContractAddress   = common.HexToAddress("0x0300000000000000000000000000000000000000")
		unexistingContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000001")
	)

	// check that precompile doesn't exist before registration
	precompiledContract, ok := evm.precompile(existingContractAddress)
	require.False(t, ok)
	require.Nil(t, precompiledContract)

	module := modules.Module{
		Address:  existingContractAddress,
		Contract: new(mockStatefulPrecompiledContract),
	}
	err := modules.RegisterModule(module)
	require.NoError(t, err)

	// check that precompile exists after registration
	precompiledContract, ok = evm.precompile(existingContractAddress)
	require.True(t, ok)
	require.NotNil(t, precompiledContract)

	// check that precompile doesn't exist
	precompiledContract, ok = evm.precompile(unexistingContractAddress)
	require.False(t, ok)
	require.Nil(t, precompiledContract)
}
