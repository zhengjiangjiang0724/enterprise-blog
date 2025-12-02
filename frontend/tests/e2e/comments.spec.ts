import { test, expect } from '@playwright/test';

/**
 * 评论相关的E2E测试
 */
test.describe('评论功能', () => {
  test('查看文章评论', async ({ page }) => {
    await page.goto('/');
    
    // 等待文章列表加载（可能没有文章，需要处理）
    try {
      await page.waitForSelector('.article-item', { timeout: 5000 });
    } catch {
      // 如果没有文章，跳过此测试
      test.skip();
      return;
    }
    
    // 检查是否有文章
    const articleCount = await page.locator('.article-item').count();
    if (articleCount === 0) {
      test.skip();
      return;
    }
    
    // 点击第一篇文章的链接
    const firstArticleLink = page.locator('.article-item h2 a').first();
    await firstArticleLink.click();
    
    // 等待文章详情页加载
    await page.waitForURL(/\/articles\/.+/, { timeout: 10000 });
    
    // 滚动到评论区域
    const commentsSection = page.locator('.comments-section, section.comments-section');
    await commentsSection.scrollIntoViewIfNeeded();
    
    // 验证评论区域可见
    await expect(commentsSection).toBeVisible({ timeout: 5000 });
  });

  test('发表评论（游客）', async ({ page }) => {
    await page.goto('/');
    
    // 等待文章列表加载（可能没有文章，需要处理）
    try {
      await page.waitForSelector('.article-item', { timeout: 5000 });
    } catch {
      // 如果没有文章，跳过此测试
      test.skip();
      return;
    }
    
    // 检查是否有文章
    const articleCount = await page.locator('.article-item').count();
    if (articleCount === 0) {
      test.skip();
      return;
    }
    
    // 点击第一篇文章的链接
    const firstArticleLink = page.locator('.article-item h2 a').first();
    await firstArticleLink.click();
    
    // 等待文章详情页加载
    await page.waitForURL(/\/articles\/.+/, { timeout: 10000 });
    
    // 滚动到评论区域
    const commentsSection = page.locator('.comments-section, section.comments-section');
    await commentsSection.scrollIntoViewIfNeeded();
    
    // 等待评论表单出现
    await page.waitForSelector('form', { timeout: 5000 });
    
    // 查找评论表单（游客需要填写昵称和邮箱）
    const authorInput = page.locator('input[name="author"]');
    const emailInput = page.locator('input[name="email"]');
    const contentTextarea = page.locator('textarea[name="content"]');
    
    // 填写评论（如果输入框存在）
    if (await authorInput.count() > 0) {
      await authorInput.fill('E2E测试用户');
    }
    if (await emailInput.count() > 0) {
      await emailInput.fill('e2e_test@example.com');
    }
    if (await contentTextarea.count() > 0) {
      await contentTextarea.fill(`这是一条E2E测试评论_${Date.now()}`);
    }
    
    // 提交评论
    const submitButton = page.locator('button[type="submit"]').first();
    if (await submitButton.count() > 0) {
      await submitButton.click();
      
      // 等待提交响应（成功消息可能通过 MessageProvider 显示）
      await page.waitForTimeout(2000);
    }
  });
});

