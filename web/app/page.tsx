import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen bg-warm-bg">
      <nav className="flex items-center justify-between px-6 py-4">
        <span className="text-xl font-bold text-warm-muted">Vielish</span>
        <div className="flex gap-3">
          <Link
            href="/login"
            className="px-4 py-2 text-sm border border-warm-accent text-warm-accent rounded-lg hover:bg-warm-accent hover:text-white transition-colors"
          >
            Đăng nhập
          </Link>
          <Link
            href="/register"
            className="px-4 py-2 text-sm bg-warm-accent text-white rounded-lg hover:bg-warm-accent-hover transition-colors"
          >
            Đăng ký
          </Link>
        </div>
      </nav>

      <main className="flex flex-col items-center justify-center px-8 py-32">
        <h1 className="text-4xl font-bold text-warm-text mb-4 text-center">
          Học từ vựng tiếng Anh — không bao giờ quên
        </h1>
        <p className="text-lg text-warm-muted mb-8 text-center">
          SRS thông minh · Giao diện Việt · Miễn phí
        </p>
        <div className="flex gap-4">
          <Link
            href="/register"
            className="px-6 py-3 bg-warm-accent text-white rounded-lg hover:bg-warm-accent-hover font-semibold transition-colors"
          >
            Đăng ký miễn phí →
          </Link>
          <Link
            href="/login"
            className="px-6 py-3 border border-warm-accent text-warm-accent rounded-lg hover:bg-warm-accent hover:text-white transition-colors"
          >
            Đăng nhập
          </Link>
        </div>
        <p className="mt-12 text-warm-subtle text-sm border-t border-dashed border-warm-border pt-4">
          ⭐ Được dùng bởi hàng nghìn học viên Việt Nam
        </p>
      </main>
    </div>
  );
}
