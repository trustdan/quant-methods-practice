import React, { useEffect, useRef } from 'react';
import DOMPurify from 'dompurify';

interface MathMarkdownProps {
  content: string;
  className?: string;
}

// Convert common markdown elements while keeping math delimiters intact
function markdownToHtml(raw: string): string {
  if (!raw) return '';

  const lines = raw.split(/\r?\n/);
  const output: string[] = [];
  let inList = false;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // Headers
    if (line.startsWith('### ')) {
      if (inList) { output.push('</ul>'); inList = false; }
      output.push(`<h3>${formatInline(line.slice(4))}</h3>`);
      continue;
    }
    if (line.startsWith('## ')) {
      if (inList) { output.push('</ul>'); inList = false; }
      output.push(`<h2>${formatInline(line.slice(3))}</h2>`);
      continue;
    }
    if (line.startsWith('# ')) {
      if (inList) { output.push('</ul>'); inList = false; }
      output.push(`<h1>${formatInline(line.slice(2))}</h1>`);
      continue;
    }

    // List items
    if (/^[-*]\s+/.test(line)) {
      if (!inList) {
        output.push('<ul>');
        inList = true;
      }
      const itemText = line.replace(/^[-*]\s+/, '');
      output.push(`<li>${formatInline(itemText)}</li>`);
      continue;
    } else if (inList) {
      output.push('</ul>');
      inList = false;
    }

    // Empty lines
    if (line.trim() === '') {
      continue;
    }

    // Display math lines ($$...$$)
    if (line.trim().startsWith('$$') && line.trim().endsWith('$$') && line.trim().length > 4) {
      output.push(`<div class="display-math">${line.trim()}</div>`);
      continue;
    }

    // Standard paragraph
    output.push(`<p>${formatInline(line)}</p>`);
  }

  if (inList) {
    output.push('</ul>');
  }

  return output.join('\n');
}

// Format inline markdown (bold, italic, code), while protecting $...$ math spans
function formatInline(text: string): string {
  // Split on math segments so inline markdown formatting does not touch formulas
  const parts = text.split(/(\$\$[\s\S]*?\$\$|\$[^$\n]+?\$)/g);
  return parts
    .map((part, index) => {
      // Odd indices are math segments ($...$ or $$...$$)
      if (index % 2 === 1) {
        return part;
      }
      // Even indices are markdown text
      let escaped = part
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');

      // Inline code
      escaped = escaped.replace(/`([^`]+)`/g, '<code>$1</code>');
      // Bold
      escaped = escaped.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
      // Italic
      escaped = escaped.replace(/(?<!\*)\*([^*]+)\*(?!\*)/g, '<em>$1</em>');

      return escaped;
    })
    .join('');
}

export const MathMarkdown: React.FC<MathMarkdownProps> = ({ content, className = '' }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const typesetSeqRef = useRef<number>(0);

  const rawHtml = markdownToHtml(content);
  const sanitizedHtml = DOMPurify.sanitize(rawHtml, {
    ALLOWED_TAGS: [
      'p', 'strong', 'em', 'code', 'pre', 'ul', 'ol', 'li',
      'h1', 'h2', 'h3', 'h4', 'br', 'span', 'div',
      'mjx-container', 'svg', 'g', 'path', 'defs', 'use', 'rect', 'text'
    ],
    ALLOWED_ATTR: ['class', 'style', 'id', 'jax', 'display', 'role', 'aria-hidden', 'viewBox', 'width', 'height', 'fill', 'd', 'transform', 'href', 'x', 'y'],
    ALLOW_DATA_ATTR: true,
  });

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const currentSeq = ++typesetSeqRef.current;

    const triggerTypeset = () => {
      const mj = typeof window !== 'undefined' ? window.MathJax : undefined;
      if (!mj) return;

      const runTypeset = () => {
        if (currentSeq !== typesetSeqRef.current || !containerRef.current) return;
        try {
          if (mj.typesetClear) {
            mj.typesetClear([containerRef.current]);
          }
          if (mj.typesetPromise) {
            mj.typesetPromise([containerRef.current]).catch((err: unknown) => {
              console.warn('MathJax typesetting error:', err);
            });
          }
        } catch (err) {
          console.warn('MathJax execution error:', err);
        }
      };

      if (mj.startup?.promise) {
        mj.startup.promise.then(runTypeset);
      } else {
        runTypeset();
      }
    };

    // If MathJax is already ready, typeset immediately; otherwise listen or retry
    if (typeof window !== 'undefined' && window.MathJax) {
      triggerTypeset();
    } else {
      const timer = setTimeout(triggerTypeset, 100);
      return () => clearTimeout(timer);
    }

    return () => {
      // Clear typeset cache for this container on unmount or re-render
      const mj = typeof window !== 'undefined' ? window.MathJax : undefined;
      if (el && mj?.typesetClear) {
        try {
          mj.typesetClear([el]);
        } catch {
          // Ignore cleanup errors
        }
      }
    };
  }, [sanitizedHtml]);

  return (
    <div
      ref={containerRef}
      className={`math-markdown ${className}`}
      dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
    />
  );
};
