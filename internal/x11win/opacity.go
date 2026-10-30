// Package x11win 通过 X11 的 EWMH 属性控制窗口。
package x11win

import (
	"encoding/binary"
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// SetOpacityByTitle 将标题为 title 的顶层窗口透明度设为 opacity（0~1）。
// 依赖 _NET_WM_WINDOW_OPACITY，仅在 X11 / XWayland 且桌面支持合成器时可见。
func SetOpacityByTitle(title string, opacity float32) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	netClientList := intern(conn, "_NET_CLIENT_LIST")
	netWMName := intern(conn, "_NET_WM_NAME")
	netOpacity := intern(conn, "_NET_WM_WINDOW_OPACITY")
	utf8 := intern(conn, "UTF8_STRING")
	wmName := intern(conn, "WM_NAME")
	str := intern(conn, "STRING")

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	reply, err := xproto.GetProperty(conn, false, root, netClientList, xproto.AtomWindow, 0, 1<<20).Reply()
	if err != nil {
		return err
	}

	for _, win := range windowList(reply.Value) {
		name := readText(conn, win, netWMName, utf8)
		if name == "" {
			name = readText(conn, win, wmName, str)
		}
		if name != title {
			continue
		}
		v := opacityToCardinal(opacity)
		xproto.ChangeProperty(conn, xproto.PropModeReplace, win, netOpacity,
			xproto.AtomCardinal, 32, 1, u32Bytes(v))
		conn.Sync()
		return nil
	}
	return fmt.Errorf("未找到标题为 %q 的窗口", title)
}

func intern(conn *xgb.Conn, name string) xproto.Atom {
	r, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return xproto.AtomNone
	}
	return r.Atom
}

func windowList(data []byte) []xproto.Window {
	var out []xproto.Window
	for i := 0; i+4 <= len(data); i += 4 {
		out = append(out, xproto.Window(binary.LittleEndian.Uint32(data[i:i+4])))
	}
	return out
}

func readText(conn *xgb.Conn, win xproto.Window, atom, typ xproto.Atom) string {
	reply, err := xproto.GetProperty(conn, false, win, atom, typ, 0, 1<<20).Reply()
	if err != nil || reply.Type == xproto.AtomNone {
		return ""
	}
	return string(reply.Value)
}

// opacityToCardinal 将透明度映射为 _NET_WM_WINDOW_OPACITY 的 32 位值。
// 用 float64 计算并钳制到 [0,1]，避免 float32(0xffffffff) 向上取整到 2^32
// 导致 uint32 回绕为 0（窗口完全透明）的 bug。
func opacityToCardinal(opacity float32) uint32 {
	f := float64(opacity)
	if f < 0 {
		f = 0
	} else if f > 1 {
		f = 1
	}
	return uint32(f * float64(0xffffffff))
}

func u32Bytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}
