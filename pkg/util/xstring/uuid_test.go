package xstring

import (
	"testing"
)

func TestGenerateShortUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(GenerateShortUUID())
	}
}
