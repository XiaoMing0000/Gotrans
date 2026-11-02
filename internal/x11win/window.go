//go:build linux

package x11win

import (
	"encoding/binary"
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// findWindowByTitle 在 _NET_CLIENT_LIST 中查找标题为 title 的顶层窗口，未找到返回 0。
func findWindowByTitle(conn *xgb.Conn, title string) xproto.Window {
	netClientList := intern(conn, "_NET_CLIENT_LIST")
	netWMName := intern(conn, "_NET_WM_NAME")
	utf8 := intern(conn, "UTF8_STRING")
	wmName := intern(conn, "WM_NAME")
	str := intern(conn, "STRING")

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	reply, err := xproto.GetProperty(conn, false, root, netClientList, xproto.AtomWindow, 0, 1<<20).Reply()
	if err != nil {
		return 0
	}
	for _, win := range windowList(reply.Value) {
		name := readText(conn, win, netWMName, utf8)
		if name == "" {
			name = readText(conn, win, wmName, str)
		}
		if name == title {
			return win
		}
	}
	return 0
}

// MoveByTitle 请求窗口管理器将标题为 title 的窗口移动到绝对坐标 (x, y)（物理像素）。
func MoveByTitle(title string, x, y int32) error {
	return configureWindow(title, x, y, 0, 0, xproto.ConfigWindowX|xproto.ConfigWindowY)
}

// ResizeByTitle 请求窗口管理器把窗口尺寸调整为 w×h（左上角不变）。
func ResizeByTitle(title string, w, h uint32) error {
	return configureWindow(title, 0, 0, w, h, xproto.ConfigWindowWidth|xproto.ConfigWindowHeight)
}

// MoveResizeByTitle 同时调整窗口位置与尺寸。
func MoveResizeByTitle(title string, x, y int32, w, h uint32) error {
	return configureWindow(title, x, y, w, h,
		xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight)
}

// configureWindow 发送 ConfigureWindow 请求；掩码决定哪些字段生效，
// valueList 按 X、Y、Width、Height 的掩码顺序拼接。
func configureWindow(title string, x, y int32, w, h uint32, mask uint16) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	win := findWindowByTitle(conn, title)
	if win == 0 {
		return fmt.Errorf("未找到标题为 %q 的窗口", title)
	}

	var values []uint32
	if mask&xproto.ConfigWindowX != 0 {
		values = append(values, uint32(x))
	}
	if mask&xproto.ConfigWindowY != 0 {
		values = append(values, uint32(y))
	}
	if mask&xproto.ConfigWindowWidth != 0 {
		values = append(values, w)
	}
	if mask&xproto.ConfigWindowHeight != 0 {
		values = append(values, h)
	}
	xproto.ConfigureWindow(conn, win, mask, values)
	conn.Sync()
	return nil
}

// RemoveDecorationsByTitle 移除窗口装饰（无边框），通过 _MOTIF_WM_HINTS 的 decorations=0 实现。
// 依赖窗口管理器支持 Motif 提示；个别 WM 可能忽略，此时窗口仍会保留标题栏。
func RemoveDecorationsByTitle(title string) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	win := findWindowByTitle(conn, title)
	if win == 0 {
		return fmt.Errorf("未找到标题为 %q 的窗口", title)
	}
	motif := intern(conn, "_MOTIF_WM_HINTS")
	// flags=2（MWM_HINTS_DECORATIONS），其余字段置 0 → 关闭全部装饰。
	hints := []uint32{2, 0, 0, 0, 0}
	xproto.ChangeProperty(conn, xproto.PropModeReplace, win, motif, motif, 32,
		uint32(len(hints)), u32s(hints))
	conn.Sync()
	return nil
}

// SetAlwaysOnTopByTitle 设置/取消窗口置顶。on 为 true 时把窗口加入
// _NET_WM_STATE_ABOVE 状态（窗口管理器将其置于普通窗口之上）；false 时移除该状态还原堆叠。
func SetAlwaysOnTopByTitle(title string, on bool) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	win := findWindowByTitle(conn, title)
	if win == 0 {
		return fmt.Errorf("未找到标题为 %q 的窗口", title)
	}

	root := xproto.Setup(conn).DefaultScreen(conn).Root
	above := intern(conn, "_NET_WM_STATE_ABOVE")
	state := intern(conn, "_NET_WM_STATE")

	// data.l[0]=action（1=add，0=remove），data.l[1]=第一个状态 atom，
	// data.l[3]=source（1=程序主动请求）。
	action := uint32(0)
	if on {
		action = 1
	}
	data := []uint32{action, uint32(above), 0, 1, 0}

	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: win,
		Type:   state,
		Data:   xproto.ClientMessageDataUnionData32New(data),
	}
	const mask = xproto.EventMaskSubstructureRedirect | xproto.EventMaskSubstructureNotify
	xproto.SendEvent(conn, false, root, uint32(mask), string(ev.Bytes()))
	conn.Sync()
	return nil
}

// PositionByTitle 返回标题为 title 的窗口左上角绝对坐标（物理像素）。
func PositionByTitle(title string) (int32, int32, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	win := findWindowByTitle(conn, title)
	if win == 0 {
		return 0, 0, fmt.Errorf("未找到标题为 %q 的窗口", title)
	}

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root
	r, err := xproto.TranslateCoordinates(conn, win, root, 0, 0).Reply()
	if err != nil {
		return 0, 0, err
	}
	return int32(r.DstX), int32(r.DstY), nil
}

// SizeByTitle 返回标题为 title 的窗口尺寸（物理像素）。
func SizeByTitle(title string) (uint32, uint32, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	win := findWindowByTitle(conn, title)
	if win == 0 {
		return 0, 0, fmt.Errorf("未找到标题为 %q 的窗口", title)
	}
	g, err := xproto.GetGeometry(conn, xproto.Drawable(win)).Reply()
	if err != nil {
		return 0, 0, err
	}
	return uint32(g.Width), uint32(g.Height), nil
}

// PointerPosition 返回指针在根窗口上的绝对坐标（物理像素），用于拖拽窗口时
// 以“绝对位移”计算目标位置，避免“窗口移动 → 相对坐标变化 → 增量失真”的抖动。
func PointerPosition() (int32, int32, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	root := xproto.Setup(conn).DefaultScreen(conn).Root
	q, err := xproto.QueryPointer(conn, root).Reply()
	if err != nil {
		return 0, 0, err
	}
	return int32(q.RootX), int32(q.RootY), nil
}

// u32s 将 []uint32 打包为小端字节序列。
func u32s(v []uint32) []byte {
	b := make([]byte, 4*len(v))
	for i, x := range v {
		binary.LittleEndian.PutUint32(b[i*4:i*4+4], x)
	}
	return b
}

// intern 缓存并返回指定名字的 atom。
func intern(conn *xgb.Conn, name string) xproto.Atom {
	r, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return xproto.AtomNone
	}
	return r.Atom
}

// windowList 把 _NET_CLIENT_LIST 的属性值（窗口 ID 小端序列）解析为窗口列表。
func windowList(data []byte) []xproto.Window {
	var out []xproto.Window
	for i := 0; i+4 <= len(data); i += 4 {
		out = append(out, xproto.Window(binary.LittleEndian.Uint32(data[i:i+4])))
	}
	return out
}

// readText 读取窗口 atom 属性的文本值；属性缺失或类型不匹配返回空串。
func readText(conn *xgb.Conn, win xproto.Window, atom, typ xproto.Atom) string {
	reply, err := xproto.GetProperty(conn, false, win, atom, typ, 0, 1<<20).Reply()
	if err != nil || reply.Type == xproto.AtomNone {
		return ""
	}
	return string(reply.Value)
}
