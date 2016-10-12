package tests

import (
	"testing"
	"time"

	strftime1 "github.com/cactus/gostrftime"
	strftime2 "github.com/hhkbp2/go-strftime"
	strftime3 "github.com/jehiah/go-strftime"
)

func BenchmarkStrftime1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		strftime1.Format("%Y-%m-%d %H:%M:%S,%L", time.Now())
	}
}

func BenchmarkStrftime2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		strftime2.Format("%Y-%m-%d %H:%M:%S,%L", time.Now())
	}
}

func BenchmarkStrftime3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		strftime3.Format("%Y-%m-%d %H:%M:%S,%L", time.Now())
	}
}
