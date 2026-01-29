package handler

import (
	"net/http"

	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) AcceptQuote(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	clientID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req dto.AcceptQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quoteID, err := uuid.Parse(req.QuoteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quote ID"})
		return
	}

	response, err := h.paymentService.AcceptQuote(c.Request.Context(), quoteID, clientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
