#!mojo mojo:dart_content_handler
// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'dart:async';

import 'gen/dart-gen/mojom/lib/mojo/examples/fortune.mojom.dart';
import 'package:v23proxy/client.dart' as v23proxy;
import 'package:mojo/application.dart';
import 'package:mojo/core.dart';

class Fortune extends Application {
  final FortuneProxy fortuneProxy = new FortuneProxy.unbound();

  Fortune.fromHandle(MojoHandle handle) : super.fromHandle(handle);

  void initialize(List<String> args, String url) {
    run(args, url);
  }

  Future run(List<String> args, String url) async {
    // args[0] is the mojo name
    // args[1] is the remote endpoint.
    // args[2+] is the phrase to add to the fortune service. If omitted, we'll get instead.
    print("$url Fortune");
    print(args);

    String remoteEndpoint = args[1];
    if (remoteEndpoint == null) {
      throw new Exception('A remote endpoint must be specified');
    }

    // The phrase is join args[2:]
    String fortunestr = args.length <= 2 ? null : args.sublist(2).join(" ");

    v23proxy.connectToRemoteService(this, fortuneProxy, remoteEndpoint);

    if (fortunestr == null) {
      // Get Fortune
      print("Asking fortune server for a fortune...");
      var response = await fortuneProxy.ptr.get(fortunestr);
      print("Received fortune: ${response.value}");
    } else {
      // Add Fortune
      print("Adding '${fortunestr}' to the fortune server...");
      var response = await fortuneProxy.ptr.add(fortunestr);
      print("Added fortune successfully: ${response}");
    }

    await this.closeApplication();
  }

  Future closeApplication() async {
    print("Closing proxy...");
    await fortuneProxy.close();
    print("Closing application...");
    await this.close();

    assert(MojoHandle.reportLeakedHandles());
  }
}

main(List args) {
  MojoHandle appHandle = new MojoHandle(args[0]);
  new Fortune.fromHandle(appHandle);
}
