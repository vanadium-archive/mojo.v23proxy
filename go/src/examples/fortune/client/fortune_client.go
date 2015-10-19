// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	v23 "v.io/x/mojo/client"

	"mojom/examples/fortune"
)

//#include "mojo/public/c/system/types.h"
import "C"

type FortuneClientDelegate struct{}

// Connects to the v23proxy and calls Get (or Add(ADD_FORTUNE) on REMOTE_ENDPOINT.
// Receives a response, if relevant, then exits.
func (delegate *FortuneClientDelegate) Initialize(ctx application.Context) {
	remoteEndpoint := os.Getenv("REMOTE_ENDPOINT")
	addFortune := os.Getenv("ADD_FORTUNE")

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
