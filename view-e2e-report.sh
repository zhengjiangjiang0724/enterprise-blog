#!/bin/bash
# 查看 E2E 测试报告
cd frontend
if [ -d "playwright-report" ]; then
  echo "正在打开 Playwright 测试报告..."
  npx playwright show-report
else
  echo "未找到测试报告，请先运行测试："
  echo "  cd frontend && npm run test:e2e"
fi
