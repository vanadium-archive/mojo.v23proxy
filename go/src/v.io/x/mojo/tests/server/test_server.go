// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"reflect"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	"mojom/tests/end_to_end_test"

	"v.io/x/mojo/tests/expected"
)

//#include "mojo/public/c/system/types.h"
import "C"

type V23ProxyTestImpl struct{}

func (i *V23ProxyTestImpl) Simple(a int32) (value string, err error) {
	if a != expected.SimpleRequestA {
		return "", fmt.Errorf("expected %v, but got %v", expected.SimpleRequestA, a)
	}
	return expected.SimpleResponseValue, nil
}

func (i *V23ProxyTestImpl) MultiArgs(a bool, b []float32, c map[string]uint8, d end_to_end_test.AStruct) (x end_to_end_test.AUnion, y string, err error) {
	if a != expected.MultiArgsRequestA {
		return nil, "", fmt.Errorf("expected %v, but got %v", expected.MultiArgsRequestA, a)
	}
	if !reflect.DeepEqual(b, expected.MultiArgsRequestB) {
		return nil, "", fmt.Errorf("expected %v, but got %v", expected.MultiArgsRequestB, b)
	}
	if !reflect.DeepEqual(c, expected.MultiArgsRequestC) {
		return nil, "", fmt.Errorf("expected %v, but got %v", expected.MultiArgsRequestC, c)
	}
	if !reflect.DeepEqual(d, expected.MultiArgsRequestD) {
		return nil, "", fmt.Errorf("expected %v, but got %v", expected.MultiArgsRequestD, d)
	}
	return expected.MultiArgsResponseX, expected.MultiArgsResponseY, nil
}

func (i *V23ProxyTestImpl) NoReturn() error {
	// TODO(bprosnitz) The test should fail if the message is not received.
	return nil
}

type V23ProxyTestServerDelegate struct {
	factory V23ProxyTestFactory
}

type V23ProxyTestFactory struct {
	stubs []*bindings.Stub
}

func (delegate *V23ProxyTestServerDelegate) Initialize(context application.Context) {
	log.Printf("V23ProxyTestServerDelegate.Initialize...")
}

func (factory *V23ProxyTestFactory) Create(request end_to_end_test.V23ProxyTest_Request) {
	log.Printf("V23ProxyTestServer's V23ProxyTestFactory.Create...")
	stub := end_to_end_test.NewV23ProxyTestStub(request, &V23ProxyTestImpl{}, bindings.GetAsyncWaiter())
	factory.stubs = append(factory.stubs, stub)
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

func (delegate *V23ProxyTestServerDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("RemoteEchoServerDelegate.AcceptConnection...")
	connection.ProvideServicesWithDescriber(
		&end_to_end_test.V23ProxyTest_ServiceFactory{&delegate.factory},
	)
}

func (delegate *V23ProxyTestServerDelegate) Quit() {
	log.Printf("V23ProxyTestServerDelegate.Quit...")
	for _, stub := range delegate.factory.stubs {
		stub.Close()
	}
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&V23ProxyTestServerDelegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
