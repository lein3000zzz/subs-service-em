package handlers

import (
	"errors"
	"net/http"
	"online-subs/pkg/subs"
	"online-subs/pkg/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubsHandler struct {
	subsRepo subs.SubscriptionsRepo
	logger   *zap.SugaredLogger
}

func NewSubsHandler(subsRepo subs.SubscriptionsRepo, logger *zap.SugaredLogger) *SubsHandler {
	return &SubsHandler{
		subsRepo: subsRepo,
		logger:   logger,
	}
}

const (
	messageSuccess = "success"
)

var (
	ErrDateFormat   = errors.New("invalid start date format, expected MM-YYYY")
	ErrInvalidParam = errors.New("invalid param")
)

type createRequest struct {
	ServiceName string    `json:"service_name"`
	Cost        int32     `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
}

func (h *SubsHandler) CreateSub() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling CreateSub()")

		var request createRequest
		responseJSON := gin.H{}

		if err := c.ShouldBindJSON(&request); err != nil {
			h.logger.Errorw("Failed to bind JSON", "error", err)
			responseJSON["error"] = "Invalid request"

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		startDate, err := time.Parse(subs.TimeParseFormat, request.StartDate)
		if err != nil {
			h.logger.Errorw(ErrDateFormat.Error(), "error", err)
			responseJSON["error"] = ErrDateFormat.Error()

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		newSub := subs.Subscription{
			Service:   request.ServiceName,
			Cost:      request.Cost,
			UserID:    request.UserID,
			StartDate: startDate,
		}

		var lastInsertedID int64
		if lastInsertedID, err = h.subsRepo.Create(&newSub); err != nil {
			h.logger.Errorw("Failed to create subscription", "error", err)

			if errors.Is(err, subs.ErrAlreadyExists) {
				responseJSON["error"] = "Subscription already exists"
				c.JSON(http.StatusBadRequest, responseJSON)
			} else {
				responseJSON["error"] = "Failed to create subscription"
				c.JSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["message"] = messageSuccess
		responseJSON["id"] = lastInsertedID
		h.logger.Infow("Successfully created subscription", "id", lastInsertedID)

		c.JSON(http.StatusCreated, responseJSON)
	}
}

func (h *SubsHandler) GetSubByID() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling GetSubByID()")
		responseJSON := gin.H{}

		idStr := c.Param("id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			h.logger.Errorw("Failed to parse id", "error", err)
			responseJSON["error"] = "Invalid id"

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		subscription, err := h.subsRepo.ReadByID(id)
		if err != nil {
			h.logger.Errorw("Failed to read subscription", "error", err)

			if errors.Is(err, subs.ErrNotFound) {
				responseJSON["error"] = "Subscription not found"
				c.JSON(http.StatusNotFound, responseJSON)
			} else {
				responseJSON["error"] = "Failed to read subscription"
				c.JSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["message"] = messageSuccess
		responseJSON["subscription"] = subscription
		h.logger.Infow("Successfully read subscription", "id", subscription.ID)

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *SubsHandler) GetByParams() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling GetByParams()")

		responseJSON := gin.H{}

		filter, err := h.constructFilterFromContextQuery(c)
		if err != nil {
			h.logger.Errorw("Failed to construct filter from context query", "error", err)
			responseJSON["error"] = err.Error()

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		if filter.Service == nil || filter.UserID == nil || filter.StartDate == nil {
			h.logger.Errorw("Insufficient filter params for a unique instance", "error", ErrInvalidParam)
			responseJSON["error"] = ErrInvalidParam

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		subscription, err := h.subsRepo.ReadByParams(filter)
		if err != nil {
			h.logger.Errorw("Failed to read subscription", "error", err)

			if errors.Is(err, subs.ErrNotFound) {
				responseJSON["error"] = "Subscription not found"
				c.JSON(http.StatusNotFound, responseJSON)
			} else {
				responseJSON["error"] = "Failed to read subscription"
				c.JSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["message"] = messageSuccess
		responseJSON["subscription"] = subscription
		h.logger.Infow("Successfully read subscription", "id", subscription.ID)

		c.JSON(http.StatusOK, responseJSON)
	}
}

type updateRequest struct {
	createRequest
	ID      int64   `json:"id"`
	EndDate *string `json:"end_date"`
}

func (h *SubsHandler) UpdateSub() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling UpdateSub()")

		var request updateRequest
		responseJSON := gin.H{}

		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			h.logger.Errorw("Failed to parse id", "error", err)
			responseJSON["error"] = ErrInvalidParam

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			h.logger.Errorw("Failed to bind JSON", "error", err)
			responseJSON["error"] = "Invalid request"

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		startDate, err := time.Parse(subs.TimeParseFormat, request.StartDate)
		if err != nil {
			h.logger.Errorw(ErrDateFormat.Error(), "error", err)
			responseJSON["error"] = ErrDateFormat.Error()

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		var endDate *time.Time
		if request.EndDate != nil {
			endDateVal, err := time.Parse(subs.TimeParseFormat, *request.EndDate)
			endDate = &endDateVal
			if err != nil {
				h.logger.Errorw("Invalid end date format", "error", err)
				responseJSON["error"] = "Invalid end date format, expected MM-YYYY"

				c.JSON(http.StatusBadRequest, responseJSON)
				return
			}
		}

		updatedSub := subs.Subscription{
			Service:   request.ServiceName,
			Cost:      request.Cost,
			UserID:    request.UserID,
			StartDate: startDate,
			EndDate:   endDate,
			ID:        &request.ID,
		}

		err = h.subsRepo.Update(id, &updatedSub)
		if err != nil {
			h.logger.Errorw("Failed to update subscription", "error", err)

			if errors.Is(err, subs.ErrNotFound) {
				responseJSON["error"] = "Subscription not found"
				c.JSON(http.StatusNotFound, responseJSON)
			} else {
				responseJSON["error"] = "Failed to update subscription"
				c.JSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["message"] = messageSuccess
		h.logger.Infow("Successfully updated subscription", "id", updatedSub.ID)

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *SubsHandler) DeleteSub() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling DeleteSub()")

		responseJSON := gin.H{}

		idStr := c.Param("id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			h.logger.Errorw("Failed to parse id", "error", err)
			responseJSON["error"] = "Invalid id"

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		err = h.subsRepo.DeleteByID(id)
		if err != nil {
			h.logger.Errorw("Failed to delete subscription", "error", err)

			if errors.Is(err, subs.ErrNotFound) {
				responseJSON["error"] = "Subscription not found"
				c.JSON(http.StatusNotFound, responseJSON)
			} else {
				responseJSON["error"] = "Failed to delete subscription"
				c.JSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["message"] = messageSuccess
		h.logger.Infow("Successfully deleted subscription", "id", id)

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *SubsHandler) List() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling List()")

		responseJSON := gin.H{}

		filter, err := h.constructFilterFromContextQuery(c)
		if err != nil {
			h.logger.Errorw("Failed to construct filter from context query", "error", err)
			responseJSON["error"] = err.Error()

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		page, limit := utils.GetPageAndLimitFromContext(c)
		filter.Limit = &limit
		offset := (page - 1) * limit
		filter.Offset = &offset

		subsData, err := h.subsRepo.List(filter)
		if err != nil {
			h.logger.Errorw("Failed to list subscriptions", "error", err)
			responseJSON["error"] = "Failed to list subscriptions"

			c.JSON(http.StatusInternalServerError, responseJSON)
			return
		}

		pages := utils.CountPages(subsData.Total, int64(limit))

		meta := gin.H{
			"total": subsData.Total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		}

		responseJSON["message"] = messageSuccess
		responseJSON["subscriptions"] = subsData.Subscriptions
		responseJSON["meta"] = meta

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *SubsHandler) GetTotalCost() func(c *gin.Context) {
	return func(c *gin.Context) {
		h.logger.Debugw("handling GetTotalCost()")
		responseJSON := gin.H{}

		filter, err := h.constructFilterFromContextQuery(c)
		if err != nil {
			h.logger.Errorw("Failed to construct filter from context query", "error", err)
			responseJSON["error"] = err.Error()

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		if filter.StartDate == nil || filter.EndDate == nil {
			h.logger.Errorw(ErrDateFormat.Error(), "error", err)
			responseJSON["error"] = ErrDateFormat.Error()

			c.JSON(http.StatusBadRequest, responseJSON)
			return
		}

		cost, err := h.subsRepo.GetTotalCost(filter)
		if err != nil {
			h.logger.Errorw("Failed to get total cost", "error", err)
			responseJSON["error"] = "Failed to get total cost"

			c.JSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["cost"] = cost
		h.logger.Infow("Successfully got total cost", "cost", cost)
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *SubsHandler) constructFilterFromContextQuery(c *gin.Context) (*subs.SubscriptionFilter, error) {
	h.logger.Debugw("constructFilterFromContextQuery()")

	var filter subs.SubscriptionFilter

	if sort := c.Query("sort"); sort != "" {
		filter.Sort = &sort
	}

	if startDateStr := c.Query("startDate"); startDateStr != "" {
		startDate, err := time.Parse(subs.TimeParseFormat, startDateStr)
		if err != nil {
			h.logger.Errorw(ErrDateFormat.Error(), "error", err)

			return nil, ErrDateFormat
		}

		filter.StartDate = &startDate
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		endDate, err := time.Parse(subs.TimeParseFormat, endDateStr)
		if err != nil {
			h.logger.Errorw(ErrDateFormat.Error(), "error", err)

			return nil, ErrDateFormat
		}

		filter.EndDate = &endDate
	}

	if service := c.Query("service"); service != "" {
		filter.Service = &service
	}

	if userIDStr := c.Query("userID"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Errorw("Failed to parse user ID", "error", err)

			return nil, err
		}

		filter.UserID = &userID
	}

	if costStr := c.Query("price"); costStr != "" {
		cost64, err := strconv.ParseInt(costStr, 10, 32)
		if err != nil {
			h.logger.Errorw("Failed to parse price", "error", err)

			return nil, ErrInvalidParam
		}
		cost := int32(cost64)
		filter.Cost = &cost
	}

	return &filter, nil
}
