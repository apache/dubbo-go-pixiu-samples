package service

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/config"
	hessian "github.com/apache/dubbo-go-hessian2"
)

func init() {
	config.SetProviderService(&GreetingService{})
	hessian.RegisterPOJO(&GreetingRequest{})
	hessian.RegisterPOJO(&GreetingResponse{})
}

type GreetingService struct {
}

func (a *GreetingService) Greeting(ctx context.Context, in *GreetingRequest) (*GreetingResponse, error) {
	return &GreetingResponse{
		Greeting: "Hello " + in.Name + ", From Dubbo-go service",
	}, nil
}

func (a *GreetingService) Reference() string {
	return "GreetingService"
}

type GreetingResponse struct {
	Greeting string `json:"greeting"`
}

func (u GreetingResponse) JavaClassName() string {
	return "com.dubbo.demo.GreetingResponse"
}

type GreetingRequest struct {
	Name string
}

func (u GreetingRequest) JavaClassName() string {
	return "com.dubbo.demo.GreetingRequest"
}
