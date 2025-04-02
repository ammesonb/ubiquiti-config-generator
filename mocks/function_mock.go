package mocks

import (
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
)

var (
	// ErrNoSuchFunction is used when a function is requested that has not been mocked
	ErrNoSuchFunction  = "no mocked function found with name %s"
	errFuncCalledExtra = "function %s was called %d times, but only has %d mocked results"
)

// FunctionName represents a function name that may be called in the application
type FunctionName string

// FunctionMock tracks called functions and mocked return values
type FunctionMock struct {
	initialized  bool
	returnValues map[FunctionName][][]any
	funcCalls    map[FunctionName]int
	calledWith   map[FunctionName][][]any
}

// ResetFunc clears the mocked return values and call count for a function
func (m *FunctionMock) ResetFunc(name FunctionName) {
	if !m.initialized {
		m.initialized = true
		m.Reset()
	}

	m.returnValues[name] = make([][]any, 0)
	m.funcCalls[name] = 0
	m.calledWith[name] = make([][]any, 0)
}

// Reset clears all mocked return values and call counts
func (m *FunctionMock) Reset() {
	m.returnValues = make(map[FunctionName][][]any)
	m.funcCalls = make(map[FunctionName]int)
	m.calledWith = make(map[FunctionName][][]any)
}

// SetNextResult adds a mocked return value to a function
func (m *FunctionMock) SetNextResult(name FunctionName, values []any) {
	if !m.initialized {
		m.initialized = true
		m.Reset()
	}

	m.returnValues[name] = append(m.returnValues[name], values)
}

// GetResult returns the next mocked return value for a function, indexed by and incrementing call count
func (m *FunctionMock) GetResult(name FunctionName, args ...any) ([]any, error) {
	values, valOk := m.returnValues[name]
	if !valOk {
		return nil, errors.ErrWithCtx(ErrNoSuchFunction, name)
	}

	m.funcCalls[name]++
	m.calledWith[name] = append(m.calledWith[name], args)

	// ensure there are enough mocks for the number of times this function is called
	// assuming one set of mocks added, first call 1 mocks < 1 calls = false, so return the values
	// on second call, 1 mocks < 2 calls = true, so return an error
	if len(values) < m.funcCalls[name] {
		return nil, errors.ErrWithVarCtx(errFuncCalledExtra, name, m.funcCalls[name], len(values))
	}

	return values[m.funcCalls[name]-1], nil
}

// GetCallCount returns the number of times a function was called
func (m *FunctionMock) GetCallCount(name FunctionName) (int, error) {
	count, ok := m.funcCalls[name]
	if !ok {
		return 0, errors.ErrWithCtx(ErrNoSuchFunction, name)
	}

	return count, nil
}

// GetCallArguments returns the arguments passed to a function
func (m *FunctionMock) GetCallArguments(name FunctionName) ([][]any, error) {
	args, ok := m.calledWith[name]
	if !ok {
		return nil, errors.ErrWithCtx(ErrNoSuchFunction, name)
	}

	return args, nil
}

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

func SetNextResult(name string, value any) error {
	if _, ok := returnValues[name]; !ok {
		return errors.ErrWithCtx(ErrNoSuchFunction, name)
	}

	returnValues[name] = append(returnValues[name], value)

	return nil
}

// GetResult retrieves the expected return for the given function and increments its call count
func GetResult(name string) (any, error) {
	values, valOk := returnValues[name]
	callCount, countOk := funcCalls[name]
	if !valOk || !countOk {
		return nil, errors.ErrWithCtx(ErrNoSuchFunction, name)
	}
	// less than or equal to, since call count is increased at the end to account for
	// zero-indexed arrays based on first vs second vs third calls
	if len(values) <= callCount {
		return nil, errors.ErrWithVarCtx(errFuncCalledExtra, name, callCount+1, len(values))
	}

	funcCalls[name]++
	return values[callCount], nil
}

// AnyToError takes any value and returns it cast as an error if non-nil
// intended to be used with mocked result values, which must be accessed as type any
func AnyToError(err any) error {
	if err == nil {
		return nil
	}
	return err.(error)
}
