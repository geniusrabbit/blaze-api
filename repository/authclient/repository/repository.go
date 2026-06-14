// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"time"

	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/generated"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

// Repository DAO which provides functionality of working with AuthClients.
// FetchList and Count are provided by the embedded generated.Repository.
type Repository struct {
	generated.Repository[authclient.AuthClient, string]
}

// NewAuthclientRepository creates a new instance of the AuthClient repository
func NewAuthclientRepository() *Repository {
	return &Repository{Repository: *generated.NewRepository[authclient.AuthClient, string]()}
}

// Get returns AuthClient by ID. Uses Find (not First) to keep the original
// no-error-on-not-found behavior required by the authclient domain interface.
func (r *Repository) Get(ctx context.Context, id string) (*authclient.AuthClient, error) {
	object := new(authclient.AuthClient)
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// Create adds a new AuthClient, auto-generating a UUID when the ID is empty.
func (r *Repository) Create(ctx context.Context, roleObj *authclient.AuthClient, opts ...authclient.QOption) (string, error) {
	if roleObj.ID == "" {
		roleObj.ID = newID()
	}
	roleObj.CreatedAt = time.Now()
	roleObj.UpdatedAt = roleObj.CreatedAt
	db := authclient.ListOptions(opts).PrepareQuery(r.Master(historylog.WithPK(ctx, roleObj.ID)))
	err := db.Create(roleObj).Error
	return roleObj.ID, err
}

// Update saves partial changes (non-zero fields only) to an existing AuthClient.
func (r *Repository) Update(ctx context.Context, id string, roleObj *authclient.AuthClient, opts ...authclient.QOption) error {
	obj := *roleObj
	obj.ID = id
	return authclient.ListOptions(opts).PrepareQuery(r.Master(ctx)).Updates(&obj).Error
}

// Delete removes an AuthClient by ID.
func (r *Repository) Delete(ctx context.Context, id string, opts ...authclient.QOption) error {
	return r.Repository.Delete(historylog.WithPK(ctx, id), id, opts...)
}
