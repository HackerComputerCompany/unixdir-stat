package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"unixdir-stat/internal/scanner"
)

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func printProgress(files, dirs int, size int64) {
	fmt.Printf("\r  [%s] %d files, %d dirs", scanner.FormatSize(size), files, dirs)
	fmt.Print("                    ")
}

type State struct {
	currentPath string
	rootEntry   *scanner.Entry
	sortField   string
	shouldExit  bool
}

func Run() {
	state := &State{
		currentPath: ".",
		sortField:   "size",
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigChan)

	fmt.Println("unixdir-stat REPL")
	fmt.Println("Type 'help' for available commands")
	fmt.Println("(Press Ctrl+C to exit)")

	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-sigChan:
			fmt.Println("\nUse 'exit' to quit, or press Ctrl+C again to force exit.")
			continue
		default:
		}

		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		state.processLine(line)
		if state.shouldExit {
			fmt.Println("Goodbye!")
			return
		}
	}
}

func (s *State) processLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "scan":
		s.handleScan(args)
	case "ls":
		s.handleLs()
	case "cd":
		s.handleCd(args)
	case "pwd":
		s.handlePwd()
	case "types":
		s.handleTypes()
	case "top":
		s.handleTop(args)
	case "sort":
		s.handleSort(args)
	case "help":
		s.handleHelp()
	case "exit", "quit", "q":
		s.shouldExit = true
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func (s *State) handleScan(args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	path = expandPath(path)

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Scanning %s...\n", absPath)

	scanner.ResetStats()

	var lastPrint int
	scanner.SetProgressCallback(func(files, dirs int, size int64) {
		if files-lastPrint >= 10 || files < lastPrint {
			printProgress(files, dirs, size)
			lastPrint = files
		}
	})

	entry, err := scanner.Scan(absPath)
	if err != nil {
		fmt.Printf("Error scanning: %v\n", err)
		return
	}

	s.rootEntry = entry
	s.currentPath = absPath

	stats := scanner.GetFileTypeStats()
	totalFiles := 0
	for _, st := range stats {
		totalFiles += st.FileCount
	}

	dirCount := 0
	for _, child := range entry.Children {
		if child.IsDir {
			dirCount++
		}
	}

	fmt.Printf("\nScan complete. Total: %s, %d files, %d directories\n", scanner.FormatSize(entry.Size), totalFiles, dirCount)
}

func (s *State) handleLs() {
	if s.rootEntry == nil {
		fmt.Println("No directory scanned yet. Use 'scan <path>' first.")
		return
	}

	entries := s.rootEntry.Children
	if s.sortField == "name" {
		sorted := make([]*scanner.Entry, len(entries))
		copy(sorted, entries)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].Name > sorted[j].Name {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		entries = sorted
	}

	fmt.Printf("%-40s %12s %8s\n", "Name", "Size", "Type")
	fmt.Println(strings.Repeat("-", 65))

	for _, e := range entries {
		dirIndicator := ""
		if e.IsDir {
			dirIndicator = "/"
		}
		fmt.Printf("%-40s %12s %8s\n",
			e.Name+dirIndicator,
			scanner.FormatSize(e.Size),
			e.FileType)
	}
}

func (s *State) handleCd(args []string) {
	if s.rootEntry == nil {
		fmt.Println("No directory scanned yet. Use 'scan <path>' first.")
		return
	}

	if len(args) == 0 {
		s.currentPath = s.rootEntry.Path
		scanner.ResetStats()
		s.rootEntry, _ = scanner.Scan(s.rootEntry.Path)
		return
	}

	target := expandPath(args[0])
	if target == ".." {
		parent := filepath.Dir(s.currentPath)
		s.currentPath = parent
		scanner.ResetStats()
		s.rootEntry, _ = scanner.Scan(parent)
	} else {
		newPath := filepath.Join(s.currentPath, target)
		s.currentPath = newPath
		scanner.ResetStats()
		entry, err := scanner.Scan(newPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		s.rootEntry = entry
	}
	fmt.Printf("Changed to: %s\n", s.currentPath)
}

func (s *State) handlePwd() {
	fmt.Println(s.currentPath)
}

func (s *State) handleTypes() {
	stats := scanner.GetFileTypeStats()
	if len(stats) == 0 {
		fmt.Println("No files scanned yet.")
		return
	}

	fmt.Printf("%-15s %12s %8s\n", "Category", "Total Size", "Files")
	fmt.Println(strings.Repeat("-", 40))

	for _, stat := range stats {
		fmt.Printf("%-15s %12s %8d\n",
			stat.Category,
			scanner.FormatSize(stat.TotalSize),
			stat.FileCount)
	}
}

func (s *State) handleTop(args []string) {
	n := 10
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &n)
	}

	if s.rootEntry == nil {
		fmt.Println("No directory scanned yet.")
		return
	}

	entries := scanner.GetTopEntries(s.rootEntry, n)

	fmt.Printf("%-40s %12s %8s\n", "Name", "Size", "Type")
	fmt.Println(strings.Repeat("-", 65))

	for i, e := range entries {
		dirIndicator := ""
		if e.IsDir {
			dirIndicator = "/"
		}
		fmt.Printf("%2d. %-37s %12s %8s\n",
			i+1,
			e.Name+dirIndicator,
			scanner.FormatSize(e.Size),
			e.FileType)
	}
}

func (s *State) handleSort(args []string) {
	if len(args) == 0 {
		fmt.Printf("Current sort: %s\n", s.sortField)
		return
	}

	field := args[0]
	switch field {
	case "size", "name", "count":
		s.sortField = field
		fmt.Printf("Sort field set to: %s\n", field)
	default:
		fmt.Printf("Unknown sort field: %s (use size, name, or count)\n", field)
	}
}

func (s *State) handleHelp() {
	fmt.Println(`Available commands:
  scan <path>    Scan a directory (default: current dir)
  ls             List directory contents with sizes
  cd <path>     Navigate to a directory (use '..' for parent)
  pwd           Print current directory
  types         Show file type breakdown
  top <n>       Show largest n entries (default: 10)
  sort <field>  Sort by size, name, or count
  help          Show this help
  exit          Exit the REPL`)
}
