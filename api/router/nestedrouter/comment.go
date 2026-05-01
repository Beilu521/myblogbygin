package nestedrouter

import (
	"github.com/GoWeb/My_Blog/api/controller"
	"github.com/gin-gonic/gin"
)

// CommentRoutes 函数：注册评论模块的路由
// 参数 r：是已拼接好前缀的路由组，且已经经过了 JWT 中间件保护
// 所有注册到这个函数的路由都需要携带有效 Token
func CommentRoutes(r *gin.RouterGroup) {
	// r.Group("/comment") 拼接最终前缀 /api/comment
	commentGroup := r.Group("/comment")
	{
		// POST /api/comment/create - 创建评论
		// 功能：给指定文章添加评论
		// 请求体：{"article_id": 文章ID, "content": "评论内容"}
		commentGroup.POST("/create", controller.CreateComment)

		// GET /api/comment/list - 获取评论列表
		// 功能：分页查询某篇文章的评论
		// 查询参数：?article_id=文章ID&page=1&page_size=10
		commentGroup.GET("/list", controller.GetCommentList)

		// DELETE /api/comment/delete/:id - 删除评论
		// 功能：删除指定评论（仅评论作者可操作）
		// URL参数：:id 是评论ID
		commentGroup.DELETE("/delete/:id", controller.DeleteComment)
	}
}
