import React from 'react';

export interface StageInfo {
  id: string;
  number: number;
  label: string;
  completed?: boolean;
  revealed?: boolean;
  locked?: boolean;
  lockReason?: string;
}

interface StageStripProps {
  stages: StageInfo[];
  currentStageIndex: number;
  onSelectStage: (index: number) => void;
}

export const StageStrip: React.FC<StageStripProps> = ({
  stages,
  currentStageIndex,
  onSelectStage,
}) => {
  return (
    <div className="stage-strip" role="tablist" aria-label="Problem Stages">
      {stages.map((stage, idx) => {
        const isActive = idx === currentStageIndex;
        const isCompleted = !!stage.completed;
        const isRevealed = !!stage.revealed;
        const isLocked = !!stage.locked;

        let statusClass = '';
        if (isActive) statusClass += ' active';
        if (isCompleted) statusClass += ' completed';
        if (isRevealed) statusClass += ' revealed';
        if (isLocked) statusClass += ' locked';

        const tooltipTitle = isLocked
          ? (stage.lockReason || `Stage ${stage.number} (${stage.label}) is locked: Complete previous stages first`)
          : isRevealed
          ? `Stage ${stage.number}: ${stage.label} (Solution Revealed)`
          : isCompleted
          ? `Stage ${stage.number}: ${stage.label} (Completed)`
          : `Stage ${stage.number}: ${stage.label}`;

        return (
          <button
            key={stage.id}
            role="tab"
            aria-selected={isActive}
            className={`stage-step${statusClass}`}
            onClick={() => onSelectStage(idx)}
            title={tooltipTitle}
          >
            <span className="stage-num">{stage.number}/7</span>
            <span className="stage-name">{stage.label}</span>
            {isCompleted && !isRevealed && <span className="stage-check" aria-hidden="true">✓</span>}
            {isRevealed && <span className="stage-revealed-tag" aria-label="Revealed" style={{ fontSize: '0.65rem', color: '#c084fc' }}>rev</span>}
            {isLocked && <span className="stage-lock" aria-label="Locked">🔒</span>}
          </button>
        );
      })}
    </div>
  );
};
