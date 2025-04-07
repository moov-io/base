package telemetry_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/ccoveille/go-safecast"
	"github.com/moov-io/base/telemetry"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestStructAttributes_inputTypes(t *testing.T) {
	type s struct {
		Field string `otel:"field,omitempty"`
	}

	// nil pointer to struct doesn't cause panic
	var m *s
	require.NotPanics(t, func() {
		require.Empty(t, telemetry.StructAttributes(m))
	})

	// non-struct types don't cause panic
	require.NotPanics(t, func() {
		require.Empty(t, telemetry.StructAttributes("abc"))
	})

	// zero value struct ignored
	require.NotPanics(t, func() {
		require.Empty(t, telemetry.StructAttributes(s{}))
		require.Empty(t, telemetry.StructAttributes(&s{}))
	})

	// empty interfaces don't cause panic
	require.NotPanics(t, func() {
		var m interface{}
		require.Empty(t, telemetry.StructAttributes(m))
		require.Empty(t, telemetry.StructAttributes(&m))
	})
}

func TestStructAttributes_strings(t *testing.T) {
	str := "string"
	type foo struct {
		Field    string  `otel:"field"`
		FieldPtr *string `otel:"field_ptr,omitempty"`
	}

	// full model
	require.Equal(t,
		[]attribute.KeyValue{
			attribute.String("field", "string"),
			attribute.String("field_ptr", "string"),
		},
		telemetry.StructAttributes(foo{Field: "string", FieldPtr: &str}),
	)

	// omitempty (nil ptr)
	require.Equal(t,
		[]attribute.KeyValue{
			attribute.String("field", "string"),
		},
		telemetry.StructAttributes(foo{Field: "string"}),
	)

	// omitempty (ptr to empty string)
	emptyStr := ""
	require.Equal(t,
		[]attribute.KeyValue{
			attribute.String("field", "string"),
		},
		telemetry.StructAttributes(foo{
			Field:    "string",
			FieldPtr: &emptyStr,
		}),
	)

	// includes empty string
	got := telemetry.StructAttributes(foo{FieldPtr: &str})
	require.Equal(t,
		[]attribute.KeyValue{
			attribute.String("field", ""),
			attribute.String("field_ptr", "string"),
		},
		got,
	)
}

type testString struct {
	s string
}

func (t testString) Value() string {
	return t.s
}

type testInt struct {
	i int
}

func (t testInt) Value() int {
	return t.i
}

type testStrPtr struct {
	s string
}

func (t *testStrPtr) Value() string {
	return t.s
}

type testIntPtr struct {
	i int
}

func (t *testIntPtr) Value() int {
	return t.i
}

func TestStructAttributes_stringValuerintValuer(t *testing.T) {
	type test struct {
		String    testString  `otel:"custom_string"`
		StringPtr *testStrPtr `otel:"custom_string_ptr"`
		Str       *string     `otel:"str_ptr"`
		Number    testInt     `otel:"custom_number"`
		NumberPtr *testIntPtr `otel:"custom_number_ptr"`
	}

	sVal := "CustomValue"
	intVal := 123

	require.Equal(t,
		[]attribute.KeyValue{
			attribute.String("custom_string", sVal),
			attribute.String("custom_string_ptr", sVal),
			attribute.String("str_ptr", sVal),
			attribute.Int("custom_number", intVal),
			attribute.Int("custom_number_ptr", intVal),
		},
		telemetry.StructAttributes(test{
			String:    testString{sVal},
			StringPtr: &testStrPtr{sVal},
			Str:       &sVal,
			Number:    testInt{intVal},
			NumberPtr: &testIntPtr{intVal},
		}),
	)
}

func TestStructAttributes_int(t *testing.T) {
	type foo struct {
		Int          int   `otel:"int"`
		IntPtr       *int  `otel:"int_ptr"`
		Int8         int8  `otel:"int8"`
		Int16        int16 `otel:"int16"`
		Int32        int32 `otel:"int32"`
		Int64        int64 `otel:"int64"`
		OmitEmptyInt int   `otel:"omit_empty_int,omitempty"`
	}
	intVal := 4

	m := foo{
		Int:          math.MaxInt,
		IntPtr:       &intVal,
		Int8:         math.MinInt8,
		Int16:        math.MaxInt16,
		Int32:        math.MaxInt32,
		Int64:        math.MaxInt64,
		OmitEmptyInt: 0, // should be excluded
	}

	require.ElementsMatch(t,
		[]attribute.KeyValue{
			attribute.Int64("int", math.MaxInt),
			attribute.Int64("int_ptr", 4),
			attribute.Int64("int8", math.MinInt8),
			attribute.Int64("int16", math.MaxInt16),
			attribute.Int64("int32", math.MaxInt32),
			attribute.Int64("int64", math.MaxInt64),
		},
		telemetry.StructAttributes(m),
	)
}

func TestStructAttributes_uint(t *testing.T) {
	type foo struct {
		Uint          uint   `otel:"uint"`
		UintPtr       *uint  `otel:"uint_ptr"`
		Uint8         uint8  `otel:"uint8"`
		Uint16        uint16 `otel:"uint16"`
		Uint32        uint32 `otel:"uint32"`
		Uint64        uint64 `otel:"uint64"`
		OmitEmptyUint uint   `otel:"omit_empty_uint,omitempty"`
	}

	uintVal := uint(123)

	m := foo{
		Uint:          123,
		UintPtr:       &uintVal,
		Uint8:         123,
		Uint16:        123,
		Uint32:        123,
		Uint64:        123,
		OmitEmptyUint: 0, // should be excluded
	}

	require.ElementsMatch(t,
		[]attribute.KeyValue{
			attribute.Int64("uint", 123),
			attribute.Int64("uint_ptr", 123),
			attribute.Int64("uint8", 123),
			attribute.Int64("uint16", 123),
			attribute.Int64("uint32", 123),
			attribute.Int64("uint64", 123),
		},
		telemetry.StructAttributes(m),
	)
}

func TestStructAttributes_uint_overflow_to_string(t *testing.T) {
	type foo struct {
		Uint64 uint64 `otel:"uint64"`
	}

	uintVal := uint64(math.MaxInt64 + 1)
	_, err := safecast.ToInt64(uintVal)
	require.Error(t, err)

	m := foo{
		Uint64: uintVal,
	}

	require.ElementsMatch(t,
		[]attribute.KeyValue{
			attribute.String("uint64", fmt.Sprintf("%d", uintVal)),
		},
		telemetry.StructAttributes(m),
	)
}

func TestStructAttributes_float(t *testing.T) {
	type foo struct {
		Float32        float32  `otel:"float32"`
		Float64        float64  `otel:"float64"`
		FloatPtr       *float64 `otel:"float_ptr"`
		OmitEmptyFloat float32  `otel:"omit_empty_float,omitempty"`
	}

	floatVal := 123.45

	m := foo{
		Float32:        123.45,
		Float64:        -583.43,
		FloatPtr:       &floatVal,
		OmitEmptyFloat: 0,
	}

	require.ElementsMatch(t,
		[]attribute.KeyValue{
			attribute.Float64("float32", float64(m.Float32)),
			attribute.Float64("float64", m.Float64),
			attribute.Float64("float_ptr", floatVal),
		},
		telemetry.StructAttributes(m),
	)
}

func TestStructAttributes_complexNumbers(t *testing.T) {
	type foo struct {
		Complex64        complex64   `otel:"complex64"`
		Complex128       complex128  `otel:"complex128"`
		ComplexPtr       *complex128 `otel:"complex_ptr"`
		OmitEmptyComplex complex64   `otel:"omit_empty_complex,omitempty"`
	}

	complexVal := complex(4, 6)

	m := foo{
		Complex64:        complex64(complex(58.3, 12)),
		Complex128:       complex(2355.3, 3.2),
		ComplexPtr:       &complexVal,
		OmitEmptyComplex: 0,
	}

	require.ElementsMatch(t,
		[]attribute.KeyValue{
			attribute.Float64("complex64_real", float64(real(m.Complex64))),
			attribute.Float64("complex64_imag", float64(imag(m.Complex64))),
			attribute.Float64("complex128_real", real(m.Complex128)),
			attribute.Float64("complex128_imag", imag(m.Complex128)),
			attribute.Float64("complex_ptr_real", 4),
			attribute.Float64("complex_ptr_imag", 6),
		},
		telemetry.StructAttributes(m),
	)
}

func TestStructAttributes_sliceAndArray(t *testing.T) {
	type bar struct {
		Field string   `otel:"field,omitempty"`
		Slice []string `otel:"slice"`
	}
	type foo struct {
		Slice          []int     `otel:"slice"`
		SlicePtr       []*int    `otel:"slice_ptr"`
		StructSlice    []bar     `otel:"struct_slice"`
		StructPtrSlice []*bar    `otel:"struct_slice_ptr"`
		Array          [3]string `otel:"array"`
		OmitEmptySlice []string  `otel:"omit_empty_slice,omitempty"`
		OmitEmptyArray [3]string `otel:"omit_empty_array,omitempty"`
	}
	intVal := 438

	// full model
	m := foo{
		Slice:    []int{43, 82, -4},
		SlicePtr: []*int{&intVal, &intVal},
		StructSlice: []bar{
			{
				Field: "field",
				Slice: []string{"foo", "bar"},
			},
			{
				Field: "field1",
				Slice: []string{"foo1", "bar1"},
			},
		},
		StructPtrSlice: []*bar{
			{
				Field: "fieldPtr",
				Slice: []string{"fooPtr", "barPtr"},
			},
			{
				Field: "", // omitempty
				Slice: []string{"fooPtr1", "barPtr1"},
			},
		},
		Array:          [3]string{"1", "2", "3"},
		OmitEmptySlice: []string{},  // should be omitted
		OmitEmptyArray: [3]string{}, // should be omitted
	}
	require.Equal(t,
		[]attribute.KeyValue{
			// foo.Slice
			attribute.Int64("slice.0", 43),
			attribute.Int64("slice.1", 82),
			attribute.Int64("slice.2", -4),

			// foo.SlicePtr
			attribute.Int64("slice_ptr.0", int64(intVal)),
			attribute.Int64("slice_ptr.1", int64(intVal)),

			// foo.StructSlice
			attribute.String("struct_slice.0.field", "field"),
			attribute.String("struct_slice.0.slice.0", "foo"),
			attribute.String("struct_slice.0.slice.1", "bar"),
			attribute.String("struct_slice.1.field", "field1"),
			attribute.String("struct_slice.1.slice.0", "foo1"),
			attribute.String("struct_slice.1.slice.1", "bar1"),

			// foo.StructPtrSlice
			attribute.String("struct_slice_ptr.0.field", "fieldPtr"),
			attribute.String("struct_slice_ptr.0.slice.0", "fooPtr"),
			attribute.String("struct_slice_ptr.0.slice.1", "barPtr"),
			attribute.String("struct_slice_ptr.1.slice.0", "fooPtr1"),
			attribute.String("struct_slice_ptr.1.slice.1", "barPtr1"),

			// foo.Array
			attribute.String("array.0", "1"),
			attribute.String("array.1", "2"),
			attribute.String("array.2", "3"),
		},
		telemetry.StructAttributes(&m),
	)
}

func TestStructAttributes_maps(t *testing.T) {
	type foo struct {
		Map        map[string]string `otel:"map"`
		IgnoredMap map[string]string
	}
	m := foo{
		Map: map[string]string{
			"key1": "val1",
			"key2": "val2",
		},
		IgnoredMap: map[string]string{
			"ignoredKey1": "ignoredVal1",
			"ignoredKey2": "ignoredVal2",
		},
	}

	got := telemetry.StructAttributes(m)
	require.Len(t, got, 2)
	require.Contains(t, got, attribute.String("map.key1", "val1"))
	require.Contains(t, got, attribute.String("map.key2", "val2"))

	gotPtr := telemetry.StructAttributes(&m)
	require.ElementsMatch(t, got, gotPtr)

	// map of structs
	type s struct {
		Field        string `otel:"field,omitempty"`
		AnotherField string `otel:"another_field,omitempty"`
		IgnoredField string
		Struct       struct {
			StructField []int `otel:"struct_field,omitempty"`
		} `otel:"struct,omitempty"`
	}
	type fooTheSecond struct {
		Map map[int]s `otel:"map"`
	}
	m1 := fooTheSecond{
		Map: map[int]s{
			0: { // make sure attributes are created for map of structs
				Field:        "found me",
				AnotherField: "and me",
				IgnoredField: "not me",
			},
			1: { // make sure omitempty is still honored
				Field:        "",
				AnotherField: "",
				IgnoredField: "",
				Struct: struct {
					StructField []int `otel:"struct_field,omitempty"`
				}{[]int{4, 564, -4}},
			},
		},
	}

	got1 := telemetry.StructAttributes(m1)
	require.Len(t, got1, 5)
	require.Contains(t, got1, attribute.String("map.0.field", "found me"))
	require.Contains(t, got1, attribute.String("map.0.another_field", "and me"))
	require.Contains(t, got1, attribute.Int("map.1.struct.struct_field.0", 4))
	require.Contains(t, got1, attribute.Int("map.1.struct.struct_field.1", 564))
	require.Contains(t, got1, attribute.Int("map.1.struct.struct_field.2", -4))

	// no attributes for empty struct containing maps
	require.Empty(t, telemetry.StructAttributes(fooTheSecond{}))
}

func TestStructAttributes_bool(t *testing.T) {
	type foo struct {
		Bool          bool  `otel:"bool"`
		BoolPtr       *bool `otel:"bool_ptr"`
		OmitEmptyBool bool  `otel:"omit_empty_bool,omitempty"`
	}
	boolVal := false

	require.Equal(t,
		[]attribute.KeyValue{
			attribute.Bool("bool", false),
			attribute.Bool("bool_ptr", false),
		},
		telemetry.StructAttributes(foo{Bool: false, BoolPtr: &boolVal}),
	)
}

func TestStructAttributes_time(t *testing.T) {
	type foo struct {
		Time    time.Time  `otel:"time"`
		TimePtr *time.Time `otel:"time_ptr"`
	}

	// handles non-zero time and ptr to time
	current := time.Now()
	m := foo{
		Time:    current,
		TimePtr: &current,
	}
	got := telemetry.StructAttributes(m)
	require.ElementsMatch(t,
		[]attribute.KeyValue{
			attribute.String("time", current.Format(time.RFC3339)),
			attribute.String("time_ptr", current.Format(time.RFC3339)),
		},
		got,
	)
	// sanity check with ptr to struct
	require.Equal(t, got, telemetry.StructAttributes(&m))

	// handles zero value time and nil ptr
	require.Empty(t, telemetry.StructAttributes(foo{}))
	require.Empty(t, telemetry.StructAttributes(&foo{}))
}

func TestStructAttributes_fieldsWithMultipleTags(t *testing.T) {
	// make sure none of the tags interfere with each other
	type foo struct {
		Field string `json:"field" xml:"field" otel:"field"`
	}
	m := foo{
		Field: "value",
	}

	require.Equal(t,
		[]attribute.KeyValue{attribute.String("field", "value")},
		telemetry.StructAttributes(m),
	)

	// sanity check json and xml to ensure the tags all still work as expected
	js, err := json.Marshal(m)
	require.NoError(t, err)
	require.Equal(t, `{"field":"value"}`, string(js))

	x, err := xml.Marshal(m)
	require.NoError(t, err)
	require.Equal(t, `<foo><field>value</field></foo>`, string(x))
}

func TestStructAttributes_maxLengthHonored(t *testing.T) {
	type foo struct {
		Map  map[int]string `otel:"map"`
		List []int          `otel:"list"`
	}

	numElements := telemetry.MaxArrayAttributes + 5

	m := foo{
		Map:  make(map[int]string),
		List: make([]int, numElements),
	}
	for i := 0; i < numElements; i++ {
		m.Map[i] = "bar"
		m.List[i] = i
	}
	require.Len(t, m.Map, numElements)
	require.Len(t, m.List, numElements)

	require.Len(t, telemetry.StructAttributes(m), telemetry.MaxArrayAttributes*2) // include map+list
}

func TestStructAttributes_noAttributesForUntaggedFields(t *testing.T) {
	type foo struct {
		NoTag struct {
			Field string
		}
	}
	m := foo{
		NoTag: struct {
			Field string
		}{
			Field: "not included",
		},
	}
	require.Empty(t, telemetry.StructAttributes(m))
}

func TestStructAttributes_supportedDataTypes(t *testing.T) {
	type nested struct {
		NestedField complex128 `otel:"nested_field"`
		Nestception struct {
			NestceptionField string `otel:"nestception_field,omitempty"`
		} `otel:"nestception,omitempty"`
	}
	type Embedded struct {
		EmbeddedField string `otel:"embedded_field"`
	}

	type model struct {
		IgnoreMe        string             // no attribute should be added for this
		ignoreMe        string             `otel:"ignore_me"` // no attribute should be added for this
		BoolField       bool               `otel:"bool_field"`
		BoolPtrField    *bool              `otel:"bool_ptr_field"`
		IntField        int                `otel:"int_field"`
		EmptyIntField   int                `otel:"empty_int_field"`
		StrField        string             `otel:"str_field"`
		StrPtrField     *string            `otel:"str_ptr_field"`
		UintField       uint32             `otel:"uint_field"`
		MapStrInt       *map[string]int    `otel:"map_str_int"`
		MapUintPtr      map[uint32]*string `otel:"map_uint_ptr"`
		SliceField      []float64          `otel:"slice_field"`
		SlicePtrField   []*string          `otel:"slice_ptr_field"`
		FloatField      float64            `otel:"float_field"`
		AnonymousStruct struct {
			Field1  string `otel:"field_1"`
			Ignored string
		} `otel:"anonymous_struct"`
		Nested         nested `otel:"nested_struct"`
		*Embedded      `otel:"embedded_struct"`
		OmitMe         string      `otel:"omit_me,omitempty"`
		FuncField      func()      `otel:"func_field"`      // should be ignored even though it has an attribute tag
		EmptyInterface interface{} `otel:"empty_interface"` // should be ignored even though it has an attribute tag
	}

	boolVal := true
	strVal := "strPtrField"
	strVal1 := "strPtrField1"
	m := model{
		IgnoreMe:     "nothing to see here",
		ignoreMe:     "nothing to see here",
		BoolField:    true,
		BoolPtrField: &boolVal,
		StrField:     "strField",
		StrPtrField:  &strVal,
		IntField:     3,
		UintField:    uint32(53),
		MapStrInt: &map[string]int{
			"foo": 21,
			"bar": -5,
		},
		MapUintPtr: map[uint32]*string{
			uint32(43):     &strVal,
			uint32(584459): &strVal,
		},
		SliceField:    []float64{458.32, -5483.6482},
		SlicePtrField: []*string{&strVal, &strVal1},
		FloatField:    4.483,
		Nested: nested{
			NestedField: complex(3, 5),
		},
		Embedded: &Embedded{EmbeddedField: "embedded"},
		FuncField: func() {
			fmt.Println("hi")
		},
		EmptyInterface: "hi",
	}
	m.AnonymousStruct.Field1 = "foo"
	m.AnonymousStruct.Ignored = "foo"
	m.Nested.Nestception.NestceptionField = "all the nesting"

	wantAttrs := []attribute.KeyValue{
		attribute.Bool("bool_field", true),
		attribute.Bool("bool_ptr_field", true),
		attribute.Int("int_field", 3),
		attribute.Int("empty_int_field", 0),
		attribute.String("str_field", "strField"),
		attribute.String("str_ptr_field", "strPtrField"),
		attribute.Int64("uint_field", 53),
		attribute.Int64("map_str_int.foo", 21),
		attribute.Int64("map_str_int.bar", -5),
		attribute.String("map_uint_ptr.43", "strPtrField"),
		attribute.String("map_uint_ptr.584459", "strPtrField"),
		attribute.Float64("slice_field.0", 458.32),
		attribute.Float64("slice_field.1", -5483.6482),
		attribute.String("slice_ptr_field.0", "strPtrField"),
		attribute.String("slice_ptr_field.1", "strPtrField1"),
		attribute.Float64("float_field", 4.483),
		attribute.String("anonymous_struct.field_1", "foo"),
		attribute.Float64("nested_struct.nested_field_real", 3),
		attribute.Float64("nested_struct.nested_field_imag", 5),
		attribute.String("nested_struct.nestception.nestception_field", "all the nesting"),
		attribute.String("embedded_struct.embedded_field", "embedded"),
	}

	require.ElementsMatch(t, wantAttrs, telemetry.StructAttributes(m))
	require.ElementsMatch(t, wantAttrs, telemetry.StructAttributes(&m))
}

func TestStructAttributes_nestedStructs(t *testing.T) {
	type s3 struct {
		Field string `otel:"field"`
	}
	type s2 struct {
		Field        string `otel:"field"`
		S3Struct     s3     `otel:"s3"`
		OmitS3Struct s3     `otel:"empty_s3,omitempty"`
	}
	type s1 struct {
		Field    []*s3 `otel:"s3_slice"`
		S2Struct *s2   `otel:"s2"`
	}
	type s struct {
		S1Struct s1 `otel:"s1,omitempty"`
	}

	m := s{
		S1Struct: s1{
			Field: []*s3{
				{
					Field: "first entry",
				},
				{
					Field: "second entry",
				},
			},
			S2Struct: &s2{
				Field: "s2Field",
				S3Struct: s3{
					Field: "s3Field",
				},
				OmitS3Struct: s3{}, // should omit zero value
			},
		},
	}

	got := telemetry.StructAttributes(&m)

	require.Contains(t, got, attribute.String("s1.s3_slice.0.field", "first entry"))
	require.Contains(t, got, attribute.String("s1.s3_slice.1.field", "second entry"))
	require.Contains(t, got, attribute.String("s1.s2.field", "s2Field"))
	require.Contains(t, got, attribute.String("s1.s2.s3.field", "s3Field"))
	require.Len(t, got, 4)

	// no attributes for zero value struct containing nested structs
	require.Empty(t, telemetry.StructAttributes(s{}))
}

func TestStructAttributes_allowedMapKeys(t *testing.T) {
	type structKey struct {
		Field string `otel:"field"`
	}
	stringKey := "stringKey"
	var interfaceKey interface{} = "key"
	type s struct {
		// supported
		BoolMap          map[bool]string    `otel:"bool_map"`
		UintMap          map[uint32]string  `otel:"uint_map"`
		FloatMap         map[float64]string `otel:"float_map"`
		StringMap        map[string]string  `otel:"string_map"`
		PointerStringMap map[*string]string `otel:"pointer_string_map"`

		// not supported
		StructMap        map[structKey]string   `otel:"struct_map"`
		PointerStructMap map[*structKey]string  `otel:"pointer_struct_map"`
		InterfaceMap     map[interface{}]string `otel:"interface_map"`
	}

	got := telemetry.StructAttributes(s{
		BoolMap: map[bool]string{
			true: "true",
		},
		UintMap: map[uint32]string{
			8: "uint",
		},
		FloatMap: map[float64]string{
			4.95: "float",
		},
		StringMap: map[string]string{
			"string": "string",
		},
		PointerStringMap: map[*string]string{
			&stringKey: "string",
		},
		StructMap: map[structKey]string{
			{Field: "ignore_this_key"}: "ignore me",
		},
		PointerStructMap: map[*structKey]string{
			{Field: "ignore_this_key"}: "ignore me",
		},
		InterfaceMap: map[interface{}]string{
			interfaceKey: "string",
		},
	})
	require.Len(t, got, 5)
	require.Contains(t, got, attribute.String("bool_map.true", "true"))
	require.Contains(t, got, attribute.String("uint_map.8", "uint"))
	require.Contains(t, got, attribute.String("float_map.4.95", "float"))
	require.Contains(t, got, attribute.String("string_map.string", "string"))
	require.Contains(t, got, attribute.String("pointer_string_map.stringKey", "string"))
}
