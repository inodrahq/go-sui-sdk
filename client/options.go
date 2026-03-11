package client

import "time"

// Option configures a SuiClient.
type Option func(*options)

type options struct {
	tls     bool
	apiKey  string
	headers map[string]string
	timeout time.Duration
	grpcWeb bool
}

func defaultOptions() *options {
	return &options{
		tls: true,
	}
}

// WithTLS enables TLS for the gRPC connection (default: true).
func WithTLS(enabled bool) Option {
	return func(o *options) {
		o.tls = enabled
	}
}

// WithAPIKey sets an API key sent as the x-api-key metadata header.
func WithAPIKey(key string) Option {
	return func(o *options) {
		o.apiKey = key
	}
}

// WithHeaders sets custom metadata headers sent with every request.
func WithHeaders(headers map[string]string) Option {
	return func(o *options) {
		o.headers = headers
	}
}

// WithTimeout sets a default timeout applied to every RPC call.
// If the caller's context already has an earlier deadline, that takes precedence.
// A zero value means no client-level timeout (callers control it via context).
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

// WithGRPCWeb enables the gRPC-Web protocol (HTTP/1.1-based) instead of
// native HTTP/2 gRPC. This is useful when connecting through proxies or
// environments that do not support HTTP/2, such as browsers or Cloudflare.
// When enabled, the target should be an HTTP(S) URL (e.g.
// "https://fullnode.testnet.sui.io:443").
func WithGRPCWeb() Option {
	return func(o *options) {
		o.grpcWeb = true
	}
}
