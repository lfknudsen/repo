//go:build darwin
// +build darwin

package main

import "os/exec"

func open(file string) error {
	return exec.Command("open", file).Run()
}
