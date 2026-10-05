import { describe, it, expect } from 'vitest';
import { resolveKeyCommand, isEditableElement } from './keymap';

describe('keymap navigation', () => {
  it('resolves stage navigation keys in practice mode', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'h' }), context)).toBe('NAV_PREV_STAGE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'l' }), context)).toBe('NAV_NEXT_STAGE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'ArrowLeft' }), context)).toBe('NAV_PREV_STAGE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'ArrowRight' }), context)).toBe('NAV_NEXT_STAGE');
  });

  it('resolves choice selection keys 1-4', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '1' }), context)).toBe('SELECT_CHOICE_1');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '2' }), context)).toBe('SELECT_CHOICE_2');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '3' }), context)).toBe('SELECT_CHOICE_3');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '4' }), context)).toBe('SELECT_CHOICE_4');
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

  it('does not hijack typing when focus is inside an editable field', () => {
    const context = { isEditableActive: true, activeMode: 'practice' as const };

    // Printable keys should return null so native typing occurs
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '1' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'h' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'l' }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: '?' }), context)).toBeNull();

    // Escape and Enter are permitted in fields
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Escape' }), context)).toBe('ESCAPE');
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'Enter' }), context)).toBe('SUBMIT');
  });

  it('ignores browser reserved combinations with Ctrl or Meta', () => {
    const context = { isEditableActive: false, activeMode: 'practice' as const };

    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'c', ctrlKey: true }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'l', ctrlKey: true }), context)).toBeNull();
    expect(resolveKeyCommand(new KeyboardEvent('keydown', { key: 'w', ctrlKey: true }), context)).toBeNull();
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
