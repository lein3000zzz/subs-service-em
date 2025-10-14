package subs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	SLATimeout = 5 * time.Second
)

type Subscription struct {
	ID        string     `gorm:"type:char(40);uniqueIndex;not null"`
	Service   string     `gorm:"type:varchar(255);primaryKey"`
	Cost      int32      `gorm:"type:int;not null"`
	UserID    uuid.UUID  `gorm:"type:uuid;primaryKey"`
	StartTime time.Time  `gorm:"type:datetime;primaryKey"`
	EndTime   *time.Time `gorm:"type:datetime"`
}

type SubscriptionFilter struct {
	Service   *string
	Cost      *int32
	UserID    *string
	StartTime *time.Time
	EndTime   *time.Time

	Limit  *int
	Offset *int
	Sort   *string
}

type SubscriptionsData struct {
	Subscriptions []*Subscription
	Total         int64
	SumCost       int64
}

// TODO время на хэндлере парсить

type SubscriptionsRepo interface {
	Create(Subscription) error
	ReadByParams() error
	Update(Subscription) error
	Delete(Subscription) error
	List() ([]Subscription, error)
}

var (
	ErrAlreadyExists = errors.New("subscription already exists")
	ErrWrongParams   = errors.New("wrong params")
	ErrNotFound      = errors.New("subscription not found")
)
