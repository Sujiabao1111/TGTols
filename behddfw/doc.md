

```markdown
# 游戏中心后端 API 文档

**版本**: v1.0.0
**基础 URL**: `https://hddfw.vazhenina.com/hddfwother/`
**API 前缀**: `/api/v1` (用户端) / `/admin` (管理端)
**鉴权方式**: Header 中添加 `Authorization: Bearer <token>`
**数据格式**: `Content-Type: application/json`

---

## 1. 用户认证 (Auth)

### 1.1 用户注册 (Register)
注册成功后会自动返回 Token，无需二次登录。支持分销关系绑定。

- **URL**: `/api/reg`
- **Method**: `POST`
- **Auth**: No
- **Body**:
  ```json
  {
    "username": "player001",
    "password": "password123",
    "invite_code": "8A93BE",  // 可选：推荐人邀请码
    "parent_id": 0            // 可选：推荐人ID (优先使用 invite_code)
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "注册成功",
    "token": "eyJhbGciOiJIUzI1Ni...", // JWT Token
    "data": {
      "user_id": 105,
      "username": "player001",
      "invite_code": "NEWCODE",
      "balance": 0
    }
  }
  ```

### 1.2 用户登录 (Login)
- **URL**: `/api/login`
- **Method**: `POST`
- **Auth**: No
- **Body**:
  ```json
  {
    "username": "player001",
    "password": "password123"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "登录成功",
    "token": "eyJhbGciOiJIUzI1Ni...",
    "data": {
      "user_id": 105,
      "username": "player001",
      "balance": 100.00,
      "vip_level": 1
    }
  }
  ```

### 1.3 获取个人信息 (User Profile)
- **URL**: `/games/info`
- **Method**: `GET`
- **Auth**: **Yes**
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "uid": 105,
      "username": "player001",
      "invite_code": "NEWCODE",
      "balance": 150.50, // 本地钱包余额
      "vip_level": 1,
      "level": 2         // 分销层级
    }
  }
  ```

### 1.4 获取推广信息 (Invite Info)
用于推广中心页面，展示邀请码、链接及团队数据。

- **URL**: `/games/inviteinfo`
- **Method**: `GET`
- **Auth**: **Yes**
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "invite_code": "NEWCODE",
      "share_link": "https://yoursite.com/register?code=NEWCODE",
      "direct_count": 5,   // 直推人数
      "team_count": 120    // 团队总人数
    }
  }
  ```

---

## 2. 游戏业务 (Game)

### 2.1 获取筛选选项 (Game Options)
一次性返回所有可用的“厂商列表”和“游戏分类列表”，用于前端渲染筛选下拉框。

- **URL**: `/games/options`
- **Method**: `POST`
- **Auth**: No
- **Body**: `{}` (空对象)
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "providers": [
        { "id": 1, "code": "PG", "name": "PG Soft" },
        { "id": 2, "code": "JILI", "name": "Jili Games" }
      ],
      "game_types": [
        { "id": 1, "code": "SLOT", "name": "Slots", "sort": 1 },
        { "id": 2, "code": "LIVE", "name": "Live Casino", "sort": 2 }
      ]
    }
  }
  ```

### 2.2 获取主游戏列表 (Main List)
支持分页、按厂商/分类筛选、按名称模糊搜索。

- **URL**: `/games/list`
- **Method**: `POST`
- **Auth**: No
- **Body**:
  ```json
  {
    "page": 1,
    "page_size": 20,
    "game_type_id": "SLOT", // 可选: "SLOT", "LIVE", "ALL"
    "game_provider": "PG",  // 可选: "PG", "ALL"
    "name": "Mahjong"       // 可选: 搜索关键词 (支持中英文)
  }
  ```
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "list": [
        {
          "id": 101,
          "game_code": "mahjong-ways",
          "name": {"CN": "麻将胡了", "EN": "Mahjong Ways"},
          "img_url": "https://img.com/mw.png",
          "provider": { "code": "PG", "name": "PG Soft" },
          "status": 1
        }
      ],
      "total": 50,
      "page": 1,
      "total_pages": 3
    }
  }
  ```

### 2.3 获取特色榜单 (Featured Games)
返回“热门游戏(Top Views)”和“推荐游戏(Top Sort)”。建议在页面初始化时并发调用。

- **URL**: `/games/featured`
- **Method**: `GET`
- **Auth**: No
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "hot_games": [ ... ],       // 浏览量前20
      "recommend_games": [ ... ]  // 权重前20
    }
  }
  ```

### 2.4 进入游戏 (Launch Game)
包含 **智能免转** 逻辑：如果本地有余额，会自动全额转入三方钱包。若未注册三方账号，会自动静默注册。

- **URL**: `/games/launch2`
- **Method**: `POST`
- **Auth**: **Yes**
- **Body**:
  ```json
  {
    "game_code": "vs20olympgate",
    "is_mobile": true,
    "language": "zh_CN"
  }
  ```
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "url": "https://gci.usplaynet.com/player/login/..." // 游戏启动链接
    }
  }
  ```

### 2.5 进入厂商大厅 (Open Lobby)
跳转/获取到厂商的聚合大厅页。

- **URL**: `/games/lobby2`
- **Method**: `POST`
- **Auth**: **Yes**
- **Body**:
  ```json
  {
    "provider_code": "PG",
    "is_mobile": true,
    "language": "zh_CN"
  }
  ```
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "url": "https://gci.usplaynet.com/lobby/..."
    }
  }
  ```

---

## 3. 资金与钱包 (Wallet)

### 3.1 一键回收余额 (Recycle Balance)
从三方统一钱包查询余额，如果有钱，全部转回本地钱包。建议在“退出游戏”或“进入个人中心”时调用。

- **URL**: `/games/recycle`
- **Method**: `POST`
- **Auth**: **Yes**
- **Body**: `{}`
- **Response**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "recycled_amount": 500.00,    // 本次从游戏回收的金额
      "current_balance": 1500.00    // 回收后本地总余额
    }
  }
  ```

### 3.2 本地充值 (Deposit)
*模拟接口，实际应对接支付网关。*

- **URL**: `/wallet/deposit`
- **Method**: `POST`
- **Auth**: **Yes**
- **Body**:
  ```json
  { "amount": 1000, "channel": "usdt" }
  ```
- **Response**: `{"code": 0, "message": "充值成功"}`

### 3.3 本地提现申请 (Withdraw)
- **URL**: `/wallet/withdraw`
- **Method**: `POST`
- **Auth**: **Yes**
- **Body**:
  ```json
  { "amount": 500, "bank_info": "Bank of China, 123456" }
  ```
- **Response**: `{"code": 0, "message": "申请已提交"}`

---

## 4. 内容展示 (CMS)

### 4.1 获取首页 Banner
返回结构支持多态（带按钮、带跳转链接、纯展示），前端需根据字段判断。

- **URL**: `/games/banners`
- **Method**: `GET`
- **Auth**: No
- **Response**:
  ```json
  {
    "code": 0,
    "data": [
      {
        "id": 1,
        "title": "Welcome Bonus",
        "image": "https://img.com/b1.jpg",
        "buttons": [ // 形式1: 带按钮
          { "label": "Deposit", "action": "/wallet", "primary": true }
        ]
      },
      {
        "id": 2,
        "title": "Activity",
        "image": "https://img.com/b2.jpg",
        "jumpLink": "/activity" // 形式2: 点击图片跳转
      }
    ]
  }
  ```

---

## 5. 管理员后台 (Admin)

需 Header 携带管理员 Token。

### 5.1 发送站内信
- **URL**: `/admin/mail/send`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "user_ids": [101, 102],
    "title": "维护补偿",
    "content": "补偿 100 金币",
    "attachments": [{"type": "balance", "value": 100}]
  }
  ```

### 5.2 修改游戏状态
- **URL**: `/admin/games/:id/status`
- **Method**: `PATCH`
- **Body**: `{"status": 0}` (0:下架, 1:上架)

### 5.3 修改厂商状态
禁用厂商后，该厂商下所有游戏在前端不可见。

- **URL**: `/admin/providers/:id/status`
- **Method**: `PATCH`
- **Body**: `{"status": 0}`

### 5.4 获取留存报表
- **URL**: `/admin/stats/retention`
- **Method**: `GET`
- **Response**:
  ```json
  {
    "code": 0,
    "data": [
      {
        "stat_date": "2023-10-27",
        "new_users": 100,
        "retention_1": 45, // 次日留存数
        "retention_3": 30
      }
    ]
  }
  ```
```