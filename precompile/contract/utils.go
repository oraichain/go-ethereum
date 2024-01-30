package contract

import (
	"github.com/ethereum/go-ethereum/crypto"
)

func CalculateFunctionSelector(functionSignature string) []byte {
	hash := crypto.Keccak256([]byte(functionSignature))
	return hash[:4]
}
