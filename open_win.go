//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func open(file string) error {
	runDLL32 := filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
	arg1 := "url.dll,FileProtocolHandler"
	return exec.Command(runDLL32, arg1, file).Run()
}
