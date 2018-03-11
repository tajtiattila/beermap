package keyvalue

import "errors"

// ErrNotFound is returned when a key is not found.
var ErrNotFound = errors.New("not found")

type DB interface {
	Get(key string) (value []byte, err error)
	Set(key string, value []byte) error

	// Delete deletes keys. Deleting a non-existent key does not return an error.
	Delete(key string) error

	Close() error

	// Iterator returns an iterator from start, up to but not including end.
	// If end is empty, iteration ends
	Iterator(start, end string) Iterator

	// Batch returns a new batch operation
	// that may be committed to this DB.
	Batch() Batch
}

type Iterator interface {
	// Next moves the iterator to the next key/value pair.
	// It returns false when the iterator is exhausted.
	Next() bool

	// Key returns the key of the current key/value pair.
	// Only valid after a call to Next returns true.
	Key() string

	// Value returns the value of the current key/value pair.
	// Only valid after a call to Next returns true.
	Value() []byte

	// Err returns the first non-EOF error that was encountered by the Iterator.
	Err() error

	// Close closes the iterator and returns any accumulated error. Exhausting
	// all the key/value pairs in a table is not considered to be an error.
	// It is valid to call Close multiple times. Other methods should not be
	// called after the iterator has been closed.
	Close() error
}

// Batch represents a batch operation.
// Set and Delete will be performed sequentially when Commit is called.
// It is not an error if a Batch is not committed.
type Batch interface {
	Set(key string, value []byte)
	Delete(key string)

	Commit() error
}
