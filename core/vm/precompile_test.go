package vm

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestActivePrecompiles(t *testing.T) {
	genesisTime := time.Now()
	getBlockTime := func(height *big.Int) *uint64 {
		if !height.IsInt64() {
			t.Fatalf("expected height bounded to int64")
		}
		totalBlockSeconds := time.Duration(10*height.Int64()) * time.Second
		blockTimeUnix := uint64(genesisTime.Add(totalBlockSeconds).Unix())

		return &blockTimeUnix
	}
	chainConfig := params.TestChainConfig
	chainConfig.HomesteadBlock = big.NewInt(1)
	chainConfig.ByzantiumBlock = big.NewInt(2)
	chainConfig.IstanbulBlock = big.NewInt(3)
	chainConfig.BerlinBlock = big.NewInt(4)
	chainConfig.CancunTime = getBlockTime(big.NewInt(5))

	testCases := []struct {
		name  string
		block *big.Int
	}{
		{"homestead", chainConfig.HomesteadBlock},
		{"byzantium", chainConfig.ByzantiumBlock},
		{"istanbul", chainConfig.IstanbulBlock},
		{"berlin", chainConfig.BerlinBlock},
		{"cancun", new(big.Int).Add(chainConfig.BerlinBlock, big.NewInt(1))},
	}

	// custom precompile address used for test
	contractAddress := common.HexToAddress("0x0400000000000000000000000000000000000000")

	// ensure we are not being shadowed by a core preompile address
	for _, tc := range testCases {
		rules := chainConfig.Rules(tc.block, false, *getBlockTime(tc.block))

		for _, precompileAddr := range ActivePrecompiles(rules) {
			if precompileAddr == contractAddress {
				t.Fatalf("expected precompile %s to not be returned in %s block", contractAddress, tc.name)
			}
		}
	}

	// register the precompile
	module := modules.Module{
		Address:  contractAddress,
		Contract: new(mockStatefulPrecompiledContract),
	}

	// TODO: should we allow dynamic registration to update ActivePrecompiles?
	// Or should we enforce registration only at init?
	err := modules.RegisterModule(module)
	require.NoError(t, err, "could not register precompile for test")

	for _, tc := range testCases {
		rules := chainConfig.Rules(tc.block, false, *getBlockTime(tc.block))

		exists := false
		for _, precompileAddr := range ActivePrecompiles(rules) {
			if precompileAddr == contractAddress {
				exists = true
			}
		}

		assert.True(t, exists, "expected %s block to include active stateful precompile %s", tc.name, contractAddress)
	}
}
