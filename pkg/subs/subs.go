package subs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	SLATimeout = 5 * time.Second

	TimeParseFormat = "01-2006"
)

type Subscription struct {
	ID        *int64     `gorm:"primaryKey;autoIncrement;type:bigint"`
	Service   string     `gorm:"type:varchar(255);uniqueIndex:index_subs"`
	Cost      int32      `gorm:"type:int;not null"`
	UserID    uuid.UUID  `gorm:"type:uuid;uniqueIndex:index_subs"`
	StartDate time.Time  `gorm:"type:datetime;uniqueIndex:index_subs"`
	EndDate   *time.Time `gorm:"type:datetime"`
}

type SubscriptionFilter struct {
	Service   *string
	Cost      *int32
	UserID    *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time

	Limit  *int
	Offset *int
	Sort   *string
}

type SubscriptionsData struct {
	Subscriptions []*Subscription
	Total         int64
	SumCost       int64
}

type SubscriptionsRepo interface {
	Create(subscription *Subscription) (int64, error)
	ReadByParams(filter *SubscriptionFilter) (*Subscription, error)
	ReadByID(id int64) (*Subscription, error)
	Update(subscriptionUpdated *Subscription) error
	DeleteByID(id int64) error
	List(filter *SubscriptionFilter) (*SubscriptionsData, error)
	GetTotalCost(filter *SubscriptionFilter) (int64, error)
}

var (
	ErrAlreadyExists = errors.New("subscription already exists")
	ErrWrongParams   = errors.New("wrong params")
	ErrNotFound      = errors.New("subscription not found")
)
