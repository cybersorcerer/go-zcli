# Changelog

## v0.5.2 - 2026-05-13

### New Features

- **Datasets write**: New `datasets write` command to write data from a local file or stdin to an existing sequential dataset or PDS/PDSE member; creates the member if it does not exist
- **Datasets write**: Supports all z/OSMF write headers: `X-IBM-Data-Type` (text/binary/record), `X-IBM-Migrated-Recall`, `X-IBM-Obtain-ENQ`, `X-IBM-Session-Ref`, `X-IBM-Release-ENQ`, `X-IBM-Dsname-Encoding`, `If-Match`, `X-IBM-Target-System`/User/Password
- **Datasets write**: Text mode options: `--file-encoding` (alternate EBCDIC code page), `--crlf` (CRLF terminators), `--wrap` (avoid truncation on F/FB datasets)
- **Datasets write**: ENQ auto-released at end of request when `--obtain-enq` is specified without `--session-ref` (atomic lock/write/unlock)
- **Datasets write**: Pipeline support via `--session-ref` + `--release-enq` for read-modify-write flows with ENQ held across multiple commands
- **Datasets write**: Optimistic locking via `--if-match` (ETag); mutually exclusive with `--obtain-enq`
- **Datasets write**: Pure ENQ release mode — `--session-ref` + `--release-enq` without content source releases a held ENQ without writing data
- **Datasets write**: ETag and `X-IBM-Session-Ref` printed from response headers when present

## v0.5.1 - 2026-04-06

### New Features

- **Parmlib**: New `parmlib validate` command for z/OS parmlib member syntax validation using `PUT /zosmf/parmlib/v1/<membertype>/validate`; supports all 38 parmlib member types
- **Parmlib**: Scenario 1 — validate all active members of a type (active LOADxx)
- **Parmlib**: Scenario 2 — validate all members of a type via specific LOADxx (`--load-member`, `--load-dataset`, `--load-volser`)
- **Parmlib**: Scenario 3 — validate a specific member by name (`--member`, `--dataset`, `--volser`)
- **Parmlib**: Scenario 4 — validate parmlib content piped from stdin
- **Parmlib**: `--deep` flag for deep validation of all parmlib types via LOADxx (LOAD membertype only)
- **Parmlib**: `--text` output with per-member result, failure location (line/position), current/expected fields
- **TSO command**: Rewritten to use new stateless `PUT /zosmf/tsoApp/v1/tso` API with `{"tsoCmd": "..."}` body; removed incorrect start/stop address space logic

### Changes

- **Console log**: `--time` flag now accepts human-friendly formats: `2006-01-02`, `2006-01-02 15:04`, `2006-01-02 15:04:05`, `15:04`, `15:04:05`, and ISO 8601 — converted internally to ISO 8601 UTC
- **Console log**: New `--ago` flag for relative time input: `30s`, `10m`, `2h`, `1h30m`, `1h30m45s` — converted to UNIX timestamp; takes precedence over `--time`
- **Console log**: `--text` output now correctly handles multi-line messages containing `\r` — continuation lines are joined with a space and the terminal wraps naturally

### Bug Fixes

- Fixed `tso command` failing with `no servletKey in TSO start response` — the previous implementation incorrectly used the legacy session-based start/issue/stop flow instead of the single-call stateless API

## v0.5.0 - 2026-04-01

### New Features

- **Jobs submit**: Added `--remote-file` flag for submitting JCL from a host dataset or UNIX file (sends `{"file": "..."}` JSON body); mutually exclusive with `--file-name`
- **Jobs submit**: Auto-detection of Content-Type based on `--intrdr-mode` (text/plain for TEXT, application/octet-stream for RECORD/BINARY, application/json for remote-file)
- **Jobs submit**: Added internal reader headers: `--intrdr-class`, `--intrdr-recfm`, `--intrdr-lrecl`, `--intrdr-mode`, `--intrdr-file-encoding`
- **Jobs submit**: Added `--user-correlator`, `--jcl-symbol` (repeatable, NAME=VALUE format → `X-IBM-JCL-Symbol-{NAME}` headers), `--notification-url`, `--notification-options`
- **Jobs submit**: Added `--target-system`, `--target-user`, `--target-password` for remote system targeting
- **Jobs hold/release/cancel**: Added `--async` flag for version 1.0 asynchronous processing (default: sync v2.0)
- **Jobs hold/release/cancel**: Added `--target-system`, `--target-user`, `--target-password` flags
- **Jobs change-class**: Added `--async` flag (version 1.0 vs 2.0 processing)
- **Jobs change-class**: Added `--target-system`, `--target-user`, `--target-password` flags
- **Filesystems TUI**: Expanded create form from 3 to 11 fields — added owner, group, perms (POSIX permissions), storage class, management class, data class, volumes (comma-separated), timeout
- **Filesystems TUI**: New `fsCreateOpts` struct and `doFsCreate` handles all optional API fields including volumes as comma-split array and timeout as query parameter

### Changes

- **All TUIs**: Changed `esc` key to `F3` for back/cancel actions; `esc` no longer works — updated all footer descriptions accordingly (`F3 back`, `[F3] cancel`)
- **All TUIs**: Added alternative keys for environments where F-keys are intercepted (e.g. Omnissa Horizon Client): `esc` as alternative for `F3`, `Ctrl+U` for `F7` (page up), `Ctrl+D` for `F8` (page down)
- **All commands**: Added `Args: cobra.NoArgs` to all 60+ leaf commands — unknown positional arguments now produce an error with exit code 1 instead of being silently ignored
- **Filesystems mount/unmount**: Added AUTHORIZATION section in command description — requires `BPX.SUPERUSER` (FACILITY) or `SUPERUSER.FILESYS.MOUNT` (UNIXPRIV) RACF access
- **Build on z/OS**: Updated `make build-zos` instructions and README to document required extended module updates (`golang.org/x/sys`, `golang.org/x/sync`, `golang.org/x/text`) before building natively on z/OS with the Open Enterprise SDK for Go

### Bug Fixes

- Fixed `datasets create` sending empty/zero optional fields (`like`, `volser`, `unit`, `storclass`, `mgntclass`, `dataclass`, `dsntype`, `avgblk`, `blksize`) in JSON body — z/OSMF rejected empty `like` field with rc=4, reason=13

## v0.4.0 - 2026-03-28

### New Features

- **Console**: Full z/OSMF REST Console API support (async, sol-key, unsol-key, auth, routcode, mscope, storage, auto, custom console names)
- **Console**: `get-response` subcommand for retrieving delayed command responses
- **Console**: `get-detection` subcommand for unsolicited keyword detection results
- **Console**: `log` subcommand for hardcopy log retrieval (OPERLOG/SYSLOG) with time filtering
- **Sysvar**: Complete system variable management (get, create, import, export, delete)
- **Sysvar**: TUI with inline create and delete actions (with confirmation dialog)
- **Sysvar**: Local file transfer for import/export with automatic ISO8859-1 tagging
- **Subsystems**: New MVS subsystems list command using `/rest/mvssubs` API
- **Subsystems**: TUI for browsing MVS subsystems
- **Datasets list**: Configurable `--attributes` and `--max-items` flags
- **Datasets members**: Added `--start`, `--attributes`, `--max-items`, `--migrated-recall` flags
- **Datasets read**: Full API support with search/research query params, record-range, data-type, obtain-enq, session-ref, migrated-recall, return-etag, if-none-match headers
- **Datasets create**: Added custom headers: data-type, file-encoding, crlf, wrap, migrated-recall, obtain-enq, session-ref, release-enq, encoding
- **Datasets utilities rename**: Corrected API to use target dataset in URL path with `from-dataset` in body; added `--enq` (SHRW/EXCLU) and `--migrated-recall`
- **Datasets utilities copy**: New command for copying datasets, members, and z/OS UNIX files with full options (from-dataset, from-file, volser, alias, enq, replace, bpxk-autocvt)
- **Datasets utilities ams**: New IDCAMS Access Method Services command for invoking AMS statements
- **CI/CD**: GitHub Actions release workflow for tag-based cross-platform builds
- **Build**: Added `make build-zos` target with native build instructions for z/OS

### Bug Fixes

- Fixed reverse video display in CLI help output (removed `cc.Italic` from coloredcobra config)
- Fixed MFS TUI showing "0 filesystems loaded" (`mode` field type mismatch: `[]int` -> `[]string`)
- Fixed subsystems command using wrong API endpoint (`/restfiles/mfs` -> `/rest/mvssubs`)
- Fixed `--dsn-level/-d` shorthand collision with global `--debug/-d` on datasets list
- Fixed datasets members `pattern` query parameter using `&` instead of `?` as first separator
- Fixed datasets utilities hrecall/hmigrate `wait` field serialized as string instead of boolean in JSON body
- Fixed datasets utilities rename URL path pointing to source instead of target dataset
- Fixed datasets create not passing custom headers to HTTP client (was `nil`)

## v0.3.0 - 2025-12-01

- First official version
- Jobs, Datasets, Files, Filesystems, MFS, Software, Console (basic), TSO, Topology, Notifications, Info, Profile commands
