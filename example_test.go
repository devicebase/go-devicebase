package devicebase_test

import (
	"fmt"
	"os"

	devicebase "github.com/devicebase/go-devicebase"
)

func ExampleNewClient() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerialno("db-mttul4i41di8"),
	)
	_ = client
}

// ExampleClient_ListDevices discovers a serialno — the first step for every
// platform, and the only method that works without a bound serialno.
func ExampleClient_ListDevices() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey(os.Getenv("DEVICEBASE_API_KEY")),
	)

	browsers, err := client.ListDevices(devicebase.ListDevicesRequest{Type: "browser"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, device := range browsers {
		fmt.Println(device.Serialno, device.State)
	}
}

func ExampleClient_GetDeviceInfo() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey(os.Getenv("DEVICEBASE_API_KEY")),
		devicebase.WithSerialno("device123"),
	)

	info, err := client.GetDeviceInfo()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("device:", info.Serialno)
}

func ExampleClient_Tap() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerialno("device123"),
	)

	result, err := client.Tap(100, 200)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("success:", result.Success)
}

func ExampleClient_LaunchApp() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerialno("device123"),
	)

	result, err := client.LaunchApp("com.tencent.mm")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("success:", result.Success)
}

func ExampleClient_Swipe() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerialno("device123"),
	)

	result, err := client.Swipe(0, 500, 500, 500)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("success:", result.Success)
}

func ExampleClient_GetScreenshot() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerialno("device123"),
	)

	screenshot, err := client.GetScreenshot()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("screenshot size: %d bytes\n", len(screenshot))
}

// ExampleClient_BrowserNavigate drives a registered browser over CDP. The
// serialno comes from ListDevices with Type "browser".
func ExampleClient_BrowserNavigate() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
	)

	serialno := "db-mtsi49bf0mqb"
	if _, err := client.BrowserNavigate(serialno, "https://example.com"); err != nil {
		fmt.Println("error:", err)
		return
	}
	if _, err := client.BrowserFill(serialno, "#search", "devicebase"); err != nil {
		fmt.Println("error:", err)
		return
	}

	text, err := client.BrowserText(serialno, "#search")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("value:", text.Data)
}

// ExampleClient_ComputerBash runs a command on the host that owns the desktop
// device. The command's own exit status arrives in data.exitCode rather than as
// an error, because the API call itself succeeded.
func ExampleClient_ComputerBash() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
	)

	result, err := client.ComputerBash("db-mtthisv311f1", "echo hello", 30)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("exit code:", result.Data["exitCode"])
}

// ExampleClient_ComputerWait takes milliseconds and widens the HTTP deadline to
// cover the wait, so a sleep longer than the default 30s is not aborted.
func ExampleClient_ComputerWait() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
	)

	if _, err := client.ComputerWait("db-mtthisv311f1", 2000); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("waited 2s")
}
