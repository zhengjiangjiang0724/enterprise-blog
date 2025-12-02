import { test, expect } from '@playwright/test';

/**
 * 文章相关的E2E测试
 */
test.describe('文章功能', () => {
  test('查看文章列表', async ({ page }) => {
    await page.goto('/');
    
    // 验证文章列表存在
    await expect(page.locator('article, .article-list, [data-testid="article-list"]')).toBeVisible();
  });

  test('查看文章详情', async ({ page }) => {
    await page.goto('/');
    
    // 点击第一篇文章
    const firstArticle = page.locator('article, .article-item, [data-testid="article-item"]').first();
    await firstArticle.click();
    
    // 验证文章详情页
    await expect(page).toHaveURL(/\/articles\/.+/);
    await expect(page.locator('h1, .article-title')).toBeVisible();
  });

  test('创建文章（需要登录）', async ({ page }) => {
    // 先登录
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/$|\/articles/);
    
    // 导航到创建文章页面
    await page.goto('/articles/new');
    
    // 填写文章表单
    await page.fill('input[name="title"], textarea[name="title"]', 'E2E测试文章');
    await page.fill('textarea[name="content"], [contenteditable="true"]', '这是E2E测试创建的文章内容');
    
    // 保存为草稿
    await page.click('text=保存草稿, button:has-text("草稿")');
    
    // 验证文章创建成功
    await expect(page.locator('.success, [data-testid="success-message"]')).toBeVisible();
  });

  test('搜索文章', async ({ page }) => {
    await page.goto('/');
    
    // 查找搜索框
    const searchInput = page.locator('input[type="search"], input[placeholder*="搜索"]');
    if (await searchInput.count() > 0) {
      await searchInput.fill('测试');
      await searchInput.press('Enter');
      
      // 验证搜索结果
      await expect(page).toHaveURL(/.*search.*|.*q=.*/);
    }
  });
});

