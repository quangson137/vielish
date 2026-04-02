export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-4xl font-bold mb-4">Vielish</h1>
      <p className="text-lg text-gray-600 mb-8">
        Learn English the Vietnamese way
      </p>
      <div className="flex gap-4">
        <a
          href="/login"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Đăng nhập
        </a>
        <a
          href="/register"
          className="px-6 py-3 border border-blue-600 text-blue-600 rounded-lg hover:bg-blue-50"
        >
          Đăng ký
        </a>
      </div>
    </main>
  );
}
