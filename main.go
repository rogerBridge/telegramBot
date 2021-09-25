package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	_ "time/tzdata" // find tzdata even not found in system

	_ "github.com/breml/rootcerts"
	"github.com/rogerBridage/telegramBot/components"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	log.Printf("%+v\n", components.Config)
	components.InitSqlite()

	followProductIDs := components.Config.FollowProductIDs
	statsProductIDs := components.Config.StatsProductIDs
	// timing delete outdated ticker data
	go components.DeleteOutdateTickerTiming(followProductIDs...)
	// timing get newest ticker data to sqlite
	go components.CreateSpecificTickersContinuousToSqlite(followProductIDs...)

	// unionPay
	e := new(components.ExchangeRateCache)
	// freecurrency
	ee := new(components.ExchangeMajor)

	b := components.NewBot()

	// !Daemon
	go components.MultiWeatherDaemon(b, components.Config.FollowCities)
	go components.ExchangeDaemon(b)
	go components.CryptoCurrencyDaemon(b, followProductIDs...)
	// !Daemon

	b.Handle("/hello", func(m *tb.Message) {
		b.Reply(m, "Hello @"+m.Sender.Username)
	})

	b.Handle("/traffic", func(m *tb.Message) {
		if m.Sender.ID != 615491801 {
			b.Send(m.Chat, "you haven't privilege trigger this command")
			return
		}
		b.Send(m.Chat, components.TencentLighthouseTrafficUsageShow())
	})

	b.Handle("/echo", func(m *tb.Message) {
		b.Reply(m, "Your Name is: "+m.Sender.Username+"\nYour ID is: "+strconv.Itoa(m.Sender.ID)+"\nYour Chat ID is: "+strconv.Itoa(int(m.Chat.ID)))
	})

	b.Handle("/if_bot_started", func(m *tb.Message) {
		b.Send(m.Sender, "Your have start the msg bot, congratulations")
	})

	waitMap := make(map[int]string, 1024)
	b.Handle(tb.OnText, func(m *tb.Message) {
		if v, ok := waitMap[m.Sender.ID]; ok {
			switch v {
			case "/test":
				b.Send(m.Sender, fmt.Sprintf("your input is:%v\n", m.Text))
				delete(waitMap, m.Sender.ID)
			case "/clockin":
				b.Send(m.Sender, fmt.Sprintf("your input is:%v\n", m.Text))
				delete(waitMap, m.Sender.ID)
			case "/weather":
				currentWeather := components.ShowCurrentWeather(m.Text)
				b.Reply(m, fmt.Sprintf("%s\n", currentWeather))
				delete(waitMap, m.Sender.ID)
			case "/exchange_rate_query":
				if _, ok := e.Value[m.Text]; !ok {
					b.Reply(m, fmt.Sprintf("%s\n", "can't find it"))
					delete(waitMap, m.Sender.ID)
					return
				}
				b.Reply(m, fmt.Sprintf("%s:  %s %.4f\n", e.ReqDate, m.Text, e.Value[m.Text]))
				delete(waitMap, m.Sender.ID)
			default:
			}
		}
	})

	b.Handle("/clockin", func(m *tb.Message) {
		b.Reply(m, "input your username and password\nexample: 300100:123456")
		waitMap[m.Sender.ID] = "/clockin"
	})

	b.Send(&tb.User{ID: 615491801}, "bot start at: "+time.Now().String())

	b.Handle("/test", func(m *tb.Message) {
		b.Reply(m, "input your content:")
		waitMap[m.Sender.ID] = "/test"
	})

	b.Handle("/weather", func(m *tb.Message) {
		if int64(m.Sender.ID) != m.Chat.ID {
			b.Reply(m, "chat with bot only")
			return
		}
		b.Reply(m, "input the city name you want to know:\nfor example: hangzhou")
		waitMap[m.Sender.ID] = "/weather"
		//b.Reply(m, components.GetCurrentWeather())
	})

	b.Handle("/weather_hangzhou", func(m *tb.Message) {
		b.Reply(m, components.ShowCurrentWeather("hangzhou"))
	})

	b.Handle("/exchange_rate", func(m *tb.Message) {
		if e.Timestamp == 0 || (time.Now().Unix()-e.Timestamp > 3600) {
			var err error
			e, err = components.GetExchangeRateFromUnionPay()
			if err != nil {
				b.Reply(m, err.Error())
				return
			}
		}
		b.Reply(m, fmt.Sprintf("%s: USD/CNY: %.4f\n", e.ReqDate, e.Value["USD-CNY"]))
	})

	b.Handle("/exchange_rate_query", func(m *tb.Message) {
		if int64(m.Sender.ID) != m.Chat.ID {
			b.Reply(m, "chat with bot only")
			return
		}
		if e.Timestamp == 0 || (time.Now().Unix()-e.Timestamp > 3600) {
			var err error
			e, err = components.GetExchangeRateFromUnionPay()
			if err != nil {
				b.Reply(m, err.Error())
				return
			}
		}
		b.Reply(m, "please input search pattern, for example, usd/cny => USD-CNY:")
		waitMap[m.Sender.ID] = "/exchange_rate_query"

	})

	b.Handle("/exchange_rate_freecurrency", func(m *tb.Message) {
		// cache one hour
		if ee.Timestamp == 0 || (time.Now().Unix()-ee.Timestamp > 3600) {
			var err error
			ee, err = components.GetExchangeRateFromFreeCurrencies()
			if err != nil {
				b.Reply(m, err.Error())
				return
			}
		}
		timestamp, err := strconv.ParseFloat(ee.Updated, 64)
		if err != nil {
			b.Reply(m, err)
			return
		}
		tm := time.Unix(int64(timestamp), 0).In(components.Loc)

		result := fmt.Sprintf("data updated time: %s\n1 usd could convert to: \n%s:%.6f, %s:%.6f, %s:%.6f, %s:%.6f, %s:%.6f, %s:%.6f, %s:%.6f, %s:%.6f\n", tm.String(), "CNH", ee.CNH, "CNY", ee.CNY, "AUD", ee.AUD, "CAD", ee.CAD, "CHF", ee.CHF, "EUR", ee.EUR, "GBP", ee.GBP, "JPY", ee.JPY)
		b.Reply(m, result)
	})

	b.Handle("/exchange_rate_oanda", func(m *tb.Message) {
		data, err := components.GetExchangeRateFromOanda("CNY", "USD")
		if err != nil {
			b.Reply(m, err.Error())
			return
		}
		result, err := strconv.ParseFloat(data.Data.BidAskData["ask"], 64)
		if err != nil {
			b.Reply(m, err.Error())
			return
		}
		resultAfterRound := math.Round(result*1e6) / 1e6
		b.Reply(m, fmt.Sprintf("USD/CNY: %v\n", resultAfterRound))
	})

	b.Handle("/crypto_currency_stats", func(m *tb.Message) {
		data, err := components.ProcessSpecificTicker(statsProductIDs...)
		if err != nil {
			b.Reply(m, err.Error())
			return
		}
		b.Reply(m, data)
	})

	b.Handle("/crypto_currency_reports", func(m *tb.Message) {
		s := components.AnalysisSpecificTickers(followProductIDs...)
		b.Reply(m, s)
	})

	b.Start()

}
