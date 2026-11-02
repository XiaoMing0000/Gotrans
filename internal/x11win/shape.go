package x11win

import (
	"fmt"
	"math"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/shape"
	"github.com/jezek/xgb/xproto"
)

// SetRoundedCornersByTitle 使用 X Shape 扩展把标题为 title 的窗口裁剪为四角圆角
// （半径 radius，物理像素）。同时设置 bounding 与 input 形状，使圆角外的透明区域
// 既不渲染也不响应点击。形状以窗口左上角为原点、按当前窗口尺寸生成；窗口缩放后需重新调用。
//
// 依赖 SHAPE 扩展与窗口管理器按形状裁剪；个别 WM 可能忽略，此时窗口仍保持直角。
func SetRoundedCornersByTitle(title string, radius int) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := shape.Init(conn); err != nil {
		return err
	}

	win := findWindowByTitle(conn, title)
	if win == 0 {
		return fmt.Errorf("未找到标题为 %q 的窗口", title)
	}

	g, err := xproto.GetGeometry(conn, xproto.Drawable(win)).Reply()
	if err != nil {
		return err
	}

	rects := roundedRectStrips(int32(g.Width), int32(g.Height), int32(radius))
	// 0 表示 UnSorted 排序方式，服务器会自动排序，任意矩形顺序均可接受。
	for _, kind := range []shape.Kind{shape.SkBounding, shape.SkInput} {
		shape.Rectangles(conn, shape.SoSet, kind, 0, win, 0, 0, rects)
	}
	conn.Sync()
	return nil
}

// roundedRectStrips 把 w×h 的矩形（四角半径 r）分解为若干 1 像素高的水平条，
// 供 SHAPE 扩展用矩形列表逼近圆角。半径非正或尺寸非法时返回整块矩形。
func roundedRectStrips(w, h, r int32) []xproto.Rectangle {
	if w <= 0 || h <= 0 || r <= 0 {
		return []xproto.Rectangle{{X: 0, Y: 0, Width: uint16(w), Height: uint16(h)}}
	}
	if 2*r > w {
		r = w / 2
	}
	if 2*r > h {
		r = h / 2
	}

	rects := make([]xproto.Rectangle, 0, h)
	for y := int32(0); y < h; y++ {
		// 以像素中心（y+0.5）采样连续圆角矩形，使上/下圆角对称。
		t := float64(y) + 0.5
		var inset float64
		switch {
		case t < float64(r): // 上圆角：圆心在 (r, r)
			dy := float64(r) - t
			inset = float64(r) - math.Sqrt(float64(r)*float64(r)-dy*dy)
		case t > float64(h)-float64(r): // 下圆角：圆心在 (r, h-r)
			dy := t - (float64(h) - float64(r))
			inset = float64(r) - math.Sqrt(float64(r)*float64(r)-dy*dy)
		}

		x0 := int32(math.Round(inset))
		x1 := int32(math.Round(float64(w) - inset))
		if x1 <= x0 {
			x1 = x0 + 1
		}
		rects = append(rects, xproto.Rectangle{
			X:      int16(x0),
			Y:      int16(y),
			Width:  uint16(x1 - x0),
			Height: 1,
		})
	}
	return rects
}
