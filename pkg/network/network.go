package network

import (
	"context"
	"net"
	"time"

	"github.com/gossip-broker/gossip/pkg/types"
)

// Server represents a network server
type Server interface {
	// Start the server
	Start(ctx context.Context) error

	// Stop the server
	Stop(ctx context.Context) error

	// Get server address
	Address() string

	// Get server statistics
	Stats() ServerStats
}

// ServerStats contains server statistics
type ServerStats struct {
	ActiveConnections int64         `json:"active_connections"`
	TotalConnections  int64         `json:"total_connections"`
	BytesReceived     int64         `json:"bytes_received"`
	BytesSent         int64         `json:"bytes_sent"`
	RequestsHandled   int64         `json:"requests_handled"`
	ErrorsCount       int64         `json:"errors_count"`
	StartTime         time.Time     `json:"start_time"`
	Uptime            time.Duration `json:"uptime"`
}

// Connection represents a client connection
type Connection interface {
	// Get connection ID
	ID() string

	// Get remote address
	RemoteAddr() net.Addr

	// Get local address
	LocalAddr() net.Addr

	// Send data to the connection
	Send(ctx context.Context, data []byte) error

	// Receive data from the connection
	Receive(ctx context.Context) ([]byte, error)

	// Close the connection
	Close() error

	// Check if connection is alive
	IsAlive() bool

	// Get connection metadata
	Metadata() map[string]interface{}

	// Set connection metadata
	SetMetadata(key string, value interface{})
}

// ConnectionHandler handles incoming connections
type ConnectionHandler interface {
	// Handle a new connection
	HandleConnection(ctx context.Context, conn Connection) error
}

// RequestHandler handles individual requests
type RequestHandler interface {
	// Handle a request and return a response
	HandleRequest(ctx context.Context, req *Request) (*Response, error)

	// Get supported request types
	SupportedTypes() []RequestType
}

// Request represents a network request
type Request struct {
	ID        string                 `json:"id"`
	Type      RequestType            `json:"type"`
	Headers   map[string]string      `json:"headers"`
	Body      []byte                 `json:"body"`
	Timestamp time.Time              `json:"timestamp"`
	ClientID  string                 `json:"client_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Response represents a network response
type Response struct {
	ID        string            `json:"id"`
	RequestID string            `json:"request_id"`
	Status    ResponseStatus    `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body"`
	Error     *ResponseError    `json:"error,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// RequestType represents the type of request
type RequestType string

const (
	RequestTypeProduce       RequestType = "PRODUCE"
	RequestTypeConsume       RequestType = "CONSUME"
	RequestTypeCreateTopic   RequestType = "CREATE_TOPIC"
	RequestTypeDeleteTopic   RequestType = "DELETE_TOPIC"
	RequestTypeListTopics    RequestType = "LIST_TOPICS"
	RequestTypeDescribeTopic RequestType = "DESCRIBE_TOPIC"
	RequestTypeJoinGroup     RequestType = "JOIN_GROUP"
	RequestTypeLeaveGroup    RequestType = "LEAVE_GROUP"
	RequestTypeHeartbeat     RequestType = "HEARTBEAT"
	RequestTypeCommitOffset  RequestType = "COMMIT_OFFSET"
	RequestTypeFetchOffset   RequestType = "FETCH_OFFSET"
	RequestTypeMetadata      RequestType = "METADATA"
	RequestTypeHealth        RequestType = "HEALTH"
	RequestTypeMetrics       RequestType = "METRICS"
)

// ResponseStatus represents the status of a response
type ResponseStatus string

const (
	ResponseStatusOK    ResponseStatus = "OK"
	ResponseStatusError ResponseStatus = "ERROR"
)

// ResponseError represents an error in a response
type ResponseError struct {
	Code    types.ErrorCode        `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Protocol defines the network protocol
type Protocol interface {
	// Encode a request/response to bytes
	Encode(ctx context.Context, msg interface{}) ([]byte, error)

	// Decode bytes to a request/response
	Decode(ctx context.Context, data []byte) (interface{}, error)

	// Get protocol name
	Name() string

	// Get protocol version
	Version() string
}

// BinaryProtocol handles binary protocol encoding/decoding
type BinaryProtocol interface {
	Protocol

	// Encode request to binary format
	EncodeRequest(req *Request) ([]byte, error)

	// Decode binary data to request
	DecodeRequest(data []byte) (*Request, error)

	// Encode response to binary format
	EncodeResponse(resp *Response) ([]byte, error)

	// Decode binary data to response
	DecodeResponse(data []byte) (*Response, error)
}

// HTTPServer provides HTTP/REST API
type HTTPServer interface {
	Server

	// Register a route handler
	RegisterHandler(method, path string, handler HTTPHandler)

	// Register middleware
	RegisterMiddleware(middleware HTTPMiddleware)

	// Serve static files
	ServeStatic(path, dir string)
}

// HTTPHandler handles HTTP requests
type HTTPHandler interface {
	ServeHTTP(ctx context.Context, req *HTTPRequest, resp HTTPResponseWriter)
}

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Query      map[string]string `json:"query"`
	PathParams map[string]string `json:"path_params"`
	RemoteAddr string            `json:"remote_addr"`
	UserAgent  string            `json:"user_agent"`
}

// HTTPResponseWriter writes HTTP responses
type HTTPResponseWriter interface {
	// Set response status code
	SetStatus(code int)

	// Set response header
	SetHeader(key, value string)

	// Write response body
	Write(data []byte) (int, error)

	// Write JSON response
	WriteJSON(data interface{}) error

	// Write error response
	WriteError(code int, message string) error
}

// HTTPMiddleware provides HTTP middleware
type HTTPMiddleware interface {
	Handle(ctx context.Context, req *HTTPRequest, resp HTTPResponseWriter, next HTTPHandler)
}

// gRPCServer provides gRPC API
type GRPCServer interface {
	Server

	// Register a service
	RegisterService(desc interface{}, impl interface{})

	// Register interceptors
	RegisterUnaryInterceptor(interceptor UnaryInterceptor)
	RegisterStreamInterceptor(interceptor StreamInterceptor)
}

// UnaryInterceptor intercepts unary RPC calls
type UnaryInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error)

// StreamInterceptor intercepts streaming RPC calls
type StreamInterceptor func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error

// UnaryServerInfo contains information about a unary RPC call
type UnaryServerInfo struct {
	Server     interface{}
	FullMethod string
}

// StreamServerInfo contains information about a streaming RPC call
type StreamServerInfo struct {
	FullMethod     string
	IsClientStream bool
	IsServerStream bool
}

// UnaryHandler handles unary RPC calls
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// StreamHandler handles streaming RPC calls
type StreamHandler func(srv interface{}, stream ServerStream) error

// ServerStream represents a server-side stream
type ServerStream interface {
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

// TCPServer provides raw TCP server functionality
type TCPServer interface {
	Server

	// Set connection handler
	SetConnectionHandler(handler ConnectionHandler)

	// Set connection options
	SetConnectionOptions(opts *ConnectionOptions)
}

// ConnectionOptions contains connection configuration
type ConnectionOptions struct {
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	MaxMessageSize  int64         `json:"max_message_size"`
	KeepAlive       bool          `json:"keep_alive"`
	KeepAlivePeriod time.Duration `json:"keep_alive_period"`
	NoDelay         bool          `json:"no_delay"`
	Buffer          BufferOptions `json:"buffer"`
}

// BufferOptions contains buffer configuration
type BufferOptions struct {
	ReadBufferSize  int `json:"read_buffer_size"`
	WriteBufferSize int `json:"write_buffer_size"`
}

// LoadBalancer distributes requests across multiple backends
type LoadBalancer interface {
	// Add a backend
	AddBackend(backend Backend) error

	// Remove a backend
	RemoveBackend(backendID string) error

	// Get a backend for a request
	GetBackend(ctx context.Context, req *Request) (Backend, error)

	// Get all backends
	GetBackends() []Backend

	// Get load balancing algorithm
	Algorithm() LoadBalancingAlgorithm
}

// Backend represents a backend server
type Backend struct {
	ID       string                 `json:"id"`
	Address  string                 `json:"address"`
	Weight   int                    `json:"weight"`
	IsActive bool                   `json:"is_active"`
	Metadata map[string]interface{} `json:"metadata"`
}

// LoadBalancingAlgorithm represents load balancing algorithms
type LoadBalancingAlgorithm string

const (
	LoadBalancingAlgorithmRoundRobin LoadBalancingAlgorithm = "round_robin"
	LoadBalancingAlgorithmWeighted   LoadBalancingAlgorithm = "weighted"
	LoadBalancingAlgorithmLeastConn  LoadBalancingAlgorithm = "least_conn"
	LoadBalancingAlgorithmIPHash     LoadBalancingAlgorithm = "ip_hash"
	LoadBalancingAlgorithmConsistent LoadBalancingAlgorithm = "consistent_hash"
)

// HealthChecker checks the health of backends
type HealthChecker interface {
	// Check the health of a backend
	Check(ctx context.Context, backend Backend) error

	// Start periodic health checks
	StartChecking(ctx context.Context, interval time.Duration)

	// Stop health checks
	StopChecking()

	// Register health change callback
	OnHealthChange(callback func(backend Backend, isHealthy bool))
}

// RateLimiter limits request rates
type RateLimiter interface {
	// Check if a request is allowed
	Allow(ctx context.Context, key string) (bool, error)

	// Allow with custom limit
	AllowN(ctx context.Context, key string, n int) (bool, error)

	// Get current rate for a key
	Rate(ctx context.Context, key string) (int, error)

	// Reset rate for a key
	Reset(ctx context.Context, key string) error
}

// CircuitBreaker prevents cascading failures
type CircuitBreaker interface {
	// Execute a function with circuit breaker protection
	Execute(ctx context.Context, fn func() error) error

	// Get circuit breaker state
	State() CircuitBreakerState

	// Get circuit breaker statistics
	Stats() CircuitBreakerStats
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerStateClosed   CircuitBreakerState = "CLOSED"
	CircuitBreakerStateOpen     CircuitBreakerState = "OPEN"
	CircuitBreakerStateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	State            CircuitBreakerState `json:"state"`
	TotalRequests    int64               `json:"total_requests"`
	SuccessfulReqs   int64               `json:"successful_requests"`
	FailedRequests   int64               `json:"failed_requests"`
	ConsecutiveFails int64               `json:"consecutive_failures"`
	LastFailureTime  time.Time           `json:"last_failure_time"`
	LastSuccessTime  time.Time           `json:"last_success_time"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	// Server configuration
	BindAddress       string        `json:"bind_address"`
	Port              int           `json:"port"`
	MaxConnections    int           `json:"max_connections"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`

	// Protocol configuration
	Protocol    string `json:"protocol"` // "tcp", "http", "grpc"
	EnableTLS   bool   `json:"enable_tls"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`
	TLSCAFile   string `json:"tls_ca_file"`

	// Performance tuning
	ReadBufferSize  int           `json:"read_buffer_size"`
	WriteBufferSize int           `json:"write_buffer_size"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`

	// Load balancing
	LoadBalancer        LoadBalancingAlgorithm `json:"load_balancer"`
	HealthCheckInterval time.Duration          `json:"health_check_interval"`

	// Rate limiting
	RateLimit       int           `json:"rate_limit"`
	RateLimitWindow time.Duration `json:"rate_limit_window"`

	// Circuit breaker
	CircuitBreakerEnabled   bool          `json:"circuit_breaker_enabled"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
}

// NetworkFactory creates network components
type NetworkFactory interface {
	// Create a new TCP server
	NewTCPServer(config *NetworkConfig) (TCPServer, error)

	// Create a new HTTP server
	NewHTTPServer(config *NetworkConfig) (HTTPServer, error)

	// Create a new gRPC server
	NewGRPCServer(config *NetworkConfig) (GRPCServer, error)

	// Create a new load balancer
	NewLoadBalancer(algorithm LoadBalancingAlgorithm) (LoadBalancer, error)

	// Create a new health checker
	NewHealthChecker() (HealthChecker, error)

	// Create a new rate limiter
	NewRateLimiter(limit int, window time.Duration) (RateLimiter, error)

	// Create a new circuit breaker
	NewCircuitBreaker(threshold int, timeout time.Duration) (CircuitBreaker, error)
}
