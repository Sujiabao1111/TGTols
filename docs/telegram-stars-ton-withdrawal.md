# Telegram Stars 充值与 TON 钱包提现流程

## 1. 结论与范围

可以兼容当前系统实现。现有后端是 Go + Fiber + GORM，核心入口为 `be/controllers/payment_handler.go`、`be/services/payment_service.go`，已有 `payment_orders` 充值订单和 `withdraw_orders` 提现订单。因此建议采用“适配器接入”方式：

- Telegram Stars（`XTR`）只作为新的充值渠道，成功后按锁定汇率折算为平台余额。
- TON 作为新的提现渠道，用户提交 TON 主网地址，平台通过 TON 钱包服务/节点广播交易。
- 保留现有 TargetedPay、TokenPay、GCash、银行卡等方式，按 `method/channel` 路由隔离。
- 所有回调、提现提交和链上查询必须幂等，不能仅依赖前端返回结果。

> 重要限制：Telegram Stars 是 Telegram 内部的数字商品支付单位。Bot API 支付币种使用 `XTR`，不能把 Stars 当作 TON 直接转给用户。商户侧 Stars 结算、退款和提现受 Telegram/Fragment 规则约束；应用内 TON 提现需要平台自有 TON 资金池或合规的托管/支付服务。

## 2. 推荐架构

```text
Telegram Mini App / Bot
        |
        | 1. 创建 Stars 订单
        v
be /api/payments/telegram-stars/order
        |
        | 2. Bot API sendInvoice(currency=XTR)
        v
Telegram checkout
        |
        | 3. pre_checkout_query -> answerPreCheckoutQuery
        | 4. successful_payment -> Telegram webhook
        v
be /api/payments/telegram-stars/webhook
        |
        | 5. 校验 payload、金额、provider_payment_charge_id，幂等入账
        v
payment_orders -> 用户余额

用户申请提现 -> withdraw_orders(status=pending/processing)
        |
        | TON Connect 验证地址或服务端校验地址
        v
TON payout adapter -> toncenter/tonapi 或托管服务 -> 链上确认
        |
        v
提现成功；失败则原子退款余额
```

## 3. Telegram Stars 充值流程

### 3.1 前置配置

1. 通过 BotFather 创建或确认 Bot，配置 Mini App 域名（HTTPS）。
2. 服务端保存 `TELEGRAM_BOT_TOKEN`，只允许后端调用 Bot API。
3. 配置 webhook：`https://<domain>/api/payments/telegram-stars/webhook`，使用 `secret_token`，并校验请求头 `X-Telegram-Bot-Api-Secret-Token`。
4. 明确定价表：业务金额（如 USD/USDT）到 Stars 的兑换关系，订单创建时固化 `stars_amount`、`fiat_amount`、`rate_id`，后续不随汇率变化。
5. 在 Telegram 生产环境先完成小额真实支付、退款和异常回调测试。测试 Bot 与生产 Bot、数据库订单必须隔离。

### 3.2 创建订单

新增接口：`POST /api/payments/telegram-stars/order`（JWT）。请求至少包括：

```json
{"amount": 100, "currency": "USD"}
```

后端执行：

1. 校验金额、用户状态、活动/打码规则。
2. 生成唯一 `order_id` 和随机 `payload`（不得只放 user_id）。
3. 在 `payment_orders` 写入 `type=telegram_stars`、`status=pending`、`dst_code=TG_STARS`、Stars 数量和汇率快照。
4. 调用 `sendInvoice`：`currency=XTR`，`prices` 只能使用 Stars 整数，`payload` 使用服务端生成值，`provider_token` 按 Telegram 当前要求处理（不要写死旧版行为）。
5. 返回 `invoice_link` 或 Mini App 可直接打开的支付参数。

建议扩展字段（可放 JSON metadata，或单独迁移列）：`telegram_user_id`、`stars_amount`、`telegram_charge_id`、`provider_charge_id`、`payload`、`rate_snapshot`。

### 3.3 回调与入账

- 收到 `pre_checkout_query`：根据 payload 查询订单，校验订单未支付、金额和币种为 `XTR`，在时限内调用 `answerPreCheckoutQuery(ok=true/false)`。
- 收到 `successful_payment`：校验 payload、`total_amount`、`currency=XTR`；记录 `telegram_payment_charge_id` 和 `provider_payment_charge_id`；数据库事务内将订单改为成功并给用户余额入账。
- 使用唯一索引约束 `payload`、`provider_charge_id`，重复 webhook 直接返回 200，不重复加钱。
- Telegram webhook 只返回快速成功响应，耗时的汇率、账务和通知处理放队列/后台任务；失败可重试。
- 退款必须关联原订单，成功退款后冲销余额；禁止把退款当作普通负数充值。

## 4. TON 提现流程

### 4.1 资金模型选择

推荐初期使用合规托管服务的 payout API；若自建钱包，需要热钱包、冷钱包、限额审批、助记词/HSM 管理、TON 节点和余额监控。平台必须预先准备 TON 流动性，Telegram Stars 结算到 TON 并非实时到账。

提现资产要明确：

- 原生 TON：单位 nanoTON（1 TON = 1,000,000,000 nanoTON）。
- Jetton（如 USDT）：必须额外保存 `jetton_master`、decimals、recipient token wallet，不能把 Jetton 地址当 TON 地址。

### 4.2 用户地址绑定

1. 前端用 TON Connect 打开钱包连接，服务端接收地址和链（mainnet/testnet）。
2. 服务端使用 TON 地址解析库校验格式、workchain、bounceable/test-only 标志；生产提现禁止 testnet 地址。
3. 推荐要求用户签名一条 nonce 消息，证明地址所有权；只保存标准化地址，不保存私钥。
4. 提现请求时再次校验地址白名单、风险评分、最小/最大金额、每日次数和手续费。

### 4.3 提现订单与广播

新增提现方式配置：`TON`（或 `TON_USDT`），接入当前 `GetWithdrawMethods`、`CreateWithdrawOrder` 和 `submitApprovedWithdrawToGateway` 抽象。建议状态：`pending -> processing -> success/failed`，并增加 `broadcasted` 或用 `processing` 表示已广播。

提交接口仍可复用现有提现接口，示例：

```json
{"type":"crypto","dst_code":"TON","amount":20,"address":"UQ..."}
```

事务规则：

1. 创建订单时锁定/扣减可提现余额，写入汇率和手续费快照。
2. 自动小额提现沿用现有 `autoWithdrawLimitUSD`；大额进入后台审核和风控。
3. 使用订单号作为 TON comment/memo（如服务支持），保存 `tx_hash`、`lt`、`query_id`。
4. 广播成功后轮询 tonapi/toncenter，达到配置确认数才标记成功；查询超时保持 processing，不能直接失败退款。
5. 广播失败或明确链上失败时，事务内将订单标记 failed 并原子退回余额；已上链但未确认不可退款，需人工处理。

## 5. 兼容改造清单

### 后端 `be`

- `config.json` 增加 `telegram`、`ton` 配置：Bot token、webhook secret、API base URL、托管服务 key、network、钱包地址、手续费、确认数和限额。
- `models/dtos/payment_models.go` 增加 Stars 创建订单/回调 DTO、TON 地址和提现 DTO。
- `services/payment_service.go` 增加 Stars order/create/notify 与 TON payout adapter；复用现有余额、打码、提现审核和退款函数。
- `controllers/payment_handler.go` 增加 Stars order、`pre_checkout`、webhook 和 TON 地址校验路由；webhook 不挂 JWT，但必须 secret token + 签名/字段校验。
- 数据库迁移：为 `payment_orders`、`withdraw_orders` 增加通道、链上/Telegram 交易标识和 metadata 索引；金额使用 decimal，不使用浮点。
- 增加 webhook 去重表或唯一索引、审计日志、管理员重试/人工放行接口。

### 前端 `fe`

- 钱包充值方式列表新增 Telegram Stars，仅在 Telegram Mini App/支持环境展示。
- Stars 支付跳转 Telegram invoice，展示处理中状态并轮询订单，不以页面回跳作为成功依据。
- 提现方式列表新增 TON；集成 TON Connect，显示网络、标准化地址、手续费、到账金额和风险提示。
- 复用现有提现记录页，新增 tx hash 链接（Tonviewer）和 processing 状态。

## 6. 安全、风控与合规

- Bot token、TON 私钥/托管 API key 只能放环境变量或密钥管理服务，禁止进 `config.json`、日志和前端。
- 所有金额使用 decimal/nanoTON 整数，汇率和手续费在订单创建时快照。
- webhook、回调、广播接口全部幂等；状态只能按有限状态机前进。
- TON 地址风控：黑名单、重复地址、频率/额度限制、人工审核、大额二次确认。
- 记录 Telegram user id 与站内 user id 的绑定，防止 payload 被替换或跨用户重放。
- 上线前确认 Telegram Stars 商户资格、退款政策、当地博彩/虚拟资产/AML/KYC 要求，以及 TON 托管服务的地区可用性。

## 7. 测试与上线顺序

1. 单元测试：金额换算、payload 校验、重复回调、状态机、地址校验、手续费和退款。
2. 集成测试：Telegram webhook 重试、超时、错误 pre-checkout；TON testnet 广播、确认、节点不可用和重复提交。
3. 灰度：先开启 Stars 小额充值和 TON testnet/白名单提现，观察账务对账、余额差异和失败退款。
4. 生产切换：启用主网配置、热钱包低余额告警、每日对账和人工应急开关。
5. 验收标准：同一 Telegram payment charge 只能入账一次；同一提现订单最多一笔链上交易；所有最终状态可在后台追溯。

## 8. 需要业务方先确认的参数

- Stars 与平台余额的定价、最低充值和退款规则。
- 提现币种是原生 TON 还是 TON 上的 USDT Jetton。
- 自建钱包还是第三方托管 payout，及其地区合规范围。
- TON 提现手续费承担方、确认数、最小/最大额度、每日限额和审核阈值。
- 是否要求 KYC、地址签名绑定和提现地址白名单。

