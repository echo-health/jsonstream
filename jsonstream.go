package jsonstream

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func New(r io.Reader) *Decoder {
	return &Decoder{
		dec:      json.NewDecoder(r),
		matchers: map[string]func(key string, token json.Token) error{},
	}
}

type Decoder struct {
	dec      *json.Decoder
	matchers map[string]func(key string, token json.Token) error
}

func (d *Decoder) Decode() error {
	err := d.next([]string{"!"})
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (d *Decoder) On(path string, handler interface{}) error {
	hndlr := reflect.TypeOf(handler)
	if hndlr.Kind() != reflect.Func {
		return errors.New("handler needs to be a func")
	}

	if hndlr.NumIn() != 2 {
		return errors.New("handler needs to be a func with signature: func(key string, value Any) error {} - wrong number of arguments")
	}

	if hndlr.In(0) != reflect.TypeOf((*string)(nil)).Elem() {
		return errors.New("handler needs to be a func with signature: func(key string, value Any) error {} - first argument not string")
	}

	if !hndlr.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("handler needs to be a func with signature: func(key string, value Any) error {} - return type not an error")
	}

	fn := reflect.ValueOf(handler)

	d.matchers[path] = func(key string, t json.Token) error {
		values := []reflect.Value{
			reflect.ValueOf(key),
		}

		inType := hndlr.In(1)
		if inType.Kind() == reflect.Ptr {
			inType = inType.Elem()
		}

		obj := reflect.New(inType).Interface()

		err := d.dec.Decode(obj)
		if err != nil {
			return err
		}

		if inType.Kind() != reflect.Ptr {
			obj = reflect.ValueOf(obj).Elem().Interface()
		}

		values = append(values, reflect.ValueOf(obj))

		rtrn := fn.Call(values)
		if len(rtrn) == 0 {
			return nil
		}

		erri := rtrn[0].Interface()
		if erri != nil {
			err = erri.(error)
		}

		return err
	}

	return nil
}

func (d *Decoder) next(basePath []string) error {
	// See if we have a match for where we are
	for p, h := range d.matchers {
		if p == strings.Join(basePath, "") {
			return h(strings.Join(basePath, ""), nil)
		}
	}

	t, err := d.dec.Token()
	if err != nil {
		return err
	}

	// handle delimiter
	delim, ok := t.(json.Delim)
	if ok {
		switch delim {
		case '{':
			// start processing the object
			return d.next(append(basePath[:], "."))
		case '[':
			// process each entry in the array
			index := 0
			for d.dec.More() {
				err = d.next(append(basePath[:], fmt.Sprintf("[%d]", index)))
				if err != nil {
					return err
				}
				index++
			}
			// carry on parsing
			return d.next(basePath)
		case '}':
			return nil
		case ']':
			return nil
		default:
			return fmt.Errorf("unexpected delimiter: %s", t)
		}
	}

	// handle an object key or value
	if len(basePath) > 0 && basePath[len(basePath)-1] == "." {
		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("not a string")
		}

		// go to value (could be a nested object or array)
		err = d.next(append(basePath[:], key))
		if err != nil {
			return err
		}

		// go to next key (or possible end of object)
		return d.next(basePath)
	}

	// handled (primitive) value of object or array
	return nil
}
