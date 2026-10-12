import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  SoftwareRecipe,
  SoftwareSpec,
  getRecipeForTemplate,
  getRecipeById,
} from './softwareRecipes';

export interface SoftwareToolDrawerProps {
  recipe?: SoftwareRecipe;
  recipeId?: string;
  templateId?: string;
  title?: string;
  moduleId?: string;
  familyId?: string;
  defaultExpanded?: boolean;
  isExpanded?: boolean;
  onToggleExpanded?: (expanded: boolean) => void;
  activeTool?: 'excel' | 'r';
  onToolChange?: (tool: 'excel' | 'r') => void;
  withheld?: boolean;
  withheldReason?: string;
  compact?: boolean;
  className?: string;
}

const STORAGE_KEY_PREFERRED_TOOL = 'quant_preferred_software_tool';

function getStoredToolPreference(): 'excel' | 'r' {
  try {
    const val = localStorage.getItem(STORAGE_KEY_PREFERRED_TOOL);
    if (val === 'r' || val === 'excel') return val;
  } catch {
    // localStorage not accessible
  }
  return 'excel';
}

function setStoredToolPreference(tool: 'excel' | 'r'): void {
  try {
    localStorage.setItem(STORAGE_KEY_PREFERRED_TOOL, tool);
  } catch {
    // localStorage not accessible
  }
}

export const SoftwareToolDrawer: React.FC<SoftwareToolDrawerProps> = ({
  recipe: initialRecipe,
  recipeId,
  templateId,
  title,
  moduleId,
  familyId,
  defaultExpanded = false,
  isExpanded: controlledExpanded,
  onToggleExpanded,
  activeTool: controlledTool,
  onToolChange,
  withheld = false,
  withheldReason,
  compact = false,
  className = '',
}) => {
  // Uncontrolled vs Controlled expansion state
  const [internalExpanded, setInternalExpanded] = useState<boolean>(defaultExpanded);
  const isExpanded = controlledExpanded !== undefined ? controlledExpanded : internalExpanded;

  // Uncontrolled vs Controlled tool tab ('excel' | 'r')
  const [internalTool, setInternalTool] = useState<'excel' | 'r'>(getStoredToolPreference);
  const activeTool = controlledTool !== undefined ? controlledTool : internalTool;

  // Copy-to-clipboard state
  const [copied, setCopied] = useState<boolean>(false);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Tab button refs for accessible arrow key navigation
  const excelTabRef = useRef<HTMLButtonElement>(null);
  const rTabRef = useRef<HTMLButtonElement>(null);

  // Resolve recipe
  const resolvedRecipe: SoftwareRecipe | undefined = React.useMemo(() => {
    if (initialRecipe) return initialRecipe;
    if (recipeId) {
      const byId = getRecipeById(recipeId);
      if (byId) return byId;
    }
    return getRecipeForTemplate({
      template_id: templateId,
      title,
      module_id: moduleId,
      family_id: familyId,
    });
  }, [initialRecipe, recipeId, templateId, title, moduleId, familyId]);

  // Clean up copy timeout
  useEffect(() => {
    return () => {
      if (copyTimeoutRef.current) {
        clearTimeout(copyTimeoutRef.current);
      }
    };
  }, []);

  const handleToggle = useCallback(() => {
    const next = !isExpanded;
    if (controlledExpanded === undefined) {
      setInternalExpanded(next);
    }
    onToggleExpanded?.(next);
  }, [isExpanded, controlledExpanded, onToggleExpanded]);

  const handleSelectTool = useCallback(
    (tool: 'excel' | 'r') => {
      if (controlledTool === undefined) {
        setInternalTool(tool);
      }
      setStoredToolPreference(tool);
      onToolChange?.(tool);
    },
    [controlledTool, onToolChange]
  );

  // Handle arrow key navigation between tabs according to WAI-ARIA tab pattern
  const handleTabKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>) => {
    if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
      e.preventDefault();
      const nextTool: 'excel' | 'r' = activeTool === 'excel' ? 'r' : 'excel';
      handleSelectTool(nextTool);
      if (nextTool === 'excel') {
        excelTabRef.current?.focus();
      } else {
        rTabRef.current?.focus();
      }
    }
  };

  const handleCopy = useCallback(async (syntax: string) => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(syntax);
      } else {
        // Fallback for older browsers or test environments
        const textArea = document.createElement('textarea');
        textArea.value = syntax;
        textArea.style.position = 'fixed';
        textArea.style.opacity = '0';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
      }
      setCopied(true);
      if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
      copyTimeoutRef.current = setTimeout(() => {
        setCopied(false);
      }, 2000);
    } catch (err) {
      console.warn('Clipboard copy failed:', err);
    }
  }, []);

  if (!resolvedRecipe && !withheld) {
    return null;
  }

  const spec: SoftwareSpec | undefined = resolvedRecipe
    ? activeTool === 'excel'
      ? resolvedRecipe.excel
      : resolvedRecipe.r
    : undefined;

  return (
    <div
      className={`software-tool-drawer ${isExpanded ? 'software-drawer-open' : 'software-drawer-closed'} ${
        compact ? 'software-drawer-compact' : ''
      } ${className}`}
      data-testid="software-tool-drawer"
    >
      {/* Drawer Toggle Header / Trigger Bar */}
      <button
        type="button"
        className="software-drawer-trigger"
        onClick={handleToggle}
        aria-expanded={isExpanded}
        aria-controls="software-tool-content"
        title="Toggle software syntax guide (Press 'u')"
      >
        <div className="software-trigger-left">
          <span className="software-trigger-icon" aria-hidden="true">
            {withheld ? '🔒' : '💡'}
          </span>
          <span className="software-trigger-title">
            {withheld
              ? 'Software Syntax Guide (Withheld in Timed Assessment)'
              : 'Software Syntax Guide: Excel & R'}
          </span>
          {resolvedRecipe && !withheld && (
            <span className="software-recipe-badge" title={resolvedRecipe.statisticalTask}>
              {resolvedRecipe.title}
            </span>
          )}
        </div>

        <div className="software-trigger-right">
          <span className="software-tool-badges" aria-hidden="true">
            <span className={`software-mini-pill ${activeTool === 'excel' ? 'active-pill' : ''}`}>
              📊 Excel
            </span>
            <span className={`software-mini-pill ${activeTool === 'r' ? 'active-pill' : ''}`}>
              🔷 R
            </span>
          </span>
          <span className="software-kbd-badge" title="Shortcut key: press 'u' to toggle">
            <kbd className="kbd">u</kbd>
          </span>
          <span
            className={`software-chevron ${isExpanded ? 'software-chevron-open' : ''}`}
            aria-hidden="true"
          >
            {isExpanded ? '▲' : '▼'}
          </span>
        </div>
      </button>

      {/* Drawer Body (Rendered when expanded) */}
      {isExpanded && (
        <div id="software-tool-content" className="software-drawer-body">
          {withheld ? (
            <div className="software-withheld-notice" role="alert">
              <div className="software-withheld-icon">🔒</div>
              <div className="software-withheld-text">
                <strong>Software Aid Withheld</strong>
                <p>
                  {withheldReason ||
                    'Generic Excel formulas and R syntax are withheld during active timed assessment mode to measure independent problem retrieval.'}
                </p>
                <span className="software-withheld-hint">
                  Full worked Excel equations and R scripts will be revealed in your post-exam score report.
                </span>
              </div>
            </div>
          ) : resolvedRecipe && spec ? (
            <div className="software-content-container">
              {/* Task Header & Context */}
              <div className="software-task-header">
                <span className="software-task-label">Task</span>
                <span className="software-task-desc">{resolvedRecipe.statisticalTask}</span>
              </div>

              {/* Accessible Tool Tabs */}
              <div className="software-tab-bar" role="tablist" aria-label="Statistical Software Tool">
                <button
                  ref={excelTabRef}
                  id="tab-excel"
                  role="tab"
                  type="button"
                  aria-selected={activeTool === 'excel'}
                  aria-controls="panel-excel"
                  className={`software-tab-btn ${activeTool === 'excel' ? 'software-tab-active-excel' : ''}`}
                  onClick={() => handleSelectTool('excel')}
                  onKeyDown={handleTabKeyDown}
                >
                  <span className="software-tab-icon">📊</span>
                  <span className="software-tab-text">Microsoft Excel</span>
                  <span className="software-tab-tag">Spreadsheets</span>
                </button>

                <button
                  ref={rTabRef}
                  id="tab-r"
                  role="tab"
                  type="button"
                  aria-selected={activeTool === 'r'}
                  aria-controls="panel-r"
                  className={`software-tab-btn ${activeTool === 'r' ? 'software-tab-active-r' : ''}`}
                  onClick={() => handleSelectTool('r')}
                  onKeyDown={handleTabKeyDown}
                >
                  <span className="software-tab-icon">🔷</span>
                  <span className="software-tab-text">R / RStudio</span>
                  <span className="software-tab-tag">Statistical Computing</span>
                </button>
              </div>

              {/* Active Tool Panel */}
              <div
                id={`panel-${activeTool}`}
                role="tabpanel"
                aria-labelledby={`tab-${activeTool}`}
                className="software-tabpanel"
              >
                {/* Syntax Card with Copy Action */}
                <div className="software-syntax-card">
                  <div className="software-syntax-topbar">
                    <div className="software-function-name">
                      <span className="software-fn-badge">
                        {activeTool === 'excel' ? 'Formula' : 'Function'}
                      </span>
                      <code>{spec.functionName}</code>
                    </div>

                    <button
                      type="button"
                      className={`btn btn-sm software-copy-btn ${copied ? 'software-copied' : ''}`}
                      onClick={() => handleCopy(spec.genericSyntax)}
                      title="Copy formula/code to clipboard"
                    >
                      {copied ? (
                        <>
                          <span className="software-check-icon" aria-hidden="true">✓</span>
                          <span>Copied!</span>
                        </>
                      ) : (
                        <>
                          <span className="software-copy-icon" aria-hidden="true">📋</span>
                          <span>Copy Syntax</span>
                        </>
                      )}
                    </button>
                  </div>

                  <pre className="software-syntax-box" data-testid="software-syntax-code">
                    <code>{spec.genericSyntax}</code>
                  </pre>
                </div>

                {/* Argument Definitions Table */}
                {spec.argumentDescriptions && spec.argumentDescriptions.length > 0 && (
                  <div className="software-args-section">
                    <h4 className="software-section-heading">Arguments &amp; Parameters</h4>
                    <div className="software-args-table-wrapper">
                      <table className="software-args-table">
                        <thead>
                          <tr>
                            <th scope="col" style={{ width: '28%' }}>Parameter</th>
                            <th scope="col">Definition &amp; Problem Variable</th>
                          </tr>
                        </thead>
                        <tbody>
                          {spec.argumentDescriptions.map((arg, idx) => (
                            <tr key={idx}>
                              <td>
                                <code className="software-arg-code">{arg.name}</code>
                              </td>
                              <td className="software-arg-desc">{arg.description}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                {/* Common Pitfalls / Gotchas */}
                {spec.commonPitfalls && (
                  <div className="software-callout-warning" role="note">
                    <div className="software-callout-header">
                      <span className="software-callout-icon" aria-hidden="true">⚠️</span>
                      <strong>Common Trap / Pitfall to Avoid</strong>
                    </div>
                    <p className="software-callout-body">{spec.commonPitfalls}</p>
                  </div>
                )}

                {/* Pedagogical Notes */}
                {spec.notes && (
                  <div className="software-callout-note" role="note">
                    <div className="software-callout-header">
                      <span className="software-callout-icon" aria-hidden="true">💡</span>
                      <strong>Practical Note &amp; Context</strong>
                    </div>
                    <p className="software-callout-body">{spec.notes}</p>
                  </div>
                )}

                {/* Optional Example Formula */}
                {spec.exampleFormula && (
                  <div className="software-example-row">
                    <span className="software-example-label">Example:</span>
                    <code className="software-example-code">{spec.exampleFormula}</code>
                  </div>
                )}
              </div>
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
};
