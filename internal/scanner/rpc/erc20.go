// Package rpc implements Opera node communication wrappers through an adapter.
package rpc

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Erc20Name collects name of the given ERC20 contract if possible.
// Solidity: function name() view returns(string)
func (a *Adapter) Erc20Name(adr common.Address) (string, error) {
	// call ERC20
	data, err := a.ftm.CallContract(context.Background(), ethereum.CallMsg{
		From: common.Address{},
		To:   &adr,
		Data: common.Hex2Bytes("06fdde03"),
	}, nil)

	if err != nil {
		fmt.Println("can not get ERC20 name", err.Error())
		return "", err
	}

	return decodeAbiString(data), nil
}

// Erc20Symbol collects symbol of the given ERC20 contract if possible.
// Solidity: function symbol() view returns(string)
func (a *Adapter) Erc20Symbol(adr common.Address) (string, error) {
	// call ERC20
	data, err := a.ftm.CallContract(context.Background(), ethereum.CallMsg{
		From: common.Address{},
		To:   &adr,
		Data: common.Hex2Bytes("95d89b41"),
	}, nil)

	if err != nil {
		fmt.Println("can not get ERC20 symbol", err.Error())
		return "", err
	}

	return decodeAbiString(data), nil
}

// Erc20Decimals collects number of decimals of the given ERC20 contract.
// Solidity: function decimals() view returns(uint8)
func (a *Adapter) Erc20Decimals(adr common.Address) (uint8, error) {
	// call ERC20
	data, err := a.ftm.CallContract(context.Background(), ethereum.CallMsg{
		From: common.Address{},
		To:   &adr,
		Data: common.Hex2Bytes("313ce567"),
	}, nil)

	if err != nil {
		fmt.Println("can not get ERC20 decimals", err.Error())
		return 0, err
	}

	// even uint8 is encoded in 32 bytes by ABI
	return data[31], nil
}

// decodeAbiString decodes string from ABI format.
func decodeAbiString(data []byte) string {
	// does it even make sense?
	if len(data) < 64 {
		return ""
	}

	// where the string starts and ends?
	offset := new(big.Int).SetBytes(data[:32]).Uint64() + 32
	length := new(big.Int).SetBytes(data[offset-32 : offset]).Uint64()

	return string(data[offset : offset+length])
}
