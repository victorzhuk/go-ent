package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCode(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   string
		input       ASTGenerateInput
		wantErr     bool
		errContains string
		checkOutput func(t *testing.T, result *generateResult)
	}{
		{
			name: "generate test scaffold",
			setupFile: `package test

func CreateUser(email string) (*User, error) {
	return nil, nil
}`,
			input: ASTGenerateInput{
				Type:     "test",
				File:     "test.go",
				Function: "CreateUser",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, result *generateResult) {
				assert.NotEmpty(t, result.Generated)
				assert.Contains(t, result.Generated, "func TestCreateUser")
				assert.Contains(t, result.Generated, "t.Parallel()")
				assert.Contains(t, result.Generated, "t.Run")
				assert.Contains(t, result.File, "test_test.go")
			},
		},
		{
			name: "empty type",
			setupFile: `package test

func Test() {}`,
			input: ASTGenerateInput{
				Type:     "",
				File:     "test.go",
				Function: "Test",
			},
			wantErr:     true,
			errContains: "type is required",
		},
		{
			name: "empty file",
			setupFile: `package test

func Test() {}`,
			input: ASTGenerateInput{
				Type:     "test",
				File:     "",
				Function: "Test",
			},
			wantErr:     true,
			errContains: "file path is required",
			checkOutput: func(t *testing.T, result *generateResult) {},
		},
		{
			name: "empty function",
			setupFile: `package test

func Test() {}`,
			input: ASTGenerateInput{
				Type:     "test",
				File:     "test.go",
				Function: "",
			},
			wantErr:     true,
			errContains: "function name is required",
		},
		{
			name: "unsupported type",
			setupFile: `package test

func Test() {}`,
			input: ASTGenerateInput{
				Type:     "invalid",
				File:     "test.go",
				Function: "Test",
			},
			wantErr:     true,
			errContains: "unsupported type",
		},
		{
			name: "function not found",
			setupFile: `package test

func Test() {}`,
			input: ASTGenerateInput{
				Type:     "test",
				File:     "test.go",
				Function: "NonExistent",
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			if tt.input.File == "" {
				result, err := generateCode(tt.input)

				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
				return
			}

			testFile := filepath.Join(tmpDir, "test.go")
			if tt.setupFile != "" {
				err := os.WriteFile(testFile, []byte(tt.setupFile), 0o600)
				require.NoError(t, err)
			}

			tt.input.File = testFile

			result, err := generateCode(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkOutput != nil {
					tt.checkOutput(t, result)
				}
			}
		})
	}
}

func TestGetTestFileName(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "simple go file",
			filePath: "test.go",
			want:     "test_test.go",
		},
		{
			name:     "path with directory",
			filePath: "pkg/test.go",
			want:     "pkg/test_test.go",
		},
		{
			name:     "file without go extension",
			filePath: "test",
			want:     "test_test.go",
		},
		{
			name:     "absolute path",
			filePath: "/home/user/test.go",
			want:     "/home/user/test_test.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getTestFileName(tt.filePath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatGenerateResult(t *testing.T) {
	result := &generateResult{
		Generated: `func TestExample(t *testing.T) {
	// test code
}`,
		File: "example_test.go",
	}

	output := formatGenerateResult(result)

	assert.Contains(t, output, "Generated Code:")
	assert.Contains(t, output, "==============")
	assert.Contains(t, output, result.Generated)
	assert.Contains(t, output, "Save to file: example_test.go")
}
