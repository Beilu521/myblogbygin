package model

import "gorm.io/gorm"

// User 结构体：用户数据模型
// 对应数据库中的 users 表
// GORM 会自动将结构体转换为表结构
type User struct {
	// gorm.Model 内嵌字段，自动包含：
	// ID        uint      // 主键，自增
	// CreatedAt time.Time // 创建时间
	// UpdatedAt time.Time // 更新时间
	// DeletedAt gorm.DeletedAt // 软删除时间（如果启用了软删除）
	gorm.Model

	Nickname string `json:"nickname" gorm:"type:varchar(32);uniqueIndex:not null"` // 昵称：varchar(32)，唯一且非空
	Avatar   string `json:"avatar"`  // 头像：存储头像图片的 URL
	Abstract string `json:"abstract"` // 个人简介：一段自我介绍文字

	Email string `json:"email" gorm:"type:varchar(128);uniqueIndex:not null"` // 邮箱：varchar(128)，唯一且非空

	Password string `json:"-" gorm:"type:varchar(64);not null"` // 密码：varchar(64)，加密存储
	// json:"-" 的作用：序列化 JSON 时忽略这个字段，防止密码泄露

	Status int `json:"status" gorm:"type:tinyint(1);default:1;comment:状态 0:禁用 1:正常"` // 状态：tinyint(1)，默认1正常，0禁用
}
