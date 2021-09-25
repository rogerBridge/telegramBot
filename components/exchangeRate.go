package components

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

type UnionPay struct {
	ExchangeRateId            int `json:"exchangeRateId"`
	CurDate                   int
	BaseCurrency              string
	TransactionCurrency       string
	ExchangeRate              float64
	CreateDate                int
	CreateUser                int
	UpdateDate                int
	UpdateUser                int
	EffecTiveDate             int
	TransactionCurrencyOption interface{}
}

var client = FastHttpClient

type ExchangeRateJson struct {
	BaseCur  string
	TransCur string
	RateData float64
}

type ExchangeRateResponse struct {
	ExchangeRateJson []*ExchangeRateJson
	CurDate          string
}

type ExchangeRateCache struct {
	Value     map[string]float64
	Timestamp int64
	ReqDate   string
}

func GetExchangeRateFromUnionPay() (*ExchangeRateCache, error) {
	e := new(ExchangeRateCache)
	e.Value = make(map[string]float64)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	// get closest work day from today
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Println(err)
		return e, err
	}
	now := time.Now().In(loc)
	nowWeekday := now.Weekday()
	if nowWeekday.String() == "Sunday" {
		now = now.Add(-2 * 1e9 * 24 * 3600)
	} else if nowWeekday.String() == "Saturday" {
		now = now.Add(-1 * 1e9 * 24 * 3600)
	}
	log.Println("After correct: ", now)
	fileName := "/" + now.Format("20060102") + ".json"

	req.Header.SetMethod(fasthttp.MethodGet)
	u := url.URL{
		Scheme: "https",
		Host:   "www.unionpayintl.com",
		Path:   "upload/jfimg" + fileName,
	}
	log.Println(u.String())
	req.SetRequestURI(u.String())
	req.Header.Set("Host", "www.unionpayintl.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Referer", "https://www.unionpayintl.com/cardholderServ/serviceCenter/rate?language=en")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.unionpayintl.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = client.Do(req, res)
	if err != nil {
		log.Println(err)
		return e, err
	}

	value := new(ExchangeRateResponse)
	err = json.Unmarshal(res.Body(), value)
	if err != nil {
		return e, err
	}
	// convert value to map[string][string]float64
	for _, v := range value.ExchangeRateJson {
		e.Value[v.TransCur+"-"+v.BaseCur] = v.RateData
	}
	e.ReqDate = value.CurDate
	e.Timestamp = time.Now().Unix()
	log.Printf("%+v", e)
	return e, nil
}

type ExchangeMajor struct {
	Updated   string  `json:"updated"`
	CNY       float64 `json:"CNY"`
	CNH       float64 `json:"CNH"`
	EUR       float64 `json:"EUR"`
	GBP       float64 `json:"GBP"`
	JPY       float64 `json:"JPY"`
	CHF       float64 `json:"CHF"`
	CAD       float64 `json:"CAD"`
	AUD       float64 `json:"AUD"`
	Timestamp int64   `json:"reqDate"`
}

func GetExchangeRateFromFreeCurrencies() (*ExchangeMajor, error) {
	client := FastHttpClient

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(fasthttp.MethodGet)

	host := "freecurrencyrates.com"
	u := &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/api/action.php",
	}
	q := make(url.Values)
	q.Set("s", "fcr")
	q.Set("iso", "CNY-CNH-EUR-GBP-JPY-CHF-CAD-AUD")
	q.Set("f", "USD")
	q.Set("v", "1")
	q.Set("do", "cvals")
	q.Set("ln", "en")
	u.RawQuery = q.Encode()

	req.Header.Set("Host", host)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-GPC", "1")
	log.Println("request uri: ", u.String())
	req.SetRequestURI(u.String())

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	result := new(ExchangeMajor)

	err := client.Do(req, res)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(res.Body(), result)
	if err != nil {
		return result, err
	}
	result.Timestamp = time.Now().Unix()
	log.Printf("%+v\n", result)
	return result, nil
}

type OandaExchangeRateData struct {
	BidAskData map[string]string `json:"bid_ask_data"`
}

type OandaExchangeRateResponse struct {
	ID   string                `json:"id"`
	Data OandaExchangeRateData `json:"data"`
}

func GetExchangeRateFromOanda(from string, to string) (*OandaExchangeRateResponse, error) {
	log.Println("get exchange rate from onada...")
	client := FastHttpClient

	// req
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("Host", "www1.oanda.com")
	req.Header.Set("User-Agent", " Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*;")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://www1.oanda.com/currency/converter/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("TE", "trailers")
	req.Header.Set("X-Prototype-Version", "1.7")
	req.Header.Set("Cookie", "mru_base0=CNH%2CCNY%2CEUR%2CUSD%2CGBP; mru_quote=EUR%2CUSD%2CGBP%2CCAD%2CAUD; base_currency_0=CNY; quote_currency=USD; ncc_session=e47b7341d892eea2793b12be5618b383d161a161; __cf_bm=L2wCVe3Kcmb5s7MxiNyrsMmHf8VWexyKg7DVx1GIYBs-1631061083-0-AbGrBzG3swnJFP7dzFihJN7u/Xa9BZXt5ucoc8mMlBd18w2SNHt/RdFzBiFKCE3KRZ9RKNFScisVJzaGyWmxUe8jbNhAJu4BrruF8MBsK7wW; __cfruid=968506895547dc887ec9e6f5198f4b2ff26398fe-1631061083; tc=1")
	req.Header.Set("X-requested-With", "XMLHttpRequest")
	u := url.URL{
		Scheme: "https",
		Host:   "www1.oanda.com",
		Path:   "/currency/converter/update",
	}
	q := u.Query()
	today := time.Now().Format("2006-01-02")
	q.Set("base_currency_0", from)
	q.Set("quote_currency", to)
	q.Set("end_date", today)
	q.Set("view", "details")
	q.Set("id", "1")
	q.Set("action", "C")
	u.RawQuery = q.Encode()
	log.Println("request uri is: ", u.String())
	req.SetRequestURI(u.String())
	// res
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := client.Do(req, res)
	if err != nil {
		log.Println("client do req, error happen: ", err)
		return nil, err
	}
	// process res data
	// ungzip response data
	rawData, err := res.BodyGunzip()
	if err != nil {
		return nil, err
	}
	data := new(OandaExchangeRateResponse)
	err = json.Unmarshal(rawData, data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Printf("%+v\n", data)
	return data, nil
}

func GetUSDTotcTicker() (float64, error) {
	log.Println("get usdt otc ticker ...")
	client := FastHttpClient

	// req
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("Host", "www.okex.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("App-Type", "web")
	req.Header.Set("devId", "fa3f1f4c-3721-41f5-aa74-47aae7dfe9f3")
	req.Header.Set("ftID", "5210057794189.00115423a51eb1c0776ce21fdd6a35e20f581.1000L8o0.ADE9A6395767C058")
	req.Header.Set("x-utc", "8")
	req.Header.Set("x-cdn", "https://static.okex.com")
	req.Header.Set("Cookie", "__cfduid=daee173d38f9b0c1b04794d73e8ff839c1619704243; first_ref=https%3A%2F%2Fwww.okex.com%2F; u_ip=My4xMTMuMjIyLjE0MQ; G_ENABLED_IDPS=google; locale=zh_CN; PVTAG=274.408.Zdd5LflK10Mc5J01t97yuGXzr9ml6D6DQUz44G1a; u_pid=D6D6lm9rzXGuy79t")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("TE", "Trailers")
	u := url.URL{
		Scheme: "https",
		Host:   "www.okex.com",
		Path:   "/v3/c2c/otc-ticker",
	}
	q := u.Query()
	u.RawQuery = q.Encode()
	q.Set("t", strconv.Itoa(int(time.Now().UnixNano()/1e6)))
	q.Set("baseCurrency", "usdt")
	q.Set("quoteCurrency", "cny")
	u.RawQuery = q.Encode()
	req.SetRequestURI(u.String())

	log.Println("request uri is: ", u.String())

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := client.Do(req, res)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	// process res bytes
	type Data struct {
		OtcTicker string `json:"otcTicker"`
	}
	type Response struct {
		Code int  `json:"code"`
		Data Data `json:"data"`
	}
	d := new(Response)
	err = json.Unmarshal(res.Body(), d)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	// log.Println(d.Data.OtcTicker)
	result, err := strconv.ParseFloat(d.Data.OtcTicker, 64)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return result, nil
}

func ExchangeBetweenUSDandUSDT() (string, float64, error) {
	result := ""
	obj, err := GetExchangeRateFromOanda("CNY", "USD")
	if err != nil {
		return "", 0, err
	}
	usdString := obj.Data.BidAskData["ask"]
	usd, err := strconv.ParseFloat(usdString, 64)
	if err != nil {
		return "", 0, err
	}
	usdt, err := GetUSDTotcTicker()
	if err != nil {
		log.Println(err)
		return "", 0, err
	}
	comparison := usdt / usd
	result += fmt.Sprintf("usdt: %.3f, usd: %.3f, usdt/usd: %.4f\n", usdt, usd, comparison)
	log.Println("comparison result: ", result)
	return result, comparison, err
}

func ExchangeDaemon(b *tb.Bot) {
	myGroup := &tb.User{ID: Config.SendToID}
	for {
		now := time.Now().In(Loc)
		content, comparison, err := ExchangeBetweenUSDandUSDT()
		// push every day
		switch now.Hour() {
		case 8, 18:
			if now.Minute() == 0 {
				if err != nil {
					b.Send(myGroup, err.Error())
				} else {
					b.Send(myGroup, content)
				}
			}
		}
		// good condition
		if err != nil {
			b.Send(myGroup, err.Error())
		} else {
			if comparison <= Config.CompareRange.Min || comparison >= Config.CompareRange.Max {
				b.Send(myGroup, content)
				time.Sleep(time.Hour)
			}
		}
		time.Sleep(time.Minute)
	}
}
