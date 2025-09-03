package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	FromBlock   int64        `json:"from_block"`
	DbPath      string       `json:"db_path"`
	Server      string       `json:"server"`
	Chain       Chain        `json:"chain"`
	ChainConfig ChainConfig  `json:"chain_config"`
}

type Chain struct {
	ChainName string `json:"chain_name"`
	RPC       string `json:"rpc"`
	UserName  string `json:"user_name"`
	PassWord  string `json:"pass_word"`
}

type ChainConfig struct {
	PubKeyHashAddrID        int     `json:"pub_key_hash_addr_id"`
	ScriptHashAddrID        int     `json:"script_hash_addr_id"`
	PrivateKeyID            int     `json:"private_key_id"`
	WitnessPubKeyHashAddrID int     `json:"witness_pub_key_hash_addr_id"`
	WitnessScriptHashAddrID int     `json:"witness_script_hash_addr_id"`
	HDPublicKeyID           []int   `json:"hd_public_key_id"`
	HDPrivateKeyID          []int   `json:"hd_private_key_id"`
	HDCoinType              int     `json:"hd_coin_type"`
}

func LoadConfig(cfg *Config, filep string) {

	// Default config.
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	if filep != "" {
		configFileName = filep
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func (cfg *Config) GetConfig() *Config {
	return cfg
}
