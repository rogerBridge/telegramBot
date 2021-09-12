package components

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

type Ticker struct {
	gorm.Model
	BestAsk      float64 `json:"best_ask"`
	BestBid      float64 `json:"best_bid"`
	InstrumentId string  `json:"instrument_id"`
	ProductId    string  `json:"product_id" gorm:"index"`
	Open24h      float64 `json:"open_24h"`
	High24h      float64 `json:"high_24h"`
	Low24h       float64 `json:"low_24h"`
	Timestamp    string  `json:"timestamp"`
}

type TickerOriginal struct {
	gorm.Model
	BestAsk      string `json:"best_ask"`
	BestBid      string `json:"best_bid"`
	InstrumentId string `json:"instrument_id"`
	ProductId    string `json:"product_id" gorm:"index"`
	Open24h      string `json:"open_24h"`
	High24h      string `json:"high_24h"`
	Low24h       string `json:"low_24h"`
	Timestamp    string `json:"timestamp"`
}

func GetAllTicker() ([]*Ticker, error) {
	client := FastHttpClient

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(fasthttp.MethodGet)

	host := "www.okex.com"
	u := &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/api/spot/v3/instruments/ticker",
	}

	req.Header.Set("Host", host)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	log.Println("request uri: ", u.String())
	req.SetRequestURI(u.String())

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := client.Do(req, res)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("client do:")

	data, err := res.BodyGunzip()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	result := make([]*TickerOriginal, 0, 128)
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// convert []*Ticker to []*TickerOriginal
	tickers := make([]*Ticker, 0, 1024)
	for _, v := range result {
		tickers = append(tickers, &Ticker{
			BestAsk:      stringToFloat64(v.BestAsk),
			BestBid:      stringToFloat64(v.BestBid),
			InstrumentId: v.InstrumentId,
			ProductId:    v.ProductId,
			Open24h:      stringToFloat64(v.Open24h),
			High24h:      stringToFloat64(v.High24h),
			Low24h:       stringToFloat64(v.Low24h),
			Timestamp:    v.Timestamp,
		})
	}
	return tickers, nil
}

func stringToFloat64(input string) float64 {
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		log.Printf("while parse string to float64, error: %v\n", err)
		log.Fatalln(err)
	}
	return v
}

func GetAllTickerMap() (map[string]*Ticker, error) {
	allTicker, err := GetAllTicker()
	if err != nil {
		return nil, err
	}
	allTickerMap := make(map[string]*Ticker)
	for _, v := range allTicker {
		allTickerMap[v.ProductId] = v
	}
	return allTickerMap, nil
}

func ProcessSpecificTicker(args ...string) (string, error) {
	allTickerMap, err := GetAllTickerMap()
	if err != nil {
		return "", err
	}

	specificTickers := args
	if len(args) == 0 {
		specificTickers = []string{"BTC-USDT"}
	}
	s := ""
	for _, v := range specificTickers {
		tString := allTickerMap[v].Timestamp
		t, err := time.Parse("2006-01-02T15:04:05.999Z", tString)
		if err != nil {
			log.Println(err)
			return "", err
		}
		t = t.In(Loc)
		s += fmt.Sprintf("%s: ask: %v, bid: %v\n24h_high: %v, 24h_low: %v, wave motion: %.3f%%, timestamp: %v\n", v, allTickerMap[v].BestAsk, allTickerMap[v].BestBid, allTickerMap[v].High24h, allTickerMap[v].Low24h, 100*(allTickerMap[v].High24h-allTickerMap[v].Low24h)/allTickerMap[v].Low24h, t)
	}
	return s, nil
}

func GetSpecificTicker(args ...string) ([]*Ticker, error) {
	allTickerMap, err := GetAllTickerMap()
	if err != nil {
		return nil, err
	}
	// log.Printf("%+v\n", allTickerMap)
	// get ticker from allTickerMap
	specificTickers := make([]*Ticker, 0, 16)
	for _, v := range args {
		specificTickers = append(specificTickers, allTickerMap[v])
	}
	return specificTickers, nil
}

func CreateSpecificTickersContinuousToSqlite(args ...string) {
	for {
		tickers, err := GetSpecificTicker(args...)
		if err != nil {
			log.Println(err)
		}
		err = CreateSpecificTickers(SqliteConn, tickers)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Minute / EVERY_MINUTE_SAMPLING)
	}
}

// AnalysisSpecificTickers("BTC-USDT", "ETH-USDT")
func AnalysisSpecificTickers(args ...string) string {
	s := ""
	log.Printf("analysis specific tickers...\n")
	for _, v := range args {
		tickers := QuerySpecificTicker(v)
		if len(tickers) >= FIVE_MINUTES {
			readyForAnalysis := tickers[:FIVE_MINUTES]
			s += AnalysisTickers(readyForAnalysis, "In last 5min "+v)
		} else {
			s += fmt.Sprintf("%s: %s\n", v, "data sample is too little\n")
			continue
		}
		if len(tickers) >= ONE_HOUR {
			readyForAnalysis := tickers[:ONE_HOUR]
			s += AnalysisTickers(readyForAnalysis, "In last 1 hour "+v)
		} else {
			s += "\n"
			continue
		}
		if len(tickers) >= ONE_DAY {
			readyForAnalysis := tickers[:ONE_DAY]
			s += AnalysisTickers(readyForAnalysis, "In last 1 day "+v)
		} else {
			s += "\n"
			continue
		}
		if len(tickers) >= ONE_WEEK {
			readyForAnalysis := tickers[:ONE_WEEK]
			s += AnalysisTickers(readyForAnalysis, "In last 7 day "+v)
			s += "\n"
		} else {
			s += "\n"
			continue
		}
	}
	return s
}

func AnalysisTickers(t []*Ticker, notify string) string {
	min := t[0]
	max := t[0]
	for _, v := range t {
		if v.BestAsk < min.BestAsk {
			min = v
		}
		if v.BestAsk > max.BestAsk {
			max = v
		}
	}
	maxTimestamp, err := time.Parse("2006-01-02T15:04:05.999Z", max.Timestamp)
	if err != nil {
		log.Fatalln(err)
	}
	maxTimestamp = maxTimestamp.In(Loc)
	minTimestamp, err := time.Parse("2006-01-02T15:04:05.999Z", min.Timestamp)
	if err != nil {
		log.Fatalln(err)
	}
	minTimestamp = minTimestamp.In(Loc)
	return fmt.Sprintf("%s:\n max: %v %s, min: %v %s, change: %.3f%%\n", notify, max.BestAsk, maxTimestamp, min.BestAsk, minTimestamp, 100*(max.BestAsk-min.BestAsk)/min.BestAsk)
}

func AnalysisTickersAndOutputByPercent(t []*Ticker, notify string, compare float64) string {
	min := t[0]
	max := t[0]
	for _, v := range t {
		if v.BestAsk < min.BestAsk {
			min = v
		}
		if v.BestAsk > max.BestAsk {
			max = v
		}
	}
	if (max.BestAsk-min.BestAsk)/min.BestAsk >= compare {
		maxTimestamp, err := time.Parse("2006-01-02T15:04:05.999Z", max.Timestamp)
		if err != nil {
			log.Fatalln(err)
		}
		maxTimestamp = maxTimestamp.In(Loc)
		minTimestamp, err := time.Parse("2006-01-02T15:04:05.999Z", min.Timestamp)
		if err != nil {
			log.Fatalln(err)
		}
		minTimestamp = minTimestamp.In(Loc)
		return fmt.Sprintf("%s:\n max: %v %s, min: %v %s, minus: %.3f%%\n", notify, max.BestAsk, maxTimestamp, min.BestAsk, minTimestamp, 100*(max.BestAsk-min.BestAsk)/min.BestAsk)
	}
	return ""
}

const EVERY_MINUTE_SAMPLING = 1
const FIVE_MINUTES = 5 * EVERY_MINUTE_SAMPLING
const ONE_HOUR = 60 * EVERY_MINUTE_SAMPLING
const ONE_DAY = 24 * 60 * EVERY_MINUTE_SAMPLING
const ONE_WEEK = 7 * 24 * 60 * EVERY_MINUTE_SAMPLING

func CryptoCurrencyDaemon(b *tb.Bot, args ...string) {
	myGroup := &tb.User{ID: -1001524256686}

	type Flag struct {
		LastSend       int64
		SendTimesCount int
	}

	lastSendTimestampMap := make(map[string]*Flag)
	for _, v := range args {
		lastSendTimestampMap[v] = new(Flag)
	}
	const INTERVAL_ONE = 120
	const INTERVAL_TWO = 600

	for {
		sendFlag := false
		reportString := ""
		log.Printf("analysis specific tickers...\n")
		for _, v := range args {
			tickers := QuerySpecificTicker(v)
			nowTimestamp := time.Now().Unix()
			if len(tickers) >= FIVE_MINUTES {
				readyForAnalysis := tickers[:FIVE_MINUTES]
				r := AnalysisTickersAndOutputByPercent(readyForAnalysis, "In last 5 min "+v, 0.025)
				if r != "" {
					// first 3 times, push interval 2mins, then, push interval 10mins
					if lastSendTimestampMap[v].SendTimesCount < 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_ONE {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else if lastSendTimestampMap[v].SendTimesCount >= 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_TWO {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else {
						log.Println("don't match send condition", lastSendTimestampMap[v])
					}
				}
			}

			if len(tickers) >= ONE_HOUR {
				readyForAnalysis := tickers[:ONE_HOUR]
				r := AnalysisTickersAndOutputByPercent(readyForAnalysis, "In last 1 hour "+v, 0.05)
				if r != "" {
					// first 3 times, push interval 2mins, then, push interval 10mins
					if lastSendTimestampMap[v].SendTimesCount < 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_ONE {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else if lastSendTimestampMap[v].SendTimesCount >= 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_TWO {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else {
						log.Println("don't match send condition", lastSendTimestampMap[v])
					}
				}
			}

			if len(tickers) >= ONE_DAY {
				readyForAnalysis := tickers[:ONE_DAY]
				r := AnalysisTickersAndOutputByPercent(readyForAnalysis, "In last 1 day "+v, 0.1)
				if r != "" {
					// first 3 times, push interval 2mins, then, push interval 10mins
					if lastSendTimestampMap[v].SendTimesCount < 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_ONE {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else if lastSendTimestampMap[v].SendTimesCount >= 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_TWO {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else {
						log.Println("don't match send condition", lastSendTimestampMap[v])
					}
				}
			}

			if len(tickers) >= ONE_WEEK {
				readyForAnalysis := tickers[:ONE_WEEK]
				r := AnalysisTickersAndOutputByPercent(readyForAnalysis, "In last one week "+v, 0.2)
				if r != "" {
					// first 3 times, push interval 2mins, then, push interval 10mins
					if lastSendTimestampMap[v].SendTimesCount < 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_ONE {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else if lastSendTimestampMap[v].SendTimesCount >= 3 && nowTimestamp-lastSendTimestampMap[v].LastSend >= INTERVAL_TWO {
						sendFlag = true
						lastSendTimestampMap[v].LastSend = nowTimestamp
						lastSendTimestampMap[v].SendTimesCount += 1
						reportString += r
					} else {
						log.Println("don't match send condition", lastSendTimestampMap[v])
					}
				}
			} else {
				reportString += "\n"
			}
			// reset sendTimesCount is no new reportString generated
			if len(strings.ReplaceAll(reportString, "\n", "")) == 0 && lastSendTimestampMap[v].SendTimesCount != 0 {
				lastSendTimestampMap[v].SendTimesCount = 0
			}
		}
		log.Printf("reportString: %s, sendFlag: %v\n", reportString, sendFlag)
		if sendFlag {
			b.Send(myGroup, reportString)
			time.Sleep(time.Minute)
		} else {
			time.Sleep(time.Minute)
		}
	}
}
