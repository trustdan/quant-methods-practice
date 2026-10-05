export type NavigationCommand =
  | 'NAV_PREV_STAGE'
  | 'NAV_NEXT_STAGE'
  | 'NAV_UP'
  | 'NAV_DOWN'
  | 'SELECT_CHOICE_1'
  | 'SELECT_CHOICE_2'
  | 'SELECT_CHOICE_3'
  | 'SELECT_CHOICE_4'
  | 'SUBMIT'
  | 'CONTINUE'
  | 'HINT'
  | 'HELP'
  | 'ESCAPE'
  | 'QUIT';

export interface KeyContext {
  isEditableActive: boolean;
  activeMode: 'practice' | 'reading' | 'modal';
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

export function resolveKeyCommand(event: KeyboardEvent, context: KeyContext): NavigationCommand | null {
  // If user is actively typing in an input field, do not hijack normal printable keys
  if (context.isEditableActive) {
    if (event.key === 'Escape') return 'ESCAPE';
    if (event.key === 'Enter' && !event.shiftKey) return 'SUBMIT';
    return null;
  }

  // Intercept reserved browser combinations like Ctrl+C, Ctrl+L, Ctrl+W: do NOT handle them!
  if (event.ctrlKey || event.metaKey || event.altKey) {
    return null;
  }

  switch (event.key) {
    // Stage navigation
    case 'h':
    case 'ArrowLeft':
      return 'NAV_PREV_STAGE';
    case 'l':
    case 'ArrowRight':
      return 'NAV_NEXT_STAGE';

    // Item movement
    case 'k':
    case 'ArrowUp':
      return 'NAV_UP';
    case 'j':
    case 'ArrowDown':
      return 'NAV_DOWN';

    // Choice selection
    case '1':
      return 'SELECT_CHOICE_1';
    case '2':
      return 'SELECT_CHOICE_2';
    case '3':
      return 'SELECT_CHOICE_3';
    case '4':
      return 'SELECT_CHOICE_4';

    // Actions
    case 'Enter':
      return 'SUBMIT';
    case ' ':
      return 'CONTINUE';
    case '?':
    case 'e':
      return 'HINT';
    case 'F1':
      return 'HELP';
    case 'Escape':
      return 'ESCAPE';
    case 'q':
      return 'QUIT';

    default:
      return null;
  }
}
