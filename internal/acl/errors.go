package acl

import "errors"

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

var ErrNoPermissions = errors.New("no permissions")
