#!/bin/bash
set -e

SVG_FILE=".github/assets/hotplex-logo.svg"
OUTPUT_DIR=".github/assets"

mkdir -p $OUTPUT_DIR

echo "1/4: 渲染基础底图..."
# 针对暗色背景提取白色文字版本的 SVG (用于 OG 图和 Favicon)
TMP_SVG="$OUTPUT_DIR/logo-dark-temp.svg"
sed -e 's/class="hot-stop-1"/stop-color="#FFFFFF"/' \
    -e 's/class="hot-stop-2"/stop-color="#E6EDF3"/' \
    $SVG_FILE > $TMP_SVG

# 强制输出 1024x1024 底图
cairosvg $TMP_SVG -W 1024 -H 1024 -o $OUTPUT_DIR/logo-base.png
rm $TMP_SVG

echo "2/4: 生成 hotplex-logo.png (1024x1024, 原色)..."
cairosvg $SVG_FILE -W 1024 -H 1024 -o $OUTPUT_DIR/hotplex-logo.png

echo "3/4: 生成 Open Graph 社交预览图 (1200x630)..."
magick -size 1200x630 xc:"#0D1117" \
  \( $OUTPUT_DIR/logo-base.png -resize 600x600 \) \
  -gravity center -composite \
  $OUTPUT_DIR/hotplex-og.png

echo "4/4: 生成多尺寸 favicon.ico..."
magick -background none $OUTPUT_DIR/logo-base.png -define icon:auto-resize=256,128,64,48,32,16 $OUTPUT_DIR/favicon.ico

# 清理中间产物
rm $OUTPUT_DIR/logo-base.png

echo "完成！资产已基于 SSOT ($SVG_FILE) 重新生成。"
