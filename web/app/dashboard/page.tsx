"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { fetchStats, UserStats } from "@/lib/vocab-api";

export default function DashboardPage() {
  const { displayName } = useAuth();
  const [stats, setStats] = useState<UserStats | null>(null);

  useEffect(() => {
    fetchStats()
      .then(setStats)
      .catch(console.error);
  }, []);

  const greeting = displayName
    ? `Xin chào, ${displayName}! 👋`
    : "Xin chào! 👋";

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6 text-warm-text">{greeting}</h2>

      <div className="grid grid-cols-3 gap-4 mb-8">
        <div className="bg-warm-surface border border-warm-border rounded-lg p-4 text-center">
          <p className="text-3xl font-bold text-warm-accent">
            {stats?.streak ?? "–"}
          </p>
          <p className="text-sm text-warm-muted">ngày streak 🔥</p>
        </div>
        <div className="bg-warm-surface border border-warm-border rounded-lg p-4 text-center">
          <p className="text-3xl font-bold text-green-600">
            {stats?.total_learned ?? "–"}
          </p>
          <p className="text-sm text-warm-muted">từ đã học 📚</p>
        </div>
        <div className="bg-warm-surface border border-warm-border rounded-lg p-4 text-center">
          <p className="text-3xl font-bold text-blue-600">
            {stats?.due_today ?? "–"}
          </p>
          <p className="text-sm text-warm-muted">cần ôn hôm nay 🔔</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Link
          href="/dashboard/topics"
          className="block p-6 bg-warm-accent rounded-lg hover:bg-warm-accent-hover transition-colors text-white"
        >
          <p className="text-2xl mb-1">🃏</p>
          <h3 className="text-lg font-semibold mb-1">Chủ đề từ vựng</h3>
          <p className="text-sm opacity-80">
            Học từ mới theo chủ đề với flashcard và SRS.
          </p>
        </Link>
        <Link
          href="/dashboard/review"
          className="block p-6 bg-warm-surface border border-warm-border rounded-lg hover:shadow-md transition-shadow"
        >
          <p className="text-2xl mb-1">🔁</p>
          <h3 className="text-lg font-semibold text-warm-text mb-1">
            Ôn tập hôm nay
          </h3>
          <p className="text-sm text-warm-muted">
            {stats && stats.due_today > 0
              ? `${stats.due_today} từ đang chờ bạn`
              : "Ôn lại các từ đã học theo lịch SRS."}
          </p>
        </Link>
      </div>
    </div>
  );
}
