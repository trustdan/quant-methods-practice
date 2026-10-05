import '@testing-library/jest-dom';

// Provide minimal MathJax mock for unit tests in jsdom
if (typeof window !== 'undefined') {
  window.MathJax = {
    typesetPromise: async (elements?: (HTMLElement | Element)[]) => {
      // Mock typeset by marking elements with data-mathjax-typeset
      if (elements) {
        elements.forEach((el) => {
          if ('dataset' in el) {
            el.dataset.mathjaxTypeset = 'true';
          }
        });
      }
    },
    typesetClear: (_elements?: (HTMLElement | Element)[]) => {},
    startup: {
      promise: Promise.resolve(),
    },
  };
}
