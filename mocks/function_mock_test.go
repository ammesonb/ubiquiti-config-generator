package mocks

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMockFunctionResults(t *testing.T) {
	funcName := "testfunc"
	InitOrClearFuncReturn(funcName)
	res, err := GetResult(funcName)
	assert.Nil(t, res, "No results if function return not mocked")
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errFuncCalledExtra, funcName, 1, 0), "No results registered")

	firstRes := []any{"foo", nil}
	secondRes := []any{"bar", nil}
	assert.NoError(t, SetNextResult(funcName, firstRes))
	assert.NoError(t, SetNextResult(funcName, secondRes))

	res, err = GetResult(funcName)
	assert.True(t, reflect.DeepEqual(res, firstRes))
	assert.NoError(t, err)
	res, err = GetResult(funcName)
	assert.True(t, reflect.DeepEqual(res, secondRes))
	assert.NoError(t, err)

	ClearAll()
	res, err = GetResult(funcName)
	assert.Nil(t, res, "No result if function was not registered")
	assert.ErrorIs(t, err, utils.ErrWithCtx(errNoSuchFunction, funcName))

	err = SetNextResult(funcName, true)
	assert.ErrorIs(t, err, utils.ErrWithCtx(errNoSuchFunction, funcName))

}
