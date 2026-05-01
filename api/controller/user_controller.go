package controller

import (
	"time"

	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/response"
	"github.com/GoWeb/My_Blog/models/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterReq 结构体：注册请求参数
// 前端提交 JSON 时，Gin 会自动绑定到这里
// binding 标签用于参数校验
type RegisterReq struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=32"` // 昵称：必填，长度2-32
	Password string `json:"password" binding:"required,min=6,max=20"` // 密码：必填，长度6-20
	Email    string `json:"email" binding:"required,email"`          // 邮箱：必填，格式必须是邮箱
}

// LoginReq 结构体：登录请求参数
// 支持用户名或邮箱登录
type LoginReq struct {
	Login    string `json:"login" binding:"required"`     // 登录名：可以是昵称或邮箱
	Password string `json:"password" binding:"required"` // 密码
}

// Register 函数：用户注册
// POST /api/user/register
// 功能：创建新用户账号
func Register(c *gin.Context) {
	// ========== 第1步：解析请求参数 ==========
	// c.ShouldBindJSON 自动解析 JSON 请求体到结构体
	// 如果格式不对或校验失败，返回错误
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error()) // 返回400和错误信息
		return
	}

	// ========== 第2步：获取数据库连接 ==========
	// core.GetDB() 获取全局数据库实例
	db := core.GetDB()

	// ========== 第3步：检查昵称是否已被使用 ==========
	// db.Model(&model.User{}) 指定要查询的表
	// Where("nickname = ?", req.Nickname) 条件查询
	// Count(&count) 统计符合条件的数量
	var count int64
	db.Model(&model.User{}).Where("nickname = ?", req.Nickname).Count(&count)
	if count > 0 {
		// 昵称重复，返回40002错误码
		response.DuplicateError(c, "昵称已被使用")
		return
	}

	// ========== 第4步：检查邮箱是否已被注册 ==========
	// 同样的方式检查邮箱唯一性
	db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		response.DuplicateError(c, "邮箱已被注册")
		return
	}

	// ========== 第5步：密码加密 ==========
	// bcrypt 是 Go 官方推荐的密码加密库
	// GenerateFromPassword 将明文密码转为加密哈希
	// bcrypt.DefaultCost = 10，表示加密强度
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c) // 加密失败，返回500
		return
	}

	// ========== 第6步：创建用户记录 ==========
	// 构造 User 结构体实例
	newUser := model.User{
		Nickname: req.Nickname,             // 昵称
		Password: string(hashedPassword),   // 加密后的密码
		Email:    req.Email,                // 邮箱
		Status:   1,                        // 状态：1=正常，0=禁用
	}

	// db.Create(&newUser) 插入数据库
	if err := db.Create(&newUser).Error; err != nil {
		response.DatabaseError(c) // 数据库错误，返回50001
		return
	}

	// ========== 第7步：返回成功响应 ==========
	// gin.H{"id": newUser.ID} 相当于 map[string]interface{}
	// 返回新创建用户的 ID
	response.SuccessWithMsg(c, gin.H{"id": newUser.ID}, "注册成功")
}

// Login 函数：用户登录
// POST /api/user/login
// 功能：验证用户身份，返回 JWT Token
func Login(c *gin.Context) {
	// ========== 第1步：解析登录参数 ==========
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	db := core.GetDB()

	// ========== 第2步：查询用户 ==========
	// 支持两种登录方式：昵称登录 或 邮箱登录
	// .Where("nickname = ? OR email = ?", req.Login, req.Login)
	// 查询条件是：nickname = 输入值 或 email = 输入值
	var user model.User
	err := db.Where("nickname = ? OR email = ?", req.Login, req.Login).First(&user).Error
	if err != nil {
		// db.First() 找不到记录会返回 error
		response.UnauthorizedWithMsg(c, "用户不存在")
		return
	}

	// ========== 第3步：检查账号状态 ==========
	// Status = 0 表示账号被禁用
	if user.Status == 0 {
		response.ForbiddenWithMsg(c, "账号已被禁用")
		return
	}

	// ========== 第4步：验证密码 ==========
	// bcrypt.CompareHashAndPassword 比较哈希值和明文密码
	// 如果匹配返回 nil，不匹配返回 error
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.UnauthorizedWithMsg(c, "密码错误")
		return
	}

	// ========== 第5步：生成 JWT Token ==========
	// Token 包含用户身份信息，用于后续接口的认证

	// 设置过期时间：从配置文件读取，core.GlobalConfig.JWT.Expire 的单位是小时
	expireTime := time.Now().Add(time.Duration(core.GlobalConfig.JWT.Expire) * time.Hour)

	// 创建 Claims（JWT 的 payload，负载）
	// MapClaims 是 map 类型的 Claims，使用键值对存储
	claims := jwt.MapClaims{
		"user_id":  user.ID,       // 用户ID
		"nickname": user.Nickname, // 用户昵称
		"exp":      expireTime.Unix(),  // 过期时间（时间戳）
		"iat":      time.Now().Unix(),   // 签发时间（时间戳）
	}

	// 创建 Token 对象，使用 HS256 签名算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用 JWT Secret 对 Token 进行签名，生成最终的 tokenString
	// 签名后的 Token 才能在网络上安全传输
	tokenString, err := token.SignedString([]byte(core.GlobalConfig.JWT.Secret))
	if err != nil {
		response.InternalError(c) // 签名失败
		return
	}

	// ========== 第6步：返回登录成功信息 ==========
	// 返回 Token 和用户基本信息
	response.Success(c, gin.H{
		"token":    tokenString,    // JWT Token，前端需要保存
		"user_id":  user.ID,       // 用户ID
		"nickname": user.Nickname, // 昵称
		"email":    user.Email,    // 邮箱
		"avatar":   user.Avatar,   // 头像URL
	})
}

// GetUserInfo 函数：获取当前用户信息
// GET /api/user/info
// 功能：获取登录用户的详细信息（需要携带 Token）
// 注意：这个接口需要 JWT 中间件保护，所以能从 Context 获取 user_id
func GetUserInfo(c *gin.Context) {
	// ========== 第1步：从 Context 获取当前用户ID ==========
	// JWT 中间件已经将用户ID存入了 Context
	// c.Get() 返回 (value, exists)
	userID, exists := c.Get("user_id")
	if !exists {
		// 如果没有找到，说明没有登录（JWT 中间件应该已经拦截了）
		response.UnauthorizedWithMsg(c, "用户未登录")
		return
	}

	db := core.GetDB()

	// ========== 第2步：查询用户信息 ==========
	// db.First(&user, userID) 根据主键查询
	// 相当于 SELECT * FROM users WHERE id = userID
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		// 用户不存在
		response.NotFoundWithMsg(c, "用户不存在")
		return
	}

	// ========== 第3步：返回用户信息 ==========
	response.Success(c, gin.H{
		"id":         user.ID,              // 用户ID
		"nickname":   user.Nickname,       // 昵称
		"email":      user.Email,          // 邮箱
		"avatar":     user.Avatar,        // 头像
		"abstract":   user.Abstract,       // 个人简介
		"status":     user.Status,         // 状态
		"created_at": user.CreatedAt.Format(time.DateTime), // 创建时间，格式化为易读格式
	})
}
