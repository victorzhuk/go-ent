package execution

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandbox_MemoryLimitStorage(t *testing.T) {
	tests := []struct {
		name       string
		memoryMB   int
		expectedMB int
	}{
		{"default memory limit", 0, 128},
		{"custom 256MB limit", 256, 256},
		{"custom 512MB limit", 512, 512},
		{"custom 1GB limit", 1024, 1024},
		{"minimal 16MB limit", 16, 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			limits := ResourceLimits{
				MaxMemoryMB: tt.memoryMB,
				MaxCPUTime:  30 * time.Second,
				MaxExecTime: 60 * time.Second,
			}
			if tt.memoryMB == 0 {
				limits = DefaultResourceLimits()
			}
			sandbox := NewSandbox(limits)
			retrieved := sandbox.GetLimits()
			assert.Equal(t, tt.expectedMB, retrieved.MaxMemoryMB)
		})
	}
}

func TestSandbox_DefaultResourceLimits(t *testing.T) {
	t.Parallel()
	limits := DefaultResourceLimits()
	assert.Equal(t, 128, limits.MaxMemoryMB)
	assert.Equal(t, 30*time.Second, limits.MaxCPUTime)
	assert.Equal(t, 60*time.Second, limits.MaxExecTime)
}

func TestSandbox_GetLimits(t *testing.T) {
	t.Parallel()
	expectedLimits := ResourceLimits{
		MaxMemoryMB: 512,
		MaxCPUTime:  45 * time.Second,
		MaxExecTime: 90 * time.Second,
	}
	sandbox := NewSandbox(expectedLimits)
	actualLimits := sandbox.GetLimits()
	assert.Equal(t, expectedLimits.MaxMemoryMB, actualLimits.MaxMemoryMB)
	assert.Equal(t, expectedLimits.MaxCPUTime, actualLimits.MaxCPUTime)
	assert.Equal(t, expectedLimits.MaxExecTime, actualLimits.MaxExecTime)
}

func TestSandbox_WithFileAccess(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	paths := []string{"/tmp/file1.txt", "/tmp/file2.txt"}
	sandbox = sandbox.WithFileAccess(paths...)
	err := sandbox.CheckFileAccess("/tmp/file1.txt")
	assert.NoError(t, err)
	err = sandbox.CheckFileAccess("/tmp/file3.txt")
	assert.Error(t, err)
}

func TestSandbox_WithAPIAccess(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	apis := []string{"http.get", "http.post"}
	sandbox = sandbox.WithAPIAccess(apis...)
	err := sandbox.CheckAPIAccess("http.get")
	assert.NoError(t, err)
	err = sandbox.CheckAPIAccess("http.put")
	assert.Error(t, err)
}

func TestCodeMode_WithMemoryLimits(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 64,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 20 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	assert.NotNil(t, codeMode)
	assert.Equal(t, sandbox, codeMode.sandbox)
}

func TestSandbox_FileAccessNoRestrictions(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	err := sandbox.CheckFileAccess("/any/path")
	assert.NoError(t, err)
}

func TestSandbox_APIAccessNoRestrictions(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	err := sandbox.CheckAPIAccess("any.api.call")
	assert.NoError(t, err)
}

func TestSandbox_MultipleFileAccess(t *testing.T) {
	t.Parallel()
	sandbox := NewSandbox(ResourceLimits{
		MaxMemoryMB: 128,
	})
	sandbox = sandbox.WithFileAccess("/tmp/data.txt").
		WithFileAccess("/var/log/app.log")
	tests := []struct {
		path     string
		expected bool
	}{
		{"/tmp/data.txt", true},
		{"/var/log/app.log", true},
		{"/etc/passwd", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := sandbox.CheckFileAccess(tt.path)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSandbox_ZeroMemoryLimit(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 0,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	retrieved := sandbox.GetLimits()
	assert.Equal(t, 0, retrieved.MaxMemoryMB)
}

func TestCodeMode_ExecuteSimpleScript(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()
	result, err := codeMode.Execute(ctx, "1 + 1", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSandbox_CPULimitStorage(t *testing.T) {
	tests := []struct {
		name         string
		cpuTime      time.Duration
		expectedTime time.Duration
	}{
		{"default CPU limit", 0, 30 * time.Second},
		{"custom 5s limit", 5 * time.Second, 5 * time.Second},
		{"custom 10s limit", 10 * time.Second, 10 * time.Second},
		{"custom 30s limit", 30 * time.Second, 30 * time.Second},
		{"custom 60s limit", 60 * time.Second, 60 * time.Second},
		{"minimal 1s limit", 1 * time.Second, 1 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			limits := ResourceLimits{
				MaxMemoryMB: 128,
				MaxCPUTime:  tt.cpuTime,
				MaxExecTime: 60 * time.Second,
			}
			if tt.cpuTime == 0 {
				limits = DefaultResourceLimits()
			}
			sandbox := NewSandbox(limits)
			retrieved := sandbox.GetLimits()
			assert.Equal(t, tt.expectedTime, retrieved.MaxCPUTime)
		})
	}
}

func TestSandbox_DefaultCPULimits(t *testing.T) {
	t.Parallel()
	limits := DefaultResourceLimits()
	assert.Equal(t, 30*time.Second, limits.MaxCPUTime)
}

func TestSandbox_GetCPULimits(t *testing.T) {
	t.Parallel()
	expectedLimits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  15 * time.Second,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(expectedLimits)
	actualLimits := sandbox.GetLimits()
	assert.Equal(t, expectedLimits.MaxCPUTime, actualLimits.MaxCPUTime)
}

func TestSandbox_ZeroCPULimit(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  0,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	retrieved := sandbox.GetLimits()
	assert.Equal(t, time.Duration(0), retrieved.MaxCPUTime)
}

func TestCodeMode_WithCPULimits(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 64,
		MaxCPUTime:  5 * time.Second,
		MaxExecTime: 10 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	assert.NotNil(t, codeMode)
	assert.Equal(t, 5*time.Second, codeMode.sandbox.GetLimits().MaxCPUTime)
}

func TestSandbox_ExecTimeLimitStorage(t *testing.T) {
	tests := []struct {
		name         string
		execTime     time.Duration
		expectedTime time.Duration
	}{
		{"default exec time limit", 0, 60 * time.Second},
		{"custom 5s limit", 5 * time.Second, 5 * time.Second},
		{"custom 10s limit", 10 * time.Second, 10 * time.Second},
		{"custom 30s limit", 30 * time.Second, 30 * time.Second},
		{"custom 60s limit", 60 * time.Second, 60 * time.Second},
		{"custom 120s limit", 120 * time.Second, 120 * time.Second},
		{"minimal 1s limit", 1 * time.Second, 1 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			limits := ResourceLimits{
				MaxMemoryMB: 128,
				MaxCPUTime:  30 * time.Second,
				MaxExecTime: tt.execTime,
			}
			if tt.execTime == 0 {
				limits = DefaultResourceLimits()
			}
			sandbox := NewSandbox(limits)
			retrieved := sandbox.GetLimits()
			assert.Equal(t, tt.expectedTime, retrieved.MaxExecTime)
		})
	}
}

func TestSandbox_DefaultExecTimeLimits(t *testing.T) {
	t.Parallel()
	limits := DefaultResourceLimits()
	assert.Equal(t, 60*time.Second, limits.MaxExecTime)
}

func TestSandbox_GetExecTimeLimits(t *testing.T) {
	t.Parallel()
	expectedLimits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 45 * time.Second,
	}
	sandbox := NewSandbox(expectedLimits)
	actualLimits := sandbox.GetLimits()
	assert.Equal(t, expectedLimits.MaxExecTime, actualLimits.MaxExecTime)
}

func TestSandbox_ZeroExecTimeLimit(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 0,
	}
	sandbox := NewSandbox(limits)
	retrieved := sandbox.GetLimits()
	assert.Equal(t, time.Duration(0), retrieved.MaxExecTime)
}

func TestCodeMode_WithExecTimeLimits(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 64,
		MaxCPUTime:  5 * time.Second,
		MaxExecTime: 10 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	assert.NotNil(t, codeMode)
	assert.Equal(t, 10*time.Second, codeMode.sandbox.GetLimits().MaxExecTime)
}

func TestSandbox_TimeoutEnforcement(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 100 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	_ = NewCodeMode(sandbox)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	select {
	case <-time.After(200 * time.Millisecond):
		assert.Fail(t, "timeout should have been triggered")
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	}
}

func TestSandbox_TimeoutCleanup(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	_ = sandbox.GetLimits()
	sandbox = NewSandbox(ResourceLimits{
		MaxMemoryMB: 256,
		MaxCPUTime:  60 * time.Second,
		MaxExecTime: 120 * time.Second,
	})
	retrieved := sandbox.GetLimits()
	assert.Equal(t, 120*time.Second, retrieved.MaxExecTime)
	assert.Equal(t, 256, retrieved.MaxMemoryMB)
}

func TestSandbox_TimeoutViolationDetection(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 75 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	retrieved := sandbox.GetLimits()
	assert.Equal(t, 75*time.Millisecond, retrieved.MaxExecTime)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	select {
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "timeout should have been triggered")
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	}
}

func TestCodeMode_PanicRecovery(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	_, err := codeMode.Execute(ctx, `throw new Error("test panic")`, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test panic")
}

func TestCodeMode_PanicRecoveryMultiple(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := codeMode.Execute(ctx, `throw new Error("panic `+strconv.Itoa(i)+`")`, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), strconv.Itoa(i))
	}
}

func TestCodeMode_PanicCleanup(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	_, err := codeMode.Execute(ctx, `throw new Error("cleanup test")`, nil)
	assert.Error(t, err)

	result, err := codeMode.Execute(ctx, "42", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestCodeMode_FunctionPanicRecovery(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	err := codeMode.DefineFunction("panicFunc", "() => { throw new Error('func panic') }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(ctx, "panicFunc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "func panic")
}

func TestCodeMode_FunctionPanicCleanup(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	err := codeMode.DefineFunction("panicFunc", "() => { throw new Error('func cleanup panic') }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(ctx, "panicFunc")
	assert.Error(t, err)

	err = codeMode.DefineFunction("safeFunc", "() => { return 99 }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(ctx, "safeFunc")
	require.NoError(t, err)
	assert.Equal(t, int64(99), result)
}

func TestSandbox_MemoryExhaustionDetection(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 1,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckMemoryLimit()
	if err != nil {
		var resErr *ResourceExceededError
		if errors.As(err, &resErr) {
			assert.Equal(t, "memory", resErr.Resource)
			assert.Contains(t, err.Error(), "memory")
		}
	}
}

func TestSandbox_MemoryExhaustionErrorMessage(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckMemoryLimit()
	if err != nil {
		assert.Contains(t, err.Error(), "resource limit exceeded")
		assert.Contains(t, err.Error(), "memory")
	}
}

func TestSandbox_CPUExhaustionDetection(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  100 * time.Millisecond,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckCPULimit(200 * time.Millisecond)
	assert.Error(t, err)

	var resErr *ResourceExceededError
	assert.True(t, errors.As(err, &resErr))
	if resErr != nil {
		assert.Equal(t, "cpu", resErr.Resource)
		assert.Equal(t, 100*time.Millisecond, resErr.Limit)
	}
}

func TestSandbox_CPUExhaustionErrorMessage(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  50 * time.Millisecond,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckCPULimit(100 * time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource limit exceeded")
	assert.Contains(t, err.Error(), "cpu")
}

func TestSandbox_ExecTimeExhaustionDetection(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 75 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckExecLimit(150 * time.Millisecond)
	assert.Error(t, err)

	var resErr *ResourceExceededError
	assert.True(t, errors.As(err, &resErr))
	if resErr != nil {
		assert.Equal(t, "execution", resErr.Resource)
		assert.Equal(t, 75*time.Millisecond, resErr.Limit)
	}
}

func TestSandbox_ExecTimeExhaustionErrorMessage(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckExecLimit(100 * time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource limit exceeded")
	assert.Contains(t, err.Error(), "execution")
}

func TestSandbox_CPULimitNotExceeded(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckCPULimit(10 * time.Second)
	assert.NoError(t, err)
}

func TestSandbox_ExecLimitNotExceeded(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 60 * time.Second,
	}
	sandbox := NewSandbox(limits)
	err := sandbox.CheckExecLimit(30 * time.Second)
	assert.NoError(t, err)
}

func TestSandbox_ZeroLimitsNoCheck(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 0,
		MaxCPUTime:  0,
		MaxExecTime: 0,
	}
	sandbox := NewSandbox(limits)

	err := sandbox.CheckMemoryLimit()
	assert.NoError(t, err)

	err = sandbox.CheckCPULimit(1 * time.Hour)
	assert.NoError(t, err)

	err = sandbox.CheckExecLimit(1 * time.Hour)
	assert.NoError(t, err)
}

func TestSandbox_ResourceExceededErrorUnwrap(t *testing.T) {
	t.Parallel()
	err := &ResourceExceededError{
		Resource: "test",
		Limit:    100,
	}
	assert.True(t, errors.Is(err, ErrResourceExceeded))
}

func TestSandbox_ResourceExceededErrorFormat(t *testing.T) {
	t.Parallel()
	err := &ResourceExceededError{
		Resource: "memory",
		Limit:    "128MB",
	}
	expected := "resource limit exceeded: memory limit 128MB"
	assert.Equal(t, expected, err.Error())
}

func TestCodeMode_ScriptTimeoutDetection(t *testing.T) {

	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	infiniteLoop := "while(true) { }"
	_, err := codeMode.Execute(ctx, infiniteLoop, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestCodeMode_FunctionTimeoutDetection(t *testing.T) {

	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	err := codeMode.DefineFunction("infiniteFunc", "() => { while(true) { } }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(ctx, "infiniteFunc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestCodeMode_TimeoutErrorContainsDuration(t *testing.T) {

	timeout := 75 * time.Millisecond
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: timeout,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	infiniteLoop := "while(true) { }"
	_, err := codeMode.Execute(ctx, infiniteLoop, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), timeout.String())
}

func TestCodeMode_TimeoutNoResourceLeak(t *testing.T) {

	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	infiniteLoop := "while(true) { }"
	_, err := codeMode.Execute(ctx, infiniteLoop, nil)
	assert.Error(t, err)

	result, err := codeMode.Execute(ctx, "1 + 1", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func TestCodeMode_ContextTimeout(t *testing.T) {

	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 30 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	infiniteLoop := "while(true) { }"
	_, err := codeMode.Execute(ctx, infiniteLoop, nil)
	assert.Error(t, err)
}

func TestCodeMode_FunctionTimeoutNoResourceLeak(t *testing.T) {

	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	err := codeMode.DefineFunction("infiniteFunc", "() => { while(true) { } }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(ctx, "infiniteFunc")
	assert.Error(t, err)

	err = codeMode.DefineFunction("safeFunc", "() => { return 42 }")
	require.NoError(t, err)

	result, err := codeMode.ExecuteFunction(ctx, "safeFunc")
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestCodeMode_VMCleanupAfterPanic(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	_, err := codeMode.Execute(ctx, `throw new Error("panic")`, nil)
	assert.Error(t, err)

	assert.NotNil(t, codeMode.vm)
}

func TestCodeMode_VMCleanupAfterTimeout(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 50 * time.Millisecond,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	infiniteLoop := "while(true) { }"
	_, err := codeMode.Execute(ctx, infiniteLoop, nil)
	assert.Error(t, err)

	assert.NotNil(t, codeMode.vm)
}

func TestCodeMode_VMCleanupAfterContextCancel(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	infiniteLoop := "while(true) { }"
	_, err := codeMode.Execute(ctx, infiniteLoop, nil)
	assert.Error(t, err)

	assert.NotNil(t, codeMode.vm)
}

func TestCodeMode_GoroutineCleanupAfterError(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	initialGoroutines := runtime.NumGoroutine()

	_, err := codeMode.Execute(ctx, `throw new Error("panic")`, nil)
	assert.Error(t, err)

	time.Sleep(10 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()

	diff := finalGoroutines - initialGoroutines
	assert.True(t, diff <= 2, "goroutine leak detected: %d", diff)
}

func TestCodeMode_FunctionVMCleanupAfterError(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	err := codeMode.DefineFunction("panicFunc", "() => { throw new Error('error') }")
	require.NoError(t, err)

	_, err = codeMode.ExecuteFunction(ctx, "panicFunc")
	assert.Error(t, err)

	assert.NotNil(t, codeMode.vm)
}

func TestCodeMode_FunctionGoroutineCleanupAfterError(t *testing.T) {
	t.Parallel()
	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  10 * time.Second,
		MaxExecTime: 5 * time.Second,
	}
	sandbox := NewSandbox(limits)
	codeMode := NewCodeMode(sandbox)
	ctx := context.Background()

	err := codeMode.DefineFunction("panicFunc", "() => { throw new Error('error') }")
	require.NoError(t, err)

	initialGoroutines := runtime.NumGoroutine()

	_, err = codeMode.ExecuteFunction(ctx, "panicFunc")
	assert.Error(t, err)

	time.Sleep(10 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()

	diff := finalGoroutines - initialGoroutines
	assert.True(t, diff <= 2, "goroutine leak detected: %d", diff)
}
