import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const webDir = path.resolve(__dirname, '..');
const nodeModules = path.resolve(webDir, 'node_modules');

const env = {
  ...process.env,
  NODE_PATH: nodeModules + (process.env.NODE_PATH ? path.delimiter + process.env.NODE_PATH : ''),
};

const cliPath = path.resolve(nodeModules, '@playwright/test/cli.js');
const child = spawn(process.execPath, [cliPath, 'test', ...process.argv.slice(2)], {
  cwd: webDir,
  env,
  stdio: 'inherit',
});

child.on('exit', (code) => {
  process.exit(code ?? 0);
});
