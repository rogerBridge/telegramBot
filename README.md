# bot小助手

> 开发这个小工具的源动力是: "懒" 😁

## Introduction
  小工具入口: [Click it](https://t.me/mh5l7760_msg_bot)

  群组: [Click it](https://t.me/joinchat/WPfiERfoj6wzMGY5)

## Function
- 天气查询, 输入城市名称(ps: 目前仅支持拼音, 例如: hangzhou, beijing)
- 天气预告, 每天8:00, 18:00 查询未来24小时内的天气, 将有雨的情况推送给用户(支持自定义, 在配置文件里面可以自己修改, 默认: 杭州市), 每天除[23, 6]之外, 每小时统计一次最近三小时内的天气情况, 如果下雨, 推送消息给用户
- 腾讯云VPS流量查询
- 打卡(废除)
- 汇率查询(银联数据(时效不好), freecurrency数据(不太准), oanda数据(推荐))
- 加密货币当下信息统计(可在配置文件内自定义观测的ProductID)
- 加密货币报告分析(5分钟, 一小时, 一天, 七天)
- USDT/USD 比值推送, (大于1.05 或 小于0.95, 可在配置文件里自定义)
- 加密货币波动分析(2.5% in 5min, 5% in 1hour, 10% in 1day, 20% in 7day, 可在配置文件里面自定义观测的ProductID), 每分钟统计一次, 波动超过设定数值将会推送
- continue