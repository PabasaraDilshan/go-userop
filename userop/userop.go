package userop

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	address, _ = abi.NewType("address", "", nil)
	uint256, _ = abi.NewType("uint256", "", nil)
	bytes32, _ = abi.NewType("bytes32", "", nil)

	// UserOpPrimitives is the primitive ABI types for each UserOperation field.
	UserOpPrimitives = []abi.ArgumentMarshaling{
		{Name: "sender", InternalType: "Sender", Type: "address"},
		{Name: "nonce", InternalType: "Nonce", Type: "uint256"},
		{Name: "initCode", InternalType: "InitCode", Type: "bytes"},
		{Name: "callData", InternalType: "CallData", Type: "bytes"},
		{Name: "callGasLimit", InternalType: "CallGasLimit", Type: "uint256"},
		{Name: "verificationGasLimit", InternalType: "VerificationGasLimit", Type: "uint256"},
		{Name: "preVerificationGas", InternalType: "PreVerificationGas", Type: "uint256"},
		{Name: "maxFeePerGas", InternalType: "MaxFeePerGas", Type: "uint256"},
		{Name: "maxPriorityFeePerGas", InternalType: "MaxPriorityFeePerGas", Type: "uint256"},
		{Name: "paymasterAndData", InternalType: "PaymasterAndData", Type: "bytes"},
		{Name: "signature", InternalType: "Signature", Type: "bytes"},
	}

	// UserOpType is the ABI type of a UserOperation.
	UserOpType, _ = abi.NewType("tuple", "op", UserOpPrimitives)

	// UserOpArr is the ABI type for an array of UserOperations.
	UserOpArr, _ = abi.NewType("tuple[]", "ops", UserOpPrimitives)
)


type UserOperation struct {
	Sender               common.Address `json:"sender"               mapstructure:"sender"               validate:"required"`
	Nonce                *big.Int       `json:"nonce"                mapstructure:"nonce"                validate:"required"`
	InitCode             []byte         `json:"initCode"             mapstructure:"initCode"             validate:"required"`
	CallData             []byte         `json:"callData"             mapstructure:"callData"             validate:"required"`
	CallGasLimit         *big.Int       `json:"callGasLimit"         mapstructure:"callGasLimit"         validate:"required"`
	VerificationGasLimit *big.Int       `json:"verificationGasLimit" mapstructure:"verificationGasLimit" validate:"required"`
	PreVerificationGas   *big.Int       `json:"preVerificationGas"   mapstructure:"preVerificationGas"   validate:"required"`
	MaxFeePerGas         *big.Int       `json:"maxFeePerGas"         mapstructure:"maxFeePerGas"         validate:"required"`
	MaxPriorityFeePerGas *big.Int       `json:"maxPriorityFeePerGas" mapstructure:"maxPriorityFeePerGas" validate:"required"`
	PaymasterAndData     []byte         `json:"paymasterAndData"     mapstructure:"paymasterAndData"     validate:"required"`
	Signature            []byte         `json:"signature"            mapstructure:"signature"            validate:"required"`
}


func (op *UserOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Sender               string `json:"sender"`
		Nonce                string `json:"nonce"`
		InitCode             string `json:"initCode"`
		CallData             string `json:"callData"`
		CallGasLimit         string `json:"callGasLimit"`
		VerificationGasLimit string `json:"verificationGasLimit"`
		PreVerificationGas   string `json:"preVerificationGas"`
		MaxFeePerGas         string `json:"maxFeePerGas"`
		MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
		PaymasterAndData     string `json:"paymasterAndData"`
		Signature            string `json:"signature"`
	}{
		Sender:               op.Sender.String(),
		Nonce:                hexutil.EncodeBig(op.Nonce),
		InitCode:             hexutil.Encode(op.InitCode),
		CallData:             hexutil.Encode(op.CallData),
		CallGasLimit:         hexutil.EncodeBig(op.CallGasLimit),
		VerificationGasLimit: hexutil.EncodeBig(op.VerificationGasLimit),
		PreVerificationGas:   hexutil.EncodeBig(op.PreVerificationGas),
		MaxFeePerGas:         hexutil.EncodeBig(op.MaxFeePerGas),
		MaxPriorityFeePerGas: hexutil.EncodeBig(op.MaxPriorityFeePerGas),
		PaymasterAndData:     hexutil.Encode(op.PaymasterAndData),
		Signature:            hexutil.Encode(op.Signature),
	})
}

// ToMap returns the current UserOp struct as a map type.
func (op *UserOperation) ToMap() (map[string]any, error) {
	data, err := op.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var opData map[string]any
	if err := json.Unmarshal(data, &opData); err != nil {
		return nil, err
	}
	return opData, nil
}

func (op *UserOperation) Pack() []byte {
	args := abi.Arguments{
		{Name: "UserOp", Type: UserOpType},
	}
	packed, _ := args.Pack(&struct {
		Sender               common.Address
		Nonce                *big.Int
		InitCode             []byte
		CallData             []byte
		CallGasLimit         *big.Int
		VerificationGasLimit *big.Int
		PreVerificationGas   *big.Int
		MaxFeePerGas         *big.Int
		MaxPriorityFeePerGas *big.Int
		PaymasterAndData     []byte
		Signature            []byte
	}{
		op.Sender,
		op.Nonce,
		op.InitCode,
		op.CallData,
		op.CallGasLimit,
		op.VerificationGasLimit,
		op.PreVerificationGas,
		op.MaxFeePerGas,
		op.MaxPriorityFeePerGas,
		op.PaymasterAndData,
		op.Signature,
	})

	enc := hexutil.Encode(packed)
	enc = "0x" + enc[66:]
	return (hexutil.MustDecode(enc))
}

