/**
 * AI Coding Guide - Interactive Features
 * Neo-Terminal Hacker Style
 */

(function() {
    'use strict';

    // ============================================
    // Configuration
    // ============================================
    const CONFIG = {
        typewriterTexts: [
            'AI 代码入门指南',
            '代码效率提升器',
            '智能编程助手',
            '开发加速工具'
        ],
        typewriterSpeed: 100,
        typewriterDeleteSpeed: 50,
        typewriterPauseTime: 2000
    };

    // ============================================
    // DOM Elements
    // ============================================
    const elements = {
        nav: document.querySelector('.nav'),
        navToggle: document.querySelector('.nav-toggle'),
        navLinks: document.querySelector('.nav-links'),
        typewriter: document.getElementById('typewriter'),
        backToTop: document.querySelector('.back-to-top'),
        scenarioItems: document.querySelectorAll('.scenario-item'),
        copyBtns: document.querySelectorAll('.copy-btn')
    };

    // ============================================
    // Typewriter Effect
    // ============================================
    class Typewriter {
        constructor(element, texts, config) {
            this.element = element;
            this.texts = texts;
            this.config = config;
            this.textIndex = 0;
            this.charIndex = 0;
            this.isDeleting = false;
            this.init();
        }

        init() {
            this.type();
        }

        type() {
            const currentText = this.texts[this.textIndex];

            if (this.isDeleting) {
                this.element.textContent = currentText.substring(0, this.charIndex - 1);
                this.charIndex--;
            } else {
                this.element.textContent = currentText.substring(0, this.charIndex + 1);
                this.charIndex++;
            }

            let typeSpeed = this.isDeleting ? this.config.typewriterDeleteSpeed : this.config.typewriterSpeed;

            if (!this.isDeleting && this.charIndex === currentText.length) {
                typeSpeed = this.config.typewriterPauseTime;
                this.isDeleting = true;
            } else if (this.isDeleting && this.charIndex === 0) {
                this.isDeleting = false;
                this.textIndex = (this.textIndex + 1) % this.texts.length;
                typeSpeed = 500;
            }

            setTimeout(() => this.type(), typeSpeed);
        }
    }

    // ============================================
    // Navigation
    // ============================================
    function initNavigation() {
        // Mobile menu toggle
        if (elements.navToggle && elements.navLinks) {
            elements.navToggle.addEventListener('click', () => {
                elements.navLinks.classList.toggle('active');
            });

            // Close menu on link click
            elements.navLinks.querySelectorAll('a').forEach(link => {
                link.addEventListener('click', () => {
                    elements.navLinks.classList.remove('active');
                });
            });
        }

        // Smooth scroll for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function(e) {
                e.preventDefault();
                const targetId = this.getAttribute('href');
                const targetElement = document.querySelector(targetId);

                if (targetElement) {
                    const navHeight = elements.nav ? elements.nav.offsetHeight : 0;
                    const targetPosition = targetElement.offsetTop - navHeight - 20;

                    window.scrollTo({
                        top: targetPosition,
                        behavior: 'smooth'
                    });
                }
            });
        });

        // Nav background on scroll
        let lastScroll = 0;
        window.addEventListener('scroll', () => {
            const currentScroll = window.pageYOffset;

            if (elements.nav) {
                if (currentScroll > 100) {
                    elements.nav.style.background = 'rgba(10, 10, 15, 0.98)';
                } else {
                    elements.nav.style.background = 'rgba(10, 10, 15, 0.9)';
                }
            }

            lastScroll = currentScroll;
        });
    }

    // ============================================
    // Back to Top Button
    // ============================================
    function initBackToTop() {
        if (!elements.backToTop) return;

        window.addEventListener('scroll', () => {
            if (window.pageYOffset > 500) {
                elements.backToTop.classList.add('visible');
            } else {
                elements.backToTop.classList.remove('visible');
            }
        });

        elements.backToTop.addEventListener('click', () => {
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        });
    }

    // ============================================
    // Scenario Accordion
    // ============================================
    function initAccordion() {
        elements.scenarioItems.forEach(item => {
            const header = item.querySelector('.scenario-header');

            header.addEventListener('click', () => {
                const isActive = item.classList.contains('active');

                // Close all other items
                elements.scenarioItems.forEach(otherItem => {
                    otherItem.classList.remove('active');
                });

                // Toggle current item
                if (!isActive) {
                    item.classList.add('active');
                }
            });
        });
    }

    // ============================================
    // Copy to Clipboard
    // ============================================
    function initCopyButtons() {
        elements.copyBtns.forEach(btn => {
            btn.addEventListener('click', async () => {
                const codeId = btn.dataset.code;
                const pre = btn.closest('.step-example, .scenario-prompt').querySelector('pre code');
                const text = pre ? pre.textContent : '';

                try {
                    await navigator.clipboard.writeText(text);
                    btn.textContent = '已复制';
                    btn.classList.add('copied');

                    setTimeout(() => {
                        btn.textContent = '复制';
                        btn.classList.remove('copied');
                    }, 2000);
                } catch (err) {
                    console.error('复制失败:', err);
                    btn.textContent = '复制失败';

                    setTimeout(() => {
                        btn.textContent = '复制';
                    }, 2000);
                }
            });
        });
    }

    // ============================================
    // Scroll Animations
    // ============================================
    function initScrollAnimations() {
        const observerOptions = {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach((entry, index) => {
                if (entry.isIntersecting) {
                    // Add staggered delay
                    setTimeout(() => {
                        entry.target.classList.add('visible');
                    }, index * 100);
                }
            });
        }, observerOptions);

        // Observe elements for animation
        const animatedElements = document.querySelectorAll('.tool-card, .step, .scenario-item, .resource-card');
        animatedElements.forEach(el => observer.observe(el));
    }

    // ============================================
    // Active Navigation Highlight
    // ============================================
    function initActiveNavHighlight() {
        const sections = document.querySelectorAll('section[id]');
        const navLinks = document.querySelectorAll('.nav-links a');

        const highlightNav = () => {
            const scrollPosition = window.scrollY + 100;

            sections.forEach(section => {
                const sectionTop = section.offsetTop;
                const sectionHeight = section.offsetHeight;
                const sectionId = section.getAttribute('id');

                if (scrollPosition >= sectionTop && scrollPosition < sectionTop + sectionHeight) {
                    navLinks.forEach(link => {
                        link.classList.remove('active');
                        if (link.getAttribute('href') === `#${sectionId}`) {
                            link.style.color = '#fbbf24';
                        } else {
                            link.style.color = '';
                        }
                    });
                }
            });
        };

        window.addEventListener('scroll', highlightNav);
        highlightNav(); // Initial call
    }

    // ============================================
    // Code Syntax Enhancement
    // ============================================
    function enhanceCodeBlocks() {
        // Add line numbers to code blocks
        document.querySelectorAll('pre code').forEach(block => {
            const lines = block.textContent.split('\n');
            if (lines.length > 3) {
                block.style.counterReset = 'line';
            }
        });

        // Re-highlight with Prism if available
        if (typeof Prism !== 'undefined') {
            Prism.highlightAll();
        }
    }

    // ============================================
    // Terminal Animation
    // ============================================
    function initTerminalAnimation() {
        const terminal = document.querySelector('.terminal');
        if (!terminal) return;

        // Add typing effect to terminal command
        const commandEl = terminal.querySelector('.command');
        if (commandEl) {
            const originalText = commandEl.textContent;
            commandEl.textContent = '';

            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        let i = 0;
                        const typeInterval = setInterval(() => {
                            if (i < originalText.length) {
                                commandEl.textContent += originalText.charAt(i);
                                i++;
                            } else {
                                clearInterval(typeInterval);
                            }
                        }, 50);
                        observer.disconnect();
                    }
                });
            }, { threshold: 0.5 });

            observer.observe(terminal);
        }
    }

    // ============================================
    // Keyboard Navigation
    // ============================================
    function initKeyboardNavigation() {
        document.addEventListener('keydown', (e) => {
            // ESC to close mobile menu
            if (e.key === 'Escape' && elements.navLinks) {
                elements.navLinks.classList.remove('active');
            }

            // Number keys to navigate sections (when not in input)
            if (!['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName)) {
                const sectionKeys = {
                    '1': '#concepts',
                    '2': '#tools',
                    '3': '#workflow',
                    '4': '#scenarios',
                    '5': '#resources'
                };

                if (sectionKeys[e.key]) {
                    const target = document.querySelector(sectionKeys[e.key]);
                    if (target) {
                        target.scrollIntoView({ behavior: 'smooth' });
                    }
                }
            }
        });
    }

    // ============================================
    // Performance: Debounce & Throttle
    // ============================================
    function debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    function throttle(func, limit) {
        let inThrottle;
        return function executedFunction(...args) {
            if (!inThrottle) {
                func(...args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // ============================================
    // Initialize Everything
    // ============================================
    function init() {
        // Core features
        initNavigation();
        initBackToTop();
        initAccordion();
        initCopyButtons();
        initScrollAnimations();
        initActiveNavHighlight();
        enhanceCodeBlocks();
        initTerminalAnimation();
        initKeyboardNavigation();

        // Typewriter effect
        if (elements.typewriter) {
            new Typewriter(
                elements.typewriter,
                CONFIG.typewriterTexts,
                CONFIG
            );
        }

        // Log initialization
        console.log('%c> AI_CODE Guide Initialized',
            'color: #fbbf24; font-family: monospace; font-size: 14px;');
        console.log('%c> Press 1-5 to navigate sections',
            'color: #10b981; font-family: monospace; font-size: 12px;');
    }

    // ============================================
    // Run on DOM Ready
    // ============================================
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // ============================================
    // Expose for debugging
    // ============================================
    window.AICodeGuide = {
        config: CONFIG,
        elements: elements
    };

})();
