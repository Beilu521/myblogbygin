package controller

import (
	"strconv"
	"time"

	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/response"
	"github.com/GoWeb/My_Blog/models/model"
	"github.com/gin-gonic/gin"
)

// ArticleReq 结构体：创建文章请求参数
// 前端 POST JSON 时自动绑定
type ArticleReq struct {
	Title   string `json:"title" binding:"required,min=1,max=128"`   // 标题：必填，1-128字符
	Content string `json:"content" binding:"required,min=1"`          // 内容：必填
}

// ArticleListReq 结构体：文章列表查询参数
// 前端通过 URL query 传递，如 /api/article/list?page=1&page_size=10&keyword=xxx
type ArticleListReq struct {
	Page     int    `form:"page"`      // 页码，默认1
	PageSize int    `form:"page_size"` // 每页数量，默认10
	Keyword  string `form:"keyword"`   // 搜索关键词
}

// ArticleUpdateReq 结构体：更新文章请求参数
type ArticleUpdateReq struct {
	Title   string `json:"title" binding:"required,min=1,max=128"`   // 标题：必填
	Content string `json:"content" binding:"required,min=1"`          // 内容：必填
}

// CreateArticle 函数：创建文章
// POST /api/article/create
// 功能：当前登录用户创建一篇新文章
// 注意：需要携带有效 Token（走 JWT 中间件）
func CreateArticle(c *gin.Context) {
	// ========== 第1步：解析请求参数 ==========
	// c.ShouldBindJSON 解析 JSON 请求体
	var req ArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	db := core.GetDB()

	// ========== 第2步：从 Context 获取当前用户ID ==========
	// JWT 中间件在验证 Token 成功后，会把 user_id 存入 Context
	// c.Get("user_id") 获取存入的值
	var userID uint
	if v, exists := c.Get("user_id"); exists {
		// 类型断言：从 interface{} 转成 uint
		userID = v.(uint)
	} else {
		// 理论上不会走到这里，因为 JWT 中间件会拦截未登录请求
		response.UnauthorizedWithMsg(c, "用户未登录")
		return
	}

	// ========== 第3步：创建文章记录 ==========
	// 构造 Article 结构体，UserID 设为当前登录用户的 ID
	newArticle := model.Article{
		Title:   req.Title,    // 文章标题
		Content: req.Content,  // 文章内容
		UserID:  userID,       // 作者ID（当前登录用户）
	}

	// db.Create() 插入数据库
	if err := db.Create(&newArticle).Error; err != nil {
		response.DatabaseError(c) // 插入失败
		return
	}

	// ========== 第4步：返回成功响应 ==========
	// 返回新文章的 ID
	response.SuccessWithMsg(c, gin.H{"id": newArticle.ID}, "文章创建成功")
}

// GetArticleList 函数：获取文章列表
// GET /api/article/list
// 功能：分页查询文章列表，支持关键词搜索
// 注意：需要携带 Token
func GetArticleList(c *gin.Context) {
	// ========== 第1步：解析查询参数 ==========
	// c.ShouldBindQuery 解析 URL query 参数（如 ?page=1&page_size=10）
	var req ArticleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	// ========== 第2步：设置分页默认值 ==========
	// 如果前端没传 page，默认第1页
	if req.Page < 1 {
		req.Page = 1
	}
	// 如果没传 page_size，或超出范围，默认10条/页
	if req.PageSize < 1 || req.PageSize > 50 {
		req.PageSize = 10
	}

	db := core.GetDB()

	// ========== 第3步：计算偏移量 ==========
	// offset = (页码 - 1) * 每页数量
	// 例如第1页：offset=0，第2页：offset=10
	offset := (req.Page - 1) * req.PageSize

	// 用于存储查询结果
	var total int64                       // 总记录数（用于计算总页数）
	var articles []model.Article          // 文章列表

	// ========== 第4步：构建查询 ==========
	// db.Model(&model.Article{}) 指定要查询的表
	query := db.Model(&model.Article{})

	// ========== 第5步：关键字搜索（可选） ==========
	// 如果有关键词，添加 WHERE 条件
	// LIKE '%关键词%' 可以匹配标题或内容中包含该词的文章
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%" // 前后加%表示任意字符
		// 搜索 title 或 content 字段
		query = query.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
	}

	// ========== 第6步：统计总数 ==========
	// COUNT(*) 用于分页计算总页数
	// 这个查询会在 LIMIT 之前执行，统计符合条件的总记录数
	query.Count(&total)

	// ========== 第7步：查询列表数据 ==========
	// Preload("User") 预加载关联的用户信息
	// 相当于 JOIN users 表，把文章作者的信息也查出来
	// Order("created_at DESC") 按创建时间倒序，最新文章在前
	// Offset(offset) 跳过前面的记录（分页）
	// Limit(req.PageSize) 只取指定数量的记录
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&articles).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// ========== 第8步：返回分页结果 ==========
	// response.SuccessPage 会返回：list、total、page、page_size、total_pages
	response.SuccessPage(c, articles, total, req.Page, req.PageSize)
}

// GetArticleDetail 函数：获取文章详情
// GET /api/article/detail/:id
// 功能：根据文章ID查询单篇文章的详细信息
// :id 是 URL 参数，通过 c.Param("id") 获取
func GetArticleDetail(c *gin.Context) {
	// ========== 第1步：获取 URL 参数中的文章ID ==========
	// c.Param("id") 获取路由参数，如 /api/article/detail/123 中的 123
	idStr := c.Param("id")

	// ========== 第2步：转换 ID 类型 ==========
	// URL 参数都是字符串，需要转成数字类型
	// strconv.ParseUint 将字符串转成 uint64
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		// 转换失败，说明 id 格式不对（比如传了 "abc"）
		response.ParamErrorWithMsg(c, "无效的文章ID")
		return
	}

	db := core.GetDB()
	var article model.Article

	// ========== 第3步：查询文章详情 ==========
	// db.First(&article, id) 根据主键查询
	// Preload("User") 预加载作者信息
	if err := db.Preload("User").First(&article, id).Error; err != nil {
		// 找不到文章
		response.NotFoundWithMsg(c, "文章不存在")
		return
	}

	// ========== 第4步：返回文章详情 ==========
	response.Success(c, article)
}

// UpdateArticle 函数：更新文章
// PUT /api/article/update/:id
// 功能：更新指定文章的内容（仅作者可操作）
// :id 是要更新的文章ID
func UpdateArticle(c *gin.Context) {
	// ========== 第1步：获取要更新的文章ID ==========
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamErrorWithMsg(c, "无效的文章ID")
		return
	}

	// ========== 第2步：解析更新内容 ==========
	var req ArticleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamErrorWithMsg(c, "参数错误："+err.Error())
		return
	}

	db := core.GetDB()
	var article model.Article

	// ========== 第3步：查询原文章 ==========
	// 先查出原文章内容
	if err := db.First(&article, id).Error; err != nil {
		response.NotFoundWithMsg(c, "文章不存在")
		return
	}

	// ========== 第4步：权限检查 ==========
	// 只有文章作者才能修改自己的文章
	// c.GetUint("user_id") 从 Context 获取当前登录用户的 ID
	userID := c.GetUint("user_id")
	if article.UserID != userID {
		// 不是作者，无权修改
		response.ForbiddenWithMsg(c, "无权限修改他人的文章")
		return
	}

	// ========== 第5步：执行更新 ==========
	// 使用 map 来更新多个字段
	updates := map[string]interface{}{
		"title":      req.Title,           // 新标题
		"content":    req.Content,         // 新内容
		"updated_at": time.Now(),           // 更新时间
	}

	// db.Model(&article) 指定要更新的表
	// Updates(updates) 更新多个字段
	if err := db.Model(&article).Updates(updates).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// ========== 第6步：返回成功响应 ==========
	response.SuccessWithMsg(c, nil, "文章更新成功")
}

// DeleteArticle 函数：删除文章
// DELETE /api/article/delete/:id
// 功能：删除指定文章（仅作者可操作）
func DeleteArticle(c *gin.Context) {
	// ========== 第1步：获取要删除的文章ID ==========
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamErrorWithMsg(c, "无效的文章ID")
		return
	}

	db := core.GetDB()
	var article model.Article

	// ========== 第2步：查询文章 ==========
	if err := db.First(&article, id).Error; err != nil {
		response.NotFoundWithMsg(c, "文章不存在")
		return
	}

	// ========== 第3步：权限检查 ==========
	// 只有作者才能删除自己的文章
	userID := c.GetUint("user_id")
	if article.UserID != userID {
		response.ForbiddenWithMsg(c, "无权限删除他人的文章")
		return
	}

	// ========== 第4步：执行删除 ==========
	// db.Delete(&article) 软删除（会设置 deleted_at 字段）
	// 如果模型没有软删除字段，则为永久删除
	if err := db.Delete(&article).Error; err != nil {
		response.DatabaseError(c)
		return
	}

	// ========== 第5步：返回成功响应 ==========
	response.SuccessWithMsg(c, nil, "文章删除成功")
}
