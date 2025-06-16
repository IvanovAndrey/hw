package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"testing"
)

func readTestData(b *testing.B) io.Reader {
	b.Helper()
	zipFile, err := os.ReadFile("testdata/users.dat.zip")
	if err != nil {
		b.Fatalf("cannot read zip file: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		b.Fatalf("cannot open zip reader: %v", err)
	}

	var file *zip.File
	for _, f := range zr.File {
		if f.Name == "users.dat" {
			file = f
			break
		}
	}
	if file == nil {
		b.Fatalf("users.dat not found in zip")
	}

	rc, err := file.Open()
	if err != nil {
		b.Fatalf("cannot open users.dat in zip: %v", err)
	}
	b.Cleanup(func() { rc.Close() })
	return rc
}

func BenchmarkGetUsersOld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := readTestData(b)

		_, err := getUsersOld(f)
		if err != nil {
			b.Fatalf("getUsersOld failed: %v", err)
		}
	}
}

func BenchmarkGetUsers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := readTestData(b)

		_, err := getUsers(f)
		if err != nil {
			b.Fatalf("getUsers failed: %v", err)
		}
	}
}
