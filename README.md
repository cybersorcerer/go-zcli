# zcli - Command Line Interface to z/OS

zcli is a command line interface (CLI) to IBM z/OS REST services, allowing you to interact with z/OS from your local machine. It communicates with z/OSMF (z/OS Management Facility) REST APIs to manage jobs, datasets, files, filesystems, software instances, and more.

## Features

- **Jobs** - List, submit, cancel, hold, release, change class; browse spool files in a TUI
- **Datasets** - List, read, create, delete; list members; utilities (copy, rename, hrecall, hmigrate, hdelete, AMS)
- **Files** - List, retrieve, write, create, delete z/OS UNIX files; chmod, chown, chtag, extattr
- **Filesystems** - Create, delete, mount, unmount, list z/OS UNIX filesystems (zFS)
- **MFS** - List mounted filesystems with optional TUI browser
- **Software** - Query SMP/E CSI, list software instances, check critical/software/fixcat updates, export, dataset listing
- **Console** - Issue z/OS console commands with full support for solicited/unsolicited message detection, async mode, custom console names, and console attributes (auth, routcode, mscope, storage, auto); retrieve delayed responses, detection results, and hardcopy logs
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

### Build on z/OS

Cross-compilation for z/OS is not supported. Build natively on z/OS:

```bash
go build -ldflags '-X main.version=v0.4.0 -X main.commit=$(git rev-parse --short HEAD)' -o zcli .
```

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

# List datasets with volume filter
zcli datasets list --dsn-level 'SYS1.**' --volser SYSRES

# List datasets with custom attributes and max items
zcli datasets list --dsn-level 'MYUSER.**' --attributes base --max-items 50

# List members of a PDS
zcli datasets members --ds-name 'SYS1.PROCLIB'

# List members with base attributes
zcli datasets members --ds-name 'SYS1.PROCLIB' --attributes base,total

# List members with pattern filter
zcli datasets members --ds-name 'SYS1.PROCLIB' --pattern 'IEF*' --max-items 20

# Read a dataset member
zcli datasets read --ds-name 'MYUSER.JCL' --member-name TESTJOB

# Read as binary
zcli datasets read --ds-name 'MYUSER.LOAD' --member-name MYPROG --data-type binary

# Read with record range (first 10 records)
zcli datasets read --ds-name 'MYUSER.DATA' --record-range '0-9'

# Search for a string in a dataset
zcli datasets read --ds-name 'MYUSER.JCL' --member-name TESTJOB --search 'EXEC PGM'

# Search with regex
zcli datasets read --ds-name 'MYUSER.JCL' --member-name TESTJOB --research 'EXEC +PGM' --insensitive false

# Read an uncataloged dataset
zcli datasets read --ds-name 'OLD.DATA' --volser VOL001

# Read with ENQ lock for editing
zcli datasets read --ds-name 'MYUSER.DATA' --obtain-enq SHRW

# Create a PDS
zcli datasets create --ds-name 'MYUSER.NEW.PDS' --dsorg PO --recfm FB --lrecl 80

# Create a sequential dataset
zcli datasets create --ds-name 'MYUSER.NEW.SEQ' --dsorg PS --primary 20 --secondary 10

# Delete a dataset
zcli datasets delete --ds-name 'MYUSER.OLD.DATA'

# Delete a PDS member
zcli datasets delete --ds-name 'MYUSER.JCL' --member-name OLDJOB

# Rename a dataset
zcli datasets utilities rename --ds-name 'MYUSER.NEW.NAME' --from-ds-name 'MYUSER.OLD.NAME'

# Rename a PDS member
zcli datasets utilities rename --ds-name 'MYUSER.JCL' --member-name NEWMEM --from-ds-name 'MYUSER.JCL' --from-member OLDMEM

# Copy a dataset
zcli datasets utilities copy --ds-name 'MYUSER.TARGET' --from-ds-name 'MYUSER.SOURCE' --replace

# Copy all members
zcli datasets utilities copy --ds-name 'MYUSER.TARGET.PDS' --from-ds-name 'MYUSER.SOURCE.PDS' --from-member '*' --replace

# Copy a single member to a new name
zcli datasets utilities copy --ds-name 'MYUSER.PDS' --member-name NEWMEM --from-ds-name 'MYUSER.PDS' --from-member OLDMEM

# Copy from a z/OS UNIX file to a dataset
zcli datasets utilities copy --ds-name 'MYUSER.DATA' --from-file '/u/myuser/input.txt'

# Recall a migrated dataset
zcli datasets utilities hrecall --ds-name 'MYUSER.ARCHIVED' --wait

# Migrate a dataset
zcli datasets utilities hmigrate --ds-name 'MYUSER.OLDDATA'

# Delete a migrated dataset backup
zcli datasets utilities hdelete --ds-name 'MYUSER.OLDBACKUP' --purge

# Invoke IDCAMS Access Method Services
zcli datasets utilities ams \
  --input "DEFINE CLUSTER(NAME (MYUSER.KSDS) VOLUMES(VSER05)) -" \
  --input "DATA  (KILOBYTES (50 5))"

# Delete a VSAM cluster via AMS
zcli datasets utilities ams --input "DELETE MYUSER.KSDS CLUSTER"

# AMS on a remote system
zcli datasets utilities ams --input "LISTCAT ALL" --target-system SYSB

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

# Retrieve hardcopy log entries
zcli console log --hardcopy OPERLOG --time 2026-03-28T10:00:00Z

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
make build-zos      # Show z/OS native build instructions
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
