#!mojo mojo:dart_content_handler
// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'gen/dart-gen/mojom/lib/mojo/examples/echo.mojom.dart';
import 'package:mojo/application.dart';
import 'package:mojo/core.dart';

class RemoteEchoImpl implements RemoteEcho {
  RemoteEchoStub _stub; // TODO(alexfandrianto): Do we need to _stub.close()?
  Application _application; // This isn't needed, is it...?

  RemoteEchoImpl(this._application, MojoMessagePipeEndpoint endpoint) {
    _stub = new RemoteEchoStub.fromEndpoint(endpoint, this);
  }

  @override
  dynamic echoString(String value,[Function responseFactory]) {
    print("Dart Server: EchoString got ${value}!");
    return responseFactory(value);
  }

  @override
  dynamic echoX(List<bool> arg1, AInArg arg2, [Function responseFactory]) {
    print("Dart Server: EchoX got ${arg1} and ${arg2}");
    return responseFactory(new OutArgTypes()..res = Result.b);
  }
}

class EchoServer extends Application {
  EchoServer.fromHandle(MojoHandle handle) : super.fromHandle(handle);

  @override
  void acceptConnection(String requestorUrl, String resolvedUrl,
      ApplicationConnection connection) {
    connection.provideService(RemoteEcho.serviceName,
        (endpoint) => new RemoteEchoImpl(this, endpoint),
        description: RemoteEchoStub.serviceDescription);
  }
}

main(List args, Object handleToken) {
  MojoHandle appHandle = new MojoHandle(handleToken);
  new EchoServer.fromHandle(appHandle);
}
