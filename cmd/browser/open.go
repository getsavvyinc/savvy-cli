package browser

import (
	"os/exec"
	"runtime"
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
