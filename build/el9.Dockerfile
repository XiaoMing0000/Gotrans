# AlmaLinux 9 构建镜像：用于在 CI 里编译 Linux 版并把 glibc 下限压到 2.34，
# 保证产出的二进制与 .rpm 能跑在 EL9/EL10 及较新的发行版上。
#
# 构建（amd64 / arm64）：
#   docker build -f build/el9.Dockerfile --build-arg TARGETARCH=amd64 -t gotrans-el9 .
# 使用：
#   docker run --rm -v "$PWD":/src -w /src gotrans-el9 \
#     release --clean --skip=publish -f .goreleaser-linux-<arch>.yaml
FROM almalinux:9

# X11/OpenGL 开发库（与 README 一致，glfw cgo 需要）+ 打包工具。
# 注意：libxkbcommon-x11-devel、libXxf86vm-devel 只在 CRB(CodeReady Builder) 仓库，
# 而 AlmaLinux 9 默认未启用 CRB，必须先打开再安装。
RUN dnf install -y --setopt=install_weak_deps=False dnf-plugins-core \
  && dnf config-manager --set-enabled crb \
  && dnf install -y --setopt=install_weak_deps=False \
    gcc git ca-certificates curl tar \
    pkgconf-pkg-config \
    libxkbcommon-devel libxkbcommon-x11-devel \
    libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel \
    libXxf86vm-devel libXfixes-devel libXi-devel libXext-devel \
    libXrender-devel libXft-devel mesa-libGL-devel \
  && dnf clean all

# 容器以 root 运行、bind-mount 的 /src 归宿主机 uid，规避 git "dubious ownership" 报错。
RUN git config --global --add safe.directory /src

# 安装 Go 工具链（版本与项目 go.mod 的 go 指令一致）。
ARG GO_VERSION=1.26.10
ARG TARGETARCH=amd64
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" -o /tmp/go.tgz \
  && tar -C /usr/local -xzf /tmp/go.tgz \
  && rm /tmp/go.tgz
ENV PATH="/usr/local/go/bin:${PATH}" \
    GOTOOLCHAIN=local

# 安装 GoReleaser 预编译二进制（不依赖容器内 Go 版本）。
ARG GORELEASER_VERSION=2.19.0
RUN set -eux; \
    case "${TARGETARCH}" in \
      amd64) GR_ARCH=x86_64 ;; \
      arm64) GR_ARCH=arm64 ;; \
      *) GR_ARCH=x86_64 ;; \
    esac; \
    curl -fsSL "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_${GR_ARCH}.tar.gz" -o /tmp/gr.tgz \
    && tar -C /usr/local/bin -xzf /tmp/gr.tgz goreleaser \
    && rm /tmp/gr.tgz \
    && goreleaser --version

ENTRYPOINT ["goreleaser"]
