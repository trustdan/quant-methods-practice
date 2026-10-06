import React, { useState, useEffect, useCallback, useRef } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';
import { SavedExplanationDTO } from '../../types/tutor';

export interface NoteLibraryViewProps {
  onClose: () => void;
  onOpenTutorWithNote?: (note: SavedExplanationDTO) => void;
}

export const NoteLibraryView: React.FC<NoteLibraryViewProps> = ({
  onClose,
  onOpenTutorWithNote,
}) => {
  const [notes, setNotes] = useState<SavedExplanationDTO[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [selectedIndex, setSelectedIndex] = useState<number>(0);
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [topicFilter, setTopicFilter] = useState<string>('all');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const searchInputRef = useRef<HTMLInputElement>(null);

  // Fetch notes from authoritative local SQLite endpoint
  const fetchNotes = useCallback(async (query: string, topic: string) => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const params = new URLSearchParams();
      if (query.trim()) params.set('q', query.trim());
      if (topic !== 'all' && topic.trim()) params.set('topic', topic.trim());

      const res = await fetch(`/api/notes?${params.toString()}`);
      if (!res.ok) {
        throw new Error(`Failed to load saved notes (${res.status})`);
      }
      const data: SavedExplanationDTO[] = await res.json();
      setNotes(data || []);
      setSelectedIndex(0);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch notes';
      setErrorMessage(msg);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchNotes(searchQuery, topicFilter);
  }, [fetchNotes, searchQuery, topicFilter]);

  // Extract distinct topics for topic filter pills
  const availableTopics = React.useMemo(() => {
    const set = new Set<string>();
    notes.forEach((n) => {
      if (n.topic) set.add(n.topic);
    });
    return Array.from(set);
  }, [notes]);

  const selectedNote = notes[selectedIndex] || null;

  // Keyboard navigation complying with docs/NAVIGATION.md
  // j/k or Up/Down move list selection, n/p browse notes, Esc closes
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const activeEl = document.activeElement;
      const isSearchActive = activeEl === searchInputRef.current;

      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
        return;
      }

      if (isSearchActive) {
        return;
      }

      if (e.key === 'j' || e.key === 'ArrowDown' || e.key === 'n') {
        e.preventDefault();
        if (notes.length > 0) {
          setSelectedIndex((prev) => (prev < notes.length - 1 ? prev + 1 : prev));
        }
      } else if (e.key === 'k' || e.key === 'ArrowUp' || e.key === 'p') {
        e.preventDefault();
        if (notes.length > 0) {
          setSelectedIndex((prev) => (prev > 0 ? prev - 1 : 0));
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [notes.length, onClose]);

  // Handle note deletion
  const handleDeleteNote = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this saved note? This does not alter practice records.')) {
      return;
    }

    try {
      const res = await fetch(`/api/notes/${id}`, { method: 'DELETE' });
      if (!res.ok) {
        throw new Error('Failed to delete note');
      }
      fetchNotes(searchQuery, topicFilter);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  // Handle Markdown export download
  const handleExportNote = (note: SavedExplanationDTO) => {
    const filename = `note-${note.id}.md`;
    const exportUrl = `/api/notes/${note.id}/export`;

    // Trigger direct browser download
    const link = document.createElement('a');
    link.href = exportUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div
      className="note-library-container"
      id="note-library-view"
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        minHeight: '80vh',
        background: 'var(--bg-app)',
      }}
    >
      {/* Top Header Bar */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '1rem 1.5rem',
          background: 'var(--bg-surface)',
          borderBottom: '1px solid var(--border-subtle)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 600, color: 'var(--text-main)' }}>
            📚 Saved Explanations & Notes Library
          </h2>
          <span className="badge" style={{ background: 'var(--bg-surface-elevated)', color: 'var(--text-muted)' }}>
            {notes.length} saved
          </span>
        </div>

        <button className="btn btn-secondary" id="btn-notes-close" onClick={onClose} title="Return to practice (Esc)">
          ← Return to Practice (Esc)
        </button>
      </div>

      {/* Filter and Search Bar */}
      <div
        style={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '1rem',
          padding: '0.85rem 1.5rem',
          background: 'var(--bg-surface-elevated)',
          borderBottom: '1px solid var(--border-subtle)',
        }}
      >
        {/* Search Input */}
        <div style={{ flex: 1, minWidth: '260px', maxWidth: '450px' }}>
          <input
            id="notes-search-input"
            ref={searchInputRef}
            type="text"
            className="input-field"
            placeholder="Search notes by title, topic, concept, formula... (Esc to leave)"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ width: '100%' }}
          />
        </div>

        {/* Topic Filter Pills */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', flexWrap: 'wrap' }}>
          <button
            className={`btn ${topicFilter === 'all' ? 'btn-primary' : 'btn-secondary'}`}
            style={{ fontSize: '0.75rem', padding: '0.25rem 0.65rem' }}
            onClick={() => setTopicFilter('all')}
          >
            All Topics
          </button>
          {availableTopics.map((top) => (
            <button
              key={top}
              className={`btn ${topicFilter === top ? 'btn-primary' : 'btn-secondary'}`}
              style={{ fontSize: '0.75rem', padding: '0.25rem 0.65rem' }}
              onClick={() => setTopicFilter(top)}
            >
              {top}
            </button>
          ))}
        </div>
      </div>

      {/* Main Dual-Pane Content */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Left: Notes List */}
        <div
          id="notes-list-pane"
          style={{
            width: '340px',
            borderRight: '1px solid var(--border-subtle)',
            background: 'var(--bg-surface)',
            overflowY: 'auto',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {isLoading ? (
            <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
              Loading notes...
            </div>
          ) : errorMessage ? (
            <div style={{ padding: '1rem', color: 'var(--accent-rose)' }}>{errorMessage}</div>
          ) : notes.length === 0 ? (
            <div style={{ padding: '2rem 1rem', textAlign: 'center', color: 'var(--text-subtle)' }}>
              <p style={{ marginBottom: '0.5rem' }}>No saved notes found.</p>
              <p style={{ fontSize: '0.8rem' }}>
                Use the AI Tutor during practice drills to save step-by-step explanations and math derivations here.
              </p>
            </div>
          ) : (
            notes.map((note, idx) => {
              const isSelected = idx === selectedIndex;
              const title = note.provider_info?.title || `${note.topic} Note`;
              const dateStr = note.updated_at
                ? new Date(note.updated_at).toLocaleDateString(undefined, {
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })
                : '';

              return (
                <div
                  key={note.id}
                  id={`note-item-${idx}`}
                  className={`note-list-item ${isSelected ? 'selected' : ''}`}
                  onClick={() => setSelectedIndex(idx)}
                  style={{
                    padding: '0.85rem 1rem',
                    borderBottom: '1px solid var(--border-subtle)',
                    cursor: 'pointer',
                    background: isSelected ? 'var(--bg-surface-elevated)' : 'transparent',
                    borderLeft: isSelected ? '3px solid var(--accent-cyan)' : '3px solid transparent',
                    transition: 'background 0.15s ease',
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--accent-cyan)', fontWeight: 600 }}>
                      {note.topic}
                    </span>
                    <span style={{ fontSize: '0.7rem', color: 'var(--text-subtle)' }}>{dateStr}</span>
                  </div>

                  <h4
                    style={{
                      fontSize: '0.9rem',
                      fontWeight: 600,
                      color: isSelected ? 'var(--text-main)' : 'var(--text-muted)',
                      marginBottom: '0.35rem',
                    }}
                  >
                    {title}
                  </h4>

                  {note.provider_info?.concepts && note.provider_info.concepts.length > 0 && (
                    <div style={{ display: 'flex', gap: '0.3rem', flexWrap: 'wrap' }}>
                      {note.provider_info.concepts.slice(0, 2).map((c) => (
                        <span key={c} className="group-tag" style={{ fontSize: '0.65rem' }}>
                          {c}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>

        {/* Right: Note Detail View */}
        <div
          id="note-detail-pane"
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: '1.5rem 2rem',
            background: 'var(--bg-app)',
          }}
        >
          {selectedNote ? (
            <div>
              {/* Note Header & Action Buttons */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                  marginBottom: '1rem',
                  paddingBottom: '1rem',
                  borderBottom: '1px solid var(--border-subtle)',
                }}
              >
                <div>
                  <h1
                    id="selected-note-title"
                    style={{ fontSize: '1.4rem', fontWeight: 700, marginBottom: '0.35rem' }}
                  >
                    {selectedNote.provider_info?.title || `${selectedNote.topic} Explanation`}
                  </h1>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                    <span>Topic: <strong>{selectedNote.topic}</strong></span>
                    <span>•</span>
                    <span>Provider: <strong>{selectedNote.provider_info?.provider || 'offline'}</strong></span>
                    <span>•</span>
                    <span>ID: <code style={{ color: 'var(--text-subtle)' }}>{selectedNote.id}</code></span>
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button
                    id="btn-note-export"
                    className="btn btn-secondary"
                    onClick={() => handleExportNote(selectedNote)}
                    title="Export as Obsidian/Typora Markdown file"
                  >
                    📥 Export (.md)
                  </button>

                  {onOpenTutorWithNote && (
                    <button
                      id="btn-note-tutor"
                      className="btn btn-secondary"
                      onClick={() => onOpenTutorWithNote(selectedNote)}
                      title="Ask follow-up questions about this note"
                    >
                      💡 Ask Follow-up
                    </button>
                  )}

                  <button
                    id="btn-note-delete"
                    className="btn btn-ghost"
                    style={{ color: 'var(--accent-rose)' }}
                    onClick={() => handleDeleteNote(selectedNote.id)}
                    title="Delete note"
                  >
                    🗑 Delete
                  </button>
                </div>
              </div>

              {/* Advisory Callout */}
              <div
                style={{
                  padding: '0.65rem 0.85rem',
                  background: 'rgba(251, 191, 36, 0.08)',
                  border: '1px solid rgba(251, 191, 36, 0.25)',
                  borderRadius: 'var(--radius-sm)',
                  fontSize: '0.8rem',
                  color: 'var(--accent-amber)',
                  marginBottom: '1.5rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                }}
              >
                <span>ℹ️</span>
                <span>
                  <strong>Advisory Note:</strong> Saved for self-study and reference. Does not alter graded evidence or official exam keys.
                </span>
              </div>

              {/* Concepts Tags */}
              {selectedNote.provider_info?.concepts && selectedNote.provider_info.concepts.length > 0 && (
                <div style={{ marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>Covered Concepts:</span>
                  {selectedNote.provider_info.concepts.map((c) => (
                    <span key={c} className="group-tag" style={{ fontSize: '0.75rem' }}>
                      {c}
                    </span>
                  ))}
                </div>
              )}

              {/* Rendered Math Markdown */}
              <div
                id="selected-note-content"
                style={{
                  background: 'var(--bg-surface)',
                  padding: '1.5rem',
                  borderRadius: 'var(--radius-md)',
                  border: '1px solid var(--border-subtle)',
                }}
              >
                <MathMarkdown content={selectedNote.raw_markdown} />
              </div>
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: '4rem 2rem', color: 'var(--text-subtle)' }}>
              Select a note from the left list or search by keyword to view math and explanations.
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
