package jsonstream_test

import (
	"strings"
	"testing"

	"github.com/echo-health/jsonstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReflectionForPathMatchers(t *testing.T) {
	jsonStream := strings.NewReader(`
		{
			"one": "foo",
			"two": [1,2,3],
			"three": true,
			"four": [
				{
					"inner": "first"
				},
				{
					"inner": "second"
				}
			]
		}
	`)

	type nested struct {
		Inner string
	}

	streaming := jsonstream.New(jsonStream)

	var oneValue string
	var twoValue []int
	var threeValue bool
	var fourValue []*nested

	streaming.On("!.one", func(key string, value string) error {
		oneValue = value
		return nil
	})

	streaming.On("!.two", func(key string, value []int) error {
		twoValue = value
		return nil
	})

	streaming.On("!.three", func(key string, value bool) error {
		threeValue = value
		return nil
	})

	streaming.On("!.four", func(key string, value []*nested) error {
		fourValue = value
		return nil
	})

	err := streaming.Decode()
	require.NoError(t, err)

	assert.Equal(t, "foo", oneValue)
	assert.Equal(t, []int{1, 2, 3}, twoValue)
	assert.Equal(t, true, threeValue)
	assert.Equal(t, "first", fourValue[0].Inner)
	assert.Equal(t, "second", fourValue[1].Inner)
}
