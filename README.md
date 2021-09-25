# bot 小助手

> 开发这个小工具的源动力是: "懒" 😁

## Introduction

小工具入口: [Click it](https://t.me/mh5l7760_msg_bot)

群组: [Click it](https://t.me/joinchat/WPfiERfoj6wzMGY5)

## Function

- 天气查询, 输入城市名称(ps: 目前仅支持拼音, 例如: hangzhou, beijing)
- 天气预告, 每天 8:00, 18:00 查询未来 24 小时内的天气, 将有雨的情况推送给用户(支持自定义, 支持多城市观察, 在配置文件里面可以自己修改, 默认: []string{"Hangzhou", "Beijing"}), 每天除[23, 6]之外, 每 5 分钟统计一次最近 6 小时内的天气情况, 如果下雨, 推送消息给用户, 推送之后 1 小时之内不推送 (因为 openWeatherAPI 限制调用次数: 100k/month, 每天大约可以调用 3000 次, 所以最多同时观察 15 个城市)
- 腾讯云 VPS 流量查询
- 打卡(废除)
- 汇率查询(银联数据(时效不好), freecurrency 数据(不太准), oanda 数据(推荐))
- 加密货币当下信息统计(可在配置文件内自定义观测的 statsProductIDs)
- 加密货币报告分析(时间间隔为: 5 分钟, 一小时, 一天, 七天, 主动查询)
- 加密货币波动分析(3% in 5min, 6% in 1hour, 10% in 1day, 20% in 7day, 可在配置文件里面自定义观测的 followProductIDs), 每分钟统计一次, 波动超过设定数值将会推送
- USDT/USD 比值推送, (上限 1.05, 下限 0.95, 可在配置文件里自定义)
- 加密货币波动分析推送方式优化: 前 2 次, 推送间隔 2mins, 之后, 推送间隔 60mins, 最大次数 5 次, 直到被重置, 参数可以在配置文件中自己配置, 推送分为两个阶段, 第一阶段, 推送间隔 x mins, 推送次数为 y, 第二阶段, 推送间隔为 a mins, 推送间隔为 b, 直到重置(重置条件: 一轮检测下来没有触发推送 && 此时的推送次数!=0), 参数可在配置文件中自定义

## Settings

```jsonc
{
  "token": "xxxxx", // bot的token
  "openWeatherAPI": "xxxxx",
  "tencentKeyOne": "xxx", // tencent cloud API, secretID
  "tencentKeyTwo": "xxx", // tencent cloud API, secretKEY
  "compareRange": {
    "max": 1.05, // usdt/usd 波动上限
    "min": 0.95 // usdt/usd 波动下限
  },
  "followProductIDs": ["BTC-USDT", "ETH-USDT", "ALGO-USDT", "SOL-USDT"], // 关注的加密货币交易对
  "statsProductIDs": ["BTC-USDT", "ETH-USDT", "ALGO-USDT", "SOL-USDT"], // 查询的加密货币交易对
  "followCities": ["Hangzhou", "Beijing"], // 天气推送的关注城市
  "intervalOne": 120, // 第一个推送阶段的间隔(秒)
  "intervalTwo": 3600, // 第二个推送阶段的间隔(秒)
  "firstNotifyNum": 2, // 第一个推送阶段的推送次数
  "secondNotifyNum": 5, // 第二个推送阶段的累计推送次数
  "fiveMinutesRange": 0.025, // 五分钟变化范围超过2.5%就会推送
  "oneHourRange": 0.05, // 一小时变化范围超过5%就会推送
  "oneDayRange": 0.1, // 一天变化范围超过10%就会推送
  "oneWeekRange": 0.2, // 一周变化范围超过20%就会推送
  "sendToID": 615491801 // 消息推送receiver的ID, 比如: 你的用户ID, 你的群组ID
}
```

## 使用方法

1. 自己编译 docker image

```bash
# 默认你已安装docker
mv config.example.json bot-config.json
user=test # 自己的用户名, 例如: test, 如果你打算推送到自己的docke-hub, 那就写docker-hub的名字, 否则随便填
version=0.0 # 版本号, 随便填, 例如: 0.0
docker build -t $(user)/botmsg:$(version)
docker run -d --name=botmsg $(user)/botmsg:$(version)
# 查看执行情况
docker logs -f --tail=200 botmsg
```
