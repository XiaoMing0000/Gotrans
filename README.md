# Gotrans

一个用 Go + Fyne 编写的桌面翻译工具，界面类似 Google 翻译。默认使用免费的 MyMemory 翻译引擎，架构可插拔。

## 功能

- 基础翻译：原文 / 译文双栏、源/目标语言选择、⇄ 互换、复制译文。
- 划词翻译：在系统其它应用中框选文字（X11 PRIMARY 选区），译文自动显示在原文框下方的小面板。

## 构建

Fyne 桌面应用在 Linux 上依赖 X11 / Wayland / OpenGL 开发库（Fyne v2.8 默认同时编译 X11 与 Wayland 两个后端，缺一不可）。先安装（RHEL / AlmaLinux / Fedora）：

```bash
sudo dnf install -y gcc pkgconfig \
  wayland-devel wayland-protocols-devel \
  libxkbcommon-devel libxkbcommon-x11-devel \
  libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel \
  libXxf86vm-devel libXfixes-devel libXi-devel libXext-devel \
  libXrender-devel libXft-devel mesa-libGL-devel
```

> 若不想装 Wayland 相关包，也可只装 X11 系列（上面的 `libX11*`、`libxkbcommon*`、`mesa-libGL-devel`），然后用 `go build -tags x11 -o gotrans .` 强制只编译 X11 后端。

然后：

```bash
go build -o gotrans .
```

## 运行

```bash
./gotrans
```

## 配置

- `GOTRANS_PROVIDER`：选择翻译引擎，默认 `mymemory`（免 Key）。未知值会回退到默认引擎。
- 默认语言：英语 → 中文（简体），可在界面上切换或点「⇄」互换。

## 注意事项

- MyMemory 匿名调用有每日配额（约数千字符），仅适合 demo。正式使用可接入 Google Cloud Translation / DeepL，在 `internal/translator` 下新增实现并注册到 `New()` 即可。
- 划词翻译读取的是 X11 PRIMARY 选区，仅在 X11 / XWayland 下生效；原生 Wayland 应用（部分 GTK4 / Qt6）的框选可能读不到。
