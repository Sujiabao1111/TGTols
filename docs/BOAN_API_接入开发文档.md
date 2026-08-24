# BOAN API 接入开发文档（Retou 测试环境）

> 文档用途：指导 Retou 团队完成 BOAN 游戏聚合平台的测试环境接入。  
> 整理日期：2026-07-31  
> 依据：`BOAN接口文档-ZH.pdf` 及 BOAN 提供的测试环境资料。  
> 注意：本文包含测试环境密钥和后台密码，只允许在受控团队内部使用，禁止提交到公开仓库、前端代码、日志或工单截图。正式上线前必须更换全部凭证。

## 1. 接入模式

BOAN 支持两类钱包模式：

1. **无缝钱包（Single Wallet）**：玩家余额保存在我方系统。BOAN 在下注、派彩、回滚等场景调用我方 `/wallet/*` 回调。
2. **转账钱包（Transfer Wallet / OneAPI）**：通过 `/cash/deposit`、`/cash/withdraw` 在我方钱包与游戏钱包之间转账。

无论采用哪种模式，以下能力均由 BOAN 提供：

- 获取厂商和游戏列表；
- 获取游戏启动 URL；
- 终止玩家游戏会话；
- 查询注单列表和注单详情。

应在联调前向 BOAN 确认 Retou 实际启用的模式、币种、厂商和游戏产品。不要同时实现两种资金模式并在业务中混用。

## 2. 测试环境资料

### 2.1 代理后台

| 项目 | 值 |
| --- | --- |
| 总代理名称 | `Retou` |
| 代理后台地址 | `https://test-web-agent.boan.games` |
| 总代理登录码 | `PsDqTw` |
| 总代理账号 | `Retou_admin` |
| 总代理密码 | `Retou_admin` |
| 子代理登录码 | `51iOvu` |
| 子代理账号 | `RetouSTG_admin` |
| 子代理密码 | `RetouSTG_admin` |

### 2.2 API 凭证

| 项目 | 值 |
| --- | --- |
| API Base URL | `https://test-gw.boan.games` |
| `X-API-Key` | `ZvWiNki4WY4iVnzH6M6wNpkvHRGIiBm10MiNNsgWaHWBErHe6LLXbihF5JeIfc3z` |
| `X-API-Secret` | `mDKnqUxaQefehTqDRbFnILY6dARqGR3nZ5IAce9uRfrzPNqHR75qlbar6Cf44jTw` |
| 我方服务出口 IP | `47.83.239.79`、`8.218.128.50` |

### 2.3 推荐环境变量

```dotenv
BOAN_BASE_URL=https://test-gw.boan.games
BOAN_API_KEY=ZvWiNki4WY4iVnzH6M6wNpkvHRGIiBm10MiNNsgWaHWBErHe6LLXbihF5JeIfc3z
BOAN_API_SECRET=mDKnqUxaQefehTqDRbFnILY6dARqGR3nZ5IAce9uRfrzPNqHR75qlbar6Cf44jTw
BOAN_AGENT_URL=https://test-web-agent.boan.games
BOAN_HTTP_TIMEOUT_MS=10000
```

凭证必须只保存在服务端密钥管理或环境变量中。`X-API-Secret` 不得发送给浏览器或移动端。

## 3. 网络与安全要求

- 全部通信必须使用 HTTPS。
- 请 BOAN 将 `47.83.239.79`、`8.218.128.50` 加入测试环境 API 白名单。
- 若使用无缝钱包，我方回调服务还需允许 BOAN 的来源 IP；具体 IP 段应向 BOAN 技术支持索取。
- 回调地址必须使用公网 HTTPS，证书链完整，不得依赖内网 DNS。
- API 日志需要脱敏：不得记录 API Secret、完整签名、后台密码和玩家敏感信息。
- 金额使用定点十进制类型（如 `DECIMAL`），禁止使用二进制浮点数进行账务计算。
- 所有时间字段按 Unix 时间戳毫秒处理。

## 4. 请求认证与签名

调用 BOAN API 时统一使用：

```http
Content-Type: application/json
X-API-Key: <BOAN_API_KEY>
X-Signature: <HMAC-SHA256_HEX>
```

签名算法：

```text
X-Signature = hex_lowercase(HMAC-SHA256(API_SECRET, raw_request_body))
```

关键规则：

- 签名对象是实际发送的完整原始 JSON 请求体，不是排序后的参数，也不是 Base64 字符串。
- JSON 序列化后不得在签名和发送之间再次修改空格、换行、字段顺序或数字格式。
- 输出为 64 位小写十六进制字符串。
- 接收 BOAN 钱包回调时，我方使用同一 Secret 对收到的原始请求体重新计算签名，并使用恒定时间比较。
- PDF 的钱包回调示例仅列出 `X-Signature`。联调时应向 BOAN 确认回调是否也携带 `X-API-Key`；若携带则一并校验。

### 4.1 Node.js 签名示例

```js
import crypto from "node:crypto";

export function signBoan(rawBody, apiSecret) {
  return crypto
    .createHmac("sha256", apiSecret)
    .update(rawBody, "utf8")
    .digest("hex");
}

const body = {
  traceId: crypto.randomUUID(),
  displayLanguage: "zh",
  currency: "CNY"
};
const rawBody = JSON.stringify(body);
const signature = signBoan(rawBody, process.env.BOAN_API_SECRET);

const response = await fetch(`${process.env.BOAN_BASE_URL}/game/vendors`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": process.env.BOAN_API_KEY,
    "X-Signature": signature
  },
  body: rawBody
});
```

### 4.2 cURL 签名示例

```bash
BODY='{"traceId":"f8c3de3d-1fea-4d7c-a8b0-29f63c4c3455","displayLanguage":"zh","currency":"CNY"}'
SIGNATURE=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$BOAN_API_SECRET" -hex | awk '{print $2}')

curl -X POST "$BOAN_BASE_URL/game/vendors" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $BOAN_API_KEY" \
  -H "X-Signature: $SIGNATURE" \
  --data-raw "$BODY"
```

## 5. 通用约定

### 5.1 响应结构

BOAN 业务结果主要通过响应体中的 `status` 表示。即使业务失败，HTTP 状态也可能仍为 `200 OK`，因此不能只判断 HTTP 状态码。

```json
{
  "traceId": "e62ceec8-1c18-48fc-af0e-9f9f943b2554",
  "status": "SC_INVALID_REQUEST",
  "message": "请求体中发送了错误/缺失的参数。",
  "validation": {
    "platform": "不支持该平台。"
  }
}
```

处理顺序：检查网络/HTTP 状态，解析 JSON，校验 `traceId`，最后根据 `status` 判断业务成功或失败。仅 `SC_OK` 视为成功。

### 5.2 标识与幂等

- `traceId`：每次 API 请求生成新的 UUID，用于全链路排查；重试策略按具体接口要求执行。
- `transactionId`：钱包交易的幂等键。同一 ID 重复到达时，不得重复增减余额，应返回首次成功处理后的当前余额。
- `referenceId`：转账钱包中由我方生成的唯一请求号。发起新转账必须使用新值；查询和指定错误重试必须复用原值。
- `betId`：下注标识。一个 `betId` 可能对应多次派彩，不能把它作为派彩接口的唯一幂等键。
- `roundId`：游戏轮次标识，仅用于归组，不适合作为交易唯一键。

建议钱包流水唯一索引：`(api_name, transaction_id)`；转账流水唯一索引：`reference_id`。

## 6. BOAN 游戏 API（我方调用）

### 6.1 接口总览

| 功能 | 方法与路径 | 主要用途 |
| --- | --- | --- |
| 获取游戏 URL | `POST /game/url` | 启动玩家游戏 |
| 厂商列表 | `POST /game/vendors` | 获取已开通厂商 |
| 游戏列表 | `POST /game/list` | 获取指定厂商游戏 |
| 终止会话 | `POST /game/terminate` | 强制结束玩家游戏会话 |

### 6.2 获取厂商列表

`POST https://test-gw.boan.games/game/vendors`

```json
{
  "traceId": "3b37a2f9-f0c3-47d4-b396-ff8e1309a84d",
  "displayLanguage": "zh",
  "currency": "CNY"
}
```

请求字段：`traceId` 必填；`displayLanguage` 用于厂商名称本地化；`currency` 为 ISO-4217 货币代码。响应 `data[]` 包含 `name`、`code`、`currencyCode`、`categoryCode`。

### 6.3 获取游戏列表

`POST https://test-gw.boan.games/game/list`

```json
{
  "traceId": "706bb933-0c18-4b79-bed9-bf0b99c724ee",
  "vendorCode": "PP",
  "pageNo": 1,
  "pageSize": 100,
  "displayLanguage": "zh",
  "currency": "CNY"
}
```

`vendorCode`、`pageNo` 必填，`pageSize` 默认 100。响应使用 `headers` 定义数组列下标，`games` 为二维数组；解析时必须按 `headers` 映射，不要硬编码固定列顺序。常见列包括 `gameCode`、`gameName`、`categoryCode`、图片、语言、平台和币种。

### 6.4 获取游戏启动 URL

`POST https://test-gw.boan.games/game/url`

```json
{
  "traceId": "f8c3de3d-1fea-4d7c-a8b0-29f63c4c3455",
  "username": "player_10001",
  "gameCode": "PP_1302",
  "language": "zh",
  "platform": "web",
  "currency": "CNY",
  "lobbyUrl": "https://example.com/game-lobby",
  "ipAddress": "47.83.239.79"
}
```

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `traceId` | String | 是 | 每次请求唯一 UUID |
| `username` | String | 是 | 我方玩家唯一账号，建议稳定且不可变 |
| `gameCode` | String | 是 | 来自游戏列表 |
| `language` | String | 是 | 默认 `en`，中文简体为 `zh` |
| `platform` | String | 是 | `web` 或 `H5`，具体大小写以联调响应为准 |
| `currency` | String | 是 | ISO-4217 代码，如 `CNY` |
| `lobbyUrl` | String | 是 | 玩家退出游戏后的返回地址 |
| `ipAddress` | String | 是 | 玩家真实 IPv4/IPv6，不应固定填写服务器出口 IP |

成功响应：

```json
{
  "status": "SC_OK",
  "traceId": "f8c3de3d-1fea-4d7c-a8b0-29f63c4c3455",
  "data": {
    "gameUrl": "https://...",
    "token": "f8c3de3d-1fea-4d7c-a8b0-29f63c4ab146"
  }
}
```

`token` 在无缝钱包的余额、下注和派彩回调中用于校验玩家游戏会话。游戏 URL 应由服务端取得后返回前端，避免暴露 BOAN 凭证。

### 6.5 终止游戏会话

`POST https://test-gw.boan.games/game/terminate`

```json
{
  "traceId": "639cb889-1a49-448b-a1e0-82e9d2ecacb5",
  "username": "player_10001"
}
```

## 7. 无缝钱包 API（由我方实现）

我方需要向 BOAN 提供以下公网回调。路径建议严格与文档一致，响应时间应尽量控制在 3 秒内，并保证数据库事务原子性。

| 方法与路径 | 余额操作 | 说明 |
| --- | --- | --- |
| `POST /wallet/balance` | 无 | 查询余额 |
| `POST /wallet/bet` | 扣款 | 普通下注 |
| `POST /wallet/bet_result` | 加款/扣款 | 下注与派彩结果 |
| `POST /wallet/rollback` | 反向冲正 | 回滚原下注影响 |
| `POST /wallet/adjustment` | 正数加款、负数扣款 | 调整历史轮次 |
| `POST /wallet/bet_debit` | 扣款 | 进入特定游戏房间 |
| `POST /wallet/bet_credit` | 加款 | 离开/结算特定游戏房间 |

### 7.1 统一成功响应

```json
{
  "traceId": "f8c3de3d-1fea-4d7c-a8b0-29f63c4c3456",
  "status": "SC_OK",
  "data": {
    "username": "player_10001",
    "currency": "CNY",
    "balance": 100.00,
    "timestamp": 1785460000000
  }
}
```

返回的 `balance` 必须是本次交易提交后的最新可用余额，`timestamp` 为我方钱包交易时间戳（毫秒）。失败响应至少返回原 `traceId` 和准确的 `status`。

### 7.2 `/wallet/balance`

字段：`traceId`、`username`、`currency`、`token`。PDF 标注 `currency` 已废弃、勿校验，但仍建议返回玩家钱包的实际币种；`token` 应与启动游戏返回的 token 一致。

### 7.3 `/wallet/bet`

请求字段：

`traceId`、`username`、`transactionId`、`betId`、`externalTransactionId`、`amount`、`currency`、`token`、`vendorCode`、`gameCode`、`roundId`、`timestamp`。

处理要求：

1. 验签、校验 token、用户和币种。
2. 用 `transactionId` 查询幂等记录；已成功则直接返回当前余额。
3. 在同一数据库事务中锁定钱包、检查余额、扣除 `amount`、写入流水。
4. 余额不足返回 `SC_INSUFFICIENT_FUNDS`，不得产生部分扣款。

### 7.4 `/wallet/bet_result`

主要字段：`transactionId`、`betId`、`roundId`、`betAmount`、`winAmount`、`effectiveTurnover`、`winLoss`、`jackpotAmount`、`resultType`、`isFreespin`、`isEndRound`、`currency`、`token`、`vendorCode`、`gameCode`、`betTime`、`settledTime`。

`resultType`：

| 值 | 钱包处理 |
| --- | --- |
| `WIN` | 玩家赢，按请求金额入账 |
| `BET_WIN` | 同次请求含下注和赢利，按 BOAN 定义处理扣款与入账 |
| `BET_LOSE` | 同次请求含下注并输，处理下注扣款 |
| `LOSE` | 玩家输，无赢利入账 |
| `END` | 仅通知轮次结束，不操作余额 |

同一 `betId` 可能多次派彩，必须以 `transactionId` 幂等；每次新的 `transactionId` 都应独立处理其 `winAmount`。`jackpotAmount > 0` 时也需入账。涉及 `BET_WIN`/`BET_LOSE` 的净额规则应使用 BOAN 测试用例再次确认后固化。

### 7.5 `/wallet/rollback`

字段：`traceId`、`transactionId`、`betId`、`externalTransactionId`、`roundId`、`vendorCode`、`gameCode`、`username`、`currency`、`timestamp`。

根据 `betId` 找到原下注，反向恢复其钱包影响。回滚请求本身仍以新的 `transactionId` 幂等。原交易不存在时返回 `SC_TRANSACTION_NOT_EXISTS`，不要凭请求金额直接加款。

### 7.6 `/wallet/adjustment`

字段：`traceId`、`username`、`transactionId`、`externalTransactionId`、`roundId`、`amount`、`currency`、`gameCode`、`timestamp`。`amount > 0` 增加余额，`amount < 0` 扣除余额。

### 7.7 `/wallet/bet_debit` 与 `/wallet/bet_credit`

这两个端点只适用于 BOAN 兼容性列表中的特定游戏产品。

- `bet_debit`：主要字段为 `transactionId`、`roundId`、`takeAll`、`amount`、`currency`、`gameCode`、`timestamp`。`takeAll=1` 时文档要求全额扣除钱包余额且请求 `amount=0`；`takeAll=0` 时扣除指定 `amount`。
- `bet_credit`：主要字段为 `transactionId`、`betId`、`roundId`、`isRefund`、`amount`、`betAmount`、`winAmount`、`effectiveTurnover`、`winLoss`、`jackpotAmount`、`currency`、`token`、`gameCode`、`betTime`、`settledTime`、`timestamp`。钱包实际加款以 `amount` 为准，`betAmount` 和 `winAmount` 仅供参考。

## 8. 转账钱包 API（我方调用）

仅在 BOAN 确认 Retou 使用转账钱包模式时接入。

| 功能 | 方法与路径 |
| --- | --- |
| 存款至游戏钱包 | `POST /cash/deposit` |
| 从游戏钱包提款 | `POST /cash/withdraw` |
| 查询游戏钱包余额 | `POST /cash/balance` |
| 查询单笔转账 | `POST /cash/getsingletransaction` |

### 8.1 存款/提款

`POST https://test-gw.boan.games/cash/deposit`  
`POST https://test-gw.boan.games/cash/withdraw`

```json
{
  "username": "player_10001",
  "vendorCode": "PP",
  "traceId": "a6cb8549-bf0e-4144-8181-6e786f768d84",
  "transferAmount": "100.00",
  "currency": "CNY",
  "referenceId": "81a8c42e-bb71-4769-b8ea-1aaaf95834fe"
}
```

`transferAmount` 最多 8 位小数且不得低于 `0.00000001`。新转账使用新的 `referenceId`。网络超时或结果未知时，先调用 `/cash/getsingletransaction` 查询，不要立即创建新 referenceId 重试，否则可能重复转账。

### 8.2 查询余额

```json
{
  "username": "player_10001",
  "vendorCode": "PP",
  "traceId": "9eff88b6-41bf-40d1-b593-32c95d4b362c",
  "currency": "CNY"
}
```

### 8.3 查询单笔转账

```json
{
  "username": "player_10001",
  "vendorCode": "PP",
  "traceId": "9eff88b6-41bf-40d1-b593-32c95d4b362c",
  "currency": "CNY",
  "referenceId": "81a8c42e-bb71-4769-b8ea-1aaaf95834fe"
}
```

响应中的 `transactionStatus` 为 `PROCESSING`、`SUCCESS` 或 `FAIL`；`transactionType` 为 `DEPOSIT` 或 `WITHDRAWAL`。

## 9. 交易查询 API（我方调用）

### 9.1 交易列表

`POST https://test-gw.boan.games/transaction/list`

```json
{
  "traceId": "04cb48ce-346f-4dcc-ae49-7955805c50fc",
  "fromTime": 1785456000000,
  "toTime": 1785542399999,
  "pageNo": 1,
  "pageSize": 2000
}
```

`pageSize` 默认 2000，最大 5000。响应与游戏列表相同，使用 `headers` + 二维 `transactions` 数组；必须按响应头映射。下注状态：`0` 未结算、`1` 已结算、`2` 已取消、`3` 已退款。

### 9.2 单笔注单详情

`POST https://test-gw.boan.games/v2/transaction/detail`

```json
{
  "traceId": "6caee2f4-f636-449d-be67-e81d63394dd0",
  "betId": "1646444515999514624",
  "fromTime": 1785456000000,
  "toTime": 1785542399999,
  "displayLanguage": "zh"
}
```

`fromTime` 与 `toTime` 的范围不得超过 7 天。响应可能包含：

- `detailUrl`：厂商提供的注单详情 HTML 地址，未提供时为空；
- `betDetail`：标准化注单数据；
- `vendorBetDetail`：厂商原始/扩展注单详情，结构随厂商变化，必须按 JSON 存储并容错解析。

## 10. 常用状态码

| 状态码 | 含义/处理建议 |
| --- | --- |
| `SC_OK` | 成功 |
| `SC_INVALID_OPERATOR` / `SC_AUTHENTICATION_FAILED` | API Key 无效或缺失，停止重试并检查配置 |
| `SC_INVALID_SIGNATURE` | 签名不一致，检查原始 JSON、编码和 Secret |
| `SC_INVALID_REQUEST` / `SC_WRONG_PARAMETERS` | 参数缺失或错误，记录 `validation` 后修正请求 |
| `SC_MISMATCHED_DATA_TYPE` | 字段类型错误 |
| `SC_INVALID_TOKEN` | 游戏会话 token 无效 |
| `SC_INVALID_GAME` / `SC_GAME_DISABLED` | 游戏无效或禁用 |
| `SC_INVALID_VENDOR` | 厂商无效 |
| `SC_WRONG_CURRENCY` | 交易币种与玩家钱包币种不一致 |
| `SC_CURRENCY_NOT_SUPPORTED` | 平台不支持该币种 |
| `SC_INSUFFICIENT_FUNDS` | 余额不足 |
| `SC_USER_NOT_EXISTS` / `SC_USER_DISABLED` | 用户不存在或已禁用 |
| `SC_TRANSACTION_DUPLICATED` | 交易 ID 重复；按幂等结果处理 |
| `SC_TRANSACTION_NOT_EXISTS` / `SC_TRANSACTION_DOES_NOT_EXIST` | 原交易或 referenceId 不存在 |
| `SC_DUPLICATE_REQUEST` | 重复请求 |
| `SC_REFERENCE_ID_DUPLICATED` | referenceId 已被使用 |
| `SC_TRANSACTION_STILL_PROCESSING` | 交易处理中，延迟后查询 |
| `SC_OPERATOR_TIMEOUT` | 运营商回调超时，检查性能与可用性 |
| `SC_UNDER_MAINTENANCE` | 游戏维护中 |
| `SC_WALLET_NOT_SUPPORTED` | 当前钱包类型不支持 |
| `SC_UNKNOWN_ERROR` / `SC_INTERNAL_ERROR` | 未知/内部错误，保留 traceId 联系 BOAN |

## 11. 语言、平台和币种

PDF 列出的语言包括：`zh` 简体中文、`hk` 繁体中文、`en` 英语、`id` 印尼语、`tl` 他加禄语、`vi` 越南语、`hi` 印地语、`pt` 葡萄牙语、`th` 泰语、`es` 西班牙语。

平台主要为 `web`、`H5`。游戏实际支持的语言、平台、币种必须以 `/game/list` 返回的对应字段为准。

PDF 示例包含 `CNY`、`USD`，币种清单还列出 `PHP`、`IDR(K)`、`VND(K)` 等。Retou 测试账号实际开放币种需从 `/game/vendors`、`/game/list` 结果和 BOAN 后台共同确认。

## 12. 钱包实现建议

钱包交易推荐采用以下事务流程：

```text
接收原始请求体
  -> 验签与基础校验
  -> 查询 transactionId 幂等记录
  -> 锁定用户钱包行
  -> 校验用户、币种、token、余额
  -> 原子更新余额
  -> 写入不可变钱包流水及请求快照哈希
  -> 提交事务
  -> 返回最新余额
```

最低数据字段建议：`transaction_id`、`trace_id`、`bet_id`、`round_id`、`username`、`currency`、`operation`、`amount`、`balance_before`、`balance_after`、`status`、`request_hash`、`vendor_code`、`game_code`、`created_at`。

并发要求：同一玩家的资金变更必须串行化或使用数据库行锁/原子条件更新；余额不足判断与扣款必须处于同一事务。不要通过“先查余额、再单独更新”实现。

## 13. 重试与超时策略

- 查询类接口可对网络错误、HTTP 5xx 做有限指数退避重试。
- 资金接口网络超时时，结果属于未知；必须先按 `transactionId` 或 `referenceId` 查询/重放相同幂等键，不得生成新键直接重试。
- `SC_INVALID_SIGNATURE`、参数错误、币种错误等确定性错误不应自动重试。
- 记录 `traceId`、`transactionId`、`referenceId`、路径、耗时、HTTP 状态和业务状态，敏感字段脱敏。
- 建议连接超时 3 秒、总超时 10 秒；最终值以 BOAN SLA 和联调结果为准。

## 14. 联调步骤

1. BOAN 确认并完成我方出口 IP 白名单。
2. 使用固定 JSON 验证 HMAC-SHA256 签名，调用 `/game/vendors`。
3. 调用 `/game/list`，确认 Retou 已开放的厂商、游戏、币种、语言和平台。
4. 创建测试玩家并调用 `/game/url`，验证 `gameUrl` 和 token。
5. 无缝钱包模式下，配置我方公网回调地址并依次验证余额、下注、重复下注、余额不足、派彩、多次派彩、回滚和超时重试。
6. 转账钱包模式下，验证存款、查询、提款、重复 referenceId 和网络超时后的单笔查询。
7. 使用 `/transaction/list` 和 `/v2/transaction/detail` 对账，核对钱包流水、下注额、派彩额和状态。
8. 完成异常测试：错误签名、错误币种、无效 token、禁用用户、并发下注、重复回调和乱序回调。

## 15. 上线前检查清单

- [ ] BOAN 已确认钱包模式、正式 Base URL、正式 API Key/Secret。
- [ ] 测试凭证已从代码和文档交付物之外的公开位置清除。
- [ ] 正式出口 IP 和回调来源 IP 白名单已配置。
- [ ] 所有钱包接口通过签名、幂等、并发、回滚和超时测试。
- [ ] 金额字段使用定点十进制，时间戳统一为毫秒。
- [ ] 游戏列表定时同步且按 `headers` 动态解析。
- [ ] 日志、指标、告警和每日对账任务已启用。
- [ ] 后台账号已修改默认密码并启用可用的二次认证措施。
- [ ] 已取得 BOAN 对 `resultType` 净额规则、回调请求头、超时 SLA 和重试次数的书面确认。

## 16. 待 BOAN 确认事项

1. Retou 使用无缝钱包还是转账钱包，或不同厂商分别使用不同模式。
2. 钱包回调完整公网域名、BOAN 回调来源 IP、回调是否携带 `X-API-Key`。
3. `BET_WIN`、`BET_LOSE` 下 `betAmount`、`winAmount` 的准确余额计算方式。
4. `/game/url` 的 `platform` 在实际请求中要求 `web/H5` 还是 `WEB/H5`。
5. Retou 测试与正式环境实际开放的币种、厂商、游戏和钱包端点兼容范围。
6. API 超时、最大重试次数、回调乱序以及派彩晚到的处理窗口。

