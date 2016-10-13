package tests

import (
	"bytes"
	"strings"
	"testing"
)

func BenchmarkStringConcat(b *testing.B) {
	x := "a"
	y := "b"
	z := "c"
	for n := 0; n < b.N; n++ {
		_ = x + y + z
	}
}

func BenchmarkStringsJoin(b *testing.B) {
	x := "a"
	y := "b"
	z := "c"
	for n := 0; n < b.N; n++ {
		_ = strings.Join([]string{x, y, z}, "")
	}
}

func BenchmarkStringByteConcat(b *testing.B) {
	x := []byte("a")
	y := []byte("b")
	z := []byte("c")
	for n := 0; n < b.N; n++ {
		a := append(x, y...)
		_ = append(a, z...)
	}
}

func BenchmarkStringByteConcatJoin(b *testing.B) {
	x := []byte("a")
	y := []byte("b")
	z := []byte("c")
	for n := 0; n < b.N; n++ {
		_ = bytes.Join([][]byte{x, y, z}, []byte(""))
	}
}
