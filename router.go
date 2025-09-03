package main

import (
	"bytes"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"

	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/wire"
	"github.com/gin-gonic/gin"
)

type Router struct {
	rawdb *RawDB
}

func NewRouter(rawdb *RawDB) *Router {
	return &Router{
		rawdb: rawdb,
	}
}

func (r *Router) GetUtxo(c *gin.Context) {
	address := c.PostForm("address")
	amount := c.PostForm("amount")
	count := c.PostForm("count")
	smallChange := c.PostForm("small_change")

	amountF, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	countF, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	if countF == 0 {
		countF = math.MaxInt64
	}

	if smallChange == "" {
		smallChange = "0"
	}
	smallChangeF, err := strconv.ParseInt(smallChange, 10, 64)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	allUtxo, amountA, err := r.rawdb.GetAllUtxo(address, amountF, countF, smallChangeF)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"utxo":   allUtxo,
		"amount": amountA,
	})
}

func (r *Router) GetBalance(c *gin.Context) {
	address := c.PostForm("address")

	balance, err := r.rawdb.GetBalance(address)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"balance": balance,
	})
}

func (r *Router) GetTxByAddress(c *gin.Context) {

	address := c.PostForm("address")
	limit := c.PostForm("limit")
	offset := c.PostForm("offset")

	limitF, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	offsetF, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	if limitF > 50 {
		limitF = 50
	}

	tx, state, err := r.rawdb.GetAddressTxs(address, limitF, offsetF)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
			"state": state,
		})
		return
	}

	c.JSON(200, gin.H{
		"tx":    tx,
		"state": state,
	})

}

func (r *Router) GetTx(c *gin.Context) {

	txhash := c.PostForm("txhash")
	hash, _ := chainhash.NewHashFromStr(txhash)
	transactionVerbose, err := r.rawdb.Node.GetRawTransactionVerboseBool(hash)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"tx": transactionVerbose,
	})

}

func (r *Router) Broadcast(c *gin.Context) {
	type params struct {
		TxHex string `json:"tx_hex"`
	}

	p := &params{}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	bytesData, err := hex.DecodeString(p.TxHex)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	msgTx := new(wire.MsgTx)
	err = msgTx.Deserialize(bytes.NewReader(bytesData))
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	txhash, err := r.rawdb.Node.SendRawTransaction(msgTx, true)
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	type HttpResult struct {
		Code  int         `json:"code"`
		Msg   string      `json:"msg"`
		Data  interface{} `json:"data"`
		Total int64       `json:"total"`
	}

	data := make(map[string]interface{})
	data["tx_hash"] = txhash.String()
	result := &HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = data

	c.JSON(http.StatusOK, result)
}

func (r *Router) GetCurrentBlock(c *gin.Context) {
	height, err := r.rawdb.GetHeight()
	if err != nil {
		c.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"current_block": height,
		"status":        "success",
	})
}


