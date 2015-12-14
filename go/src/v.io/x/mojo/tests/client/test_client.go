// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	"mojom/tests/end_to_end_test"

	v23 "v.io/x/mojo/client"
	"v.io/x/mojo/tests/expected"
)

//#include "mojo/public/c/system/types.h"
import "C"

func init() {
	// Add flag placeholders to suppress warnings on unhandled mojo flags.
	flag.String("child-connection-id", "", "")
	flag.String("platform-channel-handle-info", "", "")
}

func TestSimple(t *testing.T, ctx application.Context) {
	proxy := createProxy(ctx)
	defer proxy.Close_Proxy()

	value, err := proxy.Simple(expected.SimpleRequestA)
	if err != nil {
		t.Fatal(err)
	}
	if value != expected.SimpleResponseValue {
		t.Errorf("expected %v, but got %v", expected.SimpleResponseValue, value)
	}
}

func TestMultiArgs(t *testing.T, ctx application.Context) {
	proxy := createProxy(ctx)
	defer proxy.Close_Proxy()

	x, y, err := proxy.MultiArgs(expected.MultiArgsRequestA, expected.MultiArgsRequestB, expected.MultiArgsRequestC, expected.MultiArgsRequestD)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(x, expected.MultiArgsResponseX) {
		t.Errorf("expected %v, but got %v", expected.MultiArgsResponseX, x)
	}
	if y != expected.MultiArgsResponseY {
		t.Errorf("expected %v, but got %v", expected.MultiArgsResponseY, y)
	}
}

func TestReuseProxy(t *testing.T, ctx application.Context) {
	proxy := createProxy(ctx)
	defer proxy.Close_Proxy()

	value, err := proxy.Simple(expected.SimpleRequestA)
	if err != nil {
		t.Fatal(err)
	}
	if value != expected.SimpleResponseValue {
		t.Errorf("expected %v, but got %v", expected.SimpleResponseValue, value)
	}
	x, y, err := proxy.MultiArgs(expected.MultiArgsRequestA, expected.MultiArgsRequestB, expected.MultiArgsRequestC, expected.MultiArgsRequestD)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(x, expected.MultiArgsResponseX) {
		t.Errorf("expected %v, but got %v", expected.MultiArgsResponseX, x)
	}
	if y != expected.MultiArgsResponseY {
		t.Errorf("expected %v, but got %v", expected.MultiArgsResponseY, y)
	}
}

// This test stores a value on the server (through a no-out args RPC)
// and calls a no-in args RPC to retrieve the value and confirm
// it matches the value originally sent.
func TestNoOutArgs(t *testing.T, ctx application.Context) {
	const msg = "message-for-no-return"

	proxy := createProxy(ctx)
	defer proxy.Close_Proxy()

	err := proxy.NoOutArgsPut(msg)
	if err != nil {
		t.Fatal(err)
	}

	outMsg, err := proxy.FetchMsgFromNoOutArgsPut()
	if err != nil {
		t.Fatal(err)
	}
	if outMsg != msg {
		t.Errorf("expected %v, but got %v", msg, outMsg)
	}
}

func createProxy(ctx application.Context) *end_to_end_test.V23ProxyTest_Proxy {
	// Parse arguments. Note: May panic if not enough args are given.
	remoteName := ctx.Args()[1]

	r, p := end_to_end_test.CreateMessagePipeForV23ProxyTest()
	v23.ConnectToRemoteService(ctx, &r, remoteName)
	return end_to_end_test.NewV23ProxyTestProxy(p, bindings.GetAsyncWaiter())
}

type TestClientDelegate struct{}

func funcName(f func(*testing.T, application.Context)) string {
	qualified := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	return qualified[strings.LastIndex(qualified, ".")+1:]
}
func convertTests(testFuncs []func(*testing.T, application.Context), ctx application.Context) []testing.InternalTest {
	tests := make([]testing.InternalTest, len(testFuncs))
	for i, _ := range testFuncs {
		f := testFuncs[i]
		tests[i] = testing.InternalTest{
			Name: funcName(f),
			F:    func(t *testing.T) { f(t, ctx) },
		}
	}
	return tests
}

func (delegate *TestClientDelegate) Initialize(ctx application.Context) {
	log.Printf("TestClientDelegate.Initialize...")

	tests := []func(*testing.T, application.Context){
		TestSimple, TestMultiArgs, TestReuseProxy, TestNoOutArgs,
	}

	matchAllTests := func(pat, str string) (bool, error) { return true, nil }
	exitCode := testing.MainStart(matchAllTests, convertTests(tests, ctx), nil, nil).Run()
	if exitCode == 0 {
		fmt.Printf("%s\n", expected.SuccessMessage)
	} else {
		fmt.Printf("%s\n", expected.FailureMessage)
	}

	ctx.Close()
	os.Exit(exitCode)
}

func (delegate *TestClientDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("TestClientDelegate.AcceptConnection...")
	connection.Close()
}

func (delegate *TestClientDelegate) Quit() {
	log.Printf("TestClientDelegate.Quit...")
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&TestClientDelegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
