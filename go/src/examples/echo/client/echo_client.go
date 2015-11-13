// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	v23 "v.io/x/mojo/client"

	"mojom/examples/echo"
)

//#include "mojo/public/c/system/types.h"
import "C"

type RemoteEchoClientDelegate struct{}

// When running echo_client, ctx.Args() should contain:
// 0: mojo app name
// 1: remote endpoint
// 2+: string to echo
func (delegate *RemoteEchoClientDelegate) Initialize(ctx application.Context) {
	log.Printf("RemoteEchoClientDelegate.Initialize...")

	// Parse arguments. Note: May panic if not enough args are given.
	remoteEndpoint := ctx.Args()[1]
	echoString := "Hello, Go world!"
	if len(ctx.Args()) > 2 {
		echoString = strings.Join(ctx.Args()[2:], " ")
	}

	r, p := echo.CreateMessagePipeForRemoteEcho()

	v23.ConnectToRemoteService(ctx, &r, remoteEndpoint)
	echoProxy := echo.NewRemoteEchoProxy(p, bindings.GetAsyncWaiter())

	log.Printf("RemoteEchoClientDelegate.Initialize calling EchoString...")
	response, err := echoProxy.EchoString(echoString)
	if err == nil {
		fmt.Printf("client: %s\n", response)
	} else {
		log.Println(err)
	}

	log.Printf("RemoteEchoClientDelegate.Initialize calling EchoX...")
	response2, err := echoProxy.EchoX([]bool{true, false, false, true}, echo.AInArg{"A String"})
	if err == nil {
		fmt.Printf("client: %v\n", response2)
	} else {
		log.Println("Error: ", err)
	}

	fmt.Printf("(done)\n")
	echoProxy.Close_Proxy()
	ctx.Close()
}

func (delegate *RemoteEchoClientDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("RemoteEchoClientDelegate.AcceptConnection...")
	connection.Close()
}

func (delegate *RemoteEchoClientDelegate) Quit() {
	log.Printf("RemoteEchoClientDelegate.Quit...")
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&RemoteEchoClientDelegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
