"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import Flashcard from "@/components/flashcard";
import { fetchTopicWords, submitReview, Word } from "@/lib/vocab-api";

export default function LearnPage() {
  const params = useParams<{ id: string }>();
  const [words, setWords] = useState<Word[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [finished, setFinished] = useState(false);

  useEffect(() => {
    fetchTopicWords(params.id)
      .then(setWords)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [params.id]);

  const handleRate = async (quality: 1 | 3 | 5) => {
    const word = words[currentIndex];
    try {
      await submitReview(word.id, quality);
    } catch (err) {
      console.error("Review submit failed:", err);
    }

    if (currentIndex + 1 < words.length) {
      setCurrentIndex(currentIndex + 1);
    } else {
      setFinished(true);
    }
  };

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  if (words.length === 0) {
    return <p className="text-gray-500">Chưa có từ nào trong chủ đề này.</p>;
  }

  if (finished) {
    return (
      <div className="text-center py-12">
        <p className="text-2xl font-bold mb-4">Hoàn thành!</p>
        <p className="text-gray-600 mb-6">
          Bạn đã học xong {words.length} từ.
        </p>
        <div className="flex justify-center gap-3">
          <Link
            href={`/dashboard/topics/${params.id}`}
            className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
          >
            Quay lại
          </Link>
          <Link
            href={`/dashboard/topics/${params.id}/quiz`}
            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            Làm bài kiểm tra
          </Link>
        </div>
      </div>
    );
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
        <span className="text-sm text-gray-500">
          {currentIndex + 1} / {words.length}
        </span>
      </div>

      <Flashcard word={words[currentIndex]} onRate={handleRate} />
    </div>
  );
}
