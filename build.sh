#!/bin/bash

# 停止脚本在遇到错误时
set -e

# 获取脚本所在的项目目录，避免依赖执行脚本时的当前目录。
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 帮助信息
usage() {
  echo "Usage: $0 -d <target-dir> [--all] [--no-git] [-m <module>]"
  echo "  -d, --target-dir <target-dir>: 必填，本次构建产物的目标目录"
  echo "  --all: 构建全部模块"
  echo "  --no-git: 跳过git拉取代码"
  echo "  -m <module>: 指定一个或多个模块 (web, serve)，用逗号分隔"
}

# 解析命令行参数
BUILD_WEB=false
BUILD_SERVICE=false
BUILD_ALL=false
GIT_UPDATE=true
TARGET_DIR=""

while [[ "$#" -gt 0 ]]; do
  case $1 in
    -d|--target-dir)
      if [[ "$#" -lt 2 || -z "$2" ]]; then
        echo "错误：$1 后必须提供目标目录。"
        usage
        exit 1
      fi
      TARGET_DIR="$2"
      shift
      ;;
    --all)
      BUILD_ALL=true
      ;;
    -m)
      if [[ "$#" -lt 2 || -z "$2" ]]; then
        echo "错误：-m 后必须提供模块名称。"
        usage
        exit 1
      fi
      MODULES=$2
      shift
      IFS=',' read -ra MODULE_ARR <<< "$MODULES"
      for MODULE in "${MODULE_ARR[@]}"; do
        case $MODULE in
          web)
            BUILD_WEB=true
            ;;
          serve)
            BUILD_SERVICE=true
            ;;
          *)
            echo "Invalid module: $MODULE"
            usage
            exit 1
            ;;
        esac
      done
      ;;
    --no-git)
      GIT_UPDATE=false
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown parameter passed: $1"
      usage
      exit 1
      ;;
  esac
  shift
done

# 目标目录必须由本次执行显式传入。
if [[ -z "$TARGET_DIR" ]]; then
  echo "错误：缺少必填参数 -d/--target-dir。"
  usage
  exit 1
fi

# 判断当前是否以sudo/root权限执行。
if [ "$EUID" -ne 0 ]; then
  echo "请使用 sudo 或 root 权限运行此脚本。"
  exit 1
fi

# 如果传入 --all 参数，则构建全部模块
if $BUILD_ALL; then
  BUILD_WEB=true
  BUILD_SERVICE=true
fi

# 如果没有传入任何模块参数，默认构建 web 和 serve
if ! $BUILD_WEB && ! $BUILD_SERVICE && ! $BUILD_ALL; then
  BUILD_WEB=true
  BUILD_SERVICE=true
fi

# 检查目标目录是否存在，如果不存在则创建
if [ ! -d "$TARGET_DIR" ]; then
  echo "目标目录 $TARGET_DIR 不存在，正在创建..."
  mkdir -p "$TARGET_DIR"
fi
TARGET_DIR="$(cd "$TARGET_DIR" && pwd -P)"
echo "本次构建目标目录：$TARGET_DIR"

# 进入项目目录并且更新代码
cd "$PROJECT_DIR"
if $GIT_UPDATE; then
  echo "更新代码..."
  git checkout -- . && git pull
fi

if $BUILD_SERVICE; then
  echo "初始化go依赖包..."
  cd "$PROJECT_DIR"/server && go generate
  echo "生成二进制文件..."
  go build -o server main.go
  echo "拷贝资源文件..."
  cp config.yaml $TARGET_DIR
  cp -r resource $TARGET_DIR
  echo "临时终止守护进程"
  supervisorctl stop order_food_server
  echo "拷贝go核心文件到指定目录..."
  cp server $TARGET_DIR
  echo "拷贝完成，启动守护进程..."
  supervisorctl start order_food_server
fi

if $BUILD_WEB; then
  cd "$PROJECT_DIR"/web
  echo "npm更新依赖..."
  npm i
  echo "npm打包资源文件..."
  npm run build
  echo "拷贝资源文件..."
  cp -r dist $TARGET_DIR
fi

echo "构建成功！"
