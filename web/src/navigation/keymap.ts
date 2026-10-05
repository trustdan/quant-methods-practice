export type NavigationCommand =
  // Stage navigation
  | 'NAV_PREV_STAGE'
  | 'NAV_NEXT_STAGE'
  // Big problem navigation (matches accounting behavior)
  | 'NAV_PREV_PROBLEM'
  | 'NAV_NEXT_PROBLEM'
  // Choice movement & list navigation
  | 'NAV_UP'
  | 'NAV_DOWN'
  // Reading mode scrolling
  | 'SCROLL_UP'
  | 'SCROLL_DOWN'
  | 'SCROLL_HALF_UP'
  | 'SCROLL_HALF_DOWN'
  | 'SCROLL_TOP'
  | 'SCROLL_BOTTOM'
  // Choice selection (1-4, a-d)
  | 'SELECT_CHOICE_1'
  | 'SELECT_CHOICE_2'
  | 'SELECT_CHOICE_3'
  | 'SELECT_CHOICE_4'
  // Submission & feedback
  | 'SUBMIT'
  | 'CONTINUE'
  // Assistance & guidance
  | 'HINT'
  | 'HELP'
  // View shortcuts
  | 'VIEW_MASTERY'
  | 'VIEW_SAVED_NOTES'
  | 'VIEW_REFERENCE'
  | 'VIEW_SETTINGS'
  | 'VIEW_CANDIDATE_REVIEW'
  | 'VIEW_AI_REQUEST'
  | 'VIEW_FULL_SOLUTION'
  | 'VIEW_DATASET_CASE'
  | 'VIEW_EXAM'
  | 'VIEW_ARCADE'
  | 'VIEW_LEADERBOARD'
  | 'CYCLE_INTENSITY'
  | 'REDUCE_SESSION_SIZE'
  | 'INCREASE_SESSION_SIZE'
  // Escape & Quit
  | 'ESCAPE'
  | 'QUIT'
  // Leave-intent actions in confirmation modal
  | 'LEAVE_INTENT_SAVE'
  | 'LEAVE_INTENT_DISCARD'
  | 'LEAVE_INTENT_CANCEL';

export type KeyMode =
  | 'practice'
  | 'practice_choice'
  | 'practice_numeric'
  | 'feedback'
  | 'reading'
  | 'modal'
  | 'leave_intent';

export interface KeyContext {
  isEditableActive: boolean;
  activeMode: KeyMode;
}

export function isEditableElement(element: Element | null): boolean {
  if (!element) return false;
  const tagName = element.tagName.toUpperCase();
  if (tagName === 'INPUT' || tagName === 'TEXTAREA' || tagName === 'SELECT') {
    return true;
  }
  const htmlEl = element as HTMLElement;
  if (
    htmlEl.isContentEditable ||
    htmlEl.contentEditable === 'true' ||
    htmlEl.getAttribute?.('contenteditable') === 'true'
  ) {
    return true;
  }
  return false;
}

// Sequence tracker for vim-like 'gg' top-scroll command with 500ms timeout
let lastGTime = 0;

export function resetKeymapState(): void {
  lastGTime = 0;
}

export function resolveKeyCommand(event: KeyboardEvent, context: KeyContext): NavigationCommand | null {
  // 1. Ignore IME composition events
  if (event.isComposing) {
    return null;
  }

  // 2. If focus is inside an editable field, printable keys and editing combinations retain normal meaning
  if (context.isEditableActive) {
    if (event.key === 'Escape') return 'ESCAPE';
    if (event.key === 'Enter' && !event.shiftKey) return 'SUBMIT';
    return null;
  }

  // 3. Leave-intent modal mode: handle y/n/Esc/Enter
  if (context.activeMode === 'leave_intent') {
    if (event.key === 'y' || event.key === 'Y' || event.key === 'Enter') {
      return 'LEAVE_INTENT_SAVE';
    }
    if (event.key === 'n' || event.key === 'N') {
      return 'LEAVE_INTENT_DISCARD';
    }
    if (event.key === 'Escape') {
      return 'LEAVE_INTENT_CANCEL';
    }
    return null;
  }

  // 4. Modal mode: Escape dismisses, Tab cycles
  if (context.activeMode === 'modal') {
    if (event.key === 'Escape') return 'ESCAPE';
    return null;
  }

  // 5. Check modifier combinations
  if (event.ctrlKey || event.metaKey) {
    // Matches accounting behavior: Ctrl+Left / Ctrl+Right for previous / next big problem
    if (event.key === 'ArrowLeft') {
      return 'NAV_PREV_PROBLEM';
    }
    if (event.key === 'ArrowRight') {
      return 'NAV_NEXT_PROBLEM';
    }
    // Strictly pass through all other Ctrl combinations (Ctrl+C, Ctrl+L, Ctrl+W, Ctrl+R, Ctrl +/-, etc.)
    return null;
  }

  if (event.altKey) {
    return null;
  }

  // 6. Shift combinations for big problem navigation avoiding reserved shortcuts
  if (event.shiftKey) {
    if (event.key === 'H') {
      return 'NAV_PREV_PROBLEM';
    }
    if (event.key === 'L') {
      return 'NAV_NEXT_PROBLEM';
    }
    if (event.key === 'G' && context.activeMode === 'reading') {
      return 'SCROLL_BOTTOM';
    }
    if (event.key === 'V') {
      return 'VIEW_SAVED_NOTES';
    }
    if (event.key === 'J') {
      return 'VIEW_FULL_SOLUTION';
    }
    if (event.key === 'F') {
      return 'VIEW_DATASET_CASE';
    }
    if (event.key === 'E') {
      return 'VIEW_EXAM';
    }
    if (event.key === 'A') {
      return 'VIEW_ARCADE';
    }
  }

  // 7. Reading mode (Formula Gallery, Recap, etc.)
  if (context.activeMode === 'reading') {
    switch (event.key) {
      case 'j':
      case 'ArrowDown':
        return 'NAV_DOWN';
      case 'k':
      case 'ArrowUp':
        return 'NAV_UP';
      case 'u':
      case 'PageUp':
        return 'SCROLL_HALF_UP';
      case 'd':
      case 'PageDown':
        return 'SCROLL_HALF_DOWN';
      case 'g': {
        const now = Date.now();
        if (now - lastGTime < 500) {
          lastGTime = 0;
          return 'SCROLL_TOP';
        }
        lastGTime = now;
        return null;
      }
      case 'G':
        return 'SCROLL_BOTTOM';
      // In reading mode, h/F1 opens help (no stage h conflict)
      case 'h':
      case 'F1':
        return 'HELP';
      case 'Escape':
        return 'ESCAPE';
      case 'q':
        return 'QUIT';
      default:
        break;
    }
  }

  // 8. Practice modes
  const isChoiceMode = context.activeMode === 'practice' || context.activeMode === 'practice_choice';

  switch (event.key) {
    // Stage navigation
    case 'h':
    case 'ArrowLeft':
      return 'NAV_PREV_STAGE';
    case 'l':
    case 'ArrowRight':
      return 'NAV_NEXT_STAGE';

    // Choice / item movement
    case 'k':
    case 'ArrowUp':
      return 'NAV_UP';
    case 'j':
    case 'ArrowDown':
      return 'NAV_DOWN';

    // Choice selections (only in choice-capable modes)
    case '1':
      return isChoiceMode ? 'SELECT_CHOICE_1' : null;
    case '2':
      return isChoiceMode ? 'SELECT_CHOICE_2' : null;
    case '3':
      return isChoiceMode ? 'SELECT_CHOICE_3' : null;
    case '4':
      return isChoiceMode ? 'SELECT_CHOICE_4' : null;

    // Optional choice letter aliases (only in choice mode)
    case 'a':
      return isChoiceMode ? 'SELECT_CHOICE_1' : null;
    case 'b':
      return isChoiceMode ? 'SELECT_CHOICE_2' : null;
    case 'c':
      return isChoiceMode ? 'SELECT_CHOICE_3' : null;
    case 'd':
      return isChoiceMode ? 'SELECT_CHOICE_4' : null;

    // Actions
    case 'Enter':
      return 'SUBMIT';
    case ' ':
      return 'CONTINUE';
    case '?':
    case 'e':
      return 'HINT';
    case 'F1':
      // Resolve h conflict: F1 and button open Help in practice
      return 'HELP';
    case 'Escape':
      return 'ESCAPE';
    case 'q':
      return 'QUIT';

    // Views and application commands outside text fields
    case 's':
      return 'VIEW_MASTERY';
    case 'r':
      return 'VIEW_REFERENCE';
    case 't':
      return 'VIEW_SETTINGS';
    case 'p':
      return 'VIEW_CANDIDATE_REVIEW';
    case 'n':
      return 'VIEW_AI_REQUEST';
    case 'i':
      return 'CYCLE_INTENSITY';
    case '[':
      return 'REDUCE_SESSION_SIZE';
    case ']':
      return 'INCREASE_SESSION_SIZE';
    case 'L':
      return 'VIEW_LEADERBOARD';

    default:
      return null;
  }
}
