#!/usr/bin/env bash
# 组装 macOS .app 并打成 .dmg（未签名/未公证，首次打开需右键->打开绕过 Gatekeeper）。
# 用法: package-macos.sh <binary> <version> <arch>
#   <binary>  goreleaser 编译出的 darwin 可执行文件
set -euo pipefail

BIN="${1:?binary path required}"
VERSION="${2:?version required}"
ARCH="${3:?arch required}"

APP="Gotrans.app"
OUT="dist/gotrans_${VERSION}_darwin_${ARCH}.dmg"
ICONSET="logo.iconset"
SRC_ICON="internal/ui/assets/logo.png"

rm -rf "$APP" "$ICONSET"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

cp "$BIN" "$APP/Contents/MacOS/gotrans"
chmod +x "$APP/Contents/MacOS/gotrans"

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>Gotrans</string>
  <key>CFBundleDisplayName</key><string>Gotrans</string>
  <key>CFBundleIdentifier</key><string>com.example.gotrans</string>
  <key>CFBundleVersion</key><string>${VERSION}</string>
  <key>CFBundleShortVersionString</key><string>${VERSION}</string>
  <key>CFBundleExecutable</key><string>gotrans</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleIconFile</key><string>logo</string>
  <key>LSMinimumSystemVersion</key><string>10.15</string>
  <key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST

# PNG -> icns（macOS 自带 sips/iconutil）。
mkdir -p "$ICONSET"
# 生成含 @2x 的 Retina iconset（16/32/128/256/512 各带 2 倍尺寸），供 iconutil 转 icns。
# 注意 512@2x 实为 1024，由 512 源图放大；如需锐利的大尺寸图标，源图应不小于 1024。
for s in 16 32 128 256 512; do
  sips -z "$s" "$s" "$SRC_ICON" --out "$ICONSET/icon_${s}x${s}.png" >/dev/null
  d=$((s * 2))
  sips -z "$d" "$d" "$SRC_ICON" --out "$ICONSET/icon_${s}x${s}@2x.png" >/dev/null
done
iconutil -c icns "$ICONSET" -o "$APP/Contents/Resources/logo.icns"
rm -rf "$ICONSET"

hdiutil create -volname Gotrans -srcfolder "$APP" -ov -format UDZO "$OUT"
echo "wrote $OUT"
