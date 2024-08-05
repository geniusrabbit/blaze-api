package repository

import "gorm.io/gorm"

// QueryPreparer prepare query
type QueryPreparer interface {
	PrepareQuery(query *gorm.DB) *gorm.DB
}

// List select options
type listOptions[F, O QueryPreparer] struct {
	Filter   F
	Order    O
	Page     *Pagination
	Preloads []string
}

// Option for list query
type ListOption[F, O QueryPreparer] func(*listOptions[F, O])

// WithFilter option for list query
func WithFilter[F, O QueryPreparer](filter F) ListOption[F, O] {
	return func(opts *listOptions[F, O]) {
		opts.Filter = filter
	}
}

// WithOrder option for list query
func WithOrder[F, O QueryPreparer](order O) ListOption[F, O] {
	return func(opts *listOptions[F, O]) {
		opts.Order = order
	}
}

// WithPagination option for list query
func WithPagination[F, O QueryPreparer](page *Pagination) ListOption[F, O] {
	return func(opts *listOptions[F, O]) {
		opts.Page = page
	}
}

// WithPreloads option for list query
func WithPreloads[F, O QueryPreparer](preloads ...string) ListOption[F, O] {
	return func(opts *listOptions[F, O]) {
		opts.Preloads = preloads
	}
}

// PrepareQuery prepare query with options
func (opts *listOptions[F, O]) PrepareQuery(query *gorm.DB) *gorm.DB {
	if len(opts.Preloads) > 0 {
		for _, preload := range opts.Preloads {
			query = query.Preload(preload)
		}
	}
	query = opts.Filter.PrepareQuery(query)
	query = opts.Order.PrepareQuery(query)
	query = opts.Page.PrepareQuery(query)
	return query
}

// ListOptionsInit initialize list options
func ListOptionsInit[F, O QueryPreparer](opts ...ListOption[F, O]) *listOptions[F, O] {
	var options listOptions[F, O]
	for _, opt := range opts {
		opt(&options)
	}
	return &options
}

// PrepareListOptions prepare query with options
func PrepareListOptions[F, O QueryPreparer](db *gorm.DB, opts ...ListOption[F, O]) *gorm.DB {
	return ListOptionsInit(opts...).PrepareQuery(db)
}

// ListOptions for query preparation
type ListOptions[F QueryPreparer, O QueryPreparer] []ListOption[F, O]

func (opts ListOptions[F, O]) WithFilter(filter F) ListOptions[F, O] {
	return append(opts, WithFilter[F, O](filter))
}

func (opts ListOptions[F, O]) WithOrder(order O) ListOptions[F, O] {
	return append(opts, WithOrder[F, O](order))
}

func (opts ListOptions[F, O]) WithPagination(page *Pagination) ListOptions[F, O] {
	return append(opts, WithPagination[F, O](page))
}

func (opts ListOptions[F, O]) WithPreloads(preloads ...string) ListOptions[F, O] {
	return append(opts, WithPreloads[F, O](preloads...))
}

func (opts ListOptions[F, O]) PrepareQuery(query *gorm.DB) *gorm.DB {
	return PrepareListOptions(query, opts...)
}
