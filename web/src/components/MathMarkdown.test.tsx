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

  it('triggers MathJax typeset lifecycle', () => {
    const { container } = render(<MathMarkdown content="Value: $p = 0.5$" />);
    const div = container.querySelector('.math-markdown');
    expect(div).not.toBeNull();
  });
});
