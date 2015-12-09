// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package expected

import (
	"mojom/tests/end_to_end_test"
)

var (
	SimpleRequestA      int32 = 123
	SimpleResponseValue       = "TheValue"

	MultiArgsRequestA  = true
	MultiArgsRequestB  = []float32{1, 2, 3}
	MultiArgsRequestC  = map[string]uint8{"X": 1, "Y": 2}
	MultiArgsRequestD  = end_to_end_test.AStruct{3, 300, 129}
	MultiArgsResponseX = &end_to_end_test.AUnionB{Value: "TheUnion"}
	MultiArgsResponseY = "yresponse"

	SuccessMessage = "ALL TESTS PASSED"
	FailureMessage = "Failed"
)
