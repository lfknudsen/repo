# Repo

Utility for conveniently creating repositories with multiple online hosts.
Currently, the supported hosts are:\
[Codeberg](https://codeberg.org)\
[GitLab](https://gitlab.com)\
[GitHub](https://www.github.com)

## Dependencies
Each host has their own CLI utility that you will need to download and add to
your PATH:\
[Forgejo CLI (Codeberg)](https://codeberg.org/forgejo-contrib/forgejo-cli)\
[GitLab CLI](https://gitlab.com/gitlab-org/cli)\
[GitHub CLI](https://github.com/cli/cli)

## Compile

```
go build
```

## Setup

### Authentication
If you haven't used the required CLI tools before, you will have to authenticate
them with their respective host. To do this, execute
```
repo auth [-d|--dry] [--only <host name>] [--skip <host name>]
```
Each app will take you through its process, but it should be very quick.

If `-d` or `--dry`, the programme will print the commands that would have been
executed if omitted, and then exit without actually executing them.\
If `--skip <host name>`, the programme will skip whichever host is named.
To skip multiple, include this argument multiple times.\
If `--only <host name>`, the programme will exclusively use the specified
host. Can only be used once.

### Configuration
The first time the programme is run, it will generate a configuration file,
`.repo.csv` in the same location as the executable, which contains a list of
host names and remote names in the format
```
<host name>,<remote name>
```

You can execute the following to reinitialise this file at any time:
```
repo init
```

To open the configuration file in your computer's default application, execute
```
repo config
```

To print the path to the configuration file, execute
```
repo config path
```

To print the list of hosts and their remote names, execute
```
repo hosts
```

## Usage

If you include the flag `-h` or `--help`, usage instructions will print to
standard output.

To create new repositories:
```
repo create [-d|--dry] [--only <host name>] [--skip <host name>]
```
If `-d` or `--dry`, the programme will print the commands that would have been
executed if omitted, and then exit without actually executing them.\
If `--skip <host name>`, the programme will skip whichever host is named.
To skip multiple, include this argument multiple times.\
If `--only <host name>`, the programme will exclusively use the specified
host. Can only be used once, as it simply removes all other hosts from the
pool.

### Other commands

To update repository description:
```
repo describe [OPTIONS] <description>
```

Options are the same as for `repo create` and `repo auth`.\
Aliases: `desc` and `description`

To update repository visibility:
```
repo vis[ibility] [OPTIONS] public|private|internal
```
Options are the same as for `repo create` and `repo auth`.\
NOTE: Only GitLab and GitHub support the 'internal' setting.\
If this is chosen, Codeberg will be skipped.
