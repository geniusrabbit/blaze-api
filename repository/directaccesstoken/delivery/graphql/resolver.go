package graphql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	datModels "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	datokenrepo "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/repository"
	datokenusecase "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolver handles GraphQL queries for direct access tokens.
type QueryResolver struct {
	uc directaccesstoken.Usecase
}

// NewQueryResolver creates a new QueryResolver instance.
func NewQueryResolver(uc directaccesstoken.Usecase) *QueryResolver {
	return &QueryResolver{uc: uc}
}

// NewDefaultQueryResolver returns a new QueryResolver with the default usecase.
func NewDefaultQueryResolver() *QueryResolver {
	return &QueryResolver{uc: datokenusecase.New(datokenrepo.NewDirectAccessTokenRepository())}
}

// Generate creates a new direct access token with the specified parameters.
func (r *QueryResolver) Generate(ctx context.Context, userID *uint64, description string, expiresAt *time.Time) (*models.DirectAccessTokenPayload, error) {
	// Initialize expiration time if nil
	if expiresAt == nil {
		expiresAt = &time.Time{}
	}

	// Set default expiration to 30 days if zero
	if expiresAt.IsZero() {
		*expiresAt = time.Now().Add(time.Hour * 24 * 30)
	}

	// Validate expiration is in the future
	if expiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("expiresAt should be in future")
	}

	// Generate token using the usecase
	token, err := r.uc.Generate(ctx,
		gocast.IfThenExec(userID != nil, func() uint64 { return *userID }, func() uint64 { return 0 }),
		session.AccountID(ctx),
		description,
		*expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &models.DirectAccessTokenPayload{
		ClientMutationID: requestid.Get(ctx),
		Token:            FromDirectAccessToken(token),
	}, nil
}

// Revoke revokes direct access tokens matching the provided filter.
func (r *QueryResolver) Revoke(ctx context.Context, filter models.DirectAccessTokenListFilter) (*models.StatusResponse, error) {
	err := r.uc.Revoke(ctx, FromFilterGraphQL(&filter))
	if err != nil {
		return nil, err
	}

	return &models.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           "ok",
		Message:          gocast.Ptr("token(s) revoked"),
	}, nil
}

// Get retrieves a direct access token by its ID.
func (r *QueryResolver) Get(ctx context.Context, id uint64) (*models.DirectAccessTokenPayload, error) {
	token, err := r.uc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.DirectAccessTokenPayload{
		ClientMutationID: requestid.Get(ctx),
		Token:            FromDirectAccessToken(token),
	}, nil
}

// List retrieves a paginated collection of direct access tokens with optional filtering and ordering.
// Tokens created within the last 5 minutes are returned as-is; older tokens have their values masked.
func (r *QueryResolver) List(ctx context.Context, filter *models.DirectAccessTokenListFilter, order []*models.DirectAccessTokenListOrder, page *models.Page) (*DirectAccessTokenConnection, error) {
	return NewDirectAccessTokenConnection(ctx, r.uc, filter, order, page,
		func(dat *datModels.DirectAccessToken) *datModels.DirectAccessToken {
			// Return token unmasked if recently created
			if dat.CreatedAt.After(time.Now().Add(-time.Minute * 5)) {
				return dat
			}

			// Mask token value for older tokens
			m := new(datModels.DirectAccessToken)
			*m = *dat
			m.Token = strings.Repeat("*", len(m.Token))
			return m
		}), nil
}
