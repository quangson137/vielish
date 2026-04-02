"use client";

import { useRouter } from "next/navigation";
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
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-3xl font-bold mb-8">Tạo tài khoản</h1>
      <AuthForm mode="register" onSubmit={handleRegister} />
      <p className="mt-4 text-sm text-gray-600">
        Đã có tài khoản?{" "}
        <a href="/login" className="text-blue-600 hover:underline">
          Đăng nhập
        </a>
      </p>
    </main>
  );
}
