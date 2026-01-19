package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// CodeMode provides JavaScript sandbox for dynamic tool composition.
type CodeMode struct {
	vm      *goja.Runtime
	sandbox *Sandbox
	timeout time.Duration
}

// NewCodeMode creates a new code-mode executor with the given sandbox.
func NewCodeMode(sandbox *Sandbox) *CodeMode {
	vm := goja.New()

	// Disable dangerous globals
	_ = vm.Set("require", goja.Undefined())
	_ = vm.Set("process", goja.Undefined())
	_ = vm.Set("global", goja.Undefined())

	// Set up safe globals
	_ = vm.Set("console", vm.NewObject()) // Stub console

	timeout := sandbox.GetLimits().MaxExecTime
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &CodeMode{
		vm:      vm,
		sandbox: sandbox,
		timeout: timeout,
	}
}

// Execute runs JavaScript code in the isolated sandbox.
func (c *CodeMode) Execute(ctx context.Context, script string, input map[string]interface{}) (interface{}, error) {
	for k, v := range input {
		if err := c.vm.Set(k, v); err != nil {
			return nil, fmt.Errorf("set input %s: %w", k, err)
		}
	}

	done := make(chan interface{}, 1)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("panic recovered: %v", r)
			}
		}()
		result, err := c.vm.RunString(script)
		if err != nil {
			errCh <- err
			return
		}
		done <- result.Export()
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errCh:
		return nil, fmt.Errorf("script execution: %w", err)
	case <-time.After(c.timeout):
		c.vm.Interrupt("timeout")
		return nil, fmt.Errorf("script timeout after %v", c.timeout)
	case <-ctx.Done():
		c.vm.Interrupt("cancelled")
		return nil, ctx.Err()
	}
}

// ExecuteFunction executes a JavaScript function with arguments.
func (c *CodeMode) ExecuteFunction(ctx context.Context, funcName string, args ...interface{}) (interface{}, error) {
	fn, ok := goja.AssertFunction(c.vm.Get(funcName))
	if !ok {
		return nil, fmt.Errorf("function not found: %s", funcName)
	}

	done := make(chan interface{}, 1)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("panic recovered: %v", r)
			}
		}()
		gojaArgs := make([]goja.Value, len(args))
		for i, arg := range args {
			gojaArgs[i] = c.vm.ToValue(arg)
		}

		result, err := fn(goja.Undefined(), gojaArgs...)
		if err != nil {
			errCh <- err
			return
		}
		done <- result.Export()
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errCh:
		return nil, fmt.Errorf("function execution: %w", err)
	case <-time.After(c.timeout):
		c.vm.Interrupt("timeout")
		return nil, fmt.Errorf("function timeout after %v", c.timeout)
	case <-ctx.Done():
		c.vm.Interrupt("cancelled")
		return nil, ctx.Err()
	}
}

// DefineFunction defines a JavaScript function from code.
func (c *CodeMode) DefineFunction(name, code string) error {
	_, err := c.vm.RunString(fmt.Sprintf("var %s = %s", name, code))
	if err != nil {
		return fmt.Errorf("define function %s: %w", name, err)
	}
	return nil
}

// SetGlobal sets a global variable in the VM.
func (c *CodeMode) SetGlobal(name string, value interface{}) error {
	return c.vm.Set(name, value)
}

// GetGlobal gets a global variable from the VM.
func (c *CodeMode) GetGlobal(name string) interface{} {
	val := c.vm.Get(name)
	if val == nil {
		return nil
	}
	return val.Export()
}

// Reset resets the VM state (clears all user-defined globals).
func (c *CodeMode) Reset() {
	// Create a new VM instance
	c.vm = goja.New()

	// Re-disable dangerous globals
	_ = c.vm.Set("require", goja.Undefined())
	_ = c.vm.Set("process", goja.Undefined())
	_ = c.vm.Set("global", goja.Undefined())

	// Set up safe globals
	_ = c.vm.Set("console", c.vm.NewObject())
}
