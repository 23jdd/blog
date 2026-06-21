# Blog 后端服务

一个基于 Go + Gin 的博客后端项目，提供用户认证、文章发布、评论互动、点赞收藏、草稿管理、文件上传、动态配置等能力。

## 功能概览

- 用户体系：注册、登录、刷新令牌、登出、鉴权校验
- 邮箱验证码：发送注册验证码（可通过配置开关控制）
- 文章模块：创建、更新、删除、按作者/标签/分类查询、热门文章
- 互动模块：评论、评论审核、点赞、取消点赞、收藏、取消收藏
- 用户侧聚合：我的收藏、个性化 Feed
- 草稿模块：保存、列表、更新、删除草稿
- 文件模块：头像上传、文章内容文件上传
- 动态配置：通过 etcd 在线更新限流等配置

## 技术栈

- Web 框架：`Gin`
- 配置管理：`Viper`
- 数据库：`MySQL`（`sqlx`）
- 缓存/会话：`Redis`
- 服务配置中心：`etcd`
- 身份认证：`JWT`

## 目录结构（核心）

```text
.
├─ main.go                 # 程序入口与路由注册
├─ config.yaml             # 本地配置文件
├─ internal/
│  ├─ handlers/            # HTTP 处理器
│  ├─ sql/                 # 数据访问层（MySQL）
│  ├─ redis/               # Redis 连接与业务缓存
│  ├─ etcd/                # etcd 客户端
│  ├─ Middle/              # 中间件（鉴权、限流、日志、跨域）
│  ├─ model/               # 数据模型
│  ├─ types/               # 请求/响应类型定义
│  └─ utils/               # 工具函数（JWT、密码、校验等）
└─ uploads/                # 上传文件目录（运行时生成）
```

## 运行环境

建议版本：

- Go `>= 1.25`
- MySQL `>= 8.0`
- Redis `>= 6`
- etcd `>= 3.5`

## 快速开始

1. 克隆并进入项目目录

```bash
git clone <your-repo-url>
cd Blog
```

2. 安装依赖

```bash
go mod tidy
```

3. 准备基础服务（MySQL / Redis / etcd）

- MySQL 需可访问并存在数据库 `data`
- Redis 需可访问
- etcd 默认读取 `localhost:2379`

4. 配置 `config.yaml`（见下文“配置说明”）

5. 启动服务

```bash
go run main.go
```

默认监听端口：`8080`

## 前端部署（Go Gin 托管）

后端会在运行时优先托管 `front/dist` 目录；如果你还没构建 `dist`，则会临时托管 `front/` 下的开发版页面（便于联调）。

1. 构建前端

```bash
cd front
npm install
npm run build
```

2. 启动后端

```bash
cd ..
go run main.go
```

启动后访问：`http://localhost:8080`

## 配置说明

项目使用 `Viper` 读取 `config.yaml`，主要配置如下：

```yaml
server:
  port: 8080

limit:
  read: 100
  rate: 1

smtp:
  host: "smtp.example.com"
  port: 465
  username: "your_email@example.com"
  password: "your_smtp_password"
  from: "your_email@example.com"

upload:
  dir: "uploads"
  url_prefix: "/uploads"

feature:
  enable_email_verify: true
  enable_markdown_api: true
```

另外可选配置（未写入 `config.yaml` 时会使用默认值）：

- `jwt.secret`：JWT 密钥（强烈建议在生产环境覆盖默认值）
- `jwt.access_expire_minutes`：访问令牌过期分钟数（默认 `30`）
- `jwt.refresh_expire_hours`：刷新令牌过期小时数（默认 `168`）
- `redis.addr` / `redis.password` / `redis.db` 等：Redis 连接参数

## 数据库说明

- 程序启动时会自动连接 MySQL（当前代码默认连接串：`root:1234@tcp(127.0.0.1:3306)/data?parseTime=true`）
- 首次启动会自动确保核心表存在：`user`、`article`、`review`、`collect`、`article_like`、`draft`

> 建议你后续将数据库连接串改为配置文件或环境变量方式，避免硬编码。

## 接口概览

### 认证 ` /auth `

- `POST /auth/login`：登录
- `POST /auth/register`：注册
- `POST /auth/refresh`：刷新令牌
- `GET /auth/verification/send`：发送验证码
- `GET /auth/judgeToken`：校验令牌有效性（需鉴权）
- `POST /auth/logout`：登出（需鉴权）

### 文章 ` /articles `

- `GET /articles/:id`：文章详情
- `GET /articles/author/:authid`：作者文章列表
- `GET /articles/by-tag`：按标签筛选
- `GET /articles/category/:categoryID`：按分类筛选
- `GET /articles/:id/comments`：评论列表
- `GET /articles/:id/stats`：文章统计
- `GET /articles/hot`：热门文章
- `POST /articles`：创建文章（需鉴权）
- `PUT /articles/:id`：更新文章（需鉴权）
- `PATCH /articles/:id/status`：更新文章状态（需鉴权）
- `DELETE /articles/:id`：删除文章（需鉴权）

### 互动 ` /articles/* + /interactions/* `

- `POST /articles/:id/comments`：发表评论（需鉴权）
- `DELETE /articles/comments/:commentID`：删除评论（需鉴权）
- `PATCH /articles/comments/:commentID/status`：审核评论（需鉴权）
- `POST /articles/:id/likes`：点赞（需鉴权）
- `DELETE /articles/:id/likes`：取消点赞（需鉴权）
- `POST /articles/:id/collections`：收藏（需鉴权）
- `DELETE /articles/:id/collections`：取消收藏（需鉴权）
- `GET /interactions/my-collections`：我的收藏（需鉴权）
- `GET /interactions/feed`：我的 Feed（需鉴权）

### 草稿 ` /drafts `

- `POST /drafts`：保存草稿（需鉴权）
- `GET /drafts`：草稿列表（需鉴权）
- `PUT /drafts/:id`：更新草稿（需鉴权）
- `DELETE /drafts/:id`：删除草稿（需鉴权）

### 其他

- `POST /file/setPersonImage`：上传头像（需鉴权）
- `POST /file/uploadArticle`：上传文章文件（需鉴权）
- `GET /config`：查询动态配置（需鉴权）
- `POST /config`：更新动态配置（需鉴权）
- `POST /markdown`：Markdown 转 HTML（由功能开关控制）

## 鉴权方式

需要鉴权的接口请在请求头中携带：

```text
Authorization: Bearer <access_token>
```

## 日志与静态文件

- 服务日志输出到根目录 `log.txt`
- 上传目录由 `upload.dir` 控制，静态访问前缀由 `upload.url_prefix` 控制

## 常见问题

1. 启动即报 MySQL 连接失败  
   请检查 MySQL 是否启动、`data` 数据库是否存在，以及账号密码是否正确。

2. 登录/注册报 Redis 相关错误  
   请检查 Redis 地址、密码配置，并确认服务可达。

3. 动态配置接口报错  
   请确认 etcd 服务运行正常，且地址与代码一致（默认 `localhost:2379`）。



