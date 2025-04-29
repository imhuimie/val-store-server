#!/bin/bash

# Valorant商店数据获取工具启动脚本
# 此脚本帮助编译并运行get_store_data.go文件

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# 确保bin目录存在
mkdir -p "${PROJECT_ROOT}/bin"

# 处理命令行参数
DEBUG=false
COOKIES=""
REGION="ap"
OUTPUT=""

for arg in "$@"; do
  case $arg in
    -d|--debug)
      DEBUG=true
      shift # 移除处理过的参数
      ;;
    -c=*|--cookies=*)
      COOKIES="${arg#*=}"
      shift
      ;;
    -r=*|--region=*)
      REGION="${arg#*=}"
      shift
      ;;
    -o=*|--output=*)
      OUTPUT="${arg#*=}"
      shift
      ;;
  esac
done

# 输出信息
echo -e "${BLUE}Valorant商店数据获取工具${NC}"
echo "==============================="

# 检查是否安装了Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未安装Go语言环境${NC}"
    echo "请先安装Go: https://golang.org/doc/install"
    exit 1
fi

# 编译go文件
echo -e "${YELLOW}编译脚本...${NC}"
cd "${PROJECT_ROOT}" || exit 1

if [ "$DEBUG" = true ]; then
    echo -e "${YELLOW}调试模式: 添加更多日志输出${NC}"
    # 以调试模式编译
    go build -o "${PROJECT_ROOT}/bin/vstore-get" "${SCRIPT_DIR}/get_store_data.go"
else 
    # 普通模式编译
    go build -o "${PROJECT_ROOT}/bin/vstore-get" "${SCRIPT_DIR}/get_store_data.go"
fi

# 检查编译结果
if [ $? -ne 0 ]; then
    echo -e "${RED}编译失败!${NC}"
    exit 1
fi

echo -e "${GREEN}编译成功!${NC}"
echo -e "${YELLOW}运行工具...${NC}"
echo ""

# 构建命令行参数
CMD="${PROJECT_ROOT}/bin/vstore-get"
if [ -n "$COOKIES" ]; then
    CMD="$CMD -cookies \"$COOKIES\""
fi
if [ -n "$REGION" ]; then
    CMD="$CMD -region $REGION"
fi
if [ -n "$OUTPUT" ]; then
    CMD="$CMD -output $OUTPUT"
fi

# 设置更详细的日志
if [ "$DEBUG" = true ]; then
    export VSTORE_LOG_LEVEL=debug
fi

# 运行编译后的程序，并传递所有参数
eval "$CMD" 

echo ""
echo -e "${GREEN}完成!${NC}" 