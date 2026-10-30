package selection

import (
	"testing"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// TestReadPrimary 在真实 X 服务器上模拟一个 PRIMARY 选区拥有者，
// 验证 Watcher 能按 ICCCM 协议读回文本。无 X11 环境时跳过。
func TestReadPrimary(t *testing.T) {
	ownerConn, err := xgb.NewConn()
	if err != nil {
		t.Skipf("无 X11 环境，跳过: %v", err)
	}
	defer ownerConn.Close()

	setup := xproto.Setup(ownerConn)
	screen := setup.DefaultScreen(ownerConn)
	ownerWin, err := xproto.NewWindowId(ownerConn)
	if err != nil {
		t.Fatal(err)
	}
	xproto.CreateWindow(ownerConn, screen.RootDepth, ownerWin, screen.Root,
		0, 0, 1, 1, 0, xproto.WindowClassInputOutput, screen.RootVisual, 0, nil)

	primary := internAtom(ownerConn, "PRIMARY")
	xproto.SetSelectionOwner(ownerConn, ownerWin, primary, xproto.TimeCurrentTime)
	ownerConn.Sync()

	const want = "hello world"
	go func() {
		for {
			ev, err := ownerConn.WaitForEvent()
			if err != nil {
				return
			}
			req, ok := ev.(xproto.SelectionRequestEvent)
			if !ok {
				continue
			}
			xproto.ChangeProperty(ownerConn, xproto.PropModeReplace,
				req.Requestor, req.Property, req.Target, 8, uint32(len(want)), []byte(want))
			notify := xproto.SelectionNotifyEvent{
				Time:      req.Time,
				Requestor: req.Requestor,
				Selection: req.Selection,
				Target:    req.Target,
				Property:  req.Property,
			}
			xproto.SendEvent(ownerConn, false, req.Requestor, 0, string(notify.Bytes()))
			ownerConn.Sync()
		}
	}()

	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	got, err := w.ReadPrimary()
	if err != nil {
		t.Fatalf("ReadPrimary: %v", err)
	}
	if got != want {
		t.Fatalf("ReadPrimary = %q, want %q", got, want)
	}
}
