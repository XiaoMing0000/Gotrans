//go:build linux

// Package selection 读取 X11 的 PRIMARY 选区——即用鼠标「框选」产生的选区，
// 与 Ctrl+C 写入的 CLIPBOARD 不同。仅在 X11 / XWayland 下可用，原生 Wayland 下不可用。
package selection

import (
	"fmt"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Watcher 持有一个 X11 连接和一个用于接收选区转换结果的隐藏窗口。
type Watcher struct {
	conn    *xgb.Conn
	win     xproto.Window
	primary xproto.Atom
	utf8    xproto.Atom
	str     xproto.Atom
	prop    xproto.Atom
}

// New 连接 X 服务器并创建隐藏窗口。失败时（如无 X11 环境）返回错误。
func New() (*Watcher, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("无法连接 X11（原生 Wayland 环境下不可用）: %w", err)
	}

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)

	win, err := xproto.NewWindowId(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	xproto.CreateWindow(conn, screen.RootDepth, win, screen.Root,
		0, 0, 1, 1, 0, xproto.WindowClassInputOutput, screen.RootVisual, 0, nil)

	return &Watcher{
		conn:    conn,
		win:     win,
		primary: internAtom(conn, "PRIMARY"),
		utf8:    internAtom(conn, "UTF8_STRING"),
		str:     internAtom(conn, "STRING"),
		prop:    internAtom(conn, "GOTRANS_SELECTION"),
	}, nil
}

// Close 释放 X11 资源。
func (w *Watcher) Close() {
	if w.conn != nil {
		w.conn.Close()
	}
}

// ReadPrimary 返回当前 PRIMARY 选区文本；无选区时返回空串。
// 按 ICCCM 协议：找到选区拥有者，请求把选区转换为文本并读回。
func (w *Watcher) ReadPrimary() (string, error) {
	owner, err := xproto.GetSelectionOwner(w.conn, w.primary).Reply()
	if err != nil {
		return "", err
	}
	if owner.Owner == xproto.WindowNone {
		return "", nil
	}

	// 优先 UTF8_STRING，被拒绝或超时则回退到 STRING。
	xproto.ConvertSelection(w.conn, w.win, w.primary, w.utf8, w.prop, xproto.TimeCurrentTime)
	w.conn.Sync()
	if text, err := w.awaitNotify(w.utf8); err == nil && text != "" {
		return text, nil
	}

	xproto.ConvertSelection(w.conn, w.win, w.primary, w.str, w.prop, xproto.TimeCurrentTime)
	w.conn.Sync()
	return w.awaitNotify(w.str)
}

func (w *Watcher) awaitNotify(target xproto.Atom) (string, error) {
	deadline := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(deadline) {
		ev, err := w.conn.PollForEvent()
		if err != nil {
			return "", err
		}
		if ev == nil {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		notify, ok := ev.(xproto.SelectionNotifyEvent)
		if !ok {
			continue
		}
		if notify.Requestor != w.win || notify.Selection != w.primary || notify.Target != target {
			continue
		}
		if notify.Property == xproto.AtomNone {
			return "", fmt.Errorf("选区转换被拒绝")
		}
		return w.readProperty(notify.Property)
	}
	return "", fmt.Errorf("读取选区超时")
}

func (w *Watcher) readProperty(prop xproto.Atom) (string, error) {
	reply, err := xproto.GetProperty(w.conn, true, w.win, prop,
		xproto.AtomAny, 0, 1<<20).Reply()
	if err != nil {
		return "", err
	}
	if reply.Type == xproto.AtomNone {
		return "", nil
	}
	return string(reply.Value), nil
}

func internAtom(conn *xgb.Conn, name string) xproto.Atom {
	reply, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return xproto.AtomNone
	}
	return reply.Atom
}
