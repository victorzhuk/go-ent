# Cross-Layer Error Propagation

Example showing error chain across application layers with proper context.

## Example

<example>
<input>Build error chain across application layers with proper context</input>
<output>
```go
package user

import (
    "fmt"
    "net/http"
)

type Service struct {
    repo *Repository
}

func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
    if err := s.validate(req); err != nil {
        return nil, fmt.Errorf("validate request: %w", err)
    }
    
    u, err := s.repo.Create(req.User)
    if err != nil {
        if errors.Is(err, ErrUserExists) {
            return nil, fmt.Errorf("create user: %w", ErrUserExists)
        }
        return nil, fmt.Errorf("create user in repository: %w", err)
    }
    
    return u, nil
}

func (s *Service) validate(req CreateUserRequest) error {
    if req.Name == "" {
        return &ValidationError{Field: "name", Message: "name is required"}
    }
    if req.Email == "" {
        return &ValidationError{Field: "email", Message: "email is required"}
    }
    return nil
}

type Handler struct {
    service *Service
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) error {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return fmt.Errorf("decode request: %w", err)
    }
    
    u, err := h.service.CreateUser(req)
    if err != nil {
        var ve *ValidationError
        if errors.As(err, &ve) {
            return fmt.Errorf("invalid request: %w", ve)
        }
        if errors.Is(err, ErrUserExists) {
            return fmt.Errorf("user already exists: %w", err)
        }
        return fmt.Errorf("create user: %w", err)
    }
    
    return json.NewEncoder(w).Encode(u)
}

func main() {
    handler := &Handler{service: &Service{repo: &Repository{}}}
    
    err := handler.HandleCreateUser(w, r)
    if err != nil {
        var ve *ValidationError
        if errors.As(err, &ve) {
            http.Error(w, ve.Error(), http.StatusBadRequest)
            return
        }
        if errors.Is(err, ErrUserExists) {
            http.Error(w, "user already exists", http.StatusConflict)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        log.Printf("create user failed: %v", err)
    }
}
```
</output>
</example>
