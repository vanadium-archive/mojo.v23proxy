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
	"mojo/public/interfaces/bindings/mojom_types"
	"mojo/public/interfaces/bindings/service_describer"

	"mojom/v23serverproxy"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/rpc"
	"v.io/v23/security"
	"v.io/v23/vdl"
	"v.io/v23/vdlroot/signature"
	"v.io/x/mojo/proxy/util"
	"v.io/x/mojo/transcoder"
	_ "v.io/x/ref/runtime/factories/roaming"
	"v.io/v23/vom"
)

//#include "mojo/public/c/system/types.h"
import "C"

// As long as fakeService meets the Invoker interface, it is allowed to pass as
// a universal v23 service.
// See the function objectToInvoker in v.io/x/ref/runtime/internal/rpc/server.go
type fakeService struct {
	appctx application.Context
	suffix string
	router *bindings.Router
	ids    bindings.Counter
}

// Prepare is used by the Fake Service to prepare the placeholders for the
// input data.
func (fs fakeService) Prepare(ctx *context.T, method string, numArgs int) (argptrs []interface{}, tags []*vdl.Value, _ error) {
	inargs := make([]*vom.RawBytes, numArgs)
	inptrs := make([]interface{}, len(inargs))
	for i := range inargs {
		inptrs[i] = &inargs[i]
	}
	return inptrs, nil, nil
}

// Wraps the interface request and the name of the requested mojo service.
type v23ServiceRequest struct {
	request bindings.InterfaceRequest
	name    string
}

func (v *v23ServiceRequest) Name() string {
	return v.name
}

func (v *v23ServiceRequest) ServiceDescription() service_describer.ServiceDescription {
	panic("not supported")
}

func (v *v23ServiceRequest) PassMessagePipe() system.MessagePipeHandle {
	return v.request.PassMessagePipe()
}

// Invoke calls the mojom service based on the suffix and converts the mojom
// results (a struct) to Vanadium results (a slice of *vom.RawBytes).
// Note: The argptrs from Prepare are reused here. The vom bytes should have
// been decoded into these argptrs, so there are actual values inside now.
func (fs fakeService) Invoke(ctx *context.T, call rpc.StreamServerCall, method string, argptrs []interface{}) (results []interface{}, _ error) {
	// fs.suffix consists of the mojo url and the application/interface name.
	// The last part should be the name; everything else is the url.
	parts := strings.Split(fs.suffix, "/")
	mojourl := strings.Join(parts[:len(parts)-1], "/") // e.g., mojo:go_remote_echo_server. May be defined in a BUILD.gn file.
	mojoname := parts[len(parts)-1]                    // e.g., mojo::examples::RemoteEcho. Defined from the interface + module.

	// Create the generic message pipe. r is a bindings.InterfaceRequest, and
	// p is a bindings.InterfacePointer.
	r, p := bindings.CreateMessagePipeForMojoInterface()
	v := v23ServiceRequest{
		request: r,
		name:    mojoname,
	} // v is an application.ServiceRequest with mojoname

	// Connect to the mojourl.
	fs.appctx.ConnectToApplication(mojourl).ConnectToService(&v)

	// Then assign a new router the FakeService.
	// This will never conflict because each FakeService is only invoked once.
	fs.router = bindings.NewRouter(p.PassMessagePipe(), bindings.GetAsyncWaiter())
	defer fs.Close_Proxy()

	ctx.Infof("Fake Service Invoke (Remote Signature: %q -- %q)", mojourl, mojoname)

	// Vanadium relies on type information, so we will retrieve that first.
	mojomInterface, desc, err := fs.callRemoteSignature(mojourl, mojoname)
	if err != nil {
		return nil, err
	}

	ctx.Infof("Fake Service Invoke Signature %v", mojomInterface)
	ctx.Infof("Fake Service Invoke (Remote Method: %v)", method)

	// With the type information, we can make the method call to the remote interface.
	methodResults, err := fs.callRemoteMethod(ctx, method, mojomInterface, desc, argptrs)
	if err != nil {
		ctx.Errorf("Method called failed: %v", err)
		return nil, err
	}

	ctx.Infof("Fake Service Invoke Results %v", methodResults)

	// Convert methodResult to results.
	results = make([]interface{}, len(methodResults))
	for i := range methodResults {
		results[i] = &methodResults[i]
	}
	return results, nil
}

func (fs fakeService) Close_Proxy() {
	fs.router.Close()
}

// callRemoteSignature obtains type and header information from the remote
// mojo service. Remote mojo interfaces all define a signature method.
func (fs fakeService) callRemoteSignature(mojourl string, mojoname string) (mojomInterface mojom_types.MojomInterface, desc map[string]mojom_types.UserDefinedType, err error) {
	// TODO(afandria): The service_describer mojom file defines the constant, but
	// it is not actually present in the generated code:
	// https://github.com/domokit/mojo/issues/469
	// serviceDescriberInterfaceName := "_ServiceDescriber"

	r, p := service_describer.CreateMessagePipeForServiceDescriber()
	fs.appctx.ConnectToApplication(mojourl).ConnectToService(&r)
	sDescriber := service_describer.NewServiceDescriberProxy(p, bindings.GetAsyncWaiter())
	defer sDescriber.Close_Proxy()

	r2, p2 := service_describer.CreateMessagePipeForServiceDescription()
	err = sDescriber.DescribeService(mojoname, r2)
	if err != nil {
		return
	}
	sDescription := service_describer.NewServiceDescriptionProxy(p2, bindings.GetAsyncWaiter())
	defer sDescription.Close_Proxy()

	mojomInterface, err = sDescription.GetTopLevelInterface()
	if err != nil {
		return
	}
	descPtr, err := sDescription.GetAllTypeDefinitions()
	if err != nil {
		return
	}
	return mojomInterface, *descPtr, nil
}

// A helper function that sends a remote message that expects a response.
func (fs fakeService) callRemoteWithResponse(ctx *context.T, message *bindings.Message) (outMessage *bindings.Message, err error) {
	ctx.Infof("callRemoteGeneric: Send message along the router")

	readResult := <-fs.router.AcceptWithResponse(message)
	if err = readResult.Error; err != nil {
		return
	}

	ctx.Infof("callRemoteGeneric: Audit response message header flag")
	// The message flag we receive back must be a bindings.MessageIsResponseFlag
	if readResult.Message.Header.Flags != bindings.MessageIsResponseFlag {
		err = &bindings.ValidationError{bindings.MessageHeaderInvalidFlags,
			fmt.Sprintf("invalid message header flag: %v", readResult.Message.Header.Flags),
		}
		return
	}

	ctx.Infof("callRemoteGeneric: Audit response message header type")
	// While the mojo service we called into will return a header whose
	// type must match our outgoing one.
	if got, want := readResult.Message.Header.Type, message.Header.Type; got != want {
		err = &bindings.ValidationError{bindings.MessageHeaderUnknownMethod,
			fmt.Sprintf("invalid method in response: expected %v, got %v", want, got),
		}
		return
	}

	return readResult.Message, nil
}

type mojoService struct {
	delegate *delegate
}

func (r *mojoService) Endpoints() (endpoints []string, err error) {
	endpointObjs := r.delegate.v23Server.Status().Endpoints
	endpoints = make([]string, len(endpointObjs))
	for i, endpointObj := range endpointObjs {
		endpoints[i] = endpointObj.String()
	}
	return endpoints, nil
}

// callRemoteMethod calls the method remotely in a generic way.
// Produces []*vom.RawBytes at the end for the invoker to return.
func (fs fakeService) callRemoteMethod(ctx *context.T, method string, mi mojom_types.MojomInterface, desc map[string]mojom_types.UserDefinedType, argptrs []interface{}) ([]*vom.RawBytes, error) {
	// We need to parse the signature result to get the method relevant info out.
	found := false
	var ordinal uint32
	for ord, mm := range mi.Methods {
		if *mm.DeclData.ShortName == method {
			ordinal = ord
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("callRemoteMethod: method %s does not exist", method)
	}

	mm := mi.Methods[ordinal]

	// A void function must have request id of 0, whereas one with response params
	// should  have a unique request id.
	header := bindings.MessageHeader{
		Type:      ordinal,
		Flags:     bindings.MessageExpectsResponseFlag,
		RequestId: fs.ids.Count(),
	}

	// Now produce the *bindings.Message that we will send to the other side.
	inType, err := transcoder.MojomStructToVDLType(mm.Parameters, desc)
	if err != nil {
		return nil, err
	}
	message, err := encodeMessageFromVom(header, argptrs, inType)
	if err != nil {
		return nil, err
	}

	// Otherwise, make a generic call with the message.
	outMessage, err := fs.callRemoteWithResponse(ctx, message)
	if err != nil {
		return nil, err
	}

	// Decode the *vom.RawBytes from the mojom bytes and mojom type.
	outType, err := transcoder.MojomStructToVDLType(*mm.ResponseParams, desc)
	if err != nil {
		return nil, err
	}
	target := util.StructSplitTarget()
	if err := transcoder.FromMojo(target, outMessage.Payload, outType); err != nil {
		return nil, fmt.Errorf("transcoder.FromMojo failed: %v", err)
	}
	return target.Fields(), nil
}


func encodeMessageFromVom(header bindings.MessageHeader, argptrs []interface{}, t *vdl.Type) (*bindings.Message, error) {
	// Convert argptrs into their true form: []*vom.RawBytes
	inargs := make([]*vom.RawBytes, len(argptrs))
	for i := range argptrs {
		inargs[i] = *argptrs[i].(**vom.RawBytes)
	}

	encoder := bindings.NewEncoder()
	if err := header.Encode(encoder); err != nil {
		return nil, err
	}
	if bytes, handles, err := encoder.Data(); err != nil {
		return nil, err
	} else {
		target := transcoder.ToMojomTarget()
		if err := util.JoinRawBytesAsStruct(target, t, inargs); err != nil {
			return nil, err
		}
		moreBytes := target.Bytes()
		// Append the encoded "payload" to the end of the slice.
		bytes = append(bytes, moreBytes...)

		return &bindings.Message{
			Header:  header,
			Bytes:   bytes,
			Handles: handles,
			Payload: moreBytes,
		}, nil
	}
}

// The fake service has no signature.
func (fs fakeService) Signature(ctx *context.T, call rpc.ServerCall) ([]signature.Interface, error) {
	ctx.Infof("Fake Service Signature???")
	return nil, nil
}

// The fake service knows nothing about method signatures.
func (fs fakeService) MethodSignature(ctx *context.T, call rpc.ServerCall, method string) (signature.Method, error) {
	ctx.Infof("Fake Service Method Signature???")
	return signature.Method{}, nil
}

// The fake service will never need to glob.
func (fs fakeService) Globber() *rpc.GlobState {
	log.Printf("Fake Service Globber???")
	return nil
}

type dispatcher struct {
	appctx application.Context
}

func (v23pd *dispatcher) Lookup(ctx *context.T, suffix string) (interface{}, security.Authorizer, error) {
	ctx.Infof("dispatcher.Lookup for suffix: %s", suffix)
	return fakeService{
		appctx: v23pd.appctx,
		suffix: suffix,
		ids:    bindings.NewCounter(),
	}, security.AllowEveryone(), nil
}

type delegate struct {
	ctx       *context.T
	shutdown  v23.Shutdown
	stubs     []*bindings.Stub
	v23Server rpc.Server
}

func (delegate *delegate) Initialize(context application.Context) {
	// Start up v23 whenever a v23proxy is begun.
	// This is done regardless of whether we are initializing this v23proxy for use
	// as a client or as a server.
	ctx, shutdown := v23.Init(context)
	delegate.ctx = ctx
	delegate.shutdown = shutdown
	ctx.Infof("delegate.Initialize...")

	// TODO(alexfandrianto): Does Mojo stop us from creating too many v23proxy?
	// Is it 1 per shell? Ideally, each device will only serve 1 of these v23proxy,
	// but it is not problematic to have extra.
	_, s, err := v23.WithNewDispatchingServer(ctx, "", &dispatcher{
		appctx: context,
	})
	if err != nil {
		ctx.Fatal("Error serving service: ", err)
	}
	delegate.v23Server = s
	fmt.Println("Listening at:", s.Status().Endpoints[0].Name())
}

func (delegate *delegate) Create(request v23serverproxy.V23ServerProxy_Request) {
	svc := &mojoService{delegate: delegate}
	v23Stub := v23serverproxy.NewV23ServerProxyStub(request, svc, bindings.GetAsyncWaiter())
	delegate.stubs = append(delegate.stubs, v23Stub)

	go func() {
		// Read header message
		if err := v23Stub.ServeRequest(); err != nil {
			connectionError, ok := err.(*bindings.ConnectionError)
			if !ok || !connectionError.Closed() {
				delegate.ctx.Errorf("%v", err)
			}
			return
		}
	}()
}

func (delegate *delegate) AcceptConnection(connection *application.Connection) {
	delegate.ctx.Infof("delegate.AcceptConnection...")
	connection.ProvideServices(&v23serverproxy.V23ServerProxy_ServiceFactory{delegate})
}

func (delegate *delegate) Quit() {
	delegate.ctx.Infof("delegate.Quit...")
	for _, stub := range delegate.stubs {
		stub.Close()
	}
	delegate.shutdown()
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&delegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
