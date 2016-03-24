// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"v.io/x/mojo/tests/expected"
	"v.io/x/mojo/tests/util"
)

var runBench *bool = flag.Bool("bench", false, "run benchmarks instead of tests")
var clientType *string = flag.String("client", "go", "run test with a different client")
var serverType *string = flag.String("server", "go", "run test with a different server")

var (
	// Maps the client type to client mojo files.
	clientMap = map[string]string{
		"go":   "test_client.mojo",
		"dart": "dart-tests/end_to_end_test/lib/client.dart",
	}

	// Maps the server type to server mojo files.
	serverMap = map[string]string{
		"go":   "test_server.mojo",
		"dart": "dart-tests/end_to_end_test/lib/server.dart",
	}
)

func main() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("DEBUG LOG: Starting server proxy...")
	v23proxy, err := util.StartV23ServerProxy(wd)
	if err != nil {
		panic(err)
	}

	fmt.Println("DEBUG LOG: Waiting for v23 endpoint...")
	timeout := make(chan bool, 1)
	endpoint_chan := make(chan string, 1)
	go func() {
		time.Sleep(15 * time.Second)
		timeout <- true
	}()
	go func() {
		endpoint, err := v23proxy.Endpoint()
		if err != nil {
			panic(err)
		}
		endpoint_chan <- endpoint
	}()

	var endpoint string
	select {
	case endpoint = <-endpoint_chan:
		// keep going!
		break
	case <-timeout:
		panic("Failed to read endpoint!")
	}
	fmt.Println("DEBUG LOG: Starting client proxy...")

	endpointFlag := fmt.Sprintf("-endpoint=%s/https://mojo.v.io/%s/mojo::v23proxy::tests::V23ProxyTest", endpoint, serverMap[*serverType])
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
	cmd := util.RunMojoShellForV23ProxyTests(clientMap[*clientType], v23ProxyRoot, args)

	// A lock is put in home for the url response cache. Change HOME for v23proxy, since
	// two mojo shells will be run.
	tempHome, err := ioutil.TempDir("", "")
	defer os.Remove(tempHome)
	if err != nil {
		return err
	}
	cmd.Env = append(cmd.Env, "HOME="+tempHome)

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
