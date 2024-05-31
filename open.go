package serve

import (
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser opens the specified URL in the user's default browser
func OpenBrowser(url string) error {
	var cmd []string
	switch {
	case runtime.GOOS == "windows" || isWSL():
		cmd = []string{"cmd.exe", "/c", "start"}
	case runtime.GOOS == "darwin":
		cmd = []string{"open"}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = []string{"xdg-open"}
	}
	return exec.Command(cmd[0], append(cmd[1:], url)...).Start()
}

// isWSL checks if the Go program is running inside Windows Subsystem for Linux
func isWSL() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}
