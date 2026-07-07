package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/geniusrabbit/blaze-api/pkg/models"
)

// Constants of the order of data
type Ordering string

const (
	// Ascending ordering of data
	OrderingAsc Ordering = "ASC"
	// Descending ordering of data
	OrderingDesc Ordering = "DESC"
)

var AllOrdering = []Ordering{
	OrderingAsc,
	OrderingDesc,
}

func (e Ordering) IsValid() bool {
	switch e {
	case OrderingAsc, OrderingDesc:
		return true
	}
	return false
}

func (e Ordering) String() string {
	return string(e)
}

func (e *Ordering) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Ordering(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Ordering", str)
	}
	return nil
}

func (e Ordering) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *Ordering) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	return e.UnmarshalGQL(s)
}

func (e Ordering) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)
	return buf.Bytes(), nil
}

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
