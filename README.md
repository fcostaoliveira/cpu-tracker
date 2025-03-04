# CPU Tracker

CPU Tracker is a simple tool to monitor and track CPU usage of specific processes.

## Installation

### Prerequisites
- Go (if building from source)
- wget (for downloading prebuilt binaries)
- jq (for processing JSON output)

### Download Standalone Binaries (No Go Required)
If you don't have Go installed and just want to use the prebuilt binaries, you can download them from the latest release:

[CPU Tracker Releases](https://github.com/fcostaoliveira/cpu-tracker/releases/latest)

| OS     | Arch             | Link |
|--------|-----------------|------|
| Linux  | amd64 (x86_64)  | [cpu-tracker-linux-amd64](https://github.com/fcostaoliveira/cpu-tracker/releases/latest/download/cpu-tracker-linux-amd64.tar.gz) |
| Linux  | arm64 (ARM 64)  | [cpu-tracker-linux-arm64](https://github.com/fcostaoliveira/cpu-tracker/releases/latest/download/cpu-tracker-linux-arm64.tar.gz) |
| macOS  | amd64 (x86_64)  | [cpu-tracker-darwin-amd64](https://github.com/fcostaoliveira/cpu-tracker/releases/latest/download/cpu-tracker-darwin-amd64.tar.gz) |
| macOS  | arm64 (M1/M2)   | [cpu-tracker-darwin-arm64](https://github.com/fcostaoliveira/cpu-tracker/releases/latest/download/cpu-tracker-darwin-arm64.tar.gz) |

#### Quick Install Script
```bash
wget -c https://github.com/fcostaoliveira/cpu-tracker/releases/latest/download/cpu-tracker-$(uname -mrs | awk '{ print tolower($1) }')-$(dpkg --print-architecture).tar.gz -O - | tar -xz

# Run the tool
./cpu-tracker --help
```

### Installation via Go
If you prefer to build from source, run:
```bash
go install github.com/fcostaoliveira/cpu-tracker@latest
```

## Usage

### Start Tracking a Process
```bash
curl -X POST http://localhost:5000/start/pgrep/redis-server
```

### Stop Tracking a Process
```bash
curl -s -X POST http://localhost:5000/stop/redis-server | jq
```
Example output:
```json
{
  "message": "Stopped process redis-server",
  "median_cpu": 10.5,
  "p95_cpu": 15.2,
  "p99_cpu": 18.7,
  "cpu_details": [
    { "timestamp": "2025-03-04T12:00:00Z", "user_cpu": 5.2, "system_cpu": 3.1, "total_cpu": 8.3 }
  ]
}
```

### Stop Tracking All Processes
```bash
curl -s -X POST http://localhost:5000/stop/ | jq
```

## License
MIT License. See [LICENSE](LICENSE) for details.

---

This project is maintained by [fcostaoliveira](https://github.com/fcostaoliveira).

