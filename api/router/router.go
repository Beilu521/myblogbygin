package router

import (
	"github.com/GoWeb/My_Blog/api/middleware"
	"github.com/GoWeb/My_Blog/api/router/nestedrouter"
	"github.com/gin-gonic/gin"
)

// InitRouter 函数：初始化 Gin 路由
// 返回配置好的 gin.Engine 实例
// 这是整个应用的路由总入口
func InitRouter() *gin.Engine {
	// ========== 第1步：创建默认的 Gin 引擎 ==========
	// gin.Default() 创建一个默认的 Gin 引擎
	// 默认集成 Logger（日志中间件）和 Recovery（崩溃恢复中间件）
	r := gin.Default()

	// ========== 第2步：创建公开路由组 ==========
	// r.Group("/api") 创建一个路由组，前缀是 /api
	// 这个组的路由不需要登录就可以访问
	apiGroup := r.Group("/api")
	{
		// 注册用户相关的路由
		// api/user/register - 用户注册（公开）
		// api/user/login - 用户登录（公开）
		// api/user/info - 获取用户信息（需要登录，但单独在下面保护）
		nestedrouter.UserRoutes(apiGroup)
	}

	// ========== 第3步：创建受保护的路由组 ==========
	// r.Group("/api") 再创建一个 /api 前缀的路由组
	// 使用 middleware.JWTAuth() 中间件来保护这组路由
	// 所有注册到这个组的路由都需要携带有效的 JWT Token
	protectedGroup := r.Group("/api")
	protectedGroup.Use(middleware.JWTAuth()) // 添加 JWT 认证中间件
	{
		// 注册文章相关的路由（需要登录）
		// api/article/create - 创建文章
		// api/article/list - 文章列表
		// api/article/detail/:id - 文章详情
		// api/article/update/:id - 更新文章
		// api/article/delete/:id - 删除文章
		nestedrouter.ArticleRoutes(protectedGroup)

		// 注册评论相关的路由（需要登录）
		// api/comment/create - 创建评论
		// api/comment/list - 评论列表
		// api/comment/delete/:id - 删除评论
		nestedrouter.CommentRoutes(protectedGroup)
	}

	// ========== 第4步：返回配置好的路由引擎 ==========
	return r
}
