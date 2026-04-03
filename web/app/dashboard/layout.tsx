"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";

const navLinks = [
  { href: "/dashboard", label: "Trang chủ" },
  { href: "/dashboard/topics", label: "Chủ đề" },
  { href: "/dashboard/review", label: "Ôn tập" },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-warm-bg">
        <p className="text-warm-muted">Đang tải...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-warm-bg">
      <nav className="border-b border-warm-border px-6 py-4 flex justify-between items-center bg-warm-surface">
        <div className="flex items-center gap-6">
          <h1 className="text-xl font-bold text-warm-muted">Vielish</h1>
          {navLinks.map((link) => {
            const isActive =
              link.href === "/dashboard"
                ? pathname === "/dashboard"
                : pathname.startsWith(link.href);
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`text-sm transition-colors ${
                  isActive
                    ? "text-warm-accent font-semibold border-b-2 border-warm-accent pb-1"
                    : "text-warm-muted hover:text-warm-text"
                }`}
              >
                {link.label}
              </Link>
            );
          })}
        </div>
        <button
          onClick={() => {
            logout();
            router.push("/login");
          }}
          className="text-sm text-warm-muted hover:text-warm-text"
        >
          Đăng xuất
        </button>
      </nav>
      <main className="p-6">{children}</main>
    </div>
  );
}
