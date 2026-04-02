"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-gray-500">Đang tải...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen">
      <nav className="border-b px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">Vielish</h1>
        <button
          onClick={() => {
            logout();
            router.push("/login");
          }}
          className="text-sm text-gray-600 hover:text-gray-900"
        >
          Đăng xuất
        </button>
      </nav>
      <main className="p-6">{children}</main>
    </div>
  );
}
