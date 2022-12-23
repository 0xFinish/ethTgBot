package gethfuncs

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var infura = os.Getenv("INFURA_API")

var client *ethclient.Client

func init() {
	var err error
	client, err = ethclient.Dial(infura)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateNewWallet() (newWallet string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	hexadecimalPrivateKey := hexutil.Encode(privateKeyBytes)[2:]
	fmt.Println(hexadecimalPrivateKey)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Cannot assert type!")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	newWallet = fmt.Sprintf("private key: %s \n public key/address: %s \n", hexadecimalPrivateKey, address)
	return
}

func GetCurrentBlockNum() (currentBlock string) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	currentBlock = header.Number.String()
	return
}

func GetGasSpent(block string) (gasSpent string) {
	start := time.Now()
	var transactionArray []string
	fmt.Println(block)
	blockInt, err := strconv.Atoi(block)
	if err != nil {
		log.Fatal(err)
	}
	bigBlockInt := big.NewInt(int64(blockInt))
	blockData, err := client.BlockByNumber(context.Background(), bigBlockInt)
	if err != nil {
		log.Fatal(err)
	}
	transactions := blockData.Transactions()
	var wg sync.WaitGroup
	for i, tx := range transactions {
		wg.Add(1)
		go func(tx *types.Transaction, i int) {
			defer wg.Done()
			fmt.Printf("I am goroutine running concurrently my num is %d \n", i)
			receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				log.Fatal(err)
			}
			transactionArray = append(transactionArray, tx.Hash().Hex())
			// transactionFeeUint := (big.NewInt(int64(receipt.GasUsed))  tx.GasPrice())
			transactionFeeUnit := big.NewInt(int64(receipt.GasUsed))
			var GasWei big.Int
			GasWei.Mul(transactionFeeUnit, tx.GasPrice())
			GasWeiFloat := new(big.Float).SetInt(&GasWei)
			var GasEth big.Float
			GasEth.Quo(GasWeiFloat, big.NewFloat(1e18))
			transactionArray = append(transactionArray, GasEth.String())
			fmt.Printf("I am goroutine DONE my num is %d \n", i)
		}(tx, i)
	}
	wg.Wait()
	gasSpent = strings.Join(transactionArray[:30], ", \n")
	elapsed := time.Since(start)
	fmt.Printf("Time elapsed: %s\n", elapsed)
	return
}

func GetBiggestGasSpender(block string) (spenderAddress string) {
	start := time.Now()
	blockInt, err := strconv.Atoi(block)
	if err != nil {
		log.Fatal(err)
	}
	bigBlockInt := big.NewInt(int64(blockInt))
	blockData, err := client.BlockByNumber(context.Background(), bigBlockInt)
	if err != nil {
		log.Fatal(err)
	}
	transactions := blockData.Transactions()
	var wg sync.WaitGroup
	var findBiggest big.Float
	var address string
	for i, tx := range transactions {
		wg.Add(1)
		go func(tx *types.Transaction, i int) {
			defer wg.Done()
			fmt.Printf("I am goroutine running concurrently my num is %d \n", i)
			receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				log.Fatal(err)
			}
			transactionFeeUnit := big.NewInt(int64(receipt.GasUsed))
			var GasWei big.Int
			GasWei.Mul(transactionFeeUnit, tx.GasPrice())
			GasWeiFloat := new(big.Float).SetInt(&GasWei)
			var GasEth big.Float
			GasEth.Quo(GasWeiFloat, big.NewFloat(1e18))
			if GasEth.Cmp(&findBiggest) == 1 {
				findBiggest = GasEth
				address = tx.Hash().Hex()
			}
			fmt.Printf("I am goroutine DONE my num is %d \n", i)
		}(tx, i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Time elapsed: %s\n", elapsed)
	spenderAddress = findBiggest.String() + " \n " + address
	return
}

func GetBiggestBlockWallet(block string) (biggestAddress string) {
	return
}

func GetAddressInfo(address string) (info string) {
	return
}
