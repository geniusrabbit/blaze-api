package jwthandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/api-template-base/internal/jwt"
	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	accountrepo "github.com/geniusrabbit/api-template-base/internal/repository/account/repository"
	userrepo "github.com/geniusrabbit/api-template-base/internal/repository/user/repository"
	"github.com/geniusrabbit/api-template-base/model"
)

var (
	errInvalidAccountTarget = errors.New(`invalid account target`)
)

type authRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	AccountID uint64 `json:"account"`
}

// AuthHandler endpoint
func AuthHandler(provider *jwt.Provider) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		accountID := uint64(0)
		username, password, ok := req.BasicAuth()
		if !ok {
			if req.Method != http.MethodPost {
				badRequest(wr, req)
				return
			}

			var auth authRequest
			if err := json.NewDecoder(req.Body).Decode(&auth); err != nil {
				errorResponse(wr, req, err)
				return
			}

			username = auth.Username
			password = auth.Password
			accountID = auth.AccountID
		}

		ctx := req.Context()
		usersRepo := userrepo.New()
		accountRepo := accountrepo.New()

		user, err := usersRepo.GetByPassword(ctx, username, password)
		if err != nil {
			errorResponse(wr, req, err)
			return
		}

		account, err := accountForUser(ctx, accountRepo, user.ID, accountID)
		if err != nil {
			errorResponse(wr, req, err)
			return
		}
		if account != nil {
			accountID = account.ID
		}

		token, err := provider.CreateToken(user.ID, accountID)
		if err != nil {
			errorResponse(wr, req, err)
			return
		}

		wr.WriteHeader(http.StatusOK)
		err = json.NewEncoder(wr).Encode(map[string]any{`token`: token})
		if err != nil {
			zap.L().Error(`encode response`, zap.Error(err))
		}
	}
}

func accountForUser(ctx context.Context, accountRepo account.Repository, userID, accountID uint64) (*model.Account, error) {
	accounts, err := accountRepo.FetchList(ctx, &account.Filter{UserID: []uint64{userID}})
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		if accountID != 0 {
			return nil, errInvalidAccountTarget
		}
		return nil, nil
	}
	return accounts[0], nil
}

func badRequest(wr http.ResponseWriter, req *http.Request) {
	wr.WriteHeader(http.StatusBadRequest)
	wr.Write([]byte(`{"error":"bad request"}`))
}

func errorResponse(wr http.ResponseWriter, req *http.Request, err error) {
	if err == sql.ErrNoRows || err == gorm.ErrRecordNotFound || err == userrepo.ErrInvalidPassword {
		wr.WriteHeader(http.StatusBadRequest)
	} else {
		wr.WriteHeader(http.StatusInternalServerError)
	}
	if err == userrepo.ErrInvalidPassword {
		json.NewEncoder(wr).Encode(map[string]any{`error`: `invalid login or password`})
	} else {
		json.NewEncoder(wr).Encode(map[string]any{`error`: err.Error()})
	}
}
