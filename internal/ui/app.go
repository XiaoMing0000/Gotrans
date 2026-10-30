// Package ui 负责构建 Fyne 桌面界面。
package ui

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gotrans/internal/config"
	"gotrans/internal/selection"
	"gotrans/internal/translator"
	"gotrans/internal/x11win"
)

const (
	windowTitle    = "Go 翻译 · Gotrans"
	selPlaceholder = "（在其它应用中框选文字，原文将显示在这里）"
)

// Build 构造并返回应用主窗口。provider 为翻译引擎。
func Build(a fyne.App, p translator.Provider) fyne.Window {
	cfg, _ := config.Load()

	// 字体透明度通过主题只作用于文字颜色；窗口透明度走 X11 整窗属性（见文件末尾）。
	a.Settings().SetTheme(&fontOpacityTheme{
		Theme: theme.DefaultTheme(),
		alpha: config.ClampOpacity(cfg.FontOpacity),
	})

	w := a.NewWindow(windowTitle)
	w.Resize(fyne.NewSize(860, 620))

	// 语种下拉框：以 Name 为选项，内部映射回 Language。
	langs := p.Languages()
	names := make([]string, len(langs))
	byName := make(map[string]translator.Language, len(langs))
	for i, l := range langs {
		names[i] = l.Name
		byName[l.Name] = l
	}

	srcSelect := widget.NewSelect(names, nil)
	dstSelect := widget.NewSelect(names, nil)
	srcSelect.SetSelected("英语")
	dstSelect.SetSelected("中文（简体）")

	srcEntry := widget.NewMultiLineEntry()
	srcEntry.SetPlaceHolder("输入要翻译的文本…")
	dstEntry := widget.NewMultiLineEntry()
	dstEntry.SetPlaceHolder("译文将显示在这里")

	status := widget.NewLabel("就绪")
	status.Wrapping = fyne.TextTruncate

	// 当前源/目标语言，供后台划词 goroutine 安全读取。
	var langMu sync.Mutex
	curFrom := byName["英语"]
	curTo := byName["中文（简体）"]
	srcSelect.OnChanged = func(name string) {
		langMu.Lock()
		curFrom = byName[name]
		langMu.Unlock()
	}
	dstSelect.OnChanged = func(name string) {
		langMu.Lock()
		curTo = byName[name]
		langMu.Unlock()
	}
	getLang := func() (translator.Language, translator.Language) {
		langMu.Lock()
		defer langMu.Unlock()
		return curFrom, curTo
	}

	// 划词翻译开关；selSeq 作为版本号，清空或产生新选区时递增，
	// 用于丢弃过期的在途翻译结果。
	selEnabled := &atomic.Bool{}
	selEnabled.Store(cfg.SelectionEnabled)
	selSeq := &atomic.Uint64{}

	// 划词翻译：原文显示在输入框下方（左侧），译文显示在输出框下方（右侧），
	// 仿照主输入/输出框使用可编辑输入框，支持框选文本与右键菜单（复制/全选等）。
	selSource := widget.NewMultiLineEntry()
	selSource.SetPlaceHolder(selPlaceholder)
	selSource.Wrapping = fyne.TextWrapWord
	selResult := widget.NewMultiLineEntry()
	selResult.Wrapping = fyne.TextWrapWord
	selResult.TextStyle = fyne.TextStyle{Bold: true}

	if watcher, err := selection.New(); err != nil {
		selSource.SetText("划词翻译不可用")
		selResult.SetText(err.Error())
	} else {
		go watchSelection(watcher, p, getLang, selEnabled, selSeq, selSource, selResult)
	}

	// 划词栏：显示原文时原文/译文左右并排；隐藏原文时译文全宽填充。
	selBar := container.NewGridWithColumns(2, selSource, selResult)
	setSelectionBar := func(showSource bool) {
		if showSource {
			selSource.Show()
			selBar.Objects = []fyne.CanvasObject{selSource, selResult}
			selBar.Layout = layout.NewGridLayoutWithColumns(2)
		} else {
			selSource.Hide()
			selBar.Objects = []fyne.CanvasObject{selResult}
			selBar.Layout = layout.NewGridLayoutWithColumns(1)
		}
		selBar.Refresh()
	}
	setSelectionBar(cfg.ShowSelectionSource)

	// setSelectionEnabled 在构建完 middle（VSplit）后赋值，控制划词栏显示/隐藏。
	var setSelectionEnabled func(on bool)

	translateBtn := widget.NewButton("翻译", nil)
	swapBtn := widget.NewButton("⇄", nil)
	copyBtn := widget.NewButton("复制译文", nil)
	clearInputBtn := widget.NewButton("清空输入", func() {
		// 同时清空：主输入框、主译文、划词原文、划词译文。
		srcEntry.SetText("")
		dstEntry.SetText("")
		selSeq.Add(1)
		selSource.SetText("")
		selResult.SetText("")
	})
	settingsBtn := widget.NewButton("设置", nil)

	translateBtn.OnTapped = func() {
		text := srcEntry.Text
		if strings.TrimSpace(text) == "" {
			status.SetText("请输入要翻译的内容")
			return
		}
		from := byName[srcSelect.Selected]
		to := byName[dstSelect.Selected]

		status.SetText("翻译中…")
		translateBtn.Disable()
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			res, err := p.Translate(ctx, text, from, to)
			// 网络请求在后台执行，结果必须回到 UI 线程刷新控件。
			fyne.Do(func() {
				translateBtn.Enable()
				if err != nil {
					status.SetText("翻译失败：" + err.Error())
					return
				}
				dstEntry.SetText(res)
				status.SetText("翻译完成（引擎：" + p.Name() + "）")
			})
		}()
	}

	swapBtn.OnTapped = func() {
		s, d := srcSelect.Selected, dstSelect.Selected
		srcSelect.SetSelected(d)
		dstSelect.SetSelected(s)
	}

	copyBtn.OnTapped = func() {
		if strings.TrimSpace(dstEntry.Text) == "" {
			status.SetText("没有可复制的译文")
			return
		}
		w.Clipboard().SetContent(dstEntry.Text)
		status.SetText("已复制到剪贴板")
	}

	settingsBtn.OnTapped = func() {
		showSettings(a, w, &cfg, setSelectionEnabled, setSelectionBar)
	}

	top := container.NewGridWithColumns(3, srcSelect, swapBtn, dstSelect)
	// 主区域上下分割：上方为原文/译文输入输出框，下方为划词栏。
	// 分隔条可拖拽调整划词栏高度；划词关闭时下方收起、上方全高度填充。
	topArea := container.NewGridWithColumns(2, srcEntry, dstEntry)
	middle := container.NewVSplit(topArea, selBar)
	middle.SetOffset(0.8)
	bottom := container.NewHBox(translateBtn, copyBtn, clearInputBtn, settingsBtn, status)

	setSelectionEnabled = func(on bool) {
		selEnabled.Store(on)
		if on {
			selBar.Show()
		} else {
			selBar.Hide()
		}
		middle.Refresh()
	}
	setSelectionEnabled(cfg.SelectionEnabled)

	w.SetContent(container.NewBorder(top, bottom, nil, nil, middle))

	// 窗口显示后按保存的配置应用窗口透明度（需等 X11 窗口出现）。
	if cfg.WindowOpacity < 0.999 {
		applyOpacityWithRetry(cfg.WindowOpacity)
	}
	return w
}

// showSettings 弹出设置对话框：划词开关 + 显示原文 + 窗口/字体透明度滑杆。
func showSettings(a fyne.App, parent fyne.Window, cfg *config.Config,
	setSelectionEnabled func(on bool), setSelectionBar func(showSource bool)) {

	showSource := widget.NewCheck("显示原文", func(on bool) {
		cfg.ShowSelectionSource = on
		setSelectionBar(on)
		_ = cfg.Save()
	})
	showSource.SetChecked(cfg.ShowSelectionSource)

	check := widget.NewCheck("启用划词翻译", func(on bool) {
		setSelectionEnabled(on)
		cfg.SelectionEnabled = on
		if on {
			showSource.Enable()
		} else {
			showSource.Disable()
		}
		_ = cfg.Save()
	})
	check.SetChecked(cfg.SelectionEnabled)
	if !cfg.SelectionEnabled {
		showSource.Disable()
	}

	// 子选项缩进显示在「启用划词翻译」下方。
	selectionSub := container.NewHBox(widget.NewLabel("    "), showSource)

	// applyFontTheme 按当前字体透明度重建主题并立即生效（只作用于文字颜色）。
	applyFontTheme := func() {
		a.Settings().SetTheme(&fontOpacityTheme{
			Theme: theme.DefaultTheme(),
			alpha: config.ClampOpacity(cfg.FontOpacity),
		})
	}

	opacityLabel := widget.NewLabel(fmt.Sprintf("窗口透明度：%.0f%%", cfg.WindowOpacity*100))
	slider := widget.NewSlider(config.MinOpacity, config.MaxOpacity)
	slider.Step = 0.01
	slider.Value = config.ClampOpacity(cfg.WindowOpacity)
	slider.OnChanged = func(v float64) {
		v = config.ClampOpacity(v)
		cfg.WindowOpacity = v
		opacityLabel.SetText(fmt.Sprintf("窗口透明度：%.0f%%", v*100))
		go func() { _ = x11win.SetOpacityByTitle(windowTitle, float32(v)) }()
	}
	slider.OnChangeEnded = func(v float64) {
		cfg.WindowOpacity = config.ClampOpacity(v)
		_ = cfg.Save()
	}

	fontOpacityLabel := widget.NewLabel(fmt.Sprintf("字体透明度：%.0f%%", cfg.FontOpacity*100))
	fontSlider := widget.NewSlider(config.MinOpacity, config.MaxOpacity)
	fontSlider.Step = 0.01
	fontSlider.Value = config.ClampOpacity(cfg.FontOpacity)
	fontSlider.OnChanged = func(v float64) {
		v = config.ClampOpacity(v)
		cfg.FontOpacity = v
		fontOpacityLabel.SetText(fmt.Sprintf("字体透明度：%.0f%%", v*100))
		applyFontTheme()
	}
	fontSlider.OnChangeEnded = func(v float64) {
		cfg.FontOpacity = config.ClampOpacity(v)
		_ = cfg.Save()
	}

	content := container.NewVBox(
		check,
		selectionSub,
		opacityLabel,
		slider,
		fontOpacityLabel,
		fontSlider,
		widget.NewLabel("提示：两项透明度下限均为 20%，防止窗口或文字完全透明后无法操作。"),
	)
	dialog.ShowCustom("设置", "关闭", content, parent)
}

// fontOpacityTheme 在默认主题上只对文字颜色附加透明度，实现「字体透明度」。
// 窗口透明度走 X11 整窗属性，两者互不影响。
type fontOpacityTheme struct {
	fyne.Theme
	alpha float64
}

func (t *fontOpacityTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	c := t.Theme.Color(name, variant)
	if !isTextColor(name) {
		return c
	}
	n, ok := color.NRGBAModel.Convert(c).(color.NRGBA)
	if !ok {
		return c
	}
	n.A = uint8(t.alpha*255 + 0.5)
	return n
}

// isTextColor 判断颜色名是否为文字用色，字体透明度只作用于这些颜色。
func isTextColor(name fyne.ThemeColorName) bool {
	switch name {
	case theme.ColorNameForeground, theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return true
	}
	return false
}

// applyOpacityWithRetry 在窗口出现后重试设置窗口透明度。
func applyOpacityWithRetry(opacity float64) {
	go func() {
		for i := 0; i < 24; i++ {
			time.Sleep(250 * time.Millisecond)
			if x11win.SetOpacityByTitle(windowTitle, float32(opacity)) == nil {
				return
			}
		}
	}()
}

// watchSelection 轮询全局 PRIMARY 选区，变化时翻译并刷新面板。
func watchSelection(watcher *selection.Watcher, p translator.Provider,
	getLang func() (translator.Language, translator.Language),
	enabled *atomic.Bool, seq *atomic.Uint64, selSource, selResult *widget.Entry) {

	var last string
	for {
		if !enabled.Load() {
			time.Sleep(400 * time.Millisecond)
			continue
		}

		text, err := watcher.ReadPrimary()
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		text = strings.TrimSpace(text)
		// 注意：清空时不重置 last，避免仍然选中的旧文字被当成新选区重新显示。
		if text == "" || text == last {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		// 新选区：递增版本号，使之前尚未返回的翻译结果失效。
		token := seq.Add(1)
		last = text
		from, to := getLang()

		fyne.Do(func() {
			selSource.SetText(truncate(text, 300))
			selResult.SetText("翻译中…")
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		res, err := p.Translate(ctx, text, from, to)
		cancel()

		// 期间若点击了清空或产生了新一轮选区，则丢弃本次结果。
		if seq.Load() != token {
			continue
		}
		fyne.Do(func() {
			if err != nil {
				selResult.SetText("翻译失败：" + err.Error())
			} else {
				selResult.SetText(res)
			}
		})
	}
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
