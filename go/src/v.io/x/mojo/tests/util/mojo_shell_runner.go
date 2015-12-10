// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"v.io/jiri/jiri"
	"v.io/jiri/profiles"
	"v.io/x/lib/cmdline"
	"v.io/x/lib/envvar"
	"v.io/x/lib/timing"
)

func RunMojoShell(mojoUrl, configFile string, configAliases map[string]string, argsFor map[string][]string, target profiles.Target) *exec.Cmd {
	// ensure the profiles are loaded
	jirix, err := jiri.NewX(&cmdline.Env{Timer: timing.NewTimer("root")})
	if err != nil {
		panic(err)
	}
	_, err = profiles.NewConfigHelper(jirix, profiles.UseProfiles, filepath.Join(jirix.Root, ".jiri_v23_profiles"))
	if err != nil {
		panic(err)
	}

	envslice := profiles.EnvFromProfile(target, mojoProfileName())
	env := envvar.VarsFromSlice(envslice)
	jiri.ExpandEnv(jirix, env)
	var mojoDevtools, mojoShell, mojoServices string
	for _, e := range env.ToSlice() {
		parts := strings.SplitN(e, "=", 2)
		switch parts[0] {
		case "MOJO_DEVTOOLS":
			mojoDevtools = parts[1]
		case "MOJO_SHELL":
			mojoShell = parts[1]
		case "MOJO_SERVICES":
			mojoServices = parts[1]
		}
	}
	args := []string{
		mojoUrl,
		"--config-file", configFile,
		"--shell-path", mojoShell,
		"--enable-multiprocess"}
	if target.OS() == "android" {
		args = append(args, "--android")
		args = append(args, "--origin", mojoServices)
	}
	for alias, value := range configAliases {
		args = append(args, "--config-alias", fmt.Sprintf("%s=%s", alias, value))
	}
	for key, value := range argsFor {
		args = append(args, fmt.Sprintf("--args-for=%s %s", key, strings.Join(value, " ")))
	}
	return exec.Command(filepath.Join(mojoDevtools, "mojo_run"), args...)
}

func RunMojoShellForV23ProxyTests(mojoFile, v23ProxyRoot string, args []string) *exec.Cmd {
	configFile := filepath.Join(v23ProxyRoot, "mojoconfig")
	mojoUrl := fmt.Sprintf("https://mojo.v.io/%s", mojoFile)
	buildDir := filepath.Join(v23ProxyRoot, "gen", "mojo", "linux_amd64")
	configAliases := map[string]string{
		"V23PROXY_DIR":       v23ProxyRoot,
		"V23PROXY_BUILD_DIR": buildDir,
	}
	argsFor := map[string][]string{
		mojoUrl:                     args,
		"mojo:dart_content_handler": []string{"--enable-strict-mode"},
	}
	target := profiles.DefaultTarget()
	return RunMojoShell(mojoUrl, configFile, configAliases, argsFor, target)
}

func mojoProfileName() string {
	if val, ok := os.LookupEnv("USE_MOJO_DEV_PROFILE"); ok && val == "true" {
		fmt.Printf("Using dev profile\n")
		return "mojo-dev"
	}
	return "mojo"
}
