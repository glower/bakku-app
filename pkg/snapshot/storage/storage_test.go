package storage

import (
	"fmt"
	"testing"
)

type FakeStorage struct {
	path       string
	DBFilePath string // /foo/bar/.snapshot
	DBFileName string // .snapshot
}

func (f *FakeStorage) Add(filePath, bucketName string, value []byte) error {
	return nil
}
func (f *FakeStorage) Exist() bool {
	return true
}
func (f *FakeStorage) Path() string {
	return f.path
}
func (f *FakeStorage) FilePath() string {
	return f.DBFilePath //"/foo/bar/.fake"
}
func (f *FakeStorage) FileName() string {
	return f.DBFileName // ".fake"
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

	testBackupPath := "/foo/bar"
	testDBFileName := ".fake"
	testDBFilePath := fmt.Sprintf("%s/%s", testBackupPath, testDBFileName)

	fake := &FakeStorage{
		path:       testBackupPath,
		DBFilePath: testDBFilePath,
		DBFileName: testDBFileName,
	}

	tests := []struct {
		name        string
		args        args
		path        string
		errExpected bool
	}{
		{
			name: "1: register a snapshot storage",
			args: args{
				s: fake,
			},
			path:        testBackupPath,
			errExpected: false,
		},
		{
			name: "2: register empty backup storage",
			args: args{
				s: nil,
			},
			path:        "/some/path",
			errExpected: true,
		},
		{
			name: "3: register empty path backup storage",
			args: args{
				s: &FakeStorage{
					path:       "",
					DBFilePath: "",
					DBFileName: "",
				},
			},
			path:        "/some/path",
			errExpected: true,
		},
		{
			name: "4: get wrong path backup storage",
			args: args{
				s: fake,
			},
			path:        "/some/path",
			errExpected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Register(tt.args.s)
			storage, err := GetByPath(tt.path)

			if !tt.errExpected && err != nil {
				t.Errorf("storage.FileName(): error was not expected: [%v]", err)
			}

			if tt.errExpected && err == nil {
				t.Errorf("storage.FileName(): error was expected but it's empty")
			}

			if !tt.errExpected && storage.FileName() != testDBFileName {
				t.Errorf("storage.FileName(): value [%s], want [%s]", storage.FileName(), testDBFileName)
			}
		})
	}
}
