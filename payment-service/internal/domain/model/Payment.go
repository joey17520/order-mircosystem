package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentModel struct {
	gorm.Model
	PaymentID  uuid.UUID `gorm:"type:varchar(128);not null;comment:支付ID"`
	OrderID    uuid.UUID `gorm:"type:varchar(128);not null;comment:订单ID"`
	UserID     uuid.UUID `gorm:"type:varchar(128);not null;comment:用户ID"`
	TotalPrice int64     `gorm:"type:bigint;not null;comment:支付总金额"`
}
