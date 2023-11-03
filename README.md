## Task0

新人任务

---

一个简易用户管理系统，没有前端

框架 echo，鉴权 jwt，使用 cookie 存 token

数据库 postgresql

需要提前配置好数据库，启动服务，并相应修改连接数据库的代码。

参考 `db/db.go` 里这一行

```go
var dsn = "host=localhost port=5432 user=postgres dbname=byddb sslmode=disable"
```

---

### 启动！

你可能需要先 `go mod tidy` 来安装依赖包

然后 `go run .` 启动后端，端口 `11451`（臭口

终端里会有一些 log 输出

### 交互方式

可以通过在浏览器地址栏输入 url + 参数，浏览器自己会管理 cookie

或者 `curl` 发送请求，此时由于鉴权需要用 cookie 存的 token，你要手动设置 cookie 文件

`curl -b <文件>` 可以使请求携带某文件作为 cookie

`curl -c <文件>` 指示获得的响应中的 cookie 写入该文件中

个人一般写成这样的命令 `curl -b ck.cookie -c ck.cookie <其他参数> `

---

下面提到的所有路由都不限请求方法，GET，POST 都可以

请求参数支持以下几种形式发送，例如：

```
// JSON 数据
curl -X POST http://localhost:1323/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Joe","email":"joe@labstack"}'
```

```
// Form 表单数据
curl -X POST http://localhost:1323/users \
  -d 'name=Joe' -d 'email=joe@labstack.com'
```

```
// url+参数
curl -X GET \
  'http://localhost:1323/users?name=Joe&email=joe@labstack.com'
```

以上仅是例子，不代表具体路由和参数

不同形式的请求参数名称一致，效果相同

响应统一用 JSON 数据

### 注册用户

访问 `localhost:11451/register` 注册

参数：

- `name` 用户名。不能为空。若该用户名已被注册会注册失败
- `email` 可空
- `passwd` 密码。不能为空
- `repeat` 重复密码。必须与密码相同

注册成功后，对于一个用户还会自动生成：

-  `uid` 数据库中记录的主值
- `auth` 用户权限。新注册的用户统一为 `normal`

注册成功后不会自动登录，需要去登录用户才能访问主页

示例命令：

```
curl -b ck.cookie -c ck.cookie -X POST \
  'localhost:11451/register?name=xiwon&passwd=123&repeat=123&email=xiwon'
```

### 登录

访问 `localhost:11451/login` 登录

参数：

- `name` 用户名。不能为空。若没有找到该用户则报错
- `passwd` 密码。若密码错误则报错

示例命令：

```
curl -b ck.cookie -c ck.cookie -X POST \
  'localhost:11451/login?name=xiwon&passwd=123'
```

如果登录成功则会生成一个 token 写到你本地的 cookie 中，生成方式依照标准 jwt

每个新生成的 token 有效时间为 1 分钟，在此期间保持登录状态，可以以该用户的身份进行各种操作

在不同的登录状态下进行 login 会有不同的回复：

- 如果处于未登录状态，响应为 `you've login in user <username> from anonymous`
- 如果已登录某一另外的账号，响应为 `you've switch to user <u2> from <u1>` 
- 如果你再次登录自己的账号，响应为 `you've flashed your cookie`

任何一种合法的 login 都会刷新本地的 token

### 主页

访问 `localhost:11451/` 获取主页

无参数

- 一切正常时，返回一句话和你的个人资料
- 未登录状态下访问，响应 `you did't login, or your token had expired`
- 如果你的账户已经被删除，但是你的 token 还没过期，则会响应 `deleted user`

示例命令：

```
curl -b ck.cookie -c ck.cookie -X GET 'localhost:11451/'
```

### 登出

访问 `localhost:11451/logout` 来登出账户

无参数

- 正常情况返回 `logout from user <username>`，并清空你的 cookie
- 若未登录时 logout，响应 `logout failed, you didn't login`

示例命令：

```
curl -b ck.cookie -c ck.cookie -X GET 'localhost:11451/logout'
```

### 修改权限

uid 为 1 的用户为 root，root 每次启动后端时都会被再次写入 `admin` 权限

如果用户 profile 的 Auth 为 `admin`，则该用户有管理权限

访问 `localhost:11451/setauth` 修改某用户的权限

参数：

- `name` 被修改的用户名。如果不存在响应 `empty target user`
- `to` 权限字段被修改为的值。不能为空

只有具有 `admin` 权限的用户可以修改别人的 Auth，如果 Auth 字段不为 `admin` 则响应 `you are not an admin`。

未登录访问则报错

示例命令：

```
curl -b ck.cookie -c ck.cookie -X GET \
  'localhost:11451/setauth?name=adm&to=admin'
```

### 修改信息

访问 `localhost:11451/setinfo` 修改某用户的信息

**参数与 register 相同**。未填写的字段则不做修改

未登录访问会报错，不存在该用户则报错

管理员可以修改除 root 以外其他任何人的 info

只有 root 能修改 root 的 info

普通用户只能修改自己的 info

可以修改的字段有 passwd，email

若操作合法则响应信息会返回修改了什么

示例命令：

```
curl -b ck.cookie -c ck.cookie -X GET \
  'localhost:11451/setinfo?name=xiwon&email=aaaaa'
```

### 删除用户

访问 `localhost:11451/delete` 删除用户

参数：

- `name` 要删除的用户名。不能为空

未登录访问则报错，用户不存在则报错

管理员能删除 root 以外其他人

**任何用户都不能删除 root**

普通用户可以注销自己







