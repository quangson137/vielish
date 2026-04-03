import Link from "next/link";

export default function DashboardPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Bảng điều khiển</h2>
      <p className="text-gray-600 mb-6">
        Chào mừng bạn đến với Vielish! Chọn một hoạt động để bắt đầu.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Link
          href="/dashboard/topics"
          className="block p-6 border rounded-lg hover:shadow-md transition-shadow"
        >
          <h3 className="text-lg font-semibold mb-1">Chủ đề từ vựng</h3>
          <p className="text-gray-500 text-sm">
            Học từ mới theo chủ đề với flashcard và SRS.
          </p>
        </Link>
        <Link
          href="/dashboard/review"
          className="block p-6 border rounded-lg hover:shadow-md transition-shadow"
        >
          <h3 className="text-lg font-semibold mb-1">Ôn tập hôm nay</h3>
          <p className="text-gray-500 text-sm">
            Ôn lại các từ đã học theo lịch SRS.
          </p>
        </Link>
      </div>
    </div>
  );
}
