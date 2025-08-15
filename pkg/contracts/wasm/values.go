package wasm

// WASMI32Value represents a 32-bit integer value
type WASMI32Value struct {
	value int32
}

// NewWASMI32Value creates a new 32-bit integer value
func NewWASMI32Value(value int32) *WASMI32Value {
	return &WASMI32Value{value: value}
}

// Type returns the type of this value
func (v *WASMI32Value) Type() WASMValueType {
	return WASMValueTypeI32
}

// Value returns the underlying value
func (v *WASMI32Value) Value() interface{} {
	return v.value
}

// Clone creates a deep copy of this value
func (v *WASMI32Value) Clone() WASMValue {
	return NewWASMI32Value(v.value)
}

// Int32 returns the value as int32
func (v *WASMI32Value) Int32() int32 {
	return v.value
}

// WASMI64Value represents a 64-bit integer value
type WASMI64Value struct {
	value int64
}

// NewWASMI64Value creates a new 64-bit integer value
func NewWASMI64Value(value int64) *WASMI64Value {
	return &WASMI64Value{value: value}
}

// Type returns the type of this value
func (v *WASMI64Value) Type() WASMValueType {
	return WASMValueTypeI64
}

// Value returns the underlying value
func (v *WASMI64Value) Value() interface{} {
	return v.value
}

// Clone creates a deep copy of this value
func (v *WASMI64Value) Clone() WASMValue {
	return NewWASMI64Value(v.value)
}

// Int64 returns the value as int64
func (v *WASMI64Value) Int64() int64 {
	return v.value
}

// WASMF32Value represents a 32-bit float value
type WASMF32Value struct {
	value float32
}

// NewWASMF32Value creates a new 32-bit float value
func NewWASMF32Value(value float32) *WASMF32Value {
	return &WASMF32Value{value: value}
}

// Type returns the type of this value
func (v *WASMF32Value) Type() WASMValueType {
	return WASMValueTypeF32
}

// Value returns the underlying value
func (v *WASMF32Value) Value() interface{} {
	return v.value
}

// Clone creates a deep copy of this value
func (v *WASMF32Value) Clone() WASMValue {
	return NewWASMF32Value(v.value)
}

// Float32 returns the value as float32
func (v *WASMF32Value) Float32() float32 {
	return v.value
}

// WASMF64Value represents a 64-bit float value
type WASMF64Value struct {
	value float64
}

// NewWASMF64Value creates a new 64-bit float value
func NewWASMF64Value(value float64) *WASMF64Value {
	return &WASMF64Value{value: value}
}

// Type returns the type of this value
func (v *WASMF64Value) Type() WASMValueType {
	return WASMValueTypeF64
}

// Value returns the underlying value
func (v *WASMF64Value) Value() interface{} {
	return v.value
}

// Clone creates a deep copy of this value
func (v *WASMF64Value) Clone() WASMValue {
	return NewWASMF64Value(v.value)
}

// Float64 returns the value as float64
func (v *WASMF64Value) Float64() float64 {
	return v.value
}

// Helper functions for creating values
func NewI32(value int32) WASMValue {
	return NewWASMI32Value(value)
}

func NewI64(value int64) WASMValue {
	return NewWASMI64Value(value)
}

func NewF32(value float32) WASMValue {
	return NewWASMF32Value(value)
}

func NewF64(value float64) WASMValue {
	return NewWASMF64Value(value)
}

// Type conversion helpers
func AsI32(value WASMValue) (int32, error) {
	if value.Type() != WASMValueTypeI32 {
		return 0, ErrGlobalTypeMismatch
	}
	return value.(*WASMI32Value).Int32(), nil
}

func AsI64(value WASMValue) (int64, error) {
	if value.Type() != WASMValueTypeI64 {
		return 0, ErrGlobalTypeMismatch
	}
	return value.(*WASMI64Value).Int64(), nil
}

func AsF32(value WASMValue) (float32, error) {
	if value.Type() != WASMValueTypeF32 {
		return 0, ErrGlobalTypeMismatch
	}
	return value.(*WASMF32Value).Float32(), nil
}

func AsF64(value WASMValue) (float64, error) {
	if value.Type() != WASMValueTypeF64 {
		return 0, ErrGlobalTypeMismatch
	}
	return value.(*WASMF64Value).Float64(), nil
}
