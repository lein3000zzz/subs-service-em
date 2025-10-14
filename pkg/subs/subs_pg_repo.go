package subs

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubscriptionsPgRepo struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewSubscriptionsPgRepo(logger *zap.SugaredLogger, db *gorm.DB) *SubscriptionsPgRepo {
	return &SubscriptionsPgRepo{
		logger: logger,
		db:     db,
	}
}

func (repo *SubscriptionsPgRepo) Create(subscription *Subscription) (int64, error) {
	repo.logger.Debugw("create subscription", "subscription", subscription)

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	upsertRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(subscription)

	if upsertRes.Error != nil {
		repo.logger.Errorw("error upserting subscription", "error", upsertRes.Error, "subscription", subscription)
		return 0, upsertRes.Error
	}

	if upsertRes.RowsAffected != 1 {
		repo.logger.Warnw("failed upserting subscription", "error", ErrAlreadyExists, "subscription", subscription)
		return 0, ErrAlreadyExists
	}

	repo.logger.Infow("subscription created", "subscription", subscription)
	return *subscription.ID, nil
}

func (repo *SubscriptionsPgRepo) ReadByParams(filter *SubscriptionFilter) (*Subscription, error) {
	repo.logger.Debugw("read subscription by params", "filter", filter)

	// Вообще проверка происходит на хэндлере, но во избежание неправильного использования сделана доп. проверка здесь
	if filter.Service == nil || filter.StartDate == nil || filter.UserID == nil {
		repo.logger.Errorw("invalid filter", "filter", filter)
		return nil, ErrWrongParams
	}

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	var subscription Subscription
	res := repo.db.WithContext(ctx).Where("service = ? AND start_time = ? AND user_id = ?",
		*filter.Service, *filter.StartDate, *filter.UserID).First(&subscription)

	if res.Error != nil {
		repo.logger.Errorw("error finding subscription by params", "error", res.Error, "filter", filter)
		return nil, res.Error
	}

	repo.logger.Debugw("subscription found", "subscription", subscription)
	return &subscription, nil
}

func (repo *SubscriptionsPgRepo) ReadByID(id int64) (*Subscription, error) {
	repo.logger.Debugw("read subscription by id", "id", id)

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	var subscription Subscription
	res := repo.db.WithContext(ctx).Where("id = ?", id).First(&subscription)

	if res.Error != nil {
		repo.logger.Errorw("error finding subscription by id", "id", id, "error", res.Error)
		return nil, res.Error
	}

	repo.logger.Infow("subscription found", "subscription", subscription)
	return &subscription, nil
}

func (repo *SubscriptionsPgRepo) Update(subscriptionUpdated *Subscription) error {
	repo.logger.Debugw("update subscription", "subscription", subscriptionUpdated)

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	res := repo.db.WithContext(ctx).Model(&Subscription{}).Where("id = ?", subscriptionUpdated.ID).Updates(subscriptionUpdated)

	if res.Error != nil {
		repo.logger.Errorw("error updating subscription", "error", res.Error, "subscription", subscriptionUpdated)
		return res.Error
	}

	if res.RowsAffected == 0 {
		repo.logger.Warnw("failed subscription update", "subscription", subscriptionUpdated)
		return ErrNotFound
	}

	repo.logger.Infow("subscription updated", "subscription", subscriptionUpdated)
	return nil
}

func (repo *SubscriptionsPgRepo) DeleteByID(id int64) error {
	repo.logger.Debugw("delete subscription", "id", id)

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	res := repo.db.WithContext(ctx).Where("id = ?", id).Delete(&Subscription{})

	if res.Error != nil {
		repo.logger.Errorw("error deleting subscription", "id", id, "error", res.Error)
		return res.Error
	}

	if res.RowsAffected == 0 {
		repo.logger.Warnw("failed deleting subscription", "id", id)
		return ErrNotFound
	}

	repo.logger.Infow("subscription deleted", "id", id)
	return nil
}

func (repo *SubscriptionsPgRepo) List(filter *SubscriptionFilter) (*SubscriptionsData, error) {
	repo.logger.Debugw("list subscriptions", "filter", filter)

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	query := repo.db.WithContext(ctx).Model(&Subscription{})

	if query.Error != nil {
		repo.logger.Errorw("error listing subscriptions", "error", query.Error, "filter", filter)
		return nil, query.Error
	}

	query = repo.filterQuery(query, filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		repo.logger.Warnw("failed to count requests with filter", "err", err, "filter", filter)
		return nil, err
	}
	if total == 0 {
		repo.logger.Debugw("no subscriptions found with provided filter", "filter", filter)
		return &SubscriptionsData{
			Subscriptions: []*Subscription{},
			Total:         0,
			SumCost:       0,
		}, nil
	}

	var sumCost int64
	if err := query.Select("SUM(cost)").Scan(&sumCost).Error; err != nil {
		repo.logger.Warnw("failed to sum costs with filter", "err", err, "filter", filter)
		return nil, err
	}

	var sort string
	if filter.Sort != nil {
		sort = *filter.Sort
	}

	order := repo.getSubsListOrder(sort)

	query = query.Order(order)

	if filter.Limit != nil && filter.Offset != nil {
		query = repo.setLimitAndOffset(query, *filter.Limit, *filter.Offset)
	}

	var subscriptions []*Subscription
	if err := query.Find(&subscriptions).Error; err != nil {
		repo.logger.Warnw("failed to list subscriptions", "filter", filter, "error", err)
		return nil, err
	}

	repo.logger.Infow("subscriptions found with filter", "subscriptions", subscriptions, "filter", filter)
	return &SubscriptionsData{
		Subscriptions: subscriptions,
		Total:         total,
		SumCost:       sumCost,
	}, nil
}

func (repo *SubscriptionsPgRepo) filterQuery(query *gorm.DB, filter *SubscriptionFilter) *gorm.DB {
	repo.logger.Debugw("filter subscriptions", "filter", filter)

	if filter.Service != nil {
		query = query.Where("service = ?", *filter.Service)
	}

	if filter.StartDate != nil {
		query = query.Where("start_date <= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("end_date >= ? OR end_date IS NULL", *filter.EndDate)
	}

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.Cost != nil {
		query = query.Where("cost = ?", *filter.Cost)
	}

	return query
}

func (repo *SubscriptionsPgRepo) getSubsListOrder(sort string) string {
	repo.logger.Debugw("get subs list", "sort", sort)

	var order string
	switch sort {
	case "cost_asc":
		order = "cost ASC"
	case "cost_desc":
		order = "cost DESC"
	case "service_asc":
		order = "service ASC"
	case "service_desc":
		order = "service DESC"
	case "start_date":
		order = "start_date ASC"
	default:
		order = "start_date DESC"
	}

	return order
}

func (repo *SubscriptionsPgRepo) setLimitAndOffset(query *gorm.DB, limit, offset int) *gorm.DB {
	repo.logger.Debugw("set limit and offset", "limit", limit, "offset", offset)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	return query
}

func (repo *SubscriptionsPgRepo) GetTotalCost(filter *SubscriptionFilter) (int64, error) {
	repo.logger.Debugw("get total cost of subscriptions", "filter", filter)

	ctx, cancel := context.WithTimeout(context.Background(), SLATimeout)
	defer cancel()

	query := repo.db.WithContext(ctx).Model(&Subscription{})

	if query.Error != nil {
		repo.logger.Errorw("error getting total cost", "error", query.Error, "filter", filter)
		return 0, query.Error
	}

	query = repo.filterQuery(query, filter)

	var sumCost int64
	if err := query.Select("SUM(cost)").Scan(&sumCost).Error; err != nil {
		repo.logger.Warnw("failed to sum costs with filter", "err", err, "filter", filter)
		return 0, err
	}

	repo.logger.Infow("total cost calculated", "sumCost", sumCost, "filter", filter)
	return sumCost, nil
}
