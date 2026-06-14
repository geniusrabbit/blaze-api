package models

import (
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/models"
)

func (order *Ordering) Int8() int8 {
	if order != nil {
		if *order == OrderingAsc {
			return 1
		}
		if *order == OrderingDesc {
			return -1
		}
	}
	return 0
}

func (order *Ordering) AsOrder() models.Order {
	if order != nil {
		fmt.Println("order: ", *order)
		switch *order {
		case OrderingAsc:
			return models.OrderAsc
		case OrderingDesc:
			return models.OrderDesc
		}
	}
	return models.OrderUndefined
}
