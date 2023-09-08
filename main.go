// Originally written by @holiman, minor adjustments by me.

package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	go func() {
		logTime := time.Now()
		for {
			if time.Since(logTime) > time.Second*5 {
				fmt.Printf("Did %d attempts in %v, best score is %d\n", count.Load(), time.Since(logTime), highscore.Load())
				logTime = time.Now()
			}
		}
	}()
	for i := 0; i < runtime.NumCPU()-1; i++ {
		go brute(common.FromHex("0xbeac02000000000000000000"))
	}
	brute(common.FromHex("0xbeac02000000000000000000"))
}

var highscore atomic.Int64
var count atomic.Uint64

func brute(target []byte) {
	var (
		inner = types.LegacyTx{
			Nonce:    0,
			GasPrice: newGwei(1000),
			Gas:      250000,
			To:       nil,
			Value:    big.NewInt(0),
			Data:     common.FromHex("0x60588060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604457602036146024575f5ffd5b620180005f350680545f35146037575f5ffd5b6201800001545f5260205ff35b6201800042064281555f359062018000015500"),
			V:        big.NewInt(27),
			R:        big.NewInt(0x539),
			S:        big.NewInt(0x1337),
		}
		tx        = types.NewTx(&inner)
		signer    = types.NewEIP155Signer(big.NewInt(1))
		bigbig, _ = new(big.Int).SetString("0x1337000000000000000000", 0)
		u64       = make([]byte, 8)
	)
	for {
		sender, err := types.Sender(signer, tx)
		if err != nil {
			panic(err)
		}
		addr := crypto.CreateAddress(sender, 0)
		score := int64(compare(target[:], addr[:]))
		if score > highscore.Load() {
			highscore.Store(score)
			txjson, _ := json.MarshalIndent(tx, "", "  ")
			fmt.Printf("New highscore: %d\nAddress: %v\nTx:\n%v\n", score,
				addr, string(txjson))
		}
		crand.Read(u64)
		inner.S = new(big.Int).SetUint64(binary.BigEndian.Uint64(u64))
		inner.S.Add(inner.S, bigbig)
		tx = types.NewTx(&inner)
		count.Add(1)
	}
}

func compare(a, b []byte) int {
	for i, x := range a {
		y := b[i]
		if (x & 0xf0) != (y & 0xf0) {
			return 2 * i
		}
		if (x & 0xf) != (y & 0xf) {
			return 2*i + 1
		}
	}
	return len(a)
}

func newGwei(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), big.NewInt(params.GWei))
}