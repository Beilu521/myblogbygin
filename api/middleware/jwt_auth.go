package middleware

import (
	"net/http"
	"strings"

	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims 结构体：定义 JWT Token 中携带的用户信息
// JWT 的 payload（负载）部分存储的信息
type Claims struct {
	UserID   uint   `json:"user_id"`   // 用户ID，前端可以从 Token 中解析出用户ID
	Nickname string `json:"nickname"`   // 用户昵称
	jwt.RegisteredClaims               // JWT 内置的标准声明，包含过期时间(exp)、签发时间(iat)等
}

// JWTAuth 函数：JWT 认证中间件
// 返回一个 gin.HandlerFunc，用于保护需要登录才能访问的接口
// 使用方式：protectedGroup.Use(middleware.JWTAuth())
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ========== 第1步：从请求头获取 Authorization 字段 ==========
		// Authorization 格式："Bearer eyJhbGciOiJIUzI1NiIs..."
		// 前端需要在请求头中携带这个字段
		authHeader := c.GetHeader("Authorization")

		// ========== 第2步：检查 Authorization 是否存在 ==========
		// 如果前端没有携带 Token，直接返回 401 未授权
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code: response.CodeUnauthorized, // 401
				Msg:  "未提供认证令牌",          // 提示信息
			})
			c.Abort() // 中止后续处理，不再执行控制器
			return
		}

		// ========== 第3步：验证 Token 格式 ==========
		// 正确的格式应该是 "Bearer <token>"，中间有空格分隔
		// strings.SplitN 将字符串分割成最多2部分，避免 token 中包含空格导致分割错误
		parts := strings.SplitN(authHeader, " ", 2)

		// 验证条件：1. 必须有2部分 2. 第一部分必须是 "Bearer"
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code: response.CodeUnauthorized,
				Msg:  "认证令牌格式错误", // 格式不对，比如写成 "Bearer123" 而不是 "Bearer 123"
			})
			c.Abort()
			return
		}

		// ========== 第4步：提取 Token 字符串 ==========
		// parts[1] 才是实际的 Token 内容
		tokenString := parts[1]

		// ========== 第5步：解析并验证 Token ==========
		// 创建空的 Claims 结构体用于存储解析出的数据
		claims := &Claims{}

		// jwt.ParseWithClaims 解析 Token 并验证签名
		// 第三个参数是密钥回调函数，返回用于验证签名的密钥
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// 使用配置文件中的 JWT 密钥来验证签名
			// 注意：这里必须是与签名时相同的密钥才能验证通过
			return []byte(core.GlobalConfig.JWT.Secret), nil
		})

		// ========== 第6步：检查 Token 是否有效 ==========
		// 解析失败（密钥不匹配、Token 过期、格式错误等）或 Token 无效
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code: response.CodeUnauthorized,
				Msg:  "无效的认证令牌", // Token 过期、被篡改或密钥错误
			})
			c.Abort()
			return
		}

		// ========== 第7步：将用户信息存入 Context ==========
		// 这样在后续的控制器中可以通过 c.Get("user_id") 获取当前登录用户的 ID
		// 这是 Gin 框架的标准用法，通过 Context 在中间件和控制器之间传递数据
		c.Set("user_id", claims.UserID)   // 存入用户ID，供后续控制器使用
		c.Set("username", claims.Nickname) // 存入用户名

		// ========== 第8步：调用 c.Next() 继续处理 ==========
		// c.Next() 表示放行，让请求继续传递给下一个中间件或控制器
		// 如果不调用 c.Next()，请求就会在这里被中断
		c.Next()
	}
}
