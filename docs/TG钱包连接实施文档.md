# Telegram Mini App 连接 TON 钱包实施文档

本文档说明如何在 Telegram Mini App（TMA）中接入 TON 钱包，并发起 TON / Jetton（USDT）交易。内容对应当前项目的 `tg/wallet.html`、`tg/wallet.js` 和 `tg/tonconnect-manifest.json`。

## 1. 整体流程

```text
Telegram Bot
  └─ 打开 Mini App（HTTPS）
       └─ 页面加载 TonConnect UI
            └─ 用户选择钱包并授权
                 ├─ TON_CONNECT_UI 取得账户地址
                 └─ sendTransaction 发起交易
                      └─ 钱包确认后回到 Mini App
                           └─ 服务端通过链上数据确认到账
```

TonConnect 只负责“连接钱包”和“请求用户签名”。交易是否成功、金额是否正确，必须由服务端查询链上记录后确认，不能只相信浏览器返回值。

## 2. 前置条件

1. 一个 Telegram Bot，并通过 BotFather 配置 Mini App 菜单按钮或 direct link。
2. 一个可通过公网 HTTPS 访问的站点，例如：
   - Mini App：`https://www.youqiu666.com/hddfw/`
   - Manifest：`https://www.youqiu666.com/hddfw/tg/tonconnect-manifest.json`
3. 一个钱包地址作为收款地址。
4. 钱包应用支持 TON Connect（Tonkeeper、Telegram Wallet 等）。

Telegram Mini App 的入口地址应指向 H5 页面，而不是 Bot 对话地址。例如：`https://t.me/web333tbot/myapp`。

## 3. 配置 TonConnect Manifest

Manifest 用于在钱包选择页展示应用名称、图标、服务条款和隐私政策。

`tg/tonconnect-manifest.json` 示例：

```json
{
  "url": "https://www.youqiu666.com/hddfw/",
  "name": "RichIsland",
  "iconUrl": "https://www.youqiu666.com/hddfw/tg/logo.png",
  "termsOfUseUrl": "https://www.youqiu666.com/hddfw/tg/termOfUse.html",
  "privacyPolicyUrl": "https://www.youqiu666.com/hddfw/tg/privacyPolicy.html"
}
```

检查项目：

- 必须使用 HTTPS，且所有 URL 可从公网访问。
- JSON 响应的 `Content-Type` 应为 `application/json`。
- 图标建议使用稳定的 PNG/JPG 地址。
- `url` 应与实际 Mini App 首页一致。

## 4. 页面引入 SDK 并渲染连接按钮

`wallet.html` 中引入 TonConnect UI：

```html
<script src="https://unpkg.com/@tonconnect/ui@latest/dist/tonconnect-ui.min.js"></script>
<div id="ton-connect"></div>
<script src="./wallet.js"></script>
```

建议生产环境锁定 SDK 版本，不要长期使用 `@latest`，避免第三方升级导致页面行为变化。

## 5. 初始化 TonConnect UI

```js
const tonConnectUI = new TON_CONNECT_UI.TonConnectUI({
  manifestUrl: 'https://www.youqiu666.com/hddfw/tg/tonconnect-manifest.json',
  buttonRootId: 'ton-connect'
});

tonConnectUI.uiOptions = {
  actionsConfiguration: {
    returnStrategy: 'back',
    twaReturnUrl: 'https://t.me/web333tbot/myapp',
    notifications: ['before', 'success', 'error']
  }
};
```

`twaReturnUrl` 是钱包操作完成后的 TMA 回跳地址，必须是 Telegram Mini App 的完整链接。不要填写 Bot 普通聊天链接，否则用户会回到对话框而不是应用页面。

## 6. 连接钱包

### 6.1 用户点击连接

TonConnect UI 会自动生成连接按钮，用户点击后选择钱包并授权：

```js
async function connectWallet() {
  try {
    const wallet = await tonConnectUI.connectWallet();
    console.log('wallet connected:', wallet.account.address);
  } catch (error) {
    console.error('wallet connection failed:', error);
  }
}
```

### 6.2 恢复上次连接

页面重新打开时建议恢复连接状态：

```js
await tonConnectUI.connectionRestored;

if (tonConnectUI.connected) {
  const address = tonConnectUI.account.address;
  console.log('current wallet:', address);
}
```

如需主动连接，可调用项目中的 `connectToWallet(tonConnectUI)`。生产环境应同时处理用户取消、钱包不存在、网络错误等异常。

## 7. 获取账户信息

```js
if (!tonConnectUI.connected) {
  throw new Error('请先连接钱包');
}

const account = tonConnectUI.account;
const userAddress = account.address;
const chain = account.chain; // 主网或测试网标识
```

地址应在服务端统一规范化后比较，避免 bounceable / non-bounceable 地址格式不同导致误判。不要把私钥、助记词或钱包签名材料发送到服务器。

## 8. 发起 TON 转账

TON 转账交易结构如下：

```js
const transaction = {
  validUntil: Math.floor(Date.now() / 1000) + 360,
  messages: [{
    address: '收款 TON 地址',
    amount: TonWeb.utils.toNano('0.1').toString()
  }]
};

const result = await tonConnectUI.sendTransaction(transaction);
```

注意：`amount` 的单位是 nanotons（1 TON = 1,000,000,000 nanotons），不能直接传 `0.1`。`validUntil` 建议设置为当前时间后的 5～10 分钟。

## 9. 发起 USDT（Jetton）转账

Jetton 转账不是直接向收款人地址转 TON，而是向“发送方的 USDT Jetton Wallet 合约”发送一条带 payload 的 TON 消息。payload 中包含 Jetton 转账 opcode、数量、目标地址等字段。

当前项目 `dotransferUsdtDollar` 的关键步骤：

1. 使用 USDT Jetton Master 地址查找发送方 Jetton Wallet 地址。
2. 构造 Jetton transfer body（opcode `0x0f8a7ea5`）。
3. 按 USDT 精度填写数量（当前 USDT 通常为 6 位小数）。
4. 通过 `sendTransaction` 发送，并附带少量 TON 支付 Gas。

示意代码：

```js
const transaction = {
  validUntil: Math.floor(Date.now() / 1000) + 360,
  messages: [{
    address: senderJettonWalletAddress,
    amount: TonWeb.utils.toNano('0.05').toString(),
    payload: base64JettonTransferBody
  }]
};

await tonConnectUI.sendTransaction(transaction);
```

上线前必须确认：USDT Master 地址、网络（主网/测试网）、目标地址、Jetton Wallet 地址和 decimals 均来自可信配置，不能使用测试代码中的固定地址。

## 10. 交易结果确认（必须服务端完成）

浏览器端的 `sendTransaction` 返回值只能表示钱包提交请求的结果，不能作为“已到账”凭证。推荐流程：

1. 服务端创建订单，生成唯一订单号和期望金额。
2. 前端连接钱包并发起交易。
3. 前端将订单号、用户地址、交易返回信息提交服务端。
4. 服务端通过 TON API / 节点查询交易，校验：
   - 收款地址是否为商户地址；
   - 发送地址是否为当前用户地址；
   - 金额、Jetton Master、交易成功状态是否匹配；
   - 交易是否已被其他订单使用（幂等校验）。
5. 校验通过后再将订单标记为已支付。

轮询交易时应设置超时、最大重试次数和重复交易保护。不要采用“只要最新一笔交易 hash 变化就算成功”的判断方式。

## 11. Telegram Mini App 初始化

在 TMA 中建议调用：

```html
<script src="https://telegram.org/js/telegram-web-app.js"></script>
<script>
  Telegram.WebApp.ready();
  Telegram.WebApp.expand();
</script>
```

Telegram 用户身份应通过 `initData` 在服务端验签后使用，不能直接信任前端传来的用户 ID。

## 12. 上线前安全检查

- 当前 `wallet.js` 将 Toncenter API key 写在前端代码中。该 key 已公开，建议立即撤销并更换；节点 API 调用应迁移到服务端。
- 当前 `payment/index.js` 中还存在 Bot Token、第三方支付 key/secret 和钱包助记词等明文凭据。应视为已经泄露，立即撤销或轮换，并迁移到服务端环境变量或密钥管理服务；助记词对应钱包建议更换。
- 收款地址、Jetton Master 地址和金额不要仅由前端决定，服务端应根据订单生成并校验。
- 生产环境移除 `eruda` 调试脚本，避免暴露运行时信息。
- 不要在日志中打印敏感参数、完整用户身份数据或未脱敏订单信息。
- 前端依赖锁定版本并配合 Subresource Integrity 或自托管静态资源。
- 同时测试 Android、iOS、Telegram Desktop，以及用户取消授权、网络中断、重复点击和交易失败场景。

## 13. 常见问题

### 钱包连接后没有回到 Mini App

检查 `twaReturnUrl` 是否为正确的 `https://t.me/<bot>/<app>` 地址，并确认 BotFather 中的 Mini App 名称和 direct link 一致。

### Manifest 加载失败

直接在浏览器打开 `manifestUrl`，确认 HTTPS、状态码 200、JSON 格式和跨域策略正常。

### USDT 转账显示成功但余额未到账

通常是 Jetton Master、Jetton Wallet、目标地址或 decimals 配置错误。应在链上解析实际消息和事件，不要只看钱包 UI 提示。

### 用户重复支付

服务端必须使用订单号、交易 hash 和链上唯一标识做幂等处理，并明确订单过期时间。

## 14. 参考资料

- TON Connect：<https://docs.ton.org/develop/dapps/ton-connect>
- TonConnect Manifest：<https://docs.ton.org/develop/dapps/ton-connect/protocol/requests-responses#app-manifest>
- Telegram Mini Apps：<https://core.telegram.org/bots/webapps>
- Telegram Mini App direct link：<https://core.telegram.org/bots/webapps#direct-link-mini-apps>
- Telegram Stars：<https://core.telegram.org/api/stars>
