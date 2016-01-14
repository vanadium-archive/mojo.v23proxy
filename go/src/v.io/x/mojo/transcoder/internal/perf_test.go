// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"mojo/public/go/bindings"
	"mojom/tests/transcoder_testcases"
	"testing"
	"v.io/v23/vdl"
	"v.io/v23/vom"
	"v.io/x/mojo/transcoder"
)

var customer transcoder_testcases.Customer = transcoder_testcases.Customer{
	Name:   "John Smith",
	Id:     1,
	Active: true,
	Address: transcoder_testcases.AddressInfo{
		Street: "1 Main St.",
		City:   "Palo Alto",
		State:  "CA",
		Zip:    "94303",
	},
	Credit: transcoder_testcases.CreditReport{
		Agency: transcoder_testcases.CreditAgency_Equifax,
		Report: &transcoder_testcases.AgencyReportEquifaxReport{transcoder_testcases.EquifaxCreditReport{'A'}},
	},
}

var vdlCustomer Customer = Customer{
	Name:   "John Smith",
	Id:     1,
	Active: true,
	Address: AddressInfo{
		Street: "1 Main St.",
		City:   "Palo Alto",
		State:  "CA",
		Zip:    "94303",
	},
	Credit: CreditReport{
		Agency: CreditAgencyEquifax,
		Report: AgencyReportEquifaxReport{EquifaxCreditReport{'A'}},
	},
}

func BenchmarkVdlToMojomTranscoding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transcoder.VdlToMojom(customer)
	}
}

func BenchmarkMojomToVdlTranscoding(b *testing.B) {
	data := mojomBytesCustomer()
	t := vdl.TypeOf(customer)
	for i := 0; i < b.N; i++ {
		var c Customer
		transcoder.MojomToVdl(data, t, &c)
	}
}

func BenchmarkVomEncoding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vom.Encode(vdlCustomer)
	}
}

func BenchmarkVomDecoding(b *testing.B) {
	data := vomBytesCustomer()
	for i := 0; i < b.N; i++ {
		var c Customer
		vom.Decode(data, &c)
	}
}

func BenchmarkMojoEncoding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := bindings.NewEncoder()
		customer.Encode(enc)
		enc.Data()
	}
}

func BenchmarkMojoDecoding(b *testing.B) {
	data := mojomBytesCustomer()
	for i := 0; i < b.N; i++ {
		dec := bindings.NewDecoder(data, nil)
		customer.Decode(dec)
	}
}

func mojomBytesCustomer() []byte {
	enc := bindings.NewEncoder()
	err := customer.Encode(enc)
	if err != nil {
		panic(err)
	}
	data, _, err := enc.Data()
	if err != nil {
		panic(err)
	}
	return data
}

func vomBytesCustomer() []byte {
	data, err := vom.Encode(vdlCustomer)
	if err != nil {
		panic(err)
	}
	return data
}
