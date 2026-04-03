"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { fetchTopics, Topic } from "@/lib/vocab-api";

const LEVELS = ["beginner", "intermediate", "advanced"];
const LEVEL_LABELS: Record<string, string> = {
  beginner: "Cơ bản",
  intermediate: "Trung cấp",
  advanced: "Nâng cao",
};

export default function TopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [level, setLevel] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchTopics(level || undefined)
      .then(setTopics)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [level]);

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Chủ đề từ vựng</h2>

      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setLevel("")}
          className={`px-3 py-1 rounded text-sm ${
            level === "" ? "bg-blue-600 text-white" : "bg-gray-200"
          }`}
        >
          Tất cả
        </button>
        {LEVELS.map((l) => (
          <button
            key={l}
            onClick={() => setLevel(l)}
            className={`px-3 py-1 rounded text-sm ${
              level === l ? "bg-blue-600 text-white" : "bg-gray-200"
            }`}
          >
            {LEVEL_LABELS[l]}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-gray-500">Đang tải...</p>
      ) : topics.length === 0 ? (
        <p className="text-gray-500">Chưa có chủ đề nào.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {topics.map((topic) => (
            <Link
              key={topic.id}
              href={`/dashboard/topics/${topic.id}`}
              className="block p-4 border rounded-lg hover:shadow-md transition-shadow"
            >
              <h3 className="font-semibold text-lg">{topic.name}</h3>
              <p className="text-gray-600 text-sm">{topic.name_vi}</p>
              {topic.description && (
                <p className="text-gray-500 text-sm mt-1">
                  {topic.description}
                </p>
              )}
              <span className="inline-block mt-2 px-2 py-0.5 bg-gray-100 text-xs rounded">
                {LEVEL_LABELS[topic.level] || topic.level}
              </span>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
