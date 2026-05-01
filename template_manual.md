# Go Gin 企业级模板手册

> 📋 本手册提取自 My_Blog 项目，拿去直接用，改改表名和字段就能跑新项目

---

## 一、项目结构

```
项目名/
├── main.go                      # 程序入口（固定模板）
├── config/                      # 配置文件
│   ├── app.yaml                 # 基础配置
│   └── app-dev.yaml             # 开发环境配置
├── api/
│   ├── controller/              # 控制器（写业务逻辑）
│   │   └── user_controller.go
│   ├── middleware/              # 中间件（认证、限流等）
│   │   └── jwt_auth.go
│   └── router/                  # 路由配置
│       ├── router.go            # 主路由
│       └── nestedrouter/        # 子路由
├── core/                        # 核心模块（别动）
│   ├── response/                # 响应封装
│   │   └── response.go
│   ├── init_config.go          # 配置加载
│   ├── init_db.go              # 数据库初始化
│   └── logger/                  # 日志
├── models/
│   ├── register.go             # 模型注册
│   └── model/                   # 数据模型
│       └── user_model.go
└── flags/
    └── enter.go                # 命令行参数
```

---

## 二、main.go 模板

```go
package main

import (
	"fmt"
	"github.com/项目名/api/router"
	"github.com/项目名/core"
	"github.com/项目名/core/logger"
	"github.com/项目名/flags"
	"github.com/项目名/models"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 1. 解析命令行参数
	flags.Parse()

	// 2. 处理特殊参数（版本号）
	if flags.Opt.Version {
		fmt.Println("v1.0.0")
		return
	}

	// 3. 加载配置
	if err := core.ReadConfig(flags.Opt.Mode); err != nil {
		fmt.Println("配置加载失败:", err)
		return
	}

	// 4. 初始化日志
	level := parseLogLevel(core.GlobalConfig.Log.Level)
	logger.Init(logger.Config{Level: level})
	defer logger.Sync()

	// 5. 初始化数据库
	if err := core.InitDB(); err != nil {
		logger.S.Fatalf("数据库连接失败: %v", err)
	}
	defer core.CloseDB()

	// 6. 自动建表（仅开发/测试环境）
	if !flags.IsRelease() {
		if err := core.AutoMigrate(models.AllModels()...); err != nil {
			logger.S.Warnf("自动建表失败: %v", err)
		}
	}

	// 7. 启动服务
	addr := flags.GetAddr()
	logger.S.Infow("程序启动", "mode", flags.Opt.Mode, "addr", addr)

	r := router.InitRouter()
	if err := r.Run(addr); err != nil {
		logger.S.Fatalf("服务启动失败: %v", err)
	}
}

func parseLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":   return zapcore.DebugLevel
	case "info":    return zapcore.InfoLevel
	case "warn":    return zapcore.WarnLevel
	case "error":   return zapcore.ErrorLevel
	default:        return zapcore.InfoLevel
	}
}
```

---

## 三、core/response 响应封装

### 3.1 响应结构体

```go
type Response struct {
	Code  int         `json:"code"`           // 业务状态码
	Msg   string      `json:"msg"`            // 提示信息
	Data  interface{} `json:"data,omitempty"` // 数据
	Error string      `json:"error,omitempty"`// 错误详情
}

type PageResult struct {
	List       interface{} `json:"list"`        // 数据列表
	Total      int64       `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页
	PageSize   int         `json:"page_size"`  // 每页数量
	TotalPages int         `json:"total_pages"`// 总页数
}
```

### 3.2 状态码常量

```go
const (
	CodeSuccess         = 0      // 成功
	CodeParamError      = 400    // 参数错误
	CodeUnauthorized    = 401    // 未授权
	CodeForbidden       = 403    // 禁止访问
	CodeNotFound        = 404    // 资源不存在
	CodeInternalError   = 500    // 服务器错误
	CodeDatabaseError   = 50001  // 数据库错误
	CodeBusinessError   = 50002  // 业务错误
	CodeValidationError = 40001  // 校验错误
	CodeDuplicateError  = 40002  // 数据重复
)
```

### 3.3 响应方法（必须掌握）

| 方法 | 什么时候用 | 示例 |
|-----|-----------|------|
| `response.Success(c, data)` | 查询成功返回数据 | `response.Success(c, user)` |
| `response.SuccessWithMsg(c, data, "成功消息")` | 操作成功带提示 | `response.SuccessWithMsg(c, gin.H{"id": id}, "创建成功")` |
| `response.SuccessPage(c, list, total, page, size)` | 返回分页列表 | 文章列表、评论列表 |
| `response.ParamError(c)` | 参数错误（默认消息） | `response.ParamError(c)` |
| `response.ParamErrorWithMsg(c, "具体错误")` | 参数错误（自定义消息） | `response.ParamErrorWithMsg(c, "nickname不能为空")` |
| `response.Unauthorized(c)` | 未授权（默认消息） | Token无效 |
| `response.UnauthorizedWithMsg(c, "具体原因")` | 未授权（自定义消息） | `response.UnauthorizedWithMsg(c, "用户不存在")` |
| `response.Forbidden(c)` | 禁止访问（默认消息） | 无权限 |
| `response.ForbiddenWithMsg(c, "具体原因")` | 禁止访问（自定义消息） | `response.ForbiddenWithMsg(c, "无权限删除")` |
| `response.NotFound(c)` | 资源不存在（默认消息） | |
| `response.NotFoundWithMsg(c, "具体资源")` | 资源不存在（自定义消息） | `response.NotFoundWithMsg(c, "文章不存在")` |
| `response.InternalError(c)` | 服务器内部错误 | 数据库连接失败 |
| `response.InternalErrorWithMsg(c, "错误详情")` | 服务器错误（自定义） | |
| `response.DatabaseError(c)` | 数据库操作失败 | 50001 |
| `response.DuplicateError(c, "重复原因")` | 数据重复 | `response.DuplicateError(c, "昵称已被使用")` |

### 3.4 使用示例

```go
// 1. 成功返回数据
response.Success(c, user)

// 2. 成功返回带消息
response.SuccessWithMsg(c, gin.H{"id": newID}, "注册成功")

// 3. 返回分页列表
response.SuccessPage(c, articles, total, page, pageSize)

// 4. 参数错误
response.ParamErrorWithMsg(c, "参数错误："+err.Error())

// 5. 未授权
response.UnauthorizedWithMsg(c, "用户不存在")

// 6. 禁止访问
response.ForbiddenWithMsg(c, "无权限修改他人的文章")

// 7. 资源不存在
response.NotFoundWithMsg(c, "文章不存在")

// 8. 数据库错误
response.DatabaseError(c)

// 9. 数据重复
response.DuplicateError(c, "邮箱已被注册")
```

---

## 四、core/init_db 数据库模块

### 4.1 获取数据库实例

```go
// 获取数据库实例（整个项目就用这一个）
db := core.GetDB()
```

### 4.2 GORM 常用操作

| 操作 | 方法 | 示例 |
|-----|------|------|
| 插入 | `db.Create(&obj)` | `db.Create(&user)` |
| 查询单个 | `db.First(&obj, id)` | `db.First(&user, 1)` |
| 查询单个（条件） | `db.Where().First()` | `db.Where("nickname = ?", name).First(&user)` |
| 查询列表 | `db.Find(&list)` | `db.Find(&users)` |
| 更新 | `db.Model(&obj).Updates()` | `db.Model(&user).Updates({"nickname": "新昵称"})` |
| 删除 | `db.Delete(&obj)` | `db.Delete(&user)` |
| 统计 | `db.Count()` | `db.Model(&User{}).Count(&count)` |
| 预加载 | `db.Preload("关联")` | `db.Preload("User").Find(&articles)` |

### 4.3 完整查询示例

```go
// ========== 插入 ==========
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
response.SuccessWithMsg(c, gin.H{"id": newUser.ID}, "创建成功")

// ========== 根据ID查询 ==========
var user model.User
if err := db.First(&user, userID).Error; err != nil {
    response.NotFoundWithMsg(c, "用户不存在")
    return
}
response.Success(c, user)

// ========== 条件查询 ==========
var user model.User
err := db.Where("nickname = ? OR email = ?", login, login).First(&user).Error
if err != nil {
    response.UnauthorizedWithMsg(c, "用户不存在")
    return
}

// ========== 分页查询 ==========
offset := (page - 1) * pageSize
var total int64
var articles []model.Article
query := db.Model(&model.Article{})
if keyword != "" {
    query = query.Where("title LIKE ?", "%"+keyword+"%")
}
query.Count(&total)
if err := query.Preload("User").
    Order("created_at DESC").
    Offset(offset).
    Limit(pageSize).
    Find(&articles).Error; err != nil {
    response.DatabaseError(c)
    return
}
response.SuccessPage(c, articles, total, page, pageSize)

// ========== 更新 ==========
updates := map[string]interface{}{
    "title":   req.Title,
    "content": req.Content,
}
if err := db.Model(&article).Updates(updates).Error; err != nil {
    response.DatabaseError(c)
    return
}
response.SuccessWithMsg(c, nil, "更新成功")

// ========== 删除 ==========
if err := db.Delete(&article).Error; err != nil {
    response.DatabaseError(c)
    return
}
response.SuccessWithMsg(c, nil, "删除成功")

// ========== 检查记录是否存在 ==========
var count int64
db.Model(&model.User{}).Where("nickname = ?", nickname).Count(&count)
if count > 0 {
    response.DuplicateError(c, "昵称已被使用")
    return
}
```

### 4.4 自动建表

```go
// models/register.go
package models

import "github.com/项目名/models/model"

func AllModels() []interface{} {
	return []interface{}{
		&model.User{},
		&model.Article{},
		&model.Comment{},
	}
}
```

```go
// main.go 中调用
core.AutoMigrate(models.AllModels()...)
```

---

## 五、core/init_config 配置模块

### 5.1 配置文件结构

```yaml
# config/app.yaml（基础配置）
server:
  ip: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "123456"
  name: "my_blog"

jwt:
  secret: "your-secret-key-change-in-production"
  expire: 24

email:
  host: "smtp.qq.com"
  port: 587
  username: "example@qq.com"
  password: "your授权码"
  from: "example@qq.com"

log:
  filename: "logs/app.log"
  maxSize: 100
  maxBackups: 30
  maxAge: 7
  compress: true
  level: "debug"
```

```yaml
# config/app-dev.yaml（开发环境覆盖）
database:
  password: "dev_password"

log:
  level: "debug"
```

```yaml
# config/app-prod.yaml（生产环境覆盖）
database:
  password: "prod_password"

log:
  level: "warn"
```

### 5.2 配置读取

```go
// 加载配置（自动合并 app.yaml + app-{mode}.yaml）
core.ReadConfig(flags.Opt.Mode)

// 使用配置
core.GlobalConfig.Server.Port      // 服务器端口
core.GlobalConfig.Database.Host    // 数据库地址
core.GlobalConfig.JWT.Secret       // JWT密钥
core.GlobalConfig.JWT.Expire       // Token过期时间（小时）
core.GlobalConfig.Log.Level        // 日志级别
```

---

## 六、models/model 数据模型

### 6.1 基础模型

```go
package model

import "gorm.io/gorm"

type User struct {
	gorm.Model                         // 内含 ID、CreatedAt、UpdatedAt、DeletedAt
	Nickname string `json:"nickname" gorm:"type:varchar(32);uniqueIndex;not null"`
	Avatar   string `json:"avatar"`
	Abstract string `json:"abstract"`
	Email    string `json:"email" gorm:"type:varchar(128);uniqueIndex;not null"`
	Password string `json:"-" gorm:"type:varchar(64);not null"` // json:"-" 序列化时忽略
	Status   int    `json:"status" gorm:"type:tinyint(1);default:1"`
}

type Article struct {
	gorm.Model
	Title   string `json:"title" gorm:"type:varchar(128);not null"`
	Content string `json:"content" gorm:"type:text;not null"`
	UserID  uint   `json:"user_id" gorm:"not null"`
	User    User   `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

type Comment struct {
	gorm.Model
	Content   string  `json:"content" gorm:"type:text;not null"`
	ArticleID uint    `json:"article_id" gorm:"not null"`
	UserID    uint    `json:"user_id" gorm:"not null"`
	User      User    `json:"-" gorm:"foreignKey:UserID;references:ID"`
	Article   Article `json:"-" gorm:"foreignKey:ArticleID;references:ID"`
}
```

### 6.2 GORM 标签说明

| 标签 | 说明 | 示例 |
|-----|------|------|
| `type:varchar(32)` | 数据库字段类型 | `type:varchar(32)` |
| `type:text` | 文本类型 | `type:text` |
| `type:int` | 整数类型 | `type:int` |
| `type:tinyint(1)` | 微小整数 | `type:tinyint(1)` |
| `not null` | 非空约束 | `not null` |
| `uniqueIndex` | 唯一索引 | `uniqueIndex` |
| `default:1` | 默认值 | `default:1` |
| `comment:状态说明` | 字段注释 | `comment:状态 0:禁用 1:正常` |
| `json:"-"` | JSON序列化忽略 | `json:"-"` |
| `gorm:"foreignKey:UserID"` | 外键关联 | `gorm:"foreignKey:UserID;references:ID"` |

---

## 七、api/middleware JWT 认证中间件

### 7.1 中间件模板

```go
package middleware

import (
	"net/http"
	"strings"
	"github.com/项目名/core"
	"github.com/项目名/core/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims 结构体：JWT 负载
type Claims struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	jwt.RegisteredClaims
}

// JWTAuth 中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code: response.CodeUnauthorized,
				Msg:  "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 2. 验证格式 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code: response.CodeUnauthorized,
				Msg:  "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		// 3. 解析 Token
		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(core.GlobalConfig.JWT.Secret), nil
		})

		// 4. 验证 Token
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code: response.CodeUnauthorized,
				Msg:  "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 5. 存入 Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Nickname)

		c.Next()
	}
}
```

### 7.2 生成 Token

```go
// 登录时生成 Token
expireTime := time.Now().Add(time.Duration(core.GlobalConfig.JWT.Expire) * time.Hour)
claims := jwt.MapClaims{
	"user_id":  user.ID,
	"nickname": user.Nickname,
	"exp":      expireTime.Unix(),
	"iat":      time.Now().Unix(),
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte(core.GlobalConfig.JWT.Secret))
```

---

## 八、api/router 路由配置

### 8.1 主路由模板

```go
package router

import (
	"github.com/项目名/api/middleware"
	"github.com/项目名/api/router/nestedrouter"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 公开路由组（无需登录）
	apiGroup := r.Group("/api")
	{
		nestedrouter.UserRoutes(apiGroup)
	}

	// 受保护路由组（需要登录）
	protectedGroup := r.Group("/api")
	protectedGroup.Use(middleware.JWTAuth())
	{
		nestedrouter.ArticleRoutes(protectedGroup)
		nestedrouter.CommentRoutes(protectedGroup)
	}

	return r
}
```

### 8.2 子路由模板

```go
package nestedrouter

import (
	"github.com/项目名/api/controller"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", controller.Register)
		userGroup.POST("/login", controller.Login)
		userGroup.GET("/info", controller.GetUserInfo)
	}
}

func ArticleRoutes(r *gin.RouterGroup) {
	articleGroup := r.Group("/article")
	{
		articleGroup.POST("/create", controller.CreateArticle)
		articleGroup.GET("/list", controller.GetArticleList)
		articleGroup.GET("/detail/:id", controller.GetArticleDetail)
		articleGroup.PUT("/update/:id", controller.UpdateArticle)
		articleGroup.DELETE("/delete/:id", controller.DeleteArticle)
	}
}
```

---

## 九、api/controller 控制器模板

### 9.1 请求参数结构体

```go
// 请求参数（JSON 请求体）
type UserReq struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=32"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"required,email"`
}

// 查询参数（URL query）
type ListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
}
```

### 9.2 控制器模板

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/项目名/core"
	"github.com/项目名/core/response"
	"github.com/项目名/models/model"
)

// Create 创建资源
func Create(c *gin.Context) {
	// 1. 解析参数
	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	// 2. 获取数据库
	db := core.GetDB()

	// 3. 获取当前用户（如果需要登录）
	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedWithMsg(c, "用户未登录")
		return
	}

	// 4. 业务逻辑
	newObj := model.User{
		Nickname: req.Nickname,
		UserID:   userID.(uint),
	}

	// 5. 数据库操作
	if err := db.Create(&newObj).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// 6. 返回响应
	response.SuccessWithMsg(c, gin.H{"id": newObj.ID}, "创建成功")
}

// GetList 获取列表（分页）
func GetList(c *gin.Context) {
	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 50 {
		req.PageSize = 10
	}

	db := core.GetDB()
	offset := (req.Page - 1) * req.PageSize

	var total int64
	var list []model.User

	query := db.Model(&model.User{})
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Keyword+"%")
	}

	query.Count(&total)
	if err := query.Offset(offset).Limit(req.PageSize).Find(&list).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	response.SuccessPage(c, list, total, req.Page, req.PageSize)
}

// GetDetail 获取详情
func GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamErrorWithMsg(c, "无效的ID")
		return
	}

	db := core.GetDB()
	var obj model.User
	if err := db.First(&obj, id).Error; err != nil {
		response.NotFoundWithMsg(c, "资源不存在")
		return
	}

	response.Success(c, obj)
}

// Update 更新
func Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamErrorWithMsg(c, "无效的ID")
		return
	}

	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	db := core.GetDB()
	var obj model.User
	if err := db.First(&obj, id).Error; err != nil {
		response.NotFoundWithMsg(c, "资源不存在")
		return
	}

	// 权限检查
	userID := c.GetUint("user_id")
	if obj.UserID != userID {
		response.ForbiddenWithMsg(c, "无权限")
		return
	}

	updates := map[string]interface{}{
		"nickname": req.Nickname,
	}
	if err := db.Model(&obj).Updates(updates).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	response.SuccessWithMsg(c, nil, "更新成功")
}

// Delete 删除
func Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamErrorWithMsg(c, "无效的ID")
		return
	}

	db := core.GetDB()
	var obj model.User
	if err := db.First(&obj, id).Error; err != nil {
		response.NotFoundWithMsg(c, "资源不存在")
		return
	}

	userID := c.GetUint("user_id")
	if obj.UserID != userID {
		response.ForbiddenWithMsg(c, "无权限")
		return
	}

	if err := db.Delete(&obj).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	response.SuccessWithMsg(c, nil, "删除成功")
}
```

---

## 十、flags 命令行参数

### 10.1 常用 API

| API | 说明 | 示例 |
|-----|------|------|
| `flags.Opt.Port` | 获取端口 | `flags.Opt.Port` |
| `flags.Opt.Mode` | 获取模式 | `dev/test/release` |
| `flags.Opt.ConfigFile` | 配置文件路径 | |
| `flags.IsDev()` | 是否开发模式 | `if flags.IsDev()` |
| `flags.IsRelease()` | 是否生产模式 | `if flags.IsRelease()` |
| `flags.GetAddr()` | 获取地址 | `"0.0.0.0:8080"` |

### 10.2 启动命令示例

```bash
# 开发环境
./myapp -mode dev -port 8080

# 测试环境
./myapp -mode test -port 8081

# 生产环境
./myapp -mode release -port 80 -host 0.0.0.0

# 初始化数据库
./myapp -db

# 查看版本
./myapp -v
```

---

## 十一、bcrypt 密码加密

```go
import "golang.org/x/crypto/bcrypt"

// 注册时加密
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 登录时验证
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputPassword))
// err == nil 表示密码正确
```

---

## 十二、快速复制清单

### 新建项目时需要创建的文件

- [ ] `main.go` - 入口文件
- [ ] `config/app.yaml` - 基础配置
- [ ] `config/app-dev.yaml` - 开发配置
- [ ] `core/response/response.go` - 响应封装
- [ ] `core/init_config.go` - 配置加载
- [ ] `core/init_db.go` - 数据库初始化
- [ ] `core/logger/` - 日志模块
- [ ] `flags/enter.go` - 命令行参数
- [ ] `models/register.go` - 模型注册
- [ ] `models/model/*.go` - 数据模型
- [ ] `api/middleware/jwt_auth.go` - JWT中间件
- [ ] `api/controller/*.go` - 控制器
- [ ] `api/router/router.go` - 路由配置

### 复制后需要修改的地方

1. `github.com/项目名/` → 改成实际的项目路径
2. `models/model/*.go` → 定义自己的表结构
3. `api/controller/*.go` → 写自己的业务逻辑
4. `config/app.yaml` → 改成自己的配置
5. `config/app-*.yaml` → 环境特定配置

---

## 十三、常见错误排查

| 错误 | 原因 | 解决 |
|-----|------|------|
| `Unauthorized` | Token 无效或过期 | 重新登录获取新 Token |
| `Forbidden` | 无权限操作 | 检查是否是资源所有者 |
| `Not Found` | 资源不存在 | 检查 ID 是否正确 |
| `Database Error` | 数据库操作失败 | 检查数据库连接和 SQL |
| `Param Error` | 参数校验失败 | 检查请求参数格式 |
| `配置加载失败` | 配置文件不存在或格式错误 | 检查 config/ 目录 |

---

> 📌 本手册由 AI 生成，基于 My_Blog 项目提取
> 📌 遇到问题请对照 learning.md 查看详细注释
