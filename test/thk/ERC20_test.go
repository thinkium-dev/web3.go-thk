package test

import (
	"encoding/json"
	"fmt"
	"github.com/Alex-Chris/log/log"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"
	"wallet/models"
	"web3.go/common/cryp/crypto"
	"web3.go/common/hexutil"
	"web3.go/web3"
	"web3.go/web3/providers"
	"web3.go/web3/thk/util"
)

var ERC20JsonName = "../resources/ERC20.json"
var TokenVestingJsonName = "../resources/TokenVesting.json"

func TestDeployERC20(t *testing.T) {
	ctct, err := CompileContract("../resources/contract/ERC20.sol",
		"../resources/contract/IERC20.sol", "../resources/contract/Pausable.sol",
		"../resources/contract/SafeMath.sol", "../resources/contract/Ownable.sol",
		"../resources/contract/Address.sol", "../resources/contract/SafeERC20.sol",
		"../resources/contract/TokenVesting.sol")
	var ERC20Json, TokenVestingJson ContractJson
	amount := new(big.Int).SetUint64(uint64(100000))
	decimal := uint8(8)
	ERC20Json, TokenVestingJson, err = GetJson(ctct)
	fmt.Println("contractJson的值\n", ctct)
	dataERC20, err := json.MarshalIndent(ERC20Json, "", "  ")
	if ioutil.WriteFile(ERC20JsonName, dataERC20, 0644) == nil {
		fmt.Println("写入文件成功")
	}
	contentERC20, err := ioutil.ReadFile(ERC20JsonName)
	if err != nil {
		log.Error(err)

	}

	dataTokenVesting, err := json.MarshalIndent(TokenVestingJson, "", "  ")
	if ioutil.WriteFile(TokenVestingJsonName, dataTokenVesting, 0644) == nil {
		fmt.Println("写入文件成功")
	}
	contentTokenVesting, err := ioutil.ReadFile(TokenVestingJsonName)
	if err != nil {
		log.Error(err)

	}

	var ERC20Response TruffleContract
	_ = json.Unmarshal(contentERC20, &ERC20Response)

	var TokenVestingResponse TruffleContract
	_ = json.Unmarshal(contentTokenVesting, &TokenVestingResponse)

	var connection = web3.NewWeb3(providers.NewHTTPProvider("192.168.1.13:8089", 10, false))

	bytecodeERC20 := ERC20Response.Bytecode
	contractERC20, err := connection.Thk.NewContract(ERC20Response.Abi)
	bytecodeTokenVesting := TokenVestingResponse.Bytecode
	contractTokenVesting, err := connection.Thk.NewContract(TokenVestingResponse.Abi)
	from := "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23"
	nonce, err := connection.Thk.GetNonce(from, Chain)
	if err != nil {
		log.Error(err)
		println("get nonce error")
		return
	}
	transaction := util.Transaction{
		ChainId: Chain, FromChainId: Chain, ToChainId: Chain, From: from,
		To: "", Value: "0", Input: "", Nonce: strconv.Itoa(int(nonce)),
	}

	privateKey, err := crypto.HexToECDSA(key)
	hashERC20, err := contractERC20.Deploy(transaction, bytecodeERC20, privateKey, Symbol, Name, decimal, amount)
	if err != nil {
		log.Error(err)
		fmt.Println("get hash error")
		return
	}
	log.Info("contractERC20 hash:", hashERC20)

	time.Sleep(time.Second * 10)
	receiptERC20, err := connection.Thk.GetTransactionByHash(Chain, hashERC20)
	if err != nil {
		log.Error(err)
		fmt.Println("get hash error")
		return
	}
	log.Info("contract addr:", receiptERC20.ContractAddress)
	toERC20 := receiptERC20.ContractAddress
	newtransaction := util.Transaction{
		ChainId: Chain, FromChainId: Chain, ToChainId: Chain, From: from,
		To: toERC20, Value: "0", Input: "", Nonce: strconv.Itoa(int(nonce)),
	}
	result, err := contractERC20.Call(newtransaction, "symbol")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log("result:", result)
	var str string
	err = contractERC20.Parse(result, "symbol", &str)
	if err != nil {
		log.Error(err)
		println("failed")
		return
	}
	//TokenVesting
	nonce, err = connection.Thk.GetNonce(from, Chain)
	if err != nil {
		log.Error(err)
		println("get nonce error")
		return
	}
	deployTokenVesting := util.Transaction{
		ChainId: Chain, FromChainId: Chain, ToChainId: Chain, From: from,
		To: "", Value: "0", Input: "", Nonce: strconv.Itoa(int(nonce)),
	}
	toERC20Bytes, err := hexutil.Decode(toERC20)
	addressERC20 := models.BytesToAddress(toERC20Bytes)
	hashTokenVesting, err := contractTokenVesting.Deploy(deployTokenVesting, bytecodeTokenVesting, privateKey, addressERC20)
	if err != nil {
		log.Error(err)
		fmt.Println("get contractTokenVesting hash error")
		return
	}
	log.Info("contractTokenVesting hash:", hashTokenVesting)
	time.Sleep(time.Second * 10)
	receiptTokenVesting, err := connection.Thk.GetTransactionByHash(Chain, hashTokenVesting)
	if err != nil {
		log.Error(err)
		fmt.Println("get hash error")
		return
	}
	log.Info("contractTokenVesting addr:", receiptTokenVesting.ContractAddress)
	toTokenVesting := receiptTokenVesting.ContractAddress

	//Approve
	addrefrom, err := hexutil.Decode(from)
	addressfrom := models.BytesToAddress(addrefrom)
	value := new(big.Int).SetUint64(uint64(100000))
	input, err := contractERC20.GetInput("approve", addressfrom, value)
	println(input)

	nonce, err = connection.Thk.GetNonce(from, Chain)
	if err != nil {
		log.Error(err)
		println("get nonce error")
		return
	}
	transactionApprove := util.Transaction{
		ChainId: Chain, FromChainId: Chain, ToChainId: Chain, From: from,
		To: toERC20, Value: "0", Input: input,
		Nonce: strconv.Itoa(int(nonce)),
	}
	privatekey, err := crypto.HexToECDSA(key)
	err = connection.Thk.SignTransaction(&transactionApprove, privatekey)

	txhash, err := connection.Thk.SendTx(&transactionApprove)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	println(txhash)
	t.Log("Approve txhash:", txhash)


	time.Sleep(time.Second * 10)

	//Transfer
	addr := strings.ToLower(toTokenVesting)
	addreto, err := hexutil.Decode(addr)
	addressto := models.BytesToAddress(addreto)
	value = new(big.Int).SetUint64(uint64(100000))
	input, err = contractERC20.GetInput("transferFrom", addressfrom, addressto, value)
	//input, err := contract.GetInput("transfer",  addressto, value)
	println(input)
	nonce, err = connection.Thk.GetNonce(from, Chain)
	if err != nil {
		log.Error(err)
		println("get nonce error")
		return
	}
	transactionTransfer := util.Transaction{
		ChainId: Chain, FromChainId: Chain, ToChainId: Chain, From: from,
		To: toERC20, Value: "0", Input: input,
		Nonce: strconv.Itoa(int(nonce)),
	}
	privatekey, err = crypto.HexToECDSA(key)
	err = connection.Thk.SignTransaction(&transactionTransfer, privatekey)

	txhash, err = connection.Thk.SendTx(&transactionTransfer)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log("Transfer txhash:", txhash)


	time.Sleep(time.Second * 10)
	//Vesting
	nonce, err = connection.Thk.GetNonce(from, Chain)
	if err != nil {
		log.Error(err)
		println("get nonce error")
		return
	}
	transactionTokenVesting := util.Transaction{
		ChainId: Chain, FromChainId: Chain, ToChainId: Chain, From: from,
		To: toTokenVesting, Value: "0", Input: "", Nonce: strconv.Itoa(int(nonce)),
	}
	tmcliff, errc := time.Parse("2006-01-02 15:04:05", "2019-07-19 17:47:00")
	tmstart, errc:= time.Parse("2006-01-02 15:04:05", "2019-07-19 17:48:00")
	tmend, errc := time.Parse("2006-01-02 15:04:05", "2019-07-19 17:49:00")
	println(errc)
	toAddress, err := hexutil.Decode("0x14723a09acff6d2a60dcdf7aa4aff308fddc160c")
	vestAddress := models.BytesToAddress(toAddress)
	//vestAddressP,boolerr := new(big.Int).SetString(strings.TrimPrefix(string("0x14723a09acff6d2a60dcdf7aa4aff308fddc160c"), "0x"),16)
	//println(boolerr)
	tmcliffP:= new(big.Int).SetInt64(tmcliff.Unix())
	tmstartP:= new(big.Int).SetInt64(tmstart.Unix())
	tmendP:= new(big.Int).SetInt64(tmend.Unix())
	timesP:= new(big.Int).SetInt64(2)
	total := new(big.Int).SetUint64(uint64(100000))
	resultTokenVesting, err := contractTokenVesting.Send(transactionTokenVesting, "addPlan",privatekey,
		vestAddress, tmcliffP, tmstartP, timesP, tmendP, total, false, "上交易所后私募锁仓10分钟，之后每10分钟释放50%")
	if err != nil {
		t.Error(err)
	}
	t.Log("result:", resultTokenVesting)


	var token Token
	token.Name = Name
	token.Symbol = Symbol
	token.Total = 100000
	token.ContractAddress = receiptERC20.ContractAddress
	token.ABI = ERC20Response.Abi
	token.Icon = "icon"
	token.Website = "www.thinkey.com"
	token.Introduction = Des
	token.Date = time.Now().Format("2006-01-02")
	token.ChainId = "2"
	token.Decimal = 8
	PostInfo(token)
	//PostFile(token, "../resources/abc.sol")
}

func GetJson(ctct map[string]interface{}) (ContractJson, ContractJson, error) {
	var contractJson, ERC20Json, TokenVestingJson ContractJson
	for keyname, value := range ctct {
		contractJson.ContractName = keyname
		arr := strings.Split(contractJson.ContractName, ":")
		length := len(arr) - 1
		if (arr[length] == "ERC20") {
			mapvalue := value.(map[string]interface{})
			ERC20Json.ByteCode = mapvalue["code"].(string)
			info := mapvalue["info"].(map[string]interface{})
			abidef := info["abiDefinition"]
			abibytes, _ := json.Marshal(abidef)
			ERC20Json.ABI = string(abibytes)
		}
		if (arr[length] == "TokenVesting") {
			mapvalue := value.(map[string]interface{})
			TokenVestingJson.ByteCode = mapvalue["code"].(string)
			info := mapvalue["info"].(map[string]interface{})
			abidef := info["abiDefinition"]
			abibytes, _ := json.Marshal(abidef)
			TokenVestingJson.ABI = string(abibytes)
		}
	}
	return ERC20Json, TokenVestingJson, nil
}
