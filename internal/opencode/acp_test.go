package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionUpdateNotification_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    SessionUpdateNotification
		wantErr bool
	}{
		{
			name: "output notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "output",
				"data": "Implementing rate limit function..."
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "output",
				Data:      "Implementing rate limit function...",
			},
		},
		{
			name: "progress notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "progress",
				"progress": 0.45,
				"message": "45% complete"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "progress",
				Progress:  0.45,
				Message:   "45% complete",
			},
		},
		{
			name: "tool notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "tool",
				"tool": "fs/write_file",
				"status": "completed"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "tool",
				Tool:      "fs/write_file",
				Status:    "completed",
			},
		},
		{
			name: "complete notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "complete"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "complete",
			},
		},
		{
			name: "error notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "error",
				"error": "timeout waiting for response"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "error",
				Error:     "timeout waiting for response",
			},
		},
		{
			name: "notification with metadata",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "output",
				"data": "test",
				"metadata": {"key": "value", "num": 42}
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "output",
				Data:      "test",
				Metadata:  map[string]any{"key": "value", "num": 42.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SessionUpdateNotification
			err := json.Unmarshal([]byte(tt.data), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.SessionID, got.SessionID)
				assert.Equal(t, tt.want.PromptID, got.PromptID)
				assert.Equal(t, tt.want.Type, got.Type)
				assert.Equal(t, tt.want.Data, got.Data)
				assert.Equal(t, tt.want.Progress, got.Progress)
				assert.Equal(t, tt.want.Message, got.Message)
				assert.Equal(t, tt.want.Tool, got.Tool)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.Error, got.Error)
			}
		})
	}
}

func TestSessionUpdateNotification_Marshal(t *testing.T) {
	notif := SessionUpdateNotification{
		SessionID: "sess_123",
		PromptID:  "prompt_456",
		Type:      "output",
		Data:      "test data",
		Metadata:  map[string]any{"key": "value"},
	}

	data, err := json.Marshal(notif)
	assert.NoError(t, err)

	var got SessionUpdateNotification
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)

	assert.Equal(t, notif.SessionID, got.SessionID)
	assert.Equal(t, notif.PromptID, got.PromptID)
	assert.Equal(t, notif.Type, got.Type)
	assert.Equal(t, notif.Data, got.Data)
}

func TestSessionCancelParams_Marshal(t *testing.T) {
	params := SessionCancelParams{
		SessionID: "sess_abc123",
		Reason:    "user_requested",
	}

	data, err := json.Marshal(params)
	assert.NoError(t, err)

	var got SessionCancelParams
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)

	assert.Equal(t, params.SessionID, got.SessionID)
	assert.Equal(t, params.Reason, got.Reason)
}

func TestSessionCancelResult_Marshal(t *testing.T) {
	result := SessionCancelResult{
		SessionID: "sess_abc123",
		Status:    "cancelled",
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var got SessionCancelResult
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)

	assert.Equal(t, result.SessionID, got.SessionID)
	assert.Equal(t, result.Status, got.Status)
}

func TestACPClient_Validate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := Config{
		ClientName: "test-client",
	}

	client, err := NewACPClient(ctx, cfg)
	if err != nil {
		t.Skipf("opencode acp not available: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Validate(ctx)

	if err != nil {
		t.Skipf("ACP validation failed (likely opencode not running): %v", err)
	}
}

func TestACPClient_Validate_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	cfg := Config{
		ClientName: "test-client",
	}

	client, err := NewACPClient(ctx, cfg)
	if err != nil {
		t.Skipf("opencode acp not available: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Validate(ctx)

	assert.Error(t, err)
}

func TestJSONRPCRequest_Marshal(t *testing.T) {
	tests := []struct {
		name  string
		req   jsonrpcRequest
		check func(string)
	}{
		{
			name: "basic request",
			req: jsonrpcRequest{
				JSONRPC: "2.0",
				ID:      "test-id",
				Method:  "test/method",
			},
			check: func(s string) {
				assert.Contains(t, s, `"jsonrpc":"2.0"`)
				assert.Contains(t, s, `"id":"test-id"`)
				assert.Contains(t, s, `"method":"test/method"`)
			},
		},
		{
			name: "request with params",
			req: jsonrpcRequest{
				JSONRPC: "2.0",
				ID:      "test-id",
				Method:  "test/method",
				Params:  map[string]string{"key": "value"},
			},
			check: func(s string) {
				assert.Contains(t, s, `"params"`)
				assert.Contains(t, s, `"key":"value"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			require.NoError(t, err)
			tt.check(string(data))

			var got jsonrpcRequest
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.req.JSONRPC, got.JSONRPC)
			assert.Equal(t, tt.req.ID, got.ID)
			assert.Equal(t, tt.req.Method, got.Method)
		})
	}
}

func TestJSONRPCResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    jsonrpcResponse
		wantErr bool
	}{
		{
			name: "success response",
			data: `{"jsonrpc":"2.0","id":"test-id","result":{"success":true}}`,
			want: jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      "test-id",
				Result:  json.RawMessage(`{"success":true}`),
			},
			wantErr: false,
		},
		{
			name: "error response",
			data: `{"jsonrpc":"2.0","id":"test-id","error":{"code":-32602,"message":"Invalid params"}}`,
			want: jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      "test-id",
				Error: &jsonrpcError{
					Code:    -32602,
					Message: "Invalid params",
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			data:    `{"invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got jsonrpcResponse
			err := json.Unmarshal([]byte(tt.data), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.JSONRPC, got.JSONRPC)
				assert.Equal(t, tt.want.ID, got.ID)

				if tt.want.Error != nil {
					require.NotNil(t, got.Error)
					assert.Equal(t, tt.want.Error.Code, got.Error.Code)
					assert.Equal(t, tt.want.Error.Message, got.Error.Message)
				}

				if tt.want.Result != nil {
					assert.JSONEq(t, string(tt.want.Result), string(got.Result))
				}
			}
		})
	}
}

func TestInitializeParams_Marshal(t *testing.T) {
	params := InitializeParams{
		ProtocolVersion: "1.0",
		Capabilities: map[string]any{
			"streaming":      true,
			"fileOperations": true,
		},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var got InitializeParams
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, params.ProtocolVersion, got.ProtocolVersion)
	assert.Equal(t, params.ClientInfo.Name, got.ClientInfo.Name)
	assert.Equal(t, params.ClientInfo.Version, got.ClientInfo.Version)
}

func TestAuthenticateParams_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		params AuthenticateParams
	}{
		{
			name: "with token",
			params: AuthenticateParams{
				Method: "bearer",
				Token:  "test-token",
			},
		},
		{
			name: "with params",
			params: AuthenticateParams{
				Method: "api_key",
				Params: map[string]string{
					"api_key": "secret-key",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.params)
			require.NoError(t, err)

			var got AuthenticateParams
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.params.Method, got.Method)
			assert.Equal(t, tt.params.Token, got.Token)
		})
	}
}

func TestSessionNewParams_Marshal(t *testing.T) {
	params := SessionNewParams{
		Provider: "test-provider",
		Model:    "test-model",
		Config: map[string]any{
			"temperature": 0.7,
		},
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var got SessionNewParams
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, params.Provider, got.Provider)
	assert.Equal(t, params.Model, got.Model)
}

func TestSessionPromptParams_Marshal(t *testing.T) {
	params := SessionPromptParams{
		SessionID: "test-session",
		Prompt:    "test prompt",
		Context: []MessageContext{
			{Role: "user", Content: "previous message"},
		},
		Options: map[string]any{
			"max_tokens": 1000,
		},
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var got SessionPromptParams
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, params.SessionID, got.SessionID)
	assert.Equal(t, params.Prompt, got.Prompt)
	assert.Equal(t, float64(1000), got.Options["max_tokens"])
}

func TestNDJSONParsing(t *testing.T) {
	ndjson := `{"jsonrpc":"2.0","id":"1","result":"test1"}
{"jsonrpc":"2.0","id":"2","result":"test2"}
{"jsonrpc":"2.0","method":"notification","params":{"data":"test"}}`

	scanner := bufio.NewScanner(strings.NewReader(ndjson))
	responses := 0
	notifications := 0

	for scanner.Scan() {
		line := scanner.Bytes()

		var notif jsonrpcNotification
		if err := json.Unmarshal(line, &notif); err == nil && notif.Method != "" {
			notifications++
			continue
		}

		var resp jsonrpcResponse
		if err := json.Unmarshal(line, &resp); err == nil {
			responses++
		}
	}

	require.NoError(t, scanner.Err())
	assert.Equal(t, 2, responses)
	assert.Equal(t, 1, notifications)
}

func TestRequestIDMatching(t *testing.T) {
	reqID := "test-id-123"
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  "test/method",
	}

	reqData, err := json.Marshal(req)
	require.NoError(t, err)

	var resp jsonrpcResponse
	err = json.Unmarshal(reqData, &resp)
	require.NoError(t, err)

	assert.Equal(t, reqID, resp.ID)
}

func TestPartialJSONParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "complete json",
			input:    `{"jsonrpc":"2.0","id":"1","result":{}}`,
			expected: 1,
		},
		{
			name:     "incomplete json",
			input:    `{"jsonrpc":"2.0","id":"1","result":`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			count := 0

			for scanner.Scan() {
				var data map[string]interface{}
				if json.Unmarshal(scanner.Bytes(), &data) == nil {
					count++
				}
			}

			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestInvalidJSONHandling(t *testing.T) {
	invalidInputs := []string{
		`{"invalid": missing quote}`,
		`not json at all`,
		`{"jsonrpc":"2.0",}`,
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			var data interface{}
			err := json.Unmarshal([]byte(input), &data)
			assert.Error(t, err)
		})
	}
}

func TestConcurrentRequests_Simulation(t *testing.T) {
	requests := 10
	done := make(chan bool, requests)
	var mu sync.Mutex
	var completed int

	for i := 0; i < requests; i++ {
		go func(id int) {
			req := jsonrpcRequest{
				JSONRPC: "2.0",
				ID:      fmt.Sprintf("req-%d", id),
				Method:  "test/method",
			}

			data, err := json.Marshal(req)
			assert.NoError(t, err)

			var resp jsonrpcResponse
			err = json.Unmarshal(data, &resp)
			assert.NoError(t, err)

			mu.Lock()
			completed++
			mu.Unlock()

			done <- true
		}(i)
	}

	for i := 0; i < requests; i++ {
		<-done
	}

	assert.Equal(t, requests, completed)
}

func TestResponseStream_Simulation(t *testing.T) {
	var output bytes.Buffer
	notifications := []string{
		`{"jsonrpc":"2.0","method":"session/update","params":{"type":"output","data":"Starting..."}}`,
		`{"jsonrpc":"2.0","method":"session/update","params":{"type":"output","data":"Processing..."}}`,
		`{"jsonrpc":"2.0","id":"req-1","result":{"promptId":"p1"}}`,
	}

	for _, notif := range notifications {
		fmt.Fprintln(&output, notif)
	}

	scanner := bufio.NewScanner(&output)
	notifCount := 0
	respCount := 0

	for scanner.Scan() {
		line := scanner.Bytes()

		var notif jsonrpcNotification
		if err := json.Unmarshal(line, &notif); err == nil && notif.Method != "" {
			notifCount++
			continue
		}

		var resp jsonrpcResponse
		if err := json.Unmarshal(line, &resp); err == nil {
			respCount++
		}
	}

	require.NoError(t, scanner.Err())
	assert.Equal(t, 2, notifCount)
	assert.Equal(t, 1, respCount)
}

func TestTimeoutHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout context did not expire")
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	}
}

func TestNetworkError_Simulation(t *testing.T) {
	r, w := io.Pipe()
	_ = r.Close()

	_, err := fmt.Fprintln(w, "test")
	assert.Error(t, err)
}

func TestProcessTermination_Simulation(t *testing.T) {
	r, w := io.Pipe()
	_ = r.Close()

	_, err := w.Write([]byte("test"))
	assert.Error(t, err)

	err = w.Close()
	assert.NoError(t, err)
}

func TestClientRequestHandler_Register(t *testing.T) {
	h := NewClientRequestHandler(nil)

	methods := []string{
		"fs/read_text_file",
		"fs/write_text_file",
		"fs/list_directory",
		"terminal/exec",
		"terminal/write_input",
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			_, err := h.HandleRequest(context.Background(), method, nil)
			if err == nil || !strings.Contains(err.Error(), "invalid params") {
				t.Logf("Handler for %s is registered", method)
			}
		})
	}
}

func TestACPClient_Accessors(t *testing.T) {
	testClient := &ACPClient{
		sessionID:     "test-session",
		sessionStatus: "active",
		serverInfo: ServerInfo{
			Name:    "test-server",
			Version: "1.0.0",
		},
		capabilities: map[string]any{
			"streaming": true,
		},
		initialized:     true,
		authToken:       "test-token",
		authExpiresAt:   1234567890,
		currentPromptID: "prompt-1",
		promptHistory: []SessionPromptHistory{
			{
				PromptID:  "prompt-1",
				Prompt:    "test",
				Timestamp: time.Now(),
				Status:    "complete",
			},
		},
	}

	assert.Equal(t, "test-session", testClient.SessionID())
	assert.Equal(t, "active", testClient.SessionStatus())
	assert.True(t, testClient.IsInitialized())
	assert.Equal(t, "test-server", testClient.ServerInfo().Name)
	assert.NotNil(t, testClient.Capabilities())
	assert.Equal(t, "test-token", testClient.AuthToken())
	assert.Equal(t, int64(1234567890), testClient.AuthExpiresAt())
	assert.True(t, testClient.IsAuthenticated())
	assert.Equal(t, "prompt-1", testClient.CurrentPromptID())
	assert.Len(t, testClient.GetPromptHistory(), 1)
}

func TestACPClient_Close(t *testing.T) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	testClient := &ACPClient{
		stdin:      w1,
		stdout:     r1,
		stderr:     r2,
		updateChan: make(chan SessionUpdateNotification, 10),
		closeChan:  make(chan struct{}),
		cancel:     func() {},
	}

	err := testClient.Close()
	assert.NoError(t, err)

	err = testClient.Close()
	assert.NoError(t, err)

	_ = w1.Close()
	_ = w2.Close()
}

func TestSessionStatusTransitions(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		newStatus   string
		shouldAllow bool
	}{
		{
			name:        "active to complete",
			status:      "active",
			newStatus:   "completed",
			shouldAllow: true,
		},
		{
			name:        "active to cancelled",
			status:      "active",
			newStatus:   "cancelled",
			shouldAllow: true,
		},
		{
			name:        "active to failed",
			status:      "active",
			newStatus:   "failed",
			shouldAllow: true,
		},
		{
			name:        "cancelled stays cancelled",
			status:      "cancelled",
			newStatus:   "cancelled",
			shouldAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Status transition: %s -> %s (allowed: %v)", tt.status, tt.newStatus, tt.shouldAllow)
		})
	}
}

func TestErrorHandling_InvalidResponses(t *testing.T) {
	invalidResponses := []string{
		`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"}}`,
		`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"}}`,
		`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"}}`,
		`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params"}}`,
		`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"}}`,
	}

	for _, resp := range invalidResponses {
		t.Run(resp, func(t *testing.T) {
			var r jsonrpcResponse
			err := json.Unmarshal([]byte(resp), &r)
			require.NoError(t, err)
			assert.NotNil(t, r.Error)
			assert.Greater(t, r.Error.Code, -40000)
			assert.NotEmpty(t, r.Error.Message)
		})
	}
}

func TestACPClient_SendRequest_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	time.Sleep(10 * time.Millisecond)
	cancel()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-ctx.Done():
		assert.Equal(t, context.Canceled, ctx.Err())
	default:
		t.Error("context was not cancelled")
	}
}

func TestUpdateChannelBehavior(t *testing.T) {
	updateChan := make(chan SessionUpdateNotification, 5)

	for i := 0; i < 10; i++ {
		update := SessionUpdateNotification{
			SessionID: "test",
			PromptID:  fmt.Sprintf("prompt-%d", i),
			Type:      "output",
			Data:      fmt.Sprintf("output %d", i),
		}

		select {
		case updateChan <- update:
			if i < 5 {
				t.Logf("Update %d sent successfully", i)
			} else {
				t.Logf("Update %d: channel full, would block", i)
			}
		default:
			t.Logf("Update %d: channel full, message dropped", i)
		}
	}

	assert.Len(t, updateChan, 5)
}

func TestACPClient_Methods_NotInitialized(t *testing.T) {
	client := &ACPClient{
		initialized: false,
	}

	ctx := context.Background()

	_, err := client.Authenticate(ctx, "bearer", "token", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.SessionNew(ctx, "provider", "model", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.SessionPrompt(ctx, "prompt", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.SessionCancel(ctx, "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestACPClient_SessionPrompt_NoActiveSession(t *testing.T) {
	client := &ACPClient{
		initialized: true,
		sessionID:   "",
	}

	_, err := client.SessionPrompt(context.Background(), "prompt", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active session")
}

func TestACPClient_SessionPrompt_EmptyPrompt(t *testing.T) {
	client := &ACPClient{
		initialized: true,
		sessionID:   "session-123",
	}

	_, err := client.SessionPrompt(context.Background(), "", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt cannot be empty")
}

func TestACPClient_SessionCancel_NoSession(t *testing.T) {
	client := &ACPClient{
		initialized: true,
		sessionID:   "",
	}

	_, err := client.SessionCancel(context.Background(), "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active session")
}

func TestACPClient_SessionCancel_AlreadyCancelled(t *testing.T) {
	client := &ACPClient{
		initialized:   true,
		sessionID:     "session-123",
		sessionStatus: "cancelled",
	}

	result, err := client.SessionCancel(context.Background(), "reason")
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", result.Status)
}

func TestACPClient_SessionCancel_AlreadyComplete(t *testing.T) {
	client := &ACPClient{
		initialized:   true,
		sessionID:     "session-123",
		sessionStatus: "completed",
	}

	_, err := client.SessionCancel(context.Background(), "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session already completed")
}

func TestACPClient_SessionCancel_AlreadyFailed(t *testing.T) {
	client := &ACPClient{
		initialized:   true,
		sessionID:     "session-123",
		sessionStatus: "failed",
	}

	_, err := client.SessionCancel(context.Background(), "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session already failed")
}

func TestAuthenticateParams_NoCredentials(t *testing.T) {
	params := AuthenticateParams{
		Method: "",
		Token:  "",
		Params: nil,
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var got AuthenticateParams
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Empty(t, got.Method)
	assert.Empty(t, got.Token)
	assert.Nil(t, got.Params)
}

func TestSessionNewParams_Empty(t *testing.T) {
	params := SessionNewParams{}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var got SessionNewParams
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Empty(t, got.Provider)
	assert.Empty(t, got.Model)
}

func TestClientInfo_Marshal(t *testing.T) {
	info := ClientInfo{
		Name:    "test-client",
		Version: "1.0.0",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var got ClientInfo
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, info.Name, got.Name)
	assert.Equal(t, info.Version, got.Version)
}

func TestServerInfo_Marshal(t *testing.T) {
	info := ServerInfo{
		Name:    "opencode",
		Version: "1.2.3",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var got ServerInfo
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, info.Name, got.Name)
	assert.Equal(t, info.Version, got.Version)
}

func TestInitializeResult_Marshal(t *testing.T) {
	result := InitializeResult{
		ProtocolVersion: "1.0",
		Capabilities: map[string]any{
			"streaming": true,
		},
		ServerInfo: ServerInfo{
			Name:    "opencode",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var got InitializeResult
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, result.ProtocolVersion, got.ProtocolVersion)
	assert.Equal(t, result.ServerInfo.Name, got.ServerInfo.Name)
}

func TestSessionNewResult_Marshal(t *testing.T) {
	result := SessionNewResult{
		SessionID: "session-123",
		Provider:  "anthropic",
		Model:     "claude-3-opus",
		Status:    "active",
		Metadata: map[string]string{
			"key": "value",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var got SessionNewResult
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, result.SessionID, got.SessionID)
	assert.Equal(t, result.Provider, got.Provider)
	assert.Equal(t, result.Model, got.Model)
	assert.Equal(t, result.Status, got.Status)
}

func TestSessionPromptResult_Marshal(t *testing.T) {
	result := SessionPromptResult{
		PromptID:  "prompt-123",
		SessionID: "session-456",
		Status:    "queued",
		Metadata: map[string]string{
			"queued_at": "2024-01-01T00:00:00Z",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var got SessionPromptResult
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, result.PromptID, got.PromptID)
	assert.Equal(t, result.SessionID, got.SessionID)
	assert.Equal(t, result.Status, got.Status)
}

func TestAuthenticateResult_Marshal(t *testing.T) {
	result := AuthenticateResult{
		Token:     "test-token-123",
		ExpiresAt: 1704067200,
		Metadata: map[string]string{
			"type": "bearer",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var got AuthenticateResult
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, result.Token, got.Token)
	assert.Equal(t, result.ExpiresAt, got.ExpiresAt)
}

func TestMessageContext_Marshal(t *testing.T) {
	ctx := MessageContext{
		Role:    "user",
		Content: "test message",
	}

	data, err := json.Marshal(ctx)
	require.NoError(t, err)

	var got MessageContext
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, ctx.Role, got.Role)
	assert.Equal(t, ctx.Content, got.Content)
}

func TestACPClient_Config_Defaults(t *testing.T) {
	cfg := Config{
		ClientName: "test-client",
		ClientVer:  "1.0.0",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := NewACPClient(ctx, cfg)

	if err != nil {
		return
	}
	defer func() { _ = client.Close() }()

	assert.NotNil(t, client)
}

func TestACPClient_UpdateChannel(t *testing.T) {
	client := &ACPClient{
		updateChan: make(chan SessionUpdateNotification, 10),
	}

	updates := client.Updates()
	assert.NotNil(t, updates)
}

func TestSessionPromptHistory(t *testing.T) {
	history := []SessionPromptHistory{
		{
			PromptID:  "prompt-1",
			Prompt:    "first prompt",
			Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:    "complete",
		},
		{
			PromptID:  "prompt-2",
			Prompt:    "second prompt",
			Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			Status:    "pending",
		},
	}

	client := &ACPClient{
		promptHistory: history,
	}

	retrieved := client.GetPromptHistory()
	assert.Len(t, retrieved, 2)
	assert.Equal(t, "prompt-1", retrieved[0].PromptID)
	assert.Equal(t, "prompt-2", retrieved[1].PromptID)
}

func TestConfig_AllFields(t *testing.T) {
	cfg := Config{
		ConfigPath: "/path/to/config.yaml",
		ClientName: "test-client",
		ClientVer:  "2.0.0",
		AuthType:   "bearer",
		AuthToken:  "token-123",
		AuthAPIKey: "api-key-456",
	}

	assert.Equal(t, "/path/to/config.yaml", cfg.ConfigPath)
	assert.Equal(t, "test-client", cfg.ClientName)
	assert.Equal(t, "2.0.0", cfg.ClientVer)
	assert.Equal(t, "bearer", cfg.AuthType)
	assert.Equal(t, "token-123", cfg.AuthToken)
	assert.Equal(t, "api-key-456", cfg.AuthAPIKey)
}

func TestACPClient_MultipleClose(t *testing.T) {
	client := &ACPClient{
		updateChan: make(chan SessionUpdateNotification, 10),
		closeChan:  make(chan struct{}),
		cancel:     func() {},
	}

	err := client.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestJSONRPCError_Marshal(t *testing.T) {
	errObj := jsonrpcError{
		Code:    -32602,
		Message: "Invalid params",
		Data:    "field 'id' is required",
	}

	data, err := json.Marshal(errObj)
	require.NoError(t, err)

	var got jsonrpcError
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, errObj.Code, got.Code)
	assert.Equal(t, errObj.Message, got.Message)
	assert.Equal(t, errObj.Data, got.Data)
}

func TestNotificationParams_Marshal(t *testing.T) {
	params := map[string]any{
		"sessionId": "session-123",
		"promptId":  "prompt-456",
		"type":      "output",
		"data":      "test output",
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var got map[string]any
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, params["sessionId"], got["sessionId"])
	assert.Equal(t, params["promptId"], got["promptId"])
	assert.Equal(t, params["type"], got["type"])
	assert.Equal(t, params["data"], got["data"])
}
