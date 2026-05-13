package devicebase_test

import (
	"fmt"
	"os"

	devicebase "github.com/devicebase/go-devicebase"
)

func ExampleNewClient() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerial("device-serial-number"),
	)
	_ = client
}

func ExampleClient_GetDeviceInfo() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey(os.Getenv("DEVICEBASE_API_KEY")),
		devicebase.WithSerial("device123"),
		devicebase.WithBaseURL("https://api.devicebase.cn"),
	)

	info, err := client.GetDeviceInfo()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("device:", info.Serial)
}

func ExampleClient_tap() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerial("device123"),
	)

	result, err := client.Tap(100, 200)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("success:", result.Success)
}

func ExampleClient_launchApp() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerial("device123"),
	)

	result, err := client.LaunchApp("com.tencent.mm")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("success:", result.Success)
}

func ExampleClient_swipe() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerial("device123"),
	)

	result, err := client.Swipe(0, 500, 500, 500)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("success:", result.Success)
}

func ExampleClient_getScreenshot() {
	client := devicebase.NewClient(
		devicebase.WithAPIKey("your-api-key"),
		devicebase.WithSerial("device123"),
	)

	screenshot, err := client.GetScreenshot()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("screenshot size: %d bytes\n", len(screenshot))
}
