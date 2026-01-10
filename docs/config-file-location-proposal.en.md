# Configuration File Location Proposal

## Issue
This document addresses the question raised in issue #7, comment #3733671954: "Since it will run on Windows or Mac, I want to decide an appropriate location for the configuration file."

## Recommended Configuration File Locations

### 1. Standard Locations by Platform

#### Windows
```
%APPDATA%\midiwav\config.toml
```
Example: `C:\Users\<Username>\AppData\Roaming\midiwav\config.toml`

#### macOS
```
~/Library/Application Support/midiwav/config.toml
```
or
```
~/.config/midiwav/config.toml
```

#### Linux (for reference)
```
~/.config/midiwav/config.toml
```

### 2. Recommended Implementation Approach

#### Option A: XDG Base Directory Specification (Recommended)

Using the standard Go library `github.com/adrg/xdg` makes cross-platform support easy.

**Benefits:**
- Automatically selects appropriate locations for Windows, macOS, and Linux
- Follows industry standards
- Easy to maintain

**Implementation example:**
```go
import "github.com/adrg/xdg"

configPath, err := xdg.ConfigFile("midiwav/config.toml")
```

This resolves to:
- Windows: `%APPDATA%\midiwav\config.toml`
- macOS: `~/Library/Application Support/midiwav/config.toml`
- Linux: `~/.config/midiwav/config.toml`

#### Option B: Custom Implementation

If using only standard library:

```go
import (
    "os"
    "path/filepath"
    "runtime"
)

func getConfigPath() (string, error) {
    var baseDir string
    
    switch runtime.GOOS {
    case "windows":
        baseDir = os.Getenv("APPDATA")
    case "darwin": // macOS
        home, _ := os.UserHomeDir()
        baseDir = filepath.Join(home, "Library", "Application Support")
    default: // Linux, etc.
        configDir := os.Getenv("XDG_CONFIG_HOME")
        if configDir == "" {
            home, _ := os.UserHomeDir()
            configDir = filepath.Join(home, ".config")
        }
        baseDir = configDir
    }
    
    return filepath.Join(baseDir, "midiwav", "config.toml"), nil
}
```

### 3. Configuration File Search Order (Recommended)

For flexibility, search for configuration files in the following order:

1. Path specified by command-line argument: `--config /path/to/config.toml`
2. Environment variable: `MIDIWAV_CONFIG=/path/to/config.toml`
3. Current directory: `./config.toml`
4. User configuration directory (as described above)

This allows convenience during development and testing while ensuring proper placement in production.

### 4. Configuration File Format

The TOML format proposed in issue #7 is an appropriate choice:

```toml
[[target]]
dir = "/abs/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"

[[target]]
dir = "/abs/path/to/dir2"
recursive = true
# Omitting discord_webhook_url means no Discord posting
```

Benefits of TOML:
- Human-readable and editable
- Mature libraries available in Go (e.g., `github.com/BurntSushi/toml`)
- Intuitive representation of arrays and tables

## Recommended Implementation Steps

1. Add `github.com/adrg/xdg` and `github.com/BurntSushi/toml` as dependencies
2. Define configuration file structures
3. Implement configuration file loading functionality
4. Maintain backward compatibility with existing environment variable configuration (environment variables take precedence over config file)

## Conclusion

**Recommendation: Option A (using xdg library)**

Reasons:
- Cross-platform support is automatic
- Follows industry standards, making behavior predictable for users
- Code is simple and maintainable
- Proven track record with many Go applications

This proposal enables appropriate configuration file placement for both Windows and macOS.
