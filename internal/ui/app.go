// Package ui 负责构建 Fyne 桌面界面。
package ui

import (
	"context"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gotrans/internal/selection"
	"gotrans/internal/translator"
)

// Build 构造并返回应用主窗口。provider 为翻译引擎。
func Build(a fyne.App, p translator.Provider) fyne.Window {
	w := a.NewWindow("Go 翻译 · Gotrans")
	w.Resize(fyne.NewSize(840, 580))

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

	translateBtn := widget.NewButton("翻译", nil)
	swapBtn := widget.NewButton("⇄", nil)
	copyBtn := widget.NewButton("复制译文", nil)

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

	// 划词翻译面板：显示全局框选文字及其译文。
	selSource := widget.NewLabel("（在其它应用中框选文字，译文将自动显示在这里）")
	selSource.Wrapping = fyne.TextWrapWord
	selResult := widget.NewLabel("")
	selResult.Wrapping = fyne.TextWrapWord
	selResult.TextStyle = fyne.TextStyle{Bold: true}
	selCard := widget.NewCard("划词翻译", "", container.NewVBox(selSource, selResult))

	if watcher, err := selection.New(); err != nil {
		selSource.SetText("划词翻译不可用")
		selResult.SetText(err.Error())
	} else {
		go watchSelection(watcher, p, getLang, selSource, selResult)
	}

	top := container.NewGridWithColumns(3, srcSelect, swapBtn, dstSelect)
	left := container.NewBorder(nil, selCard, nil, nil, srcEntry)
	middle := container.NewGridWithColumns(2, left, dstEntry)
	bottom := container.NewHBox(translateBtn, copyBtn, status)

	w.SetContent(container.NewBorder(top, bottom, nil, nil, middle))
	return w
}

// watchSelection 轮询全局 PRIMARY 选区，变化时翻译并刷新面板。
func watchSelection(watcher *selection.Watcher, p translator.Provider,
	getLang func() (translator.Language, translator.Language),
	selSource, selResult *widget.Label) {

	var last string
	for {
		text, err := watcher.ReadPrimary()
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" || text == last {
			time.Sleep(300 * time.Millisecond)
			continue
		}
		last = text
		from, to := getLang()

		fyne.Do(func() {
			selSource.SetText(truncate(text, 300))
			selResult.SetText("翻译中…")
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		res, err := p.Translate(ctx, text, from, to)
		cancel()
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
