#!/bin/bash

# 验证安装脚本的正确性

echo "=== 验证 dashGO 安装脚本 ==="
echo ""

# 1. 检查脚本语法
echo "1. 检查主安装脚本语法..."
if bash -n install.sh 2>&1; then
    echo "   ✓ 主脚本语法正确"
else
    echo "   ✗ 主脚本语法错误"
    exit 1
fi

echo ""
echo "2. 检查 Agent 安装脚本语法..."
if bash -n agent/install.sh 2>&1; then
    echo "   ✓ Agent 脚本语法正确"
else
    echo "   ✗ Agent 脚本语法错误"
    exit 1
fi

echo ""
echo "3. 检查关键变量..."
echo "   主脚本:"
grep "^GITHUB_REPO=" install.sh
grep "^INSTALL_DIR=" install.sh
grep "^AGENT_DIR=" install.sh
grep "^TEMP_DIR=" install.sh

echo ""
echo "   Agent 脚本:"
grep "^GITHUB_REPO=" agent/install.sh
grep "^INSTALL_DIR=" agent/install.sh
grep "^SERVICE_NAME=" agent/install.sh
grep "^TEMP_DIR=" agent/install.sh

echo ""
echo "4. 检查下载 URL..."
echo "   主脚本中的下载 URL:"
grep -n "download.sharon.wiki" install.sh | head -5

echo ""
echo "   Agent 脚本中的下载 URL:"
grep -n "download.sharon.wiki" agent/install.sh

echo ""
echo "5. 检查文件命名..."
echo "   主脚本中的二进制文件名:"
grep -n "dashgo-server\|dashgo-agent" install.sh | head -10

echo ""
echo "   Agent 脚本中的二进制文件名:"
grep -n "dashgo-agent" agent/install.sh | head -5

echo ""
echo "6. 检查目录命名..."
echo "   主脚本中的目录引用:"
grep -n "dashgo-main" install.sh | head -5

echo ""
echo "=== 验证完成 ==="
echo ""
echo "如果所有检查都通过，脚本应该可以正常工作。"
echo "如果仍然遇到问题，请运行 debug-install.sh 进行详细调试。"