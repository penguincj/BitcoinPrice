package main

import (
	"fmt"
	"time"
	"strings"

	"github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"

	"github.com/penguincj/tokenMarket/models"
	"github.com/penguincj/tokenMarket/services"
)

const EPSINON = 0.0000001

type transactionPair struct {
	name string
	pair string
}

var tokenMap = map[int64]transactionPair{
    1: transactionPair{"BTC", "btcusdt"},
    2: transactionPair{"ETH", "ethusdt"},
    3: transactionPair{"EOS", "eosusdt"},
    4: transactionPair{"HT", "htusdt"},
    5: transactionPair{"LTC", "ltcusdt"},
    6: transactionPair{"XRP", "xrpusdt"},
    7: transactionPair{"DASH", "dashusdt"},
    8: transactionPair{"ETC", "etcusdt"},
}

func getTokenPrice(pair transactionPair) models.TokenPrice {
	htKLine := services.GetKLine(pair.pair, "1day", 1)
	if htKLine.Status != "ok" {
		fmt.Printf("get %s failed", pair.pair)
		return models.TokenPrice{}
	}

	kLineData := htKLine.Data[0]

	token := models.TokenPrice{
		Name:  pair.name,
		Pair:  pair.pair,
		Open:  kLineData.Open,
		Close: kLineData.Close,
	}

	if (token.Open >= -EPSINON) && (token.Open <= EPSINON) {
		fmt.Println("open price is zero")
		return models.TokenPrice{}
	}

	token.CloseS = fmt.Sprintf("%0.2f", token.Close)
	floatPercent := ((token.Close - token.Open) / token.Open) * 100
	floatPercentS := fmt.Sprintf("%0.2f", floatPercent)
	token.FloatPercent = floatPercentS
	if strings.Contains(token.FloatPercent, "-") {
		token.Fluctuation = "negative"
	} else {
		token.Fluctuation = "positive"
	}
	//fmt.Printf("%s float %0.2f \n", pair.name, floatPercent)

	return token
}

type Profile struct {
    Id          int
    Age         int16
}

func main() {
	fmt.Println("start")

	var tokenPriceMap = make(map[string]models.TokenPrice, 8)

	for id, tokenName := range tokenMap {
        tokenPrice := models.TokenPrice {
            Id: id,
            Name: tokenName.name,
            Pair: tokenName.pair,
        }
		err := tokenPrice.Insert()
		if err != nil {
			fmt.Println("insert %s failed, err: %s", tokenName, err)
		}
	}

	t := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-t.C:
			t.Reset(1 * time.Minute)
			for id, tokenName := range tokenMap {
				token := getTokenPrice(tokenName)
				tokenPriceMap[tokenName.name] = token
				token.Id = id
				err := token.Update("name", "pair", "open", "close", "close_s", "float_percent", "fluctuation")
				if err != nil {
					fmt.Println("update %s failed, err: %s", tokenName, err)
				}
			}

			/*
			for name, tokenPrice := range tokenPriceMap {
				fmt.Printf("%s float %s \n", name, tokenPrice.FloatPercent)
			}
			*/
		}
	}
}

func init() {
    orm.RegisterModel(new(Profile))
    orm.RegisterDriver("mysql", orm.DRMySQL)

    orm.RegisterDataBase("default", "mysql", "root:mysqlroot@/airdrop?charset=utf8")
	orm.RunSyncdb("default", false, true)
}

