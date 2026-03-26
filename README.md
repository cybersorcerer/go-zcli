# zcli - Command Line Interface to z/OS

zcli is a command line interface (CLI) to IBM z/OS REST services, allowing you to interact with z/OS from your local machine. It communicates with z/OSMF (z/OS Management Facility) REST APIs to manage jobs, datasets, files, filesystems, software instances, and more.

## Features

- **Jobs** - List, submit, cancel, hold, release, change class; browse spool files in a TUI
- **Datasets** - List, read, create, delete, rename; utilities (hrecall, hmigrate, hdelete)
- **Files** - List, retrieve, write, create, delete z/OS UNIX files; chmod, chown, chtag, extattr
- **Filesystems** - Create, delete, mount, unmount, list z/OS UNIX filesystems (zFS)
- **MFS** - List mounted filesystems with optional TUI browser
- **Software** - Query SMP/E CSI, list software instances, check critical/software/fixcat updates, export, dataset listing
- **Console** - Issue z/OS console commands
- **TSO** - Issue TSO/E commands
- **Topology** - List groups, sysplexes, systems; validate system/plex connectivity
- **Notifications** - List and send z/OSMF notifications
- **Subsystems** - List z/OS subsystems
- **Sysvar** - Retrieve z/OS system variables
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
zcli console command --command 'd a,l'

# Issue a TSO command
zcli tso command --command 'LISTCAT'

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

The interactive TUI (Terminal User Interface) is available for `jobs ls --tui` and `mfs --tui`.

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
