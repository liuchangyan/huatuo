package sml

import "fmt"

type Return int32

const (
	SUCCESS Return = iota
	ERROR_FAILURE
	ERROR_NO_DEVICE
	ERROR_OPERATION_NOT_SUPPORTED
)

// String returns the string representation of a MetaxSmlReturnCode.
func (r Return) String() string {
	return r.Error()
}

// Error returns the string representation of a MetaxSmlReturnCode.
func (r Return) Error() string {
	return errorStringFunc(r)
}

// errorStringFunc can be assigned if the system metax-sml library is in use.
var errorStringFunc = defaultErrorStringFunc

// defaultErrorStringFunc provides a basic implementation for MetaxSmlReturnCode string representation.
var defaultErrorStringFunc = func(r Return) string {
	switch r {
	case SUCCESS:
		return "METAX_SML_SUCCESS"
	case ERROR_FAILURE:
		return "METAX_SML_FAILURE"
	case ERROR_NO_DEVICE:
		return "METAX_SML_NO_DEVICE"
	case ERROR_OPERATION_NOT_SUPPORTED:
		return "METAX_SML_OPERATION_NOT_SUPPORTED"
	default:
		return fmt.Sprintf("unknown MetaxSML return code: %d", r)
	}
}
