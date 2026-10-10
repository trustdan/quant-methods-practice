import React, { useEffect, useRef, useState } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';
interface Proposal { family_id: string; n: number; p: number; k: number; title: string; scenario_markdown: string; success_label: string }
interface Stage {
  id: string; prompt_markdown: string; explanation_markdown: string;
  options: { id: string; text_markdown: string; hint_markdown: string | null; misconception_id: string | null }[];
  expected_answer: { option_id?: string; value?: number; units?: string };
  numeric_policy: { absolute_tolerance: number; relative_tolerance: number; allowed_forms: string[] } | null;
}
interface Candidate {
  id: string; revision: number; status: string; proposal: Proposal;
  source: { mode: string; route: string; model: string; seed: number }; created_at: string;
  template: { title: string; version: number; scenario_markdown: string; assumptions: string[]; stages: Stage[]; setting_group: string };
  reviews: { action: string; reviewer: string; notes: string; timestamp: string }[];
}
async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  if (!response.ok) throw new Error(await response.text());
  return response.json();
}
export const CandidateReview: React.FC<{ active: boolean }> = ({ active }) => {
  const [records, setRecords] = useState<Candidate[]>([]);
  const [selected, setSelected] = useState<Candidate | null>(null);
  const [seed, setSeed] = useState('0');
  const [manual, setManual] = useState('');
  const importRef = useRef<HTMLDetailsElement>(null);
  const [reviewer, setReviewer] = useState('');
  const [notes, setNotes] = useState('');
  const [confirmed, setConfirmed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [route, setRoute] = useState('offline');
  const [model, setModel] = useState('');
  const controller = useRef<AbortController | null>(null);
  const previewRef = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    if (!active) return;
    const abort = new AbortController();
    request<Candidate[]>('/api/candidates', { signal: abort.signal }).then(setRecords).catch((e: unknown) => {
      if (!abort.signal.aborted) setError(e instanceof Error ? e.message : 'Cannot load candidates');
    });
    request<{ active_route: string; active_model: string }>('/api/providers', { signal: abort.signal }).then(data => {
      setRoute(data.active_route); setModel(data.active_model);
    }).catch(() => { /* local variation still works */ });
    return () => { abort.abort(); controller.current?.abort(); };
  }, [active]);
  function choose(record: Candidate) {
    setSelected(record); setConfirmed(false); setNotes(''); setError(''); setNotice('');
    setTimeout(() => previewRef.current?.focus(), 0);
  }
  async function generate(mode: 'local' | 'manual' | 'ai') {
    const numericSeed = Number(seed);
    if (!Number.isSafeInteger(numericSeed)) { setError('Seed must be a safe integer.'); return; }
    let proposal: unknown;
    if (mode === 'manual') {
      try { proposal = JSON.parse(manual); } catch { setError('Enter a valid proposal JSON object.'); return; }
    }
    const abort = new AbortController(); controller.current = abort;
    setBusy(true); setError(''); setNotice('');
    try {
      const rec = await request<Candidate>('/api/candidates', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: abort.signal,
        body: JSON.stringify({ mode, seed: numericSeed, ...(mode === 'manual' ? { proposal } : {}) }),
      });
      if (abort.signal.aborted) return;
      setRecords(old => [rec, ...old]); choose(rec); setNotice('Draft saved. Review every stage before approval.');
    } catch (e) { if (!abort.signal.aborted) setError(e instanceof Error ? e.message : 'Candidate could not be saved'); }
    finally { if (controller.current === abort) { controller.current = null; setBusy(false); } }
  }
  async function review(action: 'approve' | 'reject' | 'retire') {
    if (!selected) return;
    const abort = new AbortController(); controller.current = abort;
    setBusy(true); setError(''); setNotice('');
    try {
      const rec = await request<Candidate>(`/api/candidates/${selected.id}/review`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: abort.signal,
        body: JSON.stringify({ action, expected_revision: selected.revision, reviewer, notes, semantic_confirmed: confirmed }),
      });
      if (abort.signal.aborted) return;
      setRecords(old => old.map(item => item.id === rec.id ? rec : item)); setSelected(rec); setConfirmed(false);
      setNotice(action === 'approve' ? 'Approved for future practice sessions. Existing sessions keep their saved questions.' : action === 'retire' ? 'Retired from future practice. Past sessions are preserved.' : 'Candidate rejected and kept in review history.');
    } catch (e) { if (!abort.signal.aborted) setError(e instanceof Error ? e.message : 'Review could not be saved'); }
    finally { if (controller.current === abort) { controller.current = null; setBusy(false); } }
  }
  async function exportBank() {
    setError('');
    try {
      const data = await request<unknown[]>('/api/candidates/export');
      const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }));
      const link = document.createElement('a'); link.href = url; link.download = 'approved-bank.json'; link.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (e) { setError(e instanceof Error ? e.message : 'Export failed'); }
  }
  const canReview = !!reviewer.trim() && !!notes.trim() && !busy;
  return <section className="candidate-review" aria-label="Question candidates" hidden={!active}>
    <div className="card">
      <div className="card-header"><h2 className="card-title">Question candidates</h2><button className="btn btn-outline" onClick={exportBank}>Export approved bank</button></div>
      <p>Start with a binomial “exactly” problem. Go derives the answers. Every new scenario needs your review before it enters practice.</p>
      <p className="candidate-muted">Wording and parameter variations share the existing reasoning group and do not establish transfer on their own.</p>
      <div className="candidate-actions">
        <label>Variation seed <input type="number" step="1" value={seed} onChange={e => setSeed(e.target.value)} disabled={busy} /></label>
        <button className="btn btn-primary" disabled={busy} onClick={() => generate('local')}>Generate local variation</button>
        <button className="btn btn-secondary" disabled={busy || route === 'offline'} onClick={() => generate('ai')}>Request AI wording</button>
        {busy && <button className="btn btn-outline" onClick={() => controller.current?.abort()}>Cancel request</button>}
      </div>
      <p className="candidate-muted">AI route: {route}{model ? ` / ${model}` : ''}. An AI wording request uses the selected provider’s request budget and billing route. Select it in Settings first.</p>
      <details ref={importRef}><summary>Import or revise proposal wording</summary>
        <p>Paste a proposal with only family_id, n, p, k, title, scenario_markdown, and success_label. Saving creates a separate draft; approved content stays immutable.</p>
        <label htmlFor="candidate-proposal-json">Proposal JSON</label><textarea id="candidate-proposal-json" rows={8} value={manual} onChange={e => setManual(e.target.value)} />
        <button className="btn btn-outline" disabled={busy || !manual.trim()} onClick={() => generate('manual')}>Save proposal as new draft</button>
      </details>
      {busy && <p role="status">Working…</p>}
      {error && <p role="alert">{error} Your review text is preserved. Reload the candidate list if another tab changed it.</p>}
      {notice && <p role="status">{notice}</p>}
    </div>
    <div className="candidate-layout">
      <aside className="card" aria-label="Saved candidates">
        <h3>Saved candidates</h3>
        {!records.length && <p>No candidates yet.</p>}
        {records.map(rec => <button key={rec.id} className={`btn ${selected?.id === rec.id ? 'btn-primary' : 'btn-outline'}`} disabled={busy} onClick={() => choose(rec)}>{rec.template.title} — {rec.status}</button>)}
      </aside>
      {selected && <article className="card" aria-label="Candidate preview">
        <h3 tabIndex={-1} ref={previewRef}>Preview: {selected.template.title}</h3>
        <p>Status: <strong>{selected.status}</strong> · version {selected.template.version} · preview revision {selected.revision}</p>
        <p className="candidate-muted">Source: {selected.source.mode} / {selected.source.route} / {selected.source.model} · seed {selected.source.seed} · {selected.created_at}</p>
        <MathMarkdown content={selected.template.scenario_markdown} />
        <button className="btn btn-outline" disabled={busy} onClick={() => {
          setManual(JSON.stringify(selected.proposal, null, 2));
          if (importRef.current) { importRef.current.open = true; importRef.current.scrollIntoView({ block: 'start' }); }
        }}>Copy proposal to editor</button>
        <h4>Required assumptions — verify the scenario actually supports each</h4>
        <ul>{selected.template.assumptions.map(a => <li key={a}>{a}</li>)}</ul>
        <p>Parameters: n={selected.proposal.n}, p={selected.proposal.p}, k={selected.proposal.k}; event: exactly k; probability has no physical units.</p>
        <p>Validation passed for the supported family, parameter bounds and template structure. Semantic correctness of the story requires your review.</p>
        {selected.template.stages.map((stage, index) => <section className="candidate-stage" key={stage.id}>
          <h4>Stage {index + 1}: {stage.id.replaceAll('_', ' ')}</h4>
          <MathMarkdown content={stage.prompt_markdown} />
          {stage.options.map(option => <div className="candidate-option" key={option.id}>
            <strong>{option.id === stage.expected_answer.option_id ? 'Correct option' : 'Distractor'}</strong>
            <MathMarkdown content={option.text_markdown} />
            {option.hint_markdown && <><span>Hint ({option.misconception_id}):</span><MathMarkdown content={option.hint_markdown} /></>}
          </div>)}
          {stage.expected_answer.value !== undefined && <p>Canonical answer from Go: <strong>{stage.expected_answer.value}</strong> {stage.expected_answer.units}</p>}
          {stage.numeric_policy && <p>Tolerance: absolute {stage.numeric_policy.absolute_tolerance}, relative {stage.numeric_policy.relative_tolerance}. Forms: {stage.numeric_policy.allowed_forms.join(', ')}.</p>}
          <MathMarkdown content={stage.explanation_markdown} />
        </section>)}
        {selected.reviews.map((r, index) => <div key={index}><p>{r.action} by {r.reviewer} at {r.timestamp}</p><p>{r.notes}</p></div>)}
        {(selected.status === 'pending' || selected.status === 'approved') && <div className="candidate-review-form">
          <label>Reviewer name<input value={reviewer} maxLength={200} onChange={e => setReviewer(e.target.value)} /></label>
          <label htmlFor="candidate-review-notes">Review notes</label><textarea id="candidate-review-notes" rows={3} value={notes} maxLength={4000} onChange={e => setNotes(e.target.value)} />
          {selected.status === 'pending' && <label className="candidate-confirm"><input type="checkbox" checked={confirmed} onChange={e => setConfirmed(e.target.checked)} /> I reviewed the scenario, assumptions, event, every stage, answers, distractors and hints, and confirm they agree.</label>}
          <div className="candidate-actions">
            {selected.status === 'pending' ? <>
              <button className="btn btn-primary" disabled={!canReview || !confirmed} onClick={() => review('approve')}>Approve for practice</button>
              <button className="btn btn-outline" disabled={!canReview} onClick={() => review('reject')}>Reject candidate</button>
            </> : <button className="btn btn-outline" disabled={!canReview} onClick={() => review('retire')}>Retire from future practice</button>}
          </div>
        </div>}
      </article>}
    </div>
  </section>;
};
