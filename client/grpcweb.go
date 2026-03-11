package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// grpcWebConn implements grpc.ClientConnInterface by sending gRPC-Web
// requests over HTTP/1.1. It supports both unary and server-streaming RPCs.
type grpcWebConn struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	timeout    time.Duration
}

// newGRPCWebConn creates a new gRPC-Web connection to the given base URL.
// The base URL should include the scheme (http:// or https://).
func newGRPCWebConn(baseURL string, headers map[string]string, timeout time.Duration) *grpcWebConn {
	// Trim trailing slash for consistency.
	baseURL = strings.TrimRight(baseURL, "/")
	return &grpcWebConn{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		headers:    headers,
		timeout:    timeout,
	}
}

// grpcWebFrame encodes a protobuf message into gRPC-Web frame format:
// 1 byte flag (0x00 for data) + 4 bytes big-endian length + payload.
func grpcWebFrame(msg []byte) []byte {
	frame := make([]byte, 5+len(msg))
	frame[0] = 0x00 // data frame
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(msg)))
	copy(frame[5:], msg)
	return frame
}

// parseGRPCWebFrame reads one frame from the reader.
// Returns (flag, payload, error). Flag 0x00 = data, 0x80 = trailers.
func parseGRPCWebFrame(r io.Reader) (byte, []byte, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, err
	}
	flag := header[0]
	length := binary.BigEndian.Uint32(header[1:5])
	if length > 4*1024*1024 { // 4 MiB sanity limit
		return 0, nil, fmt.Errorf("grpc-web: frame too large: %d bytes", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, fmt.Errorf("grpc-web: incomplete frame payload: %w", err)
	}
	return flag, payload, nil
}

// parseTrailers parses gRPC trailers from a trailer frame payload.
// Trailers are encoded as HTTP header lines: "key: value\r\n".
func parseTrailers(data []byte) map[string]string {
	trailers := make(map[string]string)
	lines := strings.Split(string(data), "\r\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			trailers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return trailers
}

// statusFromTrailers extracts a gRPC status from trailer values.
func statusFromTrailers(trailers map[string]string) *status.Status {
	codeStr, ok := trailers["grpc-status"]
	if !ok {
		return status.New(codes.OK, "")
	}
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return status.New(codes.Unknown, "invalid grpc-status: "+codeStr)
	}
	msg := trailers["grpc-message"]
	return status.New(codes.Code(code), msg)
}

// buildRequest creates an HTTP request for a gRPC-Web call.
func (c *grpcWebConn) buildRequest(ctx context.Context, method string, body []byte) (*http.Request, error) {
	url := c.baseURL + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("Accept", "application/grpc-web+proto")
	req.Header.Set("X-Grpc-Web", "1")
	req.Header.Set("X-User-Agent", "grpc-web-go/1.0")

	// Apply custom headers.
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Apply metadata from context (outgoing metadata).
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		for k, vals := range md {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
	}

	return req, nil
}

// Invoke implements grpc.ClientConnInterface for unary RPCs.
func (c *grpcWebConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	ctx, cancel := applyTimeout(ctx, c.timeout)
	defer cancel()

	reqMsg, ok := args.(proto.Message)
	if !ok {
		return fmt.Errorf("grpc-web: args must implement proto.Message")
	}
	replyMsg, ok := reply.(proto.Message)
	if !ok {
		return fmt.Errorf("grpc-web: reply must implement proto.Message")
	}

	// Marshal and frame the request.
	reqBytes, err := proto.Marshal(reqMsg)
	if err != nil {
		return fmt.Errorf("grpc-web: marshal request: %w", err)
	}
	body := grpcWebFrame(reqBytes)

	req, err := c.buildRequest(ctx, method, body)
	if err != nil {
		return fmt.Errorf("grpc-web: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("grpc-web: http request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP-level errors.
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return status.Errorf(codes.Unavailable, "grpc-web: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Check for grpc-status in response headers (early error).
	if grpcStatus := resp.Header.Get("Grpc-Status"); grpcStatus != "" && grpcStatus != "0" {
		code, _ := strconv.Atoi(grpcStatus)
		msg := resp.Header.Get("Grpc-Message")
		return status.Error(codes.Code(code), msg)
	}

	// Read the response body entirely.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("grpc-web: read response: %w", err)
	}

	reader := bytes.NewReader(respBody)
	var dataPayload []byte

	// Parse frames from response.
	for reader.Len() > 0 {
		flag, payload, err := parseGRPCWebFrame(reader)
		if err != nil {
			return fmt.Errorf("grpc-web: parse frame: %w", err)
		}
		if flag == 0x80 {
			// Trailer frame.
			trailers := parseTrailers(payload)
			st := statusFromTrailers(trailers)
			if st.Code() != codes.OK {
				return st.Err()
			}
		} else {
			// Data frame.
			dataPayload = payload
		}
	}

	if dataPayload == nil {
		return status.Error(codes.Internal, "grpc-web: no data frame in response")
	}

	if err := proto.Unmarshal(dataPayload, replyMsg); err != nil {
		return fmt.Errorf("grpc-web: unmarshal response: %w", err)
	}

	return nil
}

// NewStream implements grpc.ClientConnInterface for streaming RPCs.
// It supports server-streaming (the server sends multiple response frames).
func (c *grpcWebConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return &grpcWebClientStream{
		ctx:    ctx,
		conn:   c,
		method: method,
		desc:   desc,
	}, nil
}

// grpcWebClientStream implements grpc.ClientStream for gRPC-Web.
type grpcWebClientStream struct {
	ctx    context.Context
	conn   *grpcWebConn
	method string
	desc   *grpc.StreamDesc

	mu     sync.Mutex
	reader *bytes.Reader // response body reader
	sent   bool          // whether SendMsg has been called
	trailers map[string]string
	done     bool
}

func (s *grpcWebClientStream) Header() (metadata.MD, error) {
	return nil, nil
}

func (s *grpcWebClientStream) Trailer() metadata.MD {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.trailers == nil {
		return nil
	}
	md := metadata.MD{}
	for k, v := range s.trailers {
		md.Set(k, v)
	}
	return md
}

func (s *grpcWebClientStream) CloseSend() error {
	return nil
}

func (s *grpcWebClientStream) Context() context.Context {
	return s.ctx
}

func (s *grpcWebClientStream) SendMsg(m any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := m.(proto.Message)
	if !ok {
		return fmt.Errorf("grpc-web: msg must implement proto.Message")
	}

	reqBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("grpc-web: marshal stream request: %w", err)
	}
	body := grpcWebFrame(reqBytes)

	req, err := s.conn.buildRequest(s.ctx, s.method, body)
	if err != nil {
		return fmt.Errorf("grpc-web: build stream request: %w", err)
	}

	resp, err := s.conn.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("grpc-web: stream http request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return status.Errorf(codes.Unavailable, "grpc-web: stream HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Check for grpc-status in headers (early error).
	if grpcStatus := resp.Header.Get("Grpc-Status"); grpcStatus != "" && grpcStatus != "0" {
		code, _ := strconv.Atoi(grpcStatus)
		msg := resp.Header.Get("Grpc-Message")
		resp.Body.Close()
		return status.Error(codes.Code(code), msg)
	}

	// Read entire response for frame parsing.
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("grpc-web: read stream response: %w", err)
	}

	s.reader = bytes.NewReader(respBody)
	s.sent = true
	return nil
}

func (s *grpcWebClientStream) RecvMsg(m any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return io.EOF
	}

	if s.reader == nil || s.reader.Len() == 0 {
		return io.EOF
	}

	msg, ok := m.(proto.Message)
	if !ok {
		return fmt.Errorf("grpc-web: msg must implement proto.Message")
	}

	for s.reader.Len() > 0 {
		flag, payload, err := parseGRPCWebFrame(s.reader)
		if err != nil {
			return fmt.Errorf("grpc-web: parse stream frame: %w", err)
		}

		if flag == 0x80 {
			// Trailer frame — check status and signal end of stream.
			s.trailers = parseTrailers(payload)
			s.done = true
			st := statusFromTrailers(s.trailers)
			if st.Code() != codes.OK {
				return st.Err()
			}
			return io.EOF
		}

		// Data frame — unmarshal and return.
		if err := proto.Unmarshal(payload, msg); err != nil {
			return fmt.Errorf("grpc-web: unmarshal stream message: %w", err)
		}
		return nil
	}

	return io.EOF
}
