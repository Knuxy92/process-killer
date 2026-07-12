# Process Killer

An interactive CLI tool to view, search, and kill processes on Linux/Unix systems.

## Features

- **Interactive TUI** - Navigate processes with arrow keys
- **Process List** - View all running processes with PID, Name, CPU%, Memory%, Status, and Command
- **Search/Filter** - Press `/` to search and filter processes by name or command
- **Kill Single Process** - Select a process and press Enter to kill it
- **Kill All Matching** - Press `K` to kill ALL processes matching your search filter (e.g., all Chrome processes)
- **Kill Tree Mode** - Toggle to kill a process and all its child processes
- **Confirmation** - Safety confirmation before killing
- **Auto-refresh** - Automatically refreshes the process list after killing

## Usage

```bash
# Run the tool
./process-killer

# Or build from source
go build -o process-killer .
./process-killer
```

## Controls

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate through process list |
| `Enter` | Select process (press again to confirm kill) |
| `/` | Enter search mode to filter processes by name/command |
| `K` | Kill ALL processes matching the current search filter |
| `t` | Toggle kill-tree mode (kill children too) |
| `r` | Refresh process list |
| `q` / `Ctrl+C` | Quit |
| `Esc` | Cancel confirmation / clear search |

## Examples

### Kill a single process
1. Navigate with `↑`/`↓` to the process
2. Press `Enter` to select
3. Press `Enter` again to confirm

### Kill all Chrome processes
1. Press `/` to enter search mode
2. Type "chrome" to filter all Chrome processes
3. Press `K` to kill ALL matching processes
4. Press `Enter` to confirm

### Kill a process tree
1. Select a process
2. Press `t` to enable kill-tree mode
3. Press `Enter` twice to kill the process and all its children

## How It Works

1. **List**: Shows all running processes sorted by CPU usage
2. **Search**: Press `/` to filter processes by name or command
3. **Select**: Navigate with arrow keys, press Enter to select
4. **Confirm**: Press Enter again to confirm the kill
5. **Kill All**: After searching, press `K` to kill ALL matching processes
6. **Kill Tree**: Press `t` to toggle kill-tree mode (kills all child processes too)
7. **Result**: Process list auto-refreshes after killing

## Build from Source

```bash
# Clone project
git clone https://github.com/Knuxy92/process-killer.git

# Cd to project folder
cd process-killer

# Build
go build -o process-killer .

# Run
./process-killer
```

## Requirements

- Go 1.21+
- Linux/Unix system (uses `/proc` filesystem)
