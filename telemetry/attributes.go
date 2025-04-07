package telemetry

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

const DropSpanKey = "span.drop"

// DropSpan informs the sampler to skip this event if theirs no links tied to it.
func DropSpan() attribute.KeyValue {
	return attribute.Bool(DropSpanKey, true)
}

const MoovKnownIssueKey = "moov.known_issue"

// AttributeMoovKnownIssue is an attribute to mark a trace as a previously observed issue.
// IMPORTANT: if a trace has this attribute it will NOT fire a critical PD alert defined in
// https://github.com/moovfinancial/infra/blob/master/terraform-modules/apps/go-service/honeycomb.tf#L42
func AttributeMoovKnownIssue() attribute.KeyValue {
	return attribute.Bool(MoovKnownIssueKey, true)
}

// StructAttributes creates an attribute.KeyValue for each field in the struct that has an "otel" tag defined.
// Nested structs will also be included, with attribute names formatted as "parent_attribute.nested_field_attribute".
func StructAttributes(s interface{}) (kv []attribute.KeyValue) {
	rVal := reflect.ValueOf(s)
	if !rVal.IsValid() {
		return kv // ignore values that can't be handled by reflection
	}

	return structAttributes(rVal, "") // no prefix for top level
}

func structAttributes(rVal reflect.Value, prefix string) (kv []attribute.KeyValue) {
	defer func() { // recover from panics
		if recovered := recover(); recovered != nil {
			return
		}
	}()

	// if rVal is an interface or pointer, get the underlying value
	if rVal.Kind() == reflect.Interface || rVal.Kind() == reflect.Pointer {
		if rVal.IsNil() {
			return kv
		}
		// get the underlying value from this interface or pointer
		rVal = rVal.Elem()
	}

	if rVal.Kind() != reflect.Struct {
		return kv // only structs have tags to parse
	}

	for i := 0; i < rVal.NumField(); i++ {
		field := rVal.Type().Field(i)
		// skip non-exported fields
		if !field.IsExported() && !field.Anonymous {
			continue // exported fields only
		}

		otelTag := parseTag(prefix, field.Tag)
		if otelTag == nil {
			continue
		}

		kv = append(kv, createAttributes(rVal.Field(i), otelTag)...)
	}

	return kv
}

type otelTag struct {
	attributeName string
	omitEmpty     bool
}

func parseTag(prefix string, stag reflect.StructTag) *otelTag {
	tagParts := strings.Split(stag.Get(AttributeTag), ",")
	attributeName := tagParts[0] // strings.Split is guaranteed to always return at least 1 element
	if attributeName == "" {
		return nil
	}

	if prefix != "" {
		attributeName = fmt.Sprintf("%s.%s", prefix, attributeName)
	}

	omitEmpty := false
	if len(tagParts) > 1 && tagParts[1] == "omitempty" {
		omitEmpty = true
	}

	return &otelTag{
		attributeName: attributeName,
		omitEmpty:     omitEmpty,
	}
}

type stringValuer interface {
	Value() string
}

type intValuer interface {
	Value() int
}

func createAttributes(val reflect.Value, tag *otelTag) (kv []attribute.KeyValue) {
	if tag.omitEmpty && val.IsZero() {
		return kv
	}

	switch val.Kind() {
	case reflect.String:
		kv = append(kv, attribute.String(tag.attributeName, val.String()))
	case reflect.Bool:
		kv = append(kv, attribute.Bool(tag.attributeName, val.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		kv = append(kv, attribute.Int64(tag.attributeName, val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		kv = append(kv, attribute.Int64(tag.attributeName, int64(val.Uint()))) //nolint:gosec
	case reflect.Float32, reflect.Float64:
		kv = append(kv, attribute.Float64(tag.attributeName, val.Float()))
	case reflect.Complex64, reflect.Complex128:
		complexVal := val.Complex()
		kv = append(kv,
			attribute.Float64(tag.attributeName+"_real", real(complexVal)),
			attribute.Float64(tag.attributeName+"_imag", imag(complexVal)),
		)
	case reflect.Map:
		// check map key type
		if val.Len() == 0 || len(val.MapKeys()) == 0 {
			break
		}
		keyKind := val.MapKeys()[0].Kind()
		if keyKind == reflect.Pointer {
			keyKind = val.MapKeys()[0].Elem().Kind()
		}

		if !supportedMapKeyKinds[keyKind] {
			break
		}

		mapIter := val.MapRange()
		count := 0
		for mapIter.Next() {
			// only support simple map key types
			if count == MaxArrayAttributes {
				break
			}

			key := mapIter.Key()
			if key.Kind() == reflect.Pointer {
				key = key.Elem()
			}

			kv = append(kv, createAttributes(mapIter.Value(), &otelTag{
				attributeName: fmt.Sprintf("%s.%v", tag.attributeName, key.Interface()),
				omitEmpty:     false,
			})...)
			count++
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			if i == MaxArrayAttributes {
				break
			}

			kv = append(kv, createAttributes(val.Index(i), &otelTag{
				attributeName: fmt.Sprintf("%s.%d", tag.attributeName, i),
				omitEmpty:     false,
			})...)
		}
	case reflect.Struct:
		// if this is a non-zero time.Time, format as string and append to attributes
		if t, ok := val.Interface().(time.Time); ok && !t.IsZero() {
			kv = append(kv, attribute.String(tag.attributeName, t.Format(time.RFC3339)))
		} else if t, ok := val.Interface().(stringValuer); ok {
			stringValue := t.Value()
			kv = append(kv, attribute.String(tag.attributeName, stringValue))
		} else if t, ok := val.Interface().(intValuer); ok {
			intValue := t.Value()
			kv = append(kv, attribute.Int(tag.attributeName, intValue))
		} else { // otherwise recursively handle the struct
			kv = append(kv, structAttributes(val, tag.attributeName)...)
		}
	case reflect.Pointer:
		// before we check the value element, see if the pointer implements any of
		// our known interfaces
		if t, ok := val.Interface().(stringValuer); ok {
			stringValue := t.Value()
			kv = append(kv, attribute.String(tag.attributeName, stringValue))
		} else if t, ok := val.Interface().(intValuer); ok {
			intValue := t.Value()
			kv = append(kv, attribute.Int(tag.attributeName, intValue))
		} else {
			kv = append(kv, createAttributes(val.Elem(), tag)...)
		}
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Uintptr,
		reflect.UnsafePointer, reflect.Invalid:
		return // not supported
	}

	return kv
}

// supportedMapKeyKinds defines the map key types that are supported for attribute names.
// Only types that can be easily represented as strings are allowed to be used in attribute names.
var supportedMapKeyKinds = map[reflect.Kind]bool{
	reflect.Bool:    true,
	reflect.Int:     true,
	reflect.Int8:    true,
	reflect.Int16:   true,
	reflect.Int32:   true,
	reflect.Int64:   true,
	reflect.Uint:    true,
	reflect.Uint8:   true,
	reflect.Uint16:  true,
	reflect.Uint32:  true,
	reflect.Uint64:  true,
	reflect.Float32: true,
	reflect.Float64: true,
	reflect.String:  true,
}
