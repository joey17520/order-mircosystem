package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	ProductID   int64  `gorm:"type:bigint;not null;comment:产品ID;uniqueIndex:idx_product"`
	ProductName string `gorm:"type:varchar(60);not null;comment:产品名"`
	Price       int64  `gorm:"type:bigint;not null;comment:产品单价"`
	Quantity    int64  `gorm:"type:bigint;comment:产品数量"`
}
