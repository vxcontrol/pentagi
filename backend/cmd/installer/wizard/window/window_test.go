package window

import (
	"testing"
)

func TestNew(t *testing.T) {
	w := New()

	width, height := w.GetWindowSize()
	if width != 80 {
		t.Errorf("expected default width 80, got %d", width)
	}
	if height != 24 {
		t.Errorf("expected default height 24, got %d", height)
	}

	// verify all margins start at zero
	contentWidth, contentHeight := w.GetContentSize()
	if contentWidth != 80 {
		t.Errorf("expected default content width 80, got %d", contentWidth)
	}
	if contentHeight != 24 {
		t.Errorf("expected default content height 24, got %d", contentHeight)
	}

	if !w.IsShowHeader() {
		t.Error("expected header to show with default dimensions")
	}
}

func TestSetWindowSize(t *testing.T) {
	w := New()

	w.SetWindowSize(100, 50)

	if w.GetWindowWidth() != 100 {
		t.Errorf("expected width 100, got %d", w.GetWindowWidth())
	}
	if w.GetWindowHeight() != 50 {
		t.Errorf("expected height 50, got %d", w.GetWindowHeight())
	}

	width, height := w.GetWindowSize()
	if width != 100 || height != 50 {
		t.Errorf("expected size (100, 50), got (%d, %d)", width, height)
	}
}

func TestSetHeaderHeight(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 24)

	w.SetHeaderHeight(3)

	_, contentHeight := w.GetContentSize()
	expectedHeight := 24 - 3 // window height minus header
	if contentHeight != expectedHeight {
		t.Errorf("expected content height %d, got %d", expectedHeight, contentHeight)
	}
}

func TestSetFooterHeight(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 24)

	w.SetFooterHeight(2)

	_, contentHeight := w.GetContentSize()
	expectedHeight := 24 - 2 // window height minus footer
	if contentHeight != expectedHeight {
		t.Errorf("expected content height %d, got %d", expectedHeight, contentHeight)
	}
}

func TestSetLeftSideBarWidth(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 24)

	w.SetLeftSideBarWidth(10)

	contentWidth, _ := w.GetContentSize()
	expectedWidth := 80 - 10 // window width minus left sidebar
	if contentWidth != expectedWidth {
		t.Errorf("expected content width %d, got %d", expectedWidth, contentWidth)
	}
}

func TestSetRightSideBarWidth(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 24)

	w.SetRightSideBarWidth(15)

	contentWidth, _ := w.GetContentSize()
	expectedWidth := 80 - 15 // window width minus right sidebar
	if contentWidth != expectedWidth {
		t.Errorf("expected content width %d, got %d", expectedWidth, contentWidth)
	}
}

func TestGetContentSizeWithAllMargins(t *testing.T) {
	w := New()
	w.SetWindowSize(100, 50)
	w.SetHeaderHeight(5)
	w.SetFooterHeight(3)
	w.SetLeftSideBarWidth(12)
	w.SetRightSideBarWidth(8)

	contentWidth, contentHeight := w.GetContentSize()

	expectedWidth := 100 - 12 - 8 // 80
	expectedHeight := 50 - 5 - 3  // 42

	if contentWidth != expectedWidth {
		t.Errorf("expected content width %d, got %d", expectedWidth, contentWidth)
	}
	if contentHeight != expectedHeight {
		t.Errorf("expected content height %d, got %d", expectedHeight, contentHeight)
	}
}

func TestGetContentWidth(t *testing.T) {
	w := New()
	w.SetWindowSize(120, 40)
	w.SetLeftSideBarWidth(20)
	w.SetRightSideBarWidth(30)

	contentWidth := w.GetContentWidth()
	expected := 120 - 20 - 30 // 70

	if contentWidth != expected {
		t.Errorf("expected content width %d, got %d", expected, contentWidth)
	}
}

func TestGetContentHeight(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 60)
	w.SetHeaderHeight(8)
	w.SetFooterHeight(4)

	contentHeight := w.GetContentHeight()
	expected := 60 - 8 - 4 // 48

	if contentHeight != expected {
		t.Errorf("expected content height %d, got %d", expected, contentHeight)
	}
}

func TestIsShowHeaderTrue(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 20)
	w.SetHeaderHeight(5)
	w.SetFooterHeight(3)

	// available content height: 20 - 5 - 3 = 12 >= MinContentHeight (2)
	if !w.IsShowHeader() {
		t.Error("expected header to show when sufficient space available")
	}
}

func TestIsShowHeaderFalse(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 10)
	w.SetHeaderHeight(5)
	w.SetFooterHeight(4)

	// available content height: 10 - 5 - 4 = 1 < MinContentHeight (2)
	if w.IsShowHeader() {
		t.Error("expected header to hide when insufficient space")
	}
}

func TestIsShowHeaderBoundary(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 9)
	w.SetHeaderHeight(5)
	w.SetFooterHeight(2)

	// available content height: 9 - 5 - 2 = 2 == MinContentHeight (2)
	if !w.IsShowHeader() {
		t.Error("expected header to show at boundary condition")
	}
}

func TestGetContentSizeWithHiddenHeader(t *testing.T) {
	w := New()
	w.SetWindowSize(80, 8)
	w.SetHeaderHeight(5)
	w.SetFooterHeight(4)

	// header should be hidden, so content height = window height - footer only
	_, contentHeight := w.GetContentSize()
	expected := 8 - 4 // 4 (header ignored when hidden)

	if contentHeight != expected {
		t.Errorf("expected content height %d with hidden header, got %d", expected, contentHeight)
	}
}

func TestNegativeContentDimensions(t *testing.T) {
	w := New()
	w.SetWindowSize(20, 15)
	w.SetLeftSideBarWidth(15)
	w.SetRightSideBarWidth(10)
	w.SetHeaderHeight(10)
	w.SetFooterHeight(8)

	contentWidth, contentHeight := w.GetContentSize()

	// content dimensions should not go below zero
	if contentWidth < 0 {
		t.Errorf("expected non-negative content width, got %d", contentWidth)
	}
	if contentHeight < 0 {
		t.Errorf("expected non-negative content height, got %d", contentHeight)
	}

	// verify they are actually zero in this case
	if contentWidth != 0 {
		t.Errorf("expected zero content width with excessive margins, got %d", contentWidth)
	}
}

func TestZeroDimensions(t *testing.T) {
	w := New()
	w.SetWindowSize(0, 0)

	width, height := w.GetWindowSize()
	if width != 0 || height != 0 {
		t.Errorf("expected zero window size, got (%d, %d)", width, height)
	}

	contentWidth, contentHeight := w.GetContentSize()
	if contentWidth != 0 || contentHeight != 0 {
		t.Errorf("expected zero content size, got (%d, %d)", contentWidth, contentHeight)
	}

	if w.IsShowHeader() {
		t.Error("expected header to be hidden with zero window size")
	}
}

func TestLargeMargins(t *testing.T) {
	w := New()
	w.SetWindowSize(50, 30)
	w.SetLeftSideBarWidth(25)
	w.SetRightSideBarWidth(30) // total sidebars exceed window width
	w.SetHeaderHeight(15)
	w.SetFooterHeight(20) // total margins exceed window height

	contentWidth, contentHeight := w.GetContentSize()

	// max() function should prevent negative values
	if contentWidth != 0 {
		t.Errorf("expected zero content width with excessive margins, got %d", contentWidth)
	}

	// when header is hidden due to insufficient space, height = window - footer only
	// 30 >= 15 + 20 + 2 is false, so header hidden, height = max(30 - 20, 0) = 10
	expectedHeight := 10
	if contentHeight != expectedHeight {
		t.Errorf("expected content height %d with hidden header, got %d", expectedHeight, contentHeight)
	}
}

func TestComplexLayoutScenario(t *testing.T) {
	w := New()

	// simulate realistic terminal app layout
	w.SetWindowSize(120, 40)
	w.SetHeaderHeight(3)       // title bar
	w.SetFooterHeight(2)       // status bar
	w.SetLeftSideBarWidth(20)  // navigation menu
	w.SetRightSideBarWidth(15) // info panel

	contentWidth := w.GetContentWidth()
	contentHeight := w.GetContentHeight()

	expectedWidth := 120 - 20 - 15 // 85
	expectedHeight := 40 - 3 - 2   // 35

	if contentWidth != expectedWidth {
		t.Errorf("expected content width %d in complex layout, got %d", expectedWidth, contentWidth)
	}
	if contentHeight != expectedHeight {
		t.Errorf("expected content height %d in complex layout, got %d", expectedHeight, contentHeight)
	}

	if !w.IsShowHeader() {
		t.Error("expected header to show in complex layout")
	}
}

func TestWindowResizing(t *testing.T) {
	w := New()
	w.SetHeaderHeight(4)
	w.SetFooterHeight(2)
	w.SetLeftSideBarWidth(10)

	// test multiple resize operations
	sizes := []struct{ width, height int }{
		{80, 24},
		{120, 40},
		{60, 20},
		{200, 60},
	}

	for _, size := range sizes {
		w.SetWindowSize(size.width, size.height)

		if w.GetWindowWidth() != size.width {
			t.Errorf("expected window width %d, got %d", size.width, w.GetWindowWidth())
		}
		if w.GetWindowHeight() != size.height {
			t.Errorf("expected window height %d, got %d", size.height, w.GetWindowHeight())
		}

		// verify content size updates correctly
		expectedContentWidth := size.width - 10      // left sidebar
		expectedContentHeight := size.height - 4 - 2 // header + footer

		if expectedContentWidth < 0 {
			expectedContentWidth = 0
		}
		if expectedContentHeight < 0 {
			expectedContentHeight = 0
		}

		contentWidth := w.GetContentWidth()
		contentHeight := w.GetContentHeight()

		if contentWidth != expectedContentWidth {
			t.Errorf("size %dx%d: expected content width %d, got %d",
				size.width, size.height, expectedContentWidth, contentWidth)
		}
		if contentHeight != expectedContentHeight {
			t.Errorf("size %dx%d: expected content height %d, got %d",
				size.width, size.height, expectedContentHeight, contentHeight)
		}
	}
}
