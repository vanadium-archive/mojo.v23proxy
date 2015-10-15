package client

import (
	"mojo/public/go/application"
	"mojo/public/go/bindings"

	"mojom/v23proxy"
)

func ConnectToRemoteService(ctx application.Context, r application.ServiceRequest, v23Name string) {
	v23r, v23p := v23proxy.CreateMessagePipeForV23()
	ctx.ConnectToApplication("mojo:v23proxy").ConnectToService(&v23r)
	prox := v23proxy.NewV23Proxy(v23p, bindings.GetAsyncWaiter())
	prox.SetupProxy(v23Name, r.Type(), r.Desc(), r.Name(), r.PassMessagePipe())
}
