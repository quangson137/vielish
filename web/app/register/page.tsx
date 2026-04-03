"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import AuthForm from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function RegisterPage() {
  const router = useRouter();
  const { register } = useAuth();

  const handleRegister = async (data: {
    email: string;
    password: string;
    displayName?: string;
  }) => {
    await register(data.email, data.password, data.displayName || "");
    router.push("/dashboard");
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8 bg-warm-bg">
      <Link href="/" className="text-2xl font-bold text-warm-muted mb-8 hover:text-warm-accent transition-colors">
        Vielish
      </Link>
      <h1 className="text-3xl font-bold mb-8 text-warm-text">Tạo tài khoản</h1>
      <AuthForm mode="register" onSubmit={handleRegister} />
      <p className="mt-4 text-sm text-warm-muted">
        Đã có tài khoản?{" "}
        <Link href="/login" className="text-warm-accent hover:underline">
          Đăng nhập
        </Link>
      </p>
    </main>
  );
}
