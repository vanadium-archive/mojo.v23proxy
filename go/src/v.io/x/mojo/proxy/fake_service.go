// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"

	"v.io/v23/context"
	"v.io/v23/rpc"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"

	"v.io/v23/vdl"
	"v.io/v23/vdlroot/signature"
	"v.io/x/mojo/transcoder"

	"mojo/public/interfaces/bindings/mojom_types"
	"mojo/public/interfaces/bindings/service_describer"
)

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
	inargs := make([]*vdl.Value, numArgs)
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
// results (a struct) to Vanadium results (a slice of *vdl.Value).
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
	/*log.Printf("callRemoteSignature: Prepare payload and header")

	// Prepare the input for the mojo call.
	// This consists of a payload and a header for the RemoteSignature.
	payload := mojom_types.SignatureInput{}
	header := bindings.MessageHeader{
		Type:      0xffffffff,                          // Signature is always type 0xffffffff
		Flags:     bindings.MessageExpectsResponseFlag, // It always has a response.
		RequestId: fs.ids.Count(),
	}

	log.Printf("callRemoteSignature: Encode payload and header")

	var message *bindings.Message
	if message, err = bindings.EncodeMessage(header, &payload); err != nil {
		return response, fmt.Errorf("can't encode request: %v", err.Error())
	}

	log.Printf("callRemoteSignature => callRemoteGeneric")

	outMessage, err := fs.callRemoteGeneric(message)
	if err != nil {
		return response, err
	}

	log.Printf("callRemoteSignature: Decode response")

	if err = outMessage.DecodePayload(&response); err != nil {
		return
	}

	return response, nil*/

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
func (fs fakeService) callRemoteGeneric(ctx *context.T, message *bindings.Message) (outMessage *bindings.Message, err error) {
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

// callRemoteMethod calls the method remotely in a generic way.
// Produces []*vdl.Value at the end for the invoker to return.
func (fs fakeService) callRemoteMethod(ctx *context.T, method string, mi mojom_types.MojomInterface, desc map[string]mojom_types.UserDefinedType, argptrs []interface{}) ([]*vdl.Value, error) {
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
	var rqId uint64
	var flag uint32
	if mm.ResponseParams != nil {
		rqId = fs.ids.Count()
		flag = bindings.MessageExpectsResponseFlag
	} else {
		flag = bindings.MessageNoFlag
	}

	header := bindings.MessageHeader{
		Type:      ordinal,
		Flags:     flag,
		RequestId: rqId,
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

	// Handle the 0 out-arg case first.
	if mm.ResponseParams == nil {
		if err = fs.router.Accept(message); err != nil {
			return nil, err
		}
		return make([]*vdl.Value, 0), nil
	}

	// Otherwise, make a generic call with the message.
	outMessage, err := fs.callRemoteGeneric(ctx, message)
	if err != nil {
		return nil, err
	}

	// Decode the *vdl.Value from the mojom bytes and mojom type.
	outType, err := transcoder.MojomStructToVDLType(*mm.ResponseParams, desc)
	if err != nil {
		return nil, err
	}
	var outVdlValue *vdl.Value
	if err := transcoder.MojomToVdl(outMessage.Payload, outType, &outVdlValue); err != nil {
		return nil, fmt.Errorf("transcoder.MojoToVom failed: %v", err)
	}

	// Then split the *vdl.Value (struct) into []*vdl.Value
	response := splitVdlValueByMojomType(outVdlValue, outType)
	return response, nil
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
