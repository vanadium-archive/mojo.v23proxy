#!mojo mojo:dart_content_handler
// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'dart:async';

import 'gen/dart-gen/mojom/lib/mojo/examples/echo.mojom.dart';
import 'package:v23proxy/client.dart' as v23proxy;
import 'package:mojo/application.dart';
import 'package:mojo/bindings.dart';
import 'package:mojo/core.dart';

class Echo extends Application {
  final RemoteEchoProxy echoProxy = new RemoteEchoProxy.unbound();

  Echo.fromHandle(MojoHandle handle) : super.fromHandle(handle);

  void initialize(List<String> args, String url) {
    run(args, url);
  }

  Future run(List<String> args, String url) async {
    // args[0] is the mojo name
    // args[1] is the remote endpoint.
    // args[2+] are the words to echo.
    print("$url Echo");
    print(args);

    String remoteEndpoint = args[1];
    if (remoteEndpoint == null) {
      throw new Exception('A remote endpoint must be specified');
    }

    // The phrase is join args[2:]
    String echostr = args.length <= 2 ? "hello" : args.sublist(2).join(" ");

    v23proxy.connectToRemoteService(this, echoProxy, remoteEndpoint);

    print("Sending ${echostr} to echo server...");
    var response = await echoProxy.ptr.echoString(echostr);
    // The out arg is called value. It is a string.
    print(response.value != null
        ? "Received: ${response.value}"
        : "Failed to get an echo back");

    print("Attempting EchoX to echo server...");
    // The in args are a List<bool> and struct with str String field.
    var echoXParam1 = <bool>[true, false, false, true];
    var echoXParam2 = new AInArg()..str = "A String";
    response = await echoProxy.ptr.echoX(echoXParam1, echoXParam2);
    // The out arg is a single union called out.
    // out is of type OutArgTypes. Of the int64 and enum, it'll return the enum.
    print(response.out != null
        ? "Received: ${response.out}"
        : "Failed to get an echoX back");

    await this.closeApplication();
  }

  Future closeApplication() async {
    print("Closing proxy...");
    await echoProxy.close();
    print("Closing application...");
    await this.close();

    assert(MojoHandle.reportLeakedHandles());
  }
}

main(List args) {
  MojoHandle appHandle = new MojoHandle(args[0]);
  new Echo.fromHandle(appHandle);
}
