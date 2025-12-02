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
    await page.waitForURL(/\/register/);
    
    // 填写注册表单
    await page.fill('input[name="username"]', `e2e_test_user_${Date.now()}`);
    await page.fill('input[type="email"]', `e2e_test_${Date.now()}@example.com`);
    await page.fill('input[type="password"]', 'password123');
    
    // 提交表单
    await page.click('button[type="submit"]');
    
    // 验证跳转到个人资料页或首页
    await expect(page).toHaveURL(/\/profile|\/$|\/articles/, { timeout: 10000 });
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
    // 切换到手机登录标签（实际文本是"手机登录"）
    await page.click('text=手机登录');
    
    // 等待手机号输入框出现
    await page.waitForSelector('input[type="tel"]');
    
    // 填写手机号
    await page.fill('input[type="tel"]', '13800138000');
    
    // 点击发送验证码
    await page.click('button:has-text("发送验证码")');
    
    // 等待验证码输入框出现（实际是 type="text" 且 maxLength=6）
    await page.waitForSelector('input[type="text"][maxlength="6"]', { timeout: 5000 });
    
    // 填写验证码（测试环境，验证码会在后端日志中）
    await page.fill('input[type="text"][maxlength="6"]', '123456');
    
    // 提交登录
    await page.click('button[type="submit"]');
    
    // 验证登录成功（可能需要等待）
    await expect(page).toHaveURL(/\/$|\/articles/, { timeout: 10000 });
  });

  test('登录失败 - 错误密码', async ({ page }) => {
    await page.fill('input[type="email"]', 'test@example.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    
    // 验证显示错误消息
    await expect(page.locator('.error')).toBeVisible();
  });
});

