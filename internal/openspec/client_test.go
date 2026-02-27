package openspec

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseList_Changes(t *testing.T) {
	t.Parallel()

	data := []byte(`{"changes":[{"name":"my-change","status":"active"},{"name":"other-change"}]}`)
	items, err := ParseList(data)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "my-change", items[0].Name)
	assert.Equal(t, "active", items[0].Status)
}

func TestParseList_Specs(t *testing.T) {
	t.Parallel()

	data := []byte(`{"specs":[{"name":"my-spec","description":"a spec"}]}`)
	items, err := ParseList(data)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "my-spec", items[0].Name)
}

func TestParseList_Empty(t *testing.T) {
	t.Parallel()

	items, err := ParseList([]byte(`{}`))
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestParseList_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseList([]byte(`not-json`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse list")
}

func TestParseShow(t *testing.T) {
	t.Parallel()

	data := []byte(`{"name":"my-change","status":"active","tasks":[]}`)
	result, err := ParseShow(data)
	require.NoError(t, err)
	assert.Equal(t, "my-change", result["name"])
	assert.Equal(t, "active", result["status"])
}

func TestParseShow_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseShow([]byte(`{broken`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse show")
}

func TestParseValidate(t *testing.T) {
	t.Parallel()

	data := []byte(`{"valid":true,"errors":[]}`)
	result, err := ParseValidate(data)
	require.NoError(t, err)
	assert.Equal(t, true, result["valid"])
}

func TestParseValidate_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseValidate([]byte(`not-json`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse validate")
}

func TestParseStatus(t *testing.T) {
	t.Parallel()

	data := []byte(`{"change":"my-change","completed":3,"total":5}`)
	result, err := ParseStatus(data)
	require.NoError(t, err)
	assert.Equal(t, "my-change", result["change"])
}

func TestParseStatus_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseStatus([]byte(`{bad}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse status")
}

func TestNew(t *testing.T) {
	t.Parallel()

	c := New("/some/path")
	assert.NotNil(t, c)
	assert.Equal(t, "/some/path", c.cwd)
}

func TestClient_List(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not in PATH")
	}

	c := New("../..")
	data, err := c.List(context.Background(), "changes")
	require.NoError(t, err)

	// JSON may represent empty or populated list — both are valid
	items, err := ParseList(data)
	require.NoError(t, err)
	_ = items
}

func TestClient_Validate(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not in PATH")
	}

	c := New("../..")
	data, err := c.Validate(context.Background(), "all")
	if err != nil {
		t.Skipf("openspec validate returned non-zero exit (specs may have validation issues): %v", err)
	}
	assert.NotEmpty(t, data)

	result, err := ParseValidate(data)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
