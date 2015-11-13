// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	"mojom/examples/echo"
)

//#include "mojo/public/c/system/types.h"
import "C"

type RemoteEchoImpl struct{}

// Note: This is pretty much identical to echo_server.go, except for the name changes.
func (re *RemoteEchoImpl) EchoString(inValue string) (outValue string, err error) {
	log.Printf("server EchoString: %s\n", inValue)
	return inValue, nil
}

func (re *RemoteEchoImpl) EchoX(inArg1 []bool, inArg2 echo.AInArg) (out echo.OutArgTypes, err error) {
	log.Printf("server EchoX: arg1: %v arg2: %v\n", inArg1, inArg2)
	return &echo.OutArgTypesRes{echo.Result_B}, nil
}

type RemoteEchoServerDelegate struct {
	remoteEchoFactory RemoteEchoFactory
}

type RemoteEchoFactory struct {
	stubs []*bindings.Stub
}

func (delegate *RemoteEchoServerDelegate) Initialize(context application.Context) {
	log.Printf("RemoteEchoServerDelegate.Initialize...")
}

func (remoteEchoFactory *RemoteEchoFactory) Create(request echo.RemoteEcho_Request) {
	log.Printf("RemoteEchoServer's RemoteEchoFactory.Create...")
	stub := echo.NewRemoteEchoStub(request, &RemoteEchoImpl{}, bindings.GetAsyncWaiter())
	remoteEchoFactory.stubs = append(remoteEchoFactory.stubs, stub)
	go func() {
		for {
			if err := stub.ServeRequest(); err != nil {
				connectionError, ok := err.(*bindings.ConnectionError)
				if !ok || !connectionError.Closed() {
					log.Println(err)
				}
				break
			}
		}
	}()
}

func (delegate *RemoteEchoServerDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("RemoteEchoServerDelegate.AcceptConnection...")
	connection.ProvideServicesWithDescriber(
		&echo.RemoteEcho_ServiceFactory{&delegate.remoteEchoFactory},
	)
}

func (delegate *RemoteEchoServerDelegate) Quit() {
	log.Printf("RemoteEchoServerDelegate.Quit...")
	for _, stub := range delegate.remoteEchoFactory.stubs {
		stub.Close()
	}
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&RemoteEchoServerDelegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
