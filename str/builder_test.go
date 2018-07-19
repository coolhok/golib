package str

import (
	"github.com/cjysmat/assert"
	"testing"
)

func TestStringBuilder(t *testing.T) {
	sb := NewStringBuilder()
	sb.WriteString("hello")
	sb.WriteString("world")
	assert.Equal(t, "helloworld", sb.String())
}
