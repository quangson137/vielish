"use client";

import { QuizQuestion } from "@/lib/vocab-api";

interface QuizQuestionCardProps {
  question: QuizQuestion;
  index: number;
  onAnswer: (wordId: string, answer: string) => void;
  selectedAnswer?: string;
}

export default function QuizQuestionCard({
  question,
  index,
  onAnswer,
  selectedAnswer,
}: QuizQuestionCardProps) {
  return (
    <div className="border rounded-lg p-4 mb-4">
      <p className="font-medium mb-3">
        {index + 1}. <span className="text-lg">{question.word}</span> nghĩa là
        gì?
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {question.options.map((option) => (
          <button
            key={option}
            onClick={() => onAnswer(question.word_id, option)}
            className={`p-2 text-left rounded border ${
              selectedAnswer === option
                ? "bg-blue-100 border-blue-500"
                : "hover:bg-gray-50"
            }`}
          >
            {option}
          </button>
        ))}
      </div>
    </div>
  );
}
