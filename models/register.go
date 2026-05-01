package models

import "github.com/GoWeb/My_Blog/models/model"
func AllModels() []interface{}{
	return []interface{}{
		&model.User{},
		&model.Article{},
		&model.Comment{},
	}
} 