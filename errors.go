package typedb

import "errors"

// ErrNotFound is returned when a query returns no rows.
// This wraps sql.ErrNoRows to provide a typedb-specific error.
var ErrNotFound = errors.New("typedb: record not found")

// ErrFieldNotFound is returned when a field cannot be found.
var ErrFieldNotFound = errors.New("typedb: field not found")

// ErrMethodNotFound is returned when a method cannot be found.
var ErrMethodNotFound = errors.New("typedb: method not found")

// errNotMyType is returned by handler functions when they don't handle the target type.
// This allows the main function to try the next handler without logging errors.
var errNotMyType = errors.New("typedb: not my type")
