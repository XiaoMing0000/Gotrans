; Gotrans Windows 安装器（NSIS）。
; 用法：makensis /DVERSION=1.2.3 /DARCH=amd64 /DEXE=<exe绝对路径> /DOUT=<setup.exe绝对路径> build/gotrans.nsi
; EXE/OUT 用绝对路径传入，避免相对路径随 makensis 工作目录变化而解析错位。
Unicode true

!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!ifndef ARCH
  !define ARCH "amd64"
!endif
!ifndef EXE
  !define EXE "dist\gotrans.exe"
!endif
!ifndef OUT
  !define OUT "dist\gotrans_${VERSION}_windows_${ARCH}_setup.exe"
!endif

Name "Gotrans"
OutFile "${OUT}"
InstallDir "$PROGRAMFILES64\Gotrans"
RequestExecutionLevel admin
SetCompressor /SOLID lzma

Page directory
Page instfiles

Section "Install"
  SetOutPath "$INSTDIR"
  File /oname=gotrans.exe "${EXE}"
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
