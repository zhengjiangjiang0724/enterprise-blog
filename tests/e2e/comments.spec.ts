import { test, expect } from '@playwright/test';

/**
 * 评论相关的E2E测试
 */
test.describe('评论功能', () => {
  test('查看文章评论', async ({ page }) => {
    await page.goto('/');
    
    // 点击第一篇文章
    const firstArticle = page.locator('article, .article-item').first();
    await firstArticle.click();
    
    // 滚动到评论区域
    const commentsSection = page.locator('.comments, [data-testid="comments"]');
    if (await commentsSection.count() > 0) {
      await commentsSection.scrollIntoViewIfNeeded();
      await expect(commentsSection).toBeVisible();
    }
  });

  test('发表评论（游客）', async ({ page }) => {
    await page.goto('/');
    
    // 点击第一篇文章
    const firstArticle = page.locator('article, .article-item').first();
    await firstArticle.click();
    
    // 查找评论表单
    const commentForm = page.locator('form, [data-testid="comment-form"]');
    if (await commentForm.count() > 0) {
      await commentForm.scrollIntoViewIfNeeded();
      
      // 填写评论
      await page.fill('input[name="author"], textarea[name="content"]', '测试用户');
      await page.fill('input[name="email"]', 'test@example.com');
      await page.fill('textarea[name="content"]', '这是一条E2E测试评论');
      
      // 提交评论
      await page.click('button[type="submit"]:has-text("提交"), button:has-text("发表")');
      
      // 验证评论提交成功（可能需要等待审核）
      await expect(page.locator('.success, [data-testid="success-message"]')).toBeVisible({ timeout: 5000 });
    }
  });
});

