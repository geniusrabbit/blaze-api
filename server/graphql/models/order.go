package models

import (
	"fmt"

	"github.com/geniusrabbit/blaze-api/model"
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

func (order *Ordering) AsOrder() model.Order {
	if order != nil {
		fmt.Println("order: ", *order)
		switch *order {
		case OrderingAsc:
			return model.OrderAsc
		case OrderingDesc:
			return model.OrderDesc
		}
	}
	return model.OrderUndefined
}
