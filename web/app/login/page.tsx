"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import AuthForm from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();

  const handleLogin = async (data: {
    email: string;
    password: string;
  }) => {
    await login(data.email, data.password);
    router.push("/dashboard");
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8 bg-warm-bg">
      <Link href="/" className="text-2xl font-bold text-warm-muted mb-8 hover:text-warm-accent transition-colors">
        Vielish
      </Link>
      <h1 className="text-3xl font-bold mb-8 text-warm-text">Đăng nhập</h1>
      <AuthForm mode="login" onSubmit={handleLogin} />
      <p className="mt-4 text-sm text-warm-muted">
        Chưa có tài khoản?{" "}
        <Link href="/register" className="text-warm-accent hover:underline">
          Đăng ký
        </Link>
      </p>
    </main>
  );
}
