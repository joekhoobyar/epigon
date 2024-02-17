package storage

const (
	ErrFailed = iota
	ErrLocationNotObject
	ErrLocationNotPrefix
	ErrKindNotObject
	ErrKindNotPrefix
	ErrObjectNotFound
	ErrPrefixNotFound
)

var (
	errReasons = []string{
		"failed",
		"location does not identify an object",
		"location does not identify a collection",
		"not an object record",
		"not a prefix record",
		"no such object record",
		"no such prefix record",
	}
)

type StorageError struct {
	Location   string
	reason     int
	underlying error
}

func newError(location string, reason int) *StorageError {
	return &StorageError{Location: location, reason: reason}
}

func wrapFailure(err error, location string) *StorageError {
	return wrapError(err, location, ErrFailed)
}

func wrapError(err error, location string, reason int) *StorageError {
	return &StorageError{Location: location, reason: reason, underlying: err}
}

func (e *StorageError) Error() string {
	message := e.Location
	if e.reason > ErrFailed || e.underlying == nil {
		message += ": " + errReasons[e.reason]
	}
	if e.underlying != nil {
		message += ": " + e.underlying.Error()
	}
	return message
}

func (e *StorageError) Unwrap() error { return e.underlying }

func IsLocationNotObject(err error) bool { return IsError(err, ErrLocationNotObject) }

func IsLocationNotPrefixt(err error) bool { return IsError(err, ErrLocationNotPrefix) }

func IsKindNotObject(err error) bool { return IsError(err, ErrKindNotObject) }

func IsKindNotPrefix(err error) bool { return IsError(err, ErrKindNotPrefix) }

func IsObjectNotFound(err error) bool { return IsError(err, ErrObjectNotFound) }

func IsPrefixNotFound(err error) bool { return IsError(err, ErrPrefixNotFound) }

func IsError(err error, kind int) bool {
	e, ok := err.(*StorageError)
	return ok && e.reason == kind
}
