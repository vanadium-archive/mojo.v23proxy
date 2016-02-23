// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

func StartV23ServerProxy(v23ProxyRoot string) (*V23ProxyController, error) {
	cmd := RunMojoShellForV23ProxyTests("v23serverproxy.mojo", v23ProxyRoot, []string{"--v23.tcp.address=127.0.0.1:0"})
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	// A lock is put in home for the url response cache. Change HOME for v23proxy, since
	// two mojo shells will be run.
	tempHome, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	cmd.Env = append(cmd.Env, "HOME="+tempHome)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &V23ProxyController{
		cmd:      cmd,
		stdout:   stdout,
		tempHome: tempHome,
	}, nil
}

type V23ProxyController struct {
	cmd          *exec.Cmd
	stdout       io.ReadCloser
	endpointLock sync.Mutex
	endpoint     string // empty until the endpoint is read
	tempHome     string
}

func (v *V23ProxyController) Stop() error {
	os.Remove(v.tempHome)
	childPids, err := getChildProcessPids(v.cmd.Process.Pid)
	if err != nil {
		return err
	}
	if err := v.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	for _, pid := range childPids {
		syscall.Kill(pid, syscall.SIGTERM)
	}
	return nil
}

func (v *V23ProxyController) Endpoint() (string, error) {
	v.endpointLock.Lock()
	defer v.endpointLock.Unlock()

	if v.endpoint != "" {
		return v.endpoint, nil
	}

	scanner := bufio.NewScanner(v.stdout)
	const prefix = "Listening at: "
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), prefix) {
			v.endpoint = strings.TrimPrefix(scanner.Text(), prefix)
			return v.endpoint, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("unexpected EOF when looking for endpoint")
}

func getChildProcessPids(parentPid int) ([]int, error) {
	cmd := exec.Command("ps", []string{"h", "--ppid", fmt.Sprintf("%d", parentPid), "-o", "pid"}...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdout)
	var pids []int
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		pid, err := strconv.ParseInt(strings.Trim(scanner.Text(), " "), 10, 64)
		if err != nil {
			return nil, err
		}
		pids = append(pids, int(pid))
	}
	return pids, cmd.Wait()
}
