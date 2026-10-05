import React from 'react';

interface QuestionStripProps {
  totalQuestions: number;
  currentQuestionIndex: number;
  onSelectQuestion: (index: number) => void;
}

export const QuestionStrip: React.FC<QuestionStripProps> = ({
  totalQuestions,
  currentQuestionIndex,
  onSelectQuestion,
}) => {
  return (
    <div className="question-strip" role="navigation" aria-label="Question Navigation">
      {Array.from({ length: totalQuestions }).map((_, idx) => {
        const isActive = idx === currentQuestionIndex;
        return (
          <button
            key={idx}
            className={`question-pill ${isActive ? 'active' : ''}`}
            onClick={() => onSelectQuestion(idx)}
            aria-label={`Question ${idx + 1}`}
            aria-current={isActive ? 'step' : undefined}
          >
            {idx + 1}
          </button>
        );
      })}
    </div>
  );
};
