package jsonstream_test

import (
	"strings"
	"testing"

	"github.com/echo-health/jsonstream"
	"github.com/stretchr/testify/assert"
)

type Counter struct {
	C int
}

func TestReflectionForPathMatchers(t *testing.T) {

	type fixture struct {
		Name            string
		JSON            string
		Path            string
		ExpectedAsserts int
		Handler         func(t *testing.T, c *Counter) interface{}
	}

	fixtures := []fixture{
		{
			Name:            "int",
			JSON:            `{"key": 1}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value int) error {
					c.C++
					assert.Equal(t, 1, value)
					return nil
				}
			},
		},
		{
			Name:            "int with null",
			JSON:            `{"key": null}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value int) error {
					c.C++
					assert.Equal(t, 0, value)
					return nil
				}
			},
		},
		{
			Name:            "*int with null",
			JSON:            `{"key": null}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value *int) error {
					c.C++
					assert.Nil(t, value)
					return nil
				}
			},
		},
		{
			Name:            "*int with value",
			JSON:            `{"key": 5}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value *int) error {
					c.C++
					assert.Equal(t, 5, *value)
					return nil
				}
			},
		},
		{
			Name:            "string",
			JSON:            `{"key": "foo"}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value string) error {
					c.C++
					assert.Equal(t, "foo", value)
					return nil
				}
			},
		},
		{
			Name:            "string with null",
			JSON:            `{"key": null}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value string) error {
					c.C++
					assert.Equal(t, "", value)
					return nil
				}
			},
		},
		{
			Name:            "*string with null",
			JSON:            `{"key": null}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value *string) error {
					c.C++
					assert.Nil(t, value)
					return nil
				}
			},
		},
		{
			Name:            "*string with value",
			JSON:            `{"key": "foo"}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value *string) error {
					c.C++
					assert.EqualValues(t, "foo", *value)
					return nil
				}
			},
		},
		{
			Name:            "float32",
			JSON:            `{"key": 5.5}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value float64) error {
					c.C++
					assert.Equal(t, 5.5, value)
					return nil
				}
			},
		},
		{
			Name:            "float32 with null",
			JSON:            `{"key": null}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value float64) error {
					c.C++
					assert.Equal(t, float64(0), value)
					return nil
				}
			},
		},
		{
			Name:            "*float32 with null",
			JSON:            `{"key": null}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value *float32) error {
					c.C++
					assert.Nil(t, value)
					return nil
				}
			},
		},
		{
			Name:            "*float32 with value",
			JSON:            `{"key": 5.5}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value *float64) error {
					c.C++
					assert.EqualValues(t, 5.5, *value)
					return nil
				}
			},
		},
		{
			Name:            "struct",
			JSON:            `{"key": {"something": "else"}}`,
			Path:            "$.key",
			ExpectedAsserts: 1,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value struct{ Something string }) error {
					c.C++
					assert.EqualValues(t, "else", value.Something)
					return nil
				}
			},
		},
		{
			Name:            "each array element",
			JSON:            `{"key": {"nested": [1,2,3]}}`,
			Path:            "$.key.nested[*]",
			ExpectedAsserts: 3,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value int) error {
					c.C++
					switch key {
					case "$.key.nested[0]":
						assert.Equal(t, 1, value)
					case "$.key.nested[1]":
						assert.Equal(t, 2, value)
					case "$.key.nested[2]":
						assert.Equal(t, 3, value)
					default:
						assert.Fail(t, "unexpected key", "key was %s", key)
					}
					return nil
				}
			},
		},
		{
			Name:            "each array element struct",
			JSON:            `{"key": {"nested": [{"number": 101}, {"number": 202}]}}`,
			Path:            "$.key.nested[*]",
			ExpectedAsserts: 2,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value struct{ Number int }) error {
					c.C++
					switch key {
					case "$.key.nested[0]":
						assert.Equal(t, 101, value.Number)
					case "$.key.nested[1]":
						assert.Equal(t, 202, value.Number)
					default:
						assert.Fail(t, "unexpected key", "key was %s", key)
					}
					return nil
				}
			},
		},
		{
			Name:            "each array element map",
			JSON:            `{"key": {"nested": [{"foo": "bar"}, {"baz": "bop"}]}}`,
			Path:            "$.key.nested[*]",
			ExpectedAsserts: 2,
			Handler: func(t *testing.T, c *Counter) interface{} {
				return func(key string, value map[string]string) error {
					c.C++
					switch key {
					case "$.key.nested[0]":
						assert.Equal(t, "bar", value["foo"])
					case "$.key.nested[1]":
						assert.Equal(t, "bop", value["baz"])
					default:
						assert.Fail(t, "unexpected key", "key was %s", key)
					}
					return nil
				}
			},
		},
	}

	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			streaming := jsonstream.New(strings.NewReader(f.JSON))
			c := &Counter{}
			streaming.On(f.Path, f.Handler(t, c))
			assert.NoError(t, streaming.Decode())
			assert.Equal(t, f.ExpectedAsserts, c.C)
		})
	}
}
