package wasm

import (
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// TestWASMValueTypeConstants tests WASM value type constants
func TestWASMValueTypeConstants(t *testing.T) {
	if WASMValueTypeI32 != 0 {
		t.Errorf("expected WASMValueTypeI32 to be 0, got %d", WASMValueTypeI32)
	}

	if WASMValueTypeI64 != 1 {
		t.Errorf("expected WASMValueTypeI64 to be 1, got %d", WASMValueTypeI64)
	}

	if WASMValueTypeF32 != 2 {
		t.Errorf("expected WASMValueTypeF32 to be 2, got %d", WASMValueTypeF32)
	}

	if WASMValueTypeF64 != 3 {
		t.Errorf("expected WASMValueTypeF64 to be 3, got %d", WASMValueTypeF64)
	}
}

// TestWASMExportKindConstants tests WASM export kind constants
func TestWASMExportKindConstants(t *testing.T) {
	if WASMExportKindFunction != 0 {
		t.Errorf("expected WASMExportKindFunction to be 0, got %d", WASMExportKindFunction)
	}

	if WASMExportKindTable != 1 {
		t.Errorf("expected WASMExportKindTable to be 1, got %d", WASMExportKindTable)
	}

	if WASMExportKindMemory != 2 {
		t.Errorf("expected WASMExportKindMemory to be 2, got %d", WASMExportKindMemory)
	}

	if WASMExportKindGlobal != 3 {
		t.Errorf("expected WASMExportKindGlobal to be 3, got %d", WASMExportKindGlobal)
	}
}

// TestWASMI32Value tests I32 value functionality
func TestWASMI32Value(t *testing.T) {
	// Test value creation
	value := int32(42)
	i32Val := NewWASMI32Value(value)

	if i32Val == nil {
		t.Fatal("expected non-nil I32 value")
	}

	// Test Type method
	if i32Val.Type() != WASMValueTypeI32 {
		t.Errorf("expected I32 value type, got %d", i32Val.Type())
	}

	// Test Value method
	retrievedValue := i32Val.Value()
	if retrievedValue != value {
		t.Errorf("expected value %d, got %v", value, retrievedValue)
	}

	// Test Int32 method
	intValue := i32Val.Int32()
	if intValue != value {
		t.Errorf("expected Int32 value %d, got %d", value, intValue)
	}

	// Test Clone method
	clonedValue := i32Val.Clone()
	if clonedValue == nil {
		t.Fatal("expected non-nil cloned value")
	}

	if clonedValue.Type() != WASMValueTypeI32 {
		t.Errorf("expected cloned value to have I32 type, got %d", clonedValue.Type())
	}

	clonedIntValue := clonedValue.(*WASMI32Value).Int32()
	if clonedIntValue != value {
		t.Errorf("expected cloned value %d, got %d", value, clonedIntValue)
	}

	// Test that cloned value is independent
	clonedValue.(*WASMI32Value).value = 100
	if i32Val.Int32() == 100 {
		t.Error("expected original value to be unaffected by cloned value changes")
	}
}

// TestWASMI64Value tests I64 value functionality
func TestWASMI64Value(t *testing.T) {
	// Test value creation
	value := int64(123456789)
	i64Val := NewWASMI64Value(value)

	if i64Val == nil {
		t.Fatal("expected non-nil I64 value")
	}

	// Test Type method
	if i64Val.Type() != WASMValueTypeI64 {
		t.Errorf("expected I64 value type, got %d", i64Val.Type())
	}

	// Test Value method
	retrievedValue := i64Val.Value()
	if retrievedValue != value {
		t.Errorf("expected value %d, got %v", value, retrievedValue)
	}

	// Test Int64 method
	intValue := i64Val.Int64()
	if intValue != value {
		t.Errorf("expected Int64 value %d, got %d", value, intValue)
	}

	// Test Clone method
	clonedValue := i64Val.Clone()
	if clonedValue == nil {
		t.Fatal("expected non-nil cloned value")
	}

	if clonedValue.Type() != WASMValueTypeI64 {
		t.Errorf("expected cloned value to have I64 type, got %d", clonedValue.Type())
	}

	clonedIntValue := clonedValue.(*WASMI64Value).Int64()
	if clonedIntValue != value {
		t.Errorf("expected cloned value %d, got %d", value, clonedIntValue)
	}

	// Test that cloned value is independent
	clonedValue.(*WASMI64Value).value = 999999
	if i64Val.Int64() == 999999 {
		t.Error("expected original value to be unaffected by cloned value changes")
	}
}

// TestWASMF32Value tests F32 value functionality
func TestWASMF32Value(t *testing.T) {
	// Test value creation
	value := float32(3.14159)
	f32Val := NewWASMF32Value(value)

	if f32Val == nil {
		t.Fatal("expected non-nil F32 value")
	}

	// Test Type method
	if f32Val.Type() != WASMValueTypeF32 {
		t.Errorf("expected F32 value type, got %d", f32Val.Type())
	}

	// Test Value method
	retrievedValue := f32Val.Value()
	if retrievedValue != value {
		t.Errorf("expected value %f, got %v", value, retrievedValue)
	}

	// Test Float32 method
	floatValue := f32Val.Float32()
	if floatValue != value {
		t.Errorf("expected Float32 value %f, got %f", value, floatValue)
	}

	// Test Clone method
	clonedValue := f32Val.Clone()
	if clonedValue == nil {
		t.Fatal("expected non-nil cloned value")
	}

	if clonedValue.Type() != WASMValueTypeF32 {
		t.Errorf("expected cloned value to have F32 type, got %d", clonedValue.Type())
	}

	clonedFloatValue := clonedValue.(*WASMF32Value).Float32()
	if clonedFloatValue != value {
		t.Errorf("expected cloned value %f, got %f", value, clonedFloatValue)
	}

	// Test that cloned value is independent
	clonedValue.(*WASMF32Value).value = 2.718
	if f32Val.Float32() == 2.718 {
		t.Error("expected original value to be unaffected by cloned value changes")
	}
}

// TestWASMF64Value tests F64 value functionality
func TestWASMF64Value(t *testing.T) {
	// Test value creation
	value := float64(2.718281828)
	f64Val := NewWASMF64Value(value)

	if f64Val == nil {
		t.Fatal("expected non-nil F64 value")
	}

	// Test Type method
	if f64Val.Type() != WASMValueTypeF64 {
		t.Errorf("expected F64 value type, got %d", f64Val.Type())
	}

	// Test Value method
	retrievedValue := f64Val.Value()
	if retrievedValue != value {
		t.Errorf("expected value %f, got %v", value, retrievedValue)
	}

	// Test Float64 method
	floatValue := f64Val.Float64()
	if floatValue != value {
		t.Errorf("expected Float64 value %f, got %f", value, floatValue)
	}

	// Test Clone method
	clonedValue := f64Val.Clone()
	if clonedValue == nil {
		t.Fatal("expected non-nil cloned value")
	}

	if clonedValue.Type() != WASMValueTypeF64 {
		t.Errorf("expected cloned value to have F64 type, got %d", clonedValue.Type())
	}

	clonedFloatValue := clonedValue.(*WASMF64Value).Float64()
	if clonedFloatValue != value {
		t.Errorf("expected cloned value %f, got %f", value, clonedFloatValue)
	}

	// Test that cloned value is independent
	clonedValue.(*WASMF64Value).value = 1.414
	if f64Val.Float64() == 1.414 {
		t.Error("expected original value to be unaffected by cloned value changes")
	}
}

// TestWASMValueHelperFunctions tests helper functions for creating values
func TestWASMValueHelperFunctions(t *testing.T) {
	// Test NewI32
	i32Val := NewI32(42)
	if i32Val == nil {
		t.Fatal("expected non-nil I32 value from NewI32")
	}
	if i32Val.Type() != WASMValueTypeI32 {
		t.Errorf("expected I32 type from NewI32, got %d", i32Val.Type())
	}

	// Test NewI64
	i64Val := NewI64(123456789)
	if i64Val == nil {
		t.Fatal("expected non-nil I64 value from NewI64")
	}
	if i64Val.Type() != WASMValueTypeI64 {
		t.Errorf("expected I64 type from NewI64, got %d", i64Val.Type())
	}

	// Test NewF32
	f32Val := NewF32(3.14159)
	if f32Val == nil {
		t.Fatal("expected non-nil F32 value from NewF32")
	}
	if f32Val.Type() != WASMValueTypeF32 {
		t.Errorf("expected F32 type from NewF32, got %d", f32Val.Type())
	}

	// Test NewF64
	f64Val := NewF64(2.718281828)
	if f64Val == nil {
		t.Fatal("expected non-nil F64 value from NewF64")
	}
	if f64Val.Type() != WASMValueTypeF64 {
		t.Errorf("expected F64 type from NewF64, got %d", f64Val.Type())
	}
}

// TestWASMValueTypeConversions tests type conversion helper functions
func TestWASMValueTypeConversions(t *testing.T) {
	// Test AsI32 with I32 value
	i32Val := NewI32(42)
	convertedI32, err := AsI32(i32Val)
	if err != nil {
		t.Fatalf("unexpected error converting I32: %v", err)
	}
	if convertedI32 != 42 {
		t.Errorf("expected converted I32 value 42, got %d", convertedI32)
	}

	// Test AsI32 with wrong type (should fail)
	i64Val := NewI64(123)
	_, err = AsI32(i64Val)
	if err == nil {
		t.Fatal("expected error when converting I64 to I32")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}

	// Test AsI64 with I64 value
	convertedI64, err := AsI64(i64Val)
	if err != nil {
		t.Fatalf("unexpected error converting I64: %v", err)
	}
	if convertedI64 != 123 {
		t.Errorf("expected converted I64 value 123, got %d", convertedI64)
	}

	// Test AsI64 with wrong type (should fail)
	_, err = AsI64(i32Val)
	if err == nil {
		t.Fatal("expected error when converting I32 to I64")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}

	// Test AsF32 with F32 value
	f32Val := NewF32(3.14159)
	convertedF32, err := AsF32(f32Val)
	if err != nil {
		t.Fatalf("unexpected error converting F32: %v", err)
	}
	if convertedF32 != 3.14159 {
		t.Errorf("expected converted F32 value 3.14159, got %f", convertedF32)
	}

	// Test AsF32 with wrong type (should fail)
	_, err = AsF32(i32Val)
	if err == nil {
		t.Fatal("expected error when converting I32 to F32")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}

	// Test AsF64 with F64 value
	f64Val := NewF64(2.718281828)
	convertedF64, err := AsF64(f64Val)
	if err != nil {
		t.Fatalf("unexpected error converting F64: %v", err)
	}
	if convertedF64 != 2.718281828 {
		t.Errorf("expected converted F64 value 2.718281828, got %f", convertedF64)
	}

	// Test AsF64 with wrong type (should fail)
	_, err = AsF64(i32Val)
	if err == nil {
		t.Fatal("expected error when converting I32 to F64")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}
}

// TestWASMGlobalComprehensive tests global variable functionality comprehensively
func TestWASMGlobalComprehensive(t *testing.T) {
	// Test mutable global
	initialValue := NewI32(100)
	mutableGlobal := NewWASMGlobal(WASMValueTypeI32, true, initialValue)

	if mutableGlobal == nil {
		t.Fatal("expected non-nil global")
	}

	if mutableGlobal.Type != WASMValueTypeI32 {
		t.Errorf("expected global type I32, got %d", mutableGlobal.Type)
	}

	if !mutableGlobal.Mutable {
		t.Error("expected global to be mutable")
	}

	// Test Get method
	retrievedValue := mutableGlobal.Get()
	if retrievedValue != initialValue {
		t.Error("expected retrieved value to match initial value")
	}

	// Test Set method with mutable global
	newValue := NewI32(200)
	err := mutableGlobal.Set(newValue)
	if err != nil {
		t.Errorf("unexpected error setting mutable global: %v", err)
	}

	updatedValue := mutableGlobal.Get()
	if updatedValue != newValue {
		t.Error("expected global value to be updated")
	}

	// Test immutable global
	immutableGlobal := NewWASMGlobal(WASMValueTypeI64, false, NewI64(999))

	if immutableGlobal.Mutable {
		t.Error("expected global to be immutable")
	}

	// Test Set method with immutable global (should fail)
	err = immutableGlobal.Set(NewI64(888))
	if err == nil {
		t.Fatal("expected error when setting immutable global")
	}
	if err != ErrGlobalImmutable {
		t.Errorf("expected ErrGlobalImmutable, got %v", err)
	}

	// Test Set method with type mismatch (should fail)
	err = mutableGlobal.Set(NewI64(123))
	if err == nil {
		t.Fatal("expected error when setting global with type mismatch")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}
}

// TestWASMFunctionType tests function type functionality
func TestWASMFunctionType(t *testing.T) {
	// Test function type creation
	paramTypes := []WASMValueType{WASMValueTypeI32, WASMValueTypeI64}
	resultTypes := []WASMValueType{WASMValueTypeF32, WASMValueTypeF64}

	functionType := NewWASMFunctionType(paramTypes, resultTypes)

	if functionType == nil {
		t.Fatal("expected non-nil function type")
	}

	if len(functionType.Params) != 2 {
		t.Errorf("expected 2 parameter types, got %d", len(functionType.Params))
	}

	if len(functionType.Results) != 2 {
		t.Errorf("expected 2 result types, got %d", len(functionType.Results))
	}

	if functionType.Params[0] != WASMValueTypeI32 {
		t.Error("expected first parameter type to be I32")
	}

	if functionType.Params[1] != WASMValueTypeI64 {
		t.Error("expected second parameter type to be I64")
	}

	if functionType.Results[0] != WASMValueTypeF32 {
		t.Error("expected first result type to be F32")
	}

	if functionType.Results[1] != WASMValueTypeF64 {
		t.Error("expected second result type to be F64")
	}

	// Test function type with no parameters or results
	emptyFunctionType := NewWASMFunctionType([]WASMValueType{}, []WASMValueType{})

	if emptyFunctionType == nil {
		t.Fatal("expected non-nil empty function type")
	}

	if len(emptyFunctionType.Params) != 0 {
		t.Errorf("expected 0 parameter types, got %d", len(emptyFunctionType.Params))
	}

	if len(emptyFunctionType.Results) != 0 {
		t.Errorf("expected 0 result types, got %d", len(emptyFunctionType.Results))
	}
}

// TestWASMFunctionComprehensive tests function functionality comprehensively
func TestWASMFunctionComprehensive(t *testing.T) {
	// Test function creation
	paramTypes := []WASMValueType{WASMValueTypeI32}
	resultTypes := []WASMValueType{WASMValueTypeI32}

	functionType := NewWASMFunctionType(paramTypes, resultTypes)
	code := []byte{0x01, 0x02, 0x03, 0x04}
	localTypes := []WASMValueType{WASMValueTypeI64}
	body := []byte{0x05, 0x06, 0x07, 0x08}

	function := NewWASMFunction(functionType, code, localTypes, body)

	if function == nil {
		t.Fatal("expected non-nil function")
	}

	if function.Type != functionType {
		t.Error("expected function type to match")
	}

	if len(function.Code) != 4 {
		t.Errorf("expected code length 4, got %d", len(function.Code))
	}

	if len(function.LocalTypes) != 1 {
		t.Errorf("expected 1 local type, got %d", len(function.LocalTypes))
	}

	if len(function.Body) != 4 {
		t.Errorf("expected body length 4, got %d", len(function.Body))
	}

	// Test function with empty code and body
	emptyFunction := NewWASMFunction(functionType, []byte{}, []WASMValueType{}, []byte{})

	if emptyFunction == nil {
		t.Fatal("expected non-nil empty function")
	}

	if len(emptyFunction.Code) != 0 {
		t.Errorf("expected empty code, got length %d", len(emptyFunction.Code))
	}

	if len(emptyFunction.Body) != 0 {
		t.Errorf("expected empty body, got length %d", len(emptyFunction.Body))
	}
}

// TestWASMTableComprehensive tests table functionality comprehensively
func TestWASMTableComprehensive(t *testing.T) {
	// Test table creation
	elementType := WASMValueTypeI32
	initial := uint32(10)
	maximum := uint32(100)

	table := NewWASMTable(elementType, initial, maximum)

	if table == nil {
		t.Fatal("expected non-nil table")
	}

	if table.ElementType != elementType {
		t.Errorf("expected element type %d, got %d", elementType, table.ElementType)
	}

	if table.Initial != initial {
		t.Errorf("expected initial size %d, got %d", initial, table.Initial)
	}

	if table.Maximum != maximum {
		t.Errorf("expected maximum size %d, got %d", maximum, table.Maximum)
	}

	if table.Size() != initial {
		t.Errorf("expected table size %d, got %d", initial, table.Size())
	}

	// Test setting and getting values
	testValue := NewI32(12345)
	err := table.Set(5, testValue)
	if err != nil {
		t.Errorf("unexpected error setting table value: %v", err)
	}

	retrievedValue, err := table.Get(5)
	if err != nil {
		t.Errorf("unexpected error getting table value: %v", err)
	}

	if retrievedValue != testValue {
		t.Error("expected retrieved value to match set value")
	}

	// Test setting value at invalid index (should fail)
	err = table.Set(15, testValue)
	if err == nil {
		t.Fatal("expected error when setting value at invalid index")
	}
	if err != ErrTableIndexOutOfBounds {
		t.Errorf("expected ErrTableIndexOutOfBounds, got %v", err)
	}

	// Test getting value at invalid index (should fail)
	_, err = table.Get(15)
	if err == nil {
		t.Fatal("expected error when getting value at invalid index")
	}
	if err != ErrTableIndexOutOfBounds {
		t.Errorf("expected ErrTableIndexOutOfBounds, got %v", err)
	}

	// Test setting value with wrong type (should fail)
	wrongTypeValue := NewI64(123)
	err = table.Set(6, wrongTypeValue)
	if err == nil {
		t.Fatal("expected error when setting value with wrong type")
	}
	if err != ErrTableTypeMismatch {
		t.Errorf("expected ErrTableTypeMismatch, got %v", err)
	}
}

// TestWASMTableGrow tests table growth functionality
func TestWASMTableGrow(t *testing.T) {
	// Test table growth
	table := NewWASMTable(WASMValueTypeI32, 5, 20)

	initialSize := table.Size()
	if initialSize != 5 {
		t.Errorf("expected initial table size 5, got %d", initialSize)
	}

	// Test successful growth
	newSize, err := table.Grow(3, NewI32(999))
	if err != nil {
		t.Errorf("unexpected error growing table: %v", err)
	}

	expectedSize := initialSize + 3
	if newSize != expectedSize {
		t.Errorf("expected new table size %d, got %d", expectedSize, newSize)
	}

	if table.Size() != expectedSize {
		t.Errorf("expected table size %d after growth, got %d", expectedSize, table.Size())
	}

	// Test setting value in grown area
	err = table.Set(6, NewI32(888))
	if err != nil {
		t.Errorf("unexpected error setting value in grown area: %v", err)
	}

	retrievedValue, err := table.Get(6)
	if err != nil {
		t.Errorf("unexpected error getting value from grown area: %v", err)
	}

	if retrievedValue == nil {
		t.Fatal("expected non-nil value from grown area")
	}

	// Test growth beyond maximum (should fail)
	_, err = table.Grow(20, NewI32(777))
	if err == nil {
		t.Fatal("expected error when growing table beyond maximum")
	}
	if err != ErrTableGrowExceedsMaximum {
		t.Errorf("expected ErrTableGrowExceedsMaximum, got %v", err)
	}
}

// TestWASMInstance tests instance functionality
func TestWASMInstance(t *testing.T) {
	// Test instance creation
	module := NewWASMModule()
	instance := NewWASMInstance(module)

	if instance == nil {
		t.Fatal("expected non-nil instance")
	}

	if instance.Module != module {
		t.Error("expected instance module to match")
	}

	if instance.Memory != module.Memory {
		t.Error("expected instance memory to match module memory")
	}

	if len(instance.Globals) != 0 {
		t.Errorf("expected empty globals map, got %d globals", len(instance.Globals))
	}

	if len(instance.Functions) != 0 {
		t.Errorf("expected empty functions map, got %d functions", len(instance.Functions))
	}

	if len(instance.Tables) != 0 {
		t.Errorf("expected empty tables map, got %d tables", len(instance.Tables))
	}

	if len(instance.Exports) != 0 {
		t.Errorf("expected empty exports map, got %d exports", len(instance.Exports))
	}
}

// TestWASMModule tests module functionality
func TestWASMModule(t *testing.T) {
	// Test module creation
	module := NewWASMModule()

	if module == nil {
		t.Fatal("expected non-nil module")
	}

	if len(module.Types) != 0 {
		t.Errorf("expected empty types slice, got %d types", len(module.Types))
	}

	if len(module.Functions) != 0 {
		t.Errorf("expected empty functions slice, got %d functions", len(module.Functions))
	}

	if len(module.Tables) != 0 {
		t.Errorf("expected empty tables slice, got %d tables", len(module.Tables))
	}

	if module.Memory != nil {
		t.Error("expected nil memory in new module")
	}

	if len(module.Globals) != 0 {
		t.Errorf("expected empty globals slice, got %d globals", len(module.Globals))
	}

	if len(module.Exports) != 0 {
		t.Errorf("expected empty exports slice, got %d exports", len(module.Exports))
	}

	if len(module.Imports) != 0 {
		t.Errorf("expected empty imports slice, got %d imports", len(module.Imports))
	}

	if module.Start != nil {
		t.Error("expected nil start function in new module")
	}
}

// TestWASMExport tests export functionality
func TestWASMExport(t *testing.T) {
	// Test export creation
	export := WASMExport{
		Name:  "test_function",
		Kind:  WASMExportKindFunction,
		Index: 42,
	}

	if export.Name != "test_function" {
		t.Errorf("expected export name 'test_function', got '%s'", export.Name)
	}

	if export.Kind != WASMExportKindFunction {
		t.Errorf("expected export kind Function, got %d", export.Kind)
	}

	if export.Index != 42 {
		t.Errorf("expected export index 42, got %d", export.Index)
	}
}

// TestWASMImport tests import functionality
func TestWASMImport(t *testing.T) {
	// Test import creation
	importItem := WASMImport{
		Module: "env",
		Name:   "memory",
		Kind:   WASMExportKindMemory,
		Index:  0,
	}

	if importItem.Module != "env" {
		t.Errorf("expected import module 'env', got '%s'", importItem.Module)
	}

	if importItem.Name != "memory" {
		t.Errorf("expected import name 'memory', got '%s'", importItem.Name)
	}

	if importItem.Kind != WASMExportKindMemory {
		t.Errorf("expected import kind Memory, got %d", importItem.Kind)
	}

	if importItem.Index != 0 {
		t.Errorf("expected import index 0, got %d", importItem.Index)
	}
}

// TestExecutionContext tests execution context functionality
func TestExecutionContext(t *testing.T) {
	// Test execution context creation
	contract := &engine.Contract{}
	input := []byte{0x01, 0x02, 0x03}
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(1000)
	gasPrice := big.NewInt(20)
	blockNum := uint64(12345)
	timestamp := uint64(67890)
	coinbase := engine.Address{0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D}
	difficulty := big.NewInt(100)
	chainID := big.NewInt(1)
	instance := &WASMInstance{}

	ctx := &ExecutionContext{
		Contract:   contract,
		Input:      input,
		Sender:     sender,
		Value:      value,
		GasPrice:   gasPrice,
		BlockNum:   blockNum,
		Timestamp:  timestamp,
		Coinbase:   coinbase,
		Difficulty: difficulty,
		ChainID:    chainID,
		Instance:   instance,
	}

	if ctx.Contract != contract {
		t.Error("expected contract to match")
	}

	if len(ctx.Input) != 3 {
		t.Errorf("expected input length 3, got %d", len(ctx.Input))
	}

	if ctx.Sender != sender {
		t.Error("expected sender to match")
	}

	if ctx.Value.Cmp(value) != 0 {
		t.Error("expected value to match")
	}

	if ctx.GasPrice.Cmp(gasPrice) != 0 {
		t.Error("expected gas price to match")
	}

	if ctx.BlockNum != blockNum {
		t.Errorf("expected block number %d, got %d", blockNum, ctx.BlockNum)
	}

	if ctx.Timestamp != timestamp {
		t.Errorf("expected timestamp %d, got %d", timestamp, ctx.Timestamp)
	}

	if ctx.Coinbase != coinbase {
		t.Error("expected coinbase to match")
	}

	if ctx.Difficulty.Cmp(difficulty) != 0 {
		t.Error("expected difficulty to match")
	}

	if ctx.ChainID.Cmp(chainID) != 0 {
		t.Error("expected chain ID to match")
	}

	if ctx.Instance != instance {
		t.Error("expected instance to match")
	}
}
