//go:build windows

package main

import (
	"cardlounge/internal/windowstate"
	"context"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"
)

func TestNativeWindowWorkArea(t *testing.T) {
	if os.Getenv("CARDLOUNGE_WINDOW_NATIVE_TEST") != "1" {
		t.Skip("optional native window geometry check")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	platformPrepare()
	screens, index := platformDisplays(context.Background(), 0)
	if len(screens) == 0 {
		t.Fatal("no Windows monitor work area returned")
	}
	desired := windowstate.Fit(windowstate.State{}, screens, index)
	// Create only our own hidden test window. Never touch another application.
	class, _ := windows.UTF16PtrFromString("CardGameLoungeWindow")
	instance, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW").Call(0)
	type windowClass struct {
		Style                              uint32
		Procedure                          uintptr
		ClassExtra, WindowExtra            int32
		Instance, Icon, Cursor, Background uintptr
		Menu, Name                         *uint16
	}
	procedure := windows.NewCallback(func(hwnd, message, wparam, lparam uintptr) uintptr {
		result, _, _ := winUser.NewProc("DefWindowProcW").Call(hwnd, message, wparam, lparam)
		return result
	})
	definition := windowClass{Procedure: procedure, Instance: instance, Name: class}
	registered, _, registerErr := winUser.NewProc("RegisterClassW").Call(uintptr(unsafe.Pointer(&definition)))
	if registered == 0 {
		t.Fatalf("RegisterClassW: %v", registerErr)
	}
	defer winUser.NewProc("UnregisterClassW").Call(uintptr(unsafe.Pointer(class)), instance)
	title, _ := windows.UTF16PtrFromString("PokerLab geometry test")
	hwnd, _, err := winUser.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)), 0x00cf0000, 0, 0, 320, 240, 0, 0, instance, 0)
	if hwnd == 0 {
		t.Fatalf("CreateWindowExW: %v", err)
	}
	defer winUser.NewProc("DestroyWindow").Call(hwnd)
	if found := platformHandle(); found != hwnd {
		t.Fatalf("own application window not found: %d want %d", found, hwnd)
	}
	platformApply(context.Background(), hwnd, desired.Normal)
	got, ok := platformCapture(context.Background(), hwnd)
	if !ok || got.Bounds != desired.Normal {
		t.Fatalf("native bounds mismatch: %+v want %+v", got, desired)
	}
	settings := filepath.Join(t.TempDir(), "window.json")
	controller := &windowController{path: settings, handle: hwnd, state: desired}
	controller.beforeClose(context.Background())
	if saved := windowstate.Load(settings); saved.Normal != desired.Normal || saved.DisplayID != desired.DisplayID {
		t.Fatalf("native close did not save window: %+v", saved)
	}
	t.Logf("native monitor: %s work=%+v scale=%.2f window=%+v", got.Display.ID, got.Display.Work, got.Display.Scale, got.Bounds)
}
