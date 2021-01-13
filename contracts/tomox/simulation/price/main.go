package main

import (
	"context"
	"fmt"
	"github.com/tomochain/tomochain/core/state"
	"github.com/tomochain/tomochain/tomoxlending/lendingstate"
	"github.com/tomochain/tomochain/contracts/tomox/simulation"
	"math/big"
	"os"
	"time"

	"github.com/tomochain/tomochain/accounts/abi/bind"
	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/contracts/tomox"
	"github.com/tomochain/tomochain/crypto"
	"github.com/tomochain/tomochain/ethclient"
)

func main() {
	client, err := ethclient.Dial(simulation.RpcEndpoint)
	if err != nil {
		fmt.Println(err, client)
	}
	MainKey, _ := crypto.HexToECDSA(os.Getenv("MAIN_ADDRESS_KEY"))
	MainAddr := crypto.PubkeyToAddress(MainKey.PublicKey)

	nonce, _ := client.NonceAt(context.Background(), MainAddr, nil)
	auth := bind.NewKeyedTransactor(MainKey)
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(4000000) // in units
	auth.GasPrice = big.NewInt(250000000000000)

	// init trc21 issuer
	auth.Nonce = big.NewInt(int64(nonce))

	price := new(big.Int)
	price.SetString("90000000000000000000000", 10)

	lendContract, _ := tomox.NewLendingRelayerRegistration(auth, common.HexToAddress("0x4d7eA2cE949216D6b120f3AA10164173615A2b6C"), client)

	token := common.HexToAddress(os.Getenv("TOKEN_ADDRESS"))
	lendingToken := common.HexToAddress(os.Getenv("LENDING_TOKEN_ADDRESS"))

	tx, err := lendContract.SetCollateralPrice(token, lendingToken, price)
	if err != nil {
		fmt.Println("Set price failed!", err)
	}

	time.Sleep(5 * time.Second)
	r, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		fmt.Println("Get receipt failed", err)
	}
	fmt.Println("Done receipt status", r.Status)

	collateralState := state.GetLocMappingAtKey(token.Hash(), lendingstate.CollateralMapSlot)
	locMapPrices := collateralState.Add(collateralState, lendingstate.CollateralStructSlots["price"])
	locLendingTokenPriceByte := crypto.Keccak256(lendingToken.Hash().Bytes(), common.BigToHash(locMapPrices).Bytes())

	locCollateralPrice := common.BigToHash(new(big.Int).Add(new(big.Int).SetBytes(locLendingTokenPriceByte), lendingstate.PriceStructSlots["price"]))
	locBlockNumber := common.BigToHash(new(big.Int).Add(new(big.Int).SetBytes(locLendingTokenPriceByte), lendingstate.PriceStructSlots["blockNumber"]))

	priceByte, err := client.StorageAt(context.Background(), common.HexToAddress(os.Getenv("LENDING_ADDRESS")), locCollateralPrice, nil)
	fmt.Println(new(big.Int).SetBytes(priceByte), err)
	blockNumberByte, err := client.StorageAt(context.Background(), common.HexToAddress(os.Getenv("LENDING_ADDRESS")), locBlockNumber, nil)
	fmt.Println(new(big.Int).SetBytes(blockNumberByte), err)
}
