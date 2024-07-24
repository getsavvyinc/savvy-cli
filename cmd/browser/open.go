package browser

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/getsavvyinc/savvy-cli/display"
)

func OpenCmd(url string) *exec.Cmd {
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
	var browserOpenError = fmt.Errorf("couldn't open your default browser. Please visit %s in your browser", url)

	cmd := OpenCmd(url)
	if cmd == nil {
		display.Error(browserOpenError)
		return
	}

	if err := cmd.Start(); err != nil {
		display.Error(browserOpenError)
		return
	}
}
