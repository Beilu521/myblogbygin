package nestedrouter

import (
	"github.com/GoWeb/My_Blog/api/controller"
	"github.com/gin-gonic/gin"
)

// UserRoutes 函数：注册用户模块的路由
// 参数 r：是已拼接好前缀的路由组（比如已经是 /api）
// 这个函数把 /api/user 相关的路由注册到传入的路由组上
func UserRoutes(r *gin.RouterGroup) {
	// r.Group("/user") 在 /api 的基础上再拼接 /user
	// 最终前缀是 /api/user
	userGroup := r.Group("/user")
	{
		// POST /api/user/register - 用户注册
		// 功能：创建新用户账号
		// 公开接口：不需要携带 Token
		userGroup.POST("/register", controller.Register)

		// POST /api/user/login - 用户登录
		// 功能：验证用户身份，返回 JWT Token
		// 公开接口：不需要携带 Token
		userGroup.POST("/login", controller.Login)

		// GET /api/user/info - 获取当前用户信息
		// 功能：获取登录用户的详细信息
		// 注意：这个路由在 router.go 中被放到了 protectedGroup
		// 所以实际上需要携带 Token 才能访问
		userGroup.GET("/info", controller.GetUserInfo)
	}
}
