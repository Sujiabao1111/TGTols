# M7 Game API 接入文档

> 来源文件：`M7 API 接入文档1.1.4.docx`、`Retou-測試.pdf`
> 整理日期：2026-06-23
> 说明：测试环境资料中的密钥、密码已脱敏，完整凭证请以原始 PDF 或安全凭据库为准。

## 目录

- [版本记录](#版本记录)
- [接入概览](#接入概览)
- [API 请求限制](#api-请求限制)
- [转账钱包游戏 API](#转账钱包游戏-api)
- [单一钱包游戏 API](#单一钱包游戏-api)
- [附录](#附录)
- [测试环境资料](#测试环境资料)

## 版本记录

| 日期 | 版本 | 说明 |
| --- | --- | --- |
| 2025/05/16 | v1.0.0 | 初始版本 |
| 2025/08/06 | v1.0.1 | 新增玩家未完成回合 API |
| 2025/12/31 | v1.1.2 | 新增 `currency` 参数 |
| 2025/12/31 | v1.1.3 | `gameList` QTech 回传游戏品牌 |
| 2025/01/06 | v1.1.4 | 回调 / `transaction` 增加 `orderId` |
| 2026/05/06 | v1.1.5 | 新增试玩游戏 API |

## 接入概览

M7 游戏接入包含两种钱包模式：

| 模式 | 说明 | 主要调用方 |
| --- | --- | --- |
| 转账钱包 | 平台先创建玩家，再通过上下分接口把余额转入/转出 M7 钱包。 | 我方主动调用 M7 API |
| 单一钱包 | M7 在游戏过程中回调我方余额和加扣款接口，我方系统作为钱包余额来源。 | M7 回调我方 API |

### 流程图

#### 建立使用者流程

![建立使用者流程](assets/m7-api/image1.png)

#### 查询余额

![查询余额](assets/m7-api/image2.png)

#### 使用者上下分流程

![使用者上下分流程 1](assets/m7-api/image3.png)

![使用者上下分流程 2](assets/m7-api/image4.png)

## API 请求限制

| 方法 | 限制秒数 | 限制对象 |
| --- | ---: | --- |
| `entryGame` | 5 | Player |
| `deposit` | 1 | Player |
| `withdraw` | 3 | Player |
| `getGameHistory` | 5 | Agent |
| `getGameList` | 3 | Agent |

## 转账钱包游戏 API

### 1. 建立使用者

`POST /busway/external/api/createPlayer`

#### 请求参数

```json
{
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "agentId": "agt011",
  "userId": "player005",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `agentId` | 是 | 一级代理名称 |
| `userId` | 是 | 会员名称，合法字符为 `0-9`、`a-z`，最长 40 字符 |
| `currency` | 否 | 指定币别，可共用同一组 `clientId` / `clientKey` |

#### 响应示例

```json
{
  "result": {
    "data": {
      "result": "success"
    }
  },
  "code": "0000"
}
```

### 2. 游戏列表

`POST /busway/external/api/getGameList`

#### 请求参数

```json
{
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "thirdPartyType": "PP_SLOT"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `thirdPartyType` | 否 | 游戏商代号；传入时只返回该游戏商列表，不传返回全部 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int | 游戏 ID 唯一值 |
| `thirdPartyType` | string | 游戏商代号 |
| `thirdParty` | string | 游戏商 |
| `gameNameEn` | string | 游戏名称，英文 |
| `gameNameCh` | string | 游戏名称，简中 |
| `gameNameTw` | string | 游戏名称，繁中 |
| `gameNo` | string | 游戏商游戏代号 |
| `image1` | string | 游戏图 1 |
| `image2` | string | 游戏图 2 |
| `status` | string | 游戏状态：`OPEN` 开启，`CLOSE` 关闭 |
| `provideId` | string | 游戏品牌 ID |
| `provideName` | string | 游戏品牌名称 |
| `isDemo` | string | 是否可以试玩：`Y` 是，`N` 否 |

#### 响应示例

```json
{
  "result": {
    "data": [
      {
        "id": 237863,
        "thirdPartyType": "PP_SLOT",
        "thirdParty": "PP",
        "gameNameEn": "Pyramid Bonanza",
        "gameNameCh": "疯狂金字塔",
        "gameNameTw": "瘋狂金字塔",
        "gameNo": "vs20pbonanza",
        "image1": "https://api.prerelease-env.biz/game_pic/square/200/vs20pbonanza.png",
        "image2": "https://api.prerelease-env.biz/game_pic/square/200/vs20pbonanza.png",
        "status": "OPEN",
        "provideId": "",
        "provideName": "",
        "isDemo": ""
      }
    ]
  },
  "code": "0000"
}
```

### 3. 取得游戏链接

`POST /busway/external/api/entryGame`

#### 请求参数

```json
{
  "lang": "en",
  "gameId": "10001",
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "userId": "player005",
  "device": "1",
  "returnUrl": "https://xxxx.com",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `lang` | 否 | 多语言：`zh`、`en`、`th`、`pt`，默认 `en` |
| `gameId` | 是 | 游戏代码 |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `userId` | 是 | 会员名称 |
| `device` | 是 | 装置：`1` mobile，`2` web |
| `returnUrl` | 是 | 返回首页的 URL |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `url` | string | 游戏链接 |

### 4. 钱包余额查询

`POST /busway/external/wallet/balance`

#### 请求参数

```json
{
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "userId": "player005",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `userId` | 是 | 会员 ID |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `userId` | string | 会员账号 |
| `currency` | string | 币别 |
| `balance` | float | 钱包余额 |

### 5. 会员上下分

`POST /busway/external/wallet/transfer`

#### 请求参数

```json
{
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "userId": "player005",
  "amount": 200,
  "txId": "9aca5f03f2daf6baa2da85",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `userId` | 是 | 会员名称 |
| `amount` | 是 | 上下分金额；正数代表上分，负数代表下分 |
| `txId` | 是 | 平台代理商自行生成的交易代号，需保证唯一 |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `userId` | string | 会员账号 |
| `currency` | string | 币别 |
| `balance` | float | 钱包余额 |
| `amount` | float | 加/减金额 |
| `txId` | string | 交易代号 |

### 6. 上下分记录查询

`POST /busway/external/wallet/transferHistory`

#### 请求参数

```json
{
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "txId": "9aca5f03f2daf6baa2da85",
  "startTime": "2022-01-01 00:00:00",
  "endTime": "2022-01-01 00:00:00",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `txId` | 否 | 平台代理商自行生成的交易代号 |
| `startTime` | 否 | 开始时间；默认近七天，时间区间不能超过 7 天 |
| `endTime` | 否 | 结束时间；默认近七天，时间区间不能超过 7 天 |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `userId` | string | 会员账号 |
| `beforeBalance` | float | 交易前金额 |
| `afterBalance` | float | 交易后金额 |
| `amount` | float | 交易金额 |
| `txId` | string | 交易代号 |
| `transferType` | string | 交易类别：`1004` 会员下分，`1005` 会员上分 |
| `createTime` | dateTime | 建立时间 |

### 7. 注单查询

`POST /busway/external/api/getGameHistory`

#### 请求参数

```json
{
  "clientId": "KwmCPjPsmVDsSs0",
  "clientKey": "moBlpx1ViFuUsmdQVwvgEBEk",
  "agentId": "agt011",
  "userId": "arstest",
  "thirdPartyType": "DG_SLOT",
  "startTime": "2024-10-22 16:45:00",
  "endTime": "2024-10-22 16:49:59",
  "pageIndex": 1,
  "pageSize": 500,
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `agentId` | 是 | 一级代理名称 |
| `userId` | 否 | 会员 ID |
| `thirdPartyType` | 否 | 游戏商代号 |
| `startTime` | 是 | 查询起始时间 |
| `endTime` | 是 | 查询结束时间 |
| `pageIndex` | 是 | 分页页数，初始值 `1` |
| `pageSize` | 是 | 分页大小，上限 `5000`，初始值 `1000` |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `gameId` | string | 游戏商游戏代码 |
| `gamesId` | int | 游戏 ID |
| `settleTime` | dateTime | 结算日期 |
| `thirdPartyType` | string | 游戏代码 |
| `userViewId` | string | 账号 |
| `txId` | string | 注单 txid |
| `betIp` | string | 下注 IP |
| `userId` | string | 完整账号 |
| `gameRound` | string | 游戏局号 |
| `thirdParty` | string | 游戏商 |
| `validBetAmount` | float | 有效下注金额 |
| `betAmount` | float | 下注金额 |
| `gameName` | string | 游戏名称 |
| `winLose` | float | 游戏输赢 |
| `betTime` | dateTime | 下注时间 |
| `win` | float | 赢分 |
| `status` | int | 状态：`0` 未结算，`1` 已结算，`2` 取消 |
| `recordResult` | string | 原始游戏结果 |
| `recordContent` | string | 原始投注内容 |

### 8. 注单详情查询

`POST /busway/external/api/getGameHistoryDetail`

#### 请求参数

```json
{
  "clientId": "KwmCPjPsmVDsSs0",
  "clientKey": "moBlpx1ViFuUsmdQVwvgEBEk",
  "userId": "arstest",
  "thirdPartyType": "DG_SLOT",
  "txId": "YZF8PShSUALV9PhPbFcP6aWgv4utavLR",
  "gameId": 53687230,
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `userId` | 是 | 会员 ID |
| `thirdPartyType` | 是 | 游戏商代号 |
| `txId` | 是 | 注单 ID |
| `gameId` | 是 | 游戏列表 ID |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `url` | string | 注单详情链接 |

### 9. 玩家未完成的回合

`POST /busway/external/api/getIncompleteGames`

#### 请求参数

```json
{
  "clientId": "KwmCPjPsmVDsSs0",
  "clientKey": "moBlpx1ViFuUsmdQVwvgEBEk",
  "userId": "arstest",
  "thirdParty": "PP",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `userId` | 是 | 会员 ID |
| `thirdParty` | 是 | 游戏商 |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `gameId` | string | 游戏商游戏代码 |
| `gamesId` | int | 游戏 ID |
| `playSessionID` | string | 玩家特定游戏会话 ID，即游戏回合唯一编号 |
| `betAmount` | string | 下注金额 |

### 10. 取得试玩游戏链接

`POST /busway/external/api/tryPlayerGame`

#### 请求参数

```json
{
  "lang": "en",
  "gameId": "10001",
  "clientId": "kCa7zly7RppnGMk",
  "clientKey": "00ysh375628djsixjyew00",
  "device": "1",
  "returnUrl": "https://xxxx.com",
  "currency": "CNY"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `lang` | 否 | 多语言：`zh`、`en`、`th`、`pt`，默认 `en` |
| `gameId` | 是 | 游戏代码 |
| `clientId` | 是 | 后台建立的一级代理 Client ID |
| `clientKey` | 是 | 后台建立的一级代理密钥 |
| `device` | 是 | 装置：`1` mobile，`2` web |
| `returnUrl` | 是 | 返回首页的 URL |
| `currency` | 否 | 指定币别 |

#### 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `url` | string | 试玩游戏链接 |

## 单一钱包游戏 API

单一钱包接口由运营商提供，全部请求使用：

```http
Content-Type: application/json
```

### 1. 获取用户余额

`POST {domain}/seamless/api/balance`

#### 请求参数

```json
{
  "requestId": "b5572c05977b459ca910f7d99109dab5",
  "userId": "player005",
  "ts": 1764572087806,
  "signature": "kCa7zly7RppnGMk"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `requestId` | 是 | 请求 ID |
| `userId` | 是 | 会员名称，合法字符为 `0-9`、`a-z`，最长 40 字符 |
| `ts` | 是 | 毫秒时间戳 |
| `signature` | 是 | 签名 |

#### 签名规则

```text
method = "balance"
signature = MD5(MD5(method + userId + ts) + secretKey)
```

#### 成功响应

```json
{
  "data": {
    "balance": 100.00
  },
  "code": "0000"
}
```

余额为负数时应返回 `0.00`：

```json
{
  "data": {
    "balance": 0.00
  },
  "code": "0000"
}
```

### 2. 加/扣款

`POST {domain}/seamless/api/transaction`

#### 请求参数

```json
{
  "requestId": "b5572c05977b459ca910f7d99109dab5",
  "userId": "player005",
  "amount": "-3.0000",
  "ts": 1764572087806,
  "isOwe": 2,
  "gameId": "70105943",
  "source": "PP",
  "gameNo": "22",
  "orderId": "m-22008060314579738625",
  "signature": "kCa7zly7RppnGMk"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `requestId` | 是 | 请求 ID |
| `userId` | 是 | 会员名称，合法字符为 `0-9`、`a-z`，最长 40 字符 |
| `amount` | 是 | 字符串类型，四位小数；正数为游戏转回钱包金额，负数为转出到游戏金额 |
| `ts` | 是 | 毫秒时间戳 |
| `isOwe` | 是 | `2` 强制扣款，即使扣款后为负，也返回扣款后余额为 0；`0` 余额不足则不扣款并返回 `2201` |
| `gameId` | 是 | 游戏 ID |
| `source` | 是 | 游戏商 |
| `gameNo` | 否 | 游戏代码；为空时不传该参数，签名也不包含该参数 |
| `orderId` | 否 | 游戏投注订单号；非下注操作为空，签名也不包含该参数 |
| `signature` | 是 | 签名 |

#### 签名规则

```text
method = "transaction"
signature = MD5(MD5(method + userId + amount + isOwe + gameId + source + (gameNo) + (orderId) + ts) + secretKey)
```

`gameNo`、`orderId` 为空时，不参与签名。

#### 成功响应

```json
{
  "data": {
    "beforBalance": 100.00,
    "afterBalance": 103.00
  },
  "code": "0000"
}
```

#### 余额不足响应

```json
{
  "message": "余额不足",
  "code": "2201"
}
```

## 附录

### 游戏代码 `thirdPartyType`

| 游戏代号 | 游戏商 |
| --- | --- |
| `LEG_POKER` | 乐游棋牌 |
| `KY_POKER` | 开元棋牌 |
| `RCB_SPORT` | RCB 赛马 |
| `BOLE_POKER` | 博乐棋牌 |
| `JDB_SLOT` | JDB 电子 |
| `RICH_SLOT` | Rich88 |
| `SEXY_LIVE` | Sexy 性感真人 |
| `PP_SLOT` | PP 电子 |
| `PP_LIVE` | PP 真人 |
| `ONE_FC_FISH` | FC 捕鱼 |
| `ONE_FC_SLOT` | FC 电子 |
| `ONE_JILI_SLOT` | JILI 电子 |
| `ONE_JILI_FISH` | JILI 捕鱼 |
| `BNG_SLOT` | BNG 电子 |
| `RSG_SLOT` | RSG 电子 |
| `RSG_LIVE` | RG 真人 |
| `CG_SLOT` | CG 电子 |
| `QTECH_SLOT` | QTech 电子 |
| `BE_SLOT` | BE 电子，画面上请另外命名 |
| `MT_LIVE` | MT 真人 |
| `DIGMAAN_SPORT` | Digmaan 拳王斗鸡 |

### 错误码

| 错误码 | 说明 |
| --- | --- |
| `0000` | 成功 |
| `1001` | 此用户不存在 |
| `1049` | 请求频率过高 |
| `10197` | 游戏不存在 |
| `1053` | `txid` 重复 |
| `1048` | 非代理账号 |
| `2100` | 验证失败 |
| `1017` | 账号已被冻结 |
| `3211` | 找不到游戏开关 |
| `3212` | 游戏正在维护中 |

## 测试环境资料

来自 `Retou-測試.pdf`。

| 项目 | 值 |
| --- | --- |
| 商户 ID | `PPNET` |
| 环境 | 转账钱包测试 |
| 币别 | `USD` |
| API URL | `https://api.ddhh.work/` |
| 后台 | `https://oms.ddhh.work/login` |
| clientId | `U2Va8Fv3lln3o91` |
| clientKey | 已脱敏，见原始 PDF |
| 代理账号 | `richag01` |
| 代理密码 | 已脱敏，见原始 PDF |
| 总代账号 | `richCasino` |
| 总代密码 | 已脱敏，见原始 PDF |

### 测试品牌列表

| 品牌 | 币别 | 游戏商代号 |
| --- | --- | --- |
| JDB | USD | `JDB_SLOT` |
| PP | USD | `PP_SLOT` |
| PG | USD | `PG_SLOT` |
| Playtech | USD | `ONE_PT_SLOT` |
| Microgaming+ | USD | `ONE_MG_SLOT` |
| Hacksaw | USD | `DNG_HS_SLOT` |
| Slotmill | USD | `DNG_STM_SLOT` |
| PP 真人 | USD | `PP_LIVE` |
| Playtech 真人 | USD | `ONE_PT_LIVE` |
| Evolution | USD | `EVO_LIVE` |
