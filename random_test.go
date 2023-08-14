package gaul

import (
	"math/rand"
	"testing"
)

func BenchmarkLFSRSmall_Next(b *testing.B) {
	prng := NewLFSRSmall()
	for i := 0; i < b.N; i++ {
		prng.Next()
	}
}

func BenchmarkLFSRMedium_Next(b *testing.B) {
	prng := NewLFSRMedium()
	for i := 0; i < b.N; i++ {
		prng.Next()
	}
}

func BenchmarkLFSRLarge_Next(b *testing.B) {
	prng := NewLFSRLarge()
	for i := 0; i < b.N; i++ {
		prng.Next()
	}
}

func BenchmarkLFSRLarge_Float64(b *testing.B) {
	prng := NewLFSRLarge()
	for i := 0; i < b.N; i++ {
		prng.Float64()
	}
}

func BenchmarkStandardRand_Intn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Int63()
	}
}

func BenchmarkStandardRand_Float64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Float64()
	}
}
