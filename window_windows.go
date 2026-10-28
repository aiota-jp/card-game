//go:build windows

package main

import (
	"cardlounge/internal/windowstate"
	"context"
	"golang.org/x/sys/windows"
	"os"
	"unsafe"
)

var winUser = windows.NewLazySystemDLL("user32.dll")
var winShcore = windows.NewLazySystemDLL("shcore.dll")

type winRect struct{ Left, Top, Right, Bottom int32 }
type winPoint struct{ X, Y int32 }
type monitorInfo struct {
	Size          uint32
	Monitor, Work winRect
	Flags         uint32
	Device        [32]uint16
}

func rect(r winRect) windowstate.Rect {
	return windowstate.Rect{X: int(r.Left), Y: int(r.Top), Width: int(r.Right - r.Left), Height: int(r.Bottom - r.Top)}
}
func platformPrepare() {
	proc := winUser.NewProc("SetProcessDpiAwarenessContext")
	if proc.Find() == nil {
		proc.Call(^uintptr(3))
	} // PER_MONITOR_AWARE_V2; already-set manifests take precedence.
}
func platformHandle() uintptr {
	var handle uintptr
	callback := windows.NewCallback(func(hwnd, unused uintptr) uintptr {
		var pid uint32
		winUser.NewProc("GetWindowThreadProcessId").Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid != uint32(os.Getpid()) {
			return 1
		}
		var name [128]uint16
		winUser.NewProc("GetClassNameW").Call(hwnd, uintptr(unsafe.Pointer(&name[0])), 128)
		if windows.UTF16ToString(name[:]) == "CardGameLoungeWindow" {
			handle = hwnd
			return 0
		}
		return 1
	})
	winUser.NewProc("EnumWindows").Call(callback, 0)
	return handle
}
func monitorDisplay(handle uintptr) (windowstate.Display, bool) {
	info := monitorInfo{}
	info.Size = uint32(unsafe.Sizeof(info))
	ok, _, _ := winUser.NewProc("GetMonitorInfoW").Call(handle, uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return windowstate.Display{}, false
	}
	dpi := uint32(96)
	var y uint32
	proc := winShcore.NewProc("GetDpiForMonitor")
	if proc.Find() == nil {
		result, _, _ := proc.Call(handle, 0, uintptr(unsafe.Pointer(&dpi)), uintptr(unsafe.Pointer(&y)))
		if result != 0 || dpi == 0 {
			dpi = 96
		}
	}
	return windowstate.Display{ID: windows.UTF16ToString(info.Device[:]), Work: rect(info.Work), Scale: float64(dpi) / 96}, true
}
func platformDisplays(_ context.Context, hwnd uintptr) ([]windowstate.Display, int) {
	screens := []windowstate.Display{}
	preferred := 0
	current, _, _ := winUser.NewProc("MonitorFromWindow").Call(hwnd, 2)
	var pointer winPoint
	if ok, _, _ := winUser.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&pointer))); ok != 0 {
		point := uintptr(uint64(uint32(pointer.X)) | uint64(uint32(pointer.Y))<<32)
		current, _, _ = winUser.NewProc("MonitorFromPoint").Call(point, 2)
	}
	callback := windows.NewCallback(func(monitor, hdc, area, data uintptr) uintptr {
		if d, ok := monitorDisplay(monitor); ok {
			if monitor == current {
				preferred = len(screens)
			}
			screens = append(screens, d)
		}
		return 1
	})
	winUser.NewProc("EnumDisplayMonitors").Call(0, 0, callback, 0)
	return screens, preferred
}
func platformCapture(_ context.Context, hwnd uintptr) (windowCapture, bool) {
	if hwnd == 0 {
		return windowCapture{}, false
	}
	var bounds winRect
	ok, _, _ := winUser.NewProc("GetWindowRect").Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
	if ok == 0 {
		return windowCapture{}, false
	}
	maximized, _, _ := winUser.NewProc("IsZoomed").Call(hwnd)
	minimized, _, _ := winUser.NewProc("IsIconic").Call(hwnd)
	monitor, _, _ := winUser.NewProc("MonitorFromWindow").Call(hwnd, 2)
	display, valid := monitorDisplay(monitor)
	return windowCapture{rect(bounds), display, maximized != 0, minimized != 0}, valid
}
func platformApply(_ context.Context, hwnd uintptr, r windowstate.Rect) {
	if hwnd == 0 {
		return
	}
	winUser.NewProc("SetWindowPos").Call(hwnd, 0, uintptr(r.X), uintptr(r.Y), uintptr(r.Width), uintptr(r.Height), 0x14)
}
