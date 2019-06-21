package boltdb

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func testDBFileName() string {
	return fmt.Sprintf(".snapshot.%d.%d", time.Now().Unix(), rand.Intn(99999))
}

func testDBFilePath(testDBFileName string) string {
	return filepath.Join(os.TempDir(), testDBFileName)
}

func TestStorage_Add(t *testing.T) {
	testDBFileName := testDBFileName()
	testDBFilePath := testDBFilePath(testDBFileName)
	type args struct {
		filePath   string
		bucketName string
		value      []byte
	}

	testStorage := Storage{
		DBFileName: testDBFileName,
		DBFilePath: testDBFilePath,
	}

	tests := []struct {
		name    string
		s       *Storage
		args    args
		wantErr bool
	}{
		{
			name: "Scenario 1: add a value",
			s:    &testStorage,
			args: args{
				filePath:   "/foo",
				bucketName: "test",
				value:      []byte("bar"),
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: empty bucket name",
			s:    &testStorage,
			args: args{
				filePath:   "/foo",
				bucketName: "",
				value:      []byte("bar"),
			},
			wantErr: true,
		},
		{
			name: "Scenario 3: empty key",
			s:    &testStorage,
			args: args{
				filePath:   "",
				bucketName: "test",
				value:      []byte("bar"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = tt.s.Add(tt.args.filePath, tt.args.bucketName, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Storage.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				value, getErr := tt.s.Get(tt.args.filePath, tt.args.bucketName)
				if getErr != nil {
					t.Errorf("Storage.Get() error = %v", getErr)
				}
				if value != string(tt.args.value) {
					t.Errorf("Storage.Get() value %s, want %s", value, tt.args.value)
				}
			}
		})
	}
}

func TestStorage_Get(t *testing.T) {
	testDBFileName := testDBFileName()
	testDBFilePath := testDBFilePath(testDBFileName)
	type args struct {
		filePath   string
		bucketName string
	}
	testStorage := Storage{
		DBFileName: testDBFileName,
		DBFilePath: testDBFilePath,
	}

	err := testStorage.Add("/foo/bar", "test", []byte("buzz"))
	if err != nil {
		t.Errorf("storage.Add(): error was not expected: [%v]", err)
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Scenario 1: get a value",
			s:    &testStorage,
			args: args{
				filePath:   "/foo/bar",
				bucketName: "test",
			},
			want:    "buzz",
			wantErr: false,
		},
		{
			name: "Scenario 2: try to get a value from a wrong bucket",
			s:    &testStorage,
			args: args{
				filePath:   "/foo/bar",
				bucketName: "test2",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Scenario 3: try to get a value from a wrong key",
			s:    &testStorage,
			args: args{
				filePath:   "/foo/bar/xxx",
				bucketName: "test",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Scenario 4: try to get a value from an empty key",
			s:    &testStorage,
			args: args{
				filePath:   "",
				bucketName: "test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Scenario 5: try to get a value from an empty bucket",
			s:    &testStorage,
			args: args{
				filePath:   "/foo/bar",
				bucketName: "",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Get(tt.args.filePath, tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Storage.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	testDBFileName := testDBFileName()
	testDBFilePath := testDBFilePath(testDBFileName)
	type args struct {
		bucketName string
	}

	testStorage := Storage{
		DBFileName: testDBFileName,
		DBFilePath: testDBFilePath,
	}

	err := testStorage.Add("/foo/bar", "test", []byte("buzz"))
	if err != nil {
		t.Errorf("storage.Add(): error was not expected: [%v]", err)
	}
	err = testStorage.Add("/aaa", "test", []byte("bbb"))
	if err != nil {
		t.Errorf("storage.Add(): error was not expected: [%v]", err)
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Scenario 1: get all items from a bucket name",
			s:    &testStorage,
			args: args{
				bucketName: "test",
			},
			want: map[string]string{
				"/foo/bar": "buzz",
				"/aaa":     "bbb",
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: try to get all items from a wrong bucket name",
			s:    &testStorage,
			args: args{
				bucketName: "test2",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetAll(tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetAll(): error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetAll(): got %v, want %v", got, tt.want)
			}
		})
	}
}
