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
	"flag"

	"v.io/x/mojo/tests/expected"
	"v.io/x/mojo/tests/util"
)

var runBench *bool = flag.Bool("bench", false, "run benchmarks instead of tests")

func main() {
	flag.Parse()
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
	endpointFlag := "-endpoint=" + endpoint + "//https://mojo.v.io/test_server.mojo/mojo::v23proxy::tests::V23ProxyTest"
	args := []string{endpointFlag, "--v23.tcp.address=127.0.0.1:0"}
	if *runBench {
		args = append(args, "-test.run=XXXX", "-test.bench=.")
	}
	if err := runTestClient(wd, args...); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(2)
	}
	if err := v23proxy.Stop(); err != nil {
		panic(err)
	}
}

func runTestClient(v23ProxyRoot string, args ...string) error {
	cmd := util.RunMojoShellForV23ProxyTests("test_client.mojo", v23ProxyRoot, args)
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
