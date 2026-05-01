package model

import "gorm.io/gorm"

// Article 结构体：文章数据模型
// 对应数据库中的 articles 表
type Article struct {
	// gorm.Model 内嵌字段，自动包含：
	// ID        uint      // 主键，自增
	// CreatedAt time.Time // 创建时间
	// UpdatedAt time.Time // 更新时间
	// DeletedAt gorm.DeletedAt // 软删除时间
	gorm.Model

	Title string `json:"title" gorm:"type:varchar(128);not null"` // 标题：varchar(128)，非空

	Content string `json:"content" gorm:"type:text;not null"` // 内容：text 类型，存储富文本或 Markdown 内容

	UserID uint `json:"user_id" gorm:"type:int;not null"` // 作者ID：关联 users 表的外键

	// User 字段：关联的用户对象
	// json:"-" 表示 JSON 序列化时忽略这个字段（防止递归嵌套）
	// gorm:"foreignKey:UserID;references:ID" 表示外键关系：
	//   - foreignKey:UserID 表示本表的 UserID 字段是外键
	//   - references:ID 表示引用 users 表的 ID 字段
	User User `json:"-" gorm:"foreignKey:UserID;references:ID"`
}
