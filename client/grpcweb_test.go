package client

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
)

// buildGRPCWebResponse builds a valid gRPC-Web response body containing
// a data frame and a trailers frame with grpc-status: 0.
func buildGRPCWebResponse(msg proto.Message) ([]byte, error) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	// Data frame: flag=0x00
	buf.WriteByte(0x00)
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(payload)))
	buf.Write(lenBytes)
	buf.Write(payload)
	// Trailer frame: flag=0x80
	trailer := []byte("grpc-status: 0\r\n")
	buf.WriteByte(0x80)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(trailer)))
	buf.Write(lenBytes)
	buf.Write(trailer)
	return buf.Bytes(), nil
}

// buildGRPCWebErrorResponse builds a gRPC-Web response with only a trailers
// frame containing an error status.
func buildGRPCWebErrorResponse(code codes.Code, msg string) []byte {
	var buf bytes.Buffer
	trailer := []byte("grpc-status: " + codeToStr(code) + "\r\ngrpc-message: " + msg + "\r\n")
	buf.WriteByte(0x80)
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(trailer)))
	buf.Write(lenBytes)
	buf.Write(trailer)
	return buf.Bytes()
}

func codeToStr(c codes.Code) string {
	return []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"10", "11", "12", "13", "14", "15", "16",
	}[c]
}

func TestGRPCWebFrame(t *testing.T) {
	data := []byte("hello")
	frame := grpcWebFrame(data)
	if len(frame) != 5+len(data) {
		t.Fatalf("expected frame length %d, got %d", 5+len(data), len(frame))
	}
	if frame[0] != 0x00 {
		t.Errorf("expected flag 0x00, got 0x%02x", frame[0])
	}
	length := binary.BigEndian.Uint32(frame[1:5])
	if length != uint32(len(data)) {
		t.Errorf("expected length %d, got %d", len(data), length)
	}
	if !bytes.Equal(frame[5:], data) {
		t.Error("payload mismatch")
	}
}

func TestParseGRPCWebFrame(t *testing.T) {
	data := []byte("test payload")
	frame := grpcWebFrame(data)
	reader := bytes.NewReader(frame)

	flag, payload, err := parseGRPCWebFrame(reader)
	if err != nil {
		t.Fatal(err)
	}
	if flag != 0x00 {
		t.Errorf("expected flag 0x00, got 0x%02x", flag)
	}
	if !bytes.Equal(payload, data) {
		t.Error("payload mismatch")
	}
}

func TestParseGRPCWebFrameTrailer(t *testing.T) {
	trailer := []byte("grpc-status: 0\r\ngrpc-message: ok\r\n")
	var buf bytes.Buffer
	buf.WriteByte(0x80)
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(trailer)))
	buf.Write(lenBytes)
	buf.Write(trailer)

	flag, payload, err := parseGRPCWebFrame(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if flag != 0x80 {
		t.Errorf("expected flag 0x80, got 0x%02x", flag)
	}
	if !bytes.Equal(payload, trailer) {
		t.Error("trailer payload mismatch")
	}
}

func TestParseGRPCWebFrameTruncated(t *testing.T) {
	// Only 3 bytes — not enough for the 5-byte header.
	_, _, err := parseGRPCWebFrame(bytes.NewReader([]byte{0x00, 0x00, 0x00}))
	if err == nil {
		t.Error("expected error on truncated header")
	}
}

func TestParseTrailers(t *testing.T) {
	data := []byte("grpc-status: 2\r\ngrpc-message: not found\r\n")
	trailers := parseTrailers(data)
	if trailers["grpc-status"] != "2" {
		t.Errorf("expected grpc-status 2, got %q", trailers["grpc-status"])
	}
	if trailers["grpc-message"] != "not found" {
		t.Errorf("expected grpc-message 'not found', got %q", trailers["grpc-message"])
	}
}

func TestStatusFromTrailersOK(t *testing.T) {
	trailers := map[string]string{"grpc-status": "0"}
	st := statusFromTrailers(trailers)
	if st.Code() != codes.OK {
		t.Errorf("expected OK, got %v", st.Code())
	}
}

func TestStatusFromTrailersError(t *testing.T) {
	trailers := map[string]string{
		"grpc-status":  "5",
		"grpc-message": "not found",
	}
	st := statusFromTrailers(trailers)
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
	if st.Message() != "not found" {
		t.Errorf("expected message 'not found', got %q", st.Message())
	}
}

func TestStatusFromTrailersNoStatus(t *testing.T) {
	trailers := map[string]string{}
	st := statusFromTrailers(trailers)
	if st.Code() != codes.OK {
		t.Errorf("expected OK when no grpc-status, got %v", st.Code())
	}
}

func TestGRPCWebConnInvoke(t *testing.T) {
	// Set up a test HTTP server that returns a valid gRPC-Web response.
	expectedResp := &pb.GetServiceInfoResponse{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request properties.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/grpc-web+proto" {
			t.Errorf("expected content-type application/grpc-web+proto, got %s", ct)
		}
		if xgw := r.Header.Get("X-Grpc-Web"); xgw != "1" {
			t.Errorf("expected X-Grpc-Web: 1, got %s", xgw)
		}
		if r.URL.Path != "/sui.rpc.v2.LedgerService/GetServiceInfo" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Read and validate request frame.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(body) < 5 {
			t.Fatal("request body too short")
		}
		if body[0] != 0x00 {
			t.Errorf("expected data frame flag 0x00, got 0x%02x", body[0])
		}

		// Write response.
		respBody, err := buildGRPCWebResponse(expectedResp)
		if err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	}))
	defer server.Close()

	conn := newGRPCWebConn(server.URL, nil, 0)

	req := &pb.GetServiceInfoRequest{}
	resp := &pb.GetServiceInfoResponse{}
	err := conn.Invoke(t.Context(), "/sui.rpc.v2.LedgerService/GetServiceInfo", req, resp, nil)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
}

func TestGRPCWebConnInvokeWithHeaders(t *testing.T) {
	var receivedAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("X-Api-Key")
		respBody, _ := buildGRPCWebResponse(&pb.GetServiceInfoResponse{})
		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.Write(respBody)
	}))
	defer server.Close()

	headers := map[string]string{"x-api-key": "test-key-123"}
	conn := newGRPCWebConn(server.URL, headers, 0)

	err := conn.Invoke(t.Context(), "/sui.rpc.v2.LedgerService/GetServiceInfo",
		&pb.GetServiceInfoRequest{}, &pb.GetServiceInfoResponse{})
	if err != nil {
		t.Fatal(err)
	}
	if receivedAPIKey != "test-key-123" {
		t.Errorf("expected api key 'test-key-123', got %q", receivedAPIKey)
	}
}

func TestGRPCWebConnInvokeHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	conn := newGRPCWebConn(server.URL, nil, 0)
	err := conn.Invoke(t.Context(), "/test/Method",
		&pb.GetServiceInfoRequest{}, &pb.GetServiceInfoResponse{})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unavailable {
		t.Errorf("expected Unavailable, got %v", st.Code())
	}
}

func TestGRPCWebConnInvokeGRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respBody := buildGRPCWebErrorResponse(codes.NotFound, "object not found")
		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.Write(respBody)
	}))
	defer server.Close()

	conn := newGRPCWebConn(server.URL, nil, 0)
	err := conn.Invoke(t.Context(), "/test/Method",
		&pb.GetServiceInfoRequest{}, &pb.GetServiceInfoResponse{})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
	if st.Message() != "object not found" {
		t.Errorf("expected 'object not found', got %q", st.Message())
	}
}

func TestGRPCWebConnInvokeHeaderError(t *testing.T) {
	// Server returns grpc-status in HTTP headers (early error).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.Header().Set("Grpc-Status", "13")
		w.Header().Set("Grpc-Message", "internal error")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conn := newGRPCWebConn(server.URL, nil, 0)
	err := conn.Invoke(t.Context(), "/test/Method",
		&pb.GetServiceInfoRequest{}, &pb.GetServiceInfoResponse{})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

func TestGRPCWebConnStream(t *testing.T) {
	// Build a response with two data frames and a trailer.
	resp1 := &pb.GetServiceInfoResponse{}
	resp2 := &pb.GetServiceInfoResponse{}

	var respBuf bytes.Buffer
	for _, msg := range []proto.Message{resp1, resp2} {
		payload, _ := proto.Marshal(msg)
		respBuf.WriteByte(0x00)
		lenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBytes, uint32(len(payload)))
		respBuf.Write(lenBytes)
		respBuf.Write(payload)
	}
	// Trailer frame.
	trailer := []byte("grpc-status: 0\r\n")
	respBuf.WriteByte(0x80)
	tlen := make([]byte, 4)
	binary.BigEndian.PutUint32(tlen, uint32(len(trailer)))
	respBuf.Write(tlen)
	respBuf.Write(trailer)
	respBytes := respBuf.Bytes()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.Write(respBytes)
	}))
	defer server.Close()

	conn := newGRPCWebConn(server.URL, nil, 0)
	stream, err := conn.NewStream(t.Context(), nil, "/test/Stream")
	if err != nil {
		t.Fatal(err)
	}

	// SendMsg triggers the HTTP request.
	if err := stream.SendMsg(&pb.GetServiceInfoRequest{}); err != nil {
		t.Fatal(err)
	}

	// Read first message.
	msg1 := &pb.GetServiceInfoResponse{}
	if err := stream.RecvMsg(msg1); err != nil {
		t.Fatalf("RecvMsg 1: %v", err)
	}

	// Read second message.
	msg2 := &pb.GetServiceInfoResponse{}
	if err := stream.RecvMsg(msg2); err != nil {
		t.Fatalf("RecvMsg 2: %v", err)
	}

	// Third RecvMsg should return EOF (trailer frame).
	msg3 := &pb.GetServiceInfoResponse{}
	err = stream.RecvMsg(msg3)
	if err != io.EOF {
		t.Fatalf("expected EOF, got: %v", err)
	}
}

func TestNewGRPCWebClient(t *testing.T) {
	c, err := New("https://localhost:443", WithGRPCWeb())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if !c.IsGRPCWeb() {
		t.Error("expected IsGRPCWeb() to return true")
	}
	if c.conn != nil {
		t.Error("expected native gRPC conn to be nil")
	}
	if c.webConn == nil {
		t.Error("expected webConn to be non-nil")
	}
	if c.ledger == nil || c.state == nil || c.txExec == nil || c.pkg == nil || c.name == nil || c.sub == nil || c.sigVer == nil {
		t.Error("expected all service clients to be initialized")
	}
}

func TestNewGRPCWebClientAutoScheme(t *testing.T) {
	// Without scheme, TLS enabled (default) should prepend https://.
	c, err := New("localhost:443", WithGRPCWeb())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if c.webConn.baseURL != "https://localhost:443" {
		t.Errorf("expected https://localhost:443, got %s", c.webConn.baseURL)
	}
}

func TestNewGRPCWebClientAutoSchemeInsecure(t *testing.T) {
	c, err := New("localhost:8080", WithGRPCWeb(), WithTLS(false))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if c.webConn.baseURL != "http://localhost:8080" {
		t.Errorf("expected http://localhost:8080, got %s", c.webConn.baseURL)
	}
}

func TestWithGRPCWebOption(t *testing.T) {
	opts := defaultOptions()
	if opts.grpcWeb {
		t.Error("grpcWeb should be false by default")
	}
	WithGRPCWeb()(opts)
	if !opts.grpcWeb {
		t.Error("grpcWeb should be true after WithGRPCWeb()")
	}
}

func TestGRPCWebCloseIsNoop(t *testing.T) {
	c, err := New("https://localhost:443", WithGRPCWeb())
	if err != nil {
		t.Fatal(err)
	}
	// Close should not panic or error for gRPC-Web clients.
	if err := c.Close(); err != nil {
		t.Errorf("expected nil error from Close(), got: %v", err)
	}
}

func TestGRPCWebConnInvokeEndToEnd(t *testing.T) {
	// Full end-to-end: create SuiClient with gRPC-Web, call GetServiceInfo.
	expectedResp := &pb.GetServiceInfoResponse{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respBody, _ := buildGRPCWebResponse(expectedResp)
		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.Write(respBody)
	}))
	defer server.Close()

	c, err := New(server.URL, WithGRPCWeb(), WithTLS(false))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	resp, err := c.GetServiceInfo(t.Context())
	if err != nil {
		t.Fatalf("GetServiceInfo failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}
