package httpx

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// ToQueryParams reflects on obj and returns a URL-encoded query string,
// mirroring sdk-ts toQueryParams behavior:
//
//   - nil and zero-pointer fields are skipped
//   - primitives -> "key=value"; bools as "true"/"false"
//   - slices/arrays -> repeated "key=v1&key=v2"
//   - nested structs and nested maps with non-primitive values are skipped
//   - struct field name comes from `query:"name"` tag, then `json:"name"`,
//     then lowerCamelCase of the Go field name
//   - `query:"-"` skips the field
//   - `query:",flatten"` (or any tag containing the "flatten" option) on a
//     map[string]any field flattens its primitive/slice entries at the top
//     level; nested values inside the map are skipped
//   - omitempty (on either query: or json: tag) skips zero values
//
// obj may be a struct, a pointer to a struct, or a map[string]any. nil input
// returns "".
func ToQueryParams(obj any) string {
	if obj == nil {
		return ""
	}
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	values := url.Values{}
	switch v.Kind() {
	case reflect.Struct:
		encodeStruct(v, values)
	case reflect.Map:
		encodeTopLevelMap(v, values)
	default:
		return ""
	}
	return values.Encode()
}

func encodeStruct(v reflect.Value, values url.Values) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fv := v.Field(i)
		// Embedded structs: recurse so their fields appear at the top level.
		if field.Anonymous && fv.Kind() == reflect.Struct {
			encodeStruct(fv, values)
			continue
		}
		name, opts := fieldName(field)
		if name == "-" {
			continue
		}
		if opts.flatten {
			if fv.Kind() == reflect.Map {
				encodeTopLevelMap(fv, values)
			}
			continue
		}
		if opts.omitempty && isZero(fv) {
			continue
		}
		appendValue(values, name, fv)
	}
}

// encodeTopLevelMap flattens primitives and slices from a map[string]any (or
// any string-keyed map) into the top-level query params.
func encodeTopLevelMap(v reflect.Value, values url.Values) {
	if v.Kind() != reflect.Map {
		return
	}
	if v.Type().Key().Kind() != reflect.String {
		return
	}
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		val := iter.Value()
		// Unwrap interfaces so we can inspect the underlying kind.
		for val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer {
			if val.IsNil() {
				val = reflect.Value{}
				break
			}
			val = val.Elem()
		}
		if !val.IsValid() {
			continue
		}
		appendValue(values, key, val)
	}
}

// appendValue writes a primitive or repeated slice/array value into values.
// Nested structs and maps are silently skipped (TS parity).
func appendValue(values url.Values, key string, v reflect.Value) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		values.Add(key, v.String())
	case reflect.Bool:
		values.Add(key, strconv.FormatBool(v.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		values.Add(key, strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		values.Add(key, strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		values.Add(key, strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			for elem.Kind() == reflect.Interface || elem.Kind() == reflect.Pointer {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}
			if isPrimitive(elem.Kind()) {
				values.Add(key, primitiveString(elem))
			}
		}
	default:
		// Nested struct / map / chan / func: skip.
	}
}

func isPrimitive(k reflect.Kind) bool {
	switch k {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func primitiveString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return ""
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	return v.IsZero()
}

type tagOpts struct {
	omitempty bool
	flatten   bool
}

// fieldName resolves the query-string name for a struct field. Returns "-"
// for fields that should be skipped.
func fieldName(field reflect.StructField) (string, tagOpts) {
	var opts tagOpts
	if tag, ok := field.Tag.Lookup("query"); ok {
		name, rest := splitTag(tag)
		opts = parseOpts(rest)
		if name == "-" {
			return "-", opts
		}
		if name != "" {
			return name, opts
		}
		// query:",flatten" or query:",omitempty" — fall through to json/name fallback
	}
	if tag, ok := field.Tag.Lookup("json"); ok {
		name, rest := splitTag(tag)
		// json's omitempty contributes to opts as well.
		if jsonOpts := parseOpts(rest); jsonOpts.omitempty {
			opts.omitempty = true
		}
		if name == "-" {
			return "-", opts
		}
		if name != "" {
			return name, opts
		}
	}
	return lowerCamel(field.Name), opts
}

func splitTag(tag string) (string, string) {
	if idx := strings.IndexByte(tag, ','); idx >= 0 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

func parseOpts(rest string) tagOpts {
	var opts tagOpts
	if rest == "" {
		return opts
	}
	for _, part := range strings.Split(rest, ",") {
		switch strings.TrimSpace(part) {
		case "omitempty":
			opts.omitempty = true
		case "flatten":
			opts.flatten = true
		}
	}
	return opts
}

func lowerCamel(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
