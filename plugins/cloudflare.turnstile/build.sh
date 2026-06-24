#!/usr/bin/env bash
set -euo pipefail

PLUGIN_ID="cloudflare.turnstile"
VERSION="1.0.0"
BINARY_BASE="cloudflare-turnstile"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$SCRIPT_DIR/releases/$VERSION"
DIST_DIR="$SCRIPT_DIR/dist"
PACKAGE_DIR="$DIST_DIR/packages"

TARGET_OS=()
TARGET_ARCH=()
CLEAN=false
SKIP_FRONTEND=false

for arg in "$@"; do
  case "$arg" in
    --windows|--win|-w) TARGET_OS+=("windows") ;;
    --linux|-l) TARGET_OS+=("linux") ;;
    --mac|--darwin|-m) TARGET_OS+=("darwin") ;;
    --all|-a) TARGET_OS=("windows" "linux" "darwin") ;;
    --arm) TARGET_ARCH+=("arm64") ;;
    --x64|--amd64) TARGET_ARCH+=("amd64") ;;
    --clean|clean) CLEAN=true ;;
    --no-build-front|--skip-frontend) SKIP_FRONTEND=true ;;
  esac
done

cd "$SCRIPT_DIR"

if [ "$CLEAN" = true ]; then
  rm -rf "$RELEASE_DIR/bin" "$RELEASE_DIR/frontend" "$DIST_DIR"
  printf '{}' > "$RELEASE_DIR/checksums.json"
  echo "[SUCCESS] 清理完成"
  exit 0
fi

if [ ${#TARGET_OS[@]} -eq 0 ]; then TARGET_OS=("linux"); fi
if [ ${#TARGET_ARCH[@]} -eq 0 ]; then TARGET_ARCH=("amd64"); fi

echo "[INFO] Go: $(go version)"
go mod tidy

for os in "${TARGET_OS[@]}"; do
  for arch in "${TARGET_ARCH[@]}"; do
    platform_dir="${os}_${arch}"
    binary_name="$BINARY_BASE"
    if [ "$os" = "windows" ]; then binary_name="$binary_name.exe"; fi
    output_dir="$RELEASE_DIR/bin/$platform_dir"
    output_file="$output_dir/$binary_name"
    rm -rf "$output_dir"
    mkdir -p "$output_dir"
    echo "[INFO] 构建 $os/$arch"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w" -o "$output_file" ./src
  done
done

echo "[INFO] 构建插件前端（Next.js 静态导出）"
FRONTEND_SRC_DIR="$SCRIPT_DIR/frontend"
rm -rf "$RELEASE_DIR/frontend"
mkdir -p "$RELEASE_DIR/frontend"

if [ "$SKIP_FRONTEND" = true ]; then
  echo "[WARNING] 已跳过前端构建（--skip-frontend）；releases/frontend 为空，需确保已有产物或后续补构建"
elif [ ! -f "$FRONTEND_SRC_DIR/package.json" ]; then
  echo "[ERROR] 未找到前端工程 frontend/package.json" >&2
  exit 1
else
  if ! command -v npm >/dev/null 2>&1; then
    echo "[ERROR] 未找到 npm，无法构建前端；可在具备 Node 环境时重新构建，或使用 --skip-frontend 仅构建后端" >&2
    exit 1
  fi
  echo "[INFO] 安装前端依赖（npm ci）"
  (cd "$FRONTEND_SRC_DIR" && npm ci)
  echo "[INFO] 执行前端静态导出（next build）"
  (cd "$FRONTEND_SRC_DIR" && npm run build)

  FRONTEND_OUT_DIR="$FRONTEND_SRC_DIR/out"
  if [ ! -f "$FRONTEND_OUT_DIR/widget.html" ]; then
    echo "[ERROR] 未找到前端构建产物 out/widget.html" >&2
    exit 1
  fi
  cp "$FRONTEND_OUT_DIR/widget.html" "$RELEASE_DIR/frontend/widget.html"
  # 仅拷贝 widget.html 引用的 _next 静态资源目录（JS/CSS），其余 Next 默认产物（404/_not-found）不纳入插件包
  if [ -d "$FRONTEND_OUT_DIR/_next" ]; then
    cp -R "$FRONTEND_OUT_DIR/_next" "$RELEASE_DIR/frontend/_next"
  fi
  echo "[SUCCESS] 前端产物已输出到 $RELEASE_DIR/frontend"
fi

echo "[INFO] 生成 checksums.json"
python - "$RELEASE_DIR" <<'PY'
import hashlib, json, pathlib, sys
root = pathlib.Path(sys.argv[1])
checksums = {}
for path in sorted(root.rglob("*")):
    if path.is_file() and path.name != "checksums.json":
        checksums[path.relative_to(root).as_posix()] = hashlib.sha256(path.read_bytes()).hexdigest()
(root / "checksums.json").write_text(json.dumps(checksums, ensure_ascii=False, indent=2), encoding="utf-8")
PY

mkdir -p "$PACKAGE_DIR"
for os in "${TARGET_OS[@]}"; do
  for arch in "${TARGET_ARCH[@]}"; do
    package_path="$PACKAGE_DIR/$PLUGIN_ID-$VERSION-${os}_${arch}.ksplugin.zip"
    stage_dir="$DIST_DIR/stage/${os}_${arch}"
    rm -rf "$stage_dir"
    mkdir -p "$stage_dir/$PLUGIN_ID"
    cp -R "$SCRIPT_DIR/releases" "$stage_dir/$PLUGIN_ID/"
    rm -f "$package_path"
    (cd "$stage_dir" && python -m zipfile -c "$package_path" "$PLUGIN_ID")
    echo "[SUCCESS] 已打包: $package_path"
  done
done

echo "[SUCCESS] 插件构建完成"
