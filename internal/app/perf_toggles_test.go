package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"netease-music-mygo/internal/player"
)

// Turning off the start animation skips it while the rest of the motion stays full.
func TestNoSplashSkipsStart(t *testing.T) {
	a := playerBarLayoutApp()
	a.splashed, a.motion = false, motionFull
	a.settings.NoSplash = true
	tt := ui.NewTester(a.view, 1200, 780)
	tt.Frame()
	if !a.splashed {
		t.Fatal("the start animation ran with no-splash on")
	}
	tt.Advance(time.Second)
	if tt.FramePending() {
		t.Error("no-splash still asks for a frame")
	}
}

// Hiding the tray removes its icon, whatever the platform's usual choice.
func TestHideTray(t *testing.T) {
	if showsTray(settings{HideTray: true}) {
		t.Fatal("the tray shows while hidden")
	}
	if !showsTray(settings{}) {
		t.Fatal("the tray is gone by default")
	}
}

// The cover limit keeps the media cover at 600 even when 原图 was chosen.
func TestNoOriginalCover(t *testing.T) {
	old := player.CoverSize.Load()
	defer player.CoverSize.Store(old)
	a := newTestApp()
	a.settings.MediaCover = -1
	a.settings.NoOriginalCover = true
	a.useMediaCover(false)
	if got := player.CoverSize.Load(); got != 600 {
		t.Fatalf("media cover is %d, want 600", got)
	}
}
