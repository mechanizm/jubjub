package main

import (
	"encoding/hex"
	"log"

	"github.com/mechanizm/jubjub/extended"
	"github.com/mechanizm/jubjub/fr"
)

func main() {
	addressBytes, err := hex.DecodeString("7d09bb9aa97704719c33d1f6e7ed7e8d6c0edad0a02f7af82ab77ebc104f5f1e")
	if err != nil {
		panic(err)
	}

	point := extended.FromBytes(addressBytes)
	// mod_bytes := []byte{14, 125, 180, 234, 101, 51, 175, 169, 6, 103, 59, 1, 1, 52, 59, 0, 166, 104, 32, 147, 204, 200, 16, 130, 208, 151, 14, 94, 214, 247, 44, 183}
	point = point.Mul(fr.MODULUS.BytesNotCanonical())

	log.Printf("[DEBUG] point is %v, is identity %t", point.StringNotCanonical(), point.IsIdentity())
}
