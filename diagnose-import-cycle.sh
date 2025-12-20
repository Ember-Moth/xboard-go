#!/bin/bash

echo "========================================="
echo "Go Import Cycle Diagnostic Tool"
echo "========================================="
echo ""

# 清理缓存
echo "1. Cleaning Go cache..."
go clean -cache
echo "   ✓ Cache cleaned"
echo ""

# 检查导入关系
echo "2. Checking import relationships..."
echo ""

echo "   Checking middleware imports:"
grep -h "^import" internal/middleware/middleware.go -A 10 | head -15
echo ""

echo "   Checking if service imports middleware:"
if grep -r "dashgo/internal/middleware" internal/service/*.go > /dev/null 2>&1; then
    echo "   ❌ FOUND: service imports middleware (this causes cycle!)"
    grep -r "dashgo/internal/middleware" internal/service/*.go
else
    echo "   ✓ OK: service does not import middleware"
fi
echo ""

echo "   Checking if service imports handler:"
if grep -r "dashgo/internal/handler" internal/service/*.go > /dev/null 2>&1; then
    echo "   ❌ FOUND: service imports handler (this causes cycle!)"
    grep -r "dashgo/internal/handler" internal/service/*.go
else
    echo "   ✓ OK: service does not import handler"
fi
echo ""

# 尝试编译
echo "3. Attempting to compile..."
echo ""

if go build -o /tmp/dashgo-server ./cmd/server 2>&1 | tee /tmp/compile-output.txt; then
    echo ""
    echo "✓ Compilation successful!"
    rm -f /tmp/dashgo-server
else
    echo ""
    echo "❌ Compilation failed. Error output:"
    echo "-----------------------------------"
    cat /tmp/compile-output.txt
    echo "-----------------------------------"
    echo ""
    
    # 检查是否是循环导入错误
    if grep -q "import cycle" /tmp/compile-output.txt; then
        echo "Detected import cycle. Analyzing..."
        echo ""
        grep "import cycle" /tmp/compile-output.txt
    fi
fi

echo ""
echo "4. Checking AuthService implementation..."
if grep -q "func (s \*AuthService) GetUserFromToken" internal/service/auth.go; then
    echo "   ✓ AuthService has GetUserFromToken method"
else
    echo "   ❌ AuthService missing GetUserFromToken method"
fi

echo ""
echo "========================================="
echo "Diagnostic complete"
echo "========================================="
