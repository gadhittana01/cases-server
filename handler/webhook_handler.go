package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

type WebhookHandler struct {
	paymentService *service.PaymentService
	webhookSecret  string
}

func NewWebhookHandler(paymentService *service.PaymentService, config *utils.Config) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
		webhookSecret:  config.StripeWebhookSecret,
	}
}

func (h *WebhookHandler) HandleStripeWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty request body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing Stripe-Signature header"})
		return
	}

	if h.webhookSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook secret not configured"})
		return
	}

	event, err := webhook.ConstructEventWithOptions(body, signature, h.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook signature verification failed"})
		return
	}

	switch event.Type {
	case stripe.EventTypeChargeUpdated:
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse charge"})
			return
		}

		if err := h.paymentService.HandleChargeUpdated(c.Request.Context(), &charge); err != nil {
			log.Printf("Error processing charge updated webhook: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	default:
		c.JSON(http.StatusOK, gin.H{"status": "unhandled event type", "event_type": event.Type})
	}
}
