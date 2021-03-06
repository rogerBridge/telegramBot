package components

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// get specific ticker, for example:  "BTC-USDT"
func CreateSpecificTickers(tx *gorm.DB, tickers []*Ticker) error {
	err := tx.Create(&tickers).Error
	return err
}

// the newer, the more forward
func QuerySpecificTicker(arg string) []*Ticker {
	tickers := make([]*Ticker, 0, 1024)
	SqliteConn.Model(&Ticker{}).Where("product_id=?", arg).Order("id desc").Limit(10800).Find(&tickers)
	log.Printf("Query ticker is: %v, length of tickers is: %v\n", arg, len(tickers))
	return tickers
}

// args: "BTC-USDT", "ETH-USDT" ...
func DeleteOutdateTicker(tx *gorm.DB, args ...string) error {
	maxRange := EVERY_MINUTE_SAMPLING * 60 * 24 * 7
	for _, v := range args {
		tickers := QuerySpecificTicker(v)
		if len(tickers) >= maxRange {
			log.Printf("%v is too much, delete it now\n", v)
			flagID := tickers[len(tickers)-1].ID
			err := tx.Unscoped().Where("id < ?", flagID).Delete(&Ticker{}).Error
			if err != nil {
				log.Printf("while delete, error: %v\n", err)
				return err
			}
			// all tickers smaller than flagID should be deleted
		}
	}
	return nil
}

func DeleteOutdateTickerTiming(args ...string) {
	for {
		t := time.Now()
		if t.Hour() == 0 && t.Minute() == 0 {
			err := DeleteOutdateTicker(SqliteConn, args...)
			if err != nil {
				log.Println(err)
			}
		}
		time.Sleep(time.Minute)
	}
}
