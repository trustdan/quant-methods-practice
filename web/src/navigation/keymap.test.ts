import { describe, it, expect, beforeEach } from 'vitest';
import { resolveKeyCommand, isEditableElement, resetKeymapState } from './keymap';

describe('keymap navigation', () => {
  beforeEach(() => {
    resetKeymapState();
  });

  it('resolves stage navigation keys in practice mode', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'h' }), context)).toBe('NAV_PREV_STAGE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'l' }), context)).toBe('NAV_NEXT_STAGE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'ArrowLeft' }), context)).toBe('NAV_PREV_STAGE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'ArrowRight' }), context)).toBe('NAV_NEXT_STAGE');
  });

  it('resolves big problem navigation keys (Ctrl+Arrows and Shift+H/L)', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'ArrowLeft', ctrlKey: true }), context)).toBe('NAV_PREV_PROBLEM');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'ArrowRight', ctrlKey: true }), context)).toBe('NAV_NEXT_PROBLEM');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'H', shiftKey: true }), context)).toBe('NAV_PREV_PROBLEM');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'L', shiftKey: true }), context)).toBe('NAV_NEXT_PROBLEM');
  });

  it('resolves choice selection keys 1-4 and letter aliases a-d in choice mode', () => {
    const context = { isEditableActive: false, activeMode: 'practice_choice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '1' }), context)).toBe('SELECT_CHOICE_1');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '2' }), context)).toBe('SELECT_CHOICE_2');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '3' }), context)).toBe('SELECT_CHOICE_3');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '4' }), context)).toBe('SELECT_CHOICE_4');

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'a' }), context)).toBe('SELECT_CHOICE_1');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'b' }), context)).toBe('SELECT_CHOICE_2');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'c' }), context)).toBe('SELECT_CHOICE_3');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'd' }), context)).toBe('SELECT_CHOICE_4');
  });

  it('does not resolve choices 1-4 or a-d in numeric mode outside fields', () => {
    const context = { isEditableActive: false, activeMode: 'practice_numeric' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '1' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '2' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'a' }), context)).toBeNull();
  });

  it('resolves reading mode navigation and scrolling (j/k, u/d, gg, G)', async () => {
    const context = { isEditableActive: false, activeMode: 'reading' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'j' }), context)).toBe('NAV_DOWN');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'k' }), context)).toBe('NAV_UP');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'u' }), context)).toBe('SCROLL_HALF_UP');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'd' }), context)).toBe('SCROLL_HALF_DOWN');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'PageUp' }), context)).toBe('SCROLL_HALF_UP');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'PageDown' }), context)).toBe('SCROLL_HALF_DOWN');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'G', shiftKey: true }), context)).toBe('SCROLL_BOTTOM');

    // First 'g' primes sequence
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'g' }), context)).toBeNull();
    // Second 'g' within 500ms triggers SCROLL_TOP
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'g' }), context)).toBe('SCROLL_TOP');

    // In reading mode, h opens help (no stage h conflict)
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'h' }), context)).toBe('HELP');
  });

  it('resolves leave-intent confirmation modal keys (y, n, Esc, Enter)', () => {
    const context = { isEditableActive: false, activeMode: 'leave_intent' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'y' }), context)).toBe('LEAVE_INTENT_SAVE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Y' }), context)).toBe('LEAVE_INTENT_SAVE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Enter' }), context)).toBe('LEAVE_INTENT_SAVE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'n' }), context)).toBe('LEAVE_INTENT_DISCARD');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'N' }), context)).toBe('LEAVE_INTENT_DISCARD');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Escape' }), context)).toBe('LEAVE_INTENT_CANCEL');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'h' }), context)).toBeNull();
  });

  it('resolves action keys Enter, Space, ?, F1, Escape, q', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Enter' }), context)).toBe('SUBMIT');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: ' ' }), context)).toBe('CONTINUE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '?' }), context)).toBe('HINT');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'e' }), context)).toBe('HINT');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'F1' }), context)).toBe('HELP');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Escape' }), context)).toBe('ESCAPE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'q' }), context)).toBe('QUIT');
  });

  it('resolves view shortcut keys outside text entry (s, V, t, p, n, J, F, E, A, L)', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 's' }), context)).toBe('VIEW_MASTERY');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'V', shiftKey: true }), context)).toBe('VIEW_SAVED_NOTES');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 't' }), context)).toBe('VIEW_SETTINGS');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'p' }), context)).toBe('VIEW_CANDIDATE_REVIEW');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'n' }), context)).toBe('VIEW_AI_REQUEST');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'J', shiftKey: true }), context)).toBe('VIEW_FULL_SOLUTION');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'F', shiftKey: true }), context)).toBe('VIEW_DATASET_CASE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'E', shiftKey: true }), context)).toBe('VIEW_EXAM');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'A', shiftKey: true }), context)).toBe('VIEW_ARCADE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'L', shiftKey: false }), context)).toBe('VIEW_LEADERBOARD');
  });

  it('does not hijack typing when focus is inside an editable field', () => {
    const context = { isEditableActive: true, activeMode: 'practice' as const };

    // Printable keys should return null so native typing occurs
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '1' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'h' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'l' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '?' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 's' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'q' }), context)).toBeNull();

    // Escape and Enter are permitted in fields
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Escape' }), context)).toBe('ESCAPE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Enter' }), context)).toBe('SUBMIT');
  });

  it('ignores IME composition events', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };
    const event = new KeyboardEvent('keydown', { key: 'Enter' });
    Object.defineProperty(event, 'isComposing', { value: true });

    expect(resolveKeyCommand(event, context)).toBeNull();
  });

  it('strictly ignores browser reserved combinations with Ctrl or Meta', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'c', ctrlKey: true }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'l', ctrlKey: true }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'w', ctrlKey: true }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'r', ctrlKey: true }), context)).toBeNull();
  });

  it('correctly detects editable elements', () => {
    const input = document.createElement('input');
    const textarea = document.createElement('textarea');
    const div = document.createElement('div');
    const contentEditable = document.createElement('div');
    contentEditable.contentEditable = 'true';

    expect(isEditableElement(input)).toBe(true);
    expect(isEditableElement(textarea)).toBe(true);
    expect(isEditableElement(contentEditable)).toBe(true);
    expect(isEditableElement(div)).toBe(false);
    expect(isEditableElement(null)).toBe(false);
  });
});
