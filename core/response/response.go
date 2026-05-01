package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 结构体：统一响应格式
// 所有 API 接口都返回这个格式的数据
type Response struct {
	Code  int         `json:"code"`           // 业务状态码（非 HTTP 状态码）
	Msg   string      `json:"msg"`            // 提示信息
	Data  interface{} `json:"data,omitempty"`  // 数据（omitempty 表示空值时不序列化）
	Error string      `json:"error,omitempty"` // 错误详情（仅在出错时返回）
}

// PageResult 结构体：分页查询结果
// 用于列表接口返回分页数据
type PageResult struct {
	List       interface{} `json:"list"`        // 数据列表
	Total      int64       `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`  // 每页记录数
	TotalPages int         `json:"total_pages"` // 总页数
}

// ========== 业务状态码常量 ==========
// 状态码设计：
// - 0xxx：成功
// - 4xxx：客户端错误（参数错误、权限问题等）
// - 5xxx：服务端错误（数据库错误、系统错误等）

const (
	CodeSuccess         = 0      // 成功
	CodeParamError      = 400    // 参数错误（类似 HTTP 400）
	CodeUnauthorized    = 401    // 未授权/未登录
	CodeForbidden       = 403    // 禁止访问（已登录但无权限）
	CodeNotFound        = 404    // 资源不存在
	CodeInternalError   = 500    // 服务器内部错误
	CodeDatabaseError   = 50001  // 数据库操作失败（业务层面的数据库错误）
	CodeBusinessError   = 50002  // 业务逻辑错误
	CodeValidationError = 40001  // 数据校验失败（如格式错误、长度不符等）
	CodeDuplicateError  = 40002  // 数据重复（如昵称、邮箱已被使用）
)

// Success 函数：返回成功响应
// 参数：
//   - c: Gin 上下文
//   - data: 要返回的数据
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess, // 0
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMsg 函数：返回带自定义消息的成功响应
// 参数：
//   - c: Gin 上下文
//   - data: 要返回的数据
//   - msg: 自定义成功消息
func SuccessWithMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg, // 自定义消息，如"注册成功"、"文章创建成功"
		Data: data,
	})
}

// SuccessPage 函数：返回分页成功响应
// 用于列表查询接口，自动计算总页数
// 参数：
//   - c: Gin 上下文
//   - list: 当前页的数据列表
//   - total: 符合条件的总记录数
//   - page: 当前页码
//   - pageSize: 每页记录数
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	// ========== 计算总页数 ==========
	// 假设 total=95, pageSize=10
	// 95/10 = 9 余 5，所以需要 10 页
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	// 返回分页数据结构
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: PageResult{
			List:       list,       // 当前页的数据
			Total:      total,      // 总记录数
			Page:       page,       // 当前页码
			PageSize:   pageSize,   // 每页数量
			TotalPages: totalPages, // 计算出的总页数
		},
	})
}

// Error 函数：返回通用错误响应
func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeInternalError, // 500
		Msg:  msg,
	})
}

// ErrorWithCode 函数：返回指定错误码的错误响应
// 用于需要返回特定业务错误码的场景
func ErrorWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code, // 自定义错误码
		Msg:  msg,
	})
}

// ParamError 函数：返回参数错误响应（默认消息）
func ParamError(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: CodeParamError, // 400
		Msg:  "invalid parameter", // 默认消息
	})
}

// ParamErrorWithMsg 函数：返回参数错误响应（自定义消息）
// 用于告诉前端具体是哪个参数出了问题
func ParamErrorWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeParamError,
		Msg:  msg, // 自定义消息，如"参数错误：nickname长度不能少于2"
	})
}

// Unauthorized 函数：返回未授权响应（默认消息）
// 用于 JWT Token 无效或过期
func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: CodeUnauthorized, // 401
		Msg:  "unauthorized",   // 默认消息
	})
}

// UnauthorizedWithMsg 函数：返回未授权响应（自定义消息）
func UnauthorizedWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: CodeUnauthorized,
		Msg:  msg, // 自定义消息，如"用户不存在"、"密码错误"
	})
}

// Forbidden 函数：返回禁止访问响应（默认消息）
// 用于已登录但没有权限的场景
func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, Response{
		Code: CodeForbidden, // 403
		Msg:  "forbidden",
	})
}

// ForbiddenWithMsg 函数：返回禁止访问响应（自定义消息）
func ForbiddenWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code: CodeForbidden,
		Msg:  msg, // 自定义消息，如"无权限修改他人的文章"
	})
}

// NotFound 函数：返回资源不存在响应（默认消息）
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code: CodeNotFound, // 404
		Msg:  "resource not found",
	})
}

// NotFoundWithMsg 函数：返回资源不存在响应（自定义消息）
func NotFoundWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{
		Code: CodeNotFound,
		Msg:  msg, // 自定义消息，如"用户不存在"、"文章不存在"
	})
}

// InternalError 函数：返回服务器内部错误响应（默认消息）
// 用于系统级错误，如数据库连接失败、文件操作失败等
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code: CodeInternalError, // 500
		Msg:  "internal server error",
	})
}

// InternalErrorWithMsg 函数：返回服务器内部错误响应（自定义消息）
func InternalErrorWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code: CodeInternalError,
		Msg:  msg, // 自定义消息
	})
}

// DatabaseError 函数：返回数据库操作错误响应
// 用于数据库 CRUD 操作失败
func DatabaseError(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: CodeDatabaseError, // 50001（业务错误码）
		Msg:  "database operation failed", // 注意：HTTP 状态码仍是 200，因为这是业务层面的错误
	})
}

// BusinessError 函数：返回业务逻辑错误响应
// 用于业务规则不满足的场景
func BusinessError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeBusinessError, // 50002
		Msg:  msg, // 自定义业务错误消息
	})
}

// ValidationError 函数：返回数据校验错误响应
// 用于请求参数校验失败（如格式不对、长度不符等）
func ValidationError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeValidationError, // 40001
		Msg:  msg, // 具体是哪个字段校验失败
	})
}

// DuplicateError 函数：返回数据重复错误响应
// 用于唯一字段重复的场景，如昵称已被使用、邮箱已被注册
func DuplicateError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeDuplicateError, // 40002
		Msg:  msg, // 如"昵称已被使用"
	})
}
