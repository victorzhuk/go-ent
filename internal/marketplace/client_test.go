package marketplace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Defaults(t *testing.T) {
	t.Parallel()

	client := NewClient()

	assert.Equal(t, DefaultBaseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, Timeout, client.httpClient.Timeout)
}

func TestNewClientWithURL_CustomURL(t *testing.T) {
	t.Parallel()

	customURL := "http://custom.url"
	client := NewClientWithURL(customURL)

	assert.Equal(t, customURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, Timeout, client.httpClient.Timeout)
}

func TestClientSearch_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plugins/search", r.URL.Path)
		assert.Equal(t, "test", r.URL.Query().Get("q"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"plugins": [{"name": "test-plugin", "version": "1.0.0", "description": "Test plugin", "author": "Test"}], "total": 1}`))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	plugins, err := client.Search(context.Background(), "test", SearchOptions{})

	require.NoError(t, err)
	require.Len(t, plugins, 1)
	assert.Equal(t, "test-plugin", plugins[0].Name)
	assert.Equal(t, "1.0.0", plugins[0].Version)
}

func TestClientSearch_WithOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plugins/search", r.URL.Path)
		assert.Equal(t, "test", r.URL.Query().Get("q"))
		assert.Equal(t, "tools", r.URL.Query().Get("category"))
		assert.Equal(t, "testauthor", r.URL.Query().Get("author"))
		assert.Equal(t, "downloads", r.URL.Query().Get("sort_by"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"plugins": [], "total": 0}`))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	plugins, err := client.Search(context.Background(), "test", SearchOptions{
		Category: "tools",
		Author:   "testauthor",
		SortBy:   "downloads",
		Limit:    10,
	})

	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestClientSearch_EmptyResults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"plugins": [], "total": 0}`))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	plugins, err := client.Search(context.Background(), "nonexistent", SearchOptions{})

	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestClientSearch_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Search(context.Background(), "test", SearchOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do request")
}

func TestClientSearch_Non200Status(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("plugin not found"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Search(context.Background(), "test", SearchOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search failed")
	assert.Contains(t, err.Error(), "404")
}

func TestClientSearch_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Search(context.Background(), "test", SearchOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestClientDownload_Success(t *testing.T) {
	t.Parallel()

	testData := []byte("plugin data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plugins/test-plugin/versions/1.0.0/download", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(testData)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	data, err := client.Download(context.Background(), "test-plugin", "1.0.0")

	require.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestClientDownload_SizeLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, MaxDownloadSize))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Download(context.Background(), "test-plugin", "1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestClientDownload_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Download(context.Background(), "test-plugin", "1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do request")
}

func TestClientDownload_Non200Status(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("version not found"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Download(context.Background(), "test-plugin", "1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed")
	assert.Contains(t, err.Error(), "404")
}

func TestClientDownload_ReadError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	client.httpClient.Timeout = 1 * time.Nanosecond

	_, err := client.Download(context.Background(), "test-plugin", "1.0.0")

	assert.Error(t, err)
}

func TestClientGetPlugin_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plugins/test-plugin", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"name": "test-plugin",
			"version": "1.0.0",
			"description": "Test plugin",
			"author": "Test",
			"category": "tools",
			"downloads": 100,
			"rating": 4.5,
			"tags": ["test", "demo"],
			"skills_count": 5,
			"agents_count": 2,
			"rules_count": 10
		}`))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	plugin, err := client.GetPlugin(context.Background(), "test-plugin")

	require.NoError(t, err)
	require.NotNil(t, plugin)
	assert.Equal(t, "test-plugin", plugin.Name)
	assert.Equal(t, "1.0.0", plugin.Version)
	assert.Equal(t, "Test plugin", plugin.Description)
	assert.Equal(t, "Test", plugin.Author)
	assert.Equal(t, "tools", plugin.Category)
	assert.Equal(t, 100, plugin.Downloads)
	assert.Equal(t, 4.5, plugin.Rating)
	assert.Equal(t, []string{"test", "demo"}, plugin.Tags)
	assert.Equal(t, 5, plugin.Skills)
	assert.Equal(t, 2, plugin.Agents)
	assert.Equal(t, 10, plugin.Rules)
}

func TestClientGetPlugin_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.GetPlugin(context.Background(), "test-plugin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do request")
}

func TestClientGetPlugin_Non200Status(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("plugin not found"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.GetPlugin(context.Background(), "test-plugin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get plugin failed")
	assert.Contains(t, err.Error(), "404")
}

func TestClientGetPlugin_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.GetPlugin(context.Background(), "test-plugin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestClientSearch_QueryEscaping(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		assert.Equal(t, "test query with spaces", query)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"plugins": [], "total": 0}`))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Search(context.Background(), "test query with spaces", SearchOptions{})

	require.NoError(t, err)
}

func TestClientDownload_ContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClientWithURL(server.URL)
	_, err := client.Download(ctx, "test-plugin", "1.0.0")

	assert.Error(t, err)
}

func TestClientSearch_ResponseBodyLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(make([]byte, 2*MaxErrorBodySize))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.Search(context.Background(), "test", SearchOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search failed")
	assert.Contains(t, err.Error(), "500")

	bodyLength := len(err.Error())
	assert.LessOrEqual(t, bodyLength, len("search failed: status 500, body: ")+MaxErrorBodySize)
}
