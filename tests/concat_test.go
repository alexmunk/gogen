package tests

import (
	"bytes"
	"strings"
	"testing"
)

func BenchmarkStringConcat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		x := "a"
		y := "b"
		z := "c"
		_ = x + y + z
	}
}

func BenchmarkStringsJoin(b *testing.B) {
	for n := 0; n < b.N; n++ {
		x := "a"
		y := "b"
		z := "c"
		_ = strings.Join([]string{x, y, z}, "")
	}
}

func BenchmarkStringByteConcat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		x := []byte("a")
		y := []byte("b")
		z := []byte("c")
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
