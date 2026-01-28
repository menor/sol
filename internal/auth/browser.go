package auth

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openBrowser opens a URL in the user's default browser.
// It uses OS-specific commands:
//   - macOS: open
//   - Linux: xdg-open
//   - Windows: start (via cmd)
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		// Windows requires using cmd /c start
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
