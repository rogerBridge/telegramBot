package components

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	_ "time/tzdata" // find tzdata even not found in system

	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Config, errLoadConfig = ReadConfigFromJson("./bot-config.json")

var FastHttpClient = &fasthttp.Client{
	MaxConnsPerHost: 40960,
	Dial: func(addr string) (conn net.Conn, err error) {
		//return connLocal, err
		return fasthttp.DialTimeout(addr, 32*time.Second) // tcp 层
	},
}

var Loc, errLoadTimezoneLoc = time.LoadLocation("Asia/Shanghai")

func init() {
	if errLoadConfig != nil {
		log.Fatalln("load config error: ", errLoadConfig)
	}
	if errLoadTimezoneLoc != nil {
		log.Fatalln("load timezone local error: ", errLoadTimezoneLoc)
	}
	if errConnectToSqlite != nil {
		log.Fatalf("While connect to sqlite, error: %s", errConnectToSqlite)
	}
}

type BotConfig struct {
	Token            string       `json:"token"`
	OpenWeatherAPI   string       `json:"openWeatherAPI"`
	TencentKeyOne    string       `json:"tencentKeyOne"`
	TencentKeyTwo    string       `json:"tencentKeyTwo"`
	CompareRange     CompareRange `json:"compareRange"`
	FollowProductIDs []string     `json:"followProductIDs"`
}

type CompareRange struct {
	Max float64 `json:"max"`
	Min float64 `json:"min"`
}

func ReadConfigFromJson(path string) (*BotConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	b := new(BotConfig)
	err = json.Unmarshal(bytes, b)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return b, nil
}

func NewBot() *tb.Bot {
	poller := &tb.LongPoller{Timeout: 15 * time.Second}

	token := Config.Token

	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return b
}

var SqliteConn, errConnectToSqlite = gorm.Open(sqlite.Open("./data/weather.db"), &gorm.Config{})

// refresh tokens get from users
func InitSqlite() {
	err := SqliteConn.AutoMigrate(&GeoCode{})
	if err != nil {
		log.Fatalf("While Migrate sqlite, error: %s", err)
	}

	err = SqliteConn.AutoMigrate(&Ticker{})
	if err != nil {
		log.Fatalf("While Migrate sqlite, error: %s", err)
	}
}
