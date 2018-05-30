package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

func sync() {
	ctx := context.Background()
	nodeClient, err := ethclient.Dial(config.EthRPC)
	if err != nil {
		fmt.Println(config.EthRPC)
		log.Fatalln(err.Error())
	}

	esClient, err := elastic.NewClient(elastic.SetURL(config.ElasticURL), elastic.SetSniff(config.ElasticSniff))
	if err != nil {
		log.Fatalln(err.Error())
	}

	indices := []string{"esblock", "estx", "esaccount", "escontract"}
	for _, index := range indices {
		switch index {
		case "esblock":
			esIndex(ctx, esClient, index, blockMapping)
		case "estx":
			esIndex(ctx, esClient, index, txMapping)
		case "esaccount":
		case "escontract":
			esIndex(ctx, esClient, index, contractMapping)
		}
	}

	block, err := nodeClient.BlockByNumber(ctx, big.NewInt(3))
	if err != nil {
		log.Fatalln(err.Error())
	}

	blockParams := esBlockFunc(block)

	esClient.Index().Index("esblock").Type("block").Id(block.Number().String()).BodyJson(blockParams).Do(ctx)
	txs := block.Transactions()
	for _, tx := range txs {
		fmt.Println("tx", tx.Hash().Hex())
		fmt.Println("to", tx.To())
	}
	d, err := nodeClient.NonceAt(ctx, common.HexToAddress("0xd24400ae8BfEBb18cA49Be86258a3C749cf46853"), nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(d)
}

func esIndex(ctx context.Context, client *elastic.Client, index, mapping string) {
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if !exists {
		result, err := client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			log.Fatalf(err.Error())
		}
		if !result.Acknowledged {
			log.Fatalf("create index faild")
		}

	}
}

func esBlockFunc(block *types.Block) interface{} {
	rawTxs := block.Transactions()
	var txs []string
	for _, tx := range rawTxs {
		txs = append(txs, tx.Hash().Hex())
	}

	b := map[string]interface{}{
		"height":     block.Header().Number,
		"hash":       block.Hash().Hex(),
		"time":       block.Time().String(),
		"parenthash": block.ParentHash().Hex(),
		"sha3uncles": block.UncleHash().Hex(),
		"miner":      block.Coinbase().Hex(),
		"difficulty": block.Difficulty(),
		"size":       float64(block.Size()),
		"gasused":    block.GasUsed(),
		"gaslimit":   block.GasLimit(),
		"nonce":      block.Nonce(),
		"txs":        txs,
	}
	return b
}

func esTxFunc(from, to, bhash, thash string, value big.Int) interface{} {
	t := map[string]interface{}{
		"thash": thash,
		"bhash": bhash,
		"from":  from,
		"to":    to,
		"value": value,
	}
	return t
}

type esBlock struct {
	Height     big.Int  `json:"height"`
	Hash       string   `json:"hash"`
	Time       string   `json:"time"`
	Sha3Uncles string   `json:"sha3uncles"`
	Miner      string   `json:"miner"`
	Difficulty big.Int  `json:"difficulty"`
	Size       float64  `json:"size"`
	GasLimit   uint64   `json:"gaslimit"`
	GasUsed    uint64   `json:"gasused"`
	Nonce      uint64   `json:"nonce"`
	Txs        []string `json:"txs"`
}

type esTx struct {
	THash string  `json:"thash"`
	BHash string  `json:"bhash"`
	From  string  `json:"from"`
	To    string  `json:"to"`
	Value big.Int `json:"value"`
}

type esContract struct {
	Owner string `json:"owner"`
	Tx    string `json:"tx"`
	ABI   string `json:"abi"`
}

const blockMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "block": {
      "properties": {
        "hash": {
          "type": "keyword"
        },
        "size": {
          "type": "integer"
        },
        "height": {
          "type": "long"
        },
        "sha3uncles": {
          "type": "text"
        },
				"time": {
					"type": "long"
				},
				"miner": {
					"type": "text"
				},
				"nonce": {
					"type": "long"
				},
				"difficulty": {
					"type": "long"
				},
				"size": {
					"type": "double"
				},
				"size": {
					"type": "double"
				},
				"gaslimit": {
					"type": "long"
				},
				"gasused": {
					"type": "long"
				},
				"txs": {
					"type":"keyword"
				}
      }
    }
  }
}`

const txMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "tx": {
      "properties": {
        "thash": {
          "type": "keyword"
        },
				"bhash": {
					"type": "keyword"
				},
				"from": {
					"type": "keyword"
				},
				"to": {
					"type": "keyword"
				},
				"value": {
					"type": "double"
				}
      }
    }
  }
}`

const contractMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "contract": {
      "properties": {
        "owner": {
          "type": "keyword"
        },
				"tx": {
					"type": "text"
				},
				"abi": {
					"type": "text"
				}
      }
    }
  }
}`
