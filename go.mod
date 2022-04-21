module github.com/dubbo-go-pixiu/samples

go 1.17

require (
	dubbo.apache.org/dubbo-go/v3 v3.0.1-0.20220107110037-4496cef73dba
	github.com/apache/dubbo-go-hessian2 v1.11.0
	github.com/apache/dubbo-go-pixiu v0.0.0-20220321132145-e68ff8dd6c80
	github.com/dubbogo/gost v1.11.22
	github.com/dubbogo/grpc-go v1.42.7
	github.com/dubbogo/triple v1.1.7
	github.com/gin-gonic/gin v1.7.4
	github.com/golang/protobuf v1.5.2
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1

)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.5
