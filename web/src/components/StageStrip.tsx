import React from 'react';

export interface StageInfo {
  id: string;
  number: number;
  label: string;
  completed?: boolean;
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

        return (
          <button
            key={stage.id}
            role="tab"
            aria-selected={isActive}
            className={`stage-step ${isActive ? 'active' : ''} ${isCompleted ? 'completed' : ''}`}
            onClick={() => onSelectStage(idx)}
            title={`Stage ${stage.number}: ${stage.label}`}
          >
            <span className="stage-num">{stage.number}/7</span>
            <span className="stage-name">{stage.label}</span>
            {isCompleted && <span className="stage-check" aria-hidden="true">✓</span>}
          </button>
        );
      })}
    </div>
  );
};
