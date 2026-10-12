export interface TourHighlight {
  icon: string;
  title: string;
  description: string;
  badge?: string;
}

export type AppTab =
  | 'drill'
  | 'reference'
  | 'visuals'
  | 'mastery'
  | 'notes'
  | 'candidates'
  | 'worksheets'
  | 'exams'
  | 'arcade';

export type TourPlacement = 'top' | 'bottom' | 'left' | 'right' | 'auto';

export interface TourStep {
  id: string;
  badge: string;
  badgeColor?: string;
  title: string;
  subtitle: string;
  icon: string;
  targetSelector: string;
  tabTarget?: AppTab | 'tutor';
  placement?: TourPlacement;
  jumpLabel?: string;
  highlights: TourHighlight[];
  shortcutHint?: string;
}

export const TOUR_STEPS: TourStep[] = [
  {
    id: 'drills_pipeline',
    badge: '🎯 Practice Pipeline',
    badgeColor: 'var(--accent-cyan)',
    title: 'Multi-Stage Practice Drills',
    subtitle: 'Systematically decompose complex quantitative problems into manageable mathematical phases.',
    icon: '🎯',
    targetSelector: '#stage-pipeline-tracker',
    tabTarget: 'drill',
    placement: 'bottom',
    jumpLabel: 'Go to Practice Drill',
    shortcutHint: 'Hotkeys: h / l to navigate stages, 1-4 for choices',
    highlights: [
      {
        icon: '🧩',
        title: '7-Stage Scaffolding',
        description: 'Work through Target, Model, Formula, Values, Calculation, Interpretation, and Reflection stages.',
        badge: 'Structured Learning',
      },
      {
        icon: '🔒',
        title: 'Cognitive Gate Enforcement',
        description: 'Stages unlock sequentially to ensure foundational assumptions are validated before calculating.',
        badge: 'Deterministic',
      },
      {
        icon: '🎲',
        title: 'Clock PRNG Variations',
        description: 'Generate fresh random parameterizations on demand with the "New Drill" action.',
        badge: 'Infinite Drills',
      },
    ],
  },
  {
    id: 'drill_workspace',
    badge: '✏️ Active Workspace',
    badgeColor: 'var(--accent-emerald)',
    title: 'Calculations & Non-Punitive Hints',
    subtitle: 'Input values, select choices, and receive immediate causal feedback before answers are revealed.',
    icon: '💡',
    targetSelector: '#stage-action-row',
    tabTarget: 'drill',
    placement: 'top',
    jumpLabel: 'Focus Answer Workspace',
    shortcutHint: 'Hotkeys: 1-4 for options, Enter to submit, ? for hint',
    highlights: [
      {
        icon: '💡',
        title: 'First-Error Causal Hints',
        description: 'Make a mistake? Receive diagnostic cues and formulas for a clean retry without penalty.',
        badge: 'Pedagogical',
      },
      {
        icon: '⚡',
        title: 'Exact Numerical Tolerances',
        description: 'Graded with versioned relative and absolute floating point tolerance policies via the Go engine.',
        badge: 'Rigorous Math',
      },
      {
        icon: '⌨️',
        title: 'Keyboard-First Speed',
        description: 'Submit answers and advance stages without ever lifting your hands from the keyboard.',
        badge: 'Enter to Submit',
      },
    ],
  },
  {
    id: 'tutor',
    badge: '💡 Private Socratic AI',
    badgeColor: '#ec4899',
    title: 'Socratic AI Tutor & Local LLMs',
    subtitle: 'Conversational problem-solving guidance that prompts your thinking without spoiling the answer.',
    icon: '🧠',
    targetSelector: '#tab-tutor',
    tabTarget: 'drill',
    placement: 'bottom',
    jumpLabel: 'Open AI Tutor Panel',
    shortcutHint: 'Shortcut: Press "n" anytime to toggle AI Tutor',
    highlights: [
      {
        icon: '🔒',
        title: '100% Offline Local Loopback',
        description: 'Connect to LM Studio (:1234) or Ollama (:11434) with zero internet traffic and zero API costs.',
        badge: 'Absolute Privacy',
      },
      {
        icon: '🎯',
        title: 'Strict Socratic Guardrails',
        description: 'The tutor provides diagnostic analogies and formula checks while strictly withholding final numbers.',
        badge: 'No Spoilers',
      },
      {
        icon: '📝',
        title: 'Dual Note Persistence',
        description: 'Save AI explanations or full conversation transcripts directly into your personal Notes library.',
        badge: 'Markdown Notes',
      },
    ],
  },
  {
    id: 'mastery',
    badge: '📈 Learning Analytics',
    badgeColor: 'var(--accent-indigo)',
    title: 'Evidence-Based Concept Mastery',
    subtitle: 'Distinguish true conceptual understanding from lucky guesses and assisted attempts.',
    icon: '📈',
    targetSelector: '#mastery-stats',
    tabTarget: 'mastery',
    placement: 'bottom',
    jumpLabel: 'View Mastery Dashboard',
    shortcutHint: 'Shortcut: Press "s" to view Mastery',
    highlights: [
      {
        icon: '📊',
        title: 'Independent vs. Assisted Evidence',
        description: 'Rigorous separation between first-try independent answers, hinted retries, and revealed solutions.',
        badge: 'Retention Policy',
      },
      {
        icon: '🔄',
        title: 'Transfer Group Validation',
        description: 'Verify your mastery holds across diverse narrative contexts, variable notations, and parameter spaces.',
        badge: 'Near & Far Transfer',
      },
      {
        icon: '💾',
        title: 'SQLite Learning Audit',
        description: 'Every attempt timestamp, parameter snapshot, and hint usage is safely preserved locally in SQLite.',
        badge: 'Local DB',
      },
    ],
  },
  {
    id: 'reference',
    badge: '💻 Software Recipes & Proofs',
    badgeColor: 'var(--accent-amber)',
    title: 'Reference Library & Excel / R Cheatsheets',
    subtitle: 'Instant access to rigorous mathematical derivations and executable software syntax.',
    icon: '📚',
    targetSelector: '#tab-ref-software',
    tabTarget: 'reference',
    placement: 'bottom',
    jumpLabel: 'Open Reference Library',
    shortcutHint: 'Shortcut: Press "r" to view Reference',
    highlights: [
      {
        icon: '📐',
        title: 'Mathematical Proofs & Axioms',
        description: 'Complete LaTeX derivations detailing explicit assumptions, support conditions, and common traps.',
        badge: 'Proofs & Traps',
      },
      {
        icon: '💻',
        title: 'Excel & R Cheatsheet Catalog',
        description: 'Exact one-line execution calls for Excel (BINOM.DIST, NORM.DIST) and R (dbinom, pnorm, lm).',
        badge: 'Software Recipes',
      },
      {
        icon: '🔍',
        title: 'Interactive Search & Filters',
        description: 'Rapidly search across Foundations, Discrete & Continuous Distributions, Moments, and Inference.',
        badge: 'Instant Lookup',
      },
    ],
  },
  {
    id: 'worksheets',
    badge: '📄 Worksheets & Cases',
    badgeColor: 'var(--accent-cyan)',
    title: 'Full Solution Worksheets & Datasets',
    subtitle: 'Generate comprehensive printable study worksheets and dataset case investigations.',
    icon: '📝',
    targetSelector: '#worksheet-workspace-card',
    tabTarget: 'worksheets',
    placement: 'bottom',
    jumpLabel: 'Explore Worksheets',
    shortcutHint: 'Shortcuts: Shift+J (Solutions) or Shift+F (Datasets)',
    highlights: [
      {
        icon: '📄',
        title: 'Full-Solution Problem Sets',
        description: 'Work through complete method, calculation, and interpretation fields with clean print styling.',
        badge: 'Print Ready',
      },
      {
        icon: '📊',
        title: 'CSV Dataset Investigation',
        description: 'Load and inspect empirical coin-toss runs and experiment tables up to 500 rows.',
        badge: 'Data Analysis',
      },
      {
        icon: '🛡️',
        title: 'Supported Practice Separation',
        description: 'Worksheet exercises provide full guided support and remain distinct from independent mastery records.',
        badge: 'Support Mode',
      },
    ],
  },
  {
    id: 'visuals',
    badge: '📊 Visuals Lab',
    badgeColor: '#a855f7',
    title: 'Interactive Distribution Visuals',
    subtitle: 'Visual math lab for Normal, Binomial, Poisson, and Exponential curves with live probability figures.',
    icon: '📊',
    targetSelector: '#visuals-explorer-card',
    tabTarget: 'visuals',
    placement: 'bottom',
    jumpLabel: 'Open Visuals Lab',
    shortcutHint: 'Shortcut: Press "v" to open Visuals Lab',
    highlights: [
      {
        icon: '📈',
        title: 'Real-Time SVG Distributions',
        description: 'Drag parameter sliders and observe probability densities and cumulative distributions adjust dynamically.',
        badge: '60fps Graphics',
      },
      {
        icon: '🎲',
        title: 'Central Limit Theorem Simulator',
        description: 'Simulate empirical sample means from skewed distributions converging toward the Gaussian curve.',
        badge: 'CLT Lab',
      },
      {
        icon: '📋',
        title: 'Spreadsheet Formula Reconciliation',
        description: 'Compare curve areas directly with corresponding Excel and R cumulative distribution calls.',
        badge: 'Software Bridge',
      },
    ],
  },
  {
    id: 'exams',
    badge: '⏱️ Strict Assessment',
    badgeColor: '#f43f5e',
    title: 'Timed Exams & Assessment Replays',
    subtitle: 'Test your mastery under realistic exam conditions with strict evaluation and retrospective replay.',
    icon: '⏱️',
    targetSelector: '#exam-config-card',
    tabTarget: 'exams',
    placement: 'bottom',
    jumpLabel: 'Open Exams View',
    shortcutHint: 'Shortcut: Shift+E to open Exams',
    highlights: [
      {
        icon: '⏱️',
        title: 'UTC Wall-Clock Timers',
        description: 'Strict countdown timer tied to UTC clock time, continuing even if the browser is closed.',
        badge: 'Exam Conditions',
      },
      {
        icon: '🙈',
        title: 'Withheld Hints & Scores',
        description: 'Answers and feedback are strictly locked until final submission to simulate real exam pressure.',
        badge: 'Zero Spoilers',
      },
      {
        icon: '📑',
        title: 'Full Retrospective Audit',
        description: 'Review historical test attempts with complete stage-by-stage solutions, scores, and timestamps.',
        badge: 'Exam Replays',
      },
    ],
  },
];
