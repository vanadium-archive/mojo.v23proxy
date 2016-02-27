// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import 'gen/dart-gen/mojom/lib/mojo/v23proxy/tests/end_to_end_test.mojom.dart';

// These are constants that the test expects to both receive and return.
const int SimpleRequestA = 123;
const String SimpleResponseValue = "TheValue";
const bool MultiArgsRequestA  = true;
List<double> MultiArgsRequestB = <double>[1.0, 2.0, 3.0];
Map<String, int> MultiArgsRequestC = <String, int>{"X": 1, "Y": 2};
AStruct MultiArgsRequestD = new AStruct()
  ..x = 3
  ..y = 300
  ..z = 129;
AUnion MultiArgsResponseX = new AUnion()
  ..b = "TheUnion";
const String MultiArgsResponseY = "yresponse";

const String SuccessMessage = "ALL TESTS PASSED";
const String FailureMessage = "Failed";