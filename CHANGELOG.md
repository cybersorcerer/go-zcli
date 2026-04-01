# Changelog

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
