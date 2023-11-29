//
// @project GeniusRabbit
// @author Dmitry Ponomarev <demdxx@gmail.com> 2016
//

package gosql

import "errors"

// Set of errors
var (
	ErrInvalidScan         = errors.New("invalid field scan")
	ErrInvalidScanValue    = errors.New("invalid field scan value")
	ErrNullValueNotAllowed = errors.New("nil value not allowed")
	ErrInvalidDecodeValue  = errors.New("invalid decode value")
)
