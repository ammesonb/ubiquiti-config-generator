package mocks

import (
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
)

type funcReturn []any

// This is not designed for parallel support with shared memory

// Tracks expected return values from a function
var returnValues = make(map[string]funcReturn)

// Tracks how many times a function was called
var funcCalls = make(map[string]int)

// InitOrClearFuncReturn sets up emulated function return values, or clears them if already set
func InitOrClearFuncReturn(name string) {
	funcCalls[name] = 0
	returnValues[name] = make(funcReturn, 0)
}

// ClearAll reset the entire function mocking state
func ClearAll() {
	returnValues = make(map[string]funcReturn)
	funcCalls = make(map[string]int)
}

var errNoSuchFunction = "no mocked function found with name %s"
var errFuncCalledExtra = "function %s was called %d times, but only has %d mocked results"

func SetNextResult(name string, value any) error {
	if _, ok := returnValues[name]; !ok {
		return errors.ErrWithCtx(errNoSuchFunction, name)
	}

	returnValues[name] = append(returnValues[name], value)

	return nil
}

// GetResult retrieves the expected return for the given function and increments its call count
func GetResult(name string) (any, error) {
	values, valOk := returnValues[name]
	callCount, countOk := funcCalls[name]
	if !valOk || !countOk {
		return nil, errors.ErrWithCtx(errNoSuchFunction, name)
	}
	// less than or equal to, since call count is increased at the end to account for
	// zero-indexed arrays based on first vs second vs third calls
	if len(values) <= callCount {
		return nil, errors.ErrWithVarCtx(errFuncCalledExtra, name, callCount+1, len(values))
	}

	funcCalls[name]++
	return values[callCount], nil
}
