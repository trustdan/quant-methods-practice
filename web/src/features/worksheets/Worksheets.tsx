import React, { useEffect, useRef, useState } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';
import type { AnswerMap, CasePreview, Worksheet } from './types';

export type WorksheetMode = 'full_solution' | 'dataset';
const sample = 'experiment_id,successes\nrun_1,2\nrun_2,1\nrun_3,4\nrun_4,2\nrun_5,0\n';
const answersFrom = (record: Worksheet): AnswerMap => Object.fromEntries(record.items.filter(item => item.status !== 'completed' && (item.draft_answer || item.attempts?.length)).map(item => [item.key, item.draft_answer ?? item.attempts!.at(-1)!.submitted_answer]));

export const Worksheets: React.FC<{ active: boolean; launchMode: WorksheetMode }> = ({ active, launchMode }) => {
  const [mode, setMode] = useState<WorksheetMode>(launchMode);
  const [bank, setBank] = useState<{ id: string; title: string }[]>([]);
  const [template, setTemplate] = useState('binomial_fair_coin_exactly_two');
  const [records, setRecords] = useState<Worksheet[]>([]);
  const [record, setRecord] = useState<Worksheet | null>(null);
  const [answers, setAnswers] = useState<AnswerMap>({});
  const [dirty, setDirty] = useState(false);
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState('');
  const [error, setError] = useState('');
  const [csv, setCSV] = useState(sample);
  const [preview, setPreview] = useState<CasePreview | null>(null);
  const [reviewer, setReviewer] = useState('');
  const [sourceNote, setSourceNote] = useState('');
  const [reviewed, setReviewed] = useState(false);
  const controller = useRef<AbortController | null>(null);
  const selected = useRef<Worksheet | null>(null);
  const dirtyRef = useRef(false);
  // Retain the exact command across an uncertain network failure. A retry cannot
  // consume an extra attempt even if the first response was lost.
  const pending = useRef<{ command_id: string; expected_revision: number; type: string; answers: AnswerMap } | null>(null);
  const [uncertain, setUncertain] = useState(false);
  const fileVersion = useRef(0);
  selected.current = record;
  dirtyRef.current = dirty;

  useEffect(() => { setMode(launchMode); }, [launchMode]);
  useEffect(() => {
    if (!active) return;
    const ac = new AbortController(); controller.current = ac;
    setBusy(true);
    Promise.all([fetch('/api/worksheets', { signal: ac.signal }), fetch('/api/bank', { signal: ac.signal })])
      .then(async ([listResponse, bankResponse]) => {
        if (!listResponse.ok) throw new Error(await listResponse.text());
        if (!bankResponse.ok) throw new Error('Cannot load approved questions.');
        const [list, questions] = await Promise.all([listResponse.json(), bankResponse.json()]) as [Worksheet[], { id: string; title: string }[]];
        if (ac.signal.aborted) return;
        setRecords(list); setBank(questions);
        if (!dirtyRef.current && !pending.current) {
          const current = list.find(item => item.id === selected.current?.id) ?? list[0];
          if (current) { setRecord(current); setAnswers(answersFrom(current)); }
        }
      }).catch((err: unknown) => { if (!ac.signal.aborted) setError(String(err)); })
      .finally(() => { if (!ac.signal.aborted) setBusy(false); });
    return () => { ac.abort(); controller.current = null; fileVersion.current++; };
  }, [active]);

  useEffect(() => {
    if (!dirty && !uncertain) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ''; };
    window.addEventListener('beforeunload', warn);
    return () => window.removeEventListener('beforeunload', warn);
  }, [dirty, uncertain]);

  const choose = (next: Worksheet) => { setRecord(next); setAnswers(answersFrom(next)); setDirty(false); setNotice(''); setError(''); };
  const accept = (next: Worksheet) => {
    choose(next); setRecords(previous => [next, ...previous.filter(item => item.id !== next.id)]);
  };
  const run = async (work: (signal: AbortSignal) => Promise<void>) => {
    const ac = controller.current;
    if (!ac || ac.signal.aborted || busy) return;
    setBusy(true); setError(''); setNotice('');
    try { await work(ac.signal); }
    catch (err) { if (!ac.signal.aborted) setError(err instanceof Error ? err.message : String(err)); }
    finally { if (!ac.signal.aborted) setBusy(false); }
  };
  const post = async (path: string, data: unknown, signal: AbortSignal): Promise<Response> => {
    const response = await fetch(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data), signal });
    return response;
  };
  const create = () => run(async signal => {
    const response = await post('/api/worksheets', mode === 'dataset'
      ? { mode, seed: 42, csv, reviewer, source_note: sourceNote, reviewed }
      : { mode, seed: 42, template_id: template }, signal);
    if (!response.ok) throw new Error(await response.text());
    const next = await response.json() as Worksheet;
    if (!signal.aborted) { accept(next); setPreview(null); setReviewed(false); setNotice('Worksheet saved. Fill every field, then submit the whole solution.'); }
  });
  const command = (type: 'save_draft' | 'submit') => run(async signal => {
    if (!record) return;
    if (!pending.current) pending.current = { command_id: crypto.randomUUID(), expected_revision: record.revision, type, answers: structuredClone(answers) };
    setUncertain(true);
    const action = pending.current.type;
    const response = await post(`/api/worksheets/${record.id}/commands`, pending.current, signal);
    if (!response.ok) {
      // A validation/conflict response guarantees this command did not commit.
      if (response.status < 500) { pending.current = null; if (!signal.aborted) setUncertain(false); }
      throw new Error(await response.text());
    }
    const next = await response.json() as Worksheet;
    if (!signal.aborted) { pending.current = null; setUncertain(false); accept(next); setNotice(action === 'save_draft' ? 'Draft saved on this device.' : next.status === 'retry' ? 'Review one hint for each incorrect field, then retry those fields once.' : 'Worksheet completed. Your submission and review are saved.'); }
  });
  const previewCSV = () => run(async signal => {
    const response = await post('/api/worksheets/dataset-preview', { csv }, signal);
    if (!response.ok) throw new Error(await response.text());
    const next = await response.json() as CasePreview;
    if (!signal.aborted) { setPreview(next); setReviewed(false); }
  });
  const changeCSV = (text: string) => { setCSV(text); setPreview(null); setReviewed(false); };
  const readFile = async (file: File | undefined) => {
    if (!file) return;
    const version = ++fileVersion.current;
    if (file.size > 65536) { setError('CSV must be at most 64 KiB.'); return; }
    try { const text = await file.text(); if (version === fileVersion.current && controller.current) { changeCSV(text); setError(''); } }
    catch { if (version === fileVersion.current) setError('Could not read that file.'); }
  };
  const update = (key: string, kind: 'choice' | 'numeric', value: string) => {
    setAnswers(previous => ({ ...previous, [key]: kind === 'choice' ? { kind, option_id: value } : { kind, numeric_raw: value } })); setDirty(true); setNotice('');
  };
  const unfinished = record?.items.filter(item => item.status !== 'completed') ?? [];
  const ready = unfinished.length > 0 && unfinished.every(item => item.kind === 'choice' ? answers[item.key]?.option_id : answers[item.key]?.numeric_raw?.trim());

  return <section hidden={!active} className="worksheet-workspace" aria-label="Worksheets">
    <div className="card">
      <h2>Full solutions &amp; dataset cases</h2>
      <p>Complete the method, assumptions, calculation and interpretation together. These structured fields provide support; worksheet results are kept separate from independent mastery.</p>
      <div className="worksheet-actions">
        <label>Worksheet type <select value={mode} onChange={event => setMode(event.target.value as WorksheetMode)} disabled={busy}>
          <option value="full_solution">Full solution</option><option value="dataset">CSV dataset case</option>
        </select></label>
        {mode === 'full_solution' && <label>Approved question <select value={template} onChange={event => setTemplate(event.target.value)} disabled={busy}>
          {bank.map(question => <option key={question.id} value={question.id}>{question.title}</option>)}
        </select></label>}
      </div>
      {mode === 'dataset' && <div>
        <p>Generic four-toss case: one row is one complete experiment; successes is the number of heads (0–4). The file cannot establish fairness or independence. Maximum 500 rows and 64 KiB; formulas and extra columns are rejected.</p>
        <label htmlFor="worksheet-file">Choose CSV file</label><input id="worksheet-file" type="file" accept=".csv,text/csv" disabled={busy} onChange={event => void readFile(event.target.files?.[0])} />
        <label htmlFor="worksheet-csv">CSV data</label><textarea id="worksheet-csv" rows={7} value={csv} disabled={busy} onChange={event => changeCSV(event.target.value)} />
        <button className="btn btn-outline" disabled={busy} onClick={() => void previewCSV()}>Preview CSV</button>
        {preview && <div aria-label="CSV preview">
          <p>{preview.dataset.rows.length} experiments accepted. Showing the first {Math.min(10, preview.dataset.rows.length)} rows.</p>
          <table><thead><tr><th>Experiment</th><th>Heads in four tosses</th></tr></thead><tbody>{preview.dataset.rows.slice(0, 10).map(row => <tr key={row.experiment_id}><td>{row.experiment_id}</td><td>{row.successes}</td></tr>)}</tbody></table>
          <h3>Review case wording, conditions and keys</h3>
          <p>This author preview shows answers and hints. Creating the case records reference exposure; this is supported practice.</p>
          {preview.questions.map(question => <article key={question.id}>
            <h4>{question.title}</h4><MathMarkdown content={question.scenario_markdown} />
            <ul>{question.assumptions.map(assumption => <li key={assumption}>{assumption}</li>)}</ul>
            {question.stages.map(stage => <div key={stage.id}>
              <MathMarkdown content={stage.prompt_markdown} />
              {stage.options?.map(option => <div key={option.id}><MathMarkdown content={`${option.id}: ${option.text_markdown}${option.hint_markdown ? ` Hint: ${option.hint_markdown}` : ''}`} /></div>)}
              <MathMarkdown content={`Expected: ${stage.expected_answer.option_id ?? stage.expected_answer.value}. ${stage.explanation_markdown}`} />
              {stage.numeric_policy && <p>Numeric policy v{stage.numeric_policy.version}: absolute tolerance {stage.numeric_policy.absolute_tolerance}; relative tolerance {stage.numeric_policy.relative_tolerance}.</p>}
            </div>)}
          </article>)}
          <label htmlFor="dataset-reviewer">Dataset reviewer</label><input id="dataset-reviewer" value={reviewer} maxLength={120} disabled={busy} onChange={event => { setReviewer(event.target.value); setReviewed(false); }} />
          <label htmlFor="dataset-source">Data source and row meaning</label><textarea id="dataset-source" value={sourceNote} maxLength={2000} disabled={busy} onChange={event => { setSourceNote(event.target.value); setReviewed(false); }} />
          <label><input type="checkbox" checked={reviewed} disabled={busy} onChange={event => setReviewed(event.target.checked)} /> I reviewed every case field, option, hint and answer. Each row is one four-toss experiment and successes counts heads. Fairness and independence are assumptions of the separate theoretical exercise.</label>
        </div>}
      </div>}
      <button className="btn btn-primary" disabled={busy || dirty || uncertain || (mode === 'dataset' ? !preview || !reviewed || !reviewer.trim() || !sourceNote.trim() : bank.length === 0)} onClick={() => void create()}>Create worksheet</button>
      <p role="status">{notice || (dirty ? 'Unsaved changes — save your draft before closing the app.' : '')}</p>
      {error && <p role="alert">{error}</p>}
      {uncertain && !busy && <button className="btn btn-primary" onClick={() => void command('submit')}>Retry pending save or submission</button>}
    </div>
    <div className="card">
      <label>Saved worksheet <select value={record?.id ?? ''} disabled={busy || dirty || uncertain} onChange={event => { const next = records.find(item => item.id === event.target.value); if (next) choose(next); }}>
        {!records.length && <option value="">No saved worksheets</option>}
        {records.map(item => <option key={item.id} value={item.id}>{item.questions[0]?.title} — {item.status} — {new Date(item.updated_at).toLocaleString()} ({item.id.slice(-8)})</option>)}
      </select></label>
      {dirty && !uncertain && <button className="btn btn-outline" disabled={busy} onClick={() => record && choose(record)}>Discard unsaved edits</button>}
      {record && <>
        <p>Saved status: {record.status}. {record.evidence}</p>
        <div className="worksheet-actions">
          <button className="btn btn-outline" disabled={busy || uncertain || record.status === 'completed'} onClick={() => void command('save_draft')}>Save worksheet draft</button>
          <button className="btn btn-primary" disabled={busy || uncertain || !ready} onClick={() => void command('submit')}>{record.status === 'retry' ? 'Submit retry' : 'Submit full solution'}</button>
          <a className="btn btn-outline" href={`/api/worksheets/${record.id}/export`} download>Download worksheet</a>
          <button className="btn btn-outline" disabled={busy || uncertain} onClick={() => void run(async signal => {
            const response = await fetch(`/api/worksheets/${record.id}`, { signal }); if (!response.ok) throw new Error(await response.text());
            const next = await response.json() as Worksheet; if (!signal.aborted) accept(next);
          })}>Reload saved version</button>
        </div>
        {record.questions.map(question => <article key={question.id} aria-label={question.title}>
          <h3>{question.title}</h3><MathMarkdown content={question.scenario_markdown} />
          <ul>{question.assumptions?.map(assumption => <li key={assumption}>{assumption}</li>)}</ul>
          {record.items.filter(item => item.question_id === question.id).map((item, index) => <fieldset key={item.key} disabled={busy || uncertain || item.status === 'completed'}>
            <legend>{index + 1}. {item.stage_id.replaceAll('_', ' ')}</legend>
            <MathMarkdown content={item.prompt_markdown} />
            {item.kind === 'choice' ? item.options.map(option => <label className="worksheet-option" key={option.id}>
              <input type="radio" name={item.key} value={option.id} checked={(answers[item.key]?.option_id ?? item.attempts?.at(-1)?.submitted_answer.option_id) === option.id} onChange={() => update(item.key, 'choice', option.id)} />
              <MathMarkdown content={option.text_markdown} />
            </label>) : <>
              <label htmlFor={item.key}>Numeric answer ({item.stage_id.replaceAll('_', ' ')})</label>
              <input id={item.key} type="text" inputMode="decimal" autoComplete="off" value={answers[item.key]?.numeric_raw ?? item.attempts?.at(-1)?.submitted_answer.numeric_raw ?? ''} onChange={event => update(item.key, 'numeric', event.target.value)} />
              <p>Decimal, percentage or fraction. Absolute tolerance {item.numeric_policy?.absolute_tolerance}; relative tolerance {item.numeric_policy?.relative_tolerance}.</p>
            </>}
            {item.attempts?.map(attempt => <div key={attempt.attempt_number} role="status"><p>Attempt {attempt.attempt_number}: {attempt.is_correct ? 'Correct' : 'Incorrect'} — {attempt.assistance.join(', ').replaceAll('_', ' ')}</p><MathMarkdown content={attempt.feedback_markdown} /></div>)}
          </fieldset>)}
        </article>)}
        {record.dataset && <details><summary>Saved dataset ({record.dataset.rows.length} experiments)</summary><p>Reviewed by {record.dataset.reviewer}. {record.dataset.source_note}</p>
          <table><thead><tr><th>Experiment</th><th>Heads</th></tr></thead><tbody>{record.dataset.rows.map(row => <tr key={row.experiment_id}><td>{row.experiment_id}</td><td>{row.successes}</td></tr>)}</tbody></table>
        </details>}
      </>}
    </div>
  </section>;
};
