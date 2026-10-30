package x11win

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// TestLiveOpacity 端到端校验：需应用正在运行、窗口已显示，并以
// GOTRANS_LIVE_TEST=1 显式启用，否则跳过。
func TestLiveOpacity(t *testing.T) {
	if os.Getenv("GOTRANS_LIVE_TEST") != "1" {
		t.Skip("设置 GOTRANS_LIVE_TEST=1 且应用运行时执行")
	}
	got, err := readOpacityByTitle("Go 翻译 · Gotrans")
	if err != nil {
		t.Fatal(err)
	}
	if got < 0.65 || got > 0.75 {
		t.Fatalf("opacity = %f, 期望约 0.7", got)
	}
}

func readOpacityByTitle(title string) (float64, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	netClientList := intern(conn, "_NET_CLIENT_LIST")
	netWMName := intern(conn, "_NET_WM_NAME")
	netOpacity := intern(conn, "_NET_WM_WINDOW_OPACITY")
	utf8 := intern(conn, "UTF8_STRING")

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	reply, err := xproto.GetProperty(conn, false, root, netClientList, xproto.AtomWindow, 0, 1<<20).Reply()
	if err != nil {
		return 0, err
	}
	for _, win := range windowList(reply.Value) {
		if readText(conn, win, netWMName, utf8) != title {
			continue
		}
		prop, err := xproto.GetProperty(conn, false, win, netOpacity, xproto.AtomCardinal, 0, 1).Reply()
		if err != nil {
			return 0, err
		}
		if len(prop.Value) < 4 {
			return 0, fmt.Errorf("窗口 %q 的透明度属性为空", title)
		}
		v := binary.LittleEndian.Uint32(prop.Value[:4])
		return float64(v) / float64(0xffffffff), nil
	}
	return 0, fmt.Errorf("未找到标题为 %q 的窗口", title)
}
