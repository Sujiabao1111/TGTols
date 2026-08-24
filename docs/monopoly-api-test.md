# 大富翁游戏接口测试文档

## 接口清单

| 接口 | 地址 | 方法 | 说明 |
|------|------|------|------|
| 快速登录 | `/api/auth/quick-login` | POST | 12661服务端调用，获取用户Token |
| 随机游戏 | `/api/games/random-for-monopoly` | GET | 获取5个随机游戏信息 |
| 直达游戏 | `/games/launch-monopoly` | GET | 直接进入指定游戏 |

---

## 测试用例

### 1. 快速登录接口

**请求：**
```bash
curl -X POST http://localhost:3000/api/auth/quick-login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_monopoly_user_001"
  }'
```

**成功响应：**
```json
{
  "code": 0,
  "message": "success",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 300,
  "user_id": 123
}
```

**失败响应：**
```json
{
  "code": 400,
  "message": "username is required"
}
```

**说明：**
- 如果用户不存在，会自动创建
- Token有效期5分钟（300秒）
- 建议12661缓存token，在过期前复用

---

### 2. 随机游戏列表接口

**请求：**
```bash
# 默认返回5个
curl http://localhost:3000/api/games/random-for-monopoly

# 或指定数量（最多20个）
curl "http://localhost:3000/api/games/random-for-monopoly?count=10"
```

**成功响应：**
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
          "CN": "奥林匹斯之门",
          "ZH": "奧林匹斯之門"
        },
        "img_url": "https://example.com/img/olympus.png",
        "provider_code": "PG",
        "provider_name": "Pragmatic Play"
      },
      {
        "game_code": "vswayszombiewave",
        "name": {
          "EN": "Zombie Carnival",
          "CN": "僵尸嘉年华"
        },
        "img_url": "https://example.com/img/zombie.png",
        "provider_code": "PG",
        "provider_name": "Pragmatic Play"
      }
    ]
  }
}
```

**说明：**
- 建议12661在启动时或定时调用此接口缓存游戏列表
- 大富翁格子使用 `img_url` 展示图标，`name.EN` 或 `name.CN` 展示名称
- `game_code` 需要保存，用于后续直达游戏接口

---

### 3. 直达游戏接口

**请求：**
```bash
# 基础调用
curl "http://localhost:3000/games/launch-monopoly?token=xxx&game_code=vs20olympgate"

# 移动端
curl "http://localhost:3000/games/launch-monopoly?token=xxx&game_code=vs20olympgate&is_mobile=true"
```

**成功响应：**
返回HTML页面，浏览器会直接渲染游戏iframe

**失败响应（HTML）：**
```html
<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body>
  <h1>资金同步失败，请重试</h1>
</body>
</html>
```

**说明：**
- token来自快速登录接口
- game_code来自随机游戏列表接口
- 返回的是完整HTML页面，可直接在浏览器/游戏内嵌webview中打开

---

## 完整流程测试

### 场景：用户踩中格子进入游戏

```bash
# 步骤1: 12661获取随机游戏列表（启动时）
GAMES=$(curl -s http://localhost:3000/api/games/random-for-monopoly)
echo "随机游戏: $GAMES"

# 步骤2: 用户踩中格子时，12661先获取token
TOKEN_RESP=$(curl -s -X POST http://localhost:3000/api/auth/quick-login \
  -H "Content-Type: application/json" \
  -d '{"username": "player_12345"}')
TOKEN=$(echo $TOKEN_RESP | jq -r '.token')
echo "Token: $TOKEN"

# 步骤3: 构造游戏URL给玩家
GAME_CODE="vs20olympgate"  # 来自步骤1的列表
GAME_URL="http://localhost:3000/games/launch-monopoly?token=${TOKEN}&game_code=${GAME_CODE}"
echo "游戏链接: $GAME_URL"

# 步骤4: 玩家浏览器/客户端打开此URL即可进入游戏
```

---

## 错误处理

### 常见错误码

| 接口 | 错误场景 | 返回 |
|------|----------|------|
| quick-login | username为空 | code=400 |
| quick-login | 数据库错误 | code=500 |
| random-for-monopoly | 数据库错误 | code=500 |
| launch-monopoly | token无效 | HTML错误页 |
| launch-monopoly | game_code不存在 | HTML错误页 |
| launch-monopoly | 12661同步失败 | HTML错误页 |
| launch-monopoly | Hedoc转入失败 | HTML错误页 |

### 12661侧建议处理

1. **Token过期**：如果调用launch-monopoly返回认证失败，重新调quick-login获取新token
2. **游戏进入失败**：显示友好提示，让用户重试或选择其他游戏
3. **网络超时**：建议设置10秒超时，超时后提示用户检查网络

---

## 安全建议

### 1. IP白名单
建议在Nginx层限制 `/api/auth/quick-login` 仅允许12661服务IP访问：

```nginx
location /api/auth/quick-login {
    allow 127.0.0.1;  # 本机
    allow 10.0.0.0/8; # 内网
    deny all;
    
    proxy_pass http://backend;
}
```

### 2. Token使用原则
- 仅用于单次游戏进入
- 不要在客户端长期存储
- 5分钟过期后需重新获取

### 3. 参数校验
- username长度建议限制32字符以内
- 只允许字母数字下划线

---

## 集成检查清单

- [ ] quick-login接口可正常访问
- [ ] random-for-monopoly返回5个游戏
- [ ] launch-monopoly返回有效HTML
- [ ] 游戏内余额正确同步（12661→Hedoc）
- [ ] 游戏可正常游玩
- [ ] IP白名单已配置
- [ ] 错误场景已测试
