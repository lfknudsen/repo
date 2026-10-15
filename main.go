package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Base(filepath.Dir(dir))
}

// createGitLab returns a command  to create a new remote repository hosted on gitlab.com
var createGitLab = exec.Command("glab", "repo", "create", "--remote", "origin", "--private")

// createCodeberg returns a command  to create a new remote repository hosted on Codeberg.org
var createCodeberg = exec.Command("fj", "repo", "create", "--private", "--remote", "backup1", "--host", "codeberg.org", currentDir())

// createGitHub returns a command  to create a new remote repository hosted on GitHub.com
var createGitHub = exec.Command("gh", "repo", "create", currentDir(), "--private", "--remote", "backup2")

// authGitLab returns the command to authenticate the GitLab CLI.
var authGitLab = exec.Command("glab", "auth", "login")

// authCodeberg returns the command to authenticate the Forgejo CLI.
var authCodeberg = exec.Command("fj", "auth", "login", "--host", "codeberg.org")

// authGitHub returns the command to authenticate the GitHub CLI.
var authGitHub = exec.Command("gh", "auth", "login")

func execute(cmd *exec.Cmd) {
	if _, err := os.Stat(cmd.Path); err != nil {
		return
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func auth() {
	execute(authGitLab)
	execute(authCodeberg)
	execute(authGitHub)
}

func create() {
	execute(createGitLab)
	execute(createCodeberg)
	execute(createGitHub)
}

const version = "1.0.0"

func printHelp() {
	fmt.Printf("REPO v%s\nRepository Creation Utility\n\n"+
		"Usage:\n"+
		"  repo create\n"+
		"    Creates a new repository on GitLab, Codeberg, and GitHub in that order.\n"+
		"  repo auth\n"+
		"    Authenticates the CLI utilities to allow the management of repositories.\n"+
		"    This will require you to log in to GitLab, Codeberg, and GitHub in that order.\n"+
		"    If you've never used the hosts' CLI programmes before, you will have to do this\n"+
		"    first. This only needs to be done once.\n"+
		"  repo [-h|--help]\n"+
		"    Print this help message and exit.\n", version)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	if args[0] == "auth" {
		auth()
	}

	_, err := os.Stat("./.git")
	if err != nil {
		fmt.Println("The current directory is not a git repository. Please run `git init` first.")
		os.Exit(1)
	}

	if args[0] == "-h" || args[0] == "--help" {
		printHelp()
		os.Exit(0)
	}

	if args[0] == "create" {
		create()
	}

}
