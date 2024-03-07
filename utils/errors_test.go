package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type errFake struct{}

func (e errFake) Error() string {
	return "foo"
}

func TestErrors(t *testing.T) {
	err := ErrWithCtx("test error %s", "foo")
	assert.Equal(t, err.Error(), "test error foo", "Error formats correctly")

	err = ErrWithVarCtx("test %s error %s %d", "hello", "world", 2)
	assert.Equal(t, err.Error(), "test hello error world 2", "Error with variable ctx")

	err = Err("test error")
	assert.Equal(t, err.Error(), "test error", "Simple error returned")

	err2 := ErrWithParent("ipsum", err)
	assert.ErrorIs(t, err2, UbiConfGenError{Message: "test error", ParentErr: err}, "error matches")
	assert.ErrorIs(t, err2, err, "error contains parent")

	assert.False(t, err.Matches(os.ErrNotExist))

	assert.Equal(
		t,
		ErrWithCtx("hello world", "ignored").Error(),
		"hello world",
		"context ignored",
	)

	noParent := ErrWithCtxParent("test message %s", "foo", nil)
	withParent := ErrWithCtxParent("test message %s", "foo", os.ErrNotExist)
	assert.True(t, noParent.Matches(withParent), "Errors match regardless of parent")

	assert.Equal(t, noParent.Error(), "test message foo", "Parent error not included")
	assert.Equal(
		t,
		withParent.Error(),
		"test message foo: "+os.ErrNotExist.Error(),
		"Parent error included",
	)

	noParent = ErrWithVarCtxParent("test message %s %d", nil, "foo", 1)
	withParent = ErrWithVarCtxParent("test message %s %d", os.ErrNotExist, "foo", 1)
	assert.True(t, noParent.Matches(withParent), "Errors with vararg ctx match regardless of parent")

	one := ErrWithVarCtxParent("test message %s %d", os.ErrNotExist, "foo", 1)
	two := ErrWithVarCtxParent("test message %s %d", os.ErrNotExist, "foo", 2)
	assert.False(t, one.Matches(two), "Different context does not match")
}
