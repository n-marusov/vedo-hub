module vedo-core/src/services/auth-service

go 1.22

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/prometheus/client_golang v1.19.1
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.1
	vedo-core/src/services/shared/proto v0.0.0
)

replace vedo-core/src/services/shared/proto => ../shared/proto
