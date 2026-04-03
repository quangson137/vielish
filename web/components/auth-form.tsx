"use client";

import { useState, FormEvent } from "react";

interface AuthFormProps {
  mode: "login" | "register";
  onSubmit: (data: {
    email: string;
    password: string;
    displayName?: string;
  }) => Promise<void>;
}

export default function AuthForm({ mode, onSubmit }: AuthFormProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await onSubmit({
        email,
        password,
        ...(mode === "register" ? { displayName } : {}),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4">
      {error && (
        <div className="p-3 bg-red-50 text-red-700 rounded-lg text-sm">
          {error}
        </div>
      )}

      {mode === "register" && (
        <div>
          <label
            htmlFor="displayName"
            className="block text-sm font-medium mb-1"
          >
            Tên hiển thị
          </label>
          <input
            id="displayName"
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            required
            className="w-full px-3 py-2 border border-warm-border rounded-lg focus:outline-none focus:ring-2 focus:ring-warm-accent bg-warm-surface"
          />
        </div>
      )}

      <div>
        <label htmlFor="email" className="block text-sm font-medium mb-1">
          Email (Địa chỉ email)
        </label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="w-full px-3 py-2 border border-warm-border rounded-lg focus:outline-none focus:ring-2 focus:ring-warm-accent bg-warm-surface"
        />
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium mb-1">
          Mật khẩu
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
          className="w-full px-3 py-2 border border-warm-border rounded-lg focus:outline-none focus:ring-2 focus:ring-warm-accent bg-warm-surface"
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full py-3 bg-warm-accent text-white rounded-lg hover:bg-warm-accent-hover disabled:opacity-50"
      >
        {loading
          ? "Đang tải..."
          : mode === "login"
            ? "Đăng nhập"
            : "Tạo tài khoản"}
      </button>
    </form>
  );
}
