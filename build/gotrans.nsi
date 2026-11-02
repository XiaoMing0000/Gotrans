; Gotrans Windows 安装器（NSIS）。
; 用法：makensis /DVERSION=1.2.3 /DARCH=amd64 build/gotrans.nsi
; 前置：先把 goreleaser 编译出的 gotrans.exe 复制到 dist\gotrans.exe。
Unicode true

!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!ifndef ARCH
  !define ARCH "amd64"
!endif

Name "Gotrans"
OutFile "dist\gotrans_${VERSION}_windows_${ARCH}_setup.exe"
InstallDir "$PROGRAMFILES64\Gotrans"
RequestExecutionLevel admin
SetCompressor /SOLID lzma

Page directory
Page instfiles

Section "Install"
  SetOutPath "$INSTDIR"
  File "dist\gotrans.exe"
  WriteUninstaller "$INSTDIR\uninstall.exe"
  CreateShortcut "$DESKTOP\Gotrans.lnk" "$INSTDIR\gotrans.exe"
  CreateDirectory "$SMPROGRAMS\Gotrans"
  CreateShortcut "$SMPROGRAMS\Gotrans\Gotrans.lnk" "$INSTDIR\gotrans.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Gotrans" "DisplayName" "Gotrans"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Gotrans" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Gotrans" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\gotrans.exe"
  Delete "$INSTDIR\uninstall.exe"
  Delete "$DESKTOP\Gotrans.lnk"
  RMDir /r "$SMPROGRAMS\Gotrans"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Gotrans"
  RMDir "$INSTDIR"
SectionEnd
