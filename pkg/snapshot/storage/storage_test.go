package storage

import (
	"fmt"
	"testing"
)

type FakeStorage struct {
	DBFilePath string // /foo/bar/.snapshot
	DBFileName string // .snapshot
}

func (f *FakeStorage) Add(filePath, bucketName string, value []byte) error {
	return nil
}
func (f *FakeStorage) Exist() bool {
	return true
}
func (f *FakeStorage) FilePath() string {
	return "/foo/bar/.fake"
}
func (f *FakeStorage) FileName() string {
	return ".fake"
}
func (f *FakeStorage) Get(filePath, bucketName string) (string, error) {
	if filePath == "/foo/bar/test.txt" {
		return "abc", nil
	}
	return "", fmt.Errorf("not found")
}
func (f *FakeStorage) GetAll(filePath string) (map[string]string, error) {
	return map[string]string{
		"/foo/bar/test.txt": "abc",
	}, nil
}
func (f *FakeStorage) Remove(filePath string, bucketName string) error {
	return nil
}

func TestRegisterStorage(t *testing.T) {
	type args struct {
		path string
		s    Storage
	}

	fake := &FakeStorage{
		DBFilePath: "/foo/bar/.fake",
		DBFileName: ".fake",
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "1: register fake storage",
			args: args{
				path: "/foo/bar",
				s:    fake,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Register(tt.args.path, tt.args.s)
			storage := GetByPath(tt.args.path)
			if storage.FileName() != ".fake" {
				t.Errorf("storage.FileName(): value [%s], want [%s]", storage.FileName(), ".fake")
			}
		})
	}
}
