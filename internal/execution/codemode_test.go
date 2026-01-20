package execution

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeMode_VMInitialization(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	})
	codeMode := NewCodeMode(sandbox)

	assert.NotNil(t, codeMode)
	assert.NotNil(t, codeMode.vm)
	assert.Equal(t, sandbox, codeMode.sandbox)
	assert.Equal(t, 5*time.Second, codeMode.timeout)
}

func TestCodeMode_VMInitializationWithDefaultTimeout(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
	})
	codeMode := NewCodeMode(sandbox)

	assert.NotNil(t, codeMode)
	assert.NotNil(t, codeMode.vm)
	assert.Equal(t, 30*time.Second, codeMode.timeout)
}

func TestCodeMode_DangerousGlobalsDisabled(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	codeMode := NewCodeMode(sandbox)

	require.NotNil(t, codeMode)

	requireIsUndefined := func(name string) {
		val := codeMode.vm.Get(name)
		assert.True(t, goja.IsUndefined(val), "%s should be undefined", name)
	}

	requireIsUndefined("require")
	requireIsUndefined("process")
	requireIsUndefined("global")
}

func TestCodeMode_SafeGlobalsSet(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	codeMode := NewCodeMode(sandbox)

	require.NotNil(t, codeMode)

	console := codeMode.vm.Get("console")
	assert.False(t, goja.IsUndefined(console), "console should be defined")
	assert.IsType(t, codeMode.vm.NewObject(), console)
}

func TestCodeMode_VMReadyForExecution(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "1 + 1", nil)

	require.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func TestCodeMode_VMMultipleInstances(t *testing.T) {
	t.Parallel()

	sandbox1 := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
		MaxExecTime: 10 * time.Second,
	})
	sandbox2 := NewSandbox(ResourceLimits{
		MaxMemoryMB: 256,
		MaxExecTime: 20 * time.Second,
	})

	codeMode1 := NewCodeMode(sandbox1)
	codeMode2 := NewCodeMode(sandbox2)

	assert.NotNil(t, codeMode1)
	assert.NotNil(t, codeMode2)
	assert.NotEqual(t, codeMode1.vm, codeMode2.vm, "VMs should be separate instances")
	assert.NotEqual(t, codeMode1.timeout, codeMode2.timeout)
}

func TestCodeMode_VMConfigurationConsistency(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{"zero timeout defaults to 30s", 0, 30 * time.Second},
		{"1s timeout", 1 * time.Second, 1 * time.Second},
		{"10s timeout", 10 * time.Second, 10 * time.Second},
		{"30s timeout", 30 * time.Second, 30 * time.Second},
		{"60s timeout", 60 * time.Second, 60 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{
				MaxMemoryMB: 128,
				MaxCPUTime:  10 * time.Second,
				MaxExecTime: tt.timeout,
			})
			codeMode := NewCodeMode(sandbox)

			assert.NotNil(t, codeMode)
			assert.Equal(t, tt.expected, codeMode.timeout)
		})
	}
}

func TestCodeMode_VMReset(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.SetGlobal("testVar", 42)
	require.NoError(t, err)

	result := codeMode.GetGlobal("testVar")
	assert.Equal(t, int64(42), result)

	codeMode.Reset()

	resultAfterReset := codeMode.GetGlobal("testVar")
	assert.Nil(t, resultAfterReset)
}

func TestCodeMode_ResetPreservesConfiguration(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
		MaxExecTime: 15 * time.Second,
	})
	codeMode := NewCodeMode(sandbox)

	oldTimeout := codeMode.timeout
	oldSandbox := codeMode.sandbox

	codeMode.Reset()

	assert.Equal(t, oldTimeout, codeMode.timeout)
	assert.Equal(t, oldSandbox, codeMode.sandbox)
	assert.NotNil(t, codeMode.vm)

	console := codeMode.vm.Get("console")
	assert.False(t, goja.IsUndefined(console))
}

func TestCodeMode_Execute_SimpleArithmetic(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name     string
		script   string
		expected interface{}
	}{
		{"addition", "1 + 1", int64(2)},
		{"subtraction", "10 - 3", int64(7)},
		{"multiplication", "4 * 5", int64(20)},
		{"division", "20 / 4", int64(5)},
		{"modulo", "10 % 3", int64(1)},
		{"complex expression", "(5 + 3) * 2", int64(16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := codeMode.Execute(context.Background(), tt.script, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_Execute_StringOperations(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "'Hello, ' + 'World!'", nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

func TestCodeMode_Execute_BooleanOperations(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{"true equals true", "true === true", true},
		{"false equals false", "false === false", true},
		{"true not equals false", "true !== false", true},
		{"logical and", "true && false", false},
		{"logical or", "true || false", true},
		{"greater than", "5 > 3", true},
		{"less than", "3 < 5", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := codeMode.Execute(context.Background(), tt.script, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_Execute_LoopConstructs(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		let sum = 0;
		for (let i = 1; i <= 5; i++) {
			sum += i;
		}
		sum;
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(15), result)
}

func TestCodeMode_Execute_WhileLoop(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		let i = 0;
		let sum = 0;
		while (i < 5) {
			sum += i;
			i++;
		}
		sum;
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestCodeMode_Execute_FunctionDefinitionAndCall(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		function add(a, b) {
			return a + b;
		}
		add(3, 4);
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(7), result)
}

func TestCodeMode_Execute_ArrowFunction(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const multiply = (a, b) => a * b;
		multiply(6, 7);
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestCodeMode_Execute_ArrayOperations(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const arr = [1, 2, 3, 4, 5];
		arr.reduce((sum, val) => sum + val, 0);
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(15), result)
}

func TestCodeMode_Execute_ObjectOperations(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const obj = {a: 1, b: 2, c: 3};
		obj.a + obj.b + obj.c;
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(6), result)
}

func TestCodeMode_Execute_WithInputVariables(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	input := map[string]interface{}{
		"x": 10,
		"y": 20,
	}

	result, err := codeMode.Execute(context.Background(), "x + y", input)
	require.NoError(t, err)
	assert.Equal(t, int64(30), result)
}

func TestCodeMode_Execute_ConditionalLogic(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const x = 10;
		if (x > 5) {
			'greater';
		} else {
			'less';
		}
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, "greater", result)
}

func TestCodeMode_Execute_TernaryOperator(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const x = 15;
		x > 10 ? 'yes' : 'no';
	`, nil)
	require.NoError(t, err)
	assert.Equal(t, "yes", result)
}

func TestCodeMode_Execute_SyntaxError(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "function broken(", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_Execute_RuntimeError(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "undefinedVariable.notExists", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_Execute_ThrowError(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "throw new Error('test error')", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_ExecuteFunction_Simple(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("add", "function(a, b) { return a + b; }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "add", 5, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(8), result)
}

func TestCodeMode_ExecuteFunction_WithStrings(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("concat", "function(a, b) { return a + b; }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "concat", "Hello", " World")
	require.NoError(t, err)
	assert.Equal(t, "Hello World", result)
}

func TestCodeMode_ExecuteFunction_WithArray(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("sum", "function(arr) { return arr.reduce((a, b) => a + b, 0); }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "sum", []interface{}{1, 2, 3, 4, 5})
	require.NoError(t, err)
	assert.Equal(t, int64(15), result)
}

func TestCodeMode_ExecuteFunction_NonExistentFunction(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.ExecuteFunction(context.Background(), "nonExistentFunc", "arg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "function not found")
}

func TestCodeMode_ExecuteFunction_RuntimeErrorInFunction(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("badFunc", "function() { throw new Error('function error'); }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(context.Background(), "badFunc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "function execution")
}

func TestCodeMode_MultipleExecutionsMaintainState(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "let x = 10;", nil)
	require.NoError(t, err)

	result, err := codeMode.Execute(context.Background(), "x + 5", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(15), result)

	result, err = codeMode.Execute(context.Background(), "x * 2", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(20), result)
}

func TestCodeMode_Execute_ReturnNull(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "null", nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCodeMode_Execute_ReturnUndefined(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "undefined", nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCodeMode_VMMemoryLimitConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		memoryMB int
	}{
		{"32 MB limit", 32},
		{"64 MB limit", 64},
		{"128 MB limit", 128},
		{"256 MB limit", 256},
		{"512 MB limit", 512},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: tt.memoryMB})
			codeMode := NewCodeMode(sandbox)

			assert.NotNil(t, codeMode)
			assert.NotNil(t, codeMode.sandbox)
			assert.Equal(t, tt.memoryMB, codeMode.sandbox.GetLimits().MaxMemoryMB)
		})
	}
}

func TestCodeMode_VMMemoryLimitZeroNoCheck(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 0})
	codeMode := NewCodeMode(sandbox)

	assert.NotNil(t, codeMode)
	assert.Equal(t, 0, codeMode.sandbox.GetLimits().MaxMemoryMB)
}

func TestCodeMode_VMMemoryLimitAccess(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 256})
	codeMode := NewCodeMode(sandbox)

	limits := codeMode.sandbox.GetLimits()
	assert.Equal(t, 256, limits.MaxMemoryMB)
}

func TestCodeMode_VMMemoryLimitErrorFormat(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})

	err := sandbox.CheckMemoryLimit()
	if err != nil {
		assert.Contains(t, err.Error(), "resource limit exceeded")
		assert.Contains(t, err.Error(), "memory")
	}
}

func TestCodeMode_VMMemoryLimitVariousValues(t *testing.T) {
	tests := []struct {
		name     string
		memoryMB int
	}{
		{"16 MB minimum", 16},
		{"32 MB", 32},
		{"64 MB", 64},
		{"96 MB", 96},
		{"128 MB standard", 128},
		{"256 MB high", 256},
		{"512 MB very high", 512},
		{"1024 MB extreme", 1024},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: tt.memoryMB})
			codeMode := NewCodeMode(sandbox)

			limits := codeMode.sandbox.GetLimits()
			assert.Equal(t, tt.memoryMB, limits.MaxMemoryMB)
		})
	}
}

func TestCodeMode_VMMemoryLimitMultipleInstances(t *testing.T) {
	t.Parallel()

	sandbox1 := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	sandbox2 := NewSandbox(ResourceLimits{MaxMemoryMB: 256})

	codeMode1 := NewCodeMode(sandbox1)
	codeMode2 := NewCodeMode(sandbox2)

	assert.Equal(t, 128, codeMode1.sandbox.GetLimits().MaxMemoryMB)
	assert.Equal(t, 256, codeMode2.sandbox.GetLimits().MaxMemoryMB)
	assert.NotEqual(t, codeMode1.sandbox.GetLimits().MaxMemoryMB, codeMode2.sandbox.GetLimits().MaxMemoryMB)
}

func TestCodeMode_VMMemoryLimitConsistency(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	limits := codeMode.sandbox.GetLimits()
	assert.Equal(t, 128, limits.MaxMemoryMB)

	codeMode.Reset()
	limitsAfterReset := codeMode.sandbox.GetLimits().MaxMemoryMB
	assert.Equal(t, 128, limitsAfterReset)
}

func TestCodeMode_VMCleanupAfterReset(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.SetGlobal("testVar", 42)
	require.NoError(t, err)

	result := codeMode.GetGlobal("testVar")
	assert.Equal(t, int64(42), result)

	codeMode.Reset()

	resultAfterReset := codeMode.GetGlobal("testVar")
	assert.Nil(t, resultAfterReset)
	assert.NotNil(t, codeMode.vm)
}

func TestCodeMode_GoroutineCleanupAfterReset(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	initialGoroutines := runtime.NumGoroutine()

	_, err := codeMode.Execute(ctx, "for(let i=0; i<1000; i++) {}", nil)
	require.NoError(t, err)

	codeMode.Reset()

	time.Sleep(10 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()

	diff := finalGoroutines - initialGoroutines
	assert.True(t, diff <= 2, "goroutine leak detected: %d", diff)
}

func TestCodeMode_VMRecreationAfterReset(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	oldVM := codeMode.vm

	result1, err := codeMode.Execute(ctx, "let x = 10; x", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(10), result1)

	codeMode.Reset()

	assert.NotEqual(t, oldVM, codeMode.vm, "VM should be a new instance")

	result2, err := codeMode.Execute(ctx, "x", nil)
	assert.Error(t, err)
	assert.Nil(t, result2)

	result3, err := codeMode.Execute(ctx, "let y = 20; y", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(20), result3)
}

func TestCodeMode_AllowedFunction_ConsoleExposed(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	console := codeMode.vm.Get("console")
	assert.False(t, goja.IsUndefined(console), "console should be exposed")
	assert.IsType(t, codeMode.vm.NewObject(), console)
}

func TestCodeMode_AllowedFunction_ConsoleObjectAccessible(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "typeof console", nil)
	require.NoError(t, err)

	result, err := codeMode.Execute(context.Background(), "console !== undefined", nil)
	require.NoError(t, err)
	assert.True(t, result.(bool))
}

func TestCodeMode_AllowedFunction_DangerousGlobalsNotExposed(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name     string
		global   string
		expected bool
	}{
		{"require not exposed", "require", false},
		{"process not exposed", "process", false},
		{"global not exposed", "global", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := codeMode.vm.Get(tt.global)
			isUndefined := goja.IsUndefined(val)
			assert.Equal(t, tt.expected, !isUndefined, "%s exposure check", tt.global)
		})
	}
}

func TestCodeMode_AllowedFunction_DefineAndExposeFunction(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("multiply", "function(a, b) { return a * b; }")
	require.NoError(t, err)

	val := codeMode.vm.Get("multiply")
	assert.False(t, goja.IsUndefined(val), "function should be exposed")
}

func TestCodeMode_AllowedFunction_ExposedFunctionCanBeCalled(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("add", "function(a, b) { return a + b; }")
	require.NoError(t, err)

	fn, ok := goja.AssertFunction(codeMode.vm.Get("add"))
	assert.True(t, ok, "should be callable")

	args := []goja.Value{codeMode.vm.ToValue(5), codeMode.vm.ToValue(7)}
	result, err := fn(goja.Undefined(), args...)
	require.NoError(t, err)
	assert.Equal(t, int64(12), result.Export())
}

func TestCodeMode_AllowedFunction_InputVariablesExposed(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	input := map[string]interface{}{
		"value1": 100,
		"value2": "test string",
		"value3": true,
	}

	result, err := codeMode.Execute(context.Background(), "value1 + value2", input)
	require.NoError(t, err)
	assert.Equal(t, "100test string", result)
}

func TestCodeMode_AllowedFunction_MultipleExposedFunctions(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	functions := map[string]string{
		"add":      "function(a, b) { return a + b; }",
		"subtract": "function(a, b) { return a - b; }",
		"multiply": "function(a, b) { return a * b; }",
		"divide":   "function(a, b) { return a / b; }",
	}

	for name, fn := range functions {
		err := codeMode.DefineFunction(name, fn)
		require.NoError(t, err, "define function %s", name)

		val := codeMode.vm.Get(name)
		assert.False(t, goja.IsUndefined(val), "function %s should be exposed", name)
	}
}

func TestCodeMode_AllowedFunction_ExposedFunctionReturnsCorrectResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		funcDef  string
		args     []interface{}
		expected interface{}
	}{
		{"string concat", "function(a, b) { return a + b; }", []interface{}{"hello", " world"}, "hello world"},
		{"number addition", "function(a, b) { return a + b; }", []interface{}{10, 20}, int64(30)},
		{"boolean return", "function() { return true; }", nil, true},
		{"array return", "function() { return [1, 2, 3]; }", nil, []interface{}{int64(1), int64(2), int64(3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("testFunc", tt.funcDef)
			require.NoError(t, err)

			result, err := codeMode.ExecuteFunction(context.Background(), "testFunc", tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_AllowedFunction_JavaScriptCanAccessExposedFunctions(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("square", "function(x) { return x * x; }")
	require.NoError(t, err)

	result, err := codeMode.Execute(context.Background(), "square(5)", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(25), result)
}

func TestCodeMode_AllowedFunction_ExposedFunctionsPersistAcrossExecutions(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("counter", "function() { return (typeof count === 'undefined') ? 1 : ++count; }")
	require.NoError(t, err)

	result1, err := codeMode.Execute(context.Background(), "let count = 0; counter()", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result1)

	result2, err := codeMode.Execute(context.Background(), "counter()", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result2)
}

func TestCodeMode_BlockedFunctionAccess_DangerousGlobalsUndefined(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name   string
		global string
	}{
		{"require undefined", "require"},
		{"process undefined", "process"},
		{"global undefined", "global"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := codeMode.vm.Get(tt.global)
			assert.True(t, goja.IsUndefined(val), "%s should be undefined", tt.global)
		})
	}
}

func TestCodeMode_BlockedFunctionAccess_RequireAccessError(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "require('fs')", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_BlockedFunctionAccess_ProcessAccessError(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "process.exit()", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_BlockedFunctionAccess_GlobalAccessError(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), "global.test = 123", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_BlockedFunctionAccess_PropertyAccessOnBlocked(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name   string
		script string
	}{
		{"require property", "require.cache"},
		{"process property", "process.env"},
		{"global property", "global.process"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := codeMode.Execute(context.Background(), tt.script, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "script execution")
		})
	}
}

func TestCodeMode_BlockedFunctionAccess_FunctionCallOnBlocked(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name   string
		script string
	}{
		{"require call", "require('./test')"},
		{"process exit", "process.exit(0)"},
		{"process cwd", "process.cwd()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := codeMode.Execute(context.Background(), tt.script, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "script execution")
		})
	}
}

func TestCodeMode_BlockedFunctionAccess_TypeCheckReturnsUndefined(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	tests := []struct {
		name     string
		global   string
		expected string
	}{
		{"require type", "require", "undefined"},
		{"process type", "process", "undefined"},
		{"global type", "global", "undefined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := codeMode.Execute(context.Background(), fmt.Sprintf("typeof %s", tt.global), nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_BlockedFunctionAccess_ResetMaintainsBlocking(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	codeMode.Reset()

	tests := []struct {
		name   string
		global string
	}{
		{"require after reset", "require"},
		{"process after reset", "process"},
		{"global after reset", "global"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := codeMode.vm.Get(tt.global)
			assert.True(t, goja.IsUndefined(val), "%s should be undefined after reset", tt.global)
		})
	}
}

func TestCodeMode_BlockedFunctionAccess_LocalVarOverrideIsAllowed(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "var require = function() {}; typeof require", nil)
	require.NoError(t, err)
	assert.Equal(t, "function", result)
}

func TestCodeMode_BlockedFunctionAccess_MultipleAccessPatterns(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected interface{}
	}{
		{"direct access returns undefined", "require", nil},
		{"property access errors", "require.resolve", "error"},
		{"call access errors", "require()", "error"},
		{"typeof check returns undefined", "typeof require", "undefined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			result, err := codeMode.Execute(context.Background(), tt.script, nil)
			switch tt.expected {
			case "error":
				assert.Error(t, err)
			case "undefined":
				require.NoError(t, err)
				assert.Equal(t, "undefined", result)
			default:
				require.NoError(t, err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestCodeMode_BlockedFunctionAccess_EvalIsAccessible(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "eval('1+1')", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func TestCodeMode_BlockedFunctionAccess_FunctionIsAccessible(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "Function('return 1+1')()", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func TestCodeMode_MemoryReleasedAfterReset(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		_, _ = codeMode.Execute(ctx, "let arr = new Array(1000).fill(0); arr", nil)
	}

	codeMode.Reset()

	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	assert.NotNil(t, codeMode.vm)
	assert.NotNil(t, codeMode.sandbox)
}

func TestCodeMode_FunctionArgValidation_NumberTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg1     interface{}
		arg2     interface{}
		expected interface{}
	}{
		{"int + int", 5, 3, int64(8)},
		{"int + float64", 5, 3.5, 8.5},
		{"float64 + float64", 2.5, 3.5, int64(6)},
		{"negative numbers", -5, -3, int64(-8)},
		{"zero values", 0, 0, int64(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("add", "function(a, b) { return a + b; }")
			require.NoError(t, err)

			result, err := codeMode.ExecuteFunction(context.Background(), "add", tt.arg1, tt.arg2)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_FunctionArgValidation_StringTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg1     interface{}
		arg2     interface{}
		expected string
	}{
		{"string + string", "Hello", "World", "HelloWorld"},
		{"empty strings", "", "", ""},
		{"unicode strings", "Hello", "世界", "Hello世界"},
		{"string with spaces", "Hello ", "World", "Hello World"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("concat", "function(a, b) { return a + b; }")
			require.NoError(t, err)

			result, err := codeMode.ExecuteFunction(context.Background(), "concat", tt.arg1, tt.arg2)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_FunctionArgValidation_BooleanTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg1     interface{}
		arg2     interface{}
		expected bool
	}{
		{"true && true", true, true, true},
		{"true && false", true, false, false},
		{"false && true", false, true, false},
		{"false && false", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("logicalAnd", "function(a, b) { return a && b; }")
			require.NoError(t, err)

			result, err := codeMode.ExecuteFunction(context.Background(), "logicalAnd", tt.arg1, tt.arg2)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_FunctionArgValidation_ArrayTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg      interface{}
		expected interface{}
	}{
		{"array of ints", []interface{}{1, 2, 3, 4, 5}, int64(15)},
		{"empty array", []interface{}{}, int64(0)},
		{"single element", []interface{}{42}, int64(42)},
		{"negative numbers", []interface{}{-1, -2, -3}, int64(-6)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("sumArray", "function(arr) { let sum = 0; for(let i = 0; i < arr.length; i++) { sum += arr[i]; } return sum; }")
			require.NoError(t, err)

			result, err := codeMode.ExecuteFunction(context.Background(), "sumArray", tt.arg)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_FunctionArgValidation_ObjectTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		obj      interface{}
		key      interface{}
		expected interface{}
	}{
		{"get existing property", map[string]interface{}{"name": "John", "age": 30}, "name", "John"},
		{"get number property", map[string]interface{}{"count": 42}, "count", int64(42)},
		{"get boolean property", map[string]interface{}{"active": true}, "active", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("getProperty", "function(obj, key) { return obj[key]; }")
			require.NoError(t, err)

			result, err := codeMode.ExecuteFunction(context.Background(), "getProperty", tt.obj, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_FunctionArgValidation_NullAndUndefined(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("checkNull", "function(x) { return x === null; }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "checkNull", nil)
	require.NoError(t, err)
	assert.True(t, result.(bool))
}

func TestCodeMode_FunctionArgValidation_WrongArgCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []interface{}
	}{
		{"no args provided", []interface{}{}},
		{"one arg provided", []interface{}{5}},
		{"three args provided", []interface{}{5, 3, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("twoArgs", "function(a, b) { return a + b; }")
			require.NoError(t, err)

			_, err = codeMode.ExecuteFunction(context.Background(), "twoArgs", tt.args...)
			require.NoError(t, err)
		})
	}
}

func TestCodeMode_FunctionArgValidation_TypeMismatchInFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		arg   interface{}
		valid bool
	}{
		{"valid number", 5, true},
		{"valid float", 2.5, true},
		{"string treated as NaN", "hello", true},
		{"null treated as 0", nil, true},
		{"array treated as NaN", []interface{}{1, 2}, true},
		{"object treated as NaN", map[string]interface{}{"x": 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			err := codeMode.DefineFunction("expectNumber", "function(x) { return x * 2; }")
			require.NoError(t, err)

			var result interface{}
			result, err = codeMode.ExecuteFunction(context.Background(), "expectNumber", tt.arg)
			if tt.valid {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestCodeMode_FunctionArgValidation_MixedTypes(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("mixed", "function(a, b, c) { return typeof a + ',' + typeof b + ',' + typeof c; }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "mixed", "string", 42, true)
	require.NoError(t, err)
	assert.Equal(t, "string,number,boolean", result)
}

func TestCodeMode_FunctionArgValidation_NoArgs(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("noArgs", "function() { return 'no args'; }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "noArgs")
	require.NoError(t, err)
	assert.Equal(t, "no args", result)
}

func TestCodeMode_FunctionArgValidation_VariadicArgs(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("variadic", "function(...args) { return args.length; }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(context.Background(), "variadic", 1, 2, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result)

	result2, err := codeMode.ExecuteFunction(context.Background(), "variadic")
	require.NoError(t, err)
	assert.Equal(t, int64(0), result2)
}

func TestCodeMode_ReturnValue_SimpleTypes(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected interface{}
	}{
		{"number integer", "42", int64(42)},
		{"number float", "3.14", 3.14},
		{"number negative", "-100", int64(-100)},
		{"string", "'hello'", "hello"},
		{"boolean true", "true", true},
		{"boolean false", "false", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
			codeMode := NewCodeMode(sandbox)

			result, err := codeMode.Execute(context.Background(), tt.script, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeMode_ReturnValue_NestedObjects(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const obj = {
			user: {
				name: "John",
				age: 30,
				address: {
					city: "Boston",
					zip: "02101"
				}
			}
		};
		obj;
	`, nil)
	require.NoError(t, err)

	obj, ok := result.(map[string]interface{})
	require.True(t, ok)

	user, ok := obj["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "John", user["name"])
	assert.Equal(t, int64(30), user["age"])

	address, ok := user["address"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Boston", address["city"])
	assert.Equal(t, "02101", address["zip"])
}

func TestCodeMode_ReturnValue_Arrays(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const arr = [1, 2, 3, 4, 5];
		arr;
	`, nil)
	require.NoError(t, err)

	arr, ok := result.([]interface{})
	require.True(t, ok)
	assert.Len(t, arr, 5)
	assert.Equal(t, int64(1), arr[0])
	assert.Equal(t, int64(5), arr[4])
}

func TestCodeMode_ReturnValue_NestedArrays(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const matrix = [[1, 2], [3, 4], [5, 6]];
		matrix;
	`, nil)
	require.NoError(t, err)

	matrix, ok := result.([]interface{})
	require.True(t, ok)
	assert.Len(t, matrix, 3)

	row0, ok := matrix[0].([]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(1), row0[0])
	assert.Equal(t, int64(2), row0[1])
}

func TestCodeMode_ReturnValue_MixedObject(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const data = {
			id: 1,
			name: "Test",
			active: true,
			tags: ["a", "b", "c"],
			metadata: {
				created: 1234567890,
				updated: 1234567891
			}
		};
		data;
	`, nil)
	require.NoError(t, err)

	data, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(1), data["id"])
	assert.Equal(t, "Test", data["name"])
	assert.Equal(t, true, data["active"])

	tags, ok := data["tags"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tags, 3)

	metadata, ok := data["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(1234567890), metadata["created"])
}

func TestCodeMode_ReturnValue_Null(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "null", nil)
	require.NoError(t, err)
	assert.Nil(t, result)

	result2, err := codeMode.Execute(context.Background(), `
		function returnsNull() { return null; }
		returnsNull();
	`, nil)
	require.NoError(t, err)
	assert.Nil(t, result2)
}

func TestCodeMode_ReturnValue_Undefined(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "undefined", nil)
	require.NoError(t, err)
	assert.Nil(t, result)

	result2, err := codeMode.Execute(context.Background(), `
		function noReturn() { }
		noReturn();
	`, nil)
	require.NoError(t, err)
	assert.Nil(t, result2)
}

func TestCodeMode_ReturnValue_ErrorThrow(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), `
		throw new Error('test error');
	`, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "script execution")
}

func TestCodeMode_ReturnValue_ErrorFromFunction(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	err := codeMode.DefineFunction("errorFunc", "function() { throw new Error('function error'); }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(context.Background(), "errorFunc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "function execution")
}

func TestCodeMode_ReturnValue_ErrorMessageCapture(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	_, err := codeMode.Execute(context.Background(), `
		throw new Error('specific error message');
	`, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "specific error message")
}

func TestCodeMode_ReturnValue_VerifyNoReturn(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), `
		const x = 10;
		const y = 20;
	`, nil)
	require.NoError(t, err)
	assert.Nil(t, result, "code without explicit return should return nil")
}

func TestCodeMode_ReturnValue_LargeNumber(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "9007199254740991", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(9007199254740991), result)
}

func TestCodeMode_ReturnValue_ScientificNotation(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{MaxMemoryMB: 128})
	codeMode := NewCodeMode(sandbox)

	result, err := codeMode.Execute(context.Background(), "1.5e20", nil)
	require.NoError(t, err)
	assert.Equal(t, 1.5e20, result)
}
