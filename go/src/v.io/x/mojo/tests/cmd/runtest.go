// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"v.io/x/mojo/tests/expected"
	"v.io/x/mojo/tests/util"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	v23proxy, err := util.StartV23ServerProxy(wd)
	if err != nil {
		panic(err)
	}
	endpoint, err := v23proxy.Endpoint()
	if err != nil {
		panic(err)
	}
	name := endpoint + "//https://mojo.v.io/test_server.mojo/mojo::v23proxy::tests::V23ProxyTest"
	if err := runTestClient(wd, name); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(2)
	}
	if err := v23proxy.Stop(); err != nil {
		panic(err)
	}
}

func runTestClient(v23ProxyRoot, endpoint string) error {
	cmd := util.RunMojoShellForV23ProxyTests("test_client.mojo", v23ProxyRoot, []string{endpoint})
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stdoutBuf bytes.Buffer
	go func() {
		io.Copy(os.Stdout, io.TeeReader(stdout, &stdoutBuf))
	}()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	if err := cmd.Run(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(stdoutBuf.Bytes()))
	for scanner.Scan() {
		if strings.HasSuffix(scanner.Text(), expected.SuccessMessage) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return fmt.Errorf("TESTS FAILED")
}
