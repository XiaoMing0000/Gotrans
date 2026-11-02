# Gotrans

一个用 Go + Fyne 编写的桌面翻译工具，界面类似 Google 翻译。支持免费的 MyMemory 与 Google Cloud Translation，架构可插拔。

## 功能

- 基础翻译：原文 / 译文双栏、源/目标语言选择、⇄ 互换、复制译文、清空输入。原文与译文之间、主区域与划词栏之间均有可拖拽分隔条，位置会自动记忆，下次启动无需再次调整。
- 划词翻译：在系统其它应用中框选文字（X11 PRIMARY 选区），原文显示在输入框下方、译文显示在输出框下方（最小高度固定，超长内容滚动），可一键清空。勾选划词翻译时，可拖拽分隔条调整划词栏高度。译文区右键菜单可「仅显示划词译文窗口」：隐藏主窗口、弹出只含译文的无边框悬浮窗，译文区采用与主窗口一致的圆角输入框样式（圆角底框 + 边框 + 输入底色），窗口四角圆角，继承主窗口的窗口/字体透明度；按住内容任意处拖拽移动，鼠标移到窗口边缘拖动即可缩放，右键菜单（复制/置顶/恢复主窗口）返回主界面。置顶后悬浮窗始终位于其它窗口之上，再次右键可「取消置顶」还原堆叠。
- 设置：切换翻译引擎（MyMemory / Google Translate / 自定义引擎）、填写 Google API Key、自定义 Google / MyMemory 接口地址、启用 / 禁用划词翻译（关闭时隐藏划词栏、主区域全高度填充；含子选项：显示原文，隐藏原文后译文全宽填充）、调整窗口透明度与字体透明度（下限均 20%，二者独立：窗口透明度只作用于背景/输入底色，不影响文字）。自定义引擎可命名，每个保存独立的接口地址与 API Key（Google 兼容接口），并在引擎下拉中直接删除；删除当前引擎会自动切回 MyMemory。配置自动保存到用户目录，下次启动自动记忆。

## 构建

本项目使用 X11 后端以保证划词翻译和窗口透明度正常工作，X11 专属代码已用 build tag 隔离：**Linux 构建需加 `-tags x11`**；在 **Windows / macOS 上直接 `go build` 即可**，划词翻译与悬浮窗高级操作（拖拽 / 缩放 / 置顶 / 圆角）会自动降级为不可用，翻译主功能照常。

**Linux** 先安装 X11 / OpenGL 开发库（RHEL / AlmaLinux / Fedora）：

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

## 发布（CI/CD）

推送 `v*` 标签（或手动触发）会由 GitHub Actions（`.github/workflows/release.yml`）自动构建并发布一个 Release：

- **Linux**（amd64 / arm64）：在 AlmaLinux 9 容器内构建，产出 tar.gz、`.deb`、`.rpm`（glibc 下限 2.34，可装 EL9）。
- **Windows**（amd64）：zip（含 exe）+ NSIS 安装器 `.exe`。
- **macOS**（amd64 / arm64）：tar.gz + `.app` / `.dmg`（未签名 / 公证，首次打开需右键 → 打开绕过 Gatekeeper）。

所有产物统一生成 `SHA256SUMS`。Flatpak 清单在 `flatpak/`、AppImage 脚本在 `build/appimage.sh`，二者为参考实现，暂未接入主流程，可按需迭代后启用。

## 配置

- 翻译引擎与密钥通过「设置」窗口配置：可选 `MyMemory`（免 Key）、`Google Translate`（需在 [Google Cloud](https://cloud.google.com/translate) 申请 API Key），或在引擎下拉的「自定义」项里添加自定义引擎（Google 兼容接口，保存名称 + 接口地址 + API Key）。设置里只展示当前选中引擎的接口地址与 API Key 输入框。
- `GOTRANS_PROVIDER`：环境变量兜底选择引擎（`mymemory` / `google`），仅当配置文件中未保存引擎时生效；未知值会回退到默认引擎。
- 界面设置（翻译引擎、Google API Key、Google / MyMemory 接口地址、划词开关、窗口/字体透明度）保存在 `~/.config/gotrans/config.json`，下次启动自动加载。Google 接口地址为完整端点（默认 `https://translation.googleapis.com/language/translate/v2`）；MyMemory 接口地址为服务基地址（不含 `/get` 查询路径，默认 `https://api.mymemory.translated.net`）。两者留空均恢复各自默认。
- 默认语言：英语 → 中文（简体），可在界面上切换或点「⇄」互换。

## 注意事项

- MyMemory 匿名调用有每日配额（约数千字符），仅适合 demo。Google Cloud Translation 按量计费（有免费额度），需自备 API Key。
- 划词翻译依赖 X11（PRIMARY 选区），请在 X11 / XWayland 会话下运行。窗口透明度依赖带 alpha 通道的透明窗口（由 GLFW 透明帧缓冲 + 桌面合成器合成），文字透明度走主题、独立于窗口透明度。
- 「仅显示划词译文窗口」的无边框、移动、缩放与置顶均依赖窗口管理器支持 EWMH（`_MOTIF_WM_HINTS` 去边框、`ConfigureWindow` 移动/缩放、`_NET_WM_STATE_ABOVE` 置顶）；个别 WM（如部分平铺 WM）可能忽略去边框、几何请求或置顶状态，此时悬浮窗仍会保留标题栏、无法通过拖拽移动/缩放或置顶。悬浮窗四角圆角依赖 X11 SHAPE 扩展（`Shape.Rectangles`）裁剪窗口；个别 WM 可能忽略形状裁剪，此时窗口四角保持直角。
