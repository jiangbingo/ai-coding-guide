/**
 * 打字机效果单元测试
 * 
 * 测试目标：验证打字机动画功能
 */

describe('Typewriter Effect', () => {
  let element;
  
  beforeEach(() => {
    // 设置 DOM 环境
    document.body.innerHTML = '<span id="test-element"></span>';
    element = document.getElementById('test-element');
  });
  
  afterEach(() => {
    document.body.innerHTML = '';
  });
  
  describe('Basic Functionality', () => {
    test('should type text character by character', async () => {
      const text = 'Hello';
      const speed = 10; // 快速测试
      
      await typewriterEffect(element, text, speed);
      
      expect(element.textContent).toBe(text);
    });
    
    test('should handle empty string', async () => {
      await typewriterEffect(element, '', 10);
      
      expect(element.textContent).toBe('');
    });
    
    test('should handle single character', async () => {
      await typewriterEffect(element, 'A', 10);
      
      expect(element.textContent).toBe('A');
    });
    
    test('should handle special characters', async () => {
      const text = 'Hello\nWorld!';
      await typewriterEffect(element, text, 10);
      
      expect(element.textContent).toBe(text);
    });
  });
  
  describe('Timing', () => {
    test('should respect speed parameter', async () => {
      const text = 'Test';
      const speed = 50;
      
      const start = Date.now();
      await typewriterEffect(element, text, speed);
      const elapsed = Date.now() - start;
      
      // 4 个字符，每个 50ms，总时间应该 >= 200ms
      expect(elapsed).toBeGreaterThanOrEqual(200);
      expect(elapsed).toBeLessThan(300); // 允许一些误差
    });
  });
  
  describe('Edge Cases', () => {
    test('should handle null element gracefully', async () => {
      // 不应该抛出错误
      await expect(typewriterEffect(null, 'Test', 10)).resolves.not.toThrow();
    });
    
    test('should handle undefined text', async () => {
      await typewriterEffect(element, undefined, 10);
      
      expect(element.textContent).toBe('');
    });
    
    test('should stop typing if element is removed', async () => {
      const text = 'Long text for testing';
      const promise = typewriterEffect(element, text, 100);
      
      // 中途移除元素
      setTimeout(() => {
        element.remove();
      }, 150);
      
      // 应该优雅地结束
      await expect(promise).resolves.not.toThrow();
    });
  });
  
  describe('Performance', () => {
    test('should handle long text efficiently', async () => {
      const longText = 'A'.repeat(1000);
      const start = Date.now();
      
      await typewriterEffect(element, longText, 1);
      
      const elapsed = Date.now() - start;
      expect(elapsed).toBeLessThan(2000); // 2 秒内完成
    });
  });
});

/**
 * 打字机效果实现（用于测试）
 */
async function typewriterEffect(element, text, speed = 100) {
  if (!element || !text) return;
  
  const chars = text.split('');
  element.textContent = '';
  
  for (const char of chars) {
    if (!element.isConnected) break; // 元素被移除
    element.textContent += char;
    await sleep(speed);
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}