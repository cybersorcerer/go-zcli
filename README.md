# zcli - Command Line Interface to z/OS

zcli is a command line interface (CLI) to IBM z/OS REST services, allowing you to interact with z/OS from your local machine. It communicates with z/OSMF (z/OS Management Facility) REST APIs to manage jobs, datasets, files, filesystems, software instances, and more.

## Features

- **Jobs** - List, submit, cancel, hold, release, change class; browse spool files in a TUI
- **Datasets** - List, read, create, delete, rename; utilities (hrecall, hmigrate, hdelete)
- **Files** - List, retrieve, write, create, delete z/OS UNIX files; chmod, chown, chtag, extattr
- **Filesystems** - Create, delete, mount, unmount, list z/OS UNIX filesystems (zFS)
- **MFS** - List mounted filesystems with optional TUI browser
- **Software** - Query SMP/E CSI, list software instances, check critical/software/fixcat updates, export, dataset listing
- **Console** - Issue z/OS console commands with full support for solicited/unsolicited message detection, async mode, custom console names, and console attributes (auth, routcode, mscope, storage, auto); retrieve delayed responses and detection results
- **TSO** - Issue TSO/E commands
- **Topology** - List groups, sysplexes, systems; validate system/plex connectivity
- **Notifications** - List and send z/OSMF notifications
- **Subsystems** - List MVS subsystems with optional TUI
- **Sysvar** - Get, create, delete, import, and export z/OSMF system variables with optional TUI (including inline create and delete actions)
- **RTD** - Retrieve runtime diagnostic data
- **Info** - Retrieve z/OSMF server information
- **Profile** - Display zcli connection profiles

## Requirements

- Go 1.22 or later (for building from source)
- Access to a z/OSMF instance on z/OS 2.4 or later

## Installation

### From source

```bash
git clone <repository-url>
cd zcli
make build-local
make install
```

This builds the binary for your platform and installs it to `$HOME/bin`.

### Cross-platform build

```bash
make build
```

Produces binaries in `bin/` for:

- macOS (amd64, arm64)
- Linux (amd64, s390x)
- Windows (amd64)

### GitHub Releases

When a version tag (e.g. `v0.4.0`) is pushed, GitHub Actions automatically builds all platform binaries and creates a release with the artifacts attached.

```bash
git tag v0.4.0
git push origin v0.4.0
```

## Configuration

zcli reads its configuration from `~/.config/zcli/zcli.json`. The configuration file contains z/OSMF connection profiles and default settings.

### Example configuration

```json
{
  "profiles": {
    "my_zosmf": {
      "type": "zosmf",
      "properties": {
        "host": "zosmf.example.com",
        "port": "443",
        "protocol": "https",
        "user": "MYUSER",
        "password": "mypassword",
        "encoding": "IBM-1047",
        "rejectUnauthorized": false
      }
    }
  },
  "defaults": [
    {
      "profiles": [
        { "zosmf": "my_zosmf" }
      ]
    }
  ]
}
```

## Usage

```text
zcli [command] [subcommand] [flags]
```

### Global flags

| Flag               | Short | Description                                     |
| ------------------ | ----- | ----------------------------------------------- |
| `--profile-name`   |       | z/OSMF profile to use (default: from config)    |
| `--verify`         |       | TLS certificate verification (default: false)   |
| `--debug`          | `-d`  | Enable debug logging                            |
| `--format`         |       | Output format: json or text (default: json)     |

### Examples

```bash
# Get z/OSMF server information
zcli info

# List jobs with TUI
zcli jobs ls --tui --prefix 'TEST*' --status active

# List jobs as JSON
zcli jobs ls --owner MYUSER --prefix 'BATCH*'

# Submit a job
zcli jobs submit --file-name /path/to/job.jcl

# List datasets
zcli datasets list --dsn-level 'MYUSER.**'

# Read a dataset member
zcli datasets read --ds-name 'MYUSER.JCL' --member-name TESTJOB

# List z/OS UNIX files
zcli files list --path-name /u/myuser

# Retrieve a z/OS UNIX file
zcli files retrieve --zunix-file-name /u/myuser/hello.c

# List mounted filesystems (TUI)
zcli mfs --tui

# Issue a console command
zcli console command -c 'd a,l' --text

# Issue a console command with keyword detection
zcli console command -c 'd a,PEGASUS' --sol-key PEGASUS --text

# Issue a console command asynchronously
zcli console command -c 's PEGASUS' --async --unsol-key PEGASUS

# Issue a console command with custom console and auth
zcli console command -c 'd a,l' -n MYCONSOL --auth MASTER --routcode ALL

# Retrieve a delayed console response
zcli console get-response -k C6557643 --text

# Retrieve unsolicited keyword detection result
zcli console get-detection -k dec6800

# Issue a TSO command
zcli tso command --command 'LISTCAT'

# List MVS subsystems
zcli subsystems list

# List MVS subsystems with TUI
zcli subsystems list --tui

# Filter subsystems by ID
zcli subsystems list --ssid 'JES*'

# Get system variables
zcli sysvar get -x MAISEC -y MAIN

# Get system variables with TUI (supports create and delete)
zcli sysvar get -x MAISEC -y MAIN --tui

# Create/update system variables
zcli sysvar create -x MAISEC -y MAIN --var MYVAR=value1 --desc "My variable"

# Delete system variables
zcli sysvar delete -x MAISEC -y MAIN --var MYVAR

# Export system variables to z/OS file
zcli sysvar export -x MAISEC -y MAIN --file /u/myuser/vars.csv -w

# Export and download locally
zcli sysvar export -x MAISEC -y MAIN --file /u/myuser/vars.csv -w --local ./vars.csv

# Import system variables from z/OS file
zcli sysvar import -x MAISEC -y MAIN --file /u/myuser/vars.csv

# Upload local file and import
zcli sysvar import -x MAISEC -y MAIN --file /u/myuser/vars.csv --local ./vars.csv

# Query SMP/E CSI
zcli software query csids \
  --global-csi MY.GLOBAL.CSI \
  --entry SYSMOD \
  --subentries APPLY,ERROR \
  --filter "ename='HBB77E0'"

# List software instances
zcli software instances list

# List datasets of a software instance
zcli software instances silistds --uuid <uuid>

# List topology groups
zcli topology groups

# Validate system connectivity
zcli topology validate system --name MYSYS

# Use a specific profile
zcli --profile-name my_other_zosmf info
```

### TUI keyboard shortcuts

The interactive TUI (Terminal User Interface) is available for `jobs ls`, `mfs`, `subsystems list`, and `sysvar get` (use `--tui` flag).

**Jobs TUI:**

| Key                 | Action                           |
| ------------------- | -------------------------------- |
| `Ctrl+S` / `Enter`  | Select job files / spool content |
| `Ctrl+P` / `Esc`    | Go back to previous view         |
| `Ctrl+R`            | Refresh job list                 |
| `Ctrl+L`            | Cancel selected job              |
| `Ctrl+H`            | Hold selected job                |
| `Ctrl+E`            | Release selected job             |
| `Ctrl+C`            | Quit                             |
| `F7` / `F8`         | Page up / down                   |
| `G` / `g`           | Jump to bottom / top             |

**MFS TUI:**

| Key                 | Action                   |
| ------------------- | ------------------------ |
| `Ctrl+S` / `Enter`  | View filesystem details  |
| `Ctrl+P` / `Esc`    | Go back to list          |
| `Ctrl+R`            | Refresh                  |
| `Ctrl+C`            | Quit                     |

**Subsystems TUI:**

| Key        | Action   |
| ---------- | -------- |
| `Ctrl+R`   | Refresh  |
| `Ctrl+C`   | Quit     |
| `F7` / `F8` | Page up / down |

**Sysvar TUI:**

| Key        | Action                                      |
| ---------- | ------------------------------------------- |
| `Ctrl+N`   | Create / update a variable                  |
| `Ctrl+X`   | Delete selected variable (with confirmation)|
| `Ctrl+R`   | Refresh                                     |
| `Ctrl+C`   | Quit                                        |
| `Tab` / `Esc` | Navigate form / cancel                   |
| `F7` / `F8` | Page up / down                             |

## Console command flags

The `console command` subcommand supports the full z/OSMF REST Console API:

| Flag                    | Description |
| ----------------------- | ----------- |
| `--command/-c`          | z/OS command to issue (required) |
| `--console-name/-n`     | EMCS console name, 2-8 chars (default: `defcn`) |
| `--async/-a`            | Issue command asynchronously |
| `--system/-s`           | Target system in the sysplex |
| `--sol-key`             | Keyword to detect in command response |
| `--sol-key-regex`       | Treat sol-key as a regular expression |
| `--unsol-key`           | Keyword to detect in unsolicited messages |
| `--unsol-key-regex`     | Treat unsol-key as a regular expression |
| `--detect-time`         | Seconds to detect unsol-key (server default: 30) |
| `--unsol-detect-sync`   | Detect unsol-key synchronously |
| `--unsol-detect-timeout` | Timeout for synchronous detection (server default: 20) |
| `--auth`                | Console authority: MASTER, ALL, INFO, CONS, IO, SYS |
| `--routcode`            | Routing codes: ALL, NONE, or list |
| `--mscope`              | Message scope: ALL, LOCAL, or system names |
| `--storage`             | Storage in KB for message queuing (1-2000) |
| `--auto`                | Automation: YES or NO |
| `--text`                | Format response as readable text |

## Build

```bash
make build-local    # Build for current platform
make build          # Cross-platform build
make install        # Install to $HOME/bin
make clean          # Remove build artifacts
make dep            # Download and tidy dependencies
make vet            # Run go vet
make lint           # Run golangci-lint
make test           # Run tests
```

## License

Copyright (c) 2025, 2026 Sir Tobi aka Cybersorcerer

See [LICENSE](LICENSE) for details.
