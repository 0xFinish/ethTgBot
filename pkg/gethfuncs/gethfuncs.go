package gethfuncs

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var infura = os.Getenv("INFURA_API")
var goerliInfura = os.Getenv("INFURA_GOERLI")

var client *ethclient.Client

type TransactionInfo struct {
	TxInfo    *types.Transaction
	TxReceipt *types.Receipt
	TxBlock   *types.Block
}

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
	var TotalGasSpent big.Float
	var wg sync.WaitGroup
	for i, tx := range transactions {
		wg.Add(1)
		go func(tx *types.Transaction, i int) {
			defer wg.Done()
			fmt.Printf("I am goroutine running concurrently my num is %d \n", i)
			TxReceipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				log.Fatal(err)
			}
			var TotalGasSpentWei big.Int
			var baseFeePlusPriorityFeeWei big.Int
			baseFeePlusPriorityFeeWei.Add(tx.EffectiveGasTipValue(blockData.BaseFee()), blockData.BaseFee())
			TotalGasSpentWei.Mul(big.NewInt(int64(TxReceipt.GasUsed)), &baseFeePlusPriorityFeeWei)
			GasWeiFloat := new(big.Float).SetInt(&TotalGasSpentWei)
			var TotalGasSpentEth big.Float
			TotalGasSpentEth.Quo(GasWeiFloat, big.NewFloat(1e18))
			TotalGasSpent.Add(&TotalGasSpent, &TotalGasSpentEth)
		}(tx, i)
	}
	wg.Wait()
	gasSpent = TotalGasSpent.String() + " ETH"
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
			TxReceipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				log.Fatal(err)
			}
			var TotalGasSpentWei big.Int
			var baseFeePlusPriorityFeeWei big.Int
			baseFeePlusPriorityFeeWei.Add(tx.EffectiveGasTipValue(blockData.BaseFee()), blockData.BaseFee())
			TotalGasSpentWei.Mul(big.NewInt(int64(TxReceipt.GasUsed)), &baseFeePlusPriorityFeeWei)
			GasWeiFloat := new(big.Float).SetInt(&TotalGasSpentWei)
			var TotalGasSpentEth big.Float
			TotalGasSpentEth.Quo(GasWeiFloat, big.NewFloat(1e18))
			if TotalGasSpentEth.Cmp(&findBiggest) == 1 {
				fmt.Printf("The bigger gas spender found \n address:%s, gasSpent: %s", tx.Hash().Hex(), TotalGasSpentEth.String())
				findBiggest = TotalGasSpentEth
				address = tx.Hash().Hex()
			}
			fmt.Printf("I am goroutine DONE my num is %d \n", i)
		}(tx, i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Time elapsed: %s\n", elapsed)
	spenderAddress = findBiggest.String() + " ETH " + " \n " + address
	return
}

func GetTransactionFee(transaction string) (gasSpentMessage string, PureGasValue big.Float) {
	var transactionStruct TransactionInfo
	var wg sync.WaitGroup
	var err error
	txHash := common.HexToHash(transaction)
	wg.Add(2)
	go func(txHash common.Hash) {
		defer wg.Done()
		fmt.Println("I am goroutine running concurrently to get TxInfo")
		transactionStruct.TxInfo, _, err = client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("I am goroutine DONE concurrently to get TxInfo")
	}(txHash)
	go func(txHash common.Hash) {
		defer wg.Done()
		fmt.Println("I am goroutine running concurrently to get TxReceipt")
		transactionStruct.TxReceipt, err = client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("I am goroutine DONE concurrently to get TxReceipt")
	}(txHash)
	wg.Wait()
	blockData, err := client.BlockByNumber(context.Background(), transactionStruct.TxReceipt.BlockNumber)
	if err != nil {
		log.Fatal(err)
	}
	var TotalGasSpentWei big.Int
	var baseFeePlusPriorityFeeWei big.Int
	baseFeePlusPriorityFeeWei.Add(transactionStruct.TxInfo.EffectiveGasTipValue(blockData.BaseFee()), blockData.BaseFee())
	TotalGasSpentWei.Mul(big.NewInt(int64(transactionStruct.TxReceipt.GasUsed)), &baseFeePlusPriorityFeeWei)
	baseFeePlusPriorityFeeGwei := new(big.Float).SetInt(&baseFeePlusPriorityFeeWei)
	baseFeePlusPriorityFeeGwei.Quo(baseFeePlusPriorityFeeGwei, big.NewFloat(1000000000))
	GasWeiFloat := new(big.Float).SetInt(&TotalGasSpentWei)
	var TotalGasSpentEth big.Float
	TotalGasSpentEth.Quo(GasWeiFloat, big.NewFloat(1e18))
	gasSpentMessage = fmt.Sprintf("Total gas spent in ETH: %s \n gasUsedUint: %d \n gasPrice: %s \n gasPriceGwei: %s \n GasUsedEth => gasUsed * (baseFee + PriorityFee) == %d * (%s) / 1e18", TotalGasSpentEth.String(), transactionStruct.TxReceipt.GasUsed, baseFeePlusPriorityFeeWei.String(), baseFeePlusPriorityFeeGwei.String(), transactionStruct.TxReceipt.GasUsed, baseFeePlusPriorityFeeWei.String())
	PureGasValue = TotalGasSpentEth
	return
}

func GetTransactionSender(transaction string) (txSender string) {
	txHash := common.HexToHash(transaction)
	txInfo, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}
	txMessage, err := txInfo.AsMessage(types.LatestSignerForChainID(txInfo.ChainId()), big.NewInt(1))
	if err != nil {
		log.Fatal(err)
	}
	txSender = txMessage.From().String()
	return
}

func GetBiggestBlockWallet(block string) (biggestAddress string) {
	start := time.Now()
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
	var biggestBalance big.Int
	var biggestBalanceAddress string
	for _, tx := range transactions {
		wg.Add(1)
		go func(tx *types.Transaction) {
			defer wg.Done()
			txMessage, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), big.NewInt(1))
			if err != nil {
				log.Fatal(err)
			}
			balance, err := client.BalanceAt(context.Background(), txMessage.From(), bigBlockInt)
			if err != nil {
				log.Fatal(err)
			}
			if balance.Cmp(&biggestBalance) == 1 {
				// fmt.Printf("The bigger gas spender found \n address:%s, gasSpent: %s", tx.Hash().Hex(), TotalGasSpentEth.String())
				biggestBalance = *balance
				biggestBalanceAddress = txMessage.From().Hex()
			}
		}(tx)
	}
	wg.Wait()
	GasWeiFloat := new(big.Float).SetInt(&biggestBalance)
	var biggestBalanceFloat big.Float
	biggestBalanceFloat.Quo(GasWeiFloat, big.NewFloat(1e18))
	elapsed := time.Since(start)
	fmt.Printf("Time elapsed: %s\n", elapsed)
	biggestAddress = fmt.Sprintf("The biggest account: %s \n with balance: %s ETH \n in the block %d ", biggestBalanceAddress, biggestBalanceFloat.String(), blockInt)
	return
}

func GetAddressInfo(address string) (info string) {
	addressHash := common.HexToAddress(address)
	balanceWei, err := client.BalanceAt(context.Background(), addressHash, nil)
	if err != nil {
		log.Fatal(err)
	}
	GasWeiFloat := new(big.Float).SetInt(balanceWei)
	var biggestBalanceFloat big.Float
	biggestBalanceFloat.Quo(GasWeiFloat, big.NewFloat(1e18))
	info = fmt.Sprintf("The account balance of the address is %s wei", biggestBalanceFloat.String())
	return
}

func SendTransaction(address string) (info string) {
	goerliClient, err := ethclient.Dial(goerliInfura)
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA(os.Getenv("GOERLI_TESTNET_WALLET_PK"))
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("can not assert type")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := goerliClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	value := big.NewInt(1e3)
	gasLimit := uint64(21000)
	gasPrice, err := goerliClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	toAddress := common.HexToAddress(address)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	chainId, err := goerliClient.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	err = goerliClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	return
}
