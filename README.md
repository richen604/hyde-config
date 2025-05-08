# Hyde Configuration Parser (Go Implementation)

A Go-based tool for parsing TOML configuration files for HyDE and Hyprland. This is a direct port of the Python implementation at `~/.local/lib/hyde/parse.config.py`.

This is to shred some memory and improve performance.

## Features

- Parse TOML configuration files into environment variables
- Parse Hyprland-specific sections into Hyprland configuration
- Watch for configuration file changes and automatically reprocess
- Systemd service integration for automatic startup
- Supports exporting environment variables

## Installation

### Prerequisites

- Go 1.20 or later
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/hyde-project/hyde-config.git
cd hyde-config

# Build the application
make deps
make build


# Install the application
make install
```

## Usage

```bash
# Run with default settings (daemon mode, export enabled)
hyde-config

# Run once without daemon mode
hyde-config --no-daemon

# Run without exporting environment variables
hyde-config --no-export

# Enable verbose logging
hyde-config --verbose

# Enable debug logging (detailed diagnostic information)
hyde-config --debug

# Specify custom file locations
hyde-config --input /path/to/config.toml --env /path/to/env --hypr /path/to/hyprland.conf
```

### Command-Line Options

- `--input`: Path to TOML configuration file (default: `$XDG_CONFIG_HOME/hyde/config.toml`)
- `--env`: Path to output environment variables file (default: `$XDG_STATE_HOME/hyde/config`)
- `--hypr`: Path to output Hyprland configuration file (default: `$XDG_STATE_HOME/hyde/hyprland.conf`)
- `--no-daemon`: Run in one-off mode without watching for changes
- `--no-export`: Disable exporting environment variables
- `--verbose`: Enable verbose logging
- `--debug`: Enable debug mode with detailed logging

## Enabling as a Service

### Systemd User Service

The hyde-config tool can be set up as a systemd user service to run automatically when you log in:

```bash
# Install the service file
make service-install

# Enable the service to start automatically at login
make service-enable

# Start the service immediately
make service-start
```

### Checking Service Status

```bash
# Check if the service is running
make service-status

# or directly with systemd
systemctl --user status hyde-config.service
```

### Managing the Service

```bash
# Stop the service
make service-stop

# Disable the service (prevent it from starting at login)
make service-disable
```

## Configuration File Example

Hyde-config expects a TOML configuration file with environment variables and Hyprland settings:

```toml


"$schema" = "../../.local/share/hyde/schema/config_toml.json"

[wallpaper]
    #   backend = "mpvpaper" # swww,mpvpaper,pcmanfm (--desktop),hyprpaper
    custom_paths = [

        "$XDG_VIDEOS_DIR/Wallpapers",
    ] # List of paths to search for wallpapers
    filetypes = [ "mp4", "mkv" ] # explicit file types
    multi_display = true # Set to true if you want to use different wallpapers on different displays


[mediaplayer]
    max_length     = 20
    prefix_paused  = "⏸️"
    prefix_playing = "▶️"
    standby_text   = "📻"

[wallbash]
    skip_template = [
        # "waybar",    
    ]


[rofi]
    scale = 9

[rofi.launch]
    scale = 10
    # style = "style_2"
    drun_style = "launchpad"


[rofi.theme]
    scale = 5
    # style = 2

[rofi.cliphist]
    scale = 8

[rofi.wallpaper]
    scale = 5
```

## Development

### Building for Distribution

```bash
# Build optimized release binary
make release

# Build for multiple platforms
make build-all
```

### Cleaning Build Artifacts

```bash
make clean
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.