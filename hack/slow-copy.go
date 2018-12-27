package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var BUFFERSIZE int64

func copy(src, dst string, BUFFERSIZE int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("file %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}
	totalWritten := 0
	totalSize := sourceFileStat.Size()
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		// write a chunk
		var written = 0
		if written, err = destination.Write(buf[:n]); err != nil {
			return err
		}
		totalWritten = totalWritten + written

		var percent float64
		if int64(written) == totalSize {
			percent = float64(100)
		} else {
			percent = float64(100 * int64(totalWritten) / totalSize)
		}

		time.Sleep(1 * time.Second)
		fmt.Printf("%2.0f%%\n", percent)
	}
	return err
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("usage: %s source destination BUFFERSIZE\n", filepath.Base(os.Args[0]))
		return
	}

	source := os.Args[1]
	destination := os.Args[2]
	BUFFERSIZE, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		fmt.Printf("Invalid buffer size: %q\n", err)
		return
	}

	fmt.Printf("Copying %s to %s\n", source, destination)
	err = copy(source, destination, BUFFERSIZE)
	if err != nil {
		fmt.Printf("File copying failed: %q\n", err)
	}
}
