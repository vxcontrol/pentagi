package window

const MinContentHeight = 2

type Window interface {
	SetWindowSize(width, height int)
	SetHeaderHeight(height int)
	SetFooterHeight(height int)
	SetLeftSideBarWidth(width int)
	SetRightSideBarWidth(width int)
	GetWindowSize() (int, int)
	GetWindowWidth() int
	GetWindowHeight() int
	GetContentSize() (int, int)
	GetContentWidth() int
	GetContentHeight() int
	IsShowHeader() bool
}

// window manages terminal window dimensions and content area calculations
type window struct {
	// Total terminal window dimensions
	windowWidth  int
	windowHeight int

	// Margins that reduce content area
	headerHeight      int
	footerHeight      int
	leftSideBarWidth  int
	rightSideBarWidth int
}

// New creates a new window manager with default dimensions
func New() Window {
	return &window{
		windowWidth:       80, // default terminal width
		windowHeight:      24, // default terminal height
		headerHeight:      0,
		footerHeight:      0,
		leftSideBarWidth:  0,
		rightSideBarWidth: 0,
	}
}

// SetWindowSize updates the total terminal window dimensions
func (w *window) SetWindowSize(width, height int) {
	w.windowWidth = width
	w.windowHeight = height
}

// margin setters
func (w *window) SetHeaderHeight(height int) {
	w.headerHeight = height
}

func (w *window) SetFooterHeight(height int) {
	w.footerHeight = height
}

func (w *window) SetLeftSideBarWidth(width int) {
	w.leftSideBarWidth = width
}

func (w *window) SetRightSideBarWidth(width int) {
	w.rightSideBarWidth = width
}

// window size getters
func (w *window) GetWindowSize() (int, int) {
	return w.windowWidth, w.windowHeight
}

func (w *window) GetWindowWidth() int {
	return w.windowWidth
}

func (w *window) GetWindowHeight() int {
	return w.windowHeight
}

// content size getters (window size minus margins)
func (w *window) GetContentSize() (int, int) {
	contentWidth := max(w.windowWidth-w.leftSideBarWidth-w.rightSideBarWidth, 0)
	contentHeight := max(w.windowHeight-w.headerHeight-w.footerHeight, 0)
	if !w.IsShowHeader() {
		contentHeight = max(w.windowHeight-w.footerHeight, 0)
	}

	return contentWidth, contentHeight
}

func (w *window) GetContentWidth() int {
	width, _ := w.GetContentSize()
	return width
}

func (w *window) GetContentHeight() int {
	_, height := w.GetContentSize()
	return height
}

func (w *window) IsShowHeader() bool {
	return w.windowHeight >= w.headerHeight+w.footerHeight+MinContentHeight
}
