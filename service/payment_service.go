package service

import (
	"context"
	"fmt"
	"log"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-modules/utils"
	dbUtils "github.com/gadhittana01/cases-modules/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pusher "github.com/pusher/pusher-http-go/v5"
	"github.com/shopspring/decimal"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentlink"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
)

type PaymentService struct {
	repo         repository.Repository
	config       *utils.Config
	pusherClient *pusher.Client
}

func NewPaymentService(repo repository.Repository, config *utils.Config, pusherClient *pusher.Client) *PaymentService {
	stripe.Key = config.StripeSecret

	return &PaymentService{
		repo:         repo,
		config:       config,
		pusherClient: pusherClient,
	}
}

func (s *PaymentService) getFrontendURL() string {
	return s.config.FrontendURL
}

func (s *PaymentService) AcceptQuote(ctx context.Context, quoteID, clientID uuid.UUID) (*dto.PaymentIntentResponse, error) {
	quote, err := s.repo.GetQuoteByID(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("quote not found: %w", err)
	}

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

	if quote.Status != "proposed" {
		return nil, fmt.Errorf("quote is not available for acceptance")
	}

	amountDecimal := utils.PgtypeNumericToDecimal(quote.Amount)
	if amountDecimal == nil {
		return nil, fmt.Errorf("invalid quote amount")
	}
	amountCents := amountDecimal.Mul(decimal.NewFromInt(100)).IntPart()

	var paymentLinkURL string
	var paymentLinkID string

	err = dbUtils.ExecTxPool(ctx, s.repo.GetDB(), func(tx pgx.Tx) error {
		txRepo := s.repo.WithTx(tx)

		quoteCheck, err := txRepo.GetQuoteByID(ctx, quoteID)
		if err != nil {
			return err
		}
		if quoteCheck.Status != "proposed" {
			return fmt.Errorf("quote was already processed")
		}

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

		priceParams := &stripe.PriceParams{
			Currency:   stripe.String(string(stripe.CurrencySGD)),
			Product:    stripe.String(prod.ID),
			UnitAmount: stripe.Int64(amountCents),
		}
		priceObj, err := price.New(priceParams)
		if err != nil {
			return fmt.Errorf("failed to create price: %w", err)
		}

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
				"payment_link_id": "",
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
			log.Printf("Failed to update payment link return URL: %v", err)
		}

		_, err = txRepo.CreatePayment(ctx, &repository.CreatePaymentParams{
			QuoteID:               quoteID,
			StripePaymentIntentID: paymentLinkID,
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
		PaymentIntentID: paymentLinkID,
		PaymentLinkURL:  paymentLinkURL,
	}, nil
}

func (s *PaymentService) HandleChargeUpdated(ctx context.Context, charge *stripe.Charge) error {
	if charge.Status != stripe.ChargeStatusSucceeded || !charge.Paid {
		log.Printf("Charge %s not succeeded or not paid, status: %s, paid: %v", charge.ID, charge.Status, charge.Paid)
		return nil
	}

	if charge.PaymentIntent == nil {
		return fmt.Errorf("charge has no payment intent")
	}

	paymentIntentID := charge.PaymentIntent.ID
	if paymentIntentID == "" {
		return fmt.Errorf("payment intent ID is empty")
	}

	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	var checkoutSession *stripe.CheckoutSession
	if pi.Metadata != nil {
		if sessionID, ok := pi.Metadata["checkout_session_id"]; ok && sessionID != "" {
			session, err := session.Get(sessionID, nil)
			if err != nil {
				return fmt.Errorf("failed to retrieve checkout session: %w", err)
			}
			checkoutSession = session
		}
	}

	if checkoutSession == nil {
		return fmt.Errorf("checkout session not found for payment intent %s", paymentIntentID)
	}

	return s.HandlePaymentWebhook(ctx, checkoutSession)
}

func (s *PaymentService) HandlePaymentWebhook(ctx context.Context, checkoutSession *stripe.CheckoutSession) error {
	if checkoutSession.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return fmt.Errorf("payment not completed, status: %s", checkoutSession.PaymentStatus)
	}

	paymentLinkID := s.extractPaymentLinkID(ctx, checkoutSession)
	if paymentLinkID == "" {
		return fmt.Errorf("payment link ID not found in checkout session")
	}

	payment, err := s.repo.GetPaymentByStripePaymentIntentID(ctx, paymentLinkID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	quote, err := s.repo.GetQuoteByID(ctx, payment.QuoteID)
	if err != nil {
		return fmt.Errorf("quote not found: %w", err)
	}

	caseRecord, err := s.repo.GetCaseByID(ctx, quote.CaseID)
	if err != nil {
		return fmt.Errorf("case not found: %w", err)
	}

	if quote.Status != "proposed" {
		return fmt.Errorf("quote already processed, status: %s", quote.Status)
	}
	if caseRecord.Status != "open" {
		return fmt.Errorf("case already processed, status: %s", caseRecord.Status)
	}

	err = dbUtils.ExecTxPool(ctx, s.repo.GetDB(), func(tx pgx.Tx) error {
		txRepo := s.repo.WithTx(tx)

		quoteCheck, err := txRepo.GetQuoteByID(ctx, payment.QuoteID)
		if err != nil {
			return err
		}
		if quoteCheck.Status != "proposed" {
			return fmt.Errorf("quote was already processed")
		}

		if _, err := txRepo.AcceptQuote(ctx, payment.QuoteID); err != nil {
			return fmt.Errorf("failed to accept quote: %w", err)
		}

		if _, err := txRepo.RejectOtherQuotes(ctx, &repository.RejectOtherQuotesParams{
			CaseID: quote.CaseID,
			ID:     payment.QuoteID,
		}); err != nil {
			return fmt.Errorf("failed to reject other quotes: %w", err)
		}

		if _, err := txRepo.UpdateCaseStatus(ctx, &repository.UpdateCaseStatusParams{
			ID:     quote.CaseID,
			Status: "engaged",
		}); err != nil {
			return fmt.Errorf("failed to update case status: %w", err)
		}

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
	}

	log.Printf("Pusher event emitted: %v", eventData)

	return nil
}

func (s *PaymentService) extractPaymentLinkID(ctx context.Context, checkoutSession *stripe.CheckoutSession) string {
	if checkoutSession.PaymentLink != nil {
		return checkoutSession.PaymentLink.ID
	}

	if checkoutSession.Metadata != nil {
		if id := checkoutSession.Metadata["payment_link_id"]; id != "" {
			return id
		}

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
