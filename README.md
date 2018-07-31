## Query bitcoin balance and utxo -- Solution for wallet to construct raw tx
Hey my friends, see you again. Today I want to share with your guy programers about how to query bitcoin balance and uxto without any bitcoin node such as bitcoin-core etc.

These days, for fun and killing my borning spare time, I focus on dumping all bitcoin mainnet's chaindata to elasticsearch and keeping my database synchronously with the bitcoin mainnet best height. For two things:
- Open API for query balance and uxto
- parse all transactions in every new best block, send message by websocket when the tx is related with addresses in business system

As we blockchain developers all know, bitcoin system uses UTXO model to trace balance, big difference from account based blockchain like Ethereum, the second biggest project in the cryptocurrency market. We would not get balance and unspent vout conveniently if we're not dependent on bitcoin full data node. So, we need a constructed bitcoin ledger as a specify schema.

In this post, I'm not consider to talk about how to dump bitcoin chaindata to elasticsearch, maybe next post will. I have make [bitcoin-balance-utxo-api](https://github.com/wenweih/bitcoin-balance-utxo-api) repository public in Github,  writing in Go. Want to go deep about the repo? go ahead...

#### 0x01 Demo
```shell
/* start */
▶ bitcoin-service-external-api
{"level":"info","msg":"Using Configure file:/Users/hww/bitcoin-service-external-api.yml"}
[GIN] 2018/07/31 - 02:20:02 | 200 |   44.135717ms |       127.0.0.1 |  GET     /balance/1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nh
[GIN] 2018/07/31 - 02:20:11 | 200 |  349.226523ms |       127.0.0.1 |  GET     /utxo/1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nh
[GIN] 2018/07/31 - 02:21:54 | 404 |     458.573µs |       127.0.0.1 |  GET     /utxo/
[GIN] 2018/07/31 - 02:21:59 | 400 |     139.048µs |       127.0.0.1 |  GET     /utxo/1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nhx

---
/* Response */
▶ curl 127.0.0.1:3000/balance/1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nh
{"address":"1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nh","balance":1.1005578400000005,"status":200}%
~
▶ curl 127.0.0.1:3000/utxo/1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nh
{"address":"1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nh","status":200,"utxos":[{"txid":"43a95fc3e3095f825bc48ad16449a680580cc01da781962ec5a2c27089f5d188","amount":0.01,"voutindex":0},{"txid":"86b16b0521eab71ee15a4090ae0eef8f67aa7f8c3adb9b838f9d717b538a73b8","amount":0.01,"voutindex":0},{"txid":"8a383edddc474d22cd02e47044e13165831d8fdc880af7b1cdd9ae0d16dd7d07","amount":0.01,"voutindex":0},{"txid":"05e9cf3814106ee90cd56d7f979abd7d1d2cf645e3fc76c85ad3b2d8953ff138","amount":0.01,"voutindex":0},{"txid":"c8e499dce8cb2cef9f18e4fae145f66ff89765cf88e0eef82dd602197fac97ee","amount":0.01,"voutindex":0},{"txid":"e1c1b50a7a5c0df6c04c44c89e1d2ab19c34dfb5b8c44fe078d09542aef6396f","amount":0.01,"voutindex":0},{"txid":"6e82f8bddea5fa492e6536de5afc4d62a447cc5b7429e4e14c476d80e9ab9fcb","amount":0.01,"voutindex":0},{"txid":"6289fc7638663640a96f521ce84bcf3ad94344bc24848a0464dc44a8f889c8cd","amount":0.01,"voutindex":0},{"txid":"f96d72dff3076292ad3a9d943aa396cda070401cd12a4824f776083dc4247f7a","amount":0.01,"voutindex":0},{"txid":"24d8316b2e41674010a4f049b10cf264cdb925705d44b6f92ebca2d4f21993d4","amount":0.01,"voutindex":0}]}%
~
▶ curl 127.0.0.1:3000/utxo/
{"message":"Route Error"}%
~
▶ curl 127.0.0.1:3000/utxo/1QKv2b5EzrtHqNAE9dBy4mcd2Wtr3A2Nhx
{"message":"Address format error: checksum mismatch","status":400}
```
#### 0x02 Tool
- [dep](https://github.com/golang/dep) for Go package management
- [gin](https://github.com/gin-gonic/gin) for web API
- [btcutil](github.com/btcsuite/btcutil) for address validate
- [elastic](github.com/olivere/elastic) for elasticsearch client
- [viper](github.com/spf13/viper) for configure
- elasticsearch v6.3.0

**Thanks all contributors for the nice open ource.**
#### 0x03 elasticsearch mapping
There are four es indices in our database: block, tx, vout, balance. here are vout and balance mappings:
```go
const voutMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
		"vout": {
      "properties": {
        "txidbelongto": {
          "type": "keyword"
        },
        "value": {
          "type": "double"
        },
        "voutindex": {
          "type": "keyword"
        },
        "coinbase": {
          "type": "boolean"
        },
        "addresses": {
          "type":"keyword"
        },
				"time": {
					"type": "long"
				},
        "used": {
          "properties": {
            "txid": {
              "type": "text"
            },
            "vinindex": {
              "type": "short"
            }
          }
        }
      }
    }
  }
}`
```
In vout mapping, **used** field is a nested object, **addresses** is an array of object. The example docutment:
```json
{
  "_index" : "vout",
  "_type" : "vout",
  "_id" : "Clrk4GQBQUbZgPHUI_tg",
  "_score" : 1.0,
  "_source" : {
    "txidbelongto" : "8f7cbb5cdddaa6548f00aafac4086c0788a7221dea4b7339b321d566a70d392d",
    "value" : 500,
    "voutindex" : 0,
    "coinbase" : false,
    "addresses" : [
      "1ALwHv6FTu6WTQhCA8yEaJBYceSL2Go66i"
    ],
  "used" : {
    "txid" : "cbf201195d2ae9def2bf47396a909d3f7ef709ad7c503ff4163dbb4865ab9e9c",
    "vinindex" : 0
    }
  }
}
```
balance mapping and example docutment:
```go
const balanceMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
		"balance": {
			"properties": {
				"address": {
					"type":"keyword"
				},
				"amount": {
					"type": "double"
				}
			}
		}
  }
}`

.....
{
  "_index" : "balance",
  "_type" : "balance",
  "_id" : "AVFH4GQBQUbZgPHU2UIR",
  "_score" : 1.0,
  "_source" : {
    "address" : "1GkQmKAmHtNfnD3LHhTkewJxKHVSta4m2a",
    "amount" : 50
  }
},
```
#### 0x04 Endpoint
```Go
func main() {
	r := ginEngine()
	r.GET("/balance/:address", balanceHandler)
	r.GET("/utxo/:address", utxosHandle)
	if err := r.Run(":3000"); err != nil {
		sugar.Fatal("New Router Error:", err.Error())
	}
}
```
The Endpoints are easy to understand, let's look at the handle function.

##### balanceHandler
```Go
func balanceHandler(c *gin.Context) {
	address := c.Param("address")
	// https://github.com/btcsuite/btcutil/blob/master/address_test.go
	_, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"Address format error:", err.Error()}, " ")))
		return
	}

	searchResult, err := esClient.Search().Index("balance").Type("balance").Query(elastic.NewTermQuery("address", address)).Do(context.TODO())
	if err != nil {
		ginResponseException(c, http.StatusNotFound, err)
		return
	}

	if len(searchResult.Hits.Hits) != 1 {
		ginResponseException(c, http.StatusNotFound, errors.New(strings.Join([]string{"Fail to query balance for", address}, " ")))
		return
	}
	b := new(Balance)
	if err := json.Unmarshal(*(searchResult.Hits.Hits[0].Source), b); err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New("unmarshal result error"))
		return
	}

	c.JSON(200, gin.H{
		"status":  http.StatusOK,
		"address": address,
		"balance": b.Amount,
	})
}
```
##### utxosHandle
```Go
func utxosHandle(c *gin.Context) {
	address := c.Param("address")
	// https://github.com/btcsuite/btcutil/blob/master/address_test.go
	_, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"Address format error:", err.Error()}, " ")))
		return
	}

	// addresses is an array datatype
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/array.html

	// exists Query
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-exists-query.html
	//
	// curl -H 'Content-Type: application/json' -XPOST http://47.75.159.52:9200/vout/vout/_search\?pretty\=true -d '
	// {
	//		"query": {"bool":{"must":{"term":{"addresses":"12cbQLTFMXRnSzktFkuoG3eHoMeFtpTu3S"}},"must_not":{"exists":{"field":"used"}}}}
	// }'
	q := elastic.NewBoolQuery().Must(elastic.NewTermQuery("addresses", address)).MustNot(elastic.NewExistsQuery("used"))
	searchResult, err := esClient.Search().Index("vout").Type("vout").Query(q).SortBy(elastic.NewFieldSort("value").Asc()).Do(context.TODO())
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"Query utxo error:", err.Error()}, " ")))
		return
	}

	var utxos []*utxo
	for _, vout := range searchResult.Hits.Hits {
		newVout := new(esVout)
		if err := json.Unmarshal(*vout.Source, newVout); err != nil {
			ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"fail to unmarshal esvout", err.Error()}, " ")))
			return
		}
		utxos = append(utxos, &utxo{Txid: newVout.TxIDBelongTo, Amount: newVout.Value, VoutIndex: newVout.Voutindex})
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"address": address,
		"utxos":   utxos,
	})
}
```
#### 0x05 Link
- [bitcoin-balance-utxo-api](https://github.com/wenweih/bitcoin-balance-utxo-api)
- [elasticsearch guide: sort](https://www.elastic.co/guide/cn/elasticsearch/guide/current/_Sorting.html)
- [elasticsearch guide: exists query](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-exists-query.html)
