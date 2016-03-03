#!mojo mojo:dart_content_handler
// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'dart:async';
import 'dart:math' as math;

import 'gen/dart-gen/mojom/lib/mojo/examples/fortune.mojom.dart';
import 'package:mojo/application.dart';
import 'package:mojo/core.dart';

class FortuneImpl implements Fortune {
  FortuneStub _stub; // TODO(alexfandrianto): Do we need to _stub.close()?
  Application _application; // This isn't needed, is it...?
  final List<String> _fortunes;

  FortuneImpl(this._application, MojoMessagePipeEndpoint endpoint,
      this._fortunes) {
    _stub = new FortuneStub.fromEndpoint(endpoint, this);
  }

  @override
  dynamic get([Function responseFactory = null]) {
    print("Dart Server: Get was called.");
    String fortune = _fortunes[new math.Random().nextInt(_fortunes.length)];
    print("Dart Server: Going to send back ${fortune}");
    return responseFactory(fortune);
  }

  @override
  dynamic add(String wisdom,[Function responseFactory = null]) {
    print("Dart Server: Add called with '${wisdom}'");
    _fortunes.add(wisdom);
    return responseFactory();
  }
}

class FortuneServer extends Application {
  final List<String> _fortunes = <String>[
   "You will reach the heights of success.",
   "Conquer your fears or they will conquer you.",
   "Today is your lucky day!",
  ];

  FortuneServer.fromHandle(MojoHandle handle) : super.fromHandle(handle);

  @override
  void acceptConnection(String requestorUrl, String resolvedUrl,
      ApplicationConnection connection) {
    connection.provideService(Fortune.serviceName,
        (endpoint) => new FortuneImpl(this, endpoint, _fortunes),
        description: FortuneStub.serviceDescription);
  }
}

main(List args, Object handleToken) {
  MojoHandle appHandle = new MojoHandle(handleToken);
  new FortuneServer.fromHandle(appHandle);
}
