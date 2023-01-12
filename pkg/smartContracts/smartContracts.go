package smartContracts

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fi9ish/ethTgBot/pkg/store"
	"github.com/fi9ish/ethTgBot/pkg/token"
)

func DeploySmartContract() (info string) {
	goerliClient, err := ethclient.Dial(os.Getenv("INFURA_GOERLI"))
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

	gasPrice, err := goerliClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := goerliClient.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(500000)
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, _, err := store.DeployStore(auth, goerliClient, input)
	if err != nil {
		log.Fatal(err)
	}

	info = fmt.Sprintln(address.Hex()) + fmt.Sprintln(tx.Hash().Hex())
	return
}

func ReadContractInstance(address string) (info string) {
	goerliClient, err := ethclient.Dial(os.Getenv("INFURA_GOERLI"))
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress(address)
	instance, err := store.NewStore(contractAddress, goerliClient)
	if err != nil {
		log.Fatal(err)
	}

	version, err := instance.Version(nil)
	if err != nil {
		log.Fatal(err)
	}

	info = fmt.Sprintln(version)

	return
}

func WriteToContractInstance(address string) (info string) {
	goerliClient, err := ethclient.Dial(os.Getenv("INFURA_GOERLI"))
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
		log.Fatal("cannot assert type")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := goerliClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := goerliClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := goerliClient.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	contractAddress := common.HexToAddress(address)
	instance, err := store.NewStore(contractAddress, goerliClient)
	if err != nil {
		log.Fatal(err)
	}

	key := [32]byte{}
	value := [32]byte{}
	copy(key[:], []byte("foo"))
	copy(key[:], []byte("bar"))

	tx, err := instance.SetItem(auth, key, value)
	if err != nil {
		log.Fatal(err)
	}
	info = fmt.Sprintf("tx sent: %s\n", tx.Hash().Hex())

	result, err := instance.Items(nil, key)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(result[:])) // "bar"

	return
}

func ReadSmartContractBytecode(address string) (info string) {
	goerliClient, err := ethclient.Dial(os.Getenv("INFURA_GOERLI"))
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress(address)
	bytecode, err := goerliClient.CodeAt(context.Background(), contractAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(hex.EncodeToString(bytecode))

	return
}

func QueryingERC20(address string) (info string) {
	goerliClient, err := ethclient.Dial(os.Getenv("INFURA_GOERLI"))
	if err != nil {
		log.Fatal(err)
	}

	tokenAddress := common.HexToAddress(address)
	instance, err := token.NewToken(tokenAddress, goerliClient)
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
		log.Fatal("cannot assert type")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	bal, err := instance.BalanceOf(&bind.TransactOpts{}, fromAddress)
	if err != nil {
		fmt.Println("The error is here")
		log.Fatal(err)
	}

	fmt.Printf("wei: %s\n", bal)

	name, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("name: %s\n", name)
	fmt.Printf("symbol: %s\n", symbol)
	fmt.Printf("decimals: %v\n", decimals)

	return
}
