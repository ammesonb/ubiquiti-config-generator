package mocks

import (
	"reflect"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestFunctionMock(t *testing.T) {
	t.Run("Set result for uninitialized mock succeeds", func(t *testing.T) {
		mock := ServiceMock{}
		var testFunc FunctionName = "foo"
		expected := []any{1, 2, 3}
		mock.SetNextResult(testFunc, expected)

		actual, err := mock.GetResult(testFunc)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(expected, actual))
	})

	t.Run("Set result for initialized mock succeeds", func(t *testing.T) {
		mock := ServiceMock{}
		var testFunc FunctionName = "foo"
		mock.ResetFunc(testFunc)
		expected := []any{1, 2, 3}
		mock.SetNextResult(testFunc, expected)

		actual, err := mock.GetResult(testFunc)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(expected, actual))
	})

	t.Run("Get details for unregistered mock fails", func(t *testing.T) {
		mock := ServiceMock{}
		var testFunc FunctionName = "foo"
		result, err := mock.GetResult(testFunc)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errors.ErrWithCtx(ErrNoSuchFunction, testFunc))

		count, err := mock.GetCallCount(testFunc)
		assert.Equal(t, 0, count)
		assert.ErrorIs(t, err, errors.ErrWithCtx(ErrNoSuchFunction, testFunc))

		args, err := mock.GetCallArguments(testFunc)
		assert.Nil(t, args)
		assert.ErrorIs(t, err, errors.ErrWithCtx(ErrNoSuchFunction, testFunc))
	})

	t.Run("Getting result exceeding set mocks fails", func(t *testing.T) {
		mock := ServiceMock{}
		var testFunc FunctionName = "foo"
		mock.Reset()
		mock.ResetFunc(testFunc)
		expected := []any{1, 2, 3}
		mock.SetNextResult(testFunc, expected)

		result, err := mock.GetResult(testFunc)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(expected, result))

		result, err = mock.GetResult(testFunc)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errors.ErrWithVarCtx(errFuncCalledExtra, testFunc, 2, 1))
	})

	t.Run("Called with and call count update", func(t *testing.T) {
		mock := ServiceMock{}
		var testFunc FunctionName = "foo"
		mock.Reset()
		mock.ResetFunc(testFunc)

		expectedArgs := [][]any{{1, 2, 3}, {4, 5, 6}}
		for _, args := range expectedArgs {
			mock.SetNextResult(testFunc, args)
		}

		_, err := mock.GetResult(testFunc, 1, 2, 3)
		assert.NoError(t, err)
		_, err = mock.GetResult(testFunc, 4, 5, 6)
		assert.NoError(t, err)

		count, err := mock.GetCallCount(testFunc)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)

		args, err := mock.GetCallArguments(testFunc)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(expectedArgs, args))
	})
}
