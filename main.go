package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/gin-gonic/gin"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	cfg      Config
	ChainCfg chaincfg.Params
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	LoadConfig(&cfg, "")

	// 将配置中的数组转换为字节数组
	hdPublicKeyID := [4]byte{
		byte(cfg.ChainConfig.HDPublicKeyID[0]),
		byte(cfg.ChainConfig.HDPublicKeyID[1]),
		byte(cfg.ChainConfig.HDPublicKeyID[2]),
		byte(cfg.ChainConfig.HDPublicKeyID[3]),
	}
	hdPrivateKeyID := [4]byte{
		byte(cfg.ChainConfig.HDPrivateKeyID[0]),
		byte(cfg.ChainConfig.HDPrivateKeyID[1]),
		byte(cfg.ChainConfig.HDPrivateKeyID[2]),
		byte(cfg.ChainConfig.HDPrivateKeyID[3]),
	}

	ChainCfg = chaincfg.Params{
		PubKeyHashAddrID:        byte(cfg.ChainConfig.PubKeyHashAddrID),
		ScriptHashAddrID:        byte(cfg.ChainConfig.ScriptHashAddrID),
		PrivateKeyID:            byte(cfg.ChainConfig.PrivateKeyID),
		WitnessPubKeyHashAddrID: byte(cfg.ChainConfig.WitnessPubKeyHashAddrID),
		WitnessScriptHashAddrID: byte(cfg.ChainConfig.WitnessScriptHashAddrID),
		HDPublicKeyID:           hdPublicKeyID,
		HDPrivateKeyID:          hdPrivateKeyID,
		HDCoinType:              uint32(cfg.ChainConfig.HDCoinType),
	}

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(true)))
	glogger.Verbosity(log.Lvl(3))
	log.Root().SetHandler(glogger)

	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.Chain.RPC,
		Endpoint:     "ws",
		User:         cfg.Chain.UserName,
		Pass:         cfg.Chain.PassWord,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	rpcClient, _ := rpcclient.New(connCfg, nil)
	db, err := leveldb.OpenFile(cfg.DbPath, nil)
	if err != nil {
		panic(fmt.Sprintf("Leveldb err %s", err))
	}
	defer db.Close()
	RawDB := &RawDB{DB: db, Node: rpcClient}
	state := NewState(ctx, wg, rpcClient, RawDB)
	wg.Add(1)

	go state.Start(cfg.FromBlock)

	newRouter := NewRouter(RawDB)

	// 创建一个新的 Gin 路由器实例
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")                                               // 允许所有来源访问
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")                // 允许的请求方法
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") // 允许的请求头部信息
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")                                                 // 预检请求的有效期，单位为秒

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	router.POST("/utxo", newRouter.GetUtxo)                  // 获取指定地址的UTXO列表，支持按金额和数量筛选
	router.POST("/getBalance", newRouter.GetBalance)         // 获取指定地址的余额
	router.POST("/getTxByAddress", newRouter.GetTxByAddress) // 根据地址获取交易历史记录，支持分页
	router.POST("/getTx", newRouter.GetTx)                   // 根据交易哈希获取交易详细信息
	router.POST("/broadcast", newRouter.Broadcast)           // 广播已签名的交易到区块链网络

	// 启动 HTTP 服务器并监听端口
	go func() {
		err = router.Run(cfg.Server)
		if err != nil {
			panic(err)
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived an interrupt, stopping services...")
		cancel() // 取消 context，这将取消所有的 worker
	}()

	wg.Wait()
}
