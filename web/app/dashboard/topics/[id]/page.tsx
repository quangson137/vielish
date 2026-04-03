"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { fetchTopicWords, Word } from "@/lib/vocab-api";

export default function TopicDetailPage() {
  const params = useParams<{ id: string }>();
  const [words, setWords] = useState<Word[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchTopicWords(params.id)
      .then(setWords)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [params.id]);

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <Link
          href="/dashboard/topics"
          className="text-blue-600 hover:underline text-sm"
        >
          ← Chủ đề
        </Link>
      </div>

      <div className="flex gap-3 mb-6">
        <Link
          href={`/dashboard/topics/${params.id}/learn`}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Học từ mới
        </Link>
        <Link
          href={`/dashboard/topics/${params.id}/quiz`}
          className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
        >
          Làm bài kiểm tra
        </Link>
      </div>

      <h3 className="text-lg font-semibold mb-3">
        Danh sách từ ({words.length})
      </h3>
      <div className="space-y-2">
        {words.map((word) => (
          <div key={word.id} className="p-3 border rounded flex justify-between">
            <div>
              <span className="font-medium">{word.word}</span>
              {word.ipa_phonetic && (
                <span className="text-gray-500 text-sm ml-2">
                  {word.ipa_phonetic}
                </span>
              )}
            </div>
            <span className="text-gray-600">{word.vi_meaning}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
