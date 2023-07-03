package userop

import (
	"context"
	"log"
	"math/big"
	"userop-builder/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type GasEstimator struct{
    client *ethclient.Client
}
type GasPrice struct{
	maxFeePerGas *big.Int
	maxPriorityFeePerGas *big.Int
}


func (ge *GasEstimator) EstimateCreationGas(initCode []byte) *big.Int{
	if(len(initCode)==0){
		return common.Big0
	}
	deployerAddress := common.BytesToAddress(initCode[0:20])
	msg := ethereum.CallMsg{
		To:  &deployerAddress ,
		Data: initCode[20:],
	}

	gasLimit, err := ge.client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Fatal(err)
	}
	return new(big.Int).SetUint64(gasLimit)
}

func (ge *GasEstimator) EstimateCallDataGasLimit(sender *common.Address,callData []byte) *big.Int{
	msg := ethereum.CallMsg{
		From: common.HexToAddress(config.EntryPointAddress),
		To:  sender,
		Data: callData,
	}

	gasLimit, err := ge.client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Fatal(err)
	}
	gas := new(big.Int).SetUint64(gasLimit)
	return gas.Mul(gas,big.NewInt(5))
}


func (ge *GasEstimator) GetVerificationGasLimit(initGas big.Int) *big.Int{
	
	return new(big.Int).Add(big.NewInt(100000),&initGas)
}

func (ge *GasEstimator) GetGasPrice() GasPrice{
	var tip big.Int
	err := ge.client.Client().CallContext(context.Background(),&tip,"eth_maxPriorityFeePerGas")
	if err != nil {
		log.Fatal(err)
	}
	latestBlock, err := ge.client.BlockByNumber(context.Background(),common.Big0)
	if err != nil {
		log.Fatal(err)
	}
	buffer := new(big.Int).Div(&tip,big.NewInt(2))
	maxPriorityFeePerGas := new(big.Int).Add(&tip,buffer)
	maxFeePerGas := new(big.Int).Add(new(big.Int).Mul(latestBlock.BaseFee(),big.NewInt(2)),maxPriorityFeePerGas)
	gasPrice := GasPrice{
		maxPriorityFeePerGas: maxPriorityFeePerGas,
		maxFeePerGas: maxFeePerGas,
	}

	return gasPrice
	
}