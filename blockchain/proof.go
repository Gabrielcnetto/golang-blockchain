package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//take data from block
//create counter for nouce init. with 0
//create hash of the data +nouce
//check if hash it meet set of requirements

/*
the first few bytes must contain 0s
*/

const difficult = 20

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	nounce := 0
	for nounce < math.MaxInt64 {
		data := pow.InitData(nounce)
		hash = sha256.Sum256(data)

		intHash.SetBytes(hash[:])
		if intHash.Cmp(pow.Target) == -1 {
			fmt.Printf("\n%x\n", hash)
			break
		} else {
			nounce++
		}
	}
	fmt.Println()
	return nounce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int
	data := pow.InitData(pow.Block.Nounce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	return intHash.Cmp(pow.Target) == -1
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficult))
	pow := &ProofOfWork{
		Block:  b,
		Target: target,
	}
	return pow
}

func (pow *ProofOfWork) InitData(nounce int) []byte {
	return bytes.Join([][]byte{
		pow.Block.PrevHash,
		pow.Block.Data,
		ToHex(int64(nounce)),
		ToHex(int64(difficult)),
	}, []byte{})
}

func ToHex(nounce int64) []byte {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, nounce); err != nil {
		log.Panic(err)
	}
	return buff.Bytes()

}
