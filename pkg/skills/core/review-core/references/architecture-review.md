# Architecture-Focused Code Review

Example showing architectural review for a new payment service implementation.

## Example

<example>
<input>Architecture review for new payment service</input>
<output>
## Review Feedback: PR #78 - Payment Service Implementation

### Design Concerns (P1 - Request Changes)

**⚠️ 1. Tight coupling to Stripe API**

```go
type PaymentService struct {
    stripeClient *stripe.Client
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    charge, err := s.stripeClient.Charges.New(&stripe.ChargeParams{
        Amount:   stripe.Int64(req.Amount),
        Currency: stripe.String("usd"),
        Source:   &stripe.SourceParams{Token: stripe.String(req.Token)},
    })
}
```

**Issue**: Business logic tightly coupled to specific payment provider. Difficult to test, migrate, or support multiple providers.

**Fix**: Introduce payment gateway abstraction:
```go
type PaymentGateway interface {
    Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error)
    Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
    GetStatus(ctx context.Context, transactionID string) (TransactionStatus, error)
}

type PaymentService struct {
    gateway PaymentGateway
}
```

---

**⚠️ 2. Missing transaction boundaries**

```go
func (s *PaymentService) ProcessOrder(ctx context.Context, order *Order) error {
    if err := s.repo.CreatePayment(ctx, payment); err != nil {
        return err
    }
    charge, err := s.gateway.Charge(ctx, req)
    if err != nil {
        return err
    }
    if err := s.repo.UpdateOrder(ctx, order); err != nil {
        return err
    }
    return nil
}
```

**Issue**: No transaction boundaries. Partial failures lead to inconsistent state.

**Fix**: Implement compensating transactions:
```go
func (s *PaymentService) ProcessOrder(ctx context.Context, order *Order) error {
    payment, err := s.repo.CreatePendingPayment(ctx, order)
    if err != nil {
        return fmt.Errorf("create payment: %w", err)
    }

    charge, err := s.gateway.Charge(ctx, toChargeReq(payment))
    if err != nil {
        s.repo.MarkPaymentFailed(ctx, payment.ID, err)
        return fmt.Errorf("charge payment: %w", err)
    }

    if err := s.repo.MarkOrderPaid(ctx, order.ID, charge.ID); err != nil {
        s.gateway.Refund(ctx, toRefundReq(charge))
        s.repo.MarkPaymentRefunded(ctx, payment.ID)
        return fmt.Errorf("update order: %w", err)
    }

    s.repo.MarkPaymentConfirmed(ctx, payment.ID, charge.ID)
    return nil
}
```

---

### Code Quality Issues (P2 - Suggestions)

**💡 3. Define domain errors**

```go
var (
    ErrPaymentFailed     = errors.New("payment failed")
    ErrInsufficientFunds = errors.New("insufficient funds")
    ErrCardDeclined      = errors.New("card declined")
)

func mapGatewayError(err error) error {
    if stripeErr, ok := err.(*stripe.Error); ok {
        switch stripeErr.Code {
        case stripe.ErrorCodeCardDeclined:
            return fmt.Errorf("%w: %v", ErrCardDeclined, stripeErr.Msg)
        case stripe.ErrorCodeInsufficientFunds:
            return fmt.Errorf("%w: %v", ErrInsufficientFunds, stripeErr.Msg)
        }
    }
    return err
}
```

---

**💡 4. Add idempotency keys**

```go
type ChargeRequest struct {
    Amount         int64
    IdempotencyKey string
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error) {
    if existing, err := s.repo.FindByIdempotencyKey(ctx, req.IdempotencyKey); err == nil {
        return existing, nil
    }
}
```

---

**💡 5. Add circuit breaker**

```go
type PaymentService struct {
    gateway PaymentGateway
    breaker *gobreaker.CircuitBreaker
}

func (s *PaymentService) Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error) {
    result, err := s.breaker.Execute(func() (interface{}, error) {
        return s.gateway.Charge(ctx, req)
    })
    if err != nil {
        return nil, err
    }
    return result.(*ChargeResponse), nil
}
```

---

### Summary

**Action**: Request changes (architectural concerns)

**Blocking**: Payment gateway abstraction, transaction boundaries

**Suggestions**: Domain errors, idempotency, circuit breaker
</output>
</example>
