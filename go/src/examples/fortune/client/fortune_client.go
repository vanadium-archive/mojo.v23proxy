// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	v23 "v.io/x/mojo/client"

	"mojom/examples/fortune"
)

//#include "mojo/public/c/system/handle.h"
import "C"

type FortuneClientDelegate struct{}

// When running fortune_client, ctx.Args() should contain:
// 0: mojo app name
// 1: remote endpoint
// 2+: (optional) fortune to add
// If the fortune to add is omitted, then the fortune_client will Get a fortune.
// Otherwise, it will Add the given fortune.
func (delegate *FortuneClientDelegate) Initialize(ctx application.Context) {
	// Parse the arguments.
	remoteEndpoint := ctx.Args()[1]
	addFortune := strings.Join(ctx.Args()[2:], " ")

	log.Printf("FortuneClientDelegate.Initialize... %s", remoteEndpoint)
	fortuneRequest, fortunePointer := fortune.CreateMessagePipeForFortune()
	v23.ConnectToRemoteService(ctx, &fortuneRequest, remoteEndpoint)
	fortuneProxy := fortune.NewFortuneProxy(fortunePointer, bindings.GetAsyncWaiter())

	if addFortune != "" {
		log.Printf("FortuneClientDelegate.Initialize calling Add...")

		if err := fortuneProxy.Add(addFortune); err != nil {
			log.Println(err)
		} else {
			fmt.Printf("client added: %s\n", addFortune)
		}
	} else {
		log.Printf("FortuneClientDelegate.Initialize calling Get...")
		response, err := fortuneProxy.Get()
		if response != "" {
			fmt.Printf("client (get): %s\n", response)
		} else {
			log.Println(err)
		}
	}

	fortuneProxy.Close_Proxy()
	ctx.Close()
}

func (delegate *FortuneClientDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("FortuneClientDelegate.AcceptConnection...")
	connection.Close()
}

func (delegate *FortuneClientDelegate) Quit() {
	log.Printf("FortuneClientDelegate.Quit...")
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&FortuneClientDelegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
