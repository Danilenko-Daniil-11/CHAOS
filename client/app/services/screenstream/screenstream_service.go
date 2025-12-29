package screenstream

import (
	"github.com/tiagorlampert/CHAOS/client/app/services"
)

type Service struct {
	Screenshot services.Screenshot
}

func NewService(screenshot services.Screenshot) services.ScreenStream {
	return &Service{Screenshot: screenshot}
}

func (s Service) StartScreenStream(duration int) ([]byte, error) {
	// Placeholder implementation for real-time screen streaming
	// In a full implementation, this would capture multiple screenshots
	// and stream them over WebSocket
	// Currently returns a single screenshot
	return s.Screenshot.TakeScreenshot()
}
