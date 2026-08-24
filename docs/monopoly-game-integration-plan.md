# 大富翁游戏接入方案（修订版）

## 需求概述

游戏端（12661）的大富翁游戏需要接入我方游戏系统，包含两个需求：
1. **需求1**：大富翁游戏格子展示我方随机5个游戏的信息（图标、名称、game_code）
2. **需求2**：踩中格子时直接跳转进入对应游戏，跳过大厅和选游戏步骤

---

## 关键问题与解决方案

### Token问题
**问题**：大富翁游戏在踩格子前，用户未进入我方大厅，没有JWT Token

**解决方案**：新增"快速登录"接口，12661提供username，我方直接生成Token返回（同服务器免密验证）

**流程变更**：
```
原流程：用户踩格子 → 12661用已有token调直达游戏接口 ❌

新流程：
1. 用户踩格子 → 12661调快速登录接口（传username）→ 获取临时token
2. 12661用token + game_code调直达游戏接口 → 进入游戏
```

---

## 实施方案

### 新增接口1：快速登录（供12661调用）

**接口地址**：`POST /api/auth/quick-login`

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户用户名 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

**实现逻辑**：
1. 根据username查询users表
2. 如果不存在 → 自动创建用户（同服务器信任机制）
3. 生成JWT Token返回

**注意事项**：
- 仅允许内网/同服务器访问（通过IP白名单或Nginx层限制）
- 不验证密码，基于同服务器信任原则

---

### 新增接口2：获取随机游戏列表

**接口地址**：`GET /api/games/random-for-monopoly`

**请求参数**：无（或可选 `count=5` 指定数量，默认5）

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "games": [
      {
        "game_code": "vs20olympgate",
        "name": {
          "EN": "Gates of Olympus",
          "CN": "奥林匹斯之门"
        },
        "img_url": "https://xxx.com/img/olympus.png",
        "provider_code": "PG",
        "provider_name": "Pragmatic Play"
      }
    ]
  }
}
```

**实现逻辑**：
1. 查询 `games` 表 `status=1` 的记录
2. 按 `RAND()` 随机排序
3. 取前5条
4. Preload `Provider` 获取厂商信息

---

### 新增接口3：大富翁直达游戏

**接口地址**：`GET /games/launch-monopoly`

**请求参数**（Query String，沿用launch3风格）：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | JWT Token（来自快速登录接口） |
| game_code | string | 是 | 游戏代码（来自需求1的列表） |
| is_mobile | bool | 否 | 是否移动端，默认false |

**响应**：HTML 页面（同 `/games/launch3`）

**实现逻辑**：
```
1. 解析token获取userId（复用ParseTokenManual）
2. 根据game_code查询games表（Preload Provider）
   - 不存在 → 返回错误HTML
3. 调用SyncGameBalance(userId, 0, 1) 查询12661余额
   - 失败 → 返回错误HTML
4. 如果balance > 0：
   - 创建流水记录
   - 调用Hedoc UpdateBalance转入（Type=1）
   - 失败 → 调用SyncGameBalance冲正 → 返回错误HTML
5. 调用Hedoc OpenGame获取游戏URL
6. 返回渲染好的HTML（复用现有HTML模板）
```

**与LaunchGame3的核心区别**：
| 对比项 | LaunchGame3 | LaunchMonopoly |
|--------|-------------|----------------|
| 游戏标识 | game_code参数 | game_code参数（相同）|
| 获取方式 | Query解析 | Query解析（相同）|
| Token来源 | 用户登录大厅获取 | 快速登录接口获取 |

---

## 完整调用流程

### 需求1：获取游戏列表（12661服务端定时/启动时调用）
```
12661服务端 ──GET──▶ /api/games/random-for-monopoly
                         │
                         ▼
                  返回5个随机游戏
                         │
                  缓存到本地/内存
                         │
                  渲染大富翁格子
```

### 需求2：用户踩中格子进入游戏
```
用户踩中格子
    │
    ▼
12661调用 /api/auth/quick-login (传username)
    │
    ▼
获取临时token
    │
    ▼
12661调用 /games/launch-monopoly?token=xxx&game_code=xxx
    │
    ▼
返回游戏HTML
    │
    ▼
用户浏览器加载游戏页面
```

---

## 安全考虑

### 快速登录接口的安全
1. **IP白名单**：建议Nginx层限制仅允许12661服务IP访问
2. **内网绑定**：接口监听127.0.0.1或内网IP，不暴露公网
3. **参数校验**：username必须符合规范（长度、字符限制）

### Token有效期
- 建议设置较短有效期（如5-10分钟）
- 仅用于单次游戏进入，不用于其他接口

---

## 代码实现位置

追加到 `controllers/thirdprovider.go` 文件末尾，添加清晰注释分隔：

```go
// ==========================================
// 大富翁游戏接入专用接口 (Monopoly Integration)
// 追加日期: 2026-03-20
// 需求方: 12661游戏端
// ==========================================

// QuickLoginForMonopoly 快速登录接口（同服务器免密）
// ...

// GetRandomGamesForMonopoly 获取随机游戏列表
// ...

// LaunchMonopolyGame 大富翁直达游戏
// ...
```

---

## 路由注册

在 `router.go` 的 `InitRouteTables` 中添加：

```go
// 大富翁游戏专用接口
r.Post("/api/auth/quick-login", QuickLoginForMonopoly)        // 快速登录（建议加IP限制）
r.Get("/api/games/random-for-monopoly", GetRandomGamesForMonopoly)  // 随机游戏列表
r.Get("/games/launch-monopoly", LaunchMonopolyGame)           // 直达游戏（同launch3风格）
```

---

## 测试用例

### 测试1：快速登录
```bash
# 正常情况
curl -X POST http://localhost:3000/api/auth/quick-login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser123"}'

# 期望返回: token字段
```

### 测试2：随机游戏列表
```bash
curl http://localhost:3000/api/games/random-for-monopoly

# 期望返回: 5个游戏对象，包含game_code, name, img_url
```

### 测试3：直达游戏
```bash
# 先用测试1的token
curl "http://localhost:3000/games/launch-monopoly?token=xxx&game_code=vs20olympgate&is_mobile=false"

# 期望返回: HTML页面（包含游戏iframe）
```

---

## 下一步

确认方案后，实施代码：
1. 在 `thirdprovider.go` 追加三个接口实现（~150行）
2. 在 `router.go` 注册路由（3行）
3. 提供完整测试curl命令

如需调整（如参数名、Token有效期、返回字段），请告知。
