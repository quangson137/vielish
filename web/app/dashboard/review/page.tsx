"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Flashcard from "@/components/flashcard";
import { fetchDueReviews, submitReview, Word } from "@/lib/vocab-api";

export default function ReviewPage() {
  const [words, setWords] = useState<Word[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [finished, setFinished] = useState(false);

  useEffect(() => {
    fetchDueReviews()
      .then(setWords)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

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

  if (words.length === 0 && !finished) {
    return (
      <div className="text-center py-12">
        <p className="text-xl font-semibold mb-2">Không có từ cần ôn tập</p>
        <p className="text-gray-500 mb-4">
          Hãy học thêm từ mới hoặc quay lại sau.
        </p>
        <Link
          href="/dashboard/topics"
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Xem chủ đề
        </Link>
      </div>
    );
  }

  if (finished) {
    return (
      <div className="text-center py-12">
        <p className="text-2xl font-bold mb-4">Ôn tập xong!</p>
        <p className="text-gray-600 mb-6">
          Bạn đã ôn tập {words.length} từ hôm nay.
        </p>
        <Link
          href="/dashboard"
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Về trang chủ
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold">Ôn tập từ vựng</h2>
        <span className="text-sm text-gray-500">
          {currentIndex + 1} / {words.length}
        </span>
      </div>

      <Flashcard word={words[currentIndex]} onRate={handleRate} />
    </div>
  );
}
