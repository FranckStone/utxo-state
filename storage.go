package main

import (
	"strconv"
	"strings"

	"github.com/dogecoinw/doged/rpcclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	blockPrefix     = "block-"
	voutPrefix      = "vout-"
	balancePrefix   = "balance-"
	utxoPrefix      = "utxo-"
	txPrefix        = "tx-"
	txAddressPrefix = "tx-address-"
)

type RawDB struct {
	DB   *leveldb.DB
	Node *rpcclient.Client
}

func (d *RawDB) Stop() error {
	return d.DB.Close()
}

// 保存当前高度
func (d *RawDB) SetHeight(height int64) error {
	if err := d.DB.Put([]byte("height"), []byte(strconv.FormatInt(height, 10)), nil); err != nil {
		return err
	}
	return nil
}

// 获取当前高度
func (d *RawDB) GetHeight() (int64, error) {
	var height int64
	if data, err := d.DB.Get([]byte("height"), nil); err != nil {
		return height, err
	} else {
		height, _ = strconv.ParseInt(string(data), 10, 64)
	}
	return height, nil
}

// 保存区块信息
func (d *RawDB) SetBlock(height int64, block *Block) error {
	if data, err := rlp.EncodeToBytes(block); err != nil {
		return err
	} else {
		if err := d.DB.Put(blockKey(height), data, nil); err != nil {
			return err
		}
	}
	return nil
}

// 获取区块信息
func (d *RawDB) GetBlock(height int64) (*Block, error) {
	var block *Block
	if data, err := d.DB.Get(blockKey(height), nil); err != nil {
		return block, err
	} else {
		if err := rlp.DecodeBytes(data, &block); err != nil {
			return block, err
		}
	}
	return block, nil
}

// 删除区块信息
func (d *RawDB) DelBlock(height int64) error {
	return d.DB.Delete(blockKey(height), nil)
}

// 保存vout信息
func (d *RawDB) SetVout(txid string, index uint32, vout *Vout) error {
	if data, err := rlp.EncodeToBytes(vout); err != nil {
		return err
	} else {
		if err := d.DB.Put(voutKey(txid, index), data, nil); err != nil {
			return err
		}
	}
	return nil
}

// 获取vout信息
func (d *RawDB) GetVout(txid string, index uint32) (*Vout, error) {
	var vout *Vout
	if data, err := d.DB.Get(voutKey(txid, index), nil); err != nil {
		return vout, err
	} else {
		if err := rlp.DecodeBytes(data, &vout); err != nil {
			return vout, err
		}
	}
	return vout, nil
}

// 保存地址余额
func (d *RawDB) SetBalance(address string, balance float64) error {
	if err := d.DB.Put(balanceKey(address), []byte(strconv.FormatFloat(balance, 'f', 8, 64)), nil); err != nil {
		return err
	}
	return nil
}

// 获取地址余额
func (d *RawDB) GetBalance(address string) (float64, error) {
	if data, err := d.DB.Get(balanceKey(address), nil); err != nil {
		return 0, err
	} else {
		if balance, err := strconv.ParseFloat(string(data), 64); err != nil {
			return 0, err
		} else {
			return balance, nil
		}
	}
}

// 保存utxo
func (d *RawDB) SetUtxo(address string, txid string, index uint32, vin *Vin) error {
	if data, err := rlp.EncodeToBytes(vin); err != nil {
		return err
	} else {
		if err := d.DB.Put(utxoKey(address, txid, index), data, nil); err != nil {
			return err
		}
	}
	return nil
}

// 通过Iterator 获取所有utxo
func (d *RawDB) GetAllUtxo(address string, amount float64, count, smallChangeF int64) ([]*Vin, float64, error) {
	var vins []*Vin
	startKey := []byte(utxoPrefix + address)
	iter := d.DB.NewIterator(util.BytesPrefix(startKey), nil)
	temp := float64(0)
	temp1 := int64(0)
	for iter.Next() {
		var vin *Vin
		if err := rlp.DecodeBytes(iter.Value(), &vin); err != nil {
			return vins, 0, err
		}

		if smallChangeF == 1 && vin.Value == 0.001 {
			continue
		}

		vins = append(vins, vin)

		temp += vin.Value
		if temp >= amount {
			break
		}
		temp1++
		if temp1 > count {
			break
		}
	}
	iter.Release()
	return vins, temp, nil
}

// 删除utxo
func (d *RawDB) DelUtxo(address string, txid string, index uint32) error {
	if err := d.DB.Delete(utxoKey(address, txid, index), nil); err != nil {
		return err
	}
	return nil
}

// 保存交易信息, 根据地址
func (d *RawDB) SetAddressTx(address, txid string, height, time int64) error {

	newkey := []byte("tx-address-" + address + "-" + strconv.FormatInt(height, 10) + "-" + strconv.FormatInt(time, 10) + "-" + txid)
	if err := d.DB.Put(newkey, []byte{0}, nil); err != nil {
		return err
	}
	return nil
}

// 获取
func (d *RawDB) GetAddressTxs(address string, limit, offset int64) ([]*Tx, int8, error) {
	
	var txs []*Tx
	startKey := []byte(txAddressPrefix + address)

	iter := d.DB.NewIterator(util.BytesPrefix(startKey), nil)

	temp := int64(0)
	temp1 := int64(0)
	for iter.Last(); iter.Valid(); iter.Prev() {

		// 处理offset
		if temp < offset {
			temp++
			continue
		}

		txhash := string(iter.Key())

		// 按照- 分割txhash
		hash := strings.Split(txhash, "-")
		if len(hash) != 6 {
			continue
		}

		tx, _ := d.GetTx(hash[5])
		if tx == nil {
			continue
		}

		tx.Height, _ = strconv.ParseInt(hash[3], 10, 64)
		tx.Time, _ = strconv.ParseInt(hash[4], 10, 64)

		txs = append(txs, tx)

		temp1++
		if temp1 > limit {
			break
		}
	}
	iter.Release()
	return txs, 2, nil
}

// 保存交易信息
func (d *RawDB) SetTx(tx *Tx) error {
	if data, err := rlp.EncodeToBytes(tx); err != nil {
		return err
	} else {
		if err := d.DB.Put(txKey(tx.Txid), data, nil); err != nil {
			return err
		}
	}
	return nil
}

// 获取交易信息
func (d *RawDB) GetTx(txid string) (*Tx, error) {
	var tx *Tx
	if data, err := d.DB.Get(txKey(txid), nil); err != nil {
		return tx, err
	} else {
		if err := rlp.DecodeBytes(data, &tx); err != nil {
			return tx, err
		}
	}
	return tx, nil
}

// txreload
func (d *RawDB) SetTxReload(address string, state uint8) error {
	if err := d.DB.Put(txReloadKey(address), []byte{state}, nil); err != nil {
		return err
	}
	return nil
}

// 获取txreload
func (d *RawDB) GetTxReload(address string) (uint8, error) {
	if data, err := d.DB.Get(txReloadKey(address), nil); err != nil {
		return 0, nil
	} else {
		return data[0], nil
	}
}

// 创建key
func blockKey(height int64) []byte {
	return []byte(blockPrefix + strconv.FormatInt(height, 10))
}

func voutKey(txid string, index uint32) []byte {
	return []byte(voutPrefix + txid + "-" + strconv.FormatUint(uint64(index), 10))
}

func balanceKey(address string) []byte {
	return []byte(balancePrefix + address)
}

func utxoKey(address string, txid string, index uint32) []byte {
	return []byte(utxoPrefix + address + "-" + txid + "-" + strconv.FormatUint(uint64(index), 10))
}

func txKey(txid string) []byte {
	return []byte(txPrefix + txid)
}

func txAddressKey(address, txid string) []byte {
	return []byte(txAddressPrefix + address + "-" + txid)
}

// txreloadKey
func txReloadKey(address string) []byte {
	return []byte("-reload" + address)
}
