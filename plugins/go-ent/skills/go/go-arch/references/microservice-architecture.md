# Microservice Architecture Reference

Complete implementation of microservice architecture with bounded contexts and async communication.

## Service Structure

```
services/
├── user-service/           # Bounded context: User management
│   ├── internal/
│   │   ├── domain/          # User entities, events
│   │   ├── usecase/         # CreateUser, UpdateProfile
│   │   ├── repository/      # user_postgres/
│   │   └── transport/       # http handlers
│   ├── Dockerfile
│   └── go.mod               # Separate dependency management
├── order-service/           # Bounded context: Order processing
│   ├── internal/
│   │   ├── domain/          # Order, OrderItem, OrderStatus
│   │   ├── usecase/         # CreateOrder, ProcessPayment
│   │   ├── repository/      # order_postgres/
│   │   └── transport/       # http + grpc
│   ├── Dockerfile
│   └── go.mod
├── payment-service/         # Bounded context: Payments
│   ├── internal/
│   │   ├── domain/          # Payment, Transaction
│   │   ├── usecase/         # ProcessPayment, Refund
│   │   ├── repository/      # payment_postgres/
│   │   └── transport/       # grpc
│   ├── Dockerfile
│   └── go.mod
└── notification-service/    # Bounded context: Notifications
    ├── internal/
    │   ├── domain/          # Notification, Channel
    │   ├── usecase/         # SendEmail, SendSMS
    │   ├── repository/      # notification_redis/
    │   └── transport/       # event consumers
    ├── Dockerfile
    └── go.mod

api-gateway/                # Single entry point
├── internal/
│   ├── router/             # Route to services
│   ├── middleware/         # Auth, rate limiting
│   └── transport/          # http server
└── Dockerfile
```

## Service Communication (Async via Message Queue)

```go
// Order service emits events
type OrderCreatedEvent struct {
    OrderID     string
    UserID      string
    Items       []OrderItem
    TotalAmount decimal.Decimal
    Timestamp   time.Time
}

func (uc *createOrderUC) Execute(ctx context.Context, req CreateOrderReq) error {
    order := domain.NewOrder(req.UserID, req.Items)
    
    if err := uc.orderRepo.Save(ctx, order); err != nil {
        return fmt.Errorf("save order: %w", err)
    }
    
    event := OrderCreatedEvent{
        OrderID:     order.ID,
        UserID:      req.UserID,
        Items:       req.Items,
        TotalAmount: order.Total,
        Timestamp:   time.Now(),
    }
    
    if err := uc.publisher.Publish(ctx, "orders.created", event); err != nil {
        return fmt.Errorf("publish event: %w", err)
    }
    
    return nil
}

// Notification service consumes events
type NotificationConsumer struct {
    emailSender contract.EmailSender
    smsSender   contract.SMSSender
}

func (c *NotificationConsumer) Start(ctx context.Context) error {
    return c.mq.Subscribe(ctx, "orders.*", func(msg amqp.Delivery) error {
        switch msg.RoutingKey {
        case "orders.created":
            var e OrderCreatedEvent
            if err := json.Unmarshal(msg.Body, &e); err != nil {
                return err
            }
            return c.sendOrderConfirmation(ctx, e)
        case "orders.paid":
            var e OrderPaidEvent
            if err := json.Unmarshal(msg.Body, &e); err != nil {
                return err
            }
            return c.sendPaymentReceipt(ctx, e)
        }
        return nil
    })
}

func (c *NotificationConsumer) sendOrderConfirmation(ctx context.Context, e OrderCreatedEvent) error {
    email := Email{
        To:      fmt.Sprintf("user+%s@example.com", e.UserID),
        Subject: "Order Confirmation",
        Body:    fmt.Sprintf("Your order %s has been created", e.OrderID),
    }
    return c.emailSender.Send(ctx, email)
}
```

## API Gateway Routing

```go
func (g *gateway) setupRoutes(r *chi.Mux) {
    // Route to user service
    r.Route("/users", func(r chi.Router) {
        r.Post("/", g.proxyToUserService)
        r.Get("/{id}", g.proxyToUserService)
    })
    
    // Route to order service
    r.Route("/orders", func(r chi.Router) {
        r.Use(g.authMiddleware)
        r.Post("/", g.proxyToOrderService)
        r.Get("/", g.listUserOrders)
    })
    
    // Health check per service
    r.Get("/health/users", g.checkUserHealth)
    r.Get("/health/orders", g.checkOrderHealth)
}
```

## Benefits

- Each service owns its database (no shared data)
- Independent deployment and scaling
- Async communication prevents cascading failures
- API Gateway provides single entry point with auth/routing
- Event-driven integration with eventual consistency
- Team ownership per bounded context
