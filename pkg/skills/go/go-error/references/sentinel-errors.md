# Sentinel Errors Pattern

Example showing sentinel errors with errors.Is() comparison.

## Example

<example>
<input>Define sentinel errors and handle with errors.Is()</input>
<output>
```go
package user

import "errors"

var (
    ErrUserNotFound     = errors.New("user not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrUserExists       = errors.New("user already exists")
    ErrUnauthorized     = errors.New("unauthorized access")
    ErrInternalError    = errors.New("internal server error")
)

type Service struct {
    repo *Repository
}

func (s *Service) GetUser(id string) (*User, error) {
    u, err := s.repo.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return nil, fmt.Errorf("get user: %w", ErrUserNotFound)
        }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return u, nil
}

func (s *Service) DeleteUser(id string) error {
    if err := s.repo.Delete(id); err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return ErrUserNotFound
        }
        return fmt.Errorf("delete user: %w", err)
    }
    return nil
}

func (s *Service) HandleGetUser(id string) error {
    u, err := s.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return fmt.Errorf("user %s not found", id)
        }
        return err
    }
    
    fmt.Printf("Found user: %s\n", u.Name)
    return nil
}
```
</output>
</example>
