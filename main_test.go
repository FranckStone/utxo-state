package main

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
)

func TestReset(t *testing.T) {

	// 加载配置
	var cfg Config
	LoadConfig(&cfg, "")

	db, err := leveldb.OpenFile(cfg.DbPath, nil)
	if err != nil {
		panic(fmt.Sprintf("Leveldb err %s", err))
	}
	defer db.Close()
	RawDB := &RawDB{DB: db}
	err = RawDB.SetBalance("D7EHnqQ3asCiShoDfJWigr7j489ES8HVCi", 17.883190000000006)
	if err != nil {
		t.Error(err)
	}

}
