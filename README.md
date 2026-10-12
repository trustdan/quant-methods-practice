# 📊 Quant Methods Practice

<div align="center">

[![Offline First](https://img.shields.io/badge/Architecture-100%25%20Offline%20First-2ea44f?style=for-the-badge&logo=shield)](docs/AGENT-CONTRACT.md)
[![Question Bank](https://img.shields.io/badge/Question%20Bank-137%20Approved%20Questions-blue?style=for-the-badge&logo=book)](docs/QUESTION-BANK.md)
[![Coverage](https://img.shields.io/badge/Objectives-22%20Topic%20Modules-6f42c1?style=for-the-badge)](docs/CONTENT-BACKLOG.md)
[![Go Backend](https://img.shields.io/badge/Engine-Go%20Authoritative%20Numerics-00ADD8?style=for-the-badge&logo=go)](internal/mathengine)
[![Frontend](https://img.shields.io/badge/UI-React%20%2B%20TypeScript%20%2B%20Vite-61DAFB?style=for-the-badge&logo=react)](web)
[![E2E Tests](https://img.shields.io/badge/Playwright%20E2E-16%2F16%20Passed-success?style=for-the-badge&logo=playwright)](tests/e2e)

<p align="center">
  <strong>A local-first, keyboard-driven deliberate-practice tutor for probability, statistics, and econometrics.</strong><br>
  Powered by an authoritative Go mathematical engine, interactive MathJax/LaTeX notation, deep pedagogical feedback loops, dynamic SVG statistical visuals lab, full-solution worksheets, offline Socratic AI tutoring, and an optional retro vector arcade.
</p>

[▶️ Quick Start](#-launching-the-application) •
[🚀 Interactive Tour](#-interactive-guided-product-tour) •
[🤖 Local LLMs](#-connecting-to-local-llms-via-lm-studio) •
[📚 Question Bank](#-topic--question-bank-coverage-matrix) •
[💻 Software Cheatsheets](#-reference-library--software-cheatsheets) •
[🔬 Visuals Lab](#-interactive-mathematical-visuals-lab) •
[🤝 Contributing & Issues](#-contributing--reporting-issues-for-everyone) •
[⌨️ Controls](#️-keyboard-shortcuts-cheat-sheet) •
[📖 Documentation](#-documentation-map)

</div>

---

## 🚀 Interactive Guided Product Tour

**Quant Methods Practice** includes a modern, interactive guided walkthrough featuring **spotlight highlighting** and **anchored callout speech bubbles** with directional pointer arrows:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      INTERACTIVE PRODUCT TOUR PIPELINE                      │
├──────────────────────────┬──────────────────────────┬───────────────────────┤
│ 1. 7-Stage Scaffolding   │ 2. Answer Workspace      │ 3. Socratic AI Tutor  │
│    • Target to Reflection│    • Numerical & Choices │    • 100% Offline LLM │
│    • Stage unlock gates  │    • First-error hints   │    • LM Studio/Ollama │
│    • Hotkeys: h / l      │    • Exact Go tolerances │    • Shortcut: n      │
├──────────────────────────┼──────────────────────────┼───────────────────────┤
│ 4. Concept Mastery Matrix│ 5. Reference & Software  │ 6. Worked Worksheets  │
│    • 3-day decay model   │    • Full LaTeX proofs   │    • Printable sheets │
│    • Transfer groups     │    • Excel & R recipes   │    • CSV 500-row case │
│    • Shortcut: s         │    • Shortcut: r         │    • Shift+J / Shift+F│
├──────────────────────────┼──────────────────────────┴───────────────────────┤
│ 7. Visuals Math Lab      │ 8. Timed Exam Assessments                        │
│    • Dynamic SVG curves  │    • Strict UTC wall-clock exam simulation       │
│    • Real-time sliders   │    • Aids withheld until final submission        │
│    • Shortcut: v         │    • SQLite retrospective replay & score audit   │
└──────────────────────────┴──────────────────────────────────────────────────┘
```

- **Anchored Callout Bubbles with Spotlight Highlighting**: As you step through the tour, an animated cyan highlight ring (`.tour-spotlight-highlight`) frames the exact UI component being explained while darkening the surrounding backdrop. Dynamic directional arrows point directly to target elements with real-time viewport collision detection.
- **Automated Page Transitions**: The tour automatically transitions between tabs and smoothly scrolls target sections into view.
- **Persistent Preferences in Local SQLite**: Selecting **"Don't show again on startup"** records your preference directly in your local SQLite database (`quant-methods.db`), persisting reliably across random port allocations and new application launches.
- **On-Demand Retrieval Anytime**:
  - Click the **`🚀 Tour`** button in the top navigation header at any time.
  - Press <kbd>F1</kbd> to open the Help dialog and click **"🚀 Launch Product Tour"**.
  - Open Settings (<kbd>t</kbd>) and click **"Re-enable on Startup"** to reset automatic startup display.

---

## 🎥 Application Walkthrough Tour

Experience the core capabilities of **Quant Methods Practice** in this 720p 60&nbsp;FPS high-definition automated walkthrough tour:

<div align="center">

<video src="https://github.com/user-attachments/assets/75d48397-11a8-4c0a-9f0f-ccaf5236716c" width="100%" controls="controls" muted="muted" playsinline style="border-radius: 8px; border: 1px solid #30363d; box-shadow: 0 8px 24px rgba(0,0,0,0.25);">
  <a href="https://github.com/user-attachments/assets/75d48397-11a8-4c0a-9f0f-ccaf5236716c">▶️ Click here to watch or download the recording (quant_methods_demo_tour.mp4)</a>
</video>

<p align="center">
  <sub><strong>Walkthrough Recording:</strong> <a href="https://github.com/user-attachments/assets/75d48397-11a8-4c0a-9f0f-ccaf5236716c"><code>quant_methods_demo_tour.mp4</code></a> (1280×720 @ 60 FPS, 6.7 MB)</sub>
</p>

</div>

### 🎬 Walkthrough Highlights & Chapters

| Chapter | Feature Demonstrated | Highlights in Recording |
|:---|:---|:---|
| **Act 1** | **Multi-Stage Practice Drill & Pedagogical Feedback** | • Scenario contemplation on four-toss binomial experiment ($n = 4, p = 0.5, k = 2$)<br>• **Intentional Misconception Trigger**: Selecting distractor confuses parameter with random variable count<br>• **Causal Hint & Retry Badge**: Real-time causal explanation triggers with active `"1 Retry Remaining"` state<br>• **Resolution & Derivation**: Correcting answer reveals green solved badge with full step-by-step LaTeX derivation<br>• **Scaffolding Progression**: Advances to Stage 2 distribution model, requests on-demand **Offline Hint**, and completes Stage 3 combinatorial formula |
| **Act 2** | **Full Question Bank (137 Questions) & Randomization** | • **Settings Modal (`t`)**: Inspects all 6 quantitative modules, question count selectors (5 to 137), and session intensity modes<br>• **Clock-Seeded Randomization (`🎲 New Drill`)**: Re-samples newly shuffled questions from the approved 137-template bank<br>• **Cross-Problem Navigation**: Answers fresh questions and smoothly cycles between problems using the Question Strip |
| **Act 3** | **Comprehensive Tour of Core Learning Modes** | • **Concept Mastery Matrix (`#tab-mastery`)**: 22-objective retention profile with real-time concept search<br>• **Reference Library (`#tab-reference`)**: Formula cards, axioms, and searchable LaTeX references with Excel/R functions<br>• **Saved Notes (`#tab-notes`)**: Persistent SQLite personal notes and derivation export<br>• **Question Candidates (`#tab-candidates`)**: In-app candidate generation and human review governance<br>• **Worksheets (`#tab-worksheets`)**: Full worked solution forms and raw 4-toss coin outcome CSV cases<br>• **Exams Mode (`#tab-exams`)**: Mock assessments with module targeting, strict aids withholding, and score reporting |
| **Act 3.7** | **Mathematical Visuals Lab (Smooth Scrolling to SVG Diagrams)** | • **Normal Distribution**: Upper-tail rejection region, CDF curves, and live Excel formula companion (`=NORM.DIST`)<br>• **Uniform Distribution**: Density rectangle $f(x) = 1/(b-a)$ and interval probabilities<br>• **Exponential Distribution**: Rate decay density $f(x) = \lambda e^{-\lambda x}$ and memoryless property<br>• **Discrete PMF/CDF**: Poisson arrival bar chart ($\lambda$) with exact mass probabilities<br>• **Central Limit Theorem Simulator**: Empirical sample mean convergence to theoretical normality ($\sigma/\sqrt{n}$)<br>• **Confidence Intervals**: 95% coverage simulation lines verifying frequentist long-run capture rates<br>• **Hypothesis Testing**: Two-tailed rejection bounds, critical values, and $\alpha$-level rejection regions<br>• **Bivariate Linear Regression**: Scatter plot, OLS best-fit regression line, and individual residual error bars |
| **Act 3.8** | **Quant Flight Space Arcade & Return** | • Retro vector arcade autopilot attract mode<br>• Manual flight takeover (`Space`), thruster maneuvers, and laser combat (`f` / `Space`) targeting statistical hazards<br>• Clean return to active practice session with state and focus fully preserved |

---

## ✨ Core Pillars & Architectural Highlights

<table>
<tr>
<td width="50%" valign="top">

### 🛡️ 100% Offline & Private
- **Zero Remote Dependencies**: Bundled local MathJax, SVG assets, and fonts. No CDNs, tracking cookies, or analytics.
- **Local Loopback Bridge**: Runs strictly on `127.0.0.1` via a compiled Go executable; stores history in local SQLite (`quant-methods.db`).
- **Protected AI Credentials**: Optional AI assistance routes use backend-controlled encryption; tokens are never stored in browser localStorage or plaintext logs.

</td>
<td width="50%" valign="top">

### 🧮 Authoritative Mathematical Engine
- **Deterministic Go Numerics**: The Go engine alone derives canonical answers, standard errors, quantiles, and grades. No LLM grading hallucinations.
- **Strict Numerical Rigor**: Bessel's sample correction ($n-1$), pooled vs. Welch degrees of freedom, Cauchy exact tails, and versioned tolerances.
- **Synchronized Spreadsheets**: Every statistical result pairs with canonical Microsoft Excel and Google Sheets formulas (`=T.INV.2T`, `=NORM.S.DIST`, `=SLOPE`).

</td>
</tr>
<tr>
<td width="50%" valign="top">

### 🎯 7-Stage Pedagogical Scaffolding
- **Structured Problem Decomposition**: Decomposes complex quantitative scenarios into sequential conceptual decisions:
  1. *Identify Target* &bull; 2. *Specify Model* &bull; 3. *Verify Assumptions* &bull; 4. *Identify Event* &bull; 5. *Construct Expression* &bull; 6. *Calculate* &bull; 7. *Contextual Interpretation*
- **Misconception Detection**: Categorizes specific cognitive pitfalls (e.g., parameter vs. statistic, mutually exclusive vs. independent, off-by-one tail bounds) with immediate causal feedback.

</td>
<td width="50%" valign="top">

### ⚡ Vim-Inspired Keyboard Fluency
- **Sub-Second Fluid Navigation**: Navigate questions, toggle options, reveal hints, switch tabs, and submit attempts without reaching for the mouse.
- **Instant Mode Switching**: Jump directly to Visuals Lab (`v`), Reference (`r`), Mastery (`s`), Notes (`V`), Exams (`Shift+E`), Arcade (`Shift+A`), or Worksheets (`Shift+J`).
- **Accessible Screen Readers & Contrast**: High-contrast dark theme with explicit ARIA live regions and full focus management.

</td>
</tr>
</table>

---

## 📚 Topic & Question Bank Coverage Matrix

The repository contains **137 thoroughly audited and approved questions** spanning **22 core quantitative objectives** across 6 business statistics and econometrics modules:

```
Question Bank Structure: 137 Approved Questions across 6 Modules
├── Module 1: Descriptive Statistics & Exploratory Data Analysis (16 Questions)
│   ├── Objective 1.1: Foundations, Data Types & Sampling Bias [5 questions]
│   ├── Objective 1.2: Set Probability, Contingency Tables & Independence [5 questions]
│   └── Objective 1.3: Central Tendency, Dispersion, CV & Bayes' Rule [6 questions]
├── Module 2: Discrete Probability Distributions (16 Questions)
│   ├── Objective 2.1: Discrete RVs, PMF/CDF Axioms & Moments [5 questions]
│   ├── Objective 2.2: Affine Transformations & Linear Combinations [5 questions]
│   └── Objective 2.3: Binomial & Poisson Distribution Modeling [6 questions]
├── Module 3: Continuous Probability Distributions (17 Questions)
│   ├── Objective 3.1: Uniform & Continuous Density Fundamentals [5 questions]
│   ├── Objective 3.2: Normal Distribution, Z-Scores & Inverse Norm [6 questions]
│   └── Objective 3.3: Exponential Distribution & Poisson Relations [6 questions]
├── Module 4: Sampling Distributions & Confidence Intervals (22 Questions)
│   ├── Objective 4.1: Sampling Distributions & Central Limit Theorem [5 questions]
│   ├── Objective 4.2: Standard Error & Inverse-Square Precision Laws [5 questions]
│   ├── Objective 4.3: One-Sample Means & Proportions Confidence Intervals [6 questions]
│   └── Objective 4.4: Sample Size Planning (Continuous & Conservative) [6 questions]
├── Module 5: Hypothesis Testing & Two-Sample Inference (32 Questions)
│   ├── Objective 5.1: Null/Alternative Formulations & Rejection Logic [5 questions]
│   ├── Objective 5.2: Type I & Type II Errors, Power & Decision Traps [5 questions]
│   ├── Objective 5.3: One-Sample Z and T Tests for Means & Proportions [6 questions]
│   ├── Objective 5.4: Two-Sample Independent & Paired T-Tests [6 questions]
│   └── Objective 5.5: Excel Output Interpretation & Inference Diagnostics [10 questions]
└── Module 6: Simple & Multiple Linear Regression (34 Questions)
    ├── Objective 6.1: Scatter Plots, Covariance & Pearson Correlation [5 questions]
    ├── Objective 6.2: Simple Linear Regression OLS Estimation & Residuals [5 questions]
    ├── Objective 6.3: ANOVA Decomposition, R², Multiple R & Standard Error [6 questions]
    ├── Objective 6.4: Multiple Regression, Partial Slopes & Ceteris Paribus [6 questions]
    └── Objective 6.5: Model Diagnostics (Multicollinearity, Heteroscedasticity) [12 questions]
```

Detailed objective-by-objective breakdown:

| Module | Core Concepts & Analytical Focus | Approved Bank |
|:---|:---|:---:|
| **Module 1: Descriptive & Exploratory Data Analysis** | Parameter vs. statistic, categorical/numerical types, sampling bias, mean vs. median with outliers, Bessel's sample variance, coefficient of variation ($CV$), contingency tables, mutually exclusive vs. independent events, Bayes' Theorem. | **16 Questions** |
| **Module 2: Discrete Probability Distributions** | Discrete PMF/CDF validity axioms, solving missing probabilities, $E[X]$, second moment $E[X^2]$, $\mathrm{Var}(X)$, affine scaling $E[aX+c]$ & $\mathrm{Var}(aX+c)$, sums of independent RVs, Binomial formula $\binom{n}{k}p^k(1-p)^{n-k}$, Poisson arrival rates and interval scaling. | **16 Questions** |
| **Module 3: Continuous Probability Distributions** | Uniform distribution $U(a, b)$, Normal distribution standardizations $Z = (X-\mu)/\sigma$, empirical 68-95-99.7 rule, cumulative CDF vs. tail probabilities, Inverse Normal percentiles, Exponential density $f(x) = \lambda e^{-\lambda x}$ and memoryless property. | **17 Questions** |
| **Module 4: Sampling Distributions & Confidence Intervals** | Central Limit Theorem convergence, standard error of mean ($SE = \sigma/\sqrt{n}$) and proportion, inverse-square precision law ($n \propto 1/E^2$), Student's $t$ degrees of freedom, $Z$ vs. $t$ confidence intervals, and conservative sample size planning ($p^* = 0.5$). | **22 Questions** |
| **Module 5: Hypothesis Testing & Two-Sample Inference** | Status quo null $H_0$ vs. directional alternative $H_1$, Type I ($\alpha$) vs. Type II ($\beta$) error trade-offs, power ($1-\beta$), frequentist $p$-value definition, one-sample $Z$ and $t$ tests, paired differences $t$-test, independent pooled $t$-test, and Excel output reading. | **32 Questions** |
| **Module 6: Simple & Multiple Linear Regression** | Scatter plots, Pearson correlation $r$, OLS slope $b_1 = SS_{xy}/SS_{xx}$ and intercept $b_0$, residual calculation $e = y - \hat{y}$, ANOVA partitioning ($SST = SSR + SSE$), coefficient of determination $R^2$, slope $t$-tests ($t^2 = F$), multiple regression partial slopes (*ceteris paribus*), dummy variables, multicollinearity, and heteroscedasticity. | **34 Questions** |
| **Total Bank Coverage** | **All 22 Published Objectives Fully Satisfied with Governed Review Records** | **137 Questions** |

---

## 💻 Reference Library & Software Cheatsheets

The **Reference Library** (`r`) provides exhaustive, offline reference for both theoretical mathematics and executable software implementations:

- **Mathematical Proofs & Assumptions**: Rigorous LaTeX derivations detailing explicit assumptions (support, independence, parameterization) and common student traps.
- **Excel & R Cheatsheet Catalog**: Side-by-side executable recipes:
  - **Excel**: Exact syntax and argument breakdowns (`=BINOM.DIST`, `=NORM.DIST`, `=POISSON.DIST`, `=CONFIDENCE.NORM`, `=T.TEST`, `=LINEST`).
  - **R**: Exact function idioms (`dbinom()`, `pbinom(..., lower.tail = TRUE)`, `qnorm()`, `t.test()`, `lm()`).
  - Syntax comparison callouts detailing 0-indexed vs. 1-indexed differences, cumulative flags, and discrete inequalities ($P(X \le k)$ vs $P(X < k)$).
- **Instant Search & Filter Pills**: Filter across Foundations, Discrete & Continuous Distributions, Moments, and Inference.

---

## 🔬 Interactive Mathematical Visuals Lab

The **Visuals Lab** (`v`) provides real-time, interactive parameter-driven SVG diagrams synchronized with analytical formulas and Excel syntax:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       MATHEMATICAL VISUALS LABORATORY                       │
├─────────────────────┬───────────────────────────┬───────────────────────────┤
│ 1. Normal Curve     │ 2. Uniform Rectangle      │ 3. Exponential Decay      │
│    • Bell curve SVG │    • Constant density     │    • Rate parameter λ     │
│    • Shaded tails   │    • Area under rectangle │    • Memoryless check     │
│    • Live =NORM.DIST│    • Live spreadsheet calc│    • Live =EXPON.DIST     │
├─────────────────────┼───────────────────────────┼───────────────────────────┤
│ 4. Discrete PMF/CDF │ 5. Central Limit Theorem  │ 6. Confidence Intervals   │
│    • Poisson arrivals│   • Population vs sample  │    • Repeated samples     │
│    • Probability bars│   • Standard error decay  │    • True mean capture    │
│    • Discrete CDF   │    • Large-sample norm    │    • Coverage rate bar    │
├─────────────────────┼───────────────────────────┼───────────────────────────┤
│ 7. Hypothesis Test  │ 8. Bivariate Regression   │ 9. Formula Reference      │
│    • Rejection tail │    • Scatter plot points  │    • LaTeX equations      │
│    • Critical z / t │    • OLS line fit         │    • Variance rules       │
│    • Alpha threshold│    • Residual deviations  │    • Excel equivalents    │
└─────────────────────┴───────────────────────────┴───────────────────────────┘
```

1. **Normal Distribution**: Dynamic mean $\mu$ and standard deviation $\sigma$ sliders; toggle between PDF curve and CDF sigmoid; switch between Lower Tail, Upper Tail, and Two-Tailed shaded regions; displays exact `=NORM.DIST` formula.
2. **Uniform Distribution**: Bounded continuous support $[a, b]$; calculates sub-interval probability $P(x_1 \le X \le x_2) = (x_2 - x_1)/(b - a)$ with shaded density area.
3. **Exponential Distribution**: Rate parameter $\lambda$ curve $f(x) = \lambda e^{-\lambda x}$; interactive cutoff $x_0$; verifies cumulative probability and the memoryless conditional property $P(X > s + t \mid X > s) = P(X > t)$.
4. **Discrete PMF / CDF Explorer**: Poisson and custom discrete probability mass distributions; dynamically displays bar heights and step CDF functions.
5. **Central Limit Theorem Sampling Simulator**: Compares skewed parent distributions against simulated sampling distributions of the mean $\bar{X}$ across sample sizes ($n = 5, 30, 100$).
6. **Confidence Intervals Coverage Visualizer**: Generates repeated 95% confidence intervals from a known population to visualize the frequentist capture proportion in action.
7. **Hypothesis Testing Rejection Explorer**: Live null distribution curve showing critical rejection boundaries, rejection region shading, and test statistic placement.
8. **Linear Regression & Residual Analyzer**: Bivariate coordinate scatter plot; interactive OLS line fit $\hat{y} = b_0 + b_1 x$; displays vertical residual segments ($e_i = y_i - \hat{y}_i$).

---

## 🚀 Launching the Application

### Single-Command Start (Windows, macOS, Linux)

Clone the repository and launch immediately:

```bash
# Clone repository
git clone https://github.com/trustdan/quant-methods-practice.git
cd quant-methods-practice

# Windows PowerShell:
.\scripts\start.ps1

# Windows Command Prompt:
.\scripts\start.bat

# Linux / macOS:
./scripts/start.sh
```

Or launch the pre-compiled native binary directly:
```bash
.\bin\quant-practice.exe
```

The application starts a local loopback server, bootstraps an authenticated session token, and automatically opens your default browser.

### CLI Inspection & Offline Study Commands

The compiled Go executable includes rich command-line inspection and study tools:

```bash
# View active approved question bank
.\bin\quant-practice.exe -list-bank

# Inspect concept retention and transfer status
.\bin\quant-practice.exe -mastery

# Access personal saved notes library
.\bin\quant-practice.exe -notes

# Check status of local AI providers and credential vault
.\bin\quant-practice.exe -providers

# Launch a CLI-based assessment session (default 5 questions)
.\bin\quant-practice.exe -exam 10

# Launch a timed assessment (e.g., 20-minute limit; 0 = untimed)
.\bin\quant-practice.exe -exam 10 -exam-time 20

# Resume an existing assessment session by ID
.\bin\quant-practice.exe -resume-exam <exam-id>

# View history of completed assessments
.\bin\quant-practice.exe -exam-history

# Print comprehensive score report and mathematical derivations for an exam
.\bin\quant-practice.exe -exam-report <exam-id>
```

---

## 🤖 Connecting to Local LLMs via LM Studio

Follow these steps to run a local language model on your machine and connect it to this application for 100% private, cost-free AI tutoring with zero external network calls.

### Step 1: Install LM Studio
1. Download and install LM Studio for your operating system from [lmstudio.ai](https://lmstudio.ai).
2. Launch the application.
3. **Verification:** The LM Studio home screen opens with the sidebar visible on the left.

### Step 2: Download a Model
1. Click the **Search / Discover icon** (magnifying glass 🔍) in the left-hand navigation sidebar.
2. Search for a supported instruction-tuned model:
   - `Qwen2.5-7B-Instruct-GGUF` (recommended for quantitative and mathematical reasoning)
   - `Meta-Llama-3.1-8B-Instruct-GGUF` (or 3.2-3B for lighter hardware)
3. Under the model variants, select a quantization level suited to your hardware (e.g., `Q4_K_M` for standard GPU/RAM setups) and click **Download**.
4. **Verification:** Once downloaded, the model will appear in your local library under the **My Models** tab (folder icon).

### Step 3: Load the Model and Start the Server
1. Click the **Developer tab** (the `<->` / server icon) in the left-hand navigation sidebar (labeled *Developer* or *Local Server* depending on version).
2. At the top of the Developer screen, click the **Select a model to load** dropdown and choose your downloaded model.
3. Wait for the model to load into memory. The top bar will display the active model name along with estimated RAM/VRAM consumption.
4. Locate the server controls at the top right and click the **Start Server** toggle/button.
5. **Verification:** The status indicator switches to green, displaying `Server running on http://localhost:1234`, and live server logs appear in the central console.

### Step 4: Configure App Connection Settings
Enter the following values into this application's AI / LLM configuration settings (open via Settings modal `t` -> **AI Providers** -> **Local Offline LLM**):

| Setting | Value | Notes |
|:---|:---|:---|
| **API Base URL** | `http://localhost:1234/v1` | LM Studio exposes an OpenAI-compatible endpoint. Click the **LM Studio (:1234)** preset button. |
| **API Key** | `lm-studio` | LM Studio does not require authentication, but client adapters require a non-empty string. |
| **Model Name / ID** | Exact identifier from LM Studio | Found in the Developer tab (or click **Probe / Refresh Local Models** to detect automatically). |

### Step 5: Advanced & Environment Settings
Depending on where this application runs relative to LM Studio, adjust the following toggles in LM Studio's right-hand Server Settings panel:
- **Browser-Based Apps (CORS):** If this application runs in a web browser (e.g., `http://localhost:3000`), toggle **Enable CORS** to **ON**. Leaving this off causes the browser to block requests due to cross-origin security policies.
- **Docker / WSL / Remote Clients:** If this application runs inside Docker, WSL2, or a different device on your network, toggle **Serve on Local Network** to **ON** (binds to `0.0.0.0` instead of `127.0.0.1`). Use your host machine’s local IP address or `http://host.docker.internal:1234/v1` as the Base URL.
- **Context Length & Hardware Offload:** In the model settings panel on the right, ensure **GPU Offload** is enabled to maximize inference speed, and verify the context window size matches your application's token requirements.

### Connection Verification Checklist
To confirm the server is operating before running the app, execute this check in your terminal:
```bash
curl http://localhost:1234/v1/models
```

**Expected Result:** A JSON response listing the active model:
```json
{ "object": "list", "data": [ { "id": "model-identifier", "object": "model", "owned_by": "organization" } ] }
```

Once the terminal returns this JSON object and LM Studio shows incoming request logs, launch the application and initiate a prompt.

---

## 🤝 Contributing & Reporting Issues (For Everyone!)

We welcome questions, suggestions, bug reports, and contributions from **learners, educators, researchers, and developers alike**. You do **not** need to be a software developer or git expert to help make this tool better!

### 💡 What is an "Issue"?
On GitHub, an **Issue** is simply a message board topic or public ticket. Think of it as a forum post or feedback form where you can:
- 🐛 **Report a Bug or Glitch**: Found a typo, an equation that didn't render, or a button that didn't respond? Let us know so we can fix it!
- ❓ **Ask a Math or Concept Question**: If an explanation was confusing or an answer seemed counterintuitive, let us know where the explanation can be made clearer.
- 💡 **Suggest Improvements or New Questions**: Wish there was a question on the Hypergeometric distribution? Want an extra Excel shortcut in the Cheatsheet? Post your idea!
- 📝 **Feedback on Usability**: Tell us how the app feels on your screen size or keyboard layout.

### 📝 How to Report an Issue (Step-by-Step for Non-GitHub Users)
If you've never used GitHub before, filing an issue is straightforward:
1. **Create or Sign In to a GitHub Account**: Go to [github.com/signup](https://github.com/signup) if you don't already have one (it takes less than a minute).
2. **Go to the Issues Page**: Click on the **[Issues](https://github.com/trustdan/quant-methods-practice/issues)** tab near the top of this repository (or visit `https://github.com/trustdan/quant-methods-practice/issues`).
3. **Click "New Issue"**: Click the green **New Issue** button on the right.
4. **Fill in a Title and Description**:
   - *Title*: A short summary (e.g., *"Typo in Module 2 Poisson derivation"* or *"Feature suggestion: Add TI-84 command for normal CDF"*).
   - *Description*: Tell us what happened, what problem you were working on, or what idea you have. Screenshots are welcome (you can simply paste images directly into the text box!).
5. **Click "Submit new issue"**: That's it! The maintainers will receive a notification and reply to help.

### 🛠️ Contributing Without Writing Code
You can contribute meaningfully without touching a single line of programming code:
- **Proofreading & Copy Editing**: Review problem scenario narratives for clarity, grammatical polish, and accessibility.
- **Pedagogical Feedback**: Suggest better causal hints for student misconceptions.
- **Recipe Verification**: Test the Excel, R, or Python commands in the [Reference Library](web/src/features/reference/ReferenceLibrary.tsx) to ensure compatibility across spreadsheet versions.

### 💻 Developer Contributions
For engineers, data scientists, and developers looking to submit code:
1. Fork the repository and clone your fork locally.
2. Read [docs/AGENT-CONTRACT.md](docs/AGENT-CONTRACT.md) to understand the local-first architectural invariants (e.g., pure Go mathematical engine, deterministic evaluation, absolute offline privacy).
3. Run the verification test suite before opening a pull request:
   ```bash
   # Run all frontend tests
   cd web && npm test -- --run

   # Run full end-to-end tests and Go verification
   powershell -ExecutionPolicy Bypass -File scripts\test.ps1  # Windows
   ./scripts/test.sh                                         # Linux / macOS
   ```
4. See [CONTRIBUTING.md](CONTRIBUTING.md) for full coding conventions, review standards, and PR templates.

---

## ⌨️ Keyboard Shortcuts Cheat Sheet

The interface is engineered for complete, mouse-free fluency:

| Category | Shortcut | Action |
|:---|:---:|:---|
| **Option Selection** | `1` – `4` / `j` / `k` | Highlight multiple-choice option card |
| **Submission** | `Enter` / `Space` | Submit highlighted option or advance to next stage |
| **Assistance** | `h` / `?` | Request on-demand **Offline Hint** |
| **AI Tutor** | `n` | Toggle **Socratic AI Tutor** panel |
| **Explanations** | `r` | View step-by-step mathematical rationale and proof |
| **Session Control** | `t` | Open **Settings Modal** (filters, question count, intensity) |
| **Randomization** | Click `🎲 New Drill` | Trigger clock-seeded PRNG re-sampling across approved bank |
| **Problem Navigation** | `[` / `]` | Move to previous / next problem in the active drill |
| **Stage Navigation** | `h` / `l` | Move to previous / next stage in multi-stage drill |
| **Visuals Lab** | `v` | Switch to **Mathematical Visuals Lab** tab |
| **Exams Assessment** | `Shift + E` | Switch to **Exams Mode** |
| **Full-Solution Worksheets** | `Shift + J` | Open **Full Worked-Solution Worksheets** |
| **Dataset CSV Cases** | `Shift + F` | View **4-Toss Raw Outcome CSV Cases** |
| **Space Arcade** | `Shift + A` | Launch **Quant Flight Space Arcade** |
| **Candidate Governance** | `p` | Open **Question Candidates Review** tab |
| **Help & Tour** | `F1` | Open **Keyboard Help** and launch **Product Tour** |

---

## 🏗️ System Architecture & Engineering

```
┌────────────────────────────────────────────────────────────────────────┐
│                        BROWSER APPLICATION (UI)                        │
│   • React 18 + Vite + TypeScript (SPA)                                 │
│   • Bundled MathJax 3.2.2 (Offline TeX/LaTeX rendering)               │
│   • SVG Accessible Charts (Normal, Uniform, Exponential, CLT, Reg)     │
│   • HTML5 Canvas 2D Space Arcade (Retro vector renderer)              │
│   • Sub-second Keyboard Navigation & Focus Management                  │
│   • Interactive Spotlight Product Tour with Anchored Speech Bubbles    │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ Local Loopback HTTP (Bootstrap Token)
┌───────────────────────────────────▼────────────────────────────────────┐
│                       LOCAL GO BACKEND PROCESS                         │
│                                                                        │
│   ┌────────────────────────┐  ┌───────────────────────────────────┐    │
│   │   httpapi & auth       │  │   mathengine (Analytical Core)    │    │
│   │   • Local endpoints    │  │   • Combinatorics & Distributions │    │
│   │   • Bootstrap token    │  │   • Sample variance & CV          │    │
│   │   • Session lifecycle  │  │   • Z & T confidence intervals    │    │
│   └───────────┬────────────┘  │   • Hypothesis tests & ANOVA      │    │
│               │               │   • OLS Regression estimators     │    │
│   ┌───────────▼────────────┐  └─────────────────┬─────────────────┘    │
│   │   bank & family        │                    │                      │
│   │   • 137 approved JSONs │◄───────────────────┘                      │
│   │   • Template validator │   Authoritative Canonical Answers         │
│   │   • Clock-seeded PRNG  │                                           │
│   └───────────┬────────────┘                                           │
│               │ Attempt Records & Spaced Retrieval                     │
│   ┌───────────▼────────────┐  ┌───────────────────────────────────┐    │
│   │   storage (SQLite)     │  │   providers (Protected Storage)   │    │
│   │   • Immutable snapshots│  │   • Optional AI tutor connections │    │
│   │   • Mastery matrix     │  │   • ChatGPT sign-in / API keys    │    │
│   │   • Notes & exports    │  │   • Protected local keychain      │    │
│   └────────────────────────┘  └───────────────────────────────────┘    │
└────────────────────────────────────────────────────────────────────────┘
```

- **Separation of Concerns**: Mathematical truths are owned exclusively by Go's pure `mathengine` package. Frontend code and AI prompts never grade answers or alter mastery scores.
- **Immutable Snapshots**: Every practice attempt records the exact problem parameters, random seed, option ordering, submitted answer, assistance used, and timestamp. Replaying past assessments always uses the snapshot, immune to future bank edits.
- **Zero Hallucination Grading**: Problem answers are computed deterministically using standard numerical libraries with strict epsilon tolerances ($10^{-4}$ or context-specific).

---

## 🧪 Development, Auditing & Verification

The repository enforces end-to-end automated verification across docs, Go backend, and web frontend:

```bash
# Run full repository verification suite (scaffold, lint, unit, e2e, go tests)
.\scripts\test.ps1        # Windows PowerShell
./scripts/test.sh         # Linux / macOS

# Build production assets and compile native Go binary
.\scripts\build.ps1       # Windows PowerShell
./scripts/build.sh        # Linux / macOS

# Validate scaffold integrity and markdown link resolution
python scripts/validate_scaffold.py

# Verify question bank coverage across all 22 objectives
python scripts/check_content_coverage.py

# Run Playwright browser end-to-end test suite
cd web
npm run test:e2e

# Run Vitest frontend unit tests (166 tests)
npm run test
```

---

## 📖 Documentation Map

All specifications, architectural contracts, and educational frameworks are maintained in the repository:

| Document | Purpose |
|:---|:---|
| [REQUIREMENTS.md](REQUIREMENTS.md) | Comprehensive functional and non-functional requirements inventory |
| [PLAN.md](PLAN.md) | Ordered implementation roadmap and evidence verification gates |
| [OVERVIEW.md](OVERVIEW.md) | Product design philosophy and deliberate-practice vision |
| [docs/AGENT-CONTRACT.md](docs/AGENT-CONTRACT.md) | Canonical invariant contract governing all assistant implementations |
| [docs/HANDOFF.md](docs/HANDOFF.md) | Historical implementation log, verified milestones, and run logs |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Package decomposition, module boundaries, and system runtime model |
| [docs/DIRECTORY-STRUCTURE.md](docs/DIRECTORY-STRUCTURE.md) | Repository directory conventions and artifact organization |
| [docs/FEATURE-PARITY.md](docs/FEATURE-PARITY.md) | Pedagogical mapping to quantitative and statistical concepts |
| [docs/COURSE-MAP.md](docs/COURSE-MAP.md) | Quantitative topic inventory and coverage reconciliation |
| [docs/PEDAGOGY.md](docs/PEDAGOGY.md) | 7-stage instructional scaffolding, feedback loops, and fading |
| [docs/QUESTION-BANK.md](docs/QUESTION-BANK.md) | Deterministic problem template generation and human review rules |
| [docs/CONTENT-BACKLOG.md](docs/CONTENT-BACKLOG.md) | 22-objective topic backlog tracking ($\ge 5$ questions per row) |
| [docs/REMAINING-CONTENT-REVIEW.md](docs/REMAINING-CONTENT-REVIEW.md) | Worked previews, calculations, and explanations for recent content |
| [docs/WORKSHEETS.md](docs/WORKSHEETS.md) | Step-by-step calculation worksheets and empirical 4-toss CSV cases |
| [docs/NUMERICS.md](docs/NUMERICS.md) | Floating-point numerical methods, Bessel's correction, and tolerances |
| [docs/NAVIGATION.md](docs/NAVIGATION.md) | Vim-like keyboard navigation, focus trap design, and accessibility |
| [docs/MASTERY.md](docs/MASTERY.md) | Spaced retrieval scoring, evidence weighting, and concept transfer |
| [docs/STORAGE.md](docs/STORAGE.md) | SQLite database schema, schema migrations, and markdown exports |
| [docs/API.md](docs/API.md) | Local HTTP REST API endpoints and state lifecycle contracts |
| [docs/AI-TUTOR.md](docs/AI-TUTOR.md) | Read-only AI tutoring protocol, bounds, and note generation |
| [docs/PROVIDERS.md](docs/PROVIDERS.md) | Secure provider authentication, key storage, and model routing |
| [docs/MATH-AND-VISUALS.md](docs/MATH-AND-VISUALS.md) | TeX typesetting guidelines and SVG interactive visualization design |
| [docs/ARCADE.md](docs/ARCADE.md) | Quant Flight 2D vector flight simulator mechanics and separation |
| [docs/SECURITY.md](docs/SECURITY.md) | Local loopback authentication, credential isolation, and input safety |
| [docs/TESTING.md](docs/TESTING.md) | Automated testing strategy across unit, integration, and E2E tiers |
| [docs/RELEASE.md](docs/RELEASE.md) | Cross-platform build automation and native packaging roadmap |
| [docs/DECISIONS.md](docs/DECISIONS.md) | Architectural decision records (ADRs) and design trade-offs |
| [docs/SOURCES.md](docs/SOURCES.md) | Academic textbooks, statistical packages, and reference citations |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributor guidelines, development workflow, and coding standards |

Draft reference schemas: [schemas/question-template.schema.json](schemas/question-template.schema.json) and [schemas/saved-explanation.schema.json](schemas/saved-explanation.schema.json).

---

## ⚖️ Provenance & Guidelines

All original educational content, algorithms, and software in this repository are developed for deliberate quantitative practice, self-study, and research. Practice problem templates and datasets are synthetically generated and reviewed locally. See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor guidelines, testing practices, and redistribution terms.
