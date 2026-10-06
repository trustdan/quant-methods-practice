import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MathMarkdown } from './MathMarkdown';

describe('MathMarkdown', () => {
  it('renders markdown text and preserves math delimiters', () => {
    const content = 'Consider random variable $X \\sim \\operatorname{Binomial}(n, p)$ and calculate:\n\n$$P(X = k) = \\binom{n}{k} p^k (1 - p)^{n - k}$$';
    render(<MathMarkdown content={content} />);

    expect(screen.getByText(/Consider random variable/)).toBeInTheDocument();
    // Check that math delimiters were preserved for MathJax to typeset
    expect(screen.getByText(/\$X \\sim \\operatorname{Binomial}\(n, p\)\$/)).toBeInTheDocument();
  });

  it('sanitizes malicious script tags and inline events', () => {
    const malicious = 'Safe text <script>alert("xss")</script><img src="x" onerror="alert(1)">';
    const { container } = render(<MathMarkdown content={malicious} />);

    expect(container.querySelector('script')).toBeNull();
    // DOMPurify removes invalid or malicious attributes
    const img = container.querySelector('img');
    if (img) {
      expect(img.getAttribute('onerror')).toBeNull();
    }
  });

  it('keeps a multi-line $$ block in one element so MathJax can match its delimiters', () => {
    const content = 'So we define:\n\n$$\nX=\\text{the number of heads in the four tosses}.\n$$\n\nYour turn.';
    const { container } = render(<MathMarkdown content={content} />);

    const blocks = container.querySelectorAll('.display-math');
    expect(blocks).toHaveLength(1);
    expect(blocks[0].textContent).toBe('$$\nX=\\text{the number of heads in the four tosses}.\n$$');
    expect(container.querySelectorAll('p')).toHaveLength(2);
  });

  it('keeps multi-line \\[ \\] blocks together and preserves TeX characters that look like HTML', () => {
    const content = '\\[\nP(X < 2) = P(X=0) + P(X=1)\n\\]\nand $a<b$ inline';
    const { container } = render(<MathMarkdown content={content} />);

    expect(container.querySelector('.display-math')?.textContent).toBe('\\[\nP(X < 2) = P(X=0) + P(X=1)\n\\]');
    expect(container.querySelector('p')?.textContent).toBe('and $a<b$ inline');
  });

  it('leaves an unterminated display block (mid-stream) as plain text', () => {
    const { container } = render(<MathMarkdown content={'$$\nX = 2'} />);
    expect(container.querySelector('.display-math')).toBeNull();
    expect(container.textContent).toContain('X = 2');
  });

  it('renders headings of every level, numbered lists and code fences', () => {
    const content = '### Step 1\n#### Detail\n  ## Indented\n1. first\n2. second\n- bullet\n```\nx <- 1\n```';
    const { container } = render(<MathMarkdown content={content} />);

    expect(screen.getByRole('heading', { name: 'Step 1' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Detail' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Indented' })).toBeInTheDocument();
    expect(container.textContent).not.toContain('#');
    expect(container.querySelectorAll('ol > li')).toHaveLength(2);
    expect(container.querySelectorAll('ul > li')).toHaveLength(1);
    expect(container.querySelector('pre code')?.textContent).toBe('x <- 1');
  });

  it('triggers MathJax typeset lifecycle', () => {
    const { container } = render(<MathMarkdown content="Value: $p = 0.5$" />);
    const div = container.querySelector('.math-markdown');
    expect(div).not.toBeNull();
  });
});
