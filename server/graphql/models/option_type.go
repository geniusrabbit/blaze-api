package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type OptionType string

const (
	OptionTypeUndefined OptionType = "UNDEFINED"
	OptionTypeUser      OptionType = "USER"
	OptionTypeAccount   OptionType = "ACCOUNT"
	OptionTypeSystem    OptionType = "SYSTEM"
)

var AllOptionType = []OptionType{
	OptionTypeUndefined,
	OptionTypeUser,
	OptionTypeAccount,
	OptionTypeSystem,
}

func (e OptionType) IsValid() bool {
	switch e {
	case OptionTypeUndefined, OptionTypeUser, OptionTypeAccount, OptionTypeSystem:
		return true
	}
	return false
}

func (e OptionType) String() string {
	return string(e)
}

func (e *OptionType) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = OptionType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid OptionType", str)
	}
	return nil
}

func (e OptionType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *OptionType) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	return e.UnmarshalGQL(s)
}

func (e OptionType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)
	return buf.Bytes(), nil
}
