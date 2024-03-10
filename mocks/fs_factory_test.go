package mocks

import (
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestMockFileInfo_Name(t *testing.T) {
	mockFileInfo := &MockFileInfo{name: "sample name"}
	assert.Equal(t, "sample name", mockFileInfo.Name())
}

func TestMockFileInfo_Mode(t *testing.T) {
	mockFileInfo := &MockFileInfo{mode: 1}
	assert.Equal(t, os.FileMode(1), mockFileInfo.Mode())
}

func TestMockFileInfo_Size(t *testing.T) {
	mockFileInfo := &MockFileInfo{size: 1}
	assert.Equal(t, int64(1), mockFileInfo.Size())
}

func TestMockFileInfo_IsDir(t *testing.T) {
	mockFileInfo := &MockFileInfo{isDir: true}
	assert.True(t, mockFileInfo.IsDir())
}

func TestMockFileInfo_ModTime(t *testing.T) {
	now := time.Now()
	mockFileInfo := &MockFileInfo{modTime: now}
	assert.Equal(t, now, mockFileInfo.ModTime())
}

func TestMockFileInfo_Sys(t *testing.T) {
	mockFileInfo := &MockFileInfo{}
	assert.Nil(t, mockFileInfo.Sys())
}

func TestMockDirEntry_Name(t *testing.T) {
	mockDirEntry := &MockDirEntry{IName: "sample name"}
	assert.Equal(t, "sample name", mockDirEntry.Name())
}

func TestMockDirEntry_Type(t *testing.T) {
	mockDirEntry := &MockDirEntry{IMode: 1}
	assert.Equal(t, os.FileMode(1), mockDirEntry.Type())
}

func TestMockDirEntry_IsDir(t *testing.T) {
	mockDirEntry := &MockDirEntry{IIsDir: true}
	assert.True(t, mockDirEntry.IsDir())
}

func TestMockDirEntry_Info(t *testing.T) {
	mockDirEntry := &MockDirEntry{IPath: "./", IName: "fs_factory.go", IStat: os.Stat}
	info, err := mockDirEntry.Info()
	assert.NotNil(t, info)
	assert.NoError(t, err)

	assert.Greater(t, info.Size(), int64(0))
	assert.False(t, info.IsDir())
}

func TestGetFsWrapper(t *testing.T) {
	wrapper := GetFsWrapper()
	if wrapper != GetFsWrapper() {
		t.Errorf("GetFsWrapper should return the same instance")
	}
	if reflect.ValueOf(wrapper.Stat).Pointer() != reflect.ValueOf(os.Stat).Pointer() {
		t.Errorf("Stat function should use default implementation")
	}
	if reflect.ValueOf(wrapper.ReadDir).Pointer() != reflect.ValueOf(os.ReadDir).Pointer() {
		t.Errorf("ReadDir function should use default implementation")
	}
}
