# Changelog

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
