//go:build !windows && !darwin
// +build !windows,!darwin

package main

import "os/exec"

func open(file string) error {
	return exec.Command("xdg-open", configPath).Run()
}
