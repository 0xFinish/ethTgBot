package smartContracts

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fi9ish/ethTgBot/pkg/store"
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
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, _, err := store.DeployStore(auth, goerliClient, input)
	if err != nil {
		log.Fatal(err)
	}

	info = fmt.Sprintln(address.Hex()) + fmt.Sprintln(tx.Hash().Hex())
	return
}
