package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/doged/txscript"
	"github.com/dogecoinw/go-dogecoin/log"
)

var (
	delBlock      = int64(1000)
	startInterval = 3 * time.Second
)

type State struct {
	Node      *rpcclient.Client
	DB        *RawDB
	fromBlock int64

	ctx context.Context
	wg  *sync.WaitGroup
}

func NewState(ctx context.Context, wg *sync.WaitGroup, node *rpcclient.Client, db *RawDB) *State {
	return &State{
		Node: node,
		DB:   db,
		ctx:  ctx,
		wg:   wg,
	}
}

func (s *State) Start(fromBlock int64) {
	defer s.wg.Done()
	if fromBlock == 0 {
		height, err := s.DB.GetHeight()
		if err != nil {
			s.fromBlock = 0
		} else {
			s.fromBlock = height
		}
	} else {
		s.fromBlock = fromBlock
	}

	startTicker := time.NewTicker(startInterval)
out:
	for {
		select {
		case <-startTicker.C:
			if err := s.scan(); err != nil {
				log.Error("scanning", "scanning", err)
			}
		case <-s.ctx.Done():
			log.Info("scanning", "stop", "Done")
			break out
		}
	}
}

func (s *State) scan() error {
	blockCount, _ := s.Node.GetBlockCount()

	if blockCount-s.fromBlock > 100 {
		blockCount = s.fromBlock + 100
	}

	for ; s.fromBlock < blockCount; s.fromBlock++ {
		blockHash, err := s.Node.GetBlockHash(s.fromBlock)
		if err != nil {
			return err
		}

		log.Info("scanning", "fromBlock", s.fromBlock)
		block, err := s.Node.GetBlockVerboseBool(blockHash)
		if err != nil {
			return err
		}

		addrMap := make(map[string]float64, 0)
		for _, tx := range block.Tx {
			txhash, _ := chainhash.NewHashFromStr(tx)
			transactionVerbose, err := s.Node.GetRawTransactionVerboseBool(txhash)
			if err != nil {
				continue
			}

			vouts := make([]*Vout, 0)
			for _, vout := range transactionVerbose.Vout {

				hexb, _ := hex.DecodeString(vout.ScriptPubKey.Hex)
				_, addrs, _, err := txscript.ExtractPkScriptAddrs(hexb, &ChainCfg)
				if err != nil {
					return err
				}

				if len(addrs) == 0 {
					continue
				}
				voutDB := &Vout{
					Index:   vout.N,
					Value:   vout.Value,
					Address: addrs[0].EncodeAddress(),
				}

				vouts = append(vouts, voutDB)
				s.DB.SetVout(tx, vout.N, voutDB)

				vinDB := &Vin{
					Txid:    tx,
					Vout:    vout.N,
					Address: voutDB.Address,
					Value:   voutDB.Value,
				}

				s.DB.SetUtxo(voutDB.Address, tx, vout.N, vinDB)
				s.DB.SetAddressTx(voutDB.Address, tx, block.Height, block.Time)
				addrMap[addrs[0].EncodeAddress()] += vout.Value
			}

			vins := make([]*Vin, 0)
			for _, vin := range transactionVerbose.Vin {
				if vin.Coinbase != "" {
					continue
				}
				voutDB, _ := s.DB.GetVout(vin.Txid, vin.Vout)
				if voutDB == nil {
					fmt.Println("voutDB is nil", vin.Txid, vin.Vout)
					s.fork(vin.Txid)
					voutDB, _ = s.DB.GetVout(vin.Txid, vin.Vout)
					if voutDB == nil {
						fmt.Println("voutDB is still nil after fork, skipping", vin.Txid, vin.Vout)
						continue
					}
				}
				vinDB := &Vin{
					Txid:    vin.Txid,
					Vout:    vin.Vout,
					Address: voutDB.Address,
					Value:   voutDB.Value,
				}
				vins = append(vins, vinDB)
				addrMap[voutDB.Address] -= voutDB.Value
				s.DB.DelUtxo(voutDB.Address, vinDB.Txid, vinDB.Vout)
				s.DB.SetAddressTx(voutDB.Address, tx, block.Height, block.Time)
			}

			txDB := &Tx{
				Txid:  tx,
				Vins:  vins,
				Vouts: vouts,
			}
			s.DB.SetTx(txDB)
		}

		for addr, value := range addrMap {
			s.updateBalance(addr, value)
		}
		s.DB.SetHeight(s.fromBlock)
	}
	return nil
}

// 更新余额
func (s *State) updateBalance(address string, value float64) error {
	balance, _ := s.DB.GetBalance(address)
	balance += value
	return s.DB.SetBalance(address, balance)
}

func (s *State) fork(hash string) error {
	return s.forkWithDepth(hash, 0)
}

func (s *State) forkWithDepth(hash string, depth int) error {
	// 防止无限递归，最大递归深度为10
	if depth > 10 {
		fmt.Printf("fork recursion depth exceeded for hash: %s\n", hash)
		return nil
	}

	txhash, _ := chainhash.NewHashFromStr(hash)
	transactionVerbose, err := s.Node.GetRawTransactionVerboseBool(txhash)
	if err != nil {
		return err
	}

	vouts := make([]*Vout, 0)
	addrMap := make(map[string]float64, 0)
	for _, vout := range transactionVerbose.Vout {

		hexb, _ := hex.DecodeString(vout.ScriptPubKey.Hex)
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(hexb, &ChainCfg)
		if err != nil {
			return err
		}

		if len(addrs) == 0 {
			continue
		}

		voutDB := &Vout{
			Index:   vout.N,
			Value:   vout.Value,
			Address: addrs[0].EncodeAddress(),
		}
		vouts = append(vouts, voutDB)
		s.DB.SetVout(hash, vout.N, voutDB)

		vinDB := &Vin{
			Txid:    hash,
			Vout:    vout.N,
			Address: voutDB.Address,
			Value:   voutDB.Value,
		}

		s.DB.SetUtxo(voutDB.Address, hash, vout.N, vinDB)
		addrMap[addrs[0].EncodeAddress()] += vout.Value
	}

	vins := make([]*Vin, 0)
	for _, vin := range transactionVerbose.Vin {
		if vin.Coinbase != "" {
			continue
		}
		voutDB, _ := s.DB.GetVout(vin.Txid, vin.Vout)
		if voutDB == nil {
			fmt.Println("voutDB is nil", vin.Txid, vin.Vout)
			s.forkWithDepth(vin.Txid, depth+1)
			voutDB, _ = s.DB.GetVout(vin.Txid, vin.Vout)
			if voutDB == nil {
				fmt.Println("voutDB is still nil after fork in fork method, skipping", vin.Txid, vin.Vout)
				continue
			}
		}
		vinDB := &Vin{
			Txid:    vin.Txid,
			Vout:    vin.Vout,
			Address: voutDB.Address,
			Value:   voutDB.Value,
		}
		vins = append(vins, vinDB)
		addrMap[voutDB.Address] -= voutDB.Value
		s.DB.DelUtxo(voutDB.Address, vinDB.Txid, vinDB.Vout)
	}

	for addr, value := range addrMap {
		s.updateBalance(addr, value)
	}
	return nil

}
