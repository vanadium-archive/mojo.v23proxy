// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

library v23proxy;

import 'gen/dart-gen/mojom/lib/mojo/bindings/types/v23clientproxy.mojom.dart';

import 'package:mojo/application.dart' as application;
import 'package:mojo/bindings.dart' as bindings;
import 'package:mojo/core.dart' as core;

void connectToRemoteService(application.Application app,
  bindings.ProxyBase proxy, String v23Name) {

  // A pipe must be prepared between the given proxy and the v23proxy.
  core.MojoMessagePipe pipe = new core.MojoMessagePipe();
  proxy.impl.bind(pipe.endpoints[0]);

  V23ClientProxyProxy v23proxy = new V23ClientProxyProxy.unbound();
  app.connectToService("https://mojo.v.io/v23clientproxy.mojo", v23proxy);

  // Due to mojom type generation limitations, the proxy may not always have
  // a service description. To avoid issues with dartanalyzer, we use 'dynamic'.
  // TODO(alexfandrianto): Update once this changes.
  dynamic dynproxyimpl = proxy.impl;

  // This is a service_describer.ServiceDescription.
  var serviceDescription = dynproxyimpl.serviceDescription;
  Function identityResponseFactory = (v) => v;

  v23proxy.ptr.setupClientProxy(
    v23Name,
    serviceDescription.getTopLevelInterface(identityResponseFactory),
    serviceDescription.getAllTypeDefinitions(identityResponseFactory),
    proxy.serviceName,
    pipe.endpoints[1]);
}
