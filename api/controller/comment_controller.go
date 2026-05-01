package controller

import (
	"strconv"

	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/response"
	"github.com/GoWeb/My_Blog/models/model"
	"github.com/gin-gonic/gin"
)

// CommentReq 结构体：创建评论请求参数
// 前端 POST JSON 时自动绑定
type CommentReq struct {
	ArticleID uint   `json:"article_id" binding:"required"` // 文章ID：必填，评论属于哪篇文章
	Content  string `json:"content" binding:"required,min=1,max=500"` // 评论内容：必填，1-500字符
}

// CommentListReq 结构体：评论列表查询参数
// 前端通过 URL query 传递
type CommentListReq struct {
	ArticleID uint `form:"article_id" binding:"required"` // 文章ID：必填，查询某篇文章的评论
	Page     int  `form:"page"`       // 页码，默认1
	PageSize int  `form:"page_size"`  // 每页数量，默认10
}

// CreateComment 函数：创建评论
// POST /api/comment/create
// 功能：给指定文章添加评论
// 注意：需要携带有效 Token
func CreateComment(c *gin.Context) {
	// ========== 第1步：解析请求参数 ==========
	var req CommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	db := core.GetDB()

	// ========== 第2步：从 Context 获取当前用户ID ==========
	// JWT 中间件验证 Token 成功后，会把 user_id 存入 Context
	var userID uint
	if v, exists := c.Get("user_id"); exists {
		userID = v.(uint)
	} else {
		// 未登录，JWT 中间件应该已经拦截了
		response.UnauthorizedWithMsg(c, "用户未登录")
		return
	}

	// ========== 第3步：检查文章是否存在 ==========
	// 评论需要挂在一篇文章下，所以先检查这篇文章是否存在
	// db.First(&article, req.ArticleID) 根据主键查询
	var article model.Article
	if err := db.First(&article, req.ArticleID).Error; err != nil {
		// 找不到文章，不能评论不存在的文章
		response.NotFoundWithMsg(c, "文章不存在")
		return
	}

	// ========== 第4步：创建评论记录 ==========
	newComment := model.Comment{
		Content:   req.Content,       // 评论内容
		ArticleID: req.ArticleID,      // 所属文章ID
		UserID:    userID,            // 评论作者ID（当前登录用户）
	}

	if err := db.Create(&newComment).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// ========== 第5步：返回成功响应 ==========
	response.SuccessWithMsg(c, gin.H{"id": newComment.ID}, "评论创建成功")
}

// GetCommentList 函数：获取评论列表
// GET /api/comment/list
// 功能：分页查询某篇文章的评论列表
// 注意：需要携带 Token
func GetCommentList(c *gin.Context) {
	// ========== 第1步：解析查询参数 ==========
	var req CommentListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	// ========== 第2步：设置分页默认值 ==========
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 50 {
		req.PageSize = 10
	}

	db := core.GetDB()

	// ========== 第3步：计算偏移量 ==========
	offset := (req.Page - 1) * req.PageSize

	var total int64                       // 总评论数
	var comments []model.Comment          // 评论列表

	// ========== 第4步：构建查询 ==========
	// 只查询属于指定文章(article_id)的评论
	query := db.Model(&model.Comment{}).Where("article_id = ?", req.ArticleID)

	// ========== 第5步：统计总数 ==========
	query.Count(&total)

	// ========== 第6步：查询评论列表 ==========
	// Preload("User") 预加载评论者信息
	// Order("created_at DESC") 最新评论在前
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&comments).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// ========== 第7步：返回分页结果 ==========
	response.SuccessPage(c, comments, total, req.Page, req.PageSize)
}

// DeleteComment 函数：删除评论
// DELETE /api/comment/delete/:id
// 功能：删除指定评论（仅评论作者可操作）
func DeleteComment(c *gin.Context) {
	// ========== 第1步：获取要删除的评论ID ==========
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamErrorWithMsg(c, "无效的评论ID")
		return
	}

	db := core.GetDB()
	var comment model.Comment

	// ========== 第2步：查询评论 ==========
	if err := db.First(&comment, id).Error; err != nil {
		response.NotFoundWithMsg(c, "评论不存在")
		return
	}

	// ========== 第3步：权限检查 ==========
	// 只有评论作者才能删除自己的评论
	userID := c.GetUint("user_id")
	if comment.UserID != userID {
		response.ForbiddenWithMsg(c, "无权限删除他人的评论")
		return
	}

	// ========== 第4步：执行删除 ==========
	if err := db.Delete(&comment).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// ========== 第5步：返回成功响应 ==========
	response.SuccessWithMsg(c, nil, "评论删除成功")
}
