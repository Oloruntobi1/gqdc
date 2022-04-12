package util

import (
	"fmt"
)

var (
	ErrUserNotFound   = fmt.Errorf("user not found")
	ErrWalletNotFound = fmt.Errorf("wallet not found")
)

type DBError struct {
	Summary string
	Err     error
}

func (e *DBError) Error() string {
	return fmt.Sprintf("%s: %v", e.Summary, e.Err)
}

type CreateSchemaError = DBError

func NewCreateSchemaError(err error) *CreateSchemaError {
	return &CreateSchemaError{
		Summary: "failed creating schema resources",
		Err:     err,
	}
}

type ConnectionError = DBError

func NewConnectionError(err error) *ConnectionError {
	return &ConnectionError{
		Summary: "failed opening connection to database",
		Err:     err,
	}
}

type CreateFKError = DBError

func NewCreateFKError(err error) *CreateFKError {
	return &CreateFKError{
		Summary: "failed creating foreign key constraint",
		Err:     err,
	}
}
