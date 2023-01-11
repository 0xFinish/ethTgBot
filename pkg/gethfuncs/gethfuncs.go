package gethfuncs

import (
	"bytes"
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
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

var infura = os.Getenv("INFURA_API")
var infuraWSS = os.Getenv("INFURA_WSS")
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

func SendTransactionEth(address string) (info string) {
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
	info = fmt.Sprintf("tx sent: %s", signedTx.Hash().Hex())
	return
}

// func SendTransactionERC20(address string) (info string) {
// 	goerliClient, err := ethclient.Dial(goerliInfura)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	privateKey, err := crypto.HexToECDSA(os.Getenv("GOERLI_TESTNET_WALLET_PK"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	publicKey := privateKey.Public()
// 	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
// 	if !ok {
// 		log.Fatal("can not assert type")
// 	}
// 	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
// 	nonce, err := goerliClient.PendingNonceAt(context.Background(), fromAddress)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	value := big.NewInt(0)
// 	gasPrice, err := client.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	toAddress := common.HexToAddress(address)
// 	tokenAddress := common.HexToAddress("0x326C977E6efc84E512bB9C30f76E30c160eD06FB")

// 	transferFnSignature := []byte("transfer(address,uint256)")
// 	hash := sha3.NewLegacyKeccak256()
// 	hash.Write(transferFnSignature)
// 	methodID := hash.Sum(nil)[:4]
// 	fmt.Println(hexutil.Encode(methodID))

// 	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
// 	fmt.Println(hexutil.Encode(paddedAddress))

// 	amount := new(big.Int)
// 	amount.SetString("1000000", 10)
// 	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
// 	var data []byte
// 	data = append(data, methodID...)
// 	data = append(data, paddedAddress...)
// 	data = append(data, paddedAmount...)

// 	gasLimit, err := goerliClient.EstimateGas(context.Background(), ethereum.CallMsg{
// 		To:   &tokenAddress,
// 		Data: data,
// 	})
// 	if err != nil {
// 		fmt.Println("the error is here3")
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Gas limit %d", gasLimit)
// 	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
// 	chainID, err := client.NetworkID(context.Background())
// 	if err != nil {
// 		fmt.Println("the error is here1")

// 		log.Fatal(err)
// 	}
// 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
// 	if err != nil {
// 		fmt.Println("the error is here2")

// 		log.Fatal(err)
// 	}
// 	err = goerliClient.SendTransaction(context.Background(), signedTx)
// 	if err != nil {
// 		fmt.Println("the error is here")
// 		log.Fatal(err)
// 	}
// 	info = fmt.Sprintf("tx sent: %s", signedTx.Hash().Hex())
// 	return
// }

func SendTransactionERC20(address string) (info string) {
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
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := goerliClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(0) // in wei (0 eth)
	gasPrice, err := goerliClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(address)
	// tokenAddress := common.HexToAddress("0x326C977E6efc84E512bB9C30f76E30c160eD06FB")

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	amount := new(big.Int)
	amount.SetString("10000", 10) // sets the value to 1000 tokens, in the token denomination

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	// gasLimit, err := goerliClient.EstimateGas(context.Background(), ethereum.CallMsg{
	// 	To:   &tokenAddress,
	// 	Data: data,
	// })
	// if err != nil {
	// 	fmt.Println("The error is here")
	// 	log.Fatal(err)
	// }
	// fmt.Println(gasLimit) // 23256
	gasLimit := uint64(23256)

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := goerliClient.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = goerliClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	info = fmt.Sprintf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0xa56316b637a94c4cc0331c73ef26389d6c097506d581073f927275e7a6ece0bc
	return
}

func NewBlockInfoWSS() (info string) {
	clientWSS, err := ethclient.Dial(infuraWSS)
	if err != nil {
		log.Fatal(err)
	}
	headers := make(chan *types.Header)
	sub, err := clientWSS.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex())
			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}
			info = block.Hash().Hex() + " " + block.Number().String() // 0xbc10defa8dda384c96a17640d84de5578804945d347072e091b4e5f390ddea7f
			fmt.Println(block.Time())                                 // 1529525947
			fmt.Println(block.Nonce())                                // 130524141876765836
			fmt.Println(len(block.Transactions()))                    // 7
		}
	}
}

func RawTransaction(address string) (info string) {
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
	ts := types.Transactions{signedTx}
	b := new(bytes.Buffer)
	ts.EncodeIndex(0, b)
	rawTxBytes := b.Bytes()
	txFromRaw := new(types.Transaction)
	rlp.DecodeBytes(rawTxBytes, &txFromRaw)
	err = goerliClient.SendTransaction(context.Background(), txFromRaw)
	if err != nil {
		log.Fatal(err)
	}
	info = fmt.Sprintf("tx sent: %s", signedTx.Hash().Hex())
	return
}
