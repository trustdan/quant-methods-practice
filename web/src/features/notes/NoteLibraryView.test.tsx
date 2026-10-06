import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { NoteLibraryView } from './NoteLibraryView';
import { SavedExplanationDTO } from '../../types/tutor';

const MOCK_NOTES: SavedExplanationDTO[] = [
  {
    id: 'note_1',
    raw_markdown: 'Binomial calculation: $P(X=2) = 0.375$',
    origin_instance_id: 'inst_1',
    origin_stage_id: 'stage_1',
    topic: 'Binomial Distribution',
    provider_info: {
      title: 'Coin Toss Worked Derivation',
      concepts: ['binomial_pmf'],
      provider: 'offline',
      model: 'offline-curriculum',
      route: 'offline',
    },
    created_at: '2026-10-05T12:00:00Z',
    updated_at: '2026-10-05T12:00:00Z',
  },
  {
    id: 'note_2',
    raw_markdown: 'Poisson arrival rate $\\lambda = 3.0$',
    origin_instance_id: 'inst_2',
    origin_stage_id: 'stage_2',
    topic: 'Poisson Distribution',
    provider_info: {
      title: 'Call Center Arrivals',
      concepts: ['poisson_rate'],
      provider: 'offline',
      model: 'offline-curriculum',
      route: 'offline',
    },
    created_at: '2026-10-05T13:00:00Z',
    updated_at: '2026-10-05T13:00:00Z',
  },
];

describe('NoteLibraryView Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders note library with list of notes and selected note preview', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/notes')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(MOCK_NOTES),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    render(<NoteLibraryView onClose={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText(/Saved Explanations & Notes Library/i)).toBeInTheDocument();
      expect(screen.getAllByText('Coin Toss Worked Derivation').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Call Center Arrivals').length).toBeGreaterThan(0);
    });

    // Initial note details should be rendered on the right
    expect(screen.getByText(/Binomial calculation:/i)).toBeInTheDocument();
    expect(screen.getByText(/Advisory Note:/i)).toBeInTheDocument();
  });

  it('filters notes when searching', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('q=poisson')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([MOCK_NOTES[1]]),
        });
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve(MOCK_NOTES),
      });
    });

    render(<NoteLibraryView onClose={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getAllByText('Coin Toss Worked Derivation').length).toBeGreaterThan(0);
    });

    const searchInput = screen.getByPlaceholderText(/Search notes/i);
    fireEvent.change(searchInput, { target: { value: 'poisson' } });

    await waitFor(() => {
      expect(screen.queryByText('Coin Toss Worked Derivation')).toBeNull();
      expect(screen.getAllByText('Call Center Arrivals').length).toBeGreaterThan(0);
    });
  });

  it('navigates notes using keyboard shortcuts j and k', async () => {
    global.fetch = vi.fn().mockImplementation((_url: string) => {
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve(MOCK_NOTES),
      });
    });

    render(<NoteLibraryView onClose={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getAllByText('Coin Toss Worked Derivation').length).toBeGreaterThan(0);
    });

    // Initially note 0 is selected
    expect(screen.getByText(/Binomial calculation:/i)).toBeInTheDocument();

    // Press 'j' to move selection down
    fireEvent.keyDown(window, { key: 'j' });

    await waitFor(() => {
      expect(screen.getByText(/Poisson arrival rate/i)).toBeInTheDocument();
    });

    // Press 'k' to move selection back up
    fireEvent.keyDown(window, { key: 'k' });

    await waitFor(() => {
      expect(screen.getByText(/Binomial calculation:/i)).toBeInTheDocument();
    });
  });

  it('triggers note export when export button clicked', async () => {
    global.fetch = vi.fn().mockImplementation((_url: string) => {
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve(MOCK_NOTES),
      });
    });

    render(<NoteLibraryView onClose={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getAllByText('Coin Toss Worked Derivation').length).toBeGreaterThan(0);
    });

    const exportBtn = screen.getByRole('button', { name: /Export \(\.md\)/i });
    expect(exportBtn).toBeInTheDocument();

    // Mock document.createElement('a') click
    const clickSpy = vi.fn();
    const origCreateElement = document.createElement.bind(document);
    vi.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
      if (tagName === 'a') {
        const el = origCreateElement('a');
        el.click = clickSpy;
        return el;
      }
      return origCreateElement(tagName);
    });

    fireEvent.click(exportBtn);

    expect(clickSpy).toHaveBeenCalled();
  });
});
