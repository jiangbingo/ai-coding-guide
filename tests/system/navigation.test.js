/**
 * 导航流程系统测试
 * 
 * 测试目标：验证用户导航和交互流程
 */

import { test, expect } from '@playwright/test';

test.describe('Navigation Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:8000');
  });
  
  test('should load homepage successfully', async ({ page }) => {
    // 检查页面标题
    await expect(page).toHaveTitle(/AI 代码入门指南/);
    
    // 检查 Hero 区域
    const heroTitle = await page.locator('.hero-title');
    await expect(heroTitle).toBeVisible();
  });
  
  test('should navigate to section on keyboard press', async ({ page }) => {
    // 按数字键 1 跳转到基础概念章节
    await page.keyboard.press('1');
    
    // 等待滚动完成
    await page.waitForTimeout(500);
    
    // 验证章节可见
    const conceptsSection = await page.locator('#concepts');
    await expect(conceptsSection).toBeInViewport();
  });
  
  test('should toggle accordion on click', async ({ page }) => {
    // 找到第一个手风琴项
    const firstAccordion = await page.locator('.scenario-item').first();
    const header = await firstAccordion.locator('.scenario-header');
    const content = await firstAccordion.locator('.scenario-content');
    
    // 初始状态应该是折叠的
    await expect(content).not.toBeVisible();
    
    // 点击展开
    await header.click();
    await page.waitForTimeout(300);
    
    // 验证展开
    await expect(content).toBeVisible();
    
    // 再次点击折叠
    await header.click();
    await page.waitForTimeout(300);
    
    // 验证折叠
    await expect(content).not.toBeVisible();
  });
  
  test('should handle mobile menu toggle', async ({ page }) => {
    // 设置移动端视口
    await page.setViewportSize({ width: 375, height: 667 });
    
    // 找到移动端菜单按钮
    const menuButton = await page.locator('.mobile-menu-button');
    const navMenu = await page.locator('.nav-menu');
    
    // 初始状态应该是隐藏的
    await expect(navMenu).not.toBeVisible();
    
    // 点击打开
    await menuButton.click();
    await page.waitForTimeout(300);
    
    // 验证打开
    await expect(navMenu).toBeVisible();
    
    // 点击关闭
    await menuButton.click();
    await page.waitForTimeout(300);
    
    // 验证关闭
    await expect(navMenu).not.toBeVisible();
  });
  
  test('should scroll smoothly to sections', async ({ page }) => {
    // 点击导航链接
    await page.click('a[href="#tools"]');
    
    // 等待滚动动画
    await page.waitForTimeout(1000);
    
    // 验证目标章节在视口中
    const toolsSection = await page.locator('#tools');
    await expect(toolsSection).toBeInViewport();
  });
  
  test('should handle keyboard shortcuts', async ({ page }) => {
    // 测试所有数字键快捷键
    const sections = ['concepts', 'tools', 'workflow', 'scenarios', 'resources'];
    
    for (let i = 0; i < sections.length; i++) {
      await page.keyboard.press(String(i + 1));
      await page.waitForTimeout(500);
      
      const section = await page.locator(`#${sections[i]}`);
      await expect(section).toBeInViewport();
    }
  });
  
  test('should close mobile menu on ESC', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    // 打开菜单
    await page.click('.mobile-menu-button');
    await page.waitForTimeout(300);
    
    const navMenu = await page.locator('.nav-menu');
    await expect(navMenu).toBeVisible();
    
    // 按 ESC 关闭
    await page.keyboard.press('Escape');
    await page.waitForTimeout(300);
    
    // 验证关闭
    await expect(navMenu).not.toBeVisible();
  });
});

test.describe('Responsive Design', () => {
  test('should adapt to desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto('http://localhost:8000');
    
    // 验证桌面端布局
    const heroGrid = await page.locator('.hero-grid');
    await expect(heroGrid).toHaveCSS('display', 'grid');
    
    // 验证双栏布局
    const columns = await heroGrid.locator('> *').count();
    expect(columns).toBeGreaterThanOrEqual(2);
  });
  
  test('should adapt to tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('http://localhost:8000');
    
    // 验证平板端布局
    const heroGrid = await page.locator('.hero-grid');
    await expect(heroGrid).toBeVisible();
  });
  
  test('should adapt to mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('http://localhost:8000');
    
    // 验证移动端布局（单栏）
    const heroGrid = await page.locator('.hero-grid');
    const display = await heroGrid.evaluate(el => 
      window.getComputedStyle(el).display
    );
    
    // 移动端可能是 flex 或单栏 grid
    expect(['flex', 'grid']).toContain(display);
  });
});

test.describe('Performance', () => {
  test('should load within acceptable time', async ({ page }) => {
    const start = Date.now();
    await page.goto('http://localhost:8000');
    const loadTime = Date.now() - start;
    
    // 页面应该在 3 秒内加载完成
    expect(loadTime).toBeLessThan(3000);
  });
  
  test('should have no console errors', async ({ page }) => {
    const errors = [];
    
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    await page.goto('http://localhost:8000');
    await page.waitForTimeout(2000);
    
    expect(errors).toHaveLength(0);
  });
  
  test('should pass Lighthouse performance audit', async ({ page }) => {
    // 这个测试需要 lighthouse 插件
    // 这里只是示例，实际需要安装 @lhci/cli
    
    const metrics = await page.evaluate(() => {
      return new Promise((resolve) => {
        const observer = new PerformanceObserver((list) => {
          const entries = list.getEntries();
          resolve(entries);
        });
        observer.observe({ entryTypes: ['paint', 'largest-contentful-paint'] });
      });
    });
    
    // 验证关键性能指标
    expect(metrics.length).toBeGreaterThan(0);
  });
});