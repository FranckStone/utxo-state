# 配置文件说明

本项目支持多种区块链网络，每种网络都有独立的配置文件。

## 可用的配置文件

所有配置文件都位于 `configs/` 文件夹中：

### 1. configs/config-litecoin.json - Litecoin 主网
- **RPC 端口**: 9332
- **数据目录**: data/litecoin/db
- **地址前缀**: L (普通地址), M (P2SH地址)
- **私钥格式**: 6 (WIF格式)

### 2. configs/config-dogecoin.json - Dogecoin 主网
- **RPC 端口**: 22555
- **数据目录**: data/dogecoin/db
- **地址前缀**: D (普通地址), A (P2SH地址)
- **私钥格式**: Q (WIF格式)

### 3. configs/config-bitcoin.json - Bitcoin 主网
- **RPC 端口**: 8332
- **数据目录**: data/bitcoin/db
- **地址前缀**: 1 (普通地址), 3 (P2SH地址)
- **私钥格式**: 5 (WIF格式)

### 4. configs/config-testnet.json - 测试网
- **RPC 端口**: 18332
- **数据目录**: data/testnet/db
- **地址前缀**: m/n (普通地址), 2 (P2SH地址)
- **私钥格式**: 9 (WIF格式)



## 使用方法

### 默认配置
如果不指定配置文件，程序将使用 `config.json` 作为默认配置。

## 配置参数说明

### 基础配置
- `from_block`: 开始同步的区块高度，0表示从数据库记录的高度开始
- `db_path`: 数据库存储路径
- `server`: HTTP服务器监听地址和端口

### 链配置
- `chain_name`: 区块链网络名称
- `rpc`: RPC服务器地址和端口
- `user_name`: RPC用户名
- `pass_word`: RPC密码

### 链参数配置
- `pub_key_hash_addr_id`: 公钥哈希地址的版本字节
- `script_hash_addr_id`: 脚本哈希地址的版本字节
- `private_key_id`: 私钥WIF格式的版本字节
- `witness_pub_key_hash_addr_id`: 见证公钥哈希地址的版本字节
- `witness_script_hash_addr_id`: 见证脚本哈希地址的版本字节
- `hd_public_key_id`: HD钱包公钥的版本字节数组
- `hd_private_key_id`: HD钱包私钥的版本字节数组
- `hd_coin_type`: BIP44币种类型

### **网络参数对比**

| 网络 | RPC端口 | 地址前缀 | 私钥格式 | 币种类型 |
|------|---------|----------|----------|----------|
| Bitcoin | 8332 | 1, 3 | 5 | 0 |
| Litecoin | 9332 | L, M | 6 | 2 |
| Dogecoin | 22555 | D, A | Q | 3 |
| 测试网 | 18332 | m, n, 2 | 9 | 1 |

## 注意事项

1. **数据目录**: 不同网络使用不同的数据目录，避免数据混淆
2. **RPC端口**: 确保配置的RPC端口与您的节点配置一致
3. **密码安全**: 生产环境中请使用强密码，不要使用默认密码
4. **网络切换**: 切换网络时，建议清空或备份原有数据目录


## 配置文件内容

### 1. Bitcoin 主网配置 (config-bitcoin.json)
```json
{
  "from_block": 0,
  "db_path": "data/bitcoin/db",
  "server": ":8082",
  "chain": {
    "chain_name": "bitcoin",
    "rpc": "YOUR_RPC_HOST:8332",
    "user_name": "YOUR_RPC_USERNAME",
    "pass_word": "YOUR_RPC_PASSWORD"
  },
  "chain_config": {
    "pub_key_hash_addr_id": 0,
    "script_hash_addr_id": 5,
    "private_key_id": 128,
    "witness_pub_key_hash_addr_id": 0,
    "witness_script_hash_addr_id": 0,
    "hd_public_key_id": [4, 136, 178, 30],
    "hd_private_key_id": [4, 136, 173, 244],
    "hd_coin_type": 0
  }
}
```

### 2. Dogecoin 主网配置 (config-dogecoin.json)
```json
{
  "from_block": 0,
  "db_path": "data/dogecoin/db",
  "server": ":8082",
  "chain": {
    "chain_name": "dogecoin",
    "rpc": "YOUR_RPC_HOST:22555",
    "user_name": "YOUR_RPC_USERNAME",
    "pass_word": "YOUR_DOGECOIN_RPC_PASSWORD"
  },
  "chain_config": {
    "pub_key_hash_addr_id": 56,
    "script_hash_addr_id": 22,
    "private_key_id": 158,
    "witness_pub_key_hash_addr_id": 0,
    "witness_script_hash_addr_id": 0,
    "hd_public_key_id": [2, 250, 202, 253],
    "hd_private_key_id": [2, 250, 195, 152],
    "hd_coin_type": 3
  }
}
```

### 3. Litecoin 主网配置 (config-litecoin.json)
```json
{
  "from_block": 0,
  "db_path": "data/litecoin/db",
  "server": ":8082",
  "chain": {
    "chain_name": "litecoin",
    "rpc": "YOUR_RPC_HOST:9332",
    "user_name": "YOUR_RPC_USERNAME",
    "pass_word": "YOUR_RPC_PASSWORD"
  },
  "chain_config": {
    "pub_key_hash_addr_id": 48,
    "script_hash_addr_id": 50,
    "private_key_id": 176,
    "witness_pub_key_hash_addr_id": 0,
    "witness_script_hash_addr_id": 0,
    "hd_public_key_id": [4, 136, 178, 30],
    "hd_private_key_id": [4, 136, 173, 244],
    "hd_coin_type": 2
  }
}
```

### 4. 测试网配置 (config-testnet.json)
```json
{
  "from_block": 0,
  "db_path": "data/testnet/db",
  "server": ":8082",
  "chain": {
    "chain_name": "testnet",
    "rpc": "YOUR_RPC_HOST:18332",
    "user_name": "YOUR_RPC_USERNAME",
    "pass_word": "YOUR_RPC_PASSWORD"
  },
  "chain_config": {
    "pub_key_hash_addr_id": 111,
    "script_hash_addr_id": 196,
    "private_key_id": 239,
    "witness_pub_key_hash_addr_id": 0,
    "witness_script_hash_addr_id": 0,
    "hd_public_key_id": [4, 136, 178, 30],
    "hd_private_key_id": [4, 136, 173, 244],
    "hd_coin_type": 1
  }
}
```

## 自定义配置

您可以根据需要修改任何配置文件，或者创建新的配置文件。所有配置参数都可以根据您的实际环境进行调整。

## 安全注意事项

⚠️ **重要安全提醒**：

1. **密码安全**：配置文件中的密码已进行脱敏处理，实际使用时请替换为您的真实RPC密码
2. **RPC访问**：请确保RPC服务仅对可信网络开放，避免暴露到公网
3. **用户权限**：建议为RPC创建专用用户，避免使用管理员账户
4. **网络安全**：生产环境中建议使用HTTPS和VPN等安全措施
5. **配置备份**：请妥善保管包含真实密码的配置文件，不要提交到版本控制系统

### 脱敏说明

本文档中的敏感信息已进行脱敏处理：
- `YOUR_RPC_HOST`: 替换为您的RPC服务器地址
- `YOUR_RPC_USERNAME`: 替换为您的RPC用户名  
- `YOUR_RPC_PASSWORD`: 替换为您的RPC密码
- `YOUR_DOGECOIN_RPC_PASSWORD`: 替换为您的Dogecoin RPC密码
