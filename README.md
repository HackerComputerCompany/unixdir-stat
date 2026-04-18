# unixdir-stat Specification

## Project Overview

- **Project Name**: unixdir-stat
- **Type**: Disk space analyzer (Unix equivalent of WinDirStat)
- **Core Functionality**: Analyze directory sizes, categorize files by type, provide interactive navigation
- **Target Users**: Unix system administrators and users who need to understand disk usage

## Architecture

### Phases
1. **REPL** (Phase 1) - Interactive command-line interface
2. **TUI** (Phase 2) - Terminal UI with ncurses
3. **GUI** (Phase 3) - Graphical user interface

### Current Phase: REPL

## Functionality Specification

### Core Features

1. **Directory Scanning**
   - Recursively scan directories
   - Calculate total size of each directory
   - Handle permission errors gracefully
   - Support symlinks (option to follow or ignore)

2. **File Type Categorization**
   - Categorize files by extension
   - Categories: Source Code, Images, Documents, Archives, Media, Config, Other
   - Display size per category

3. **REPL Commands**
   - `scan <path>` - Scan a directory
   - `ls` - List current directory contents with sizes
   - `cd <path>` - Navigate to a directory
   - `pwd` - Print working directory
   - `types` - Show file type breakdown
   - `top <n>` - Show largest files/directories
   - `sort <field>` - Sort by size/name/count (size, name, count)
   - `help` - Show available commands
   - `exit` - Exit the REPL

### Data Structures

```
Entry:
  - Path: string
  - Name: string
  - Size: int64
  - IsDir: bool
  - Children: []Entry
  - FileType: string
```

```
FileTypeStats:
  - Extension: string
  - Category: string
  - TotalSize: int64
  - FileCount: int
```

## Acceptance Criteria

1. Can scan any accessible directory
2. Displays directory sizes in human-readable format (KB, MB, GB)
3. File type breakdown shows category and size
4. Navigation commands work correctly
5. Handles errors gracefully (permission denied, not found)
6. REPL responds to commands immediately
