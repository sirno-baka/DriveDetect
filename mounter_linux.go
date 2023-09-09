package usbdrivedetect

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Detect returns a list of file paths pointing to the root folder of
// USB storage devices connected to the system.
func DetectAndMount() ([]string, error) {
	var drives []string
	driveMap := make(map[string]string)
	unmountedStorages := []string{}
	udiskPattern := regexp.MustCompile("^(\\S+)\\s+\\S+\\s+\\S+\\s+\\S+\\s+\\S+\\s+(part) (.*$)")

	out, err := exec.Command("lsblk", "--list").Output()

	if err != nil {
		log.Printf("Error calling udisk: %s", err)
	}

	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		line := s.Text()
		if udiskPattern.MatchString(line) {
			device := udiskPattern.FindStringSubmatch(line)[1]
			if ok := isUSBStorage(device); ok {
				mountPoint := udiskPattern.FindStringSubmatch(line)[3]
				if mountPoint == "" {
					unmountedStorages = append(unmountedStorages, device)
				} else {
					driveMap[device] = mountPoint
				}

			}

		}
	}

	for _, mountPoint := range driveMap {
		file, err := os.Open(mountPoint)
		if err == nil {
			drives = append(drives, mountPoint)
		}
		file.Close()
	}
	if len(drives) != 0 {
		return drives, nil
	}

	// trying mount
	udiskctlMountedOutPattern := regexp.MustCompile("at (.*)")

	for _, device := range unmountedStorages {
		out, err := exec.Command("udisksctl", "mount", "-b", "/dev/"+device).Output()
		if err != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				log.Printf("Error calling udisksctl mount -b : %s", err.Stderr)
			case error:
				log.Printf("Error calling udisksctl mount -b : %s", err)
			}

			continue
		}
		if udiskctlMountedOutPattern.MatchString(string(out)) {
			mountPoint := udiskctlMountedOutPattern.FindStringSubmatch(string(out))[1]
			file, err := os.Open(mountPoint)
			if err == nil {
				drives = append(drives, mountPoint)
				file.Close()
			} else {
				exec.Command("udisksctl", "unmount", "-b", "/dev/"+device)
			}
		}
	}

	return drives, nil
}

func isUSBStorage(device string) bool {
	deviceVerifier := "ID_USB_DRIVER=usb-storage"
	cmd := "udevadm"
	args := []string{"info", "-q", "property", "-n", device}
	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		log.Printf("Error checking device %s: %s", device, err)
		return false
	}

	if strings.Contains(string(out), deviceVerifier) {
		return true
	}

	return false
}
