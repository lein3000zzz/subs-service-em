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

type basicRequest struct {
	ServiceName string    `json:"service_name"`
	Cost        int32     `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

// Для корректной генерации сваггера

type ErrorResponse struct {
	Error string `json:"error"`
}

type SubscriptionResponse struct {
	Message      string             `json:"message"`
	Subscription *subs.Subscription `json:"subscription"`
}

type BasicResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

type ListResponse struct {
	Message       string               `json:"message"`
	Subscriptions []*subs.Subscription `json:"subscriptions"`
	Meta          *Metadata            `json:"meta"`
}

type Metadata struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int64 `json:"pages"`
}

type CostResponse struct {
	Message string `json:"message"`
	SumCost int64  `json:"sum_cost"`
}

// CreateSub godoc
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body basicRequest true "Subscription payload"
// @Success 201 {object} BasicResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/create [post]
func (h *SubsHandler) CreateSub(c *gin.Context) {
	h.logger.Debugw("handling CreateSub()")

	newSub, err := h.buildSubscriptionFromContext(c)

	if err != nil {
		h.logger.Errorw("error creating new sub", "error", err)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var lastInsertedID string
	if lastInsertedID, err = h.subsRepo.Create(newSub); err != nil {
		h.logger.Errorw("Failed to create subscription", "error", err)

		if errors.Is(err, subs.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Subscription already exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "failed to create subscription",
			})
		}

		return
	}

	h.logger.Infow("Successfully created subscription", "id", lastInsertedID)

	c.JSON(http.StatusCreated, BasicResponse{
		Message: messageSuccess,
		ID:      lastInsertedID,
	})
}

func (h *SubsHandler) buildSubscriptionFromContext(c *gin.Context) (*subs.Subscription, error) {
	var request basicRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorw("Failed to bind JSON", "error", err)

		return nil, err
	}

	startDate, err := time.Parse(subs.TimeParseFormat, request.StartDate)
	if err != nil {
		h.logger.Errorw(ErrDateFormat.Error(), "error", err)

		return nil, ErrDateFormat
	}

	var endDate *time.Time
	if request.EndDate != nil {
		endDateVal, err := time.Parse(subs.TimeParseFormat, *request.EndDate)
		if err != nil {
			h.logger.Errorw("Invalid end date format", "error", err)

			return nil, ErrDateFormat
		}
		endDate = &endDateVal
	}

	return &subs.Subscription{
		Service:   request.ServiceName,
		Cost:      request.Cost,
		UserID:    request.UserID,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// GetSubByID godoc
// @Summary Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/get/{id} [get]
func (h *SubsHandler) GetSubByID(c *gin.Context) {

	h.logger.Debugw("handling GetSubByID()")

	id := c.Param("id")

	subscription, err := h.subsRepo.ReadByID(id)
	h.handleGetSubscriptionResponse(c, subscription, err)
}

// GetByParams godoc
// @Summary Get subscription by unique params
// @Tags subscriptions
// @Produce json
// @Param service query string true "Service name"
// @Param userID query string true "User UUID"
// @Param startDate query string true "Start date MM-YYYY"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/get/query [get]
func (h *SubsHandler) GetByParams(c *gin.Context) {
	h.logger.Debugw("handling GetByParams()")

	filter, err := h.constructFilterFromContextQuery(c)
	if err != nil {
		h.logger.Errorw("Failed to construct filter from context query", "error", err)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if filter.Service == nil || filter.UserID == nil || filter.StartDate == nil {
		h.logger.Errorw("Insufficient filter params for a unique instance", "error", ErrInvalidParam)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrInvalidParam.Error(),
		})
		return
	}

	subscription, err := h.subsRepo.ReadByParams(filter)
	h.handleGetSubscriptionResponse(c, subscription, err)
}

func (h *SubsHandler) handleGetSubscriptionResponse(c *gin.Context, subscription *subs.Subscription, err error) {
	if err != nil {
		h.logger.Errorw("Failed to read subscription", "error", err)
		if errors.Is(err, subs.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Subscription not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to read subscription",
			})
		}

		return
	}

	h.logger.Infow("Successfully read subscription", "id", subscription.ID)
	c.JSON(http.StatusOK, SubscriptionResponse{
		Message:      messageSuccess,
		Subscription: subscription,
	})
}

// UpdateSub godoc
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body basicRequest true "Subscription payload"
// @Success 200 {object} BasicResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/update/{id} [patch]
func (h *SubsHandler) UpdateSub(c *gin.Context) {
	h.logger.Debugw("handling UpdateSub()")

	id := c.Param("id")

	subUpdates, err := h.buildSubscriptionFromContext(c)

	if err != nil {
		h.logger.Errorw("error creating new sub", "error", err)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	err = h.subsRepo.Update(id, subUpdates)
	if err != nil {
		h.logger.Errorw("Failed to update subscription", "error", err)

		if errors.Is(err, subs.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Subscription not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to update subscription",
			})
		}

		return
	}

	h.logger.Infow("Successfully updated subscription", "id", subUpdates.ID)

	c.JSON(http.StatusOK, BasicResponse{
		Message: messageSuccess,
		ID:      id,
	})
}

// DeleteSub godoc
// @Summary Delete subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} BasicResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/delete/{id} [delete]
func (h *SubsHandler) DeleteSub(c *gin.Context) {
	h.logger.Debugw("handling DeleteSub()")

	id := c.Param("id")

	err := h.subsRepo.DeleteByID(id)
	if err != nil {
		h.logger.Errorw("Failed to delete subscription", "error", err)

		if errors.Is(err, subs.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Subscription not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to delete subscription",
			})
		}

		return
	}

	h.logger.Infow("Successfully deleted subscription", "id", id)
	c.JSON(http.StatusOK, BasicResponse{
		Message: messageSuccess,
		ID:      id,
	})
}

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Param sort query string false "Sort (cost_asc\|cost_desc\|service_asc\|service_desc\|start_date)"
// @Param service query string false "Service name"
// @Param userID query string false "User UUID"
// @Param startDate query string false "Start date MM-YYYY"
// @Param endDate query string false "End date MM-YYYY"
// @Param price query int false "Cost"
// @Success 200 {object} ListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/list [get]
func (h *SubsHandler) List(c *gin.Context) {
	h.logger.Debugw("handling List()")

	filter, err := h.constructFilterFromContextQuery(c)
	if err != nil {
		h.logger.Errorw("Failed to construct filter from context query", "error", err)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	page, limit := utils.GetPageAndLimitFromContext(c)
	filter.Limit = &limit
	offset := (page - 1) * limit
	filter.Offset = &offset

	subsData, err := h.subsRepo.List(filter)
	if err != nil {
		h.logger.Errorw("Failed to list subscriptions", "error", err)

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to list subscriptions",
		})
		return
	}

	pages := utils.CountPages(subsData.Total, int64(limit))

	meta := &Metadata{
		Total: subsData.Total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	}

	c.JSON(http.StatusOK, ListResponse{
		Message:       messageSuccess,
		Subscriptions: subsData.Subscriptions,
		Meta:          meta,
	})
}

// GetTotalCost godoc
// @Summary Get total subscription cost for period
// @Tags subscriptions
// @Produce json
// @Param startDate query string true "Start date MM-YYYY"
// @Param endDate query string true "End date MM-YYYY"
// @Param service query string false "Service name"
// @Param userID query string false "User UUID"
// @Success 200 {object} CostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/v1/total [get]
func (h *SubsHandler) GetTotalCost(c *gin.Context) {
	h.logger.Debugw("handling GetTotalCost()")

	filter, err := h.constructFilterFromContextQuery(c)
	if err != nil {
		h.logger.Errorw("Failed to construct filter from context query", "error", err)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if filter.StartDate == nil || filter.EndDate == nil {
		h.logger.Errorw(ErrDateFormat.Error(), "error", err)

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrDateFormat.Error(),
		})
		return
	}

	cost, err := h.subsRepo.GetTotalCost(filter)
	if err != nil {
		h.logger.Errorw("Failed to get total cost", "error", err)

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get total cost",
		})
		return
	}

	h.logger.Infow("Successfully got total cost", "cost", cost)
	c.JSON(http.StatusOK, CostResponse{
		Message: messageSuccess,
		SumCost: cost,
	})
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
