package uuid

import (
	"github.com/cjysmat/assert"
	"testing"
)

func TestUUID(t *testing.T) {
	uuid, _ := UUID()
	t.Logf("%v\n", uuid)
	assert.Equal(t, 32, len(uuid))
}

func BenchmarkUUID(b *testing.B) {
    for i:=0; i<b.N; i++ {
        UUID()
    }
}
