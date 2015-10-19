package main

import (
	"flag"
	"fmt"
	"log"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/options"
	"v.io/v23/security"
	"v.io/v23/vdl"

	"mojom/v23proxy"

	"mojo/public/go/application"
	"mojo/public/go/bindings"
	"mojo/public/go/system"
	"mojo/public/interfaces/bindings/mojom_types"

	"v.io/x/mojo/transcoder"
	_ "v.io/x/ref/runtime/factories/static"
)

//#include "mojo/public/c/system/types.h"
import "C"

func init() {
	flag.String("child-connection-id", "", "")
	flag.String("mojo-platform-channel-handle", "", "")
}

type v23HeaderReceiver struct {
	delegate    *V23ServerDelegate
	v23Name     string
	ifaceSig    mojom_types.MojomInterface
	desc        map[string]mojom_types.UserDefinedType
	serviceName string
	handle      system.MessagePipeHandle
}

func (r *v23HeaderReceiver) SetupProxy(v23Name string, ifaceSig mojom_types.MojomInterface, desc map[string]mojom_types.UserDefinedType, serviceName string, handle system.MessagePipeHandle) (err error) {
	log.Printf("[server] In SetupProxy(%s, %v, %v, %s, %v)\n", v23Name, ifaceSig, desc, serviceName, handle)
	r.v23Name = v23Name
	r.ifaceSig = ifaceSig
	r.desc = desc
	r.serviceName = serviceName
	r.handle = handle

	go func() {
		connector := bindings.NewConnector(r.handle, bindings.GetAsyncWaiter())

		// Read generic calls in a loop
		stub := &genericStub{
			header:    r,
			ctx:       r.delegate.ctx,
			connector: connector,
		}
		bindingStub := bindings.NewStub(connector, stub)
		for {
			if err := bindingStub.ServeRequest(); err != nil {
				connectionError, ok := err.(*bindings.ConnectionError)
				if !ok || !connectionError.Closed() {
					log.Println(err)
				}
				break
			}
		}
		r.delegate.stubs = append(r.delegate.stubs, bindingStub)
	}()
	return nil
}

// TODO(alexfandrianto): This assumes that bindings.Encoder has the method
// WriteRawBytes. See the comment block below.
// type byteCopyingPayload []byte

// func (bcp byteCopyingPayload) Encode(encoder *bindings.Encoder) error {
// 	encoder.WriteRawBytes(bcp)
// 	return nil
// }

// func (bcp byteCopyingPayload) Decode(decoder *bindings.Decoder) error {
// 	panic("not supported")
// }

type genericStub struct {
	header    *v23HeaderReceiver
	ctx       *context.T
	connector *bindings.Connector
}

func (s *genericStub) Accept(message *bindings.Message) (err error) {
	if _, ok := s.header.ifaceSig.Methods[message.Header.Type]; !ok {
		return fmt.Errorf("Method had index %d, but interface only has %d methods",
			message.Header.Type, len(s.header.ifaceSig.Methods))
	}

	methodSig := s.header.ifaceSig.Methods[message.Header.Type]
	methodName := *methodSig.DeclData.ShortName
	// Should we perform validation of flags like generated methods?
	// Does this handle 0-arg methods?

	messageBytes := message.Payload

	response, err := s.Call(s.header.v23Name, methodName, messageBytes, methodSig.Parameters, methodSig.ResponseParams)
	if err != nil {
		return err
	}

	// TODO(alexfandrianto): This assumes that bindings.Encoder has the method
	// WriteRawBytes. We will need to add this to Mojo ourselves.
	// func (e *Encoder) WriteRawBytes(data []byte) {
	// 	first := e.end
	// 	e.claimData(align(len(data), defaultAlignment))
	// 	copy(e.buf[first:], data)
	// }
	//
	// See: https://codereview.chromium.org/1416433002/

	responseHeader := bindings.MessageHeader{
		Type:      message.Header.Type,
		Flags:     bindings.MessageIsResponseFlag,
		RequestId: message.Header.RequestId,
	}
	// responseMessage, err := bindings.EncodeMessage(responseHeader, byteCopyingPayload(response))
	// if err != nil {
	// 	return err
	// }
	// return s.connector.WriteMessage(responseMessage)

	// TODO(alexfandrianto): Replace this block with the above.
	encoder := bindings.NewEncoder()
	if err := responseHeader.Encode(encoder); err != nil {
		return err
	}
	if bytes, handles, err := encoder.Data(); err != nil {
		return err
	} else {
		// response is our payload; append to the end of our slice.
		bytes = append(bytes, response...)

		// This is analogous to bindings.newMessage
		responseMessage := &bindings.Message{
			Header:  responseHeader,
			Bytes:   bytes,
			Handles: handles,
			Payload: response,
		}
		return s.connector.WriteMessage(responseMessage)
	}
}

func (s *genericStub) Call(name, method string, value []byte, inParamsType mojom_types.MojomStruct, outParamsType *mojom_types.MojomStruct) ([]byte, error) {
	log.Printf("server: %s.%s: %#v", name, method, inParamsType)

	inVType := transcoder.MojomStructToVDLType(inParamsType, s.header.desc)
	var outVType *vdl.Type
	if outParamsType != nil {
		outVType = transcoder.MojomStructToVDLType(*outParamsType, s.header.desc)
	}

	// Decode the vdl.Value from the mojom bytes and mojom type.
	inVdlValue, err := transcoder.DecodeValue(value, inVType)
	if err != nil {
		return nil, fmt.Errorf("transcoder.DecodeValue failed: %v", err)
	}

	// inVdlValue is a struct, but we need to send []interface.
	inargs := splitVdlValueByMojomType(inVdlValue, inVType)
	inargsIfc := make([]interface{}, len(inargs))
	for i := range inargs {
		inargsIfc[i] = inargs[i]
	}

	// We know that the v23proxy (on the other side) will give us back a bunch of
	// data in []interface{}. so we'll want to decode them into *vdl.Value.
	log.Printf("%s %v\n", method, outParamsType)
	outargs := make([]*vdl.Value, len(outParamsType.Fields))
	outptrs := make([]interface{}, len(outargs))
	for i := range outargs {
		outptrs[i] = &outargs[i]
	}

	// Now, run the call without any authorization.
	if err := v23.GetClient(s.ctx).Call(s.ctx, name, method, inargsIfc, outptrs, options.ServerAuthorizer{security.AllowEveryone()}); err != nil {
		return nil, err
	}

	// Now convert the []interface{} into a *vdl.Value (struct).
	outVdlValue := combineVdlValueByMojomType(outargs, outVType)

	// Finally, encode this *vdl.Value (struct) into mojom bytes and send the response.
	result, err := transcoder.Encode(outVdlValue)
	if err != nil {
		return nil, fmt.Errorf("transcoder.Encode failed: %v", err)
	}
	return result, nil
}

type V23ServerDelegate struct {
	ctx      *context.T
	stubs    []*bindings.Stub
	shutdown v23.Shutdown
}

func (delegate *V23ServerDelegate) Initialize(context application.Context) {
	log.Printf("V23ServerDelegate.Initialize...")

	// Start up v23 whenever a v23proxy is begun.
	// This is done regardless of whether we are initializing this v23proxy for use
	// as a client or as a server.
	ctx, shutdown := v23.Init()
	delegate.ctx = ctx
	delegate.shutdown = shutdown

	// TODO(alexfandrianto): Does Mojo stop us from creating too many v23proxy?
	// Is it 1 per shell? Ideally, each device will only serve 1 of these v23proxy,
	// but it is not problematic to have extra.
	/*s := MakeServer(ctx)
	err := s.ServeDispatcher("", &V23ProxyDispatcher{
		appctx: context,
	})*/
	_, s, err := v23.WithNewDispatchingServer(ctx, "", &V23ProxyDispatcher{
		appctx: context,
	})
	if err != nil {
		log.Panic("Error serving service: ", err)
	}

	endpoints := s.Status().Endpoints
	fmt.Printf("Listening at: /%v\n", endpoints[0])
}

/*func MakeServer(ctx *context.T) rpc.Server {
	s, err := v23.NewServer(ctx)
	if err != nil {
		log.Panic("Failure creating server: ", err)
	}

	endpoints, err := s.Listen(v23.GetListenSpec(ctx))
	if err != nil {
		log.Panic("Error listening to service: ", err)
	}
	fmt.Printf("Listening at: /%v\n", endpoints[0])
	return s
}*/

type V23ProxyDispatcher struct {
	appctx application.Context
}

func (v23pd *V23ProxyDispatcher) Lookup(ctx *context.T, suffix string) (interface{}, security.Authorizer, error) {
	log.Printf("Dispatcher: %s", suffix)
	return fakeService{
		appctx: v23pd.appctx,
		suffix: suffix,
		ids:    bindings.NewCounter(),
	}, security.AllowEveryone(), nil
}

func (delegate *V23ServerDelegate) Create(request v23proxy.V23_Request) {
	headerReceiver := &v23HeaderReceiver{delegate: delegate}
	v23Stub := v23proxy.NewV23Stub(request, headerReceiver, bindings.GetAsyncWaiter())
	delegate.stubs = append(delegate.stubs, v23Stub)

	go func() {
		// Read header message
		if err := v23Stub.ServeRequest(); err != nil {
			connectionError, ok := err.(*bindings.ConnectionError)
			if !ok || !connectionError.Closed() {
				log.Println(err)
			}
			return
		}
	}()
}

func (delegate *V23ServerDelegate) AcceptConnection(connection *application.Connection) {
	log.Printf("V23ServerDelegate.AcceptConnection...")
	connection.ProvideServices(&v23proxy.V23_ServiceFactory{delegate})
}

func (delegate *V23ServerDelegate) Quit() {
	log.Printf("V23ServerDelegate.Quit...")
	for _, stub := range delegate.stubs {
		stub.Close()
	}

	delegate.shutdown()
}

//export MojoMain
func MojoMain(handle C.MojoHandle) C.MojoResult {
	application.Run(&V23ServerDelegate{}, system.MojoHandle(handle))
	return C.MOJO_RESULT_OK
}

func main() {
}
