# Gotrans

一个用 Go + Fyne 编写的桌面翻译工具，界面类似 Google 翻译。支持免费的 MyMemory 与 Google Cloud Translation，架构可插拔。

## 功能

- 基础翻译：原文 / 译文双栏、源/目标语言选择、⇄ 互换、复制译文、清空输入。
- 划词翻译：在系统其它应用中框选文字（X11 PRIMARY 选区），原文显示在输入框下方、译文显示在输出框下方（最小高度固定，超长内容滚动），可一键清空。勾选划词翻译时，可拖拽分隔条调整划词栏高度。
- 设置：切换翻译引擎（MyMemory / Google Translate）、填写 Google API Key、自定义 Google / MyMemory 接口地址、启用 / 禁用划词翻译（关闭时隐藏划词栏、主区域全高度填充；含子选项：显示原文，隐藏原文后译文全宽填充）、调整窗口透明度与字体透明度（下限均 20%）。配置自动保存到用户目录，下次启动自动记忆。

## 构建

本项目强制使用 X11 后端（`-tags x11`），以保证划词翻译和窗口透明度正常工作。先安装 X11 / OpenGL 开发库（RHEL / AlmaLinux / Fedora）：

```bash
sudo dnf install -y gcc pkgconfig \
  libxkbcommon-devel libxkbcommon-x11-devel \
  libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel \
  libXxf86vm-devel libXfixes-devel libXi-devel libXext-devel \
  libXrender-devel libXft-devel mesa-libGL-devel
```

然后构建：

```bash
go build -tags x11 -o gotrans .
```

## 运行

```bash
./gotrans
```

## 配置

- 翻译引擎与密钥通过「设置」窗口配置：可选 `MyMemory`（免 Key）或 `Google Translate`（需在 [Google Cloud](https://cloud.google.com/translate) 申请 API Key）。
- `GOTRANS_PROVIDER`：环境变量兜底选择引擎（`mymemory` / `google`），仅当配置文件中未保存引擎时生效；未知值会回退到默认引擎。
- 界面设置（翻译引擎、Google API Key、Google / MyMemory 接口地址、划词开关、窗口/字体透明度）保存在 `~/.config/gotrans/config.json`，下次启动自动加载。Google 接口地址为完整端点（默认 `https://translation.googleapis.com/language/translate/v2`）；MyMemory 接口地址为服务基地址（不含 `/get` 查询路径，默认 `https://api.mymemory.translated.net`）。两者留空均恢复各自默认。
- 默认语言：英语 → 中文（简体），可在界面上切换或点「⇄」互换。

## 注意事项

- MyMemory 匿名调用有每日配额（约数千字符），仅适合 demo。Google Cloud Translation 按量计费（有免费额度），需自备 API Key。
- 划词翻译与窗口透明度均依赖 X11，请在 X11 / XWayland 会话下运行。
