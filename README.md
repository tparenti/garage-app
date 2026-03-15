# WrenchLog

A personal maintenance tracking web app for vehicles and equipment.
Built with Go + SQLite — no external database server needed.

## Features

- **Owners** — track vehicles/equipment per person (or yourself)
- **Vehicles / Equipment** — year, make, model, VIN, plate, color, notes
- **Maintenance Logs** — date, mileage, work performed, parts used, notes
- Single `.db` file for all data — easy to back up

## Setup

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- GCC (required to compile the SQLite driver)
  - **Windows**: Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or use WSL
  - **Mac**: `xcode-select --install`
  - **Linux**: `sudo apt install gcc` (or equivalent)

### Run

```bash
# 1. Clone / download the project
cd wrenchlog

# 2. Download dependencies
go mod tidy

# 3. Run
go run .
```

Open **http://localhost:8080** in your browser.

### Build a binary

```bash
go build -o wrenchlog .
./wrenchlog
```

## Project Structure

```
wrenchlog/
├── main.go          # HTTP handlers & routing
├── db.go            # Database models & queries
├── go.mod
├── wrenchlog.db     # Created automatically on first run
├── templates/
│   ├── base.html
│   ├── dashboard.html
│   ├── owners.html
│   ├── owner_detail.html
│   ├── owner_form.html
│   ├── vehicle_detail.html
│   ├── vehicle_form.html
│   └── log_form.html
└── static/          # (optional) CSS/JS overrides
```

## Data

The SQLite database (`wrenchlog.db`) is created automatically in the same folder as the binary.
Back it up by copying that single file.
