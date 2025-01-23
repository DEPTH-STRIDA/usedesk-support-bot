package model

import (
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	UserID   int64 `gorm:"unique;not null"`
	TicketID int64
	// TicketsID string `gorm:"type:jsonb"`
}
