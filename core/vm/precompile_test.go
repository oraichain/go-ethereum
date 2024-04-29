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
	var address = common.HexToAddress("0x0300000000000000000000000000000000000000")

	t.Run("test NewEVM constructor", func(t *testing.T) {
		modules.ClearRegisteredModules()
		evm := NewEVM(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{})
		require.Nil(t, evm.enabledPrecompiles)
	})

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

	addrs := func(hexAddrs []string) []common.Address {
		addrs := make([]common.Address, len(hexAddrs))
		for i, hexAddr := range hexAddrs {
			addrs[i] = common.HexToAddress(hexAddr)
		}
		return addrs
	}

	for _, tc := range []struct {
		desc            string
		registeredAddrs []common.Address
		enabledAddrs    []common.Address
		foundAddrs      []common.Address
		notFoundAddrs   []common.Address
	}{
		{
			desc:            "test case #1",
			registeredAddrs: addrs([]string{"0x101", "0x102", "0x103", "0x104", "0x105"}),
			enabledAddrs:    addrs([]string{"0x101", "0x103", "0x105"}),
			foundAddrs:      addrs([]string{"0x101", "0x103", "0x105"}),
			notFoundAddrs:   addrs([]string{"0x102", "0x104", "0x106"}),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			modules.ClearRegisteredModules()
			evm := NewEVMWithEnabledPrecompiles(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{}, tc.enabledAddrs)

			for _, registeredAddr := range tc.registeredAddrs {
				module := modules.Module{
					Address:  registeredAddr,
					Contract: new(mockStatefulPrecompiledContract),
				}
				err := modules.RegisterModule(module)
				require.NoError(t, err)
			}

			for _, addr := range tc.foundAddrs {
				precompile, ok := evm.precompile(addr)
				require.True(t, ok)
				require.NotNil(t, precompile)
			}

			for _, addr := range tc.notFoundAddrs {
				precompile, ok := evm.precompile(addr)
				require.False(t, ok)
				require.Nil(t, precompile)
			}
		})
	}
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
