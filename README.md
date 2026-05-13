# go-devicebase

Go SDK for the [Devicebase](https://github.com/devicebase) device automation API. Control Android, HarmonyOS, and iOS devices programmatically via HTTP.

## Installation

```bash
go get github.com/devicebase/go-devicebase
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    devicebase "github.com/devicebase/go-devicebase"
)

func main() {
    client := devicebase.NewClient(
        devicebase.WithAPIKey("your-api-key"),
        devicebase.WithSerial("device-serial-number"),
    )

    // Get device info
    info, err := client.GetDeviceInfo()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("device:", info.Serial)

    // Take a screenshot
    screenshot, err := client.GetScreenshot()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("screenshot: %d bytes\n", len(screenshot))

    // Control the device
    client.Tap(100, 200)
    client.LaunchApp("com.tencent.mm")
}
```

## Configuration

```go
client := devicebase.NewClient(
    devicebase.WithAPIKey("your-api-key"),           // Required
    devicebase.WithSerial("device-serial-number"),   // Required
    devicebase.WithBaseURL("https://api.devicebase.cn"), // Optional, default shown
    devicebase.WithTimeout(30 * time.Second),         // Optional
    devicebase.WithHTTPClient(httpClient),            // Optional
)
```

Environment variables are used as fallbacks:

- `DEVICEBASE_API_KEY` — API key for authentication
- `DEVICEBASE_BASE_URL` — API base URL (default: `https://api.devicebase.cn`)

## Device Control

### Touch Operations

```go
client.Tap(100, 200)           // Single tap
client.DoubleTap(100, 200)     // Double tap
client.LongPress(100, 200)     // Long press
client.Swipe(0, 500, 500, 500) // Swipe gesture
```

### Navigation

```go
client.Back()  // Back button
client.Home()  // Home button
```

### App Operations

```go
result, _ := client.LaunchApp("com.tencent.mm")
appInfo, _ := client.GetCurrentApp()
fmt.Println(appInfo.Data["package"])
```

### Text Input

```go
client.InputText("Hello World")
client.ClearText()
```

### UI Hierarchy

```go
hierarchy, _ := client.DumpHierarchy()
fmt.Println(hierarchy.Data)
```

### Screenshots

```go
jpegBytes, _ := client.GetScreenshot()
downloadBytes, _ := client.DownloadScreenshot()
```

## Error Handling

```go
import "errors"

result, err := client.Tap(100, 200)
if err != nil {
    var authErr *devicebase.AuthenticationError
    var notFoundErr *devicebase.DeviceNotFoundError
    var valErr *devicebase.ValidationError

    switch {
    case errors.As(err, &authErr):
        fmt.Println("authentication failed:", authErr.Message)
    case errors.As(err, &notFoundErr):
        fmt.Println("device not found:", notFoundErr.Message)
    case errors.As(err, &valErr):
        fmt.Println("validation error:", valErr.Message)
    default:
        fmt.Println("error:", err)
    }
}
```

## License

MIT
