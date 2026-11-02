package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gotrans/internal/x11win"
)

const (
	floatWindowTitle  = "划词译文 · Gotrans"
	floatMinWidth     = float32(220)
	floatMinHeight    = float32(120)
	floatInitW        = float32(420)
	floatInitH        = float32(260)
	edgeThickness     = float32(4)
	floatCornerRadius = float32(12) // 悬浮窗四角圆角半径（逻辑像素，按缩放换算成物理像素）
)

// 拖拽类型：整块移动，或四条边的缩放。
const (
	dragMove = iota
	dragResizeTop
	dragResizeBottom
	dragResizeLeft
	dragResizeRight
)

// selectionResultEntry 是主窗口的划词译文区：一个多行输入框，样式与其它原文/译文输入框一致，
// 保留文本选择/编辑能力，右键菜单在复制/剪切/粘贴/全选之外附加一个自定义项（「仅显示划词译文窗口」）。
// 注意：必须使用值嵌入（widget.Entry 而非 *widget.Entry），使自定义类型与内嵌 Entry 地址一致，
// 否则渲染器缓存会分裂，导致窗口缩放时译文区不跟随、内容不铺满。
type selectionResultEntry struct {
	widget.Entry
	extraLabel string
	onExtra    func()
}

func newSelectionResultEntry() *selectionResultEntry {
	e := &selectionResultEntry{}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.TextStyle = fyne.TextStyle{Bold: true}
	e.ExtendBaseWidget(e)
	return e
}

// SetExtra 设置右键菜单末尾附加的条目及其动作。
func (e *selectionResultEntry) SetExtra(label string, fn func()) {
	e.extraLabel = label
	e.onExtra = fn
}

// TappedSecondary 复刻内置菜单（复制/剪切/粘贴/全选）并附加自定义项。
func (e *selectionResultEntry) TappedSecondary(pe *fyne.PointEvent) {
	app := fyne.CurrentApp()
	c := app.Driver().CanvasForObject(e)
	if c == nil {
		return
	}
	c.Focus(e)

	clipboard := app.Clipboard()
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("复制", func() {
			e.TypedShortcut(&fyne.ShortcutCopy{Clipboard: clipboard})
		}),
		fyne.NewMenuItem("剪切", func() {
			e.TypedShortcut(&fyne.ShortcutCut{Clipboard: clipboard})
		}),
		fyne.NewMenuItem("粘贴", func() {
			e.TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
		}),
		fyne.NewMenuItem("全选", func() {
			e.TypedShortcut(&fyne.ShortcutSelectAll{})
		}),
	}
	if e.extraLabel != "" && e.onExtra != nil {
		items = append(items, fyne.NewMenuItemSeparator(), fyne.NewMenuItem(e.extraLabel, e.onExtra))
	}

	pos := app.Driver().AbsolutePositionForObject(e).Add(pe.Position)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("", items...), c, pos)
}

// floatText 是悬浮窗中的译文显示区：采用与主窗口译文区一致的输入框样式
// （圆角底框 + 边框 + 输入底色），内容为自动换行、垂直滚动的文本；
// 按住整块区域拖动即可移动窗口，右键菜单提供复制/置顶/恢复主窗口。
type floatText struct {
	widget.BaseWidget
	label     *widget.Label
	scroll    *container.Scroll
	onMove    func(dx, dy float32)
	onDragEnd func()
	restore   func()
	isPinned  func() bool
	togglePin func()
}

func newFloatText(text string, onMove func(dx, dy float32), onDragEnd, restore, togglePin func(), isPinned func() bool) *floatText {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignLeading

	// 内边距使文字与圆角边框保持间距，观感与主窗口输入框一致。
	scroll := container.NewVScroll(container.NewPadded(label))
	t := &floatText{label: label, scroll: scroll, onMove: onMove, onDragEnd: onDragEnd,
		restore: restore, isPinned: isPinned, togglePin: togglePin}
	t.ExtendBaseWidget(t)
	return t
}

// SetText 更新显示的译文。
func (t *floatText) SetText(text string) {
	t.label.SetText(text)
}

// Dragged 覆盖文本选择行为，改为回调 onMove 移动悬浮窗。
func (t *floatText) Dragged(ev *fyne.DragEvent) {
	if t.onMove != nil {
		t.onMove(ev.Dragged.DX, ev.Dragged.DY)
	}
}

// DragEnd 通知外部结束拖拽。
func (t *floatText) DragEnd() {
	if t.onDragEnd != nil {
		t.onDragEnd()
	}
}

// TappedSecondary 显示复制/恢复主窗口的右键菜单。
func (t *floatText) TappedSecondary(pe *fyne.PointEvent) {
	app := fyne.CurrentApp()
	c := app.Driver().CanvasForObject(t)
	if c == nil {
		return
	}

	items := []*fyne.MenuItem{
		fyne.NewMenuItem("复制", func() {
			app.Clipboard().SetContent(t.label.Text)
		}),
	}
	if t.togglePin != nil {
		label := "置顶"
		if t.isPinned != nil && t.isPinned() {
			label = "取消置顶"
		}
		items = append(items, fyne.NewMenuItem(label, t.togglePin))
	}
	if t.restore != nil {
		items = append(items, fyne.NewMenuItemSeparator(), fyne.NewMenuItem("恢复主窗口", t.restore))
	}

	pos := app.Driver().AbsolutePositionForObject(t).Add(pe.Position)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("", items...), c, pos)
}

func (t *floatText) CreateRenderer() fyne.WidgetRenderer {
	th := t.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	box := canvas.NewRectangle(th.Color(theme.ColorNameInputBackground, v))
	box.CornerRadius = th.Size(theme.SizeNameInputRadius)
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = th.Size(theme.SizeNameInputBorder)
	border.StrokeColor = th.Color(theme.ColorNameInputBorder, v)
	border.CornerRadius = th.Size(theme.SizeNameInputRadius)
	return &floatTextRenderer{t: t, box: box, border: border, objects: []fyne.CanvasObject{box, border, t.scroll}}
}

type floatTextRenderer struct {
	t       *floatText
	box     *canvas.Rectangle
	border  *canvas.Rectangle
	objects []fyne.CanvasObject
}

// Layout 复刻主窗口 Entry 渲染器的几何：底框内缩 border、边框描边、文本内容上下留出 border。
func (r *floatTextRenderer) Layout(s fyne.Size) {
	th := r.t.Theme()
	borderSize := th.Size(theme.SizeNameInputBorder)

	// 0.5 用于在低 DPI 下向里取整，避免右边框溢出。
	r.border.Resize(fyne.NewSize(s.Width-borderSize-.5, s.Height-borderSize-.5))
	r.border.StrokeWidth = borderSize
	r.border.Move(fyne.NewSquareOffsetPos(borderSize / 2))
	r.box.Resize(s.Subtract(fyne.NewSquareSize(borderSize * 2)))
	r.box.Move(fyne.NewSquareOffsetPos(borderSize))

	inputBorder := th.Size(theme.SizeNameInputBorder)
	r.t.scroll.Resize(s.Subtract(fyne.NewSize(0, inputBorder*2)))
	r.t.scroll.Move(fyne.NewPos(0, inputBorder))
}

func (r *floatTextRenderer) MinSize() fyne.Size {
	return r.t.scroll.MinSize().AddWidthHeight(0, r.t.Theme().Size(theme.SizeNameInputBorder)*2)
}

func (r *floatTextRenderer) Refresh() {
	th := r.t.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	r.box.FillColor = th.Color(theme.ColorNameInputBackground, v)
	r.box.CornerRadius = th.Size(theme.SizeNameInputRadius)
	r.border.StrokeColor = th.Color(theme.ColorNameInputBorder, v)
	r.border.CornerRadius = th.Size(theme.SizeNameInputRadius)
	r.t.scroll.Refresh()
}

func (r *floatTextRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *floatTextRenderer) Destroy() {}

// edgeDir 表示悬浮窗的哪条边。
type edgeDir int

const (
	edgeTop edgeDir = iota
	edgeBottom
	edgeLeft
	edgeRight
)

// edgeStrip 是悬浮窗四周的一条透明边缘：鼠标悬停显示缩放光标，按住拖动调整窗口大小。
// 顶部/左侧边缘在改变尺寸的同时会移动窗口，使对边保持固定。
type edgeStrip struct {
	widget.BaseWidget
	dir       edgeDir
	min       fyne.Size
	onDrag    func(dx, dy float32)
	onDragEnd func()
}

func newEdgeStrip(dir edgeDir, onDrag func(dx, dy float32), onDragEnd func()) *edgeStrip {
	var min fyne.Size
	switch dir {
	case edgeTop, edgeBottom:
		min = fyne.NewSize(0, edgeThickness)
	case edgeLeft, edgeRight:
		min = fyne.NewSize(edgeThickness, 0)
	}
	s := &edgeStrip{dir: dir, min: min, onDrag: onDrag, onDragEnd: onDragEnd}
	s.ExtendBaseWidget(s)
	return s
}

func (s *edgeStrip) Dragged(ev *fyne.DragEvent) {
	if s.onDrag != nil {
		s.onDrag(ev.Dragged.DX, ev.Dragged.DY)
	}
}

func (s *edgeStrip) DragEnd() {
	if s.onDragEnd != nil {
		s.onDragEnd()
	}
}

// Cursor 悬停时显示对应方向的缩放光标，提示可调整大小。
func (s *edgeStrip) Cursor() desktop.Cursor {
	switch s.dir {
	case edgeLeft, edgeRight:
		return desktop.HResizeCursor
	default:
		return desktop.VResizeCursor
	}
}

func (s *edgeStrip) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	return &edgeStripRenderer{bg: bg, min: s.min, objects: []fyne.CanvasObject{bg}}
}

type edgeStripRenderer struct {
	bg      *canvas.Rectangle
	min     fyne.Size
	objects []fyne.CanvasObject
}

func (r *edgeStripRenderer) Layout(s fyne.Size) { r.bg.Resize(s) }
func (r *edgeStripRenderer) MinSize() fyne.Size { return r.min }
func (r *edgeStripRenderer) Refresh()           { r.bg.Refresh() }
func (r *edgeStripRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
func (r *edgeStripRenderer) Destroy() {}

// floatWindow 记录悬浮窗运行状态。拖拽移动/缩放全部基于“拖拽起点锚点 + 指针绝对位移”，
// 避免用逐帧相对增量累积而造成的抖动与鼠标不同步。
type floatWindow struct {
	win   fyne.Window
	text  *floatText
	scale float32
	ready bool

	dragging               bool
	dragKind               int
	anchorX, anchorY       int32
	anchorW, anchorH       int32
	anchorPtrX, anchorPtrY int32

	pinned bool
}

// newFloatWindow 创建并显示一个仅含译文的悬浮窗，隐藏主窗口 parent。
// 窗口/字体透明度由全局主题统一控制（见 app.go 的 opacityTheme），此处无需单独设置。
// onClosed 在悬浮窗被关闭（含恢复主窗口）时回调，用于清除外部持有的引用。
func newFloatWindow(a fyne.App, parent fyne.Window, initialText string, onClosed func()) *floatWindow {
	fs := &floatWindow{}
	fs.win = a.NewWindow(floatWindowTitle)
	fs.scale = fs.win.Canvas().Scale()
	if fs.scale <= 0 {
		fs.scale = 1
	}
	fs.win.Resize(fyne.NewSize(floatInitW, floatInitH))

	restore := func() { fs.win.Close() }
	fs.text = newFloatText(initialText, fs.moveBy, fs.endDrag, restore,
		func() { fs.togglePin() }, func() bool { return fs.pinned })

	// 无边框窗口四周各留一条透明边缘用于缩放，中间内容区整块拖拽移动。
	content := container.NewBorder(
		newEdgeStrip(edgeTop, fs.resizeTop, fs.endDrag),
		newEdgeStrip(edgeBottom, fs.resizeBottom, fs.endDrag),
		newEdgeStrip(edgeLeft, fs.resizeLeft, fs.endDrag),
		newEdgeStrip(edgeRight, fs.resizeRight, fs.endDrag),
		fs.text,
	)
	fs.win.SetContent(content)
	fs.win.SetPadded(false)

	fs.win.SetOnClosed(func() {
		parent.Show()
		if onClosed != nil {
			onClosed()
		}
	})

	fs.win.Show()
	parent.Hide()

	// 窗口映射后移除边框、确认窗口可被定位。
	go fs.applyBorderless()

	return fs
}

// beginDrag 在拖拽首帧记录锚点（窗口几何 + 指针绝对位置）。已在拖拽时返回是否可继续。
func (fs *floatWindow) beginDrag(kind int) bool {
	if fs.dragging {
		return fs.dragKind == kind
	}
	if !fs.ready {
		return false
	}
	x, y, err1 := x11win.PositionByTitle(floatWindowTitle)
	if err1 != nil {
		return false
	}
	w, h, err2 := x11win.SizeByTitle(floatWindowTitle)
	if err2 != nil {
		return false
	}
	px, py, err3 := x11win.PointerPosition()
	if err3 != nil {
		return false
	}
	fs.dragging = true
	fs.dragKind = kind
	fs.anchorX, fs.anchorY = x, y
	fs.anchorW, fs.anchorH = int32(w), int32(h)
	fs.anchorPtrX, fs.anchorPtrY = px, py
	return true
}

func (fs *floatWindow) endDrag() {
	// 缩放会改变窗口尺寸，需重新按新几何裁剪圆角；移动拖拽无需重设。
	if fs.dragKind >= dragResizeTop && fs.dragKind <= dragResizeRight {
		fs.applyRoundedCorners()
	}
	fs.dragging = false
}

// applyRoundedCorners 用 X Shape 扩展把悬浮窗裁剪为圆角，半径按缩放换算成物理像素。
func (fs *floatWindow) applyRoundedCorners() {
	radius := int(floatCornerRadius * fs.scale)
	if radius < 1 {
		radius = 1
	}
	_ = x11win.SetRoundedCornersByTitle(floatWindowTitle, radius)
}

// togglePin 切换置顶状态并请求窗口管理器更新窗口堆叠层级。
func (fs *floatWindow) togglePin() {
	fs.pinned = !fs.pinned
	_ = x11win.SetAlwaysOnTopByTitle(floatWindowTitle, fs.pinned)
}

// pointerDelta 返回当前指针相对拖拽起点的绝对位移（物理像素）。
func (fs *floatWindow) pointerDelta() (dx, dy int32, ok bool) {
	px, py, err := x11win.PointerPosition()
	if err != nil {
		return 0, 0, false
	}
	return px - fs.anchorPtrX, py - fs.anchorPtrY, true
}

func (fs *floatWindow) minW() int32 { return int32(floatMinWidth * fs.scale) }
func (fs *floatWindow) minH() int32 { return int32(floatMinHeight * fs.scale) }

// moveBy 整块拖拽移动窗口：位置 = 起点 + 指针绝对位移。
func (fs *floatWindow) moveBy(dx, dy float32) {
	if !fs.beginDrag(dragMove) {
		return
	}
	tx, ty, ok := fs.pointerDelta()
	if !ok {
		return
	}
	_ = x11win.MoveByTitle(floatWindowTitle, fs.anchorX+tx, fs.anchorY+ty)
}

// resizeTop 顶部边缘：下边界固定。
func (fs *floatWindow) resizeTop(_, dy float32) {
	if !fs.beginDrag(dragResizeTop) {
		return
	}
	_, ty, ok := fs.pointerDelta()
	if !ok {
		return
	}
	h := fs.anchorH - ty
	if h < fs.minH() {
		h = fs.minH()
	}
	y := fs.anchorY + fs.anchorH - h
	fs.applyGeometry(fs.anchorX, y, fs.anchorW, h)
}

// resizeBottom 底部边缘：上边界固定。
func (fs *floatWindow) resizeBottom(_, dy float32) {
	if !fs.beginDrag(dragResizeBottom) {
		return
	}
	_, ty, ok := fs.pointerDelta()
	if !ok {
		return
	}
	h := fs.anchorH + ty
	if h < fs.minH() {
		h = fs.minH()
	}
	fs.applyGeometry(fs.anchorX, fs.anchorY, fs.anchorW, h)
}

// resizeLeft 左侧边缘：右边界固定。
func (fs *floatWindow) resizeLeft(dx, _ float32) {
	if !fs.beginDrag(dragResizeLeft) {
		return
	}
	tx, _, ok := fs.pointerDelta()
	if !ok {
		return
	}
	w := fs.anchorW - tx
	if w < fs.minW() {
		w = fs.minW()
	}
	x := fs.anchorX + fs.anchorW - w
	fs.applyGeometry(x, fs.anchorY, w, fs.anchorH)
}

// resizeRight 右侧边缘：左边界固定。
func (fs *floatWindow) resizeRight(dx, _ float32) {
	if !fs.beginDrag(dragResizeRight) {
		return
	}
	tx, _, ok := fs.pointerDelta()
	if !ok {
		return
	}
	w := fs.anchorW + tx
	if w < fs.minW() {
		w = fs.minW()
	}
	fs.applyGeometry(fs.anchorX, fs.anchorY, w, fs.anchorH)
}

// applyGeometry 同时通过 X11 直接移动+缩放窗口（立即生效、跟随指针），
// 再走 Fyne 原生缩放使内容重排铺满新尺寸。直接 X11 几何请求与内容重排相互独立，
// 因此缩放能实时跟随鼠标，而不是等到释放时才一次性显示。
func (fs *floatWindow) applyGeometry(x, y, w, h int32) {
	_ = x11win.MoveResizeByTitle(floatWindowTitle, x, y, uint32(w), uint32(h))
	fs.win.Resize(fyne.NewSize(float32(w)/fs.scale, float32(h)/fs.scale))
}

// applyBorderless 在窗口出现后移除标题栏、确认窗口可被定位。
func (fs *floatWindow) applyBorderless() {
	for i := 0; i < 24; i++ {
		time.Sleep(250 * time.Millisecond)
		if x11win.RemoveDecorationsByTitle(floatWindowTitle) != nil {
			continue
		}
		if _, _, err := x11win.PositionByTitle(floatWindowTitle); err != nil {
			continue
		}
		fyne.Do(func() { fs.ready = true })
		fs.applyRoundedCorners()
		return
	}
}
