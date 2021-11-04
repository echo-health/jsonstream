package jsonstream

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"

	"github.com/pkg/errors"
)

func New(r io.Reader) *Decoder {
	return &Decoder{
		dec:      json.NewDecoder(r),
		matchers: map[string]func(key string) error{},
	}
}

type Decoder struct {
	dec      *json.Decoder
	matchers map[string]func(key string) error
}

func (d *Decoder) Decode() error {
	err := d.next("$")
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

	d.matchers[path] = func(key string) (err error) {
		defer func() {
			if r := recover(); r != nil {
				s, ok := r.(string)
				if ok {
					err = errors.New(s)
					return
				}
				e, ok := r.(error)
				if ok {
					err = e
					return
				}

				panic(r)
			}
		}()

		inType := hndlr.In(1)

		// // if the receiving type is a pointer get it's underling type
		// if inType.Kind() == reflect.Ptr {
		// 	inType = inType.Elem()
		// }

		obj := reflect.New(inType).Interface()

		err = d.dec.Decode(obj)
		if err != nil {
			return err
		}

		// if the receiving type is NOT a pointer then deference to value
		// if hndlr.In(1).Kind() != reflect.Ptr {
		obj = reflect.ValueOf(obj).Elem().Interface()
		// }

		values := []reflect.Value{
			reflect.ValueOf(key),
			reflect.ValueOf(obj),
		}

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

var pathRegexp = regexp.MustCompile(`([\w\*]+)(\[[\w\*]+\])?(\.|$)`)

func pathMatch(path string, filter string) bool {
	pathParts := pathRegexp.FindAllStringSubmatch(path, -1)
	filterParts := pathRegexp.FindAllStringSubmatch(filter, -1)

	if len(pathParts) != len(filterParts) {
		return false
	}

	for i, part := range pathParts {
		key := part[1]
		arr := part[2]

		filterPart := filterParts[i]
		filterKey := filterPart[1]
		filterArr := filterPart[2]

		// if key doesn't match return
		if key != filterKey && filterKey != "*" {
			return false
		}

		// if the path has an array part but the filter doesn't (or vice-versa) return
		if (arr != "") != (filterArr != "") {
			return false
		}

		// if no array part continue
		if arr == "" {
			continue
		}

		// if array parts don't match return
		if arr != filterArr && filterArr != "[*]" {
			return false
		}
	}

	return true
}

func (d *Decoder) next(path string) error {
	// See if we have a match for where we are
	for p, h := range d.matchers {
		if pathMatch(path, p) {
			return h(path)
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
			return d.next(path + ".")
		case '[':
			// process each entry in the array
			index := 0
			for d.dec.More() {
				err = d.next(path + fmt.Sprintf("[%d]", index))
				if err != nil {
					return err
				}
				index++
			}
			// carry on parsing
			return d.next(path)
		case '}':
			return nil
		case ']':
			return nil
		default:
			return fmt.Errorf("unexpected delimiter: %s", t)
		}
	}

	// handle an object key or value
	if len(path) > 0 && path[len(path)-1] == '.' {
		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("not a string")
		}

		// go to value (could be a nested object or array)
		err = d.next(path + key)
		if err != nil {
			return err
		}

		// go to next key (or possible end of object)
		return d.next(path)
	}

	// handled (primitive) value of object or array
	return nil
}
