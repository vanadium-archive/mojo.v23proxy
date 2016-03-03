#!mojo mojo:dart_content_handler
// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'dart:async';

import 'expected.dart';
import 'gen/dart-gen/mojom/lib/mojo/v23proxy/tests/end_to_end_test.mojom.dart';

import 'package:v23proxy/client.dart' as v23proxy;
import 'package:mojo/application.dart';
import 'package:mojo/core.dart';

class EndToEndTestClient extends Application {
  final V23ProxyTestProxy testProxy = new V23ProxyTestProxy.unbound();

  EndToEndTestClient.fromHandle(MojoHandle handle) : super.fromHandle(handle);

  void initialize(List<String> args, String url) {
    run(args, url);
  }

  Future testSimple() async {
    print ("Running Test Simple...");
    V23ProxyTestSimpleResponseParams response = await testProxy.ptr.simple(SimpleRequestA);
    if (response.value != SimpleResponseValue) {
      throw "expected $SimpleResponseValue, but got ${response.value}";
    }
    print ("Completed Test Simple! OK");
  }

  Future testMultiArgs() async {
    print ("Running Test MultiArgs...");
    V23ProxyTestMultiArgsResponseParams response1 = await testProxy.ptr.multiArgs(MultiArgsRequestA, MultiArgsRequestB, MultiArgsRequestC, MultiArgsRequestD);
    if (response1.x.toString() != MultiArgsResponseX.toString()) { // compare strings since this is a union
      throw "expected $MultiArgsResponseX, but got ${response1.x}";
    }
    if (response1.y != MultiArgsResponseY) {
      throw "expected $MultiArgsResponseY, but got ${response1.y}";
    }
    print ("Completed Test MultiArgs! OK");
  }

  Future testNoOutArgs() async {
    print ("Running Test NoOutArgs...");
    String expectedMessage = "message-for-no-return";
    V23ProxyTestNoOutArgsPutResponseParams _ = await testProxy.ptr.noOutArgsPut(expectedMessage);
    V23ProxyTestFetchMsgFromNoOutArgsPutResponseParams response3 = await testProxy.ptr.fetchMsgFromNoOutArgsPut();
    if (response3.storedMsg != expectedMessage) {
      throw "expected $expectedMessage, but got ${response3.storedMsg}";
    }
    print ("Completed Test NoOutArgs! OK");
  }

  Future benchmarkSimple() async {
    print ("Running Benchmark Simple...");
    int benchmarkN = 100;
    DateTime start = new DateTime.now();
    for (int i = 0; i < benchmarkN; i++) {
      await testProxy.ptr.simple(SimpleRequestA);
    }
    DateTime end = new DateTime.now();
    print ("Completed Benchmark Simple! OK");

    // Print how long the benchmark took, somewhat mimicking what it would look like in Go.
    print("    $benchmarkN\t   ${(end.difference(start).inMicroseconds * 1000 / benchmarkN).floor()} ns/op");
  }

  Future run(List<String> args, String url) async {
    // args[0] is the mojo name
    // args[1] is the remote endpoint prefixed by -endpoint=
    // args[2] is v23.tcp.address
    // args[3] might be present and if so, it is -test.run=XXXX
    // args[4] might be present and if so, it is -test.bench=.
    print("$url EndtoEndTestClient");
    print(args);

    String remoteEndpoint = args[1].split("=")[1];
    if (remoteEndpoint == null || remoteEndpoint == "") {
      throw new Exception('A remote endpoint must be specified');
    }

    bool isBenchmark = (args.length >= 5 && args[4] == "-test.bench=.");

    v23proxy.connectToRemoteService(this, testProxy, remoteEndpoint);

    if (isBenchmark) {
      await benchmarkSimple();
    } else {
      await testSimple();
      await testMultiArgs();
      await testNoOutArgs();
    }

    print(SuccessMessage);

    await this.closeApplication();
  }

  Future closeApplication() async {
    print("Closing proxy...");
    await testProxy.close();
    print("Closing application...");
    await this.close();

    assert(MojoHandle.reportLeakedHandles());
  }
}

main(List args, Object handleToken) {
  MojoHandle appHandle = new MojoHandle(handleToken);
  new EndToEndTestClient.fromHandle(appHandle);
}
