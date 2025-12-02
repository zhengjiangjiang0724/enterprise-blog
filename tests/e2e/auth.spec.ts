import { test, expect } from '@playwright/test';

/**
 * 认证相关的E2E测试
 */
test.describe('用户认证', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('用户注册流程', async ({ page }) => {
    // 点击注册链接
    await page.click('text=立即注册');
    
    // 填写注册表单
    await page.fill('input[name="username"]', 'e2e_test_user');
    await page.fill('input[name="email"]', `e2e_test_${Date.now()}@example.com`);
    await page.fill('input[name="password"]', 'password123');
    
    // 提交表单
    await page.click('button[type="submit"]');
    
    // 验证跳转到首页或文章列表
    await expect(page).toHaveURL(/\/$|\/articles/);
  });

  test('用户登录流程 - 邮箱密码', async ({ page }) => {
    // 填写登录表单
    await page.fill('input[type="email"]', 'test@example.com');
    await page.fill('input[type="password"]', 'password123');
    
    // 提交表单
    await page.click('button[type="submit"]');
    
    // 验证登录成功（跳转到首页或显示用户信息）
    await expect(page).toHaveURL(/\/$|\/articles/);
  });

  test('用户登录流程 - 手机验证码', async ({ page }) => {
    // 切换到手机登录标签
    await page.click('text=手机号登录');
    
    // 填写手机号
    await page.fill('input[type="tel"]', '13800138000');
    
    // 点击发送验证码
    await page.click('text=发送验证码');
    
    // 等待验证码输入框出现
    await page.waitForSelector('input[pattern="[0-9]{6}"]');
    
    // 填写验证码（测试环境，验证码会在后端日志中）
    await page.fill('input[pattern="[0-9]{6}"]', '123456');
    
    // 提交登录
    await page.click('button[type="submit"]');
    
    // 验证登录成功
    await expect(page).toHaveURL(/\/$|\/articles/);
  });

  test('登录失败 - 错误密码', async ({ page }) => {
    await page.fill('input[type="email"]', 'test@example.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    
    // 验证显示错误消息
    await expect(page.locator('.error')).toBeVisible();
  });
});

