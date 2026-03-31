// Package observability provides distributed tracing, logging, and alerting
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// TracingProvider provides distributed tracing capabilities
type TracingProvider struct {
	serviceName string
	sampler     Sampler
	exporter    SpanExporter

	// Trace storage
	traces   sync.Map // map[string]*Trace
	spans    sync.Map // map[string]*Span

	// Metrics
	totalSpans    int64
	totalTraces   int64
	activeTraces  int64

	mu sync.RWMutex
}

// Sampler determines if a trace should be sampled
type Sampler interface {
	ShouldSample(traceID string) bool
}

// SpanExporter exports spans to external systems
type SpanExporter interface {
	Export(span *Span) error
}

// Span represents a single span in a trace
type Span struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	Operation    string            `json:"operation"`
	Service      string            `json:"service"`
	Start        time.Time         `json:"start"`
	End          time.Time         `json:"end,omitempty"`
	Duration     time.Duration     `json:"duration"`
	Status       SpanStatus        `json:"status"`
	Tags         map[string]string `json:"tags"`
	Logs         []SpanLog         `json:"logs"`
	TenantID     string            `json:"tenant_id,omitempty"`
}

// SpanStatus defines span status
type SpanStatus string

const (
	SpanOK       SpanStatus = "ok"
	SpanError    SpanStatus = "error"
	SpanCanceled SpanStatus = "canceled"
)

// SpanLog represents a log entry in a span
type SpanLog struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
}

// Trace represents a complete trace
type Trace struct {
	TraceID    string  `json:"trace_id"`
	RootSpan   *Span   `json:"root_span"`
	Spans      []*Span `json:"spans"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Duration   time.Duration `json:"duration"`
	 TenantID  string  `json:"tenant_id,omitempty"`
}

// NewTracingProvider creates a new tracing provider
func NewTracingProvider(serviceName string) *TracingProvider {
	return &TracingProvider{
		serviceName: serviceName,
		sampler:     &ProbabilitySampler{Rate: 0.1}, // 10% sampling
	}
}

// StartSpan starts a new span
func (p *TracingProvider) StartSpan(ctx context.Context, operation string) (context.Context, *Span) {
	// Get trace ID from context or create new
	traceID := getTraceIDFromContext(ctx)
	if traceID == "" {
		traceID = generateTraceID()
	}

	spanID := generateSpanID()

	// Get parent span ID
	parentSpanID := getSpanIDFromContext(ctx)

	span := &Span{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		Operation:    operation,
		Service:      p.serviceName,
		Start:        time.Now(),
		Status:       SpanOK,
		Tags:         make(map[string]string),
		Logs:         make([]SpanLog, 0),
	}

	// Get tenant ID from context
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		span.TenantID = tenantID
	}

	p.spans.Store(spanID, span)
	p.totalSpans++

	// Store trace
	if trace, ok := p.traces.Load(traceID); ok {
		t := trace.(*Trace)
		t.Spans = append(t.Spans, span)
	} else {
		trace := &Trace{
			TraceID:   traceID,
			RootSpan:  span,
			Spans:     []*Span{span},
			StartTime: time.Now(),
		}
		p.traces.Store(traceID, trace)
		p.totalTraces++
	}

	// Add span to context
	ctx = context.WithValue(ctx, "trace_id", traceID)
	ctx = context.WithValue(ctx, "span_id", spanID)

	return ctx, span
}

// EndSpan ends a span
func (p *TracingProvider) EndSpan(span *Span) {
	span.End = time.Now()
	span.Duration = span.End.Sub(span.Start)

	// Export span
	if p.exporter != nil {
		if err := p.exporter.Export(span); err != nil {
			log.Printf("Failed to export span: %v", err)
		}
	}

	// Update trace
	if trace, ok := p.traces.Load(span.TraceID); ok {
		t := trace.(*Trace)
		if t.RootSpan.SpanID == span.SpanID {
			t.EndTime = span.End
			t.Duration = span.Duration
		}
	}
}

// SetSpanError marks a span as error
func (p *TracingProvider) SetSpanError(span *Span, err error) {
	span.Status = SpanError
	span.Tags["error"] = "true"
	span.Tags["error.message"] = err.Error()
}

// AddSpanLog adds a log to a span
func (p *TracingProvider) AddSpanLog(span *Span, level, message string) {
	span.Logs = append(span.Logs, SpanLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	})
}

// GetTrace retrieves a trace by ID
func (p *TracingProvider) GetTrace(traceID string) (*Trace, error) {
	if v, ok := p.traces.Load(traceID); ok {
		return v.(*Trace), nil
	}
	return nil, fmt.Errorf("trace not found: %s", traceID)
}

// GetSpan retrieves a span by ID
func (p *TracingProvider) GetSpan(spanID string) (*Span, error) {
	if v, ok := p.spans.Load(spanID); ok {
		return v.(*Span), nil
	}
	return nil, fmt.Errorf("span not found: %s", spanID)
}

// GetStats returns tracing statistics
func (p *TracingProvider) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"service_name":    p.serviceName,
		"total_traces":    p.totalTraces,
		"total_spans":     p.totalSpans,
		"active_traces":   p.activeTraces,
	}
}

// ProbabilitySampler samples based on probability
type ProbabilitySampler struct {
	Rate float64
}

// ShouldSample implements Sampler
func (s *ProbabilitySampler) ShouldSample(traceID string) bool {
	// Simple hash-based sampling
	hash := 0
	for _, c := range traceID {
		hash += int(c)
	}
	return float64(hash%100)/100 < s.Rate
}

// Helper functions

func generateTraceID() string {
	return fmt.Sprintf("trace-%d-%s", time.Now().UnixNano(), randomString(8))
}

func generateSpanID() string {
	return fmt.Sprintf("span-%d", time.Now().UnixNano())
}

func getTraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value("trace_id").(string); ok {
		return v
	}
	return ""
}

func getSpanIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value("span_id").(string); ok {
		return v
	}
	return ""
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// JaegerExporter exports spans to Jaeger
type JaegerExporter struct {
	Endpoint string
}

// Export implements SpanExporter
func (e *JaegerExporter) Export(span *Span) error {
	// Placeholder for Jaeger export
	// In production, would use Jaeger client library
	return nil
}

// ZipkinExporter exports spans to Zipkin
type ZipkinExporter struct {
	Endpoint string
}

// Export implements SpanExporter
func (e *ZipkinExporter) Export(span *Span) error {
	// Placeholder for Zipkin export
	return nil
}

// TracingMiddleware creates HTTP middleware for tracing
type TracingMiddleware struct {
	provider *TracingProvider
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware(provider *TracingProvider) *TracingMiddleware {
	return &TracingMiddleware{
		provider: provider,
	}
}

// Wrap wraps a handler with tracing
func (m *TracingMiddleware) Wrap(handler string) func(ctx context.Context) (context.Context, func()) {
	return func(ctx context.Context) (context.Context, func()) {
		ctx, span := m.provider.StartSpan(ctx, handler)
		return ctx, func() {
			m.provider.EndSpan(span)
		}
	}
}

// SpanJSON exports span as JSON
func SpanJSON(span *Span) ([]byte, error) {
	return json.Marshal(span)
}

// TraceJSON exports trace as JSON
func TraceJSON(trace *Trace) ([]byte, error) {
	return json.Marshal(trace)
}