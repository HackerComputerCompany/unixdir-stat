# unixdir-stat

Interactive disk-usage explorer for Unix-like systems: scan a directory tree, inspect sizes, browse with simple commands, and see a coarse breakdown by file category (source, images, documents, and so on). The current interface is a **text REPL**; a TUI and GUI are planned later.

## Requirements

- **Go**: a toolchain that satisfies [`go.mod`](go.mod) (currently **Go 1.25+**). Check with:

  ```bash
  go version
  ```

- **OS**: macOS or Linux (and other platforms Go supports). No extra native libraries are required.

### Installing Go

**macOS**

- [Homebrew](https://brew.sh/): `brew install go`
- Or install a official package from [https://go.dev/dl/](https://go.dev/dl/)

**Linux**

- Debian/Ubuntu: `sudo apt install golang-go` (version may lag; use backports or a tarball if `go version` is too old).
- Fedora: `sudo dnf install golang`.
- Or unpack the Linux tarball from [https://go.dev/dl/](https://go.dev/dl/) into `/usr/local/go` and add `/usr/local/go/bin` to your `PATH` (see the install page for copy-paste steps).

If your distro’s Go is older than this module requires, prefer the **official installer** from go.dev so `go.mod` and `go test` match CI and other developers.

## Build

From the repository root:

```bash
go build -o unixdir-stat ./cmd
```

This produces a single binary, `unixdir-stat`, in the current directory. The same command works on **macOS** and **Linux**; Go cross-compiles if you set `GOOS`/`GOARCH` (optional):

```bash
GOOS=linux GOARCH=amd64 go build -o unixdir-stat-linux-amd64 ./cmd
GOOS=darwin GOARCH=arm64 go build -o unixdir-stat-darwin-arm64 ./cmd
```

## Run

```bash
./unixdir-stat
```

During development you can skip the binary:

```bash
go run ./cmd
```

There are no CLI flags yet; the app starts the REPL immediately.

## User manual (REPL)

After startup you see a `>` prompt. Type a command and press Enter. Paths can be relative, absolute, or prefixed with `~` for your home directory.

| Command | Description |
|--------|-------------|
| `scan [path]` | Scan a directory (default: `.`). Resets category stats and sets the current tree to that directory. Shows progress while scanning. |
| `ls` | List immediate children of the scanned directory with human-readable sizes and type category. Children are ordered by size unless `sort name` is active (see below). |
| `cd [path]` | **With an argument**: `cd subdir` moves into a child of the current path (re-scans that path). `cd ..` moves to the parent directory (re-scans). **With no argument**: re-scan the current directory (refresh sizes and children). |
| `pwd` | Print the absolute path of the directory you are currently viewing (after `scan` / `cd`). |
| `types` | Show aggregate size and file count per **category** (Source Code, Images, Documents, etc.) for the tree produced by the most recent `scan` or `cd` (each `scan`, `cd` into a child, `cd ..`, or bare `cd` resets category counters before scanning). |
| `top [n]` | Show the **n** largest immediate children of the current directory (default **10**). Order follows the scan (largest first). |
| `sort [field]` | Without arguments: print the current sort field. With `size`, `name`, or `count`: store the sort field. **`ls` uses this only for `name`** (alphabetical); the listing is otherwise ordered by size as produced by the scanner. |
| `help` | Print short command help. |
| `exit` | Quit (also `quit` or `q`). |

**Signals**: first **Ctrl+C** prints a hint; use `exit` or press Ctrl+C again to leave (second interrupt exits the loop).

**Permissions**: if a subdirectory cannot be read, that branch is skipped; the rest of the tree is still reported.

**Symlinks**: directory entries that resolve as directories are traversed like normal folders; behavior follows Go’s `os.ReadDir` / file walk semantics on your OS.

## Examples

**Quick session on the project itself**

```text
> scan .
> ls
> top 5
> types
> cd internal
> ls
> pwd
> cd ..
> sort name
> ls
> exit
```

**Scan your home directory (can take a while)**

```text
> scan ~
> top 20
> types
```

**One-liner from the shell** (non-interactive use is limited; this only runs the first command unless you script more):

```bash
printf 'scan .\nls\nexit\n' | ./unixdir-stat
```

## Tests

```bash
go test ./...
```

## Project status and design

| Phase | Description |
|-------|-------------|
| **1 — REPL** (current) | Interactive commands: scan, navigate, list, categories, largest children. |
| **2 — TUI** (planned) | Terminal UI (e.g. ncurses-style). |
| **3 — GUI** (planned) | Graphical disk map-style view. |

### Implementation notes

- **Entry tree**: each node has path, name, size, directory flag, children, and a **category** string for files (from extension).
- **Sizes**: displayed with KB/MB/GB/TB/PB-style units (base 1024).

### Data structures (reference)

```text
Entry: Path, Name, Size, IsDir, Children []*Entry, FileType (category)
FileTypeStats: Extension, Category, TotalSize, FileCount
```

## Roadmap (from original spec)

Planned direction: richer navigation, optional symlink policy, and clearer sort behavior for `count` / `size` on `ls`. The REPL should remain usable for scripting small workflows where a TUI is not wanted.

**Acceptance-style goals for the REPL phase**: correct totals for readable trees, human-readable sizes, working navigation commands, responsive loop, and clear handling of missing paths or permission errors on subtrees.
