#!/bin/bash

APP_NAME="app"

LANGUAGES=("English" "中文")
EN_MESSAGES=(
 "Select build option:"
 "1) 64-bit only (default)"
 "2) 32-bit only"
 "3) Both 32-bit and 64-bit"
 "Enter choice [1-3]: "
 "Error: Go is not installed. Please install Go first."
 "Error: go.mod not found. Please run 'go mod init' first."
 "Error: go.mod does not contain toml dependency."
 "Running go mod tidy..."
 "go mod tidy completed successfully."
 "go mod tidy failed."
 "Building for"
 "Success"
 "Failed"
 "Build complete!"
 "Files generated:"
 "Size:"
 "Invalid choice. Using default (64-bit only)."
)

ZH_MESSAGES=(
 "选择编译选项："
 "1) 仅64位（默认）"
 "2) 仅32位"
 "3) 32位和64位都编译"
 "输入选择 [1-3]："
 "错误：未安装Go。请先安装Go。"
 "错误：未找到go.mod。请先运行 'go mod init'。"
 "错误：go.mod未包含toml依赖。"
 "正在运行 go mod tidy..."
 "go mod tidy 完成。"
 "go mod tidy 失败。"
 "正在编译"
 "成功"
 "失败"
 "编译完成！"
 "生成的文件："
 "大小："
 "无效选择，使用默认选项（仅64位）。"
)

select_language() {
 echo "Select language / 选择语言："
 echo "1) English"
 echo "2) 中文"
 read -p "Enter 1 or 2: " lang_choice
 
 if [ "$lang_choice" = "2" ]; then
   LANG_INDEX=1
 else
   LANG_INDEX=0
 fi
}

get_msg() {
 local index=$1
 if [ "$LANG_INDEX" -eq 1 ]; then
   echo "${ZH_MESSAGES[$index]}"
 else
   echo "${EN_MESSAGES[$index]}"
 fi
}

check_go() {
 if ! command -v go &> /dev/null; then
   echo "$(get_msg 5)"
   exit 1
 fi
}

check_gomod() {
 if [ ! -f "go.mod" ]; then
   echo "$(get_msg 6)"
   exit 1
 fi
 
 if ! grep -q "github.com/BurntSushi/toml" go.mod; then
   echo "$(get_msg 7)"
   exit 1
 fi
}

run_gomod_tidy() {
 echo "$(get_msg 8)"
 if go mod tidy; then
   echo "$(get_msg 9)"
 else
   echo "$(get_msg 10)"
   exit 1
 fi
}

select_build_option() {
 echo ""
 echo "$(get_msg 0)"
 echo "$(get_msg 1)"
 echo "$(get_msg 2)"
 echo "$(get_msg 3)"
 read -p "$(get_msg 4)" choice
 
 if [ -z "$choice" ]; then
   choice="1"
 fi
 
 case $choice in
   1)
     BUILD_64=true
     BUILD_32=false
     ;;
   2)
     BUILD_64=false
     BUILD_32=true
     ;;
   3)
     BUILD_64=true
     BUILD_32=true
     ;;
   *)
     echo "$(get_msg 17)"
     BUILD_64=true
     BUILD_32=false
     ;;
 esac
}

build_platform() {
 local goos=$1
 local goarch=$2
 local output_name="${APP_NAME}_${goos}_${goarch}"
 
 if [ "$goos" = "windows" ]; then
   output_name="${output_name}.exe"
 fi
 
 echo "$(get_msg 11) $goos/$goarch..."
 
 if env GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -ldflags="-s -w" -o "./$output_name"; then
   echo "  $(get_msg 12)"
   local size=$(ls -lh "./$output_name" 2>/dev/null | awk '{print $5}')
   echo "  $(get_msg 15) $size"
 else
   echo "  $(get_msg 13)"
 fi
}

main() {
 select_language
 check_go
 check_gomod
 run_gomod_tidy
 select_build_option
 
 local platforms=()
 
 if [ "$BUILD_64" = true ]; then
   platforms+=("windows/amd64")
   platforms+=("linux/amd64")
   platforms+=("linux/arm64")
   platforms+=("darwin/amd64")
   platforms+=("darwin/arm64")
   platforms+=("android/arm64")
   platforms+=("freebsd/amd64")
 fi
 
 if [ "$BUILD_32" = true ]; then
   platforms+=("windows/386")
   platforms+=("linux/386")
   platforms+=("linux/arm")
   platforms+=("android/arm")
   platforms+=("freebsd/386")
 fi
 
 for platform in "${platforms[@]}"; do
   IFS='/' read -r goos goarch <<< "$platform"
   build_platform "$goos" "$goarch"
   echo ""
 done
 
 echo "$(get_msg 14)"
 echo "$(get_msg 15)"
 ls -lh ./ 2>/dev/null || echo "No files found"
}

main