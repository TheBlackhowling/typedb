package typedb

import "errors"

// ErrNotFound is returned when a query returns no rows.
// This wraps sql.ErrNoRows to provide a typedb-specific error.
var ErrNotFound = errors.New("typedb: record not found")
