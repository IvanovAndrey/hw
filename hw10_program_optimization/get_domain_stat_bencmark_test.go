package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"testing"
)

func openUsersDat(b *testing.B) io.Reader {
	b.Helper()
	zipData, err := os.ReadFile("testdata/users.dat.zip")
	if err != nil {
		b.Fatalf("failed to read zip: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		b.Fatalf("failed to create zip reader: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == "users.dat" {
			rc, err := f.Open()
			if err != nil {
				b.Fatalf("failed to open users.dat: %v", err)
			}
			b.Cleanup(func() { rc.Close() })
			return rc
		}
	}

	b.Fatalf("users.dat not found in zip")
	return nil
}

func BenchmarkGetDomainStat(b *testing.B) {
	domain := "example.com"

	for i := 0; i < b.N; i++ {
		r := openUsersDat(b)
		_, err := GetDomainStat(r, domain)
		if err != nil {
			b.Fatalf("GetDomainStat error: %v", err)
		}
	}
}

func BenchmarkGetDomainStatOld(b *testing.B) {
	domain := "example.com"

	for i := 0; i < b.N; i++ {
		r := openUsersDat(b)
		_, err := GetDomainStatOld(r, domain)
		if err != nil {
			b.Fatalf("GetDomainStat error: %v", err)
		}
	}
}
