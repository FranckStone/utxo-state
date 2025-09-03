package main

import (
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"strconv"
)

type Vin struct {
	Txid    string  `json:"txid"`
	Vout    uint32  `json:"vout"`
	Address string  `json:"address"`
	Value   float64 `json:"value"`
}

type extVin struct {
	Txid    string `json:"txid"`
	Vout    uint32 `json:"vout"`
	Address string `json:"address"`
	Value   []byte `json:"value"`
}

func (v *Vin) DecodeRLP(s *rlp.Stream) error {
	var ext extVin
	if err := s.Decode(&ext); err != nil {
		return err
	}
	v.Txid = ext.Txid
	v.Vout = ext.Vout
	v.Address = ext.Address
	value, err := strconv.ParseFloat(string(ext.Value), 64)
	if err != nil {
		value = 0
	}
	v.Value = value
	return nil
}

func (v *Vin) EncodeRLP(w io.Writer) error {
	value := []byte(strconv.FormatFloat(v.Value, 'f', 8, 64))
	return rlp.Encode(w, extVin{
		Txid:    v.Txid,
		Vout:    v.Vout,
		Address: v.Address,
		Value:   value,
	})
}

type Vout struct {
	Index   uint32  `json:"index"`
	Address string  `json:"address"`
	Value   float64 `json:"value"`
}

type extVout struct {
	Index   uint32 `json:"index"`
	Address string `json:"address"`
	Value   []byte `json:"value"`
}

func (v *Vout) DecodeRLP(s *rlp.Stream) error {
	var ext extVout
	if err := s.Decode(&ext); err != nil {
		return err
	}
	v.Index = ext.Index
	v.Address = ext.Address
	value, err := strconv.ParseFloat(string(ext.Value), 64)
	if err != nil {
		value = 0
	}
	v.Value = value
	return nil
}

func (v *Vout) EncodeRLP(w io.Writer) error {
	value := []byte(strconv.FormatFloat(v.Value, 'f', 8, 64))
	return rlp.Encode(w, extVout{
		Index:   v.Index,
		Address: v.Address,
		Value:   value,
	})
}

type Tx struct {
	Txid   string  `json:"txid"`
	Vins   []*Vin  `json:"vins"`
	Vouts  []*Vout `json:"vouts"`
	Height int64   `json:"height"`
	Time   int64   `json:"time"`
}

type extTx struct {
	Txid  string  `json:"txid"`
	Vins  []*Vin  `json:"vins"`
	Vouts []*Vout `json:"vouts"`
}

func (t *Tx) DecodeRLP(s *rlp.Stream) error {
	var et extTx
	if err := s.Decode(&et); err != nil {
		return err
	}
	t.Txid, t.Vins, t.Vouts = et.Txid, et.Vins, et.Vouts
	return nil
}

func (t *Tx) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, extTx{
		Txid:  t.Txid,
		Vins:  t.Vins,
		Vouts: t.Vouts,
	})
}

// Block represents a block in the blockchain
type Block struct {
	Height int64  `json:"height"`
	Hash   string `json:"hash"`
	Tx     []*Tx  `json:"tx"`
}

// DecodeRLP implements rlp.Decoder
func (b *Block) DecodeRLP(s *rlp.Stream) error {
	var data []interface{}
	if err := s.Decode(&data); err != nil {
		return err
	}
	b.Height = data[0].(int64)
	b.Hash = data[1].(string)
	for _, v := range data[2].([]interface{}) {
		tx := &Tx{}
		if err := tx.DecodeRLP(rlp.NewStream(v.(io.Reader), 0)); err != nil {
			return err
		}
		b.Tx = append(b.Tx, tx)
	}
	return nil
}

// EncodeRLP implements rlp.Encoder
func (b *Block) EncodeRLP(w io.Writer) error {
	var txs []interface{}
	for _, v := range b.Tx {
		txs = append(txs, v)
	}
	return rlp.Encode(w, []interface{}{b.Height, b.Hash, txs})
}
