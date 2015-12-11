// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package client

import (
	"mojo/public/go/application"
	"mojo/public/go/bindings"

	"mojom/v23clientproxy"
)

func ConnectToRemoteService(ctx application.Context, r application.ServiceRequest, v23Name string) {
	v23r, v23p := v23clientproxy.CreateMessagePipeForV23ClientProxy()
	ctx.ConnectToApplication("https://mojo.v.io/v23clientproxy.mojo").ConnectToService(&v23r)
	prox := v23clientproxy.NewV23ClientProxyProxy(v23p, bindings.GetAsyncWaiter())
	sd := r.ServiceDescription()
	mojomInterfaceType, err := sd.GetTopLevelInterface()
	if err != nil {
		// The service description must have the MojomInterface type.
		panic(err)
	}
	desc, err := sd.GetAllTypeDefinitions()
	if err != nil {
		// The service description must have the map of UserDefinedTypes.
		panic(err)
	}

	prox.SetupClientProxy(v23Name, mojomInterfaceType, *desc, r.Name(), r.PassMessagePipe())
}
