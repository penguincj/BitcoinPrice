package main

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"

	"github.com/penguincj/tokenMarket/models"
	"github.com/penguincj/tokenMarket/services"
)

const EPSINON = 0.0000001

var tokenMap = map[int64]string{
    1: "btcusdt",
    2: "ethusdt",
    3: "eosusdt",
    4: "htusdt",
    5: "ltcusdt",
    6: "xrpusdt",
    7: "dashusdt",
    8: "etcusdt",
}

func getTokenPrice(name string) models.TokenPrice {
	htKLine := services.GetKLine(name, "1day", 1)
	if htKLine.Status != "ok" {
		fmt.Printf("get %s failed", name)
		return models.TokenPrice{}
	}

	kLineData := htKLine.Data[0]

	token := models.TokenPrice{
		Name:  name,
		Open:  kLineData.Open,
		Close: kLineData.Close,
	}

	if (token.Open >= -EPSINON) && (token.Open <= EPSINON) {
		fmt.Println("open price is zero")
		return models.TokenPrice{}
	}

	floatPercent := ((token.Close - token.Open) / token.Open) * 100
	floatPercentS := fmt.Sprintf("%0.2f", floatPercent)
	token.FloatPercent = floatPercentS
	fmt.Printf("%s float %0.2f \n", name, floatPercent)

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
            Name: tokenName,
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
				tokenPriceMap[tokenName] = token
				token.Id = id
				err := token.Update("name", "open", "close", "float_percent")
				if err != nil {
					fmt.Println("update %s failed, err: %s", tokenName, err)
				}
			}

			for name, tokenPrice := range tokenPriceMap {
				fmt.Printf("%s float %s \n", name, tokenPrice.FloatPercent)
			}
		}
	}
}

func init() {
    orm.RegisterModel(new(Profile))
    orm.RegisterDriver("mysql", orm.DRMySQL)

    orm.RegisterDataBase("default", "mysql", "root:mysqlroot@/airdrop?charset=utf8")
	orm.RunSyncdb("default", false, true)
}

