# Error Chain Inspection

Example showing error chain inspection with errors.As() to extract error details.

## Example

<example>
<input>Inspect error chains with errors.As() to extract error details</input>
<output>
```go
package user

import (
    "errors"
    "fmt"
)

func HandleError(err error) {
    if err == nil {
        return
    }
    
    switch {
    case errors.Is(err, ErrUserNotFound):
        fmt.Println("User not found")
        
    case errors.Is(err, ErrInvalidInput):
        var ve *ValidationError
        if errors.As(err, &ve) {
            fmt.Printf("Validation error: %s - %s\n", ve.Field, ve.Message)
        }
        
    case errors.Is(err, ErrUserExists):
        var uae *UserAlreadyExistsError
        if errors.As(err, &uae) {
            fmt.Printf("User exists: %s\n", uae.Email)
        }
        
    default:
        var paymentErr *payment.PaymentError
        if errors.As(err, &paymentErr) {
            fmt.Printf("Payment error [%s]: %s\n", paymentErr.Code, paymentErr.Message)
            if paymentErr.Cause != nil {
                fmt.Printf("  Caused by: %v\n", paymentErr.Cause)
            }
            return
        }
        
        fmt.Printf("Unexpected error: %v\n", err)
    }
}

func ExampleUsage() {
    err := CreateUser(User{Email: "existing@example.com"})
    HandleError(err)
    
    err = GetUser("nonexistent")
    HandleError(err)
    
    err = CreateUser(User{Name: "John", Email: ""})
    HandleError(err)
}
```
</output>
</example>
