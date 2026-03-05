//go:build windows

package player

import (
	"log/slog"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WM_HOTKEY            = 0x0312
	MOD_NOREPEAT         = 0x4000
	VK_MEDIA_PLAY_PAUSE  = 0xB3
	VK_MEDIA_NEXT_TRACK  = 0xB0
	VK_MEDIA_PREV_TRACK  = 0xB1
)

var (
	user32             = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey = user32.NewProc("RegisterHotKey")
	procGetMessage     = user32.NewProc("GetMessageW")
)

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

func StartMediaHotkeyListener() {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		const (
			ID_PLAY_PAUSE = 1
			ID_NEXT       = 2
			ID_PREV       = 3
		)

		if !registerHotKey(ID_PLAY_PAUSE, MOD_NOREPEAT, VK_MEDIA_PLAY_PAUSE) {
			slog.Warn("failed to register play/pause hotkey")
		} else {
			slog.Info("registered play/pause hotkey")
		}
		if !registerHotKey(ID_NEXT, MOD_NOREPEAT, VK_MEDIA_NEXT_TRACK) {
			slog.Warn("failed to register next track hotkey")
		} else {
			slog.Info("registered next track hotkey")
		}
		if !registerHotKey(ID_PREV, MOD_NOREPEAT, VK_MEDIA_PREV_TRACK) {
			slog.Warn("failed to register prev track hotkey")
		} else {
			slog.Info("registered prev track hotkey")
		}

		slog.Info("Media hotkey listener started")

		var msg MSG
		for {
			ret, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(&msg)),
				0, 0, 0,
			)
			if ret == 0 {
				break
			}

			if msg.Message == WM_HOTKEY {
				switch int(msg.WParam) {
				case ID_PLAY_PAUSE:
					slog.Info("media hotkey: play/pause")
					SendMediaEvent(MediaEvent{Action: "pause"})
				case ID_NEXT:
					slog.Info("media hotkey: next")
					SendMediaEvent(MediaEvent{Action: "next"})
				case ID_PREV:
					slog.Info("media hotkey: previous")
					SendMediaEvent(MediaEvent{Action: "previous"})
				}
			}
		}
	}()
}

func registerHotKey(id int, modifiers uint32, vk uint32) bool {
	ret, _, _ := procRegisterHotKey.Call(
		0,
		uintptr(id),
		uintptr(modifiers),
		uintptr(vk),
	)
	return ret != 0
}
