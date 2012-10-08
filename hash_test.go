package main

import (
	"testing"
)

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashPassword([]byte("avgpwd3000$"), []byte("ABCDEFGH"))
	}
}
