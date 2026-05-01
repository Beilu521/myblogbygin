package model

import "gorm.io/gorm"

// Comment 结构体：评论数据模型
// 对应数据库中的 comments 表
type Comment struct {
	// gorm.Model 内嵌字段，自动包含：
	// ID        uint      // 主键，自增
	// CreatedAt time.Time // 创建时间
	// UpdatedAt time.Time // 更新时间
	// DeletedAt gorm.DeletedAt // 软删除时间
	gorm.Model

	Content string `json:"content" gorm:"type:text;not null"` // 评论内容：text 类型

	ArticleID uint `json:"article_id" gorm:"not null"` // 所属文章ID：关联 articles 表的外键

	UserID uint `json:"user_id" gorm:"not null"` // 评论者ID：关联 users 表的外键

	// User 字段：关联的评论者用户对象
	// gorm:"foreignKey:UserID;references:ID" 表示外键关系
	User User `json:"-" gorm:"foreignKey:UserID;references:ID"`

	// Article 字段：关联的文章对象
	// gorm:"foreignKey:ArticleID;references:ID" 表示外键关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;references:ID"`
}
