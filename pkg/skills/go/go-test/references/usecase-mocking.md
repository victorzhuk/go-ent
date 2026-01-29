# UseCase Testing with Mocks

## Testing UseCase with Mocked Repository

```go
func TestCreateUserUC_Execute(t *testing.T) {
    type mocks struct {
        userRepo *mock_userRepo.MockUserRepo
    }

    type args struct {
        req CreateUserReq
    }

    tests := []struct {
        name    string
        setup   func(m *mocks)
        args    args
        want    *CreateUserResp
        wantErr error
    }{
        {
            name: "success",
            setup: func(m *mocks) {
                m.userRepo.EXPECT().
                    Save(gomock.Any(), gomock.Any()).
                    Return(nil)
            },
            args: args{
                req: CreateUserReq{
                    Email: "test@example.com",
                    Name:  "Test User",
                },
            },
            want: &CreateUserResp{
                ID: uuid.Must(uuid.NewV7()),
            },
            wantErr: nil,
        },
        {
            name: "duplicate email",
            setup: func(m *mocks) {
                m.userRepo.EXPECT().
                    Save(gomock.Any(), gomock.Any()).
                    Return(contract.ErrConflict)
            },
            args: args{
                req: CreateUserReq{
                    Email: "existing@example.com",
                    Name:  "Test User",
                },
            },
            want:    nil,
            wantErr: contract.ErrConflict,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            m := &mocks{
                userRepo: mock_userRepo.NewMockUserRepo(ctrl),
            }

            if tt.setup != nil {
                tt.setup(m)
            }

            uc := NewCreateUserUC(m.userRepo, slog.New(slog.DiscardHandler))
            got, err := uc.Execute(context.Background(), tt.args.req)

            if tt.wantErr != nil {
                require.Error(t, err)
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, got)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want.ID, got.ID)
        })
    }
}
```

**Pattern**: gomock for type-safe mocks, setup function per test, gomock.Any() for context param matching.
