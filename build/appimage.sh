#!/usr/bin/env bash
# 参考脚本：把 goreleaser 编译出的 Linux 二进制打包成 AppImage。
# 未接入主 workflow（appimagetool 下载与架构名需在 CI 现场验证）。
# 用法（在 goreleaser 之后）: appimage.sh <binary> <version> <arch>
#   <arch> 取 amd64 | arm64（脚本内部映射成 x86_64 | aarch64）
set -euo pipefail

BIN="${1:?binary path required}"
VERSION="${2:?version required}"
ARCH="${3:?arch required}"

case "$ARCH" in
  amd64) APPTL=x86_64 ;;
  arm64) APPTL=aarch64 ;;
  *) echo "unknown arch: $ARCH"; exit 1 ;;
esac

APPDIR="Gotrans.AppDir"
rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin" \
         "$APPDIR/usr/share/applications" \
         "$APPDIR/usr/share/icons/hicolor/256x256/apps"

cp "$BIN" "$APPDIR/usr/bin/gotrans"
cp build/gotrans.desktop "$APPDIR/usr/share/applications/"
cp internal/ui/assets/logo.png "$APPDIR/usr/share/icons/hicolor/256x256/apps/gotrans.png"

cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
export PATH="$HERE/usr/bin:$PATH"
exec "$HERE/usr/bin/gotrans" "$@"
EOF
chmod +x "$APPDIR/AppRun"

# appimagetool 免 FUSE 运行。
curl -fsSL "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${APPTL}.AppImage" -o appimagetool
chmod +x appimagetool
APPIMAGE_EXTRACT_AND_RUN=1 ./appimagetool "$APPDIR" "dist/gotrans_${VERSION}_linux_${ARCH}.AppImage"
echo "wrote dist/gotrans_${VERSION}_linux_${ARCH}.AppImage"
