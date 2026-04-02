"use client";

import { useRouter } from "next/navigation";
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
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-3xl font-bold mb-8">Đăng nhập Vielish</h1>
      <AuthForm mode="login" onSubmit={handleLogin} />
      <p className="mt-4 text-sm text-gray-600">
        Chưa có tài khoản?{" "}
        <a href="/register" className="text-blue-600 hover:underline">
          Đăng ký
        </a>
      </p>
    </main>
  );
}
