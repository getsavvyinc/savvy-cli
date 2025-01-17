package browser

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/getsavvyinc/savvy-cli/display"
)

func OpenCmd(url string) *exec.Cmd {
	if strings.HasPrefix(url, "chrome-extension://") {
		switch runtime.GOOS {
		case "linux":
			return exec.Command("google-chrome", url)
		case "darwin":
			return exec.Command("open", "-a", "Google Chrome", url)
		default:
			return nil
		}
	}
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url)
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		return exec.Command("open", url)
	default:
	}
	return nil
}

func Open(url string) {
	display.Info("Opening your default browser to " + url)

	cmd := OpenCmd(url)
	runOpenCmd(cmd, url)
}

func runOpenCmd(cmd *exec.Cmd, target string) {
	var browserOpenError = fmt.Errorf("Please visit %s in your browser", target)
	if cmd == nil {
		display.Error(browserOpenError)
		return
	}

	if err := cmd.Start(); err != nil {
		display.Error(browserOpenError)
		return
	}
}

func OpenExtensionSidePanel() {
	extensionURL := "chrome-extension://jocphfjphhfbdccjfjjnbcnejmbojjlh/side-panel/index.html"
	display.Info("Opening Savvy's extension on Chrome...")
	cmd := OpenCmd(extensionURL)
	runOpenCmd(cmd, extensionURL)
}
