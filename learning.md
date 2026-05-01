# Go 云原生后端项目模板指南

> 本文档详细介绍 My_Blog 项目中所有企业级模板的使用方法，涵盖日志、配置、数据库、命令行参数等核心模块。看完后可直接在新项目中复用。

---

## 目录

1. [一、core/logger 日志模块](#一corelogger-日志模块)
2. [二、flags 命令行参数模块](#二flags-命令行参数模块)
3. [三、core/init_config 配置文件模块](#三coreinit_config-配置文件模块)
4. [四、core/init_db 数据库模块](#四coreinit_db-数据库模块)
5. [五、api/middleware JWT认证中间件](#五apimiddleware-jwt认证中间件)
6. [六、api/controller 控制器详解](#六apicontroller-控制器详解)
7. [七、models/model 数据模型](#七modelsmodel-数据模型)
8. [八、api/router 路由配置](#八apirouter-路由配置)
9. [九、MySQL SQL 详解](#九mysql-sql-详解)
10. [十、main.go 标准启动模板](#十maingo-标准启动模板)
11. [十一、项目架构与模块关系](#十一项目架构与模块关系)
12. [十二、快速复制模板](#十二快速复制模板)

---

## 一、core/logger 日志模块

### 1.1 模块概述

| 项目 | 说明 |
|------|------|
| 依赖库 | `go.uber.org/zap` + `github.com/natefinch/lumberjack` |
| 功能 | 高性能结构化日志，支持控制台+文件输出 |
| 全局实例 | `logger.S` (类型: `*zap.SugaredLogger`) |

### 1.2 为什么用 zap ？

| 特性 | zap | 标准 log |
|------|-----|----------|
| 性能 | 零内存分配，极快 | 有内存分配，较慢 |
| 结构化 | 原生支持 key-value | 需手动拼接 |
| 输出 | 控制台+文件+网络 | 仅控制台 |
| 日志轮转 | 原生支持 | 需第三方库 |

### 1.3 核心类型

#### Config 结构体（日志配置）

```go
type Config struct {
    Filename   string        // 日志文件路径（例："./logs/app.log"）
    MaxSize    int           // 单个日志文件最大多大，单位MB（超过则切割）
    MaxBackups int           // 最多保留几个日志文件
    MaxAge     int           // 日志保留多少天
    Compress   bool          // 是否压缩旧日志（.gz格式）
    Level      zapcore.Level // 日志输出级别
}
```

#### Config 字段详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Filename | string | `"./logs/app.log"` | 日志文件路径 |
| MaxSize | int | `10` | 单文件最大 MB，超过自动切割 |
| MaxBackups | int | `5` | 保留的旧日志文件数量 |
| MaxAge | int | `30` | 旧日志保留天数 |
| Compress | bool | `true` | 是否压缩旧日志 |
| Level | zapcore.Level | `InfoLevel` | 日志级别 |

### 1.4 全局变量

```go
var S *zap.SugaredLogger  // 全局日志实例，所有日志操作通过它
```

### 1.5 核心函数

#### Init(cfg ...Config) 初始化日志

```go
// 无参数调用：使用默认配置
logger.Init()

// 有参数调用：使用自定义配置
logger.Init(logger.Config{
    Filename:   "./logs/app.log",
    MaxSize:    10,
    MaxBackups: 5,
    MaxAge:     30,
    Compress:   true,
    Level:      zapcore.DebugLevel,
})
```

#### Sync() 刷新日志缓冲区

```go
// 程序退出前调用，确保日志写入文件
defer logger.Sync()
```

### 1.6 日志级别

| 级别 | zapcore.Level | 说明 | 典型使用场景 |
|------|---------------|------|-------------|
| Debug | `zapcore.DebugLevel` | 调试信息 | 开发环境，详细执行步骤 |
| Info | `zapcore.InfoLevel` | 一般信息 | **默认**，记录正常运行 |
| Warn | `zapcore.WarnLevel` | 警告 | 不影响运行但需关注 |
| Error | `zapcore.ErrorLevel` | 错误 | 需要关注的问题 |
| Fatal | `zapcore.FatalLevel` | 致命错误 | 程序立即退出 |

### 1.7 logger.S 日志方法详解

所有方法都在 `*zap.SugaredLogger` 类型上调用，格式：`logger.S.方法名`

#### 1.7.1 Info 系列（一般信息）

```go
// 普通日志
logger.S.Info("这是一条普通信息")
// 输出: {"level":"INFO","time":"2026-04-25T10:00:00","msg":"这是一条普通信息"}

// 格式化日志（类似 fmt.Sprintf）
username := "zhangsan"
logger.S.Infof("用户%s登录成功", username)
// 输出: {"level":"INFO","msg":"用户zhangsan登录成功"}

// 结构化日志（推荐！键值对形式，方便日志收集系统分析）
logger.S.Infow("用户登录成功",
    "username", "zhangsan",
    "ip", "192.168.1.100",
    "time_ms", 35,
)
// 输出: {"level":"INFO","msg":"用户登录成功","username":"zhangsan","ip":"192.168.1.100","time_ms":35}
```

#### 1.7.2 Error 系列（错误信息）

```go
// 普通错误
logger.S.Error("操作失败")

// 格式化错误
err := errors.New("数据库连接超时")
logger.S.Errorf("查询用户失败: %v", err)

// 结构化错误（推荐，包含上下文）
logger.S.Errorw("用户登录失败",
    "username", "zhangsan",
    "reason", "密码错误",
    "ip", "192.168.1.100",
    "attempt", 3,
)
```

#### 1.7.3 Debug 系列（调试信息）

```go
// 普通调试
logger.S.Debug("进入函数")

// 格式化调试
logger.S.Debugf("查询条件: %v", condition)

// 结构化调试
logger.S.Debugw("SQL执行",
    "sql", "SELECT * FROM users",
    "duration_ms", 25,
)
```

#### 1.7.4 Warn 系列（警告信息）

```go
logger.S.Warn("内存使用率超过80%")
logger.S.Warnf("重试次数: %d", retryCount)
logger.S.Warnw("连接超时",
    "host", "localhost",
    "timeout_ms", 5000,
)
```

#### 1.7.5 Fatal 系列（致命错误）

```go
// Fatal 会打印日志后调用 os.Exit(1) 终止程序
// 慎用！一般用于完全无法恢复的错误
logger.S.Fatal("配置文件加载失败，无法继续运行")
logger.S.Fatalf("数据库连接失败: %v", err)
logger.S.Fatalw("启动失败",
    "reason", "端口被占用",
    "port", 8080,
)
```

### 1.8 方法对比表

| 方法 | 格式 | 推荐场景 | 示例 |
|------|------|---------|------|
| `.Info()` | `logger.S.Info("msg")` | 一般信息 | `"服务启动成功"` |
| `.Infof()` | `logger.S.Infof("format", args...)` | 简单格式化 | `"用户%s登录"`, name |
| `.Infow()` | `logger.S.Infow("msg", kvs...)` | **推荐**，结构化 | `"登录"`, `"user"`, name |
| `.Error()` `.Errorf()` `.Errorw()` | 同上 | 错误信息 | 记录错误和上下文 |
| `.Debug()` `.Debugf()` `.Debugw()` | 同上 | 开发调试 | 开发环境详细日志 |
| `.Warn()` `.Warnf()` `.Warnw()` | 同上 | 警告信息 | 不影响运行的问题 |
| `.Fatal()` `.Fatalf()` `.Fatalw()` | 同上 | 致命错误 | 程序必须退出 |

### 1.9 实战示例

#### 示例1：登录函数日志

```go
func login(username, password string) error {
    // 开始登录
    logger.S.Infow("开始登录",
        "username", username,
        "time", time.Now().Format("2006-01-02 15:04:05"),
    )

    // 验证密码
    if !verifyPassword(username, password) {
        logger.S.Errorw("登录失败",
            "username", username,
            "reason", "密码错误",
            "ip", getClientIP(),
        )
        return errors.New("密码错误")
    }

    // 登录成功
    logger.S.Infow("登录成功",
        "username", username,
        "duration_ms", 35,
    )
    return nil
}
```

#### 示例2：HTTP 请求日志中间件

```go
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        logger.S.Infow("HTTP请求",
            "method", c.Request.Method,
            "path", path,
            "status", c.Writer.Status(),
            "duration_ms", time.Since(start).Milliseconds(),
            "ip", c.ClientIP(),
        )
    }
}
```

#### 示例3：数据库操作日志

```go
func GetUserByID(id uint) (*User, error) {
    var user User
    start := time.Now()

    if err := core.GetDB().First(&user, id).Error; err != nil {
        logger.S.Errorw("查询用户失败",
            "id", id,
            "error", err.Error(),
        )
        return nil, err
    }

    logger.S.Debugw("查询用户成功",
        "id", id,
        "duration_ms", time.Since(start).Milliseconds(),
    )
    return &user, nil
}
```

---

## 二、flags 命令行参数模块

### 2.1 模块概述

| 项目 | 说明 |
|------|------|
| 依赖库 | Go 标准库 `flag` |
| 功能 | 解析命令行参数 |
| 全局实例 | `flags.Opt` (类型: `*Options`) |

### 2.2 Options 结构体（所有参数）

```go
type Options struct {
    ConfigFile string // -f 配置文件路径（默认 config/dev.yaml）
    DBInit     bool   // -db 是否初始化数据库
    Version    bool   // -v 显示版本号
    Port       int    // -port 服务端口（默认 8080）
    Host       string // -host 服务绑定地址（默认 0.0.0.0）
    Mode       string // -mode 运行模式（默认 dev，选项：dev/test/release）
    LogLevel   string // -log-level 日志级别（默认 debug，选项：debug/info/warn/error）
}
```

### 2.3 Options 字段详解

| 字段 | 类型 | 命令行参数 | 默认值 | 说明 |
|------|------|-----------|--------|------|
| ConfigFile | string | `-f` | `config/dev.yaml` | 配置文件路径 |
| DBInit | bool | `-db` | `false` | 是否初始化数据库 |
| Version | bool | `-v` | `false` | 显示版本号 |
| Port | int | `-port` | `8080` | 服务端口 |
| Host | string | `-host` | `0.0.0.0` | 绑定地址 |
| Mode | string | `-mode` | `dev` | 运行模式 |
| LogLevel | string | `-log-level` | `debug` | 日志级别 |

### 2.4 核心函数

#### Parse() 解析命令行参数

```go
// 必须第一个调用！
// 调用之后才能使用 flags.Opt.XXX 获取参数值
flags.Parse()

// 获取参数值
port := flags.Opt.Port           // int 类型
mode := flags.Opt.Mode          // string 类型
configFile := flags.Opt.ConfigFile // string 类型
dbInit := flags.Opt.DBInit      // bool 类型
```

#### GetEnv(key, defaultValue string) string 读取环境变量

```go
// 读取环境变量，优先使用环境变量，没有则用默认值
envPort := flags.GetEnv("PORT", "8080")
// 如果系统有 PORT 环境变量，返回环境变量的值
// 否则返回 "8080"
```

#### PrintUsage() 打印帮助信息

```go
// 打印标准 flag 帮助信息
// 并显示示例命令
flags.PrintUsage()
```

### 2.5 快捷判断函数

| 函数 | 返回值 | 说明 |
|------|--------|------|
| `IsRelease()` | `bool` | mode == "release" 时返回 true |
| `IsDev()` | `bool` | mode == "dev" 时返回 true |
| `IsTest()` | `bool` | mode == "test" 时返回 true |

```go
// 判断运行环境
if flags.IsRelease() {
    // 生产环境逻辑
    println("生产环境")
} else if flags.IsDev() {
    // 开发环境逻辑
    println("开发环境")
} else if flags.IsTest() {
    // 测试环境逻辑
    println("测试环境")
}
```

### 2.6 GetAddr() 获取服务地址

```go
// 返回 "host:port" 格式字符串
// 用于 gin.Run() 或 http.ListenAndServe()
addr := flags.GetAddr() // 例如 "0.0.0.0:8080"
println("监听地址:", addr)
```

### 2.7 CheckRequiredFlags() 检查必需参数

```go
// 检查必需参数是否已提供
err := flags.CheckRequiredFlags([]string{"config", "port"})
if err != nil {
    println("错误:", err)
    return
}
```

### 2.8 ValidateFlags() 校验参数有效性

```go
// 自动校验 Mode 和 LogLevel 是否有效
err := flags.ValidateFlags()
if err != nil {
    println("参数错误:", err)
    return
}
```

### 2.9 启动命令大全

```bash
# 1. 默认启动（开发模式，端口8080）
myapp.exe

# 2. 指定端口
myapp.exe -port 9000

# 3. 指定运行模式
myapp.exe -mode release
myapp.exe -mode test

# 4. 指定配置文件
myapp.exe -f config/prod.yaml

# 5. 同时指定多个参数
myapp.exe -mode release -port 80 -host 0.0.0.0

# 6. 初始化数据库（单独使用，不启动服务）
myapp.exe -db

# 7. 查看版本
myapp.exe -v

# 8. 查看帮助
myapp.exe -h

# 9. 自定义日志级别
myapp.exe -log-level error

# 10. 组合使用
myapp.exe -mode release -port 8080 -host 0.0.0.0 -f config/prod.yaml -log-level error
```

### 2.10 添加新参数的方法

**什么时候需要添加新参数？**
- 需要在启动时传入特殊配置
- 需要控制程序的不同行为

**添加步骤：**

第一步：在 `flags/enter.go` 的 `Options` 结构体中添加字段
```go
type Options struct {
    // ... 现有字段不要动 ...
    Timeout int    // 添加新字段
}
```

第二步：在 `Parse()` 函数中添加绑定
```go
func Parse() {
    // ... 现有代码不要动 ...

    // 添加新参数的绑定
    flag.IntVar(&Opt.Timeout, "timeout", 30, "超时时间（秒）")
}
```

**参数类型与绑定方法对照表：**

| 参数类型 | Go 类型 | 绑定方法 | 命令行示例 |
|----------|---------|----------|-----------|
| 字符串 | string | `flag.StringVar(&ptr, "name", 默认值, "说明")` | `-name zhangsan` |
| 整数 | int | `flag.IntVar(&ptr, "port", 默认值, "说明")` | `-port 8080` |
| 布尔 | bool | `flag.BoolVar(&ptr, "debug", 默认值, "说明")` | `-debug` |

### 2.11 函数一览表

| 函数 | 签名 | 功能 |
|------|------|------|
| `Parse` | `Parse()` | 解析命令行参数（必须先调用） |
| `GetEnv` | `GetEnv(key, defaultValue string) string` | 读取环境变量 |
| `PrintUsage` | `PrintUsage()` | 打印帮助信息 |
| `IsRelease` | `IsRelease() bool` | 判断是否生产环境 |
| `IsDev` | `IsDev() bool` | 判断是否开发环境 |
| `IsTest` | `IsTest() bool` | 判断是否测试环境 |
| `GetAddr` | `GetAddr() string` | 获取 "host:port" 格式地址 |
| `CheckRequiredFlags` | `CheckRequiredFlags(required []string) error` | 检查必需参数 |
| `ValidateFlags` | `ValidateFlags() error` | 校验参数有效性 |

---

## 三、core/init_config 配置文件模块

### 3.1 模块概述

| 项目 | 说明 |
|------|------|
| 依赖库 | `github.com/spf13/viper` |
| 功能 | 读取并合并 YAML 配置文件（基础配置 + 环境配置） |
| 全局实例 | `core.GlobalConfig` (类型: `*Config`) |

### 3.2 核心设计理念

**配置分层设计**：基础配置（app.yaml）+ 环境覆盖配置（app-dev/prod/test.yaml）

```
config/
├── app.yaml        # 基础配置（所有环境共用）
├── app-dev.yaml    # 开发环境覆盖
├── app-prod.yaml   # 生产环境覆盖
└── app-test.yaml   # 测试环境覆盖
```

### 3.3 配置结构体

#### SystemConfig（服务配置）

```go
type SystemConfig struct {
    Ip   string `mapstructure:"ip"`
    Port int    `mapstructure:"port"`
}
```

#### DatabaseConfig（数据库配置）

```go
type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    Name     string `mapstructure:"name"`
}
```

#### JWTConfig（JWT配置）

```go
type JWTConfig struct {
    Secret string `mapstructure:"secret"`
    Expire int    `mapstructure:"expire"`
}
```

#### EmailConfig（邮件配置）

```go
type EmailConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    From     string `mapstructure:"from"`
}
```

#### LogConfig（日志配置）

```go
type LogConfig struct {
    Filename   string `mapstructure:"filename"`
    MaxSize    int    `mapstructure:"maxSize"`
    MaxBackups int    `mapstructure:"maxBackups"`
    MaxAge     int    `mapstructure:"maxAge"`
    Compress   bool   `mapstructure:"compress"`
    Level      string `mapstructure:"level"`
}
```

#### Config（完整配置）

```go
type Config struct {
    Server   SystemConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Email    EmailConfig    `mapstructure:"email"`
    Log      LogConfig      `mapstructure:"log"`
}
```

### 3.4 全局变量

```go
var GlobalConfig *Config  // 全局配置实例，加载后通过它访问配置
```

### 3.5 核心函数

#### ReadConfig(env string) error 读取并合并配置

```go
// 传入环境名：dev / test / release
// 自动合并 app.yaml + app-{env}.yaml
if err := core.ReadConfig("dev"); err != nil {
    panic(err)
}

// 等同于
if err := core.ReadConfig(flags.Opt.Mode); err != nil {
    panic(err)
}
```

### 3.6 配置访问方式

```go
// 加载配置（传入环境名）
core.ReadConfig("dev")

// 服务配置
core.GlobalConfig.Server.Ip     // 服务 IP
core.GlobalConfig.Server.Port   // 服务端口

// 数据库配置
core.GlobalConfig.Database.Host     // 数据库地址
core.GlobalConfig.Database.Port     // 数据库端口
core.GlobalConfig.Database.Username // 数据库用户名
core.GlobalConfig.Database.Password // 数据库密码
core.GlobalConfig.Database.Name     // 数据库名

// JWT 配置
core.GlobalConfig.JWT.Secret  // JWT 密钥
core.GlobalConfig.JWT.Expire  // Token 过期时间（小时）

// 日志配置
core.GlobalConfig.Log.Filename   // 日志文件路径
core.GlobalConfig.Log.Level      // 日志级别
core.GlobalConfig.Log.MaxSize    // 单文件最大 MB
core.GlobalConfig.Log.MaxBackups // 保留旧文件数量
core.GlobalConfig.Log.MaxAge     // 保留天数
core.GlobalConfig.Log.Compress   // 是否压缩
```

### 3.7 配置文件示例

#### app.yaml（基础配置）

```yaml
server:
  ip: "0.0.0.0"
  port: 8080
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "123456"
  name: "myblog"
jwt:
  secret: "your-secret-key-change-in-production"
  expire: 24
email:
  host: "smtp.qq.com"
  port: 587
  username: ""
  password: ""
  from: ""
log:
  filename: "./logs/app.log"
  maxSize: 10
  maxBackups: 5
  maxAge: 30
  compress: true
```

#### app-dev.yaml（开发环境覆盖）

```yaml
server:
  port: 8080
database:
  name: "myblog_dev"
jwt:
  secret: "dev-secret-key"
log:
  level: "debug"
```

#### app-prod.yaml（生产环境覆盖）

```yaml
server:
  port: 80
  ip: "0.0.0.0"
database:
  name: "myblog_prod"
  username: "prod_user"
  password: "prod_password"
jwt:
  secret: "prod-secret-key-must-be-changed"
  expire: 168
log:
  level: "error"
```

### 3.8 配置合并原理

| 层级 | 文件 | 说明 |
|------|------|------|
| 基础层 | app.yaml | 所有环境共用的默认配置 |
| 覆盖层 | app-{env}.yaml | 特定环境的覆盖配置 |

**合并规则**：
- 基础层提供所有配置的默认值
- 覆盖层只写需要修改的配置项
- 覆盖层的值会覆盖基础层（深度合并）

**例子**：
- app.yaml 定义 `server.port: 8080`
- app-prod.yaml 定义 `server.port: 80`
- 最终生效的是 `server.port: 80`

// 邮件配置
core.GlobalConfig.Email.Host     // SMTP 地址
core.GlobalConfig.Email.Port     // SMTP 端口
core.GlobalConfig.Email.Username // 用户名
core.GlobalConfig.Email.Password // 密码
core.GlobalConfig.Email.From     // 发件人
```

### 3.6 配置文件结构

```
config/
├── dev.yaml    # 开发环境配置（本地开发用）
├── test.yaml   # 测试环境配置（测试服务器用）
└── prod.yaml  # 生产环境配置（线上用）
```

### 3.7 dev.yaml 配置示例

```yaml
server:
  ip: "127.0.0.1"
  port: 8080

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "123456"
  name: "myblog_dev"

jwt:
  secret: "dev-secret-key-change-in-production"
  expire: 24

email:
  host: "smtp.qq.com"
  port: 587
  username: "your_qq@qq.com"
  password: "your_auth_code"
  from: "your_qq@qq.com"
```

### 3.8 配置加载流程

```go
// 传入环境名，自动合并基础配置 + 环境配置
if err := core.ReadConfig(flags.Opt.Mode); err != nil {
    logger.S.Fatalf("配置加载失败: %v", err)
}

// 等同于：core.ReadConfig("dev") / core.ReadConfig("test") / core.ReadConfig("release")
// 自动加载：app.yaml + app-dev.yaml / app-test.yaml / app-prod.yaml
```

---

## 四、core/init_db 数据库模块

### 4.1 模块概述

| 项目 | 说明 |
|------|------|
| 依赖库 | `gorm.io/gorm` + `gorm.io/driver/mysql` |
| 功能 | MySQL 数据库连接和操作 |
| 全局实例 | `core.DB` (类型: `*gorm.DB`) |

### 4.2 核心类型

#### DBOptions 结构体（连接池配置）

```go
type DBOptions struct {
    MaxIdleConns  int           // 最大空闲连接数（数据库连接池数量）
    MaxOpenConns  int           // 最大打开连接数（同时最大连接数）
    ConnMaxLife   time.Duration // 连接最大生命周期（超过此时间连接会被关闭）
    SlowThreshold time.Duration // 慢查询阈值（超过此时间记录为慢查询）
}
```

#### DBOptions 字段详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| MaxIdleConns | int | `10` | 空闲时保持的连接数 |
| MaxOpenConns | int | `100` | 最多同时打开的连接数 |
| ConnMaxLife | time.Duration | `1小时` | 连接最大存活时间 |
| SlowThreshold | time.Duration | `200ms` | 超过此时间记录为慢查询 |

#### defaultDBOptions 默认连接池配置

```go
var defaultDBOptions = &DBOptions{
    MaxIdleConns:  10,
    MaxOpenConns:  100,
    ConnMaxLife:   time.Hour,
    SlowThreshold: 200 * time.Millisecond,
}
```

### 4.3 全局变量

```go
var DB *gorm.DB  // 全局数据库实例
```

### 4.4 核心函数

#### InitDB(opts ...*DBOptions) error 初始化数据库

```go
// 第1步：先加载配置（InitDB 依赖配置中的数据库信息）
core.ReadConfig("config/dev.yaml")

// 第2步：使用默认连接参数初始化
if err := core.InitDB(); err != nil {
    panic(err)
}

// 或使用自定义连接参数
if err := core.InitDB(&core.DBOptions{
    MaxIdleConns:  10,
    MaxOpenConns:  50,
    ConnMaxLife:   time.Hour,
    SlowThreshold: 500 * time.Millisecond,
}); err != nil {
    panic(err)
}
```

#### InitDBWithOptions(opt *DBOptions) error 带选项初始化

```go
// InitDB 的别名，支持传入自定义连接参数
core.InitDBWithOptions(&core.DBOptions{
    MaxIdleConns: 20,
    MaxOpenConns: 100,
})
```

#### CloseDB() error 关闭数据库连接

```go
// 程序退出时调用，释放数据库连接资源
if err := core.CloseDB(); err != nil {
    println("关闭数据库失败:", err)
}

// 更推荐的方式：defer
defer core.CloseDB()
```

#### PingDB() error 健康检查

```go
// 检查数据库连接是否正常
if err := core.PingDB(); err != nil {
    logger.S.Errorw("数据库连接失败", "error", err.Error())
} else {
    logger.S.Info("数据库连接正常")
}
```

#### AutoMigrate(models ...interface{}) error 自动迁移

```go
// 定义模型
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"size:64;unique"`
    Password  string    `gorm:"size:128"`
    Email     string    `gorm:"size:128"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 自动创建或更新表结构（单个模型）
if err := core.AutoMigrate(&User{}); err != nil {
    panic(err)
}

// 多个模型一起迁移（推荐方式）
core.AutoMigrate(&User{}, &Article{}, &Comment{})

// 或使用 models.AllModels()（需先 import models 包）
core.AutoMigrate(models.AllModels()...)
```

#### GetDB() *gorm.DB 获取数据库实例

```go
// 获取数据库实例
db := core.GetDB()

// 使用 GORM 进行数据库操作
```

### 4.5 路由层详解

#### 4.5.1 目录结构

```
api/router/
├── router.go              ← 主路由入口，汇总各模块
└── nestedrouter/
    ├── user.go            ← 用户路由
    ├── article.go         ← 文章路由
    └── comment.go         ← 评论路由
```

#### 4.5.2 主路由（router.go）

```go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/GoWeb/My_Blog/api/router/nestedrouter"
)

func InitRouter() *gin.Engine {
    r := gin.Default()

    apiGroup := r.Group("/api")
    {
        nestedrouter.UserRoutes(apiGroup)
        nestedrouter.ArticleRoutes(apiGroup)
        nestedrouter.CommentRoutes(apiGroup)
    }

    return r
}
```

#### 4.5.3 模块路由示例（user.go）

```go
package nestedrouter

import (
    "github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
    userGroup := r.Group("/user")
    {
        userGroup.POST("/register", nil)  // 注册
        userGroup.POST("/login", nil)     // 登录
        userGroup.GET("/info", nil)        // 用户信息
    }
}
```

#### 4.5.4 路由层级关系

```
r = gin.Default()
    ↓
apiGroup = r.Group("/api")
    ↓
userGroup = apiGroup.Group("/user")
    ↓
实际路径 = /api/user/register
```

#### 4.5.5 添加新模块路由

1. 在 `nestedrouter/` 下创建新文件 `xxx.go`
2. 定义 `XxxRoutes(r *gin.RouterGroup)` 函数
3. 在 `router.go` 中 import 并调用

```go
// nestedrouter/xxx.go
package nestedrouter

import "github.com/gin-gonic/gin"

func XxxRoutes(r *gin.RouterGroup) {
    group := r.Group("/xxx")
    {
        group.POST("/create", nil)
        group.GET("/list", nil)
    }
}

// router.go 中添加
import "github.com/GoWeb/My_Blog/api/router/nestedrouter"

func InitRouter() *gin.Engine {
    r := gin.Default()
    apiGroup := r.Group("/api")
    {
        nestedrouter.XxxRoutes(apiGroup)  // 新增
    }
    return r
}
```

#### 4.5.6 main.go 中使用路由

```go
import "github.com/GoWeb/My_Blog/api/router"

func main() {
    // ... 初始化代码 ...

    r := router.InitRouter()
    r.Run(":8080")
}
```

#### 4.5.7 HTTP 方法对照

| Gin 方法 | HTTP 方法 | 用途 |
|----------|-----------|------|
| `r.GET()` | GET | 查询 |
| `r.POST()` | POST | 创建 |
| `r.PUT()` | PUT | 更新 |
| `r.DELETE()` | DELETE | 删除 |
| `r.PATCH()` | PATCH | 部分更新 |

### 4.6 响应封装详解

#### 4.6.1 目录结构

```
core/response/
└── response.go     ← 统一响应封装
```

#### 4.6.2 响应结构

```go
type Response struct {
    Code    int         `json:"code"`             // 状态码：0成功，非0失败
    Msg     string      `json:"msg"`              // 消息
    Data    interface{} `json:"data,omitempty"`   // 数据（成功时有值）
    Error   string      `json:"error,omitempty"`  // 错误详情（失败时有值）
}

type PageResult struct {
    List       interface{} `json:"list"`        // 数据列表
    Total      int64       `json:"total"`       // 总条数
    Page       int         `json:"page"`         // 当前页
    PageSize   int         `json:"page_size"`    // 每页条数
    TotalPages int         `json:"total_pages"`  // 总页数
}
```

#### 4.6.3 状态码常量

```go
const (
    CodeSuccess           = 0      // 成功
    CodeParamError        = 400    // 参数错误
    CodeUnauthorized      = 401   // 未认证
    CodeForbidden         = 403   // 无权限
    CodeNotFound          = 404    // 资源不存在
    CodeInternalError     = 500   // 服务器内部错误
    CodeDatabaseError     = 50001 // 数据库错误
    CodeBusinessError     = 50002 // 业务逻辑错误
    CodeValidationError   = 40001 // 数据校验失败
    CodeDuplicateError    = 40002 // 数据重复冲突
)
```

#### 4.6.4 成功响应

```go
// 最常用：成功响应（默认 msg = "success"）
response.Success(c, data)

// 自定义消息
response.SuccessWithMsg(c, data, "操作成功")

// 分页响应（列表查询时使用）
response.SuccessPage(c, list, total, page, pageSize)
```

#### 4.6.5 错误响应

```go
// 通用错误（默认 code = 500）
response.Error(c, "操作失败")

// 自定义错误码
response.ErrorWithCode(c, 1001, "参数错误")

// 参数错误
response.ParamError(c)                      // 默认 "invalid parameter"
response.ParamErrorWithMsg(c, "用户名不能为空")

// 认证失败（HTTP 401）
response.Unauthorized(c)
response.UnauthorizedWithMsg(c, "token 已过期")

// 无权限（HTTP 403）
response.Forbidden(c)
response.ForbiddenWithMsg(c, "没有操作权限")

// 资源不存在（HTTP 404）
response.NotFound(c)
response.NotFoundWithMsg(c, "用户不存在")

// 服务器内部错误（HTTP 500）
response.InternalError(c)
response.InternalErrorWithMsg(c, "服务暂不可用")

// 业务错误（HTTP 200，code = 50002）
response.BusinessError(c, "余额不足")

// 数据库错误（HTTP 200，code = 50001）
response.DatabaseError(c)

// 数据校验失败（HTTP 200，code = 40001）
response.ValidationError(c, "邮箱格式不正确")

// 数据重复冲突（HTTP 200，code = 40002）
response.DuplicateError(c, "用户名已被注册")
```

#### 4.6.6 控制器中使用

```go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/GoWeb/My_Blog/core/logger"
    "github.com/GoWeb/My_Blog/core/response"
)

func Register(c *gin.Context) {
    // 1. 解析请求参数
    var req RegisterReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数解析失败")
        return
    }

    // 2. 业务逻辑
    logger.S.Info("User register")

    // 3. 返回响应
    response.SuccessWithMsg(c, nil, "注册成功")
}

func GetUserList(c *gin.Context) {
    page := 1
    pageSize := 10

    // 查询数据库获取 list 和 total
    var users []User
    var total int64

    // ... 查询逻辑 ...

    // 分页响应
    response.SuccessPage(c, users, total, page, pageSize)
}
```

#### 4.6.7 响应示例

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1,
        "username": "zhangsan"
    }
}
```

**分页响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "list": [...],
        "total": 100,
        "page": 1,
        "page_size": 10,
        "total_pages": 10
    }
}
```

**错误响应：**
```json
{
    "code": 40001,
    "msg": "用户名不能为空",
    "data": null
}
```

#### 4.6.8 RESTful API 响应规范

**成功响应格式：**
```json
{
    "code": 0,
    "msg": "success",
    "data": { ... }
}
```

**分页响应格式：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "list": [...],
        "total": 100,
        "page": 1,
        "page_size": 10,
        "total_pages": 10
    }
}
```

**错误响应格式：**
```json
{
    "code": 非0数字,
    "msg": "错误描述",
    "data": null
}
```

### 4.7 响应封装函数一览表

| 函数 | 签名 | 功能 |
|------|------|------|
| `Success` | `Success(c *gin.Context, data interface{})` | 成功响应（默认 msg="success"） |
| `SuccessWithMsg` | `SuccessWithMsg(c *gin.Context, data interface{}, msg string)` | 成功响应（自定义消息） |
| `SuccessPage` | `SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int)` | 分页响应 |
| `Error` | `Error(c *gin.Context, msg string)` | 通用错误（code=500） |
| `ErrorWithCode` | `ErrorWithCode(c *gin.Context, code int, msg string)` | 自定义错误码 |
| `ParamError` | `ParamError(c *gin.Context)` | 参数错误（code=400） |
| `ParamErrorWithMsg` | `ParamErrorWithMsg(c *gin.Context, msg string)` | 参数错误（自定义消息） |
| `Unauthorized` | `Unauthorized(c *gin.Context)` | 未认证（HTTP 401） |
| `UnauthorizedWithMsg` | `UnauthorizedWithMsg(c *gin.Context, msg string)` | 未认证（自定义消息） |
| `Forbidden` | `Forbidden(c *gin.Context)` | 无权限（HTTP 403） |
| `ForbiddenWithMsg` | `ForbiddenWithMsg(c *gin.Context, msg string)` | 无权限（自定义消息） |
| `NotFound` | `NotFound(c *gin.Context)` | 资源不存在（HTTP 404） |
| `NotFoundWithMsg` | `NotFoundWithMsg(c *gin.Context, msg string)` | 资源不存在（自定义消息） |
| `InternalError` | `InternalError(c *gin.Context)` | 服务器内部错误（HTTP 500） |
| `InternalErrorWithMsg` | `InternalErrorWithMsg(c *gin.Context, msg string)` | 服务器错误（自定义消息） |
| `DatabaseError` | `DatabaseError(c *gin.Context)` | 数据库错误（code=50001） |
| `BusinessError` | `BusinessError(c *gin.Context, msg string)` | 业务错误（code=50002） |
| `ValidationError` | `ValidationError(c *gin.Context, msg string)` | 数据校验失败（code=40001） |
| `DuplicateError` | `DuplicateError(c *gin.Context, msg string)` | 数据重复冲突（code=40002） |

### 4.8 数据库函数一览表

| 函数 | 签名 | 功能 |
|------|------|------|
| `InitDB` | `InitDB(opts ...*DBOptions) error` | 初始化数据库连接 |
| `InitDBWithOptions` | `InitDBWithOptions(opt *DBOptions) error` | 带选项初始化 |
| `CloseDB` | `CloseDB() error` | 关闭数据库连接 |
| `PingDB` | `PingDB() error` | 健康检查 |
| `AutoMigrate` | `AutoMigrate(models ...interface{}) error` | 自动迁移表结构 |
| `GetDB` | `GetDB() *gorm.DB` | 获取数据库实例 |

### 4.9 DAO 层使用详解

获取数据库实例后，通过 GORM 进行 CRUD 操作。

#### 创建（Create）

```go
db := core.GetDB()

// 创建单条记录
user := User{
    Username: "zhangsan",
    Password: "encrypted_password",
    Email: "zhangsan@example.com",
}
result := db.Create(&user)

// 检查错误
if result.Error != nil {
    logger.S.Errorw("创建用户失败", "error", result.Error)
    return
}

// 获取插入的 ID
logger.S.Infow("用户创建成功", "id", user.ID)

// 影响行数
logger.S.Infof("插入 %d 条记录", result.RowsAffected)
```

#### 查询（Read）

```go
db := core.GetDB()

// 根据主键查询
var user User
db.First(&user, 1) // SELECT * FROM users WHERE id = 1

// 条件查询
var user User
db.Where("username = ?", "zhangsan").First(&user)
// SELECT * FROM users WHERE username = 'zhangsan' LIMIT 1

// Where 多种用法
db.Where("age > ?", 18).Find(&users)           // 大于
db.Where("age >= ? AND status = ?", 18, "active").Find(&users) // AND 条件
db.Where("username IN ?", []string{"zhangsan", "lisi"}).Find(&users) // IN 查询

// First vs Find
// First: 获取第一条记录，找不到会报错
// Find: 获取所有记录，不会报错

// Or 条件
db.Where("age > ?", 18).Or("status = ?", "vip").Find(&users)

// Select 指定字段
db.Select("username, email").Find(&users)

// Order 排序
db.Order("created_at DESC").Find(&users)

// Limit 和 Offset 分页
db.Limit(10).Offset(0).Find(&users) // 第1页，每页10条

// Pluck 获取单列值
var usernames []string
db.Model(&User{}).Pluck("username", &usernames)
```

#### 更新（Update）

```go
db := core.GetDB()

// 更新单个字段
db.Model(&user).Update("email", "new@example.com")
// UPDATE users SET email = 'new@example.com' WHERE id = ?

// 更新多个字段
db.Model(&user).Updates(map[string]interface{}{
    "email": "new@example.com",
    "username": "new_name",
})
// UPDATE users SET email = 'new@example.com', username = 'new_name' WHERE id = ?

// 更新多个字段（struct，会忽略零值）
db.Model(&user).Updates(User{
    Email: "new@example.com",
    Username: "new_name",
})

// 批量更新
db.Model(&User{}).Where("status = ?", "inactive").Update("status", "archived")
// UPDATE users SET status = 'archived' WHERE status = 'inactive'

// 自增/自减
db.Model(&user).UpdateColumn("count", gorm.Expr("count + 1"))
```

#### 删除（Delete）

```go
db := core.GetDB()

// 根据主键删除
db.Delete(&user, 1) // DELETE FROM users WHERE id = 1

// 删除当前记录
db.Delete(&user) // DELETE FROM users WHERE id = ?

// 批量删除
db.Where("status = ?", "deleted").Delete(&User{})
// DELETE FROM users WHERE status = 'deleted'

// 软删除（需要模型有 DeletedAt 字段）
db.Delete(&user) // UPDATE users SET deleted_at = NOW() WHERE id = ?

// 永久删除（忽略软删除）
db.Unscoped().Delete(&user)
```

#### 控制器开发流程

**需求：实现用户注册接口**

1. 分析请求参数（注册需要：昵称、密码、邮箱）
2. 定义请求结构体（使用 binding 校验）
3. 查询数据库检查重复（昵称、邮箱）
4. 密码加密存储（bcrypt）
5. 创建用户记录
6. 返回统一响应

```go
// 1. 定义请求结构体
type RegisterReq struct {
    Nickname string `json:"nickname" binding:"required,min=2,max=32"`
    Password string `json:"password" binding:"required,min=6,max=20"`
    Email    string `json:"email" binding:"required,email"`
}

// 2. 控制器实现
func Register(c *gin.Context) {
    var req RegisterReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    db := core.GetDB()

    // 3. 检查昵称是否重复
    var count int64
    db.Model(&model.User{}).Where("nickname = ?", req.Nickname).Count(&count)
    if count > 0 {
        response.DuplicateError(c, "昵称已被使用")
        return
    }

    // 4. 检查邮箱是否重复
    db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
    if count > 0 {
        response.DuplicateError(c, "邮箱已被注册")
        return
    }

    // 5. 密码加密（bcrypt）
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        response.InternalError(c)
        return
    }

    // 6. 创建用户
    newUser := model.User{
        Nickname: req.Nickname,
        Password: string(hashedPassword),
        Email:    req.Email,
        Status:   1,
    }
    if err := db.Create(&newUser).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    // 7. 返回成功响应
    response.SuccessWithMsg(c, nil, "注册成功")
}
```

**关键点说明：**

| 步骤 | 函数 | 说明 |
|------|------|------|
| 参数绑定 | `c.ShouldBindJSON(&req)` | 自动解析 JSON 并校验 |
| 统计数量 | `db.Model().Where().Count(&count)` | 检查记录是否存在 |
| 密码加密 | `bcrypt.GenerateFromPassword()` | bcrypt 是 Go 推荐的密码加密库 |
| 创建记录 | `db.Create(&user)` | 插入数据到数据库 |
| 统一响应 | `response.Success/DuplicateError` | RESTful 风格响应 |

**bcrypt 密码加密原理：**

```go
// 加密：明文 → 密文（不可逆）
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
// 结果：$2a$10$N9qo8uLOickgx2ZMRZoMye...（随机盐值+hash）

// 校验：明文 与 密文 比较
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte("123456"))
// 相同返回 nil，不同返回 error
```

### 4.10 GORM 标签详解

定义模型时使用 GORM 标签来控制表结构。

#### 4.7.1 基础字段标签

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`                    // 主键
    UUID      string         `gorm:"type:varchar(36);uniqueIndex"`  // UUID，唯一索引
    Username  string         `gorm:"size:64;unique;not null"`        // 大小64，唯一，非空
    Password  string         `gorm:"size:128"`                       // 大小128
    Email     string         `gorm:"size:128;index"`                 // 大小128，普通索引
    Age       int            `gorm:"index"`                          // 普通索引
    Status    string         `gorm:"size:16;default:'active'"`      // 默认值
    Balance   float64        `gorm:"type:decimal(10,2)"`            // 小数类型
    Birthday  *time.Time     `gorm:"type:date"`                     // 日期类型
    DeletedAt gorm.DeletedAt `gorm:"index"`                         // 软删除
    CreatedAt time.Time                                        // 创建时间（自动）
    UpdatedAt time.Time                                        // 更新时间（自动）
}
```

| 标签 | 说明 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `type` | 指定数据库类型 | `gorm:"type:varchar(36)"` |
| `unique` | 唯一索引 | `gorm:"unique"` |
| `uniqueIndex` | 唯一索引（可指定名称） | `gorm:"uniqueIndex"` |
| `index` | 普通索引 | `gorm:"index"` |
| `size` | 字段长度 | `gorm:"size:64"` |
| `default` | 默认值 | `gorm:"default:'active'"` |
| `not null` | 非空约束 | `gorm:"not null"` |
| `type:decimal(10,2)` | 小数类型 | `gorm:"type:decimal(10,2)"` |
| `type:date` | 日期类型 | `gorm:"type:date"` |

#### 4.7.2 关联关系标签

```go
type Comment struct {
    ID        uint      `gorm:"primaryKey"`
    Content   string    `gorm:"type:text;not null"`
    ArticleID uint      `gorm:"not null"`
    UserID    uint      `gorm:"not null"`

    // foreignKey: 指定当前表的外键字段名
    // references: 指定对方表的主键字段名（默认是 ID，可省略）
    User      User      `gorm:"foreignKey:UserID;references:ID"`
    Article   Article   `gorm:"foreignKey:ArticleID;references:ID"`
}

type Article struct {
    ID      uint   `gorm:"primaryKey"`
    Title   string `gorm:"type:varchar(128);not null"`
    Content string `gorm:"type:text;not null"`
    UserID  uint   `gorm:"not null"`

    // 关联用户
    User    User    `gorm:"foreignKey:UserID;references:ID"`
}
```

| 标签 | 说明 | 示例 |
|------|------|------|
| `foreignKey` | 当前表的外键字段名 | `gorm:"foreignKey:UserID"` |
| `references` | 对方表的主键字段名 | `gorm:"references:ID"` |
| `json:"-"` | JSON序列化时忽略此字段 | `json:"-"` |

#### 4.7.3 关联标签详解

**foreignKey 和 references 的区别：**

```
foreignKey: 本表的外键字段叫什么
references: 对方表的主键字段叫什么

例：Comment 表的 ArticleID 关联 Article 表的 ID
    ArticleID  ← foreignKey：这是本表(Comment)的外键字段名
       ↓
    Article 表的 ID  ← references：这是对方表的主键字段名

    gorm:"foreignKey:ArticleID;references:ID"
    翻译：外键是 ArticleID，引用对方表的 ID 主键
```

**什么时候省略 references？**

- 当对方表主键是 `ID` 时可以省略（GORM 默认就是 ID）
- 当想明确表达关联关系时建议保留

**为什么 Comment 的 Article 关联要指定 references？**

因为 GORM 默认引用对方表的 `ID` 主键，而 Article 的主键确实是 `ID`。所以实际上：
```go
// 完整写法
gorm:"foreignKey:ArticleID;references:ID"

// 省略 references（效果相同，因为默认就是 ID）
gorm:"foreignKey:ArticleID"
```

### 4.11 连接池详解

#### 什么是连接池？

数据库连接池是预建立并保持的数据库连接集合，避免每次请求都新建连接（新建连接耗时约 10-100ms）。

```
┌─────────────────────────────────────────┐
│                 应用                     │
│  请求1 ──┐                              │
│  请求2 ──┼──► 连接池 ───► 数据库        │
│  请求3 ──┘    │                         │
│              保持 N 个空闲连接复用       │
└─────────────────────────────────────────┘
```

#### 为什么需要连接池？

| 方式 | 耗时 | 并发支持 |
|------|------|----------|
| 每次请求新建连接 | 10-100ms | 差 |
| 使用连接池 | 0.01ms | 好 |

#### 生产环境配置建议

```go
core.InitDB(&core.DBOptions{
    MaxIdleConns:  10,              // 保持10个空闲连接
    MaxOpenConns:  50,             // 最多50个并发连接（根据服务器配置调整）
    ConnMaxLife:   time.Hour,      // 1小时后重建连接
    SlowThreshold: 500 * time.Millisecond, // 超过500ms视为慢查询
})
```

### 4.12 初始化完整示例

```go
func main() {
    // 1. 解析命令行参数
    flags.Parse()

    // 2. 初始化日志
    if flags.IsRelease() {
        logger.Init(logger.Config{Level: zapcore.ErrorLevel})
    } else {
        logger.Init(logger.Config{Level: zapcore.DebugLevel})
    }
    defer logger.Sync()

    // 3. 选择配置文件
    configFile := "config/dev.yaml"
    if flags.IsRelease() {
        configFile = "config/prod.yaml"
    }

    // 4. 加载配置
    if err := core.ReadConfig(configFile); err != nil {
        logger.S.Fatalf("配置加载失败: %v", err)
    }

    // 5. 初始化数据库
    if err := core.InitDB(); err != nil {
        logger.S.Fatalf("数据库连接失败: %v", err)
    }
    defer core.CloseDB()

    // 6. 自动迁移表
    if err := core.AutoMigrate(&User{}); err != nil {
        logger.S.Fatalf("表迁移失败: %v", err)
    }

    logger.S.Info("初始化完成")
}
```

---

## 五、api/middleware JWT认证中间件

### 5.1 模块概述

| 项目 | 说明 |
|------|------|
| 依赖库 | `github.com/golang-jwt/jwt/v5` |
| 功能 | 验证 JWT Token，保护需要认证的接口 |
| 所在文件 | `api/middleware/jwt_auth.go` |

### 5.2 工作原理

中间件是夹在请求和控制器之间的"拦截器"：

```
请求 → JWTAuth中间件 → 控制器 → 响应
        ↓
    验证失败？直接返回 401
```

### 5.3 JWT Token 格式

前端请求时需要在 Header 中携带 Token：

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 5.4 Claims 结构体

```go
type Claims struct {
    UserID   uint   `json:"user_id"`   // 用户ID
    Nickname string `json:"nickname"`  // 用户昵称
    jwt.RegisteredClaims             // JWT 内置声明（包含过期时间等）
}
```

### 5.5 JWTAuth() 中间件函数

```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 Header 获取 Authorization
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, response.Response{
                Code: response.CodeUnauthorized,
                Msg:  "未提供认证令牌",
            })
            c.Abort()
            return
        }

        // 2. 验证格式：Bearer xxx
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, response.Response{
                Code: response.CodeUnauthorized,
                Msg:  "认证令牌格式错误",
            })
            c.Abort()
            return
        }

        tokenString := parts[1]

        // 3. 解析 Token
        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte(core.GlobalConfig.JWT.Secret), nil
        })

        // 4. 验证 Token 有效性
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, response.Response{
                Code: response.CodeUnauthorized,
                Msg:  "无效的认证令牌",
            })
            c.Abort()
            return
        }

        // 5. 将用户信息存入 Context，供后续控制器使用
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Nickname)

        c.Next()
    }
}
```

### 5.6 流程图

```
请求进入中间件
    │
    ▼
检查 Authorization Header 是否存在
    │
    ├── 不存在 → 返回 401 "未提供认证令牌"
    │
    ▼
验证格式是否为 "Bearer xxx"
    │
    ├── 格式错误 → 返回 401 "认证令牌格式错误"
    │
    ▼
使用 JWT Secret 解析 Token
    │
    ├── 解析失败/Token无效 → 返回 401 "无效的认证令牌"
    │
    ▼
从 Token 中提取用户信息
    │
    ▼
存入 Context：c.Set("user_id", claims.UserID)
              c.Set("username", claims.Nickname)
    │
    ▼
c.Next() → 执行后续控制器
```

### 5.7 Token 生成方法（在登录时使用）

```go
import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// 登录成功后生成 Token
func generateToken(user model.User) (string, error) {
    // 设置过期时间
    expireTime := time.Now().Add(time.Duration(core.GlobalConfig.JWT.Expire) * time.Hour)

    // 创建 Claims（负载）
    claims := jwt.MapClaims{
        "user_id":  user.ID,       // 用户ID
        "nickname": user.Nickname, // 用户昵称
        "exp":      expireTime.Unix(),  // 过期时间
        "iat":      time.Now().Unix(),  // 签发时间
    }

    // 创建 Token（签名算法：HS256）
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // 用 Secret 签名，生成 tokenString
    tokenString, err := token.SignedString([]byte(core.GlobalConfig.JWT.Secret))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}
```

### 5.8 控制器中获取当前用户

```go
func GetUserInfo(c *gin.Context) {
    // 从 Context 中获取用户ID（中间件已存入）
    userID, exists := c.Get("user_id")
    if !exists {
        response.UnauthorizedWithMsg(c, "用户未登录")
        return
    }

    // 转换类型（interface{} 转 uint）
    uid := userID.(uint)

    // 查询用户信息
    db := core.GetDB()
    var user model.User
    if err := db.First(&user, uid).Error; err != nil {
        response.NotFoundWithMsg(c, "用户不存在")
        return
    }

    response.Success(c, gin.H{
        "id":       user.ID,
        "nickname": user.Nickname,
        "email":    user.Email,
        "avatar":   user.Avatar,
    })
}
```

### 5.9 在路由中使用中间件

```go
// router.go
func InitRouter() *gin.Engine {
    r := gin.Default()

    // 公开接口（不需要认证）
    apiGroup := r.Group("/api")
    {
        nestedrouter.UserRoutes(apiGroup)  // 注册、登录
    }

    // 受保护接口（需要认证）
    protectedGroup := r.Group("/api")
    protectedGroup.Use(middleware.JWTAuth())  // 添加 JWT 中间件
    {
        nestedrouter.ArticleRoutes(protectedGroup)  // 文章增删改查
        nestedrouter.CommentRoutes(protectedGroup)   // 评论增删改查
    }

    return r
}
```

### 5.10 路由分组说明

| 分组 | 中间件 | 说明 |
|------|--------|------|
| `apiGroup` | 无 | 公开接口，任何人都能访问 |
| `protectedGroup` | `JWTAuth()` | 受保护接口，需要携带有效 Token |

### 5.11 前端调用示例

```javascript
// 请求头中携带 Token
const token = localStorage.getItem('token');

// 创建文章（需要认证）
fetch('/api/article/create', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token  // 注意格式：Bearer + 空格 + token
    },
    body: JSON.stringify({
        title: '我的文章',
        content: '文章内容...'
    })
})
```

---

## 六、api/controller 控制器详解

### 6.1 模块概述

| 项目 | 说明 |
|------|------|
| 控制器文件 | `api/controller/user_controller.go` |
| 控制器文件 | `api/controller/article_controller.go` |
| 控制器文件 | `api/controller/comment_controller.go` |

### 6.2 User 控制器（用户模块）

#### 6.2.1 请求结构体

```go
// 注册请求
type RegisterReq struct {
    Nickname string `json:"nickname" binding:"required,min=2,max=32"`
    Password string `json:"password" binding:"required,min=6,max=20"`
    Email    string `json:"email" binding:"required,email"`
}

// 登录请求
type LoginReq struct {
    Login    string `json:"login" binding:"required"`     // 可输入昵称或邮箱
    Password string `json:"password" binding:"required"`
}
```

#### 6.2.2 Register 注册

```go
func Register(c *gin.Context) {
    // 1. 解析请求参数
    var req RegisterReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    db := core.GetDB()

    // 2. 检查昵称是否重复
    var count int64
    db.Model(&model.User{}).Where("nickname = ?", req.Nickname).Count(&count)
    if count > 0 {
        response.DuplicateError(c, "昵称已被使用")
        return
    }

    // 3. 检查邮箱是否重复
    db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
    if count > 0 {
        response.DuplicateError(c, "邮箱已被注册")
        return
    }

    // 4. 密码加密
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        response.InternalError(c)
        return
    }

    // 5. 创建用户
    newUser := model.User{
        Nickname: req.Nickname,
        Password: string(hashedPassword),
        Email:    req.Email,
        Status:   1,
    }

    if err := db.Create(&newUser).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    // 6. 返回成功响应
    response.SuccessWithMsg(c, gin.H{"id": newUser.ID}, "注册成功")
}
```

#### 6.2.3 Login 登录

```go
func Login(c *gin.Context) {
    // 1. 解析请求参数
    var req LoginReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    db := core.GetDB()

    // 2. 查询用户（支持昵称或邮箱登录）
    var user model.User
    err := db.Where("nickname = ? OR email = ?", req.Login, req.Login).First(&user).Error
    if err != nil {
        response.UnauthorizedWithMsg(c, "用户不存在")
        return
    }

    // 3. 检查账号状态
    if user.Status == 0 {
        response.ForbiddenWithMsg(c, "账号已被禁用")
        return
    }

    // 4. 验证密码
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        response.UnauthorizedWithMsg(c, "密码错误")
        return
    }

    // 5. 生成 JWT Token
    expireTime := time.Now().Add(time.Duration(core.GlobalConfig.JWT.Expire) * time.Hour)
    claims := jwt.MapClaims{
        "user_id":  user.ID,
        "nickname": user.Nickname,
        "exp":      expireTime.Unix(),
        "iat":      time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(core.GlobalConfig.JWT.Secret))
    if err != nil {
        response.InternalError(c)
        return
    }

    // 6. 返回 Token 和用户信息
    response.Success(c, gin.H{
        "token":    tokenString,
        "user_id":  user.ID,
        "nickname": user.Nickname,
        "email":    user.Email,
        "avatar":   user.Avatar,
    })
}
```

#### 6.2.4 GetUserInfo 获取用户信息

```go
func GetUserInfo(c *gin.Context) {
    // 从 Context 获取当前用户ID（JWT 中间件已存入）
    userID, exists := c.Get("user_id")
    if !exists {
        response.UnauthorizedWithMsg(c, "用户未登录")
        return
    }

    db := core.GetDB()
    var user model.User
    if err := db.First(&user, userID).Error; err != nil {
        response.NotFoundWithMsg(c, "用户不存在")
        return
    }

    response.Success(c, gin.H{
        "id":         user.ID,
        "nickname":   user.Nickname,
        "email":      user.Email,
        "avatar":     user.Avatar,
        "abstract":   user.Abstract,
        "status":     user.Status,
        "created_at": user.CreatedAt.Format(time.DateTime),
    })
}
```

### 6.3 Article 控制器（文章模块）

#### 6.3.1 请求结构体

```go
// 创建文章
type ArticleReq struct {
    Title   string `json:"title" binding:"required,min=1,max=128"`
    Content string `json:"content" binding:"required,min=1"`
}

// 文章列表（分页）
type ArticleListReq struct {
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
    Keyword  string `form:"keyword"`
}

// 更新文章
type ArticleUpdateReq struct {
    Title   string `json:"title" binding:"required,min=1,max=128"`
    Content string `json:"content" binding:"required,min=1"`
}
```

#### 6.3.2 CreateArticle 创建文章

```go
func CreateArticle(c *gin.Context) {
    // 1. 解析请求参数
    var req ArticleReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    db := core.GetDB()

    // 2. 获取当前用户ID（JWT 中间件已存入）
    var userID uint
    if v, exists := c.Get("user_id"); exists {
        userID = v.(uint)
    } else {
        response.UnauthorizedWithMsg(c, "用户未登录")
        return
    }

    // 3. 创建文章
    newArticle := model.Article{
        Title:   req.Title,
        Content: req.Content,
        UserID:  userID,
    }

    if err := db.Create(&newArticle).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    response.SuccessWithMsg(c, gin.H{"id": newArticle.ID}, "文章创建成功")
}
```

#### 6.3.3 GetArticleList 获取文章列表

```go
func GetArticleList(c *gin.Context) {
    // 1. 解析分页参数
    var req ArticleListReq
    if err := c.ShouldBindQuery(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    // 2. 设置默认值
    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 || req.PageSize > 50 {
        req.PageSize = 10
    }

    db := core.GetDB()
    offset := (req.Page - 1) * req.PageSize

    var total int64
    var articles []model.Article

    query := db.Model(&model.Article{})

    // 3. 关键字搜索
    if req.Keyword != "" {
        keyword := "%" + req.Keyword + "%"
        query = query.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
    }

    // 4. 统计总数
    query.Count(&total)

    // 5. 查询列表（预加载用户信息）
    if err := query.Preload("User").
        Order("created_at DESC").
        Offset(offset).
        Limit(req.PageSize).
        Find(&articles).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    // 6. 返回分页结果
    response.SuccessPage(c, articles, total, req.Page, req.PageSize)
}
```

#### 6.3.4 GetArticleDetail 获取文章详情

```go
func GetArticleDetail(c *gin.Context) {
    // 1. 获取文章ID
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        response.ParamErrorWithMsg(c, "无效的文章ID")
        return
    }

    db := core.GetDB()
    var article model.Article

    // 2. 查询文章（预加载用户信息）
    if err := db.Preload("User").First(&article, id).Error; err != nil {
        response.NotFoundWithMsg(c, "文章不存在")
        return
    }

    response.Success(c, article)
}
```

#### 6.3.5 UpdateArticle 更新文章

```go
func UpdateArticle(c *gin.Context) {
    // 1. 获取文章ID
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        response.ParamErrorWithMsg(c, "无效的文章ID")
        return
    }

    // 2. 解析请求参数
    var req ArticleUpdateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    db := core.GetDB()
    var article model.Article

    // 3. 查询原文章
    if err := db.First(&article, id).Error; err != nil {
        response.NotFoundWithMsg(c, "文章不存在")
        return
    }

    // 4. 权限检查（只能修改自己的文章）
    userID := c.GetUint("user_id")
    if article.UserID != userID {
        response.ForbiddenWithMsg(c, "无权限修改他人的文章")
        return
    }

    // 5. 更新文章
    updates := map[string]interface{}{
        "title":      req.Title,
        "content":    req.Content,
        "updated_at": time.Now(),
    }

    if err := db.Model(&article).Updates(updates).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    response.SuccessWithMsg(c, nil, "文章更新成功")
}
```

#### 6.3.6 DeleteArticle 删除文章

```go
func DeleteArticle(c *gin.Context) {
    // 1. 获取文章ID
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        response.ParamErrorWithMsg(c, "无效的文章ID")
        return
    }

    db := core.GetDB()
    var article model.Article

    // 2. 查询文章
    if err := db.First(&article, id).Error; err != nil {
        response.NotFoundWithMsg(c, "文章不存在")
        return
    }

    // 3. 权限检查（只能删除自己的文章）
    userID := c.GetUint("user_id")
    if article.UserID != userID {
        response.ForbiddenWithMsg(c, "无权限删除他人的文章")
        return
    }

    // 4. 删除文章
    if err := db.Delete(&article).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    response.SuccessWithMsg(c, nil, "文章删除成功")
}
```

### 6.4 Comment 控制器（评论模块）

#### 6.4.1 请求结构体

```go
// 创建评论
type CommentReq struct {
    ArticleID uint   `json:"article_id" binding:"required"`
    Content  string `json:"content" binding:"required,min=1,max=500"`
}

// 评论列表
type CommentListReq struct {
    ArticleID uint `form:"article_id" binding:"required"`
    Page     int  `form:"page"`
    PageSize int  `form:"page_size"`
}
```

#### 6.4.2 CreateComment 创建评论

```go
func CreateComment(c *gin.Context) {
    // 1. 解析请求参数
    var req CommentReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    db := core.GetDB()

    // 2. 获取当前用户ID
    var userID uint
    if v, exists := c.Get("user_id"); exists {
        userID = v.(uint)
    } else {
        response.UnauthorizedWithMsg(c, "用户未登录")
        return
    }

    // 3. 检查文章是否存在
    var article model.Article
    if err := db.First(&article, req.ArticleID).Error; err != nil {
        response.NotFoundWithMsg(c, "文章不存在")
        return
    }

    // 4. 创建评论
    newComment := model.Comment{
        Content:   req.Content,
        ArticleID: req.ArticleID,
        UserID:    userID,
    }

    if err := db.Create(&newComment).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    response.SuccessWithMsg(c, gin.H{"id": newComment.ID}, "评论创建成功")
}
```

#### 6.4.3 GetCommentList 获取评论列表

```go
func GetCommentList(c *gin.Context) {
    // 1. 解析分页参数
    var req CommentListReq
    if err := c.ShouldBindQuery(&req); err != nil {
        response.ParamErrorWithMsg(c, "参数错误："+err.Error())
        return
    }

    // 2. 设置默认值
    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 || req.PageSize > 50 {
        req.PageSize = 10
    }

    db := core.GetDB()
    offset := (req.Page - 1) * req.PageSize

    var total int64
    var comments []model.Comment

    query := db.Model(&model.Comment{}).Where("article_id = ?", req.ArticleID)

    // 3. 统计总数
    query.Count(&total)

    // 4. 查询列表（预加载用户信息）
    if err := query.Preload("User").
        Order("created_at DESC").
        Offset(offset).
        Limit(req.PageSize).
        Find(&comments).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    // 5. 返回分页结果
    response.SuccessPage(c, comments, total, req.Page, req.PageSize)
}
```

#### 6.4.4 DeleteComment 删除评论

```go
func DeleteComment(c *gin.Context) {
    // 1. 获取评论ID
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        response.ParamErrorWithMsg(c, "无效的评论ID")
        return
    }

    db := core.GetDB()
    var comment model.Comment

    // 2. 查询评论
    if err := db.First(&comment, id).Error; err != nil {
        response.NotFoundWithMsg(c, "评论不存在")
        return
    }

    // 3. 权限检查（只能删除自己的评论）
    userID := c.GetUint("user_id")
    if comment.UserID != userID {
        response.ForbiddenWithMsg(c, "无权限删除他人的评论")
        return
    }

    // 4. 删除评论
    if err := db.Delete(&comment).Error; err != nil {
        response.DatabaseError(c)
        return
    }

    response.SuccessWithMsg(c, nil, "评论删除成功")
}
```

---

## 七、models/model 数据模型

### 7.1 User 模型

```go
package model

import "gorm.io/gorm"

type User struct {
    gorm.Model              // 内嵌 ID, CreatedAt, UpdatedAt, DeletedAt
    Nickname string `json:"nickname" gorm:"type:varchar(32);uniqueIndex;not null"`
    Avatar   string `json:"avatar"`                                        // 头像
    Abstract string `json:"abstract"`                                      // 简介
    Email    string `json:"email" gorm:"type:varchar(128);uniqueIndex;not null"`
    Password string `json:"-" gorm:"type:varchar(64);not null"`           // JSON输出时隐藏
    Status   int    `json:"status" gorm:"type:tinyint(1);default:1"`    // 0:禁用 1:正常
}
```

| 字段 | 类型 | GORM标签 | 说明 |
|------|------|----------|------|
| ID | uint | `gorm.Model` | 主键，自动递增 |
| Nickname | string | `type:varchar(32);uniqueIndex;not null` | 昵称，唯一，非空 |
| Avatar | string | 无 | 头像URL |
| Abstract | string | 无 | 个人简介 |
| Email | string | `type:varchar(128);uniqueIndex;not null` | 邮箱，唯一，非空 |
| Password | string | `json:"-"` | 密码，JSON序列化时隐藏 |
| Status | int | `type:tinyint(1);default:1` | 状态：0禁用 1正常 |
| CreatedAt | time.Time | `gorm.Model` | 创建时间 |
| UpdatedAt | time.Time | `gorm.Model` | 更新时间 |
| DeletedAt | time.Time | `gorm.Model` | 软删除时间 |

### 7.2 Article 模型

```go
package model

import "gorm.io/gorm"

type Article struct {
    gorm.Model              // 内嵌 ID, CreatedAt, UpdatedAt, DeletedAt
    Title   string `json:"title" gorm:"type:varchar(128);not null"`
    Content string `json:"content" gorm:"type:text;not null"`
    UserID  uint   `json:"user_id" gorm:"not null"`
    User    User   `json:"-" gorm:"foreignKey:UserID;references:ID"`  // 关联用户
}
```

| 字段 | 类型 | GORM标签 | 说明 |
|------|------|----------|------|
| ID | uint | `gorm.Model` | 主键 |
| Title | string | `type:varchar(128);not null` | 标题 |
| Content | string | `type:text;not null` | 内容 |
| UserID | uint | `not null` | 作者ID（外键） |
| User | User | `foreignKey:UserID;references:ID` | 关联的用户 |
| CreatedAt | time.Time | `gorm.Model` | 创建时间 |
| UpdatedAt | time.Time | `gorm.Model` | 更新时间 |
| DeletedAt | time.Time | `gorm.Model` | 软删除时间 |

### 7.3 Comment 模型

```go
package model

import "gorm.io/gorm"

type Comment struct {
    gorm.Model              // 内嵌 ID, CreatedAt, UpdatedAt, DeletedAt
    Content   string  `json:"content" gorm:"type:text;not null"`
    ArticleID uint    `json:"article_id" gorm:"not null"`
    UserID    uint    `json:"user_id" gorm:"not null"`
    User      User    `json:"-" gorm:"foreignKey:UserID;references:ID"`
    Article   Article `json:"-" gorm:"foreignKey:ArticleID;references:ID"`
}
```

| 字段 | 类型 | GORM标签 | 说明 |
|------|------|----------|------|
| ID | uint | `gorm.Model` | 主键 |
| Content | string | `type:text;not null` | 评论内容 |
| ArticleID | uint | `not null` | 关联文章ID |
| UserID | uint | `not null` | 评论用户ID |
| User | User | `foreignKey:UserID` | 关联的用户 |
| Article | Article | `foreignKey:ArticleID` | 关联的文章 |
| CreatedAt | time.Time | `gorm.Model` | 创建时间 |
| UpdatedAt | time.Time | `gorm.Model` | 更新时间 |
| DeletedAt | time.Time | `gorm.Model` | 软删除时间 |

### 7.4 模型注册

```go
// models/register.go
package models

import "github.com/GoWeb/My_Blog/models/model"

func AllModels() []interface{} {
    return []interface{}{
        &model.User{},      // 用户表
        &model.Article{},   // 文章表
        &model.Comment{},   // 评论表
    }
}
```

使用方式：
```go
// 自动创建所有表
core.AutoMigrate(models.AllModels()...)
```

---

## 八、api/router 路由配置

### 8.1 主路由（router.go）

```go
package router

import (
    "github.com/GoWeb/My_Blog/api/middleware"
    "github.com/GoWeb/My_Blog/api/router/nestedrouter"
    "github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
    r := gin.Default()

    // 公开路由（无需认证）
    apiGroup := r.Group("/api")
    {
        nestedrouter.UserRoutes(apiGroup)
    }

    // 受保护路由（需要 JWT 认证）
    protectedGroup := r.Group("/api")
    protectedGroup.Use(middleware.JWTAuth())
    {
        nestedrouter.ArticleRoutes(protectedGroup)
        nestedrouter.CommentRoutes(protectedGroup)
    }

    return r
}
```

### 8.2 User 路由（nestedrouter/user.go）

```go
package nestedrouter

import (
    "github.com/GoWeb/My_Blog/api/controller"
    "github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
    userGroup := r.Group("/user")
    {
        userGroup.POST("/register", controller.Register)   // 注册（公开）
        userGroup.POST("/login", controller.Login)        // 登录（公开）
        userGroup.GET("/info", controller.GetUserInfo)    // 用户信息（需认证）
    }
}
```

### 8.3 Article 路由（nestedrouter/article.go）

```go
package nestedrouter

import (
    "github.com/GoWeb/My_Blog/api/controller"
    "github.com/gin-gonic/gin"
)

func ArticleRoutes(r *gin.RouterGroup) {
    articleGroup := r.Group("/article")
    {
        articleGroup.POST("/create", controller.CreateArticle)     // 创建（需认证）
        articleGroup.GET("/list", controller.GetArticleList)       // 列表（需认证）
        articleGroup.GET("/detail/:id", controller.GetArticleDetail) // 详情（需认证）
        articleGroup.PUT("/update/:id", controller.UpdateArticle)  // 更新（需认证）
        articleGroup.DELETE("/delete/:id", controller.DeleteArticle) // 删除（需认证）
    }
}
```

### 8.4 Comment 路由（nestedrouter/comment.go）

```go
package nestedrouter

import (
    "github.com/GoWeb/My_Blog/api/controller"
    "github.com/gin-gonic/gin"
)

func CommentRoutes(r *gin.RouterGroup) {
    commentGroup := r.Group("/comment")
    {
        commentGroup.POST("/create", controller.CreateComment)   // 创建（需认证）
        commentGroup.GET("/list", controller.GetCommentList)     // 列表（需认证）
        commentGroup.DELETE("/delete/:id", controller.DeleteComment) // 删除（需认证）
    }
}
```

---

## 九、MySQL SQL 详解

### 5.1 SQL 分类概览

| 分类 | 全称 | 关键字 | 作用 |
|------|------|--------|------|
| DDL | Data Definition Language | CREATE, DROP, ALTER | 定义数据结构 |
| DML | Data Manipulation Language | INSERT, UPDATE, DELETE | 操作数据 |
| DQL | Data Query Language | SELECT | 查询数据 |
| DCL | Data Control Language | GRANT, REVOKE | 权限控制 |
| TCL | Transaction Control Language | COMMIT, ROLLBACK | 事务控制 |

### 5.2 DDL 数据定义语言

#### 5.2.1 数据库操作

```sql
-- 创建数据库
CREATE DATABASE blog DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE blog;

-- 删除数据库（危险！生产环境禁止）
DROP DATABASE blog;

-- 查看所有数据库
SHOW DATABASES;

-- 查看当前数据库
SELECT DATABASE();
```

#### 5.2.2 表操作

```sql
-- 创建表
CREATE TABLE users (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username    VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名',
    password    VARCHAR(128) NOT NULL COMMENT '密码（加密后）',
    email       VARCHAR(128) NOT NULL UNIQUE COMMENT '邮箱',
    nickname    VARCHAR(32) DEFAULT '' COMMENT '昵称',
    avatar      VARCHAR(256) DEFAULT '' COMMENT '头像URL',
    status      TINYINT(1) DEFAULT 1 COMMENT '状态：0禁用 1正常',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at  DATETIME DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 创建文章表
CREATE TABLE articles (
    id         INT PRIMARY KEY AUTO_INCREMENT COMMENT '文章ID',
    title      VARCHAR(128) NOT NULL COMMENT '标题',
    content    TEXT NOT NULL COMMENT '内容',
    user_id    INT NOT NULL COMMENT '作者ID',
    status     TINYINT(1) DEFAULT 1 COMMENT '状态：0草稿 1发布',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';

-- 创建评论表
CREATE TABLE comments (
    id         INT PRIMARY KEY AUTO_INCREMENT COMMENT '评论ID',
    content    TEXT NOT NULL COMMENT '评论内容',
    article_id INT NOT NULL COMMENT '文章ID',
    user_id    INT NOT NULL COMMENT '用户ID',
    parent_id  INT DEFAULT NULL COMMENT '父评论ID（回复）',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_article_id (article_id),
    INDEX idx_user_id (user_id),
    INDEX idx_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评论表';
```

#### 5.2.3 修改表结构

```sql
-- 添加字段
ALTER TABLE users ADD COLUMN phone VARCHAR(20) DEFAULT '' COMMENT '手机号' AFTER email;

-- 修改字段
ALTER TABLE users MODIFY COLUMN phone VARCHAR(30) DEFAULT NULL;

-- 重命名字段
ALTER TABLE users CHANGE COLUMN phone mobile VARCHAR(20) DEFAULT NULL;

-- 删除字段
ALTER TABLE users DROP COLUMN mobile;

-- 添加索引
ALTER TABLE articles ADD INDEX idx_status (status);

-- 添加外键
ALTER TABLE comments ADD CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE comments ADD CONSTRAINT fk_comments_article FOREIGN KEY (article_id) REFERENCES articles(id);

-- 删除外键
ALTER TABLE comments DROP FOREIGN KEY fk_comments_user;

-- 查看表结构
DESC users;
SHOW CREATE TABLE users\G
```

#### 5.2.4 索引操作

```sql
-- 创建索引
CREATE INDEX idx_username ON users(username);
CREATE UNIQUE INDEX idx_email ON users(email);
CREATE INDEX idx_composite ON articles(user_id, status);

-- 查看索引
SHOW INDEX FROM users;

-- 删除索引
DROP INDEX idx_username ON users;

-- 联合索引最左前缀原则
-- 索引 idx_composite(user_id, status) 可用于：
--   ✓ WHERE user_id = 1
--   ✓ WHERE user_id = 1 AND status = 1
--   ✗ WHERE status = 1
```

### 5.3 DML 数据操作语言

#### 5.3.1 插入数据

```sql
-- 插入单条
INSERT INTO users (username, password, email) VALUES ('zhangsan', 'hashed_pwd', 'zhangsan@example.com');

-- 插入多条
INSERT INTO users (username, password, email) VALUES
    ('zhangsan', 'hashed_pwd', 'zhangsan@example.com'),
    ('lisi', 'hashed_pwd', 'lisi@example.com'),
    ('wangwu', 'hashed_pwd', 'wangwu@example.com');

-- 插入并获取自增ID
INSERT INTO users (username, password, email) VALUES ('test', 'pwd', 'test@example.com');
SELECT LAST_INSERT_ID();

-- 蠕虫复制（测试用）
INSERT INTO articles (title, content, user_id) SELECT title, content, user_id FROM articles;
```

#### 5.3.2 更新数据

```sql
-- 更新单条
UPDATE users SET nickname = '张三', status = 1 WHERE id = 1;

-- 批量更新
UPDATE users SET status = 0 WHERE deleted_at IS NOT NULL;

-- 乐观锁更新（版本号控制）
UPDATE products SET stock = stock - 1, version = version + 1 WHERE id = 1 AND version = 5;

-- 更新多个字段
UPDATE users SET
    nickname = '新昵称',
    avatar = 'https://example.com/avatar.jpg',
    updated_at = NOW()
WHERE id = 1;
```

#### 5.3.3 删除数据

```sql
-- 删除单条
DELETE FROM users WHERE id = 1;

-- 物理删除（危险！）
DELETE FROM users WHERE deleted_at < '2024-01-01';

-- 软删除（推荐）
UPDATE users SET deleted_at = NOW() WHERE id = 1;

-- 清空表（危险！自增ID重置）
TRUNCATE TABLE users;
```

### 5.4 DQL 数据查询语言

#### 5.4.1 基础查询

```sql
-- 查询所有字段
SELECT * FROM users;

-- 查询指定字段
SELECT id, username, email FROM users;

-- 去重查询
SELECT DISTINCT status FROM articles;

-- 限制条数
SELECT * FROM users LIMIT 10;

-- 分页查询（OFFSET）
SELECT * FROM users LIMIT 10 OFFSET 20;
-- 或
SELECT * FROM users LIMIT 20, 10;

-- 排序
SELECT * FROM articles ORDER BY created_at DESC;        -- 降序
SELECT * FROM articles ORDER BY created_at ASC;         -- 升序
SELECT * FROM articles ORDER BY status ASC, id DESC;    -- 多字段排序

-- 条件查询
SELECT * FROM users WHERE status = 1 AND deleted_at IS NULL;
SELECT * FROM articles WHERE title LIKE '%Go%';        -- 模糊匹配
SELECT * FROM articles WHERE id IN (1, 2, 3);           -- IN查询
SELECT * FROM articles WHERE id BETWEEN 10 AND 20;      -- 范围查询
```

#### 5.4.2 聚合函数

```sql
-- 计数
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM users WHERE status = 1;

-- 求和
SELECT SUM(view_count) FROM articles;

-- 平均值
SELECT AVG(price) FROM products;

-- 最大/最小
SELECT MAX(price), MIN(price) FROM products;

-- 分组聚合
SELECT status, COUNT(*) as total FROM articles GROUP BY status;

-- 分组后筛选（HAVING）
SELECT user_id, COUNT(*) as article_count
FROM articles
GROUP BY user_id
HAVING article_count > 5;
```

#### 5.4.3 多表查询（JOIN）

```sql
-- 内连接（只保留两边都有的）
SELECT a.title, u.username
FROM articles a
INNER JOIN users u ON a.user_id = u.id;

-- 左连接（保留左边全部）
SELECT a.title, u.username
FROM articles a
LEFT JOIN users u ON a.user_id = u.id;

-- 右连接（保留右边全部）
SELECT a.title, u.username
FROM articles a
RIGHT JOIN users u ON a.user_id = u.id;

-- 多表连接
SELECT c.content, u.username, a.title
FROM comments c
INNER JOIN users u ON c.user_id = u.id
INNER JOIN articles a ON c.article_id = a.id;

-- 连接3张以上表
SELECT o.id, u.username, p.name, oi.quantity
FROM orders o
INNER JOIN users u ON o.user_id = u.id
INNER JOIN order_items oi ON o.id = oi.order_id
INNER JOIN products p ON oi.product_id = p.id;
```

#### 5.4.4 子查询

```sql
-- WHERE 子查询
SELECT * FROM articles WHERE user_id IN (
    SELECT id FROM users WHERE status = 1
);

-- FROM 子查询
SELECT avg_count FROM (
    SELECT user_id, COUNT(*) as avg_count
    FROM articles
    GROUP BY user_id
) as tmp
WHERE avg_count > 3;

-- EXISTS 子查询
SELECT * FROM users u WHERE EXISTS (
    SELECT 1 FROM articles a WHERE a.user_id = u.id
);
```

#### 5.4.5 高级查询

```sql
-- UNION 合并查询
SELECT username FROM users WHERE status = 1
UNION
SELECT username FROM admins WHERE status = 1;

-- UNION ALL（保留重复）
SELECT username FROM users
UNION ALL
SELECT username FROM admins;

-- CASE 条件表达式
SELECT
    username,
    status,
    CASE
        WHEN status = 0 THEN '禁用'
        WHEN status = 1 THEN '正常'
        ELSE '未知'
    END as status_text
FROM users;

-- IF 函数
SELECT username, IF(status = 1, '正常', '禁用') as status_text FROM users;

-- IFNULL 处理NULL
SELECT username, IFNULL(nickname, username) as display_name FROM users;

-- COALESCE 返回第一个非NULL值
SELECT COALESCE(nickname, username, '匿名用户') as display_name FROM users;

-- GROUP_CONCAT 合并字符串
SELECT user_id, GROUP_CONCAT(title) as articles
FROM articles
GROUP BY user_id;
```

### 5.5 DCL 数据控制语言

```sql
-- 创建用户
CREATE USER 'blog'@'localhost' IDENTIFIED BY 'strong_password';

-- 授权
GRANT SELECT, INSERT, UPDATE, DELETE ON blog.* TO 'blog'@'localhost';
GRANT ALL PRIVILEGES ON blog.* TO 'blog'@'localhost';

-- 查看权限
SHOW GRANTS FOR 'blog'@'localhost';

-- 撤销权限
REVOKE DELETE ON blog.* FROM 'blog'@'localhost';

-- 刷新权限
FLUSH PRIVILEGES;

-- 删除用户
DROP USER 'blog'@'localhost';
```

### 5.6 TCL 事务控制语言

```sql
-- 开启事务
START TRANSACTION;
-- 或
BEGIN;

-- 提交事务
COMMIT;

-- 回滚事务
ROLLBACK;

-- 设置保存点
SAVEPOINT sp1;

-- 回滚到保存点
ROLLBACK TO sp1;

-- 释放保存点
RELEASE SAVEPOINT sp1;

-- 自动提交关闭（会话级）
SET autocommit = 0;

-- 事务示例
START TRANSACTION;
    INSERT INTO users (username, password, email) VALUES ('test', 'pwd', 'test@example.com');
    INSERT INTO articles (title, content, user_id) VALUES ('测试', '内容', LAST_INSERT_ID());
    -- 如果前面任何一条失败，ROLLBACK 全部回滚
COMMIT;
```

### 5.7 MySQL 事务隔离级别

| 隔离级别 | 脏读 | 不可重复读 | 幻读 |
|----------|------|-----------|------|
| READ UNCOMMITTED | 可能 | 可能 | 可能 |
| READ COMMITTED | 不可能 | 可能 | 可能 |
| REPEATABLE READ（默认） | 不可能 | 不可能 | 可能 |
| SERIALIZABLE | 不可能 | 不可能 | 不可能 |

```sql
-- 查看当前隔离级别
SELECT @@tx_isolation;
-- MySQL 8.0+
SELECT @@transaction_isolation;

-- 设置隔离级别（会话级）
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- 设置隔离级别（全局）
SET GLOBAL TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

### 5.8 连接池详解

#### 5.8.1 什么是连接池？

数据库连接池是预建立并保持的数据库连接集合，避免每次请求都新建连接。

```
┌─────────────────────────────────────────┐
│                 应用                      │
│  请求1 ──┐                              │
│  请求2 ──┼──► 连接池 ───► 数据库        │
│  请求3 ──┘    │                         │
│              保持 N 个空闲连接复用       │
└─────────────────────────────────────────┘
```

#### 5.8.2 为什么需要连接池？

| 方式 | 耗时 | 并发支持 |
|------|------|----------|
| 每次请求新建连接 | 10-100ms | 差 |
| 使用连接池 | 0.01ms | 好 |

#### 5.8.3 连接池参数配置

```sql
-- MySQL 连接池配置（在 MySQL 配置文件中）
[mysqld]
max_connections = 200          -- 最大连接数
wait_timeout = 28800           -- 空闲连接超时（秒）
interactive_timeout = 28800    -- 交互连接超时
```

#### 5.8.4 Go 连接池配置（GORM 示例）

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    ConnPool: &sql.Tx{},
})

sqlDB, _ := db.DB()

// 最大空闲连接数
sqlDB.SetMaxIdleConns(10)

// 最大打开连接数
sqlDB.SetMaxOpenConns(100)

// 连接最大生命周期
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 5.9 MySQL 性能优化

#### 5.9.1 慢查询日志

```sql
-- 查看慢查询配置
SHOW VARIABLES LIKE 'slow_query%';
SHOW VARIABLES LIKE 'long_query_time';

-- 开启慢查询日志
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;  -- 超过1秒记录

-- 查看慢查询日志
SHOW FULL PROCESSLIST;
```

#### 5.9.2 EXPLAIN 执行计划

```sql
-- 分析查询
EXPLAIN SELECT * FROM users WHERE username = 'test';
EXPLAIN SELECT a.title, u.username
FROM articles a
INNER JOIN users u ON a.user_id = u.id;

-- 关键字段解读
-- type: ALL（全表扫描） < index < range < ref < eq_ref < const
-- key: 实际使用的索引
-- rows: 预计扫描行数
-- Extra: 额外信息（Using filesort, Using index 等）
```

#### 5.9.3 SQL 优化原则

```sql
-- ✓ 推荐：使用索引字段查询
SELECT * FROM users WHERE id = 1;
SELECT * FROM users WHERE email = 'test@example.com';  -- 有UNIQUE索引

-- ✗ 避免：前导通配符模糊查询
SELECT * FROM users WHERE username LIKE '%zhang%';    -- 无法使用索引

-- ✓ 推荐：LIMIT 分页
SELECT * FROM articles ORDER BY id DESC LIMIT 10;

-- ✗ 避免：深分页（OFFSET 过大）
SELECT * FROM articles LIMIT 1000000, 10;  -- 需扫描100万行

-- ✓ 推荐：子查询改 JOIN
SELECT * FROM users WHERE id IN (SELECT user_id FROM articles);

-- ✓ 推荐：批量操作
INSERT INTO users VALUES (1, 'a'), (2, 'b'), (3, 'c');
```

### 5.10 常用命令速查

```sql
-- 库操作
SHOW DATABASES;
CREATE DATABASE xxx;
DROP DATABASE xxx;
USE xxx;

-- 表操作
SHOW TABLES;
DESC tablename;
SHOW CREATE TABLE tablename\G

-- 数据操作
SELECT ... FROM ... WHERE ...;
INSERT INTO ... VALUES ...;
UPDATE ... SET ... WHERE ...;
DELETE FROM ... WHERE ...;

-- 索引操作
SHOW INDEX FROM tablename;
EXPLAIN ...;

-- 用户权限
SHOW GRANTS FOR 'user'@'host';
GRANT ... ON ... TO 'user'@'host';

-- 系统变量
SHOW VARIABLES LIKE '%xxx%';
SET GLOBAL xxx = value;

-- 进程
SHOW FULL PROCESSLIST;
KILL process_id;

-- 表分析
ANALYZE TABLE tablename;
OPTIMIZE TABLE tablename;
CHECK TABLE tablename;
REPAIR TABLE tablename;
```

---

## 十、企业模板 API 大全

本文档统计项目中所有**模板代码**提供的 API 和方法，分为三大类：

| 分类 | 来源 | 说明 |
|------|------|------|
| **核心模块** | 项目内部 core/ | 数据库、日志、配置、响应封装 |
| **路由模块** | api/router/ | Gin 路由定义 |
| **工具模块** | flags/ | 命令行参数解析 |

---

### 10.1 core/logger 日志模块

#### 来源
```go
import "github.com/GoWeb/My_Blog/core/logger"
```

#### 核心 API

| 方法 | 签名 | 功能 |
|------|------|------|
| `logger.Init` | `Init(cfg Config)` | 初始化日志（main.go 中调用一次） |
| `logger.S.Sync()` | `Sync()` | 刷新缓冲区（defer 调用） |
| `logger.S.Info` | `Info(msg string, fields ...Field)` | 记录 Info 级别日志 |
| `logger.S.Infow` | `Infow(msg string, keysAndValues ...interface{})` | 结构化 Info 日志 |
| `logger.S.Warn` | `Warn(msg string, fields ...Field)` | 记录 Warn 级别日志 |
| `logger.S.Warnf` | `Warnf(format string, args ...interface{})` | 格式化 Warn 日志 |
| `logger.S.Error` | `Error(msg string, fields ...Field)` | 记录 Error 级别日志 |
| `logger.S.Errorw` | `Errorw(msg string, keysAndValues ...interface{})` | 结构化 Error 日志 |
| `logger.S.Debug` | `Debug(msg string, fields ...Field)` | 记录 Debug 级别日志 |
| `logger.S.Debugw` | `Debugw(msg string, keysAndValues ...interface{})` | 结构化 Debug 日志 |

#### 使用示例

```go
// 普通日志
logger.S.Info("用户注册成功")

// 结构化日志（推荐）
logger.S.Infow("用户登录",
    "user_id", 1,
    "nickname", "张三",
    "ip", "192.168.1.1",
)

// 格式化日志
logger.S.Warnf("数据库连接失败: %v", err)

// 错误日志
logger.S.Errorw("创建用户失败",
    "error", err.Error(),
    "email", email,
)
```

#### 输出效果
```json
{"level":"info","msg":"用户登录","user_id":1,"nickname":"张三","ip":"192.168.1.1","time":"2024-01-01T12:00:00Z"}
```

---

### 10.2 core/response 响应封装模块

#### 来源
```go
import "github.com/GoWeb/My_Blog/core/response"
```

#### 状态码常量

| 常量 | 值 | 说明 |
|------|------|------|
| `CodeSuccess` | 0 | 成功 |
| `CodeParamError` | 400 | 参数错误 |
| `CodeUnauthorized` | 401 | 未认证 |
| `CodeForbidden` | 403 | 无权限 |
| `CodeNotFound` | 404 | 资源不存在 |
| `CodeInternalError` | 500 | 服务器内部错误 |
| `CodeDatabaseError` | 50001 | 数据库错误 |
| `CodeBusinessError` | 50002 | 业务逻辑错误 |
| `CodeValidationError` | 40001 | 数据校验失败 |
| `CodeDuplicateError` | 40002 | 数据重复冲突 |

#### 成功响应

| 方法 | 签名 | 功能 |
|------|------|------|
| `response.Success` | `Success(c *gin.Context, data interface{})` | 成功响应（默认 msg="success"） |
| `response.SuccessWithMsg` | `SuccessWithMsg(c *gin.Context, data interface{}, msg string)` | 成功响应（自定义消息） |
| `response.SuccessPage` | `SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int)` | 分页响应 |

**使用示例：**
```go
// 最常用：成功响应
response.Success(c, nil)

// 自定义消息
response.SuccessWithMsg(c, nil, "注册成功")

// 返回数据
response.Success(c, map[string]interface{}{
    "user_id": 1,
    "token": "abc123",
})

// 分页响应（列表查询）
response.SuccessPage(c, userList, total, 1, 10)
```

**响应格式：**
```json
// Success
{"code": 0, "msg": "success", "data": {...}}

// SuccessPage
{"code": 0, "msg": "success", "data": {"list": [...], "total": 100, "page": 1, "page_size": 10, "total_pages": 10}}
```

#### 错误响应

| 方法 | 签名 | 功能 |
|------|------|------|
| `response.Error` | `Error(c *gin.Context, msg string)` | 通用错误（code=500） |
| `response.ErrorWithCode` | `ErrorWithCode(c *gin.Context, code int, msg string)` | 自定义错误码 |
| `response.ParamError` | `ParamError(c *gin.Context)` | 参数错误（默认"invalid parameter"） |
| `response.ParamErrorWithMsg` | `ParamErrorWithMsg(c *gin.Context, msg string)` | 参数错误（自定义消息） |
| `response.Unauthorized` | `Unauthorized(c *gin.Context)` | 未认证（HTTP 401） |
| `response.UnauthorizedWithMsg` | `UnauthorizedWithMsg(c *gin.Context, msg string)` | 未认证（自定义消息） |
| `response.Forbidden` | `Forbidden(c *gin.Context)` | 无权限（HTTP 403） |
| `response.ForbiddenWithMsg` | `ForbiddenWithMsg(c *gin.Context, msg string)` | 无权限（自定义消息） |
| `response.NotFound` | `NotFound(c *gin.Context)` | 资源不存在（HTTP 404） |
| `response.NotFoundWithMsg` | `NotFoundWithMsg(c *gin.Context, msg string)` | 资源不存在（自定义消息） |
| `response.InternalError` | `InternalError(c *gin.Context)` | 服务器内部错误（HTTP 500） |
| `response.InternalErrorWithMsg` | `InternalErrorWithMsg(c *gin.Context, msg string)` | 服务器错误（自定义消息） |
| `response.DatabaseError` | `DatabaseError(c *gin.Context)` | 数据库错误（code=50001） |
| `response.BusinessError` | `BusinessError(c *gin.Context, msg string)` | 业务错误（code=50002） |
| `response.ValidationError` | `ValidationError(c *gin.Context, msg string)` | 数据校验失败（code=40001） |
| `response.DuplicateError` | `DuplicateError(c *gin.Context, msg string)` | 数据重复冲突（code=40002） |

**使用示例：**
```go
// 参数校验失败
if err := c.ShouldBindJSON(&req); err != nil {
    response.ParamErrorWithMsg(c, "参数错误："+err.Error())
    return
}

// 数据已存在
if count > 0 {
    response.DuplicateError(c, "昵称已被使用")
    return
}

// 数据库操作失败
if err := db.Create(&user).Error; err != nil {
    response.DatabaseError(c)
    return
}

// 业务逻辑错误
if balance < amount {
    response.BusinessError(c, "余额不足")
    return
}
```

**错误响应格式：**
```json
{"code": 40001, "msg": "昵称已被使用", "data": null}
```

---

### 10.3 core/init_db 数据库模块

#### 来源
```go
import "github.com/GoWeb/My_Blog/core"
```

#### 核心 API

| 方法 | 签名 | 功能 |
|------|------|------|
| `core.ReadConfig` | `ReadConfig(mode string) error` | 加载配置文件 |
| `core.InitDB` | `InitDB(opts ...*DBOptions) error` | 初始化数据库连接 |
| `core.CloseDB` | `CloseDB() error` | 关闭数据库连接 |
| `core.GetDB` | `GetDB() *gorm.DB` | 获取数据库实例 |
| `core.AutoMigrate` | `AutoMigrate(models ...interface{}) error` | 自动迁移表结构 |
| `core.PingDB` | `PingDB() error` | 健康检查 |

#### 使用示例

```go
// main.go 中初始化
core.ReadConfig(flags.Opt.Mode)  // 先加载配置
core.InitDB()                     // 再连接数据库
defer core.CloseDB()             // 程序结束时关闭

// 控制器中获取数据库实例
db := core.GetDB()

// 自动建表（仅开发环境）
if !flags.IsRelease() {
    core.AutoMigrate(&model.User{}, &model.Article{})
}
```

#### GORM CRUD 常用方法

| 方法 | 签名 | 功能 |
|------|------|------|
| `db.Create` | `Create(value interface{}) *DB` | 插入数据 |
| `db.First` | `First(dest interface{}, conds ...interface{}) *DB` | 查询第一条 |
| `db.Find` | `Find(dest interface{}, conds ...interface{}) *DB` | 查询多条 |
| `db.Where` | `Where(query interface{}, args ...interface{}) *DB` | 条件查询 |
| `db.Model` | `Model(value interface{}) *DB` | 指定模型/表 |
| `db.Create` | `Create(value interface{}) *DB` | 插入数据 |
| `db.Updates` | `Updates(values interface{}) *DB` | 更新数据 |
| `db.Delete` | `Delete(value interface{}, conds ...interface{}) *DB` | 删除数据 |
| `db.Count` | `Count(count *int64) *DB` | 统计数量 |
| `db.Exec` | `Exec(sql string, values ...interface{}) *DB` | 执行原生 SQL |

#### GORM 查询示例

```go
db := core.GetDB()

// 插入
db.Create(&user)

// 查询单条
var user model.User
db.First(&user, 1)                                      // 根据主键
db.Where("nickname = ?", "张三").First(&user)           // 条件查询

// 查询多条
var users []model.User
db.Where("status = ?", 1).Find(&users)

// 统计数量
var count int64
db.Model(&model.User{}).Where("status = ?", 1).Count(&count)

// 更新
db.Model(&user).Updates(map[string]interface{}{
    "nickname": "新昵称",
    "avatar": "http://xxx.jpg",
})

// 删除（软删除，需 DeletedAt 字段）
db.Delete(&user)

// 原生 SQL
db.Exec("UPDATE users SET status = 0 WHERE id = ?", 1)
```

---

### 10.4 flags 命令行参数模块

#### 来源
```go
import "github.com/GoWeb/My_Blog/flags"
```

#### 核心 API

| 方法 | 签名 | 功能 |
|------|------|------|
| `flags.Parse` | `Parse()` | 解析命令行参数（main.go 第一步调用） |
| `flags.Opt` | `Opt` 全局变量 | 命令行选项结构体 |
| `flags.IsRelease` | `IsRelease() bool` | 是否生产环境 |
| `flags.GetAddr` | `GetAddr() string` | 获取服务地址 |

#### Opt 结构体字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `Opt.Mode` | string | 运行模式：dev/test/prod |
| `Opt.Version` | bool | 是否显示版本号 |
| `Opt.Addr` | string | 服务地址（默认 ":8080"） |

#### 使用示例

```go
// main.go 中
flags.Parse()                    // 解析命令行
if flags.Opt.Version {
    fmt.Println("v1.0.0")
    return
}
if flags.IsRelease() {
    // 生产环境逻辑
}
addr := flags.GetAddr()         // 获取地址，如 ":8080" 或 ":80"
```

#### 命令行用法

```bash
go run main.go                  # 默认 dev 模式，端口 :8080
go run main.go -m=prod         # 生产模式
go run main.go -m=test         # 测试模式
go run main.go -v              # 显示版本号
go run main.go -addr=:80       # 指定端口
go run main.go -m=prod -addr=:80  # 组合使用
```

---

### 10.5 Gin 框架路由方法

#### 来源
```go
import "github.com/gin-gonic/gin"
```

#### HTTP 方法

| 方法 | 作用 | 路由示例 |
|------|------|----------|
| `r.GET` | 查询（GET 请求） | `r.GET("/user/:id", handler)` |
| `r.POST` | 创建（POST 请求） | `r.POST("/user", handler)` |
| `r.PUT` | 更新（PUT 请求） | `r.PUT("/user/:id", handler)` |
| `r.DELETE` | 删除（DELETE 请求） | `r.DELETE("/user/:id", handler)` |
| `r.PATCH` | 部分更新（PATCH 请求） | `r.PATCH("/user/:id", handler)` |
| `r.Group` | 分组路由 | `apiGroup := r.Group("/api")` |

#### 上下文获取请求参数

| 方法 | 签名 | 功能 |
|------|------|------|
| `c.ShouldBindJSON` | `ShouldBindJSON(obj interface{}) error` | 解析 JSON 请求体 |
| `c.ShouldBindQuery` | `ShouldBindQuery(obj interface{}) error` | 解析 URL Query 参数 |
| `c.Param` | `Param(key string) string` | 获取路径参数 |
| `c.Query` | `Query(key string) string` | 获取 URL 查询参数 |
| `c.PostForm` | `PostForm(key string) string` | 获取表单参数 |

#### 使用示例

```go
// JSON 请求体
type RegisterReq struct {
    Nickname string `json:"nickname"`
    Password string `json:"password"`
}
var req RegisterReq
c.ShouldBindJSON(&req)

// URL Query 参数：/user/list?page=1&page_size=10
type ListReq struct {
    Page     int `form:"page"`
    PageSize int `form:"page_size"`
}
var req ListReq
c.ShouldBindQuery(&req)

// 路径参数：/user/:id
id := c.Param("id")  // 获取 :id 的值

// URL 参数：/search?q=keyword
q := c.Query("q")
```

---

### 10.6 bcrypt 密码加密

#### 来源
```go
import "golang.org/x/crypto/bcrypt"
```

#### 核心 API

| 方法 | 签名 | 功能 |
|------|------|------|
| `bcrypt.GenerateFromPassword` | `GenerateFromPassword(password []byte, cost int) ([]byte, error)` | 加密密码 |
| `bcrypt.CompareHashAndPassword` | `CompareHashAndPassword(hashedPassword, password []byte) error` | 验证密码 |

#### 使用示例

```go
// 加密（注册时）
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
    response.InternalError(c)
    return
}
// 存储 hashedPassword 到数据库

// 验证（登录时）
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
if err != nil {
    response.Error(c, "密码错误")
    return
}
```

#### 参数说明

| 参数 | 说明 |
|------|------|
| `password` | 明文密码 |
| `cost` | 加密强度，范围 4-31，默认 10。数字越大越安全但越慢 |

---

### 10.7 项目路由汇总

#### 路由结构
```
api/router/
├── router.go              ← 主路由入口
└── nestedrouter/
    ├── user.go            ← 用户模块路由
    ├── article.go         ← 文章模块路由
    └── comment.go         ← 评论模块路由
```

#### RESTful API 路由表

| 模块 | 方法 | 路径 | 功能 | 控制器函数 |
|------|------|------|------|------------|
| user | POST | /api/user/register | 用户注册 | Register |
| user | POST | /api/user/login | 用户登录 | Login |
| user | GET | /api/user/info | 获取用户信息 | Info |
| article | POST | /api/article/create | 创建文章 | CreateArticle |
| article | GET | /api/article/list | 文章列表 | ArticleList |
| article | GET | /api/article/detail/:id | 文章详情 | ArticleDetail |
| article | PUT | /api/article/update/:id | 更新文章 | UpdateArticle |
| article | DELETE | /api/article/delete/:id | 删除文章 | DeleteArticle |
| comment | POST | /api/comment/create | 创建评论 | CreateComment |
| comment | GET | /api/comment/list | 评论列表 | CommentList |
| comment | DELETE | /api/comment/delete/:id | 删除评论 | DeleteComment |

#### RESTful 风格对照

| HTTP 方法 | 语义 | 示例 |
|-----------|------|------|
| GET | 查询 | GET /user/1 查询用户 |
| POST | 创建 | POST /user 创建用户 |
| PUT | 更新（全部） | PUT /user/1 更新用户全部字段 |
| DELETE | 删除 | DELETE /user/1 删除用户 |

---

### 10.8 一图看懂 API 调用关系

```
前端请求
    │
    ▼
┌─────────────────────────────────────┐
│ Gin 路由 (api/router/)              │
│   r.POST("/user/register", Register)│
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│ 控制器 (api/controller/)            │
│   Register(c *gin.Context)          │
│     │                               │
│     ├── c.ShouldBindJSON(&req)     │ ← Gin 解析请求
│     │                               │
│     ├── core.GetDB()               │ ← 获取数据库
│     │                               │
│     ├── db.Where().Count()         │ ← GORM 查询
│     │                               │
│     ├── bcrypt.GenerateFromPassword│ ← bcrypt 加密
│     │                               │
│     ├── db.Create()                 │ ← GORM 插入
│     │                               │
│     └── response.SuccessWithMsg()   │ ← 统一响应
└─────────────────────────────────────┘
    │
    ▼
前端收到：{"code": 0, "msg": "注册成功", "data": null}
```

---

## 十一、main.go 标准启动模板

### 7.1 完整代码

```go
package main

import (
	"fmt"

	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/logger"
	"github.com/GoWeb/My_Blog/flags"
	"github.com/GoWeb/My_Blog/models"
	"go.uber.org/zap/zapcore"
)

func main() {
	// ========== 第1步：解析命令行参数 ==========
	flags.Parse()

	// ========== 第2步：处理特殊参数 ==========
	if flags.Opt.Version {
		fmt.Println("v1.0.0")
		return
	}

	// ========== 第3步：加载配置 ==========
	if err := core.ReadConfig(flags.Opt.Mode); err != nil {
		fmt.Println("配置加载失败:", err)
		return
	}

	// ========== 第4步：初始化日志 ==========
	// 日志级别从配置文件中读取（app.yaml 的 log.level）
	level := parseLogLevel(core.GlobalConfig.Log.Level)
	logger.Init(logger.Config{Level: level})
	defer logger.Sync()

	// ========== 第5步：初始化数据库 ==========
	if err := core.InitDB(); err != nil {
		logger.S.Fatalf("数据库连接失败: %v", err)
	}
	defer core.CloseDB()

	// ========== 第5.5步：自动建表（仅开发/测试环境） ==========
	if !flags.IsRelease() {
		if err := core.AutoMigrate(models.AllModels()...); err != nil {
			logger.S.Warnf("自动建表失败: %v", err)
		}
	}

	// ========== 第6步：获取服务地址 ==========
	addr := flags.GetAddr()

	// ========== 第7步：打印启动日志 ==========
	logger.S.Infow("程序启动",
		"mode", flags.Opt.Mode,
		"config", fmt.Sprintf("app.yaml + app-%s.yaml", flags.Opt.Mode),
		"addr", addr,
	)

	// ========== 第8步：启动 Gin HTTP 服务（需要你添加） ==========
}

func parseLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
```

### 7.2 模板使用手册（新手必看）

**Q：新项目用这个模板时，main.go 里哪些需要改？**

#### 不用改的（模板固定写法）：

```go
// 第1步：解析命令行 - 不用改
flags.Parse()

// 第2步：处理版本号 - 不用改
if flags.Opt.Version { fmt.Println("v1.0.0"); return }

// 第3步：初始化日志 - 不用改（自动根据环境选择级别）
if flags.IsRelease() { logger.Init(...) } else { logger.Init(...) }

// 第4步：加载配置 - 不用改（自动合并配置）
core.ReadConfig(flags.Opt.Mode)

// 第5步：初始化数据库 - 不用改
core.InitDB()
defer core.CloseDB()

// 第6步：获取地址 - 不用改
addr := flags.GetAddr()

// 第7步：打印日志 - 不用改
logger.S.Infow("程序启动", ...)
```

#### 需要你添加的（第8步）：

```go
// 在第7步之后添加你自己的业务代码
// 这是唯一需要你写的地方！

// 例如启动 Gin：
r := gin.Default()
r.GET("/ping", func(c *gin.Context) {
    c.JSON(200, gin.H{"msg": "pong"})
})
logger.S.Infow("Gin 服务启动", "addr", addr)
if err := r.Run(addr); err != nil {
    logger.S.Fatalf("服务启动失败: %v", err)
}
```

#### 需要改 import 的情况：

```go
// 如果你的项目名不是 "My_Blog"，需要改这里：
import (
    "github.com/你的项目名/core"        // 改
    "github.com/你的项目名/core/logger" // 改
    "github.com/你的项目名/flags"       // 改
    // 其他 import 保持不变
)
```

### 7.3 一图看懂 main.go 模板结构

```
┌─────────────────────────────────────────────────────────────┐
│                      main.go 模板结构                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  import (固定)                                              │
│       │                                                    │
│       ▼                                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 第1步 flags.Parse()           ←── 固定，不用改        │   │
│  │ 第2步 处理 -v 参数             ←── 固定，不用改        │   │
│  │ 第3步 logger.Init()            ←── 固定，不用改        │   │
│  │ 第4步 core.ReadConfig()        ←── 固定，不用改        │   │
│  │ 第5步 core.InitDB()           ←── 固定，不用改        │   │
│  │ 第6步 flags.GetAddr()         ←── 固定，不用改        │   │
│  │ 第7步 logger.S.Infow()        ←── 固定，不用改        │   │
│  │                                                     │   │
│  │ ★ 第8步 启动服务               ←── 你要写这里！        │   │
│  │   例：r := gin.Default()                          │   │
│  │        r.Run(addr)                               │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 7.4 常见问题

| 问题 | 答案 |
|------|------|
| Q: import 怎么改？ | 把 `"github.com/GoWeb/My_Blog/..."` 改成你项目的 import 路径 |
| Q: 第8步不写会怎样？ | 程序启动后什么都没做就退出了 |
| Q: 能删掉某些步吗？ | 不能，1-7步是启动必要流程 |
| Q: flags.Opt.Mode 能改吗？ | 不能，这是命令行传入的环境参数 |
| Q: 日志级别在哪配？ | 在 `app-{env}.yaml` 的 `log.level` 配置 |

| 步骤 | 代码 | 说明 |
|------|------|------|
| 1 | `flags.Parse()` | 解析命令行参数 |
| 2 | `flags.Opt.Version` | 处理 -v 参数（显示版本） |
| 3 | `core.ReadConfig()` | 加载配置文件 |
| 4 | `logger.Init()` | 从配置读取日志级别并初始化 |
| 5 | `core.InitDB()` | 连接数据库 |
| 6 | `flags.GetAddr()` | 获取监听地址 |
| 7 | `logger.S.Infow()` | 打印启动信息 |
| 8 | `r.Run(addr)` | **启动 Gin HTTP 服务（你需要添加）** |

### 7.5 启动流程图

```
程序启动（main 函数入口）
        │
        ▼
┌───────────────────┐
│ flags.Parse()      │ 读取命令行参数
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ logger.Init()     │ 初始化日志系统
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ ReadConfig()      │ 加载配置文件
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ InitDB()          │ 连接数据库
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ GetAddr()         │ 获取服务地址
└───────────────────┘
        │
        ▼
   启动 HTTP 服务
```

---

## 十二、项目架构与模块关系

### 8.1 模块依赖关系

```
┌─────────────────────────────────────────────────────┐
│                     main.go                          │
│                                                      │
│  1. flags.Parse()        ←── 命令行参数模块           │
│  2. logger.Init()        ←── 日志模块（依赖 Config） │
│  3. core.ReadConfig()    ←── 配置模块（依赖 Viper）   │
│  4. core.InitDB()        ←── 数据库模块（依赖 Config）│
└─────────────────────────────────────────────────────┘
         │              │              │
         ▼              ▼              ▼
   ┌──────────┐   ┌──────────┐   ┌──────────┐
   │  flags   │   │  logger  │   │   core   │
   └──────────┘   └──────────┘   └──────────┘
```

### 8.2 调用顺序

```
flags.Parse()          // 第1个调用，解析命令行
       ↓
logger.Init()         // 第2个调用，初始化日志
       ↓
core.ReadConfig()     // 第3个调用，加载配置
       ↓
core.InitDB()         // 第4个调用，连接数据库
       ↓
flags.GetAddr()       // 第5个调用，获取服务地址
       ↓
启动服务
```

### 8.3 各模块职责

| 模块 | 职责 | 被依赖 |
|------|------|--------|
| flags | 命令行参数解析 | main.go |
| logger | 日志记录 | main.go, init_db.go |
| init_config | 配置文件读取 | init_db.go |
| init_db | 数据库连接 | DAO 层 |

---

## 十三、快速复制模板

### 9.1 logger 使用模板

```go
// ========== 初始化（程序启动时）==========
logger.Init()
defer logger.Sync()

// ========== 使用 ==========
logger.S.Info("普通信息")
logger.S.Infof("用户%s登录成功", username)
logger.S.Infow("登录", "user", username, "ip", ip)
logger.S.Errorw("失败", "reason", "密码错误")
logger.S.Debugw("调试", "sql", sql)
logger.S.Warnw("警告", "memory", "80%")
logger.S.Fatalw("致命", "error", err)
```

### 9.2 flags 使用模板

```go
// ========== 解析（必须第一个）==========
flags.Parse()

// ========== 获取值 ==========
port := flags.Opt.Port
mode := flags.Opt.Mode
configFile := flags.Opt.ConfigFile
isDbInit := flags.Opt.DBInit

// ========== 判断 ==========
if flags.IsRelease() { }
if flags.IsDev() { }
if flags.IsTest() { }

// ========== 获取地址 ==========
addr := flags.GetAddr()

// ========== 环境变量 ==========
envPort := flags.GetEnv("PORT", "8080")

// ========== 校验 ==========
if err := flags.ValidateFlags(); err != nil { }
if err := flags.CheckRequiredFlags([]string{"config"}); err != nil { }

// ========== 帮助 ==========
flags.PrintUsage()
### 9.3 config 使用模板

```go
// ========== 加载（传入环境名，自动合并）==========
core.ReadConfig("dev")      // 合并 app.yaml + app-dev.yaml
core.ReadConfig("test")     // 合并 app.yaml + app-test.yaml
core.ReadConfig("release")  // 合并 app.yaml + app-prod.yaml

// 等同于
core.ReadConfig(flags.Opt.Mode)

// ========== 使用 ==========
core.GlobalConfig.Server.Port
core.GlobalConfig.Database.Host
core.GlobalConfig.Database.Username
core.GlobalConfig.Database.Password
core.GlobalConfig.Log.Level
core.GlobalConfig.Database.Name
core.GlobalConfig.JWT.Secret
core.GlobalConfig.JWT.Expire
core.GlobalConfig.Email.Host
```

### 9.4 db 使用模板

```go
// ========== 初始化 ==========
core.InitDB()
defer core.CloseDB()

// 或自定义连接池
core.InitDB(&core.DBOptions{
    MaxIdleConns:  10,
    MaxOpenConns:  50,
    ConnMaxLife:   time.Hour,
    SlowThreshold: 500 * time.Millisecond,
})

// ========== 获取实例 ==========
db := core.GetDB()

// ========== DAO 操作 ==========
db.Create(&user)                                      // 创建
db.First(&user, id)                                   // 查询
db.Where("username = ?", name).First(&user)          // 条件查询
db.Save(&user)                                        // 更新
db.Delete(&user)                                      // 删除

// ========== 健康检查 ==========
core.PingDB()

// ========== 自动建表 ==========
// 需要先 import "github.com/你的项目/models"
core.AutoMigrate(models.AllModels()...)  // 自动创建所有模型对应的表
```

### 9.5 main.go 完整模板

```go
package main

import (
	"fmt"

	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/logger"
	"github.com/GoWeb/My_Blog/flags"
	"github.com/GoWeb/My_Blog/models"
	"go.uber.org/zap/zapcore"
)

func main() {
	// ========== 第1步：解析命令行参数 ==========
	flags.Parse()

	// ========== 第2步：处理特殊参数 ==========
	if flags.Opt.Version {
		fmt.Println("v1.0.0")
		return
	}

	// ========== 第3步：加载配置 ==========
	if err := core.ReadConfig(flags.Opt.Mode); err != nil {
		fmt.Println("配置加载失败:", err)
		return
	}

	// ========== 第4步：初始化日志 ==========
	level := parseLogLevel(core.GlobalConfig.Log.Level)
	logger.Init(logger.Config{Level: level})
	defer logger.Sync()

	// ========== 第5步：初始化数据库 ==========
	if err := core.InitDB(); err != nil {
		logger.S.Fatalf("数据库连接失败: %v", err)
	}
	defer core.CloseDB()

	// ========== 第5.5步：自动建表（仅开发/测试环境） ==========
	if !flags.IsRelease() {
		if err := core.AutoMigrate(models.AllModels()...); err != nil {
			logger.S.Warnf("自动建表失败: %v", err)
		}
	}

	// ========== 第6步：获取服务地址 ==========
	addr := flags.GetAddr()

	// ========== 第7步：打印启动日志 ==========
	logger.S.Infow("程序启动",
		"mode", flags.Opt.Mode,
		"config", fmt.Sprintf("app.yaml + app-%s.yaml", flags.Opt.Mode),
		"addr", addr,
	)

	// ========== 第8步：启动 Gin HTTP 服务（需要你添加） ==========
}

func parseLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
```

---

## 十四、常见问题

### 10.1 日志不输出到文件
- 检查 `./logs/` 目录是否存在
- Windows 下路径写法：`".\\logs\\app.log"`

### 10.2 命令行参数不生效
- 确保 `flags.Parse()` 在使用参数前调用
- 参数名不要有下划线

### 10.3 数据库连接失败
- 检查 config/dev.yaml 中的数据库配置
- 确保 MySQL 服务已启动
- 检查用户名密码是否正确

### 10.4 为什么用 zap 不用 log
- zap 性能更高（零内存分配）
- 支持结构化日志
- 支持多种输出（控制台+文件）

### 10.5 配置不生效
- 确保 ReadConfig 在使用配置前调用
- 检查 YAML 格式是否正确（缩进要用空格）

---

## 十一、环境切换汇总

| 环境 | 启动命令 | 日志级别 | 配置文件 |
|------|----------|---------|---------|
| 开发 | `myapp.exe` | Debug | dev.yaml |
| 测试 | `myapp.exe -mode test` | Debug | test.yaml |
| 生产 | `myapp.exe -mode release` | Error | prod.yaml |
