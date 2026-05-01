package nestedrouter

import (
	"github.com/GoWeb/My_Blog/api/controller"
	"github.com/gin-gonic/gin"
)

// ArticleRoutes 函数：注册文章模块的路由
// 参数 r：是已拼接好前缀的路由组，且已经经过了 JWT 中间件保护
// 所有注册到这个函数的路由都需要携带有效 Token
func ArticleRoutes(r *gin.RouterGroup) {
	// r.Group("/article") 拼接最终前缀 /api/article
	articleGroup := r.Group("/article")
	{
		// POST /api/article/create - 创建文章
		// 功能：当前登录用户创建一篇新文章
		// 请求体：{"title": "标题", "content": "内容"}
		articleGroup.POST("/create", controller.CreateArticle)

		// GET /api/article/list - 获取文章列表
		// 功能：分页查询文章列表，支持关键词搜索
		// 查询参数：?page=1&page_size=10&keyword=关键词
		articleGroup.GET("/list", controller.GetArticleList)

		// GET /api/article/detail/:id - 获取文章详情
		// 功能：根据文章ID查询单篇文章
		// URL参数：:id 是文章ID，如 /api/article/detail/123
		articleGroup.GET("/detail/:id", controller.GetArticleDetail)

		// PUT /api/article/update/:id - 更新文章
		// 功能：更新指定文章的内容（仅作者可操作）
		// URL参数：:id 是文章ID
		// 请求体：{"title": "新标题", "content": "新内容"}
		articleGroup.PUT("/update/:id", controller.UpdateArticle)

		// DELETE /api/article/delete/:id - 删除文章
		// 功能：删除指定文章（仅作者可操作）
		// URL参数：:id 是文章ID
		articleGroup.DELETE("/delete/:id", controller.DeleteArticle)
	}
}
