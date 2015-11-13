// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"math/rand"
	"time"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	"mojom/examples/fortune"
)

//#include "mojo/public/c/system/types.h"
import "C"

type FortuneImpl struct {
	wisdom []string   // All known fortunes.
	random *rand.Rand // To pick a random index in 'wisdom'.
}

// Makes an implementation.
func NewFortuneImpl() *FortuneImpl {
	return &FortuneImpl{
		wisdom: []string{
			"You will reach the heights of success.",
			"Conquer your fears or they will conquer you.",
			"Today is your lucky day!",
		},
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (f *FortuneImpl) Add(inValue string) (err error) {
	log.Printf("server Add: %s\n", inValue)

	f.wisdom = append(f.wisdom, inValue)
	return nil
}

func (f *FortuneImpl) Get() (outValue string, err error) {
	log.Printf("server Get\n")

	if len(f.wisdom) == 0 {
		return "[empty]", nil
	}
	return f.wisdom[f.random.Intn(len(f.wisdom))], nil
}

type FortuneServerDelegate struct {
	fortuneFactory FortuneFactory
}

type FortuneFactory struct {
	stubs   []*bindings.Stub
	fortune *FortuneImpl
}

func (delegate *FortuneServerDelegate) Initialize(context application.Context) {
	log.Printf("FortuneServerDelegate.Initialize...")
}

func (ff *FortuneFactory) Create(request fortune.Fortune_Request) {
	log.Printf("FortuneServerDelegate's FortuneFactory.Create...")
	stub := fortune.NewFortuneStub(request, ff.fortune, bindings.GetAsyncWaiter())
	ff.stubs = append(ff.stubs, stub)
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

func (delegate *FortuneServerDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("FortuneServerDelegate.AcceptConnection...")
	connection.ProvideServicesWithDescriber(
		&fortune.Fortune_ServiceFactory{&delegate.fortuneFactory},
	)
}

func (delegate *FortuneServerDelegate) Quit() {
	log.Printf("FortuneServerDelegate.Quit...")
	for _, stub := range delegate.fortuneFactory.stubs {
		stub.Close()
	}
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&FortuneServerDelegate{
		fortuneFactory: FortuneFactory{
			fortune: NewFortuneImpl(),
		},
	}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
