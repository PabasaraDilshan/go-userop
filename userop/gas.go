package userop

import (
	"bytes"
	"context"
	"log"
	"math"
	"math/big"
	"userop-builder/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type GasEstimator struct{
    client *ethclient.Client
	ov *GasOverheads
}
type GasPrice struct{
	MaxFeePerGas *big.Int
	MaxPriorityFeePerGas *big.Int
}
type GasOverheads struct{

		fixed float64//: 21000,
		perUserOp float64//: 22874,
		perUserOpWord float64//: 25,
		batchFixed float64//:816,
		zeroByte float64 //: 4,
		nonZeroByte float64 //: 16,
		bundleSize float64//: 1,
		sigSize int//: 65
		paymasterDataSize int
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
	
	
	var res string
	err := ge.client.Client().CallContext(context.Background(),&res,"eth_maxPriorityFeePerGas")
	if err != nil {
		log.Fatal(err)
	}
	tip,_:= new(big.Int).SetString(res,0)
	blockNumber,err := ge.client.BlockNumber(context.Background())
	latestBlock, err := ge.client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
	if err != nil {
		log.Fatal(err)
	}
	buffer := new(big.Int).Div(tip,big.NewInt(2))
	maxPriorityFeePerGas := new(big.Int).Add(tip,buffer)
	maxFeePerGas := new(big.Int).Add(new(big.Int).Mul(latestBlock.BaseFee(),big.NewInt(2)),maxPriorityFeePerGas)
	gasPrice := GasPrice{
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		MaxFeePerGas: maxFeePerGas,
	}
	return gasPrice
}

func (ge *GasEstimator) CalcCallDataCost(userOp UserOperation) float64 {
	cost := float64(0)
	for _, b := range userOp.Pack() {
		if b == byte(0) {
			cost += ge.ov.zeroByte
		} else {
			cost += ge.ov.nonZeroByte
		}
	}

	return cost
}

func (ge *GasEstimator) CalcPerUserOpCost(op UserOperation) float64 {
	opLen := math.Floor(float64(len(op.Pack())+31) / 32)
	cost := (ge.ov.perUserOpWord * opLen) + ge.ov.perUserOp

	return cost
}

func (ge *GasEstimator) CalcPreverificationGas(userop UserOperation) *big.Int{
	uop := userop
	uop.PreVerificationGas = big.NewInt(100000)
	uop.VerificationGasLimit = big.NewInt(1000000)
	uop.CallGasLimit = big.NewInt(1000000)
	uop.Signature = bytes.Repeat([]byte{1}, ge.ov.sigSize)
	uop.PaymasterAndData = bytes.Repeat([]byte{1}, ge.ov.paymasterDataSize)

	// Calculate the additional gas for adding this userOp to a batch.
	batchOv := ((ge.ov.fixed + ge.ov.batchFixed) / ge.ov.bundleSize) + ge.CalcCallDataCost(uop)

	// The total PVG is the sum of the batch overhead and the overhead for this userOp's validation and
	// execution.
	pvg := batchOv + ge.CalcPerUserOpCost(userop)
	static := big.NewInt(int64(math.Round(pvg)))

	// Use value from CalcPreVerificationGasFunc if set, otherwise return the static value.
	return static
}