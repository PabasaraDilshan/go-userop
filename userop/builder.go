package userop

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	config "userop-builder/config"
	contracts "userop-builder/contracts"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)


type UserOpBuilder struct{
    accountAddress *common.Address
    ownerAddress *common.Address
    client *ethclient.Client
    gasEstimator *GasEstimator
}

func NewUserOpBuilder(owner common.Address, client *ethclient.Client) *UserOpBuilder{
    uob := UserOpBuilder{
        ownerAddress: &owner,
        client: client,
    }
    uob.gasEstimator = &GasEstimator{
        client: client,
    }
    return &uob
}

func (uob *UserOpBuilder) GetAccountAddress() *common.Address{
    uob.accountAddress = uob.GetCounterFactualAddress()
    return uob.accountAddress
}
func  (uob *UserOpBuilder) GetCallData(target common.Address,value big.Int,data []byte ) []byte{
    accABI,err := abi.JSON(strings.NewReader(contracts.GameAccountABI));
    encodedData, err := accABI.Pack("execute", uob.ownerAddress, value,data)
    if err != nil {
        log.Fatal(err)
    }
    return encodedData
}
    
func  (uob *UserOpBuilder) GetAccountInitCode() []byte{
    accFacABI,err := abi.JSON(strings.NewReader(contracts.AccountFactoryABI));
    data, err := accFacABI.Pack("createAccount", *uob.ownerAddress, common.Big0)
    if err != nil {
        log.Fatal(err)
    }
    return append(common.FromHex(config.AccountFactoryAddress),data...)
}
func (uob *UserOpBuilder) GetNone() *big.Int {
    instance,err := contracts.NewEntrypoint(common.HexToAddress(config.EntryPointAddress),uob.client)
    if err != nil {
        log.Fatal(err)
    }
    nonce,err := instance.GetNonce(&bind.CallOpts{},*uob.accountAddress,common.Big0)
    if err != nil {
        log.Fatal(err)
    }
    return nonce
}

func (uob *UserOpBuilder) GetCounterFactualAddress() *common.Address {
	// client, err := ethclient.Dial(ProviderUrl)
    client := uob.client
	accountFactoryAddress := common.HexToAddress(config.AccountFactoryAddress)
    instance, err := contracts.NewAccountFactory(accountFactoryAddress, client)
    if err != nil {
        log.Fatal(err)
    }


    address, err := instance.GetAddress(&bind.CallOpts{},*uob.ownerAddress,common.Big0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(address)
    return &address
}

