package model

import pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"

// Order type alias for pkg/models.Order
type Order = pkgModels.Order

// Order constants
const (
	OrderUndefined = pkgModels.OrderUndefined
	OrderAsc       = pkgModels.OrderAsc
	OrderDesc      = pkgModels.OrderDesc
)

// OrderFromStr converts a string to an Order value
var OrderFromStr = pkgModels.OrderFromStr
