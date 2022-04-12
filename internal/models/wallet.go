package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID        int64
	UUID      uuid.UUID
	UserID    int64
	Balance   decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
