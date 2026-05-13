# z-uc — 轻量跨集群用户中心

## 定位

微信是超级 UC（掌管 openid/session_key），z-uc 是开发者自己的轻量身份层。  
负责 unionid → 用户画像聚合 → JWT 签发，不替代微信做身份认证。

## 模块

```
z-uc/
├── models/                共享类型、JWT、AES 密钥派生
├── user/    (port 4461)   注册、资料、密码、注销
├── login/   (port 4462)   表单登录、Token 续期
├── auth/    (port 4463)   第三方登录、Token 验证、mini program 双令牌
├── remote/                服务间调用客户端 (gobreaker)
├── html/                  前端页面 (signin/signup/access/google/checkoutnow)
└── migrations/            建表 SQL
```

## API 端点

### user (4461)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/register` | 注册 (USERNAME / EMAIL / TEL / WX_UNION) |
| GET | `/user/profile` | 获取资料 (需 Authorization Bearer) |
| PUT | `/user/profile` | 更新资料 |
| PUT | `/user/password` | 修改密码 |
| DELETE | `/user/account` | 注销账号 |
| GET | `/internal/user/sn/:sn` | 内部查询(按 sn) |
| GET | `/internal/user/id?userId=` | 内部查询(按 id) |
| GET | `/internal/user/unionid/:unionId` | 内部查询(按微信 unionId) |

### login (4462)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/login` | 表单登录 { loginName, password } |
| POST | `/token/refresh` | 刷新 Token { refreshToken } |

### auth (4463)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/state` | 获取 OAuth state (防 CSRF) |
| POST | `/auth/google/token` | Google OAuth 登录 { code, redirectUri, state } |
| POST | `/auth/wx/token` | 微信公众号登录 { code, state } |
| POST | `/auth/wx-miniapp/token` | 小程序登录 (双令牌) { code, appId, nickName, icon, gender } |
| GET | `/auth/token/verify` | 验证 Token → 返回 claims |
| GET | `/auth/info` | 从 Token 获取完整用户信息 |

## 配置

每个服务独立配置：

```
{service}/config/
├── application-test.yml    ← 端口、MySQL、Redis
└── apps.yml                ← 小程序 app 列表 (ConfigMap mount)
```

### auth/config/apps.yml

```yaml
apps:
  - id: "mini-app-example"
    token_url: "http://mini-app-example.svc.cluster.local:8080/internal/token/generate"
```

apps.yml 作为 K8s ConfigMap 挂载，跨集群各自配置，改 app 不重部署。

## 启动

```bash
cd user  && go run ./cmd/
cd login && go run ./cmd/
cd auth  && go run ./cmd/
```

## 认证流程

### 表单登录

```
POST /login  { loginName, password }
  → t_names 查 loginName → user_id
  → t_user 查 user → bcrypt 验证
  → 返回: { access_token, refresh_token, token_type: "Bearer", expires_in, user }
```

### Google OAuth

```
login.html (点 Google 图标)
  → google/google.html (取 state)
  → Google 授权页
  → access/access.html (收 code+state)
  → POST /auth/google/token { code, state, redirectUri }
  → auth 校验 state (Redis GetDel)
  → exchange Google code → id_token / userinfo
  → signUpOrLoginByThird("google", email, ...)
  → 签发 JWT
```

### 微信小程序 — 双令牌

```
wx.login() → code
  → POST /auth/wx-miniapp/token { code, appId, nickName, icon, gender }

auth 处理:
  ① jscode2session → openid, unionid, session_key
  ② 查/建用户
  ③ 签发 x_token (z-uc 全域 JWT)
  ④ findApp(appId) → POST {token_url} 
     { sn, openId, unionId, userId, nickName, icon, cipher, timestamp }
     cipher = AES-256-GCM( 派生key, unionId )  ← app server 解密比对
  ⑤ 返回: { x_token, x_refresh_token, access_token, token_type, expires_in, user }

app server 验证:
  key = HMAC-SHA256(JwtSecret, "z-uc:app:" + appId)
  decrypt(cipher, key) == unionId  → 验证通过 → 签发自己的 token
```

## Token 格式

```json
{
  "iss": "z-uc",
  "sub": "U2711-03abcdef12",
  "iat": 1715432000,
  "exp": 1718024000,
  "sn": "U2711-03abcdef12",
  "name": "alice",
  "nickName": "Alice",
  "icon": "https://...",
  "email": "alice@x.com",
  "gender": "F",
  "createFrom": "USERNAME",
  "location": "CN",
  "userId": 10001
}
```

access_token 有效期 30 天，refresh_token 有效期 365 天。

## 密钥派生 (BIP32 模式)

```
master_secret = JwtSecret

app_key = HMAC-SHA256(master_secret, "z-uc:app:<appId>")  → 32 bytes

Go:   models.DeriveAppSecret("mini-app-example")
Python: python subsecret_generator.py mini-app-example
```

两边独立计算，不需要在配置文件里存 app secret。

## app 间认证（加密比对）

```
z-uc → app server:

{
  sn, openId, unionId, userId, nickName, icon,
  cipher: AES-256-GCM(key=app_key, plaintext=unionId),  ← 加密 unionId
  timestamp
}

app server 验证:

decrypted = AES-256-GCM_DECRYPT(key=app_key, ciphertext=cipher)
decrypted == unionId  → 请求来自 z-uc，合法
```

## 数据库

```sql
t_user    (id, sn, name, password, nick_name, icon, gender, birth,
           create_from, location, city, wx_union_id, email, tel, create_at,
           account_non_expired, account_non_locked, credentials_non_expired, enabled)

t_names   (login_name PK, user_id, app_id, create_at)
```

一个用户可以有多条 t_names (用户名/邮箱/手机/微信 unionId/Google email)，都映射到同一个 user_id。

## HTML 页面

| 路径 | 功能 |
|------|------|
| `/signin/login.html` | 登录页（表单 + 微信/Google 图标） |
| `/signup/signup.html` | 注册页（含协议复选框） |
| `/access/access.html` | 授权确认页（OAuth 回调 + 用户信息展示） |
| `/google/google.html` | Google OAuth 跳转（取 state → redirect） |
| `/connect/connect.html` | 第三方登录连接路由（微信） |
| `/checkoutnow/index.html` | 订单审批页（店铺老板确认放货） |

## 依赖

```
gin, sqlx, mysql, xb (SQL builder), x/crypto (bcrypt),
golang-jwt, viper, fsnotify, resty, redis/go-redis,
sony/gobreaker (remote)
```

Go 1.25.5
