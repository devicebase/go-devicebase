# go-devicebase

Go SDK for the [Devicebase](https://github.com/devicebase) device automation API. Control three device platforms programmatically over HTTP:

| Platform | Covers | Path family |
|----------|--------|-------------|
| **mobile** | Android, HarmonyOS, iOS | `/v1/{action}/{serialno}` |
| **browser** | Chrome / Chromium / Edge over CDP | `/api/browser/{serialno}/{action...}` |
| **computer** | macOS / Windows / Linux desktops | `/api/computer/{serialno}/{action}` |

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
        devicebase.WithSerial("db-mttul4i41di8"),
    )

    info, err := client.GetDeviceInfo()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("device:", info.Serial)

    screenshot, err := client.GetScreenshot()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("screenshot: %d bytes\n", len(screenshot))

    if _, err := client.Tap(100, 200); err != nil {
        log.Fatal(err)
    }
    if _, err := client.LaunchApp("com.tencent.mm"); err != nil {
        log.Fatal(err)
    }
}
```

## Finding a device

`ListDevices` is the entry point: it is the only call that needs no serial, and it is how a `serialno` is discovered. Filters are resolved server-side.

```go
devices, err := client.ListDevices(devicebase.ListDevicesRequest{
    Type:  "browser",   // category bucket, or a system type such as "macos"
    State: "free",
})

for _, device := range devices {
    fmt.Println(device.Serialno, device.Type, device.OSType)
}
```

`Type` accepts either a **category bucket** (`mobile`, `browser`, `computer`) or a **system type** (`android`, `harmonyos`, `ios`, `macos`, `windows`, `linux`, `chrome`, `chromium`, `edge`, `other`). Buckets match against the device's `os_type`, because a device row only carries the coarse `type` — a Chrome browser is `type=browser` with `os_type=Chrome`. `linux` means every computer that is neither macOS nor Windows, and `other` means every browser that is not Chrome/Chromium/Edge.

`Limit` caps the result (the server defaults to 10, clamped to 1-100).

### Serial numbers

The device identifier is the **`serialno`** field (e.g. `db-mttul4i41di8`), which the server issues. The `device_sn` UUID also resolves — the gateway looks devices up with `WHERE (serialno = ? OR device_sn = ?)` — but `serialno` is the primary key.

## Configuration

```go
client := devicebase.NewClient(
    devicebase.WithAPIKey("your-api-key"),               // Optional: falls back to DEVICEBASE_API_KEY
    devicebase.WithSerial("db-mttul4i41di8"),            // Optional: for the mobile methods
    devicebase.WithBaseURL("https://api.devicebase.cn"), // Optional, default shown
    devicebase.WithTimeout(30 * time.Second),            // Optional, default shown
    devicebase.WithHTTPClient(httpClient),               // Optional
)
```

Environment variables are used as fallbacks:

- `DEVICEBASE_API_KEY` — API key for authentication
- `DEVICEBASE_BASE_URL` — API base URL (default: `https://api.devicebase.cn`)

`WithSerial` binds a device for the **mobile** methods only. The browser, computer and list methods take the serialno per call, so one client can drive many devices across platforms:

```go
client := devicebase.NewClient(devicebase.WithAPIKey("your-api-key"))

client.BrowserNavigate("db-mtsi49bf0mqb", "https://example.com")
client.ComputerClick("db-mtthisv311f1", devicebase.ComputerClickRequest{X: 640, Y: 360})
```

A missing API key is reported by the first request as an `AuthenticationError`, not at construction, so a client can be built before the environment is fully set up.

## Mobile

```go
// Touch
client.Tap(100, 200)
client.DoubleTap(100, 200)
client.LongPress(100, 200)
client.Swipe(0, 500, 500, 500)

// Navigation
client.Back()
client.Home()

// Apps
client.LaunchApp("com.tencent.mm")
client.StopApp("com.tencent.mm")
client.StopCurrentApp()
appInfo, _ := client.GetCurrentApp()

// Text
client.InputText("Hello World")
client.ClearText()

// Inspection
info, _ := client.GetDeviceInfo()
hierarchy, _ := client.DumpHierarchy()

// Shell (adb/hdc platforms only)
result, _ := client.Bash("ls -la /sdcard")
fmt.Println(result.Data["exitCode"])

// Install — the path is on the agent host, not local
install, _ := client.InstallApp("/tmp/app.apk")
client.InstallStatus(install.Data["installId"].(string))
```

## Browser

The serialno of a registered browser device, from `ListDevices{Type: "browser"}`. Selectors are CSS selectors.

```go
client.BrowserNavigate(serialno, "https://example.com")
client.BrowserRefresh(serialno)
client.BrowserGoBack(serialno)
client.BrowserGoForward(serialno)

client.BrowserClick(serialno, "button#submit")
client.BrowserFill(serialno, "#search", "devicebase")
client.BrowserSelect(serialno, "#country", "CN")
client.BrowserText(serialno, "#search")
client.BrowserAttribute(serialno, "a.logo", "href")
client.BrowserExists(serialno, ".modal")
client.BrowserExecute(serialno, "document.title") // danger tier: runs in the page

client.BrowserInput(serialno, "hello 世界") // CDP insertText, reliable for CJK
client.BrowserHotkey(serialno, []string{"Meta", "a"})

client.BrowserState(serialno)
client.BrowserTabs(serialno)
client.BrowserTabOpen(serialno, "https://example.com")
client.BrowserTabSwitch(serialno, tabID)
client.BrowserTabClose(serialno, tabID)
client.BrowserTabCloseAll(serialno)

client.BrowserLaunch(serialno)
client.BrowserClose(serialno)
```

Editing shortcuts (`Meta a`, `Backspace`) act on the page. Browser-chrome shortcuts such as `Control t` are **not reachable** — CDP drives the page, not the browser UI.

## Computer

The serialno of a registered computer device, from `ListDevices{Type: "computer"}`. Coordinates are absolute screen pixels.

```go
client.ComputerClick(serialno, devicebase.ComputerClickRequest{X: 640, Y: 360})
client.ComputerClick(serialno, devicebase.ComputerClickRequest{
    X: 640, Y: 360, Button: devicebase.ButtonRight, // empty → server default "left"
})
client.ComputerDoubleClick(serialno, devicebase.Point{X: 640, Y: 360})
client.ComputerLongClick(serialno, devicebase.ComputerLongClickRequest{X: 640, Y: 360, Duration: 2})
client.ComputerMove(serialno, devicebase.Point{X: 100, Y: 100})
client.ComputerDrag(serialno, devicebase.Bounds{X1: 100, Y1: 100, X2: 800, Y2: 600})
client.ComputerScroll(serialno, devicebase.ScrollRequest{
    Direction: devicebase.ScrollDown,
    Amount:    5, // zero → driver default
})

client.ComputerTypeText(serialno, "hello")
client.ComputerPress(serialno, "Enter")
client.ComputerHotkey(serialno, []string{"Control", "Shift", "Escape"})

client.ComputerPosition(serialno)
client.ComputerScreenSize(serialno)
client.ComputerPermissions(serialno)
client.ComputerLaunchApp(serialno, "Visual Studio Code")

// Wait takes MILLISECONDS; bash --timeout takes SECONDS.
client.ComputerWait(serialno, 2000)
client.ComputerBash(serialno, "ls -la", 30)
```

`ComputerBash` runs on the host machine as the desktop user, unsandboxed, under the platform default shell — treat it as shell access. A non-zero command exit is reported in `data.exitCode`, not as an error, because the API call itself succeeded:

```go
result, err := client.ComputerBash(serialno, "exit 3", 10)
// err == nil; result.Data["exitCode"] == 3
```

`ComputerWait` and `ComputerBash` raise their HTTP deadline per call to cover however long they were asked to block, so a wait past the default 30s is not aborted client-side.

## Screenshots

`GetScreenshot` posts to `/v1/screen/{serialno}`, which is a **cross-family** route: the server dispatches it by device type, so it works for mobile, browser and computer serials alike. The server decides the format (JPEG) regardless of any file extension you choose.

```go
data, err := client.GetScreenshot()
```

`DownloadScreenshot` is an SDK-only extra with no CLI equivalent — it hits `GET /v1/screenshot/{serialno}` instead.

## Error Handling

Two failure layers are surfaced, and the second is easy to miss:

| Layer | Type | Meaning |
|-------|------|---------|
| Gateway | `AuthenticationError`, `DeviceNotFoundError`, `ValidationError`, `Error` | The HTTP status was non-2xx. `StatusCode` carries it. |
| Business | `BusinessError` | HTTP 200, but the envelope carried a non-2xx `code`. `Code` carries it. |

```go
import "errors"

_, err := client.BrowserClick(serialno, "#missing")
if err != nil {
    var bizErr *devicebase.BusinessError
    var authErr *devicebase.AuthenticationError
    var notFoundErr *devicebase.DeviceNotFoundError
    var valErr *devicebase.ValidationError

    switch {
    case errors.As(err, &bizErr):
        // The action failed in the driver, e.g. no element matched.
        fmt.Println("action failed:", bizErr.Code, bizErr.Body)
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

The control API reports action failures **inside an otherwise successful response** — a selector that matches nothing comes back as HTTP 200 with `{"code":502,...}`. Guarding on the HTTP status alone would treat that as success, which is why the envelope is inspected too.

## Testing

```bash
go test ./...              # hermetic: httptest servers, no network
go test -cover ./...
```

A live smoke suite is included and skipped by default. It only performs read-only actions, and discovers serials from the server rather than hard-coding them:

```bash
DEVICEBASE_LIVE_TEST=1 \
DEVICEBASE_API_KEY=<key> \
DEVICEBASE_BASE_URL=http://127.0.0.1:8000 \
go test -run TestLive -v ./...
```

## License

MIT
