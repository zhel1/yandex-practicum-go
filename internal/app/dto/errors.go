package dto

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrDeleted       = errors.New("marked as deleted")
	ErrAlreadyExists = errors.New("already exists")
	ErrExecutionPSQL = errors.New("execution PSQL error")
	ErrStatementPSQL = errors.New("statement PSQL error")
)
