package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Base(filepath.Dir(dir))
}

func getConfigFilePath() string {
	path, err := os.Executable()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Join(filepath.Dir(path), ".repo.csv")
}

var rxHost = regexp.MustCompile(`(?:https?://)?(?:www\.)?([^.]+)[^,]*,(.+)`)

const defaultHostFileContent = "HOST NAME,REMOTE NAME\ngitlab,origin\ncodeberg,backup1\ngithub,backup2"

func getHosts() (hosts []string, remotes []string) {
	configFilename := getConfigFilePath()
	// Start by creating and initialising the config file if it doesn't exist already.
	f, err := os.OpenFile(configFilename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		wr := bufio.NewWriter(f)
		_, _ = wr.WriteString(defaultHostFileContent)
		_ = wr.Flush()
		_ = f.Close()
	}

	f, err = os.OpenFile(configFilename, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
		return []string{}, []string{}
	}
	r := bufio.NewScanner(f)
	var str string
	hosts = make([]string, 0, 3)
	remotes = make([]string, 0, 3)
	remoteName := "origin"
	remoteNameBackupNumber := 1
	for r.Scan() {
		str = r.Text()
		if str == "HOST NAME,REMOTE NAME" {
			continue
		}
		match := rxHost.FindStringSubmatch(str)
		if len(match) < 2 {
			fmt.Println(match)
			os.Exit(1)
		}
		hosts = append(hosts, strings.ToLower(match[1]))
		if len(match) > 2 {
			remotes = append(remotes, match[2])
		} else {
			remotes = append(remotes, remoteName)
			remoteName = fmt.Sprintf("backup%d", remoteNameBackupNumber)
			remoteNameBackupNumber++
		}
	}
	_ = f.Close()
	return hosts, remotes
}

var hosts, remotes []string = getHosts()

func getRemoteName(host string) string {
	if idx := slices.Index(hosts, host); idx >= 0 {
		return remotes[idx]
	}
	return ""
}

// createGitLab returns a command  to create a new remote repository hosted on gitlab.com
var createGitLab = exec.Command("glab", "repo", "create", "--private")

// createCodeberg returns a command  to create a new remote repository hosted on Codeberg.org
var createCodeberg = exec.Command("fj", "repo", "create", "--private", "--host", "codeberg.org", currentDir())

// createGitHub returns a command  to create a new remote repository hosted on GitHub.com
var createGitHub = exec.Command("gh", "repo", "create", currentDir(), "--private")

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

func create() {
	if slices.Contains(hosts, "gitlab") {
		cmd := createGitLab
		cmd.Args = append(cmd.Args, "--remote", getRemoteName("gitlab"))
		execute(cmd)
	}
	if slices.Contains(hosts, "codeberg") {
		cmd := createCodeberg
		cmd.Args = append(cmd.Args, "--remote", getRemoteName("codeberg"))
		execute(cmd)
	}
	if slices.Contains(hosts, "github") {
		cmd := createGitHub
		cmd.Args = append(cmd.Args, "--remote", getRemoteName("github"))
		execute(cmd)
	}
}

func auth() {
	if slices.Contains(hosts, "gitlab") {
		execute(authGitLab)
	}
	if slices.Contains(hosts, "codeberg") {
		execute(authCodeberg)
	}
	if slices.Contains(hosts, "github") {
		execute(authGitHub)
	}
}

const version = "1.0.0"

func printHelp() {
	fmt.Printf("REPO v%s\nRepository Creation Utility\n\n"+
		"Usage:\n"+
		"  repo create\n"+
		"    Creates a new repository on GitLab, Codeberg, and GitHub in that order.\n"+
		"    The URL will be https://<host>/<your username>/<this directory>\n"+
		"  repo auth\n"+
		"    Authenticates the CLI utilities to allow the management of repositories.\n"+
		"    This will require you to log in to GitLab, Codeberg, and GitHub in that order.\n"+
		"    If you've never used the hosts' CLI programmes before, you will have to do this\n"+
		"    first. This only needs to be done once.\n"+
		"  repo [-h|--help]\n"+
		"    Print this help message and exit.\n",
		version)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	if len(hosts) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No hosts specified in %s.\n", getConfigFilePath())
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
