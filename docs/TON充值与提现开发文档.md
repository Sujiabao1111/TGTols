# TON 充值与提现开发文档

> 本文根据当前客户端代码整理，覆盖 TON 钱包支付充值、TON 银行提现，以及“代币兑换后提现 TON”相关流程。协议的最终校验、扣款和链上打款由服务端/第三方支付完成，客户端只负责发包、调起钱包和刷新状态。

## 1. 代码入口

| 模块 | 关键文件 |
|---|---|
| 协议号 | `assets/Script/NetWork/ProtocolHandles/MsgDefine.ts` |
| 请求结构 | `assets/Script/NetWork/GameStruct.ts` |
| TON 银行数据与回包 | `assets/Script/NetWork/ProtocolHandles/DataOperModules/TonBankModule.ts`、`assets/Script/GameData/GD_TonBank.ts` |
| 充值商品与钱包支付 | `assets/Script/NetWork/ProtocolHandles/DataOperModules/RechargeGoodsModule.ts` |
| 充值界面 | `assets/Script/View/Recharge/RechargeDialog.ts`、`RechargeItem.ts`、`assets/Script/View/FcExchange/BuySelectDialog.ts` |
| TON 银行界面 | `assets/Script/View/Ton/TonDialog.ts`、`TonAddSpeedDialog.ts` |
| TON Connect 钱包 | `assets/Script/View/Ton/WalletMgr.ts` |
| 代币提现 | `assets/Script/NetWork/ProtocolHandles/DataOperModules/CryptoCoinModule.ts`、`assets/Script/View/cryptoCoin/CryptoDrawItem.ts` |

## 2. 单位与钱包约定

- 服务端金额通常为整数最小单位。TON 展示/支付前使用 `Utils.getTon(value)`（代码中常见除以 `10000`）；充值商品价格 `Price` 为美元 `1/10000`，客户端自行换算。
- 代币数量 `Count` 注释约定为 `/1e6`；代币价格 `Price` 需除以 `1e9` 才是实际价格。
- 钱包地址统一由 `WalletMgr.getAddress()` 获取；连接、断开和转账均由 `WalletMgr` 封装。

## 3. TON 充值（购买游戏商品）

### 3.1 流程

1. 进入充值/兑换页面，客户端发送 `100340` 获取商品列表。
2. 用户选择商品和支付币种，在 `BuySelectDialog` 组装购买请求 `100341`。
3. 服务端返回订单（收款地址、金额、TransferId 等）。
4. 客户端收到订单后调用 TON Connect 钱包转账；钱包完成链上交易后由服务端确认并下发充值成功。
5. 客户端收到 `100344`（`RechargeOk`）后刷新购买次数、关闭弹窗并展示奖励。

### 3.2 C2S 协议

`MsgId_C2S_GetRechargeGoodsList = 100340`

```ts
{ MsgId: 100340, Module: string }
```

`Module` 用于区分页面（如 `farmCoin`）。

`MsgId_C2S_BuyRechargeGoods = 100341`

```ts
{
  MsgId: 100341,
  Gid: number,             // 商品 ID
  Currency: "ton" | "ton-usdt" | string,
  FromM: string,           // dailyRecharge/newbieRecharge/turntableRecharge 等
  PayChannel: "idpay" | string,
  DstCode: ""
}
```

当前 TON 支付由 `RechargeGoodsModule.MsgId_S2C_RechargeGoodsOrder` 分流：

- `Currency == ton`：`WalletMgr.dotransfer(Utils.getTon(Amount), WalletAddr, TransferId)`。
- `Currency == ton-usdt`：`WalletMgr.dotransferUsdtDollar(Utils.getTon(Amount), WalletAddr, TransferId)`。

### 3.3 S2C 订单

`MsgId_S2C_RechargeGoodsOrder = 100341`

```ts
{
  Currency: string,
  TransferId: string,      // 第三方/系统订单号，必须原样带入钱包备注或转账参数
  Amount: number,          // 客户端需按 TON 单位换算
  WalletAddr: string,      // 平台收款钱包
  InvoiceLink?: string,
  QrCodeBase64?: string,
  ExpireTime?: string,
  TokenPayId?: string,
  PayUrl?: string
}
```

`MsgId_S2C_RechargeOk = 100344`：`Rewards: number[][]`。收到后更新空投积分、关闭购买弹窗；有奖励则展示奖励，否则提示“充值成功”。购买次数变更使用 `100343`。

### 3.4 钱包要求

`WalletMgr.ts` 负责 TON Connect 初始化、地址获取、转账和断开。发起充值前应确保钱包已连接；订单过期、用户拒签、链上失败需由钱包回调或服务端订单状态处理，不能仅凭客户端点击判定成功。

## 4. TON 银行提现

TON 银行是游戏内累计 TON 的提现队列，不等同于充值购买。服务端下发银行余额和多个提现档位，客户端按档位展示任务进度。

### 4.1 初始化

`MsgId_C2S_GetTonBankData = 100330`（仅 `TonBanks.IsInit == false` 时请求）。

回包 `MsgId_S2C_TonBankData = 100330`：

```ts
{ Ton: number, Withdraws: GD_TonBank[] }
```

`GD_TonBank` 字段：`Cid` 档位 ID、`Ton` 所需余额、`Task` 任务数据、`IsFree` 是否免费提现。任务字段包括 `C/CC` 进度、`M/CM` 已提现/上限次数、`Cost` 加速所需紫晶、`State` 状态。

任务状态：0 未完成，1 已完成可提现，2 提现中，3 排队中，4 加速成功/等待处理。

### 4.2 发起提现

`MsgId_C2S_TonBanWithdraw = 100331`

```ts
{ MsgId: 100331, Cid: number, WalletAddr: string }
```

客户端前置条件：钱包地址存在；银行 TON 余额 `Ton >= Withdraw.Ton`；任务完成（除 `IsFree == 1`）；累计提现次数未达到 `Task.CM`。满足后由 `TonDialog` 发送请求。

回包 `MsgId_S2C_TonBankWithdrawRet = 100331`：`{ Ton, Withdraw }`。客户端更新余额和对应档位，提示提现成功，并派发 `E_OnRefreshWalletContext`。链上到账仍以服务端/钱包为准。

### 4.3 提现加速

`MsgId_C2S_TonBanWithdrawSpeedUp = 100332`

```ts
{ MsgId: 100332, Cid: number }
```

客户端检查紫晶余额不少于 `Task.Cost` 后发送。成功回包 `100334`（`TonBankWithdrawSpeedUpRet`）携带更新后的 `Withdraw`。

### 4.4 异步变更

- `100332` `TonBankWithdrawTaskChange`：批量更新 `Withdraws` 中任务。
- `100333` `TonBankTonChange`：只更新银行 TON 余额。

两者都会刷新钱包上下文；前端必须按 `Cid` 合并数据，避免覆盖其他档位。

## 5. 代币兑换/提现 TON（CryptoCoin）

该链路用于代币系统：`CryptoCoinModule` 接收代币余额和价格，TON 的代币 ID 在代码中约定为 `Sdid == 15`。

- 获取/同步代币：`110150`、`110151`。
- 代币兑换：`MsgId_C2S_SwapCryptoCoin = 110150`，成功回包 `110152`。
- 代币提现：`MsgId_C2S_WithdrawCryptoCoin = 110151`；成功回包 `110153`，字段 `{ Ton, Id, WithdrawCnt }`。客户端更新指定代币提现次数并提示区块链到账需要时间。

注意：`CryptoDrawItem.ts` 中提现请求代码目前存在注释，若重新启用需确认完整请求字段（至少 `MsgId`、代币 ID/数量、`WalletAddr`）与服务端协议一致，不能直接照搬 TON 银行请求。

## 6. 错误与验收建议

- 订单重复回调必须幂等：以 `TransferId/TokenPayId` 去重，充值奖励只能发放一次。
- 金额转换统一封装，禁止在界面层重复除法导致精度错误；展示建议固定小数位。
- 提现请求发送前再次校验余额、任务状态、次数上限和钱包地址；服务端必须再次校验。
- 钱包拒签、断开、订单过期、链上失败、服务端审核拒绝均应有明确状态和可重试策略。
- 测试覆盖：TON 与 TON-USDT 两种充值、订单过期/重复回调、无钱包提现、余额不足、任务未完成、次数上限、提现中断线重连，以及异步余额/任务推送。

