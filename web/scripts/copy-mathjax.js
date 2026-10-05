import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const src = path.resolve(__dirname, '../node_modules/mathjax/es5');
const dest = path.resolve(__dirname, '../public/vendor/mathjax');

if (!fs.existsSync(src)) {
  console.error('MathJax source not found in node_modules/mathjax/es5. Run npm install first.');
  process.exit(1);
}

fs.mkdirSync(dest, { recursive: true });
fs.cpSync(src, dest, { recursive: true });
console.log('MathJax assets successfully copied to web/public/vendor/mathjax');
