#!mojo mojo:dart_content_handler
// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'dart:async';

import 'expected.dart';
import 'gen/dart-gen/mojom/lib/mojo/v23proxy/tests/end_to_end_test.mojom.dart';

import 'package:mojo/application.dart';
import 'package:mojo/core.dart';

Completer forNoOutArgsPut = new Completer();

class V23ProxyTestImpl implements V23ProxyTest {
  V23ProxyTestStub _stub; // TODO(alexfandrianto): Do we need to _stub.close()?
  Application _application; // This isn't needed, is it...?

  V23ProxyTestImpl(this._application, MojoMessagePipeEndpoint endpoint) {
    print("Creating the v23proxy test impl.");
    _stub = new V23ProxyTestStub.fromEndpoint(endpoint, this);
  }

  @override
  dynamic simple(int a,[Function responseFactory = null]) {
    if (a != SimpleRequestA) {
      throw "expected $SimpleRequestA, but got $a";
    }
    return responseFactory(SimpleResponseValue);
  }

  @override
  dynamic multiArgs(bool a,List<double> b,Map<String, int> c,AStruct d,[Function responseFactory = null]) {
    if (a != MultiArgsRequestA) {
      throw "expected $MultiArgsRequestA, but got $a";
    }
    for (int i = 0; i < MultiArgsRequestB.length; i++) {
      if (b[i] != MultiArgsRequestB[i]) {
        throw "expected $MultiArgsRequestB, but got $b";
      }
    }
    MultiArgsRequestC.forEach((String key, int value) {
      if (c[key] != value) {
        throw "expected $MultiArgsRequestC, but got $c";
      }
    });
    if (d.toString() != MultiArgsRequestD.toString()) {// d is a struct, so we'll just toString both.
      throw "expected $MultiArgsRequestD, but got $d";
    }
    return responseFactory(MultiArgsResponseX, MultiArgsResponseY);
  }

  @override
  dynamic noOutArgsPut(String storedMsg,[Function responseFactory = null]) {
    forNoOutArgsPut.complete(storedMsg);
    return responseFactory();
  }

  @override
  dynamic fetchMsgFromNoOutArgsPut([Function responseFactory = null]) {
    Completer completer = new Completer();

    new Future.delayed(const Duration(seconds: 1)).then((_) {
      completer.completeError("timed out waiting for no return message");
    });

    forNoOutArgsPut.future.then((String answer) {
      if (!completer.isCompleted) {
        completer.complete(responseFactory(answer));
      }
    });
    return completer.future;
  }
}

class EndToEndTestServer extends Application {
  EndToEndTestServer.fromHandle(MojoHandle handle) : super.fromHandle(handle);

  @override
  void acceptConnection(String requestorUrl, String resolvedUrl,
      ApplicationConnection connection) {
    print("Server: Accepting a connection from ${requestorUrl} to ${resolvedUrl}");
    connection.provideService(V23ProxyTest.serviceName,
        (endpoint) => new V23ProxyTestImpl(this, endpoint),
        description: V23ProxyTestStub.serviceDescription);
  }
}

main(List args, Object handleToken) {
  MojoHandle appHandle = new MojoHandle(handleToken);
  new EndToEndTestServer.fromHandle(appHandle);
}
