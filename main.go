package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// dry indicates whether this should be a dry-run.
// If true, the execute function will only print the command that it would have otherwise run.
var dry = false

var visibility = "--private"

// arg returns true if the current command-line argument is equal to one of the given input params.
func arg(params ...string) bool {
	return slices.Contains(params, os.Args[0])
}

// next skips past the current command-line argument.
func next() {
	os.Args = os.Args[1:]
}

// checkFlags handles the boolean command-line argument flags.
func checkFlags() {
	var idx int
	idx = slices.Index(os.Args, "-d")
	if idx != -1 {
		dry = true
		slices.Delete(os.Args, idx, idx+1)
	}
	idx = slices.Index(os.Args, "--dry")
	if idx != -1 {
		dry = true
		slices.Delete(os.Args, idx, idx+1)
	}
	idx = slices.Index(os.Args, "-h")
	if idx != -1 {
		printHelp()
		os.Exit(0)
	}
	idx = slices.Index(os.Args, "--help")
	if idx != -1 {
		printHelp()
		os.Exit(0)
	}
	idx = slices.Index(os.Args, "--public")
	if idx != -1 {
		visibility = "--public"
	}
}

// getCurrentDir returns the name of the current directory.
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Base(dir)
}

// currentDir caches the result of getCurrentDir.
var currentDir = getCurrentDir()

// getConfigFilePath returns the absolute path to the file with the list of hosts and remotes.
func getConfigFilePath() string {
	path, err := os.Executable()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Join(filepath.Dir(path), ".repo.csv")
}

// configPath caches the result of getConfigFilePath.
var configPath = getConfigFilePath()

// rxHostRemote is used to parse and re-format the user-input hosts and remotes.
var rxHostRemote = regexp.MustCompile(`(?:https?://)?(?:www\.)?([^.]+)[^,]*,(.+)`)

// rxHost is used to parse and re-format a user-input host name.
var rxHost = regexp.MustCompile(`(?:https?://)?(?:www\.)?([^.]+)`)

// defaultHostFileContent contains the initial contents of the config file.
const defaultHostFileContent = "HOST NAME,REMOTE NAME\ngitlab,origin\ncodeberg,backup1\ngithub,backup2"

// reinitHostFile creates/truncates the config file and writes the default contents from
// defaultHostFileContent
func reinitHostFile() {
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err == nil {
		wr := bufio.NewWriter(f)
		_, _ = wr.WriteString(defaultHostFileContent)
		_ = wr.Flush()
		_ = f.Close()
	}
}

// initHostFile creates a new config file and writes the default contents from
// defaultHostFileContent.
//
// If the config file already exists, nothing happens and no error is thrown.
func initHostFile() {
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		wr := bufio.NewWriter(f)
		_, _ = wr.WriteString(defaultHostFileContent)
		_ = wr.Flush()
		_ = f.Close()
	}
}

// getHosts reads the contents of the config file and returns them as two arrays.
// The host name in position `i` has the remote name in position `i`. The arrays
// always have the same size.
func getHosts() (hosts []string, remotes []string) {
	// Start by creating and initialising the config file if it doesn't exist already.
	initHostFile()
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0666)
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
		match := rxHostRemote.FindStringSubmatch(str)
		if len(match) < 2 {
			_, _ = fmt.Fprintf(os.Stderr, "error in host file format: %s\n", str)
			os.Exit(1)
		}
		hosts = append(hosts, strings.ToLower(match[1]))
		if len(match) > 2 {
			remotes = append(remotes, match[2])
		} else {
			if slices.Contains(remotes, remoteName) {
				remoteName = fmt.Sprintf("backup%d", remoteNameBackupNumber)
				remoteNameBackupNumber++
			}
			remotes = append(remotes, remoteName)
			remoteName = fmt.Sprintf("backup%d", remoteNameBackupNumber)
			remoteNameBackupNumber++
		}
	}
	_ = f.Close()
	return hosts, remotes
}

// printHosts prints to standard output the list of hosts and their remotes as
// retrieved by getHosts.
func printHosts() {
	maxWidth := 0
	for _, host := range hosts {
		if len(host) > maxWidth {
			maxWidth = len(host)
		}
	}
	const headerHost = "HOST:"
	maxWidth += 5
	fmt.Print(headerHost)
	fmt.Print(strings.Repeat(" ", maxWidth-len(headerHost)))
	const headerRemote = "REMOTE:"
	fmt.Println(headerRemote)

	for i := range hosts {
		fmt.Print(hosts[i])
		fmt.Print(strings.Repeat(" ", maxWidth-len(hosts[i])))
		fmt.Println(remotes[i])
	}
}

// hosts and remotes cache the results of getHosts.
var hosts, remotes []string = getHosts()

// getRemoteName maps a host name to a remote name.
func getRemoteName(host string) string {
	if idx := slices.Index(hosts, host); idx >= 0 {
		return remotes[idx]
	}
	return ""
}

// createGitLab returns a command  to create a new remote repository hosted on gitlab.com
var createGitLab = exec.Command("glab", "repo", "create", visibility)

// createCodeberg returns a command  to create a new remote repository hosted on Codeberg.org
var createCodeberg = exec.Command("fj", "repo", "create", visibility, "--host", "codeberg.org", currentDir)

// createGitHub returns a command  to create a new remote repository hosted on GitHub.com
var createGitHub = exec.Command("gh", "repo", "create", currentDir, visibility)

// authGitLab returns the command to authenticate the GitLab CLI.
var authGitLab = exec.Command("glab", "auth", "login")

// authCodeberg returns the command to authenticate the Forgejo CLI.
var authCodeberg = exec.Command("fj", "auth", "login", "--host", "codeberg.org")

// authGitHub returns the command to authenticate the GitHub CLI.
var authGitHub = exec.Command("gh", "auth", "login")

var descriptionGitLab = exec.Command("glab", "repo", "update", "--description")
var descriptionCodeberg = exec.Command("fj", "repo", "edit", "--host", "codeberg.org", "--description")
var descriptionGitHub = exec.Command("gh", "repo", "edit", "-d")

// execute runs the given command.
// Stdout and Stderr are piped to the current console's.
//
// If the dry variable is true, it will print the command string
// and exit without doing anything else.
func execute(cmd *exec.Cmd) {
	if dry {
		fmt.Printf("Dry run. Would have executed:\n  %s\n", strings.Join(cmd.Args, " "))
		return
	}

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

// create will assemble the commands for each repo creation and execute them,
// skipping over any hosts that the user has omitted via `--skip <host name>`
func create() {
	if slices.Contains(hosts, "gitlab") {
		fmt.Println("Creating GitLab project...")
		cmd := createGitLab
		cmd.Args = append(cmd.Args, "--remoteName", getRemoteName("gitlab"))
		execute(cmd)
	}
	if slices.Contains(hosts, "codeberg") {
		fmt.Println("Creating Codeberg project...")
		cmd := createCodeberg
		cmd.Args = append(cmd.Args, "--remote", getRemoteName("codeberg"))
		execute(cmd)
	}
	if slices.Contains(hosts, "github") {
		fmt.Println("Creating GitHub project...")
		cmd := createGitHub
		cmd.Args = append(cmd.Args, "--remote", getRemoteName("github"), "--source", ".")
		execute(cmd)
	}
}

// auth assembles and executes the repo authentication commands.
func auth() {
	if slices.Contains(hosts, "gitlab") {
		fmt.Println("Requesting GitLab authorisation...")
		execute(authGitLab)
	}
	if slices.Contains(hosts, "codeberg") {
		fmt.Println("Requesting Codeberg authorisation...")
		execute(authCodeberg)
	}
	if slices.Contains(hosts, "github") {
		fmt.Println("Requesting GitHub authorisation...")
		execute(authGitHub)
	}
}

const version = "1.5.4"

func printHelp() {
	fmt.Printf("REPO v%s\nRepository Creation Utility\n\n"+
		"Usage:\n"+
		"  repo create [-d|--dry] [--only <host>|--skip <host>] [--public]\n"+
		"    Creates a new private repository on GitLab, Codeberg, and GitHub in that order.\n"+
		"    The URL will be https://<host>/<your username>/<this directory>\n\n"+
		"    If -d or --dry, the command that would have been executed is instead\n"+
		"      printed to the standard output, and the programme exits.\n"+
		"    If --skip is specified, the given host will be skipped.\n"+
		"      Multiple --skip flags will cause all the specified hosts to be skipped.\n"+
		"    If --only is specified, hosts other than the given will be skipped.\n"+
		"      Only one --only can be used. --only and --skip are mutually exclusive.\n"+
		"	 If --public, the repository will be public instead of private.\n"+
		"-------------------------------------------------------------------------------\n"+
		"  repo auth [-d|--dry] [--only <host>|--skip <host>]\n"+
		"    Authenticates the CLI utilities to allow the management of repositories.\n"+
		"    This will require you to log in to GitLab, Codeberg, and GitHub in that order.\n"+
		"    If you've never used the hosts' CLI programmes before, you will have to do this\n"+
		"    first. This only needs to be done once.\n\n"+
		"    If -d or --dry, the command that would have been executed is instead\n"+
		"      printed to the standard output, and the programme exits.\n"+
		"    If --skip is specified, the given host will be skipped.\n"+
		"      Multiple --skip flags will cause all the specified hosts to be skipped.\n"+
		"    If --only is specified, hosts other than the given will be skipped.\n"+
		"      Only one --only can be used. --only and --skip are mutually exclusive.\n"+
		"-------------------------------------------------------------------------------\n"+
		"  repo describe [OPTIONS] <description text>\n"+
		"    Adds/updates the given description text to the repository.\n"+
		"    --skip and --only work as they do with the other commands.\n"+
		"    Aliases: desc, description.\n"+
		"-------------------------------------------------------------------------------\n"+
		"  repo init\n"+
		"    Re-initialises the file with the list of hosts and their remotes.\n"+
		"    Prints the path to the file.\n"+
		"-------------------------------------------------------------------------------\n"+
		"  repo config\n"+
		"    Opens the configuration file in the default application.\n"+
		"-------------------------------------------------------------------------------\n"+
		"  repo config path\n"+
		"    Prints the path to the configuration file.\n"+
		"-------------------------------------------------------------------------------\n"+
		"  repo hosts\n"+
		"    Prints the list of hosts and their remote name.\n"+
		"-------------------------------------------------------------------------------\n"+
		"Universal flags:\n"+
		"    -h or --help: Prints this help message and exits.\n\n"+
		"A note on host names:\n"+
		"    You do not need to include the full URL;\n"+
		"    `https://www.gitlab.com/` is parsed as `gitlab`\n"+
		"    You should not name a host `-d`, `--dry`, `-h`, or `--help` and attempt to skip it.",
		version)
}

// Removes the given host name and its remote from the variables hosts and remotes,
// which causes create and auth to skip it.
func skipHost(host string) error {
	matches := rxHost.FindStringSubmatch(host)
	if len(matches) < 2 {
		return errors.New("could not parse host")
	}
	idx := slices.Index(hosts, matches[1])
	if idx == -1 {
		return errors.New("host not found")
	}
	slices.Delete(hosts, idx, idx+1)
	slices.Delete(remotes, idx, idx+1)
	return nil
}

func onlyHost(host string) error {
	matches := rxHost.FindStringSubmatch(host)
	if len(matches) < 2 {
		return errors.New("could not parse host")
	}
	idx := slices.Index(hosts, matches[1])
	if idx == -1 {
		return errors.New("host not found")
	}
	hosts = []string{hosts[idx]}
	remotes = []string{remotes[idx]}
	return nil
}

func main() {
	var err error
	next()
	if len(os.Args) == 0 {
		printHelp()
		os.Exit(1)
	}
	checkFlags()
	if arg("init") {
		reinitHostFile()
		fmt.Println(configPath)
		os.Exit(0)
	}

	if arg("config") {
		if len(os.Args) > 1 {
			if os.Args[1] == "path" {
				fmt.Println(configPath)
				os.Exit(0)
			}
			printHelp()
			os.Exit(1)
		}
		err = open(configPath)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(hosts) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No hosts specified in %s.\n", configPath)
		printHelp()
		os.Exit(1)
	}

	if arg("auth") {
		determineHosts()
		auth()
		os.Exit(0)
	}

	_, err = os.Stat("./.git")
	if err != nil {
		fmt.Println("The current directory is not a git repository. Please run `git init` first.")
		os.Exit(1)
	}

	if arg("create") {
		determineHosts()
		create()
		os.Exit(0)
	}

	if arg("hosts") {
		printHosts()
		os.Exit(0)
	}

	if arg("describe") || arg("desc") || arg("description") {
		determineHosts()
		if len(os.Args) > 1 {
			updateDescription(os.Args[1])
			os.Exit(0)
		}
		printHelp()
		os.Exit(1)
	}
}

func determineHosts() {
	next()
	var err error
	for len(os.Args) >= 2 {
		switch os.Args[0] {
		case "--only":
			err = onlyHost(os.Args[1])
		case "--skip":
			err = skipHost(os.Args[1])
		default:
			printHelp()
			os.Exit(1)
		}
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Args = os.Args[2:]
	}
}

func updateDescription(description string) {
	if slices.Contains(hosts, "gitlab") {
		fmt.Println("Updating GitLab project...")
		cmd := descriptionGitLab
		cmd.Args = append(cmd.Args, description)
		execute(cmd)
	}
	if slices.Contains(hosts, "codeberg") {
		fmt.Println("Updating Codeberg project...")
		cmd := descriptionCodeberg
		cmd.Args = append(cmd.Args, description)
		execute(cmd)
	}
	if slices.Contains(hosts, "github") {
		fmt.Println("Updating GitHub project...")
		cmd := descriptionGitHub
		cmd.Args = append(cmd.Args, description)
		execute(cmd)
	}
}
