# 大富翁游戏接入接口文档

**文档版本**: v1.4  
**更新日期**: 2026-03-20  
**适用方**: 12661游戏端  
**接口基地址**: `http://<服务器地址>:3000`

---

## 概述

本文档描述12661大富翁游戏接入游戏中心（HDD游戏大厅）的接口规范。

---

## 接入流程

```
用户踩中格子
    ↓
12661客户端 -> 12661服务端
    ↓
12661服务端调用 POST /api/auth/quick-login
    ├─ 请求体: {username: "player_123"}
    ↓
获得Token（5分钟有效）
    ↓
Token下发给客户端
    ↓
客户端WebView直接打开URL:
/games/launch-monopoly?token=xxx&game_code=xxx&is_mobile=false&language=CN
    ↓
进入游戏
```

---

## 接口列表

### 1. 快速登录接口（服务端调用）

**接口信息**
- **地址**: `POST /api/auth/quick-login`
- **Content-Type**: `application/json`
- **调用方**: 12661服务端（非前端）

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户唯一标识 |

**请求示例**
```json
{
  "username": "player_123456"
}
```

**响应参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，0为成功 |
| token | string | JWT访问令牌（5分钟有效） |
| expires_in | int | 有效期300秒 |

**成功响应**
```json
{
  "code": 0,
  "message": "success",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 300,
  "user_id": 123
}
```

---

### 2. 获取随机游戏列表

**接口信息**
- **地址**: `GET /api/games/random-for-monopoly`
- **调用方**: 12661服务端

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| count | int | 否 | 返回数量，默认5，最大20 |

**请求示例**
```
GET /api/games/random-for-monopoly?count=5
```

**响应示例**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "games": [
      {
        "game_code": "vs20olympgate",
        "name": {"EN": "Gates of Olympus", "CN": "奥林匹斯之门"},
        "img_url": "https://example.com/games/olympus.png",
        "provider_code": "PG",
        "provider_name": "Pragmatic Play"
      }
    ]
  }
}
```

---

### 3. 直达游戏接口（前端WebView打开）

**接口信息**
- **地址**: `GET /games/launch-monopoly`
- **调用方**: 12661客户端WebView直接打开

**请求参数**（Query String）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | JWT令牌（来自quick-login） |
| game_code | string | 是 | 游戏代码 |
| is_mobile | bool | 否 | 是否移动端，默认false |
| language | string | 否 | 语言，默认"EN" |

**请求示例**
```
GET /games/launch-monopoly?token=eyJhbGciOiJIUzI1NiIs...&game_code=vs20olympgate&is_mobile=false&language=CN
```

**响应说明**

此接口返回**HTML页面**，不是JSON。WebView直接加载返回的内容即可进入游戏。

**成功响应**
- Content-Type: `text/html; charset=utf-8`
- Body: 完整HTML页面（包含游戏iframe）

**失败响应**
HTML错误页面。

**注意事项**
1. **这是GET接口**，供WebView直接打开链接使用
2. Token 5分钟后过期，只能使用一次
3. 如果Token过期，需要重新调用quick-login获取新Token

---

## 完整业务流程

### 阶段1：初始化
1. 12661服务端定时调用 `GET /api/games/random-for-monopoly`
2. 缓存游戏列表（game_code, name, img_url）
3. 大富翁格子展示

### 阶段2：用户踩中格子
1. 用户踩中格子
2. 12661客户端通知服务端
3. 12661服务端调用 `POST /api/auth/quick-login` 获取Token
4. 12661服务端将Token下发给客户端
5. 12661客户端构造URL并打开WebView：
   ```
   /games/launch-monopoly?token=xxx&game_code=xxx&is_mobile=false&language=CN
   ```
6. 用户进入游戏

---

## 代码示例

### 12661服务端（Go）

```go
// 快速登录获取Token
func GetGameToken(username string) (string, error) {
    reqBody, _ := json.Marshal(map[string]string{
        "username": username,
    })
    
    resp, err := http.Post(
        "http://game-center/api/auth/quick-login",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result struct {
        Code  int    `json:"code"`
        Token string `json:"token"`
    }
    
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.Code != 0 {
        return "", errors.New("login failed")
    }
    
    return result.Token, nil
}

// 用户踩中格子时
func OnUserStepGameCell(username, gameCode string) (string, error) {
    token, err := GetGameToken(username)
    if err != nil {
        return "", err
    }
    
    // 构造游戏URL
    gameURL := fmt.Sprintf(
        "http://game-center/games/launch-monopoly?token=%s&game_code=%s&is_mobile=false&language=CN",
        url.QueryEscape(token),
        url.QueryEscape(gameCode),
    )
    
    return gameURL, nil
}
```

### 12661客户端（JavaScript）

```javascript
// 用户踩中格子
async function onStepGameCell(gameCode) {
    // 1. 向12661服务端请求游戏URL
    const res = await fetch('/api/get-game-url', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({gameCode})
    });
    
    const {gameURL} = await res.json();
    
    // 2. WebView直接打开URL
    window.open(gameURL, '_blank');
    
    // 或者在游戏内嵌WebView中加载
    // webview.src = gameURL;
}
```

### cURL测试

```bash
# 1. 快速登录（服务端调用）
curl -X POST http://localhost:3000/api/auth/quick-login \
  -H "Content-Type: application/json" \
  -d '{"username": "player_123456"}'

# 2. 获取随机游戏列表
curl "http://localhost:3000/api/games/random-for-monopoly?count=5"

# 3. 直达游戏（WebView直接打开的URL）
curl "http://localhost:3000/games/launch-monopoly?token=eyJhbGci...&game_code=vs20olympgate&is_mobile=false&language=CN"
```

---

## 安全建议

### IP白名单

```nginx
location /api/auth/quick-login {
    allow 10.0.0.0/8;
    deny all;
    proxy_pass http://backend;
}
```

### 接口权限

| 接口 | 方法 | 调用方 | 保护 |
|------|------|--------|------|
| /api/auth/quick-login | POST | 12661服务端 | IP白名单 |
| /api/games/random-for-monopoly | GET | 12661服务端 | 建议缓存 |
| /games/launch-monopoly | GET | 12661客户端 | Token验证 |

---

## 变更日志

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v1.4 | 2026-03-20 | 直达游戏接口改为GET，供WebView直接打开URL |

---

## 技术支持

如有技术问题，请联系：
- 技术对接人: [填写]
- 邮箱: [填写]
