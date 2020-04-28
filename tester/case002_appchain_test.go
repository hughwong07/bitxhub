package tester

import (
	"strconv"
	"testing"

	"github.com/meshplus/bitxhub-kit/types"
	"github.com/meshplus/bitxhub/internal/constant"

	"github.com/meshplus/bitxhub/internal/coreapi/api"

	"github.com/meshplus/bitxhub-kit/crypto"
	"github.com/meshplus/bitxhub-kit/crypto/asym/ecdsa"
	"github.com/meshplus/bitxhub-model/pb"
	rpcx "github.com/meshplus/go-bitxhub-client"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type RegisterAppchain struct {
	suite.Suite
	api     api.CoreAPI
	privKey crypto.PrivateKey
	from    types.Address
}

func (suite *RegisterAppchain) SetupSuite() {
	var err error
	suite.privKey, err = ecdsa.GenerateKey(ecdsa.Secp256r1)
	suite.Require().Nil(err)

	suite.from, err = suite.privKey.PublicKey().Address()
	suite.Require().Nil(err)
}

// Appchain registers in bitxhub
func (suite *RegisterAppchain) TestRegisterAppchain() {
	args := []*pb.Arg{
		rpcx.String(""),
		rpcx.Int32(0),
		rpcx.String("hyperchain"),
		rpcx.String("税务链"),
		rpcx.String("趣链税务链"),
		rpcx.String("1.8"),
	}

	tx, err := genBVMContractTransaction(suite.privKey, constant.InterchainContractAddr.Address(), "Register", args...)
	suite.Nil(err)

	ret, err := sendTransactionWithReceipt(suite.api, tx)
	suite.Require().Nil(err)
	suite.Require().Equal("hyperchain", gjson.Get(string(ret.Ret), "chain_type").String())
}

func (suite *RegisterAppchain) TestFetchAppchains() {
	k1, err := ecdsa.GenerateKey(ecdsa.Secp256r1)
	suite.Require().Nil(err)
	k2, err := ecdsa.GenerateKey(ecdsa.Secp256r1)
	suite.Require().Nil(err)

	args := []*pb.Arg{
		rpcx.String(""),
		rpcx.Int32(0),
		rpcx.String("hyperchain"),
		rpcx.String("税务链"),
		rpcx.String("趣链税务链"),
		rpcx.String("1.8"),
	}
	tx, err := genBVMContractTransaction(k1, constant.InterchainContractAddr.Address(), "Register", args...)
	suite.Require().Nil(err)
	ret, err := sendTransactionWithReceipt(suite.api, tx)
	suite.Require().Nil(err)
	suite.Require().True(ret.IsSuccess())

	args = []*pb.Arg{
		rpcx.String(""),
		rpcx.Int32(0),
		rpcx.String("fabric"),
		rpcx.String("政务链"),
		rpcx.String("fabric政务"),
		rpcx.String("1.4"),
	}

	tx, err = genBVMContractTransaction(k2, constant.InterchainContractAddr.Address(), "Register", args...)
	suite.Require().Nil(err)
	ret, err = sendTransactionWithReceipt(suite.api, tx)
	suite.Require().Nil(err)

	tx, err = genBVMContractTransaction(k2, constant.InterchainContractAddr.Address(), "Appchains")
	suite.Require().Nil(err)
	ret, err = sendTransactionWithReceipt(suite.api, tx)
	suite.Require().Nil(err)
	suite.Require().True(ret.IsSuccess())

	tx, err = genBVMContractTransaction(k2, constant.InterchainContractAddr.Address(), "CountAppchains")
	suite.Require().Nil(err)
	rec, err := sendTransactionWithReceipt(suite.api, tx)
	suite.Require().Nil(err)
	suite.Require().True(ret.IsSuccess())
	num, err := strconv.Atoi(string(rec.Ret))
	suite.Require().Nil(err)
	result := gjson.Parse(string(ret.Ret))
	suite.Require().GreaterOrEqual(num, len(result.Array()))

	tx, err = genBVMContractTransaction(k2, constant.InterchainContractAddr.Address(), "CountApprovedAppchains")
	suite.Require().Nil(err)
	ret, err = sendTransactionWithReceipt(suite.api, tx)
	suite.Require().Nil(err)
	suite.Require().True(ret.IsSuccess())
	num, err = strconv.Atoi(string(ret.Ret))
	suite.Require().Nil(err)
	suite.Require().EqualValues(0, num)
}

func TestRegisterAppchain(t *testing.T) {
	suite.Run(t, &RegisterAppchain{})
}
