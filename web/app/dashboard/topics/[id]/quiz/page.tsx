"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import QuizQuestionCard from "@/components/quiz-question";
import {
  fetchQuiz,
  submitQuiz,
  QuizQuestion,
  QuizResult,
} from "@/lib/vocab-api";

export default function QuizPage() {
  const params = useParams<{ id: string }>();
  const [questions, setQuestions] = useState<QuizQuestion[]>([]);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [result, setResult] = useState<QuizResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchQuiz(params.id)
      .then((data) => setQuestions(data.questions))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [params.id]);

  const handleAnswer = (wordId: string, answer: string) => {
    if (result) return;
    setAnswers((prev) => ({ ...prev, [wordId]: answer }));
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    try {
      const answerList = Object.entries(answers).map(([word_id, answer]) => ({
        word_id,
        answer,
      }));
      const res = await submitQuiz(params.id, answerList);
      setResult(res);
    } catch (err) {
      console.error("Quiz submit failed:", err);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <Link
          href={`/dashboard/topics/${params.id}`}
          className="text-blue-600 hover:underline text-sm"
        >
          ← Quay lại
        </Link>
        <h2 className="text-xl font-bold">Bài kiểm tra</h2>
      </div>

      {questions.map((q, i) => (
        <QuizQuestionCard
          key={q.word_id}
          question={q}
          index={i}
          onAnswer={handleAnswer}
          selectedAnswer={answers[q.word_id]}
        />
      ))}

      {!result ? (
        <button
          onClick={handleSubmit}
          disabled={
            submitting || Object.keys(answers).length !== questions.length
          }
          className="mt-4 px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {submitting ? "Đang nộp..." : "Nộp bài"}
        </button>
      ) : (
        <div className="mt-6 p-4 border rounded-lg bg-gray-50">
          <p className="text-xl font-bold mb-2">
            Kết quả: {result.score}/{result.total}
          </p>
          <div className="space-y-1">
            {result.results.map((r) => (
              <p
                key={r.word_id}
                className={r.correct ? "text-green-600" : "text-red-600"}
              >
                {r.correct ? "✓" : "✗"} Đáp án đúng: {r.correct_answer}
              </p>
            ))}
          </div>
          <Link
            href={`/dashboard/topics/${params.id}`}
            className="inline-block mt-4 px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
          >
            Quay lại chủ đề
          </Link>
        </div>
      )}
    </div>
  );
}
