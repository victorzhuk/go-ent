package opencode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type jsonrpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type jsonrpcNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    map[string]any    `json:"capabilities"`
	ClientInfo      ClientInfo        `json:"clientInfo"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type AuthenticateParams struct {
	Method string            `json:"method"`
	Token  string            `json:"token,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

type AuthenticateResult struct {
	Token     string            `json:"token,omitempty"`
	ExpiresAt int64             `json:"expiresAt,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type SessionNewParams struct {
	Provider string            `json:"provider"`
	Model    string            `json:"model"`
	Config   map[string]any    `json:"config,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type SessionNewResult struct {
	SessionID string            `json:"sessionId"`
	Provider  string            `json:"provider"`
	Model     string            `json:"model"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type SessionPromptParams struct {
	SessionID string            `json:"sessionId"`
	Prompt    string            `json:"prompt"`
	Context   []MessageContext  `json:"context,omitempty"`
	Options   map[string]any    `json:"options,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type MessageContext struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptContext struct {
	Files []string `json:"files,omitempty"`
	Tools []string `json:"tools,omitempty"`
}

type SessionPromptResult struct {
	PromptID  string            `json:"promptId"`
	SessionID string            `json:"sessionId"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type SessionPromptHistory struct {
	PromptID  string
	Prompt    string
	Timestamp time.Time
	Status    string
}

type SessionCancelParams struct {
	SessionID string `json:"sessionId"`
	Reason    string `json:"reason,omitempty"`
}

type SessionCancelResult struct {
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
}

type SessionUpdateNotification struct {
	SessionID string         `json:"sessionId"`
	PromptID  string         `json:"promptId"`
	Type      string         `json:"type"`
	Data      string         `json:"data,omitempty"`
	Progress  float64        `json:"progress,omitempty"`
	Message   string         `json:"message,omitempty"`
	Tool      string         `json:"tool,omitempty"`
	Status    string         `json:"status,omitempty"`
	Error     string         `json:"error,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type ACPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	logger *slog.Logger

	sessionID     string
	sessionStatus string
	serverInfo    ServerInfo
	capabilities  map[string]any
	initialized   bool

	authToken     string
	authExpiresAt int64
	authMetadata  map[string]string

	currentPromptID string
	updateChan      chan SessionUpdateNotification
	closeChan       chan struct{}
	closed          bool
	promptHistory   []SessionPromptHistory

	requestHandler *ClientRequestHandler
}

type Config struct {
	ConfigPath string
	ClientName string
	ClientVer  string
	AuthType   string
	AuthToken  string
	AuthAPIKey string
}

func NewACPClient(ctx context.Context, cfg Config) (*ACPClient, error) {
	if cfg.ClientVer == "" {
		cfg.ClientVer = "1.0.0"
	}

	ctx, cancel := context.WithCancel(ctx)

	cmd := exec.CommandContext(ctx, "opencode", "acp")
	env := cmd.Env
	if cfg.ConfigPath != "" {
		env = append(env, fmt.Sprintf("OPENCODE_CONFIG=%s", cfg.ConfigPath))
	}
	if cfg.AuthType != "" {
		env = append(env, fmt.Sprintf("OPENCODE_AUTH_TYPE=%s", cfg.AuthType))
	}
	if cfg.AuthToken != "" {
		env = append(env, fmt.Sprintf("OPENCODE_AUTH_TOKEN=%s", cfg.AuthToken))
	}
	if cfg.AuthAPIKey != "" {
		env = append(env, fmt.Sprintf("OPENCODE_AUTH_API_KEY=%s", cfg.AuthAPIKey))
	}
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		cancel()
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		cancel()
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		cancel()
		return nil, fmt.Errorf("start opencode acp: %w", err)
	}

	client := &ACPClient{
		cmd:        cmd,
		stdin:      stdin,
		stdout:     stdout,
		stderr:     stderr,
		ctx:        ctx,
		cancel:     cancel,
		logger:     slog.Default(),
		updateChan: make(chan SessionUpdateNotification, 100),
		closeChan:  make(chan struct{}),
	}

	go client.readStderr()
	go client.readIncomingRequests()

	return client, nil
}

func (c *ACPClient) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return fmt.Errorf("already initialized")
	}

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := InitializeParams{
		ProtocolVersion: "1.0",
		Capabilities: map[string]any{
			"streaming":      true,
			"fileOperations": true,
			"terminal":       true,
		},
		ClientInfo: ClientInfo{
			Name:    "go-ent",
			Version: "0.1.0",
		},
	}

	var result InitializeResult
	if err := c.sendRequest(initCtx, "initialize", params, &result); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	c.serverInfo = result.ServerInfo
	c.capabilities = result.Capabilities
	c.initialized = true

	c.logger.Info("acp client initialized",
		"server", c.serverInfo.Name,
		"version", c.serverInfo.Version,
	)

	return nil
}

func (c *ACPClient) Authenticate(ctx context.Context, method, token string, params map[string]string) (*AuthenticateResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	if method == "" && token == "" && len(params) == 0 {
		return nil, fmt.Errorf("no credentials provided")
	}

	authCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	authParams := AuthenticateParams{
		Method: method,
		Token:  token,
		Params: params,
	}

	var result AuthenticateResult
	if err := c.sendRequest(authCtx, "authenticate", authParams, &result); err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	if result.Token != "" {
		c.authToken = result.Token
	}
	if result.ExpiresAt > 0 {
		c.authExpiresAt = result.ExpiresAt
	}
	if result.Metadata != nil {
		c.authMetadata = result.Metadata
	}

	c.logger.Info("authentication successful", "method", method)

	return &result, nil
}

func (c *ACPClient) SessionNew(ctx context.Context, provider, model string, config map[string]any) (*SessionNewResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	params := SessionNewParams{
		Provider: provider,
		Model:    model,
		Config:   config,
	}

	var result SessionNewResult
	if err := c.sendRequest(ctx, "session/new", params, &result); err != nil {
		return nil, fmt.Errorf("session/new: %w", err)
	}

	c.sessionID = result.SessionID
	c.sessionStatus = result.Status

	c.logger.Info("session created",
		"session_id", result.SessionID,
		"provider", result.Provider,
		"model", result.Model,
	)

	return &result, nil
}

func (c *ACPClient) SessionPrompt(ctx context.Context, prompt string, context []MessageContext, options map[string]any) (*SessionPromptResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	if c.sessionID == "" {
		return nil, fmt.Errorf("no active session")
	}

	if prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	params := SessionPromptParams{
		SessionID: c.sessionID,
		Prompt:    prompt,
		Context:   context,
		Options:   options,
	}

	var result SessionPromptResult
	if err := c.sendRequest(ctx, "session/prompt", params, &result); err != nil {
		return nil, fmt.Errorf("session/prompt: %w", err)
	}

	c.currentPromptID = result.PromptID

	historyEntry := SessionPromptHistory{
		PromptID:  result.PromptID,
		Prompt:    prompt,
		Timestamp: time.Now(),
		Status:    result.Status,
	}
	c.promptHistory = append(c.promptHistory, historyEntry)

	c.logger.Info("prompt sent",
		"prompt_id", result.PromptID,
		"session_id", result.SessionID,
		"status", result.Status,
	)

	return &result, nil
}

func (c *ACPClient) SessionCancel(ctx context.Context, reason string) (*SessionCancelResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	if c.sessionID == "" {
		return nil, fmt.Errorf("no active session")
	}

	if c.sessionStatus == "cancelled" {
		return &SessionCancelResult{
			SessionID: c.sessionID,
			Status:    "cancelled",
		}, nil
	}

	if c.sessionStatus == "completed" || c.sessionStatus == "failed" {
		return nil, fmt.Errorf("session already %s", c.sessionStatus)
	}

	params := SessionCancelParams{
		SessionID: c.sessionID,
		Reason:    reason,
	}

	var result SessionCancelResult
	if err := c.sendRequest(ctx, "session/cancel", params, &result); err != nil {
		return nil, fmt.Errorf("session/cancel: %w", err)
	}

	c.sessionStatus = "cancelled"
	c.currentPromptID = ""
	c.promptHistory = nil

	c.logger.Info("session cancelled",
		"session_id", result.SessionID,
		"reason", reason,
	)

	return &result, nil
}

func (c *ACPClient) Updates() <-chan SessionUpdateNotification {
	return c.updateChan
}

func (c *ACPClient) sendRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
	reqID := uuid.Must(uuid.NewV7()).String()

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  params,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	if _, err := fmt.Fprintln(c.stdin, string(reqData)); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	responseCh := make(chan *jsonrpcResponse, 1)
	errorCh := make(chan error, 1)

	go func() {
		resp, err := c.readResponse(reqID)
		if err != nil {
			errorCh <- err
			return
		}
		responseCh <- resp
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("request cancelled: %w", ctx.Err())
	case err := <-errorCh:
		return err
	case resp := <-responseCh:
		if resp.Error != nil {
			return fmt.Errorf("jsonrpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		if result != nil && len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, result); err != nil {
				return fmt.Errorf("unmarshal result: %w", err)
			}
		}

		return nil
	}
}

func (c *ACPClient) readResponse(expectedID string) (*jsonrpcResponse, error) {
	scanner := bufio.NewScanner(c.stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var notif jsonrpcNotification
		if err := json.Unmarshal(line, &notif); err == nil && notif.JSONRPC == "2.0" && notif.Method != "" {
			c.handleNotification(notif)
			continue
		}

		var resp jsonrpcResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			c.logger.Warn("invalid jsonrpc message", "error", err, "line", string(line))
			continue
		}

		if resp.ID == expectedID {
			return &resp, nil
		}

		c.logger.Warn("unexpected response id", "expected", expectedID, "got", resp.ID)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return nil, io.EOF
}

func (c *ACPClient) handleNotification(notif jsonrpcNotification) {
	switch notif.Method {
	case "session/update":
		var update SessionUpdateNotification
		data, err := json.Marshal(notif.Params)
		if err == nil {
			if err := json.Unmarshal(data, &update); err == nil {
				c.mu.Lock()
				if update.SessionID == c.sessionID {
					if update.Status == "complete" {
						c.sessionStatus = "completed"
					} else if update.Status == "error" {
						c.sessionStatus = "failed"
					} else if update.Status == "cancelled" {
						c.sessionStatus = "cancelled"
					}
				}
				c.mu.Unlock()
				select {
				case c.updateChan <- update:
				default:
					c.logger.Warn("update channel full, dropping notification")
				}
			}
		}
	default:
		c.logger.Debug("unhandled notification", "method", notif.Method)
	}
}

func (c *ACPClient) readStderr() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		c.logger.Debug("opencode stderr", "line", scanner.Text())
	}
}

func (c *ACPClient) readIncomingRequests() {
	scanner := bufio.NewScanner(c.stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			c.logger.Warn("invalid incoming request", "error", err, "line", string(line))
			continue
		}

		go c.handleIncomingRequest(req)
	}

	if err := scanner.Err(); err != nil {
		c.logger.Error("read incoming requests", "error", err)
	}
}

func (c *ACPClient) handleIncomingRequest(req jsonrpcRequest) {
	if c.requestHandler == nil {
		c.requestHandler = NewClientRequestHandler(c.logger)
	}

	result, err := c.requestHandler.HandleRequest(c.ctx, req.Method, nil)
	if err != nil {
		params, _ := req.Params.(json.RawMessage)
		result, err = c.requestHandler.HandleRequest(c.ctx, req.Method, params)
	}

	resp := jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	if err != nil {
		resp.Error = &jsonrpcError{
			Code:    -32603,
			Message: err.Error(),
		}
		c.logger.Warn("request failed", "method", req.Method, "error", err)
	} else {
		data, err := json.Marshal(result)
		if err != nil {
			resp.Error = &jsonrpcError{
				Code:    -32603,
				Message: fmt.Sprintf("marshal result: %v", err),
			}
		} else {
			resp.Result = data
		}
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		c.logger.Error("marshal response", "error", err)
		return
	}

	if _, err := fmt.Fprintln(c.stdin, string(respData)); err != nil {
		c.logger.Error("write response", "error", err)
	}
}

func (c *ACPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.closeChan)

	if c.stdin != nil {
		_ = c.stdin.Close()
	}

	if c.updateChan != nil {
		close(c.updateChan)
	}

	c.sessionID = ""
	c.sessionStatus = ""
	c.currentPromptID = ""
	c.promptHistory = nil
	c.cancel()

	return nil
}

func (c *ACPClient) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

func (c *ACPClient) SessionStatus() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionStatus
}

func (c *ACPClient) IsInitialized() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.initialized
}

func (c *ACPClient) ServerInfo() ServerInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serverInfo
}

func (c *ACPClient) Capabilities() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.capabilities
}

func (c *ACPClient) AuthToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.authToken
}

func (c *ACPClient) AuthExpiresAt() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.authExpiresAt
}

func (c *ACPClient) IsAuthenticated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.authToken != ""
}

func (c *ACPClient) GetPromptHistory() []SessionPromptHistory {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.promptHistory
}

func (c *ACPClient) CurrentPromptID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentPromptID
}

func (c *ACPClient) Validate(ctx context.Context) error {
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if !c.initialized {
		if err := c.Initialize(validateCtx); err != nil {
			return fmt.Errorf("acp initialization failed: %w", err)
		}
	}

	return nil
}
