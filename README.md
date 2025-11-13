# barbecue

Barbecue uses the Starkeeper Voyager Astronomy API to control that the warmup of your camera is done before switching power off.

## Usage

### Basic Commands

```bash
# Show general help
./bin/barbecue --help

# Report camera status and check if idle
./bin/barbecue report

# Check if camera needs warming (requires warm temperature threshold)
./bin/barbecue needWarming --warm-temp 10
```

### Global Flags

- `--addr string`: Voyager TCP server address (default "127.0.0.1:5950")
- `--level string`: Set log level of barbecue (default "warn")

### Commands

#### `report`
Report camera status and check if idle.

Connects to Voyager server, retrieves camera status, and determines if the camera is idle (temperature >= ambient and power off).

**Example:**
```bash
./bin/barbecue report --addr 192.168.1.100:5950 --level debug
```

#### `needWarming`
Check if camera needs warming.

Check if the camera cooling temperature is below the specified warm temperature threshold. Connects to Voyager server, retrieves camera status, and compares the temperature.

**Flags:**
- `--warm-temp int`: Warm temperature threshold (required)

**Example:**
```bash
./bin/barbecue needWarming --warm-temp 15 --addr 192.168.1.100:5950
```

### Configuration

Barbecue reads configuration from `conf/barbecue.yaml` in the current directory. The configuration file should contain:

```yaml
voyager:
  tcpserver:
    address: 127.0.0.1
    port: 5950
```
