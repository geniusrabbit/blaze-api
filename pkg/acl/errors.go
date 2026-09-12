package acl

import "fmt"

// import (
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"

// 	"github.com/grpc-ecosystem/grpc-gateway/runtime"
// )

// // PermissionError represents common ACL error object
// type PermissionError struct {
// 	status *status.Status
// }

// // NewPermissionError creates new permission error with the message
// func NewPermissionError(code codes.Code, message string) *PermissionError {
// 	return &PermissionError{status: status.New(code, message)}
// }

// // Error returns error message
// func (err *PermissionError) Error() string {
// 	return err.status.Message()
// }

// // GRPCStatus returns GRPC status of error
// func (err *PermissionError) GRPCStatus() *status.Status {
// 	return err.status
// }

// // HTTPCode from GRPC status
// func (err *PermissionError) HTTPCode() int {
// 	return runtime.HTTPStatusFromCode(err.GRPCStatus().Code())
// }

// // The list of common errors
// var (
// 	ErrNoPermissions = NewPermissionError(codes.PermissionDenied, "no permissions")
// )

type ACLError struct {
	parent  error
	Message string
}

// Error returns the error message.
func (err *ACLError) Error() string {
	if err.parent != nil {
		return err.parent.Error() + ": " + err.Message
	}
	return err.Message
}

// WithMessage returns a new error with the message.
func (err *ACLError) WithMessage(message string) *ACLError {
	nErr := &ACLError{
		parent:  err,
		Message: message,
	}
	return nErr
}

// WithMessagef returns a new error with the message formatted.
func (err *ACLError) WithMessagef(format string, a ...any) *ACLError {
	return err.WithMessage(fmt.Sprintf(format, a...))
}

// Unwrap returns the parent error.
func (err *ACLError) Unwrap() error {
	return err.parent
}

var ErrNoPermissions = &ACLError{Message: "no permissions"}
