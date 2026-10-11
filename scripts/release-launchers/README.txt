================================================================================
  QUANT METHODS PRACTICE — STANDALONE OFFLINE RELEASE
================================================================================

Quant Methods Practice is a 100% self-contained, offline interactive quantitative
methods and probability tutor. It includes all mathematical engines, curriculum
problems, explanations, diagrams, fonts, and the Quant Flight arcade game directly
embedded in the executable.

ZERO EXTERNAL RUNTIME DEPENDENCIES:
  • No Go compiler required.
  • No Node.js / npm required.
  • No Python required.
  • No Internet connection required.
  • All LaTeX/MathJax fonts, styling, and graphics are bundled locally.

QUICK START:
  • Windows:
      Double-click "start.bat" or run "quant-practice.exe" in PowerShell/CMD.
  • macOS & Linux:
      Open a terminal in this directory and run:
          ./start.sh
      or:
          ./quant-practice

The application will bind to a local loopback port (127.0.0.1) and automatically
launch your default web browser.

OFFLINE DIAGNOSTICS & DOCTOR:
  To verify your system environment and embedded assets, run:
      ./quant-practice -diagnostics

USEFUL CLI COMMANDS:
  ./quant-practice -help             : View all available command-line flags
  ./quant-practice -list-bank        : View all 21 approved curriculum questions
  ./quant-practice -diagnostics      : Run environment and asset health checks
  ./quant-practice -exam             : Start an assessment mode session in CLI
  ./quant-practice -high-scores      : View the Quant Flight arcade leaderboard
  ./quant-practice -backup <path>    : Create a consistent backup of your progress
  ./quant-practice -export-data <f>  : Export all learner data to a JSON file
  ./quant-practice -port 8080        : Bind to a specific port
  ./quant-practice -no-browser       : Start server without auto-opening browser

DATA STORAGE & PRIVACY:
  All learner progress, notes, attempts, and high scores are stored locally on
  your machine in an encrypted/protected directory. No telemetry or network requests
  are made unless you explicitly configure an AI provider in Settings.

LICENSES:
  See LICENSE and NOTICE for open-source license information.
================================================================================
