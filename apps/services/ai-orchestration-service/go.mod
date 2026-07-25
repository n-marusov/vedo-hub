module vedo-core/src/services/ai-orchestration-service

go 1.22

require (
	github.com/prometheus/client_golang v1.19.1
	golang.org/x/text v0.14.0
	google.golang.org/grpc v1.64.0
	vedo-core/src/services/shared/llm v0.0.0-00010101000000-000000000000
	vedo-core/src/services/shared/proto v0.0.0-00010101000000-000000000000
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace vedo-core/src/services/shared/llm => ../shared/llm

replace vedo-core/src/services/shared/proto => ../shared/proto
