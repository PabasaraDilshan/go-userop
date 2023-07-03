package userop

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	config "userop-builder/config"
	contracts "userop-builder/contracts"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)


type UserOpBuilder struct{
    accountAddress *common.Address
    ownerAddress *common.Address
    client *ethclient.Client
    GasEstimator *GasEstimator
}



func NewUserOpBuilder(owner common.Address, client *ethclient.Client) *UserOpBuilder{
    uob := UserOpBuilder{
        ownerAddress: &owner,
        client: client,
    }
    ov := GasOverheads{
            fixed: 21000,
            perUserOp: 22874,
            perUserOpWord: 25,
            batchFixed:816,
            zeroByte: 4,
            nonZeroByte: 16,
            bundleSize: 1,
            sigSize: 65,
            paymasterDataSize: 174,
    }
    uob.GasEstimator = &GasEstimator{
        client: client,
        ov: &ov,
    }
    return &uob
}
func (uob *UserOpBuilder) CreateUnsignedUserOp() *UserOperation{
    fmt.Println("Point1")
    accAddress := uob.GetAccountAddress()
    fmt.Println("Point2",accAddress)
    initCode := uob.GetAccountInitCode()
    fmt.Println("Point3",initCode)
    initGas := uob.GasEstimator.EstimateCreationGas(initCode)
    fmt.Println("Point4",initGas)
    gasPrice := uob.GasEstimator.GetGasPrice()
    fmt.Println("Point5")
    nonce := uob.GetNone()
    fmt.Println("Point6")
    verificationGasLimit := uob.GasEstimator.GetVerificationGasLimit(*initGas)
    fmt.Println("Point7")
    callGasLimit := uob.GasEstimator.EstimateCallDataGasLimit(accAddress,make([]byte, 0))
    fmt.Println("Point8")
    uop := UserOperation{
        Sender: *accAddress,
        InitCode: initCode,
        MaxFeePerGas: gasPrice.MaxFeePerGas,
        MaxPriorityFeePerGas: gasPrice.MaxPriorityFeePerGas,
        Nonce: nonce,
        VerificationGasLimit: verificationGasLimit,
        CallGasLimit: callGasLimit,
        CallData: make([]byte, 0),
        PreVerificationGas: common.Big0,
        PaymasterAndData: make([]byte, 0),
        Signature: make([]byte, 0),

    }
    preVerificationGas := uob.GasEstimator.CalcPreverificationGas(uop)
    uop.PreVerificationGas = preVerificationGas 

    now := time.Now()

	// Calculate the validAfter and validUntil dates
	validAfter := now.AddDate(0, 0, -5)
	validUntil := now.AddDate(0, 0, 5)

	// Convert the dates to Unix timestamps
	validAfterUnix := validAfter.Unix()
	validUntilUnix := validUntil.Unix()
    hash := uob.getPaymasterHash(&uop,*new(big.Int).SetInt64(validUntilUnix),*new(big.Int).SetInt64(validAfterUnix))
    fmt.Println("Point9",common.Bytes2Hex(hash[:]) )
	privateKey, err := crypto.HexToECDSA("1a0ff12e07e13b32e8e6ef3965e3af9140a39180523f37a262ca2473e681d13b")
	if err != nil {
		log.Fatal(err)
	}

	// Sign the hash with the private key
	signature, err := crypto.Sign(hash[:], privateKey)
    fmt.Println("Point10",common.Bytes2Hex(signature) )
	if err != nil {
		log.Fatal(err)
	}
    
    paymasterData := bytes.Join([][]byte{
        common.FromHex(config.TokenPaymaster),
        boolToHex(true),common.LeftPadBytes(big.NewInt(100).Bytes(), 4),
        common.FromHex("0xB469dbEe15A3e7bf37f6aaBE0F6362b362c65246"),
        common.LeftPadBytes(big.NewInt(validUntilUnix).Bytes(),32),
        common.LeftPadBytes(big.NewInt(validAfterUnix).Bytes(),32), 
        signature }, []byte{})

    fmt.Println("Point10",common.Bytes2Hex(paymasterData) )
    uop.PaymasterAndData = paymasterData
    return &uop
}
func boolToHex(b bool) []byte {
	if b {
		return []byte{0x01}
	}
	return  []byte{0x00}
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
func (uob *UserOpBuilder) getPaymasterHash(userop *UserOperation,validUntil big.Int,validAfter big.Int) [32]byte{
    client := uob.client
	
    instance, err := contracts.NewPaymaster(common.HexToAddress(config.TokenPaymaster), client)
    if err != nil {
        log.Fatal(err)
    }
    uo := contracts.UserOperation{
        Sender: userop.Sender,
        Nonce: userop.Nonce,
        InitCode: userop.InitCode,
        CallData: userop.CallData,
        CallGasLimit: userop.CallGasLimit,
        VerificationGasLimit: userop.VerificationGasLimit,
        PreVerificationGas: userop.PreVerificationGas,
        MaxFeePerGas: userop.MaxFeePerGas,
        MaxPriorityFeePerGas: userop.MaxPriorityFeePerGas,
        PaymasterAndData: userop.PaymasterAndData,
        Signature: userop.Signature,
    }


    hash, err := instance.GetHash(&bind.CallOpts{},uo,true,10,common.HexToAddress("0xB469dbEe15A3e7bf37f6aaBE0F6362b362c65246"),&validUntil,&validAfter)
    if err != nil {
        log.Fatal(err)
    }
    return hash
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

