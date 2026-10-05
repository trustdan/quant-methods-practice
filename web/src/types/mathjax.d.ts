declare global {
  interface Window {
    MathJax?: {
      typesetPromise?: (elements?: (HTMLElement | Element)[]) => Promise<void>;
      typesetClear?: (elements?: (HTMLElement | Element)[]) => void;
      startup?: {
        promise?: Promise<void>;
      };
      [key: string]: unknown;
    };
  }
}

export {};
