package model

// User 用户表模型
type User struct {
	Id          int64  `gorm:"column:id;primaryKey" json:"id"`
	PhoneNumber string `gorm:"column:phoneNumber"   json:"phoneNumber"`
	Status      int    `gorm:"column:status"        json:"status"`
	ValidTime   string `gorm:"column:validTime"     json:"validTime"`
}

// TableName 指定表名，避免 GORM 自动复数化
func (User) TableName() string { return "users" }
