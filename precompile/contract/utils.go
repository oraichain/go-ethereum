package contract

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// Gas costs for stateful precompiles
const (
	WriteGasCostPerSlot = 20_000
	ReadGasCostPerSlot  = 5_000
)

func CalculateFunctionSelector(functionSignature string) []byte {
	hash := crypto.Keccak256([]byte(functionSignature))
	return hash[:4]
}

// DeductGas checks if [suppliedGas] is sufficient against [requiredGas] and deducts [requiredGas] from [suppliedGas].
func DeductGas(suppliedGas uint64, requiredGas uint64) (uint64, error) {
	if suppliedGas < requiredGas {
		return 0, fmt.Errorf("out of gas, supplied: %v, required: %v", suppliedGas, requiredGas)
	}
	return suppliedGas - requiredGas, nil
}
