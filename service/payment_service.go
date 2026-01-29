package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-app-server/utils"
	configUtils "github.com/gadhittana01/cases-modules/utils"
	dbUtils "github.com/gadhittana01/cases-modules/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pusher "github.com/pusher/pusher-http-go/v5"
	"github.com/shopspring/decimal"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentlink"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
)

type PaymentService struct {
	repo                repository.Repository
	stripeKey           string
	stripeSecret        string
	stripeWebhookSecret string
	frontendURL         string
	pusherClient        *pusher.Client
}

func NewPaymentService(repo repository.Repository, config *configUtils.Config, pusherClient *pusher.Client) *PaymentService {
	stripeKey := config.StripeKey
	stripeSecret := config.StripeSecret
	stripeWebhookSecret := config.StripeWebhookSecret

	// Set Stripe API key
	stripe.Key = stripeSecret

	// Get frontend URL from environment or use default
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	return &PaymentService{
		repo:                repo,
		stripeKey:           stripeKey,
		stripeSecret:        stripeSecret,
		stripeWebhookSecret: stripeWebhookSecret,
		frontendURL:         frontendURL,
		pusherClient:        pusherClient,
	}
}

func (s *PaymentService) getFrontendURL() string {
	return s.frontendURL
}

// AcceptQuote creates a Stripe Payment Link for the quote
// NOTE: This does NOT accept the quote or engage the case - that only happens via webhook
func (s *PaymentService) AcceptQuote(ctx context.Context, quoteID, clientID uuid.UUID) (*dto.PaymentIntentResponse, error) {
	// Get quote
	quote, err := s.repo.GetQuoteByID(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("quote not found: %w", err)
	}

	// Verify case belongs to client
	caseRecord, err := s.repo.GetCaseByID(ctx, quote.CaseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}

	if caseRecord.ClientID != clientID {
		return nil, fmt.Errorf("unauthorized: you can only accept quotes for your own cases")
	}

	if caseRecord.Status != "open" {
		return nil, fmt.Errorf("case is not open for acceptance")
	}

	// Check if quote is still proposed
	if quote.Status != "proposed" {
		return nil, fmt.Errorf("quote is not available for acceptance")
	}

	// Convert amount to cents
	amountDecimal := utils.PgtypeNumericToDecimal(quote.Amount)
	if amountDecimal == nil {
		return nil, fmt.Errorf("invalid quote amount")
	}
	amountCents := amountDecimal.Mul(decimal.NewFromInt(100)).IntPart()

	// Use transaction to ensure atomicity
	var paymentLinkURL string
	var paymentLinkID string

	err = dbUtils.ExecTxPool(ctx, s.repo.GetDB(), func(tx pgx.Tx) error {
		txRepo := s.repo.WithTx(tx)

		// Re-check quote status within transaction (prevent race conditions)
		quoteCheck, err := txRepo.GetQuoteByID(ctx, quoteID)
		if err != nil {
			return err
		}
		if quoteCheck.Status != "proposed" {
			return fmt.Errorf("quote was already processed")
		}

		// Create product first
		productParams := &stripe.ProductParams{
			Name: stripe.String(fmt.Sprintf("Legal Services - Case: %s", caseRecord.Title)),
			Metadata: map[string]string{
				"quote_id": quoteID.String(),
				"case_id":  quote.CaseID.String(),
			},
		}
		prod, err := product.New(productParams)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		// Create price for the product
		priceParams := &stripe.PriceParams{
			Currency:   stripe.String(string(stripe.CurrencySGD)),
			Product:    stripe.String(prod.ID),
			UnitAmount: stripe.Int64(amountCents),
		}
		priceObj, err := price.New(priceParams)
		if err != nil {
			return fmt.Errorf("failed to create price: %w", err)
		}

		// Create payment link with initial return URL (will update with payment link ID after creation)
		initialReturnURL := fmt.Sprintf("%s/client/cases/%s/payment/processing", s.getFrontendURL(), quote.CaseID.String())
		params := &stripe.PaymentLinkParams{
			LineItems: []*stripe.PaymentLinkLineItemParams{
				{
					Price:    stripe.String(priceObj.ID),
					Quantity: stripe.Int64(1),
				},
			},
			Metadata: map[string]string{
				"quote_id":        quoteID.String(),
				"case_id":         quote.CaseID.String(),
				"payment_link_id": "", // Will be updated after creation
			},
			AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
				Type: stripe.String("redirect"),
				Redirect: &stripe.PaymentLinkAfterCompletionRedirectParams{
					URL: stripe.String(initialReturnURL),
				},
			},
		}

		pl, err := paymentlink.New(params)
		if err != nil {
			return fmt.Errorf("failed to create payment link: %w", err)
		}

		paymentLinkID = pl.ID
		paymentLinkURL = pl.URL

		// Update payment link metadata with payment link ID and return URL
		// Redirect to loading/processing page - webhook will update status
		returnURL := fmt.Sprintf("%s/client/cases/%s/payment/processing?payment_link_id=%s", s.getFrontendURL(), quote.CaseID.String(), paymentLinkID)
		updateParams := &stripe.PaymentLinkParams{
			Metadata: map[string]string{
				"quote_id":        quoteID.String(),
				"case_id":         quote.CaseID.String(),
				"payment_link_id": paymentLinkID,
			},
			AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
				Type: stripe.String("redirect"),
				Redirect: &stripe.PaymentLinkAfterCompletionRedirectParams{
					URL: stripe.String(returnURL),
				},
			},
		}
		_, err = paymentlink.Update(paymentLinkID, updateParams)
		if err != nil {
			// Non-critical: payment link will still work with initial return URL
			log.Printf("Failed to update payment link return URL: %v", err)
		}

		// Create payment record (using payment link ID in stripe_payment_intent_id field for compatibility)
		// NOTE: Quote is NOT accepted, case is NOT engaged, lawyer gets NO access yet
		// This only happens after payment is confirmed via webhook
		_, err = txRepo.CreatePayment(ctx, &repository.CreatePaymentParams{
			QuoteID:               quoteID,
			StripePaymentIntentID: paymentLinkID, // Store payment link ID here
			Amount:                utils.DecimalToPgtypeNumeric(*amountDecimal),
			Status:                "pending",
		})
		if err != nil {
			return fmt.Errorf("failed to create payment record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.PaymentIntentResponse{
		ClientSecret:    paymentLinkURL, // For backward compatibility
		PaymentIntentID: paymentLinkID,  // Payment Link ID
		PaymentLinkURL:  paymentLinkURL, // Explicit payment link URL
	}, nil
}

// HandlePaymentWebhook processes checkout.session.completed webhook event
// This is the ONLY place where quote acceptance, case engagement, and access happen
func (s *PaymentService) HandlePaymentWebhook(ctx context.Context, checkoutSession *stripe.CheckoutSession) error {
	// Verify payment was successful
	if checkoutSession.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return fmt.Errorf("payment not completed, status: %s", checkoutSession.PaymentStatus)
	}

	// Get payment link ID from checkout session
	// Priority: PaymentLink field > metadata > quote_id lookup
	paymentLinkID := s.extractPaymentLinkID(ctx, checkoutSession)
	if paymentLinkID == "" {
		return fmt.Errorf("payment link ID not found in checkout session")
	}

	// Get payment record
	payment, err := s.repo.GetPaymentByStripePaymentIntentID(ctx, paymentLinkID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Get quote and case
	quote, err := s.repo.GetQuoteByID(ctx, payment.QuoteID)
	if err != nil {
		return fmt.Errorf("quote not found: %w", err)
	}

	caseRecord, err := s.repo.GetCaseByID(ctx, quote.CaseID)
	if err != nil {
		return fmt.Errorf("case not found: %w", err)
	}

	// Verify quote and case are still in valid state
	if quote.Status != "proposed" {
		return fmt.Errorf("quote already processed, status: %s", quote.Status)
	}
	if caseRecord.Status != "open" {
		return fmt.Errorf("case already processed, status: %s", caseRecord.Status)
	}

	// Atomic transaction: accept quote, reject others, engage case, mark payment succeeded
	err = dbUtils.ExecTxPool(ctx, s.repo.GetDB(), func(tx pgx.Tx) error {
		txRepo := s.repo.WithTx(tx)

		// Re-check quote status within transaction (prevent race conditions)
		quoteCheck, err := txRepo.GetQuoteByID(ctx, payment.QuoteID)
		if err != nil {
			return err
		}
		if quoteCheck.Status != "proposed" {
			return fmt.Errorf("quote was already processed")
		}

		// Accept quote
		if _, err := txRepo.AcceptQuote(ctx, payment.QuoteID); err != nil {
			return fmt.Errorf("failed to accept quote: %w", err)
		}

		// Reject other quotes
		if _, err := txRepo.RejectOtherQuotes(ctx, &repository.RejectOtherQuotesParams{
			CaseID: quote.CaseID,
			ID:     payment.QuoteID,
		}); err != nil {
			return fmt.Errorf("failed to reject other quotes: %w", err)
		}

		// Mark case as engaged
		if _, err := txRepo.UpdateCaseStatus(ctx, &repository.UpdateCaseStatusParams{
			ID:     quote.CaseID,
			Status: "engaged",
		}); err != nil {
			return fmt.Errorf("failed to update case status: %w", err)
		}

		// Mark payment as succeeded
		if _, err := txRepo.UpdatePaymentStatus(ctx, &repository.UpdatePaymentStatusParams{
			ID:     payment.ID,
			Status: "succeeded",
		}); err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	// Emit Pusher event after successful payment processing
	channel := fmt.Sprintf("payment-%s", paymentLinkID)
	eventData := map[string]interface{}{
		"payment_id":     payment.ID.String(),
		"payment_status": "succeeded",
		"quote_id":       quote.ID.String(),
		"quote_status":   "accepted",
		"case_id":        quote.CaseID.String(),
		"case_status":    "engaged",
		"is_completed":   true,
	}
	if err := s.pusherClient.Trigger(channel, "payment-completed", eventData); err != nil {
		log.Printf("Failed to emit Pusher event: %v", err)
		// Don't fail the transaction if Pusher fails
	}

	return nil
}

// extractPaymentLinkID extracts payment link ID from checkout session using multiple fallback strategies
func (s *PaymentService) extractPaymentLinkID(ctx context.Context, checkoutSession *stripe.CheckoutSession) string {
	// Strategy 1: Direct PaymentLink field (most reliable for Payment Links)
	if checkoutSession.PaymentLink != nil {
		return checkoutSession.PaymentLink.ID
	}

	// Strategy 2: Metadata
	if checkoutSession.Metadata != nil {
		if id := checkoutSession.Metadata["payment_link_id"]; id != "" {
			return id
		}

		// Strategy 3: Lookup by quote_id from metadata
		if quoteIDStr := checkoutSession.Metadata["quote_id"]; quoteIDStr != "" {
			if quoteID, err := uuid.Parse(quoteIDStr); err == nil {
				if payment, err := s.repo.GetPaymentByQuoteID(ctx, quoteID); err == nil && payment != nil {
					return payment.StripePaymentIntentID
				}
			}
		}
	}

	return ""
}
