import React, { useEffect, useRef } from 'react';
import DOMPurify from 'dompurify';

interface MathMarkdownProps {
  content: string;
  className?: string;
}

function escapeHtml(text: string): string {
  return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

// Convert common markdown elements while keeping math delimiters intact.
// Display math ($$ ... $$ or \[ ... \]) may span several lines and is kept in
// one element so MathJax can match its delimiters.
function markdownToHtml(raw: string): string {
  if (!raw) return '';

  const lines = raw.split(/\r?\n/);
  const output: string[] = [];
  let list: 'ul' | 'ol' | null = null;

  const closeList = () => {
    if (list) {
      output.push(`</${list}>`);
      list = null;
    }
  };

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();

    // Fenced code blocks
    if (trimmed.startsWith('```')) {
      closeList();
      const code: string[] = [];
      while (i + 1 < lines.length && !lines[i + 1].trim().startsWith('```')) {
        code.push(lines[++i]);
      }
      i++; // skip the closing fence (or run past the end of an unterminated stream)
      output.push(`<pre><code>${escapeHtml(code.join('\n'))}</code></pre>`);
      continue;
    }

    // Display math, single- or multi-line
    const opener = trimmed.startsWith('$$') ? '$$' : trimmed.startsWith('\\[') ? '\\[' : null;
    if (opener) {
      const closer = opener === '$$' ? '$$' : '\\]';
      const block: string[] = [trimmed];
      let closed = trimmed.length >= opener.length + closer.length && trimmed.endsWith(closer);
      while (!closed && i + 1 < lines.length) {
        const next = lines[++i];
        block.push(next);
        closed = next.trim().endsWith(closer);
      }
      closeList();
      const tex = escapeHtml(block.join('\n'));
      // An unterminated block (e.g. still streaming) shows as raw text until it closes.
      output.push(closed ? `<div class="display-math">${tex}</div>` : `<p>${tex}</p>`);
      continue;
    }

    // Headings (# through ######, up to three leading spaces)
    const heading = /^ {0,3}(#{1,6})\s+(.*?)\s*#*\s*$/.exec(line);
    if (heading) {
      closeList();
      const level = Math.min(heading[1].length + 1, 4); // h1 is reserved for the page
      output.push(`<h${level}>${formatInline(heading[2])}</h${level}>`);
      continue;
    }

    // Horizontal rules
    if (/^ {0,3}([-*_])(\s*\1){2,}\s*$/.test(line)) {
      closeList();
      output.push('<hr>');
      continue;
    }

    // List items
    const item = /^\s*(?:([-*+])|\d+[.)])\s+(.*)$/.exec(line);
    if (item) {
      const kind = item[1] ? 'ul' : 'ol';
      if (list !== kind) {
        closeList();
        output.push(`<${kind}>`);
        list = kind;
      }
      output.push(`<li>${formatInline(item[2])}</li>`);
      continue;
    }

    // Empty lines end a list
    if (trimmed === '') {
      closeList();
      continue;
    }

    closeList();
    output.push(`<p>${formatInline(line)}</p>`);
  }

  closeList();
  return output.join('\n');
}

// Format inline markdown (bold, italic, code), while protecting $...$ math spans
function formatInline(text: string): string {
  // Split on math segments so inline markdown formatting does not touch formulas
  const parts = text.split(/(\$\$[\s\S]*?\$\$|\$[^$\n]+?\$|\\\([\s\S]*?\\\))/g);
  return parts
    .map((part, index) => {
      // Odd indices are math segments ($...$, $$...$$ or \(...\)). Escape HTML only;
      // MathJax reads the original TeX back from the text content.
      if (index % 2 === 1) {
        return escapeHtml(part);
      }
      // Even indices are markdown text
      let escaped = escapeHtml(part);

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
      'h1', 'h2', 'h3', 'h4', 'br', 'hr', 'span', 'div',
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
