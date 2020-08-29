package rand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringWithLetters(t *testing.T) {
	assert := assert.New(t)

	letters := "123456789"
	num := 10
	randStr := StringWithLetters(letters, num)
	assert.NotEqual(randStr, "")
	assert.Equal(len(randStr), num)
	t.Log("random string1", randStr)
	t.Log("random string2", StringWithLetters(letters, num))
	t.Log("random string3", StringWithLetters(letters, num))
	t.Log("random string4", StringWithLetters(letters, num))
}

func BenchmarkStringWithLetters_Len6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringWithLetters(letterBytes, 6)
	}
}
