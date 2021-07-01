package main

import (
	"context"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
)

func hello(ctx context.Context, event events.APIGatewayRequest) (events.APIGatewayRequest, error) {

	return event, nil
}

func main() {
	cloudfunction.Start(hello)
}
