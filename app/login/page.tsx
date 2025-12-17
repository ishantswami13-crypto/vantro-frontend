"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { AuthAPI } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const data = await AuthAPI.login(email, password);

      // save JWT
      localStorage.setItem("token", data.token);

      router.push("/dashboard");
    } catch (err: any) {
      setError("Invalid email or password");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex items-center justify-center h-screen bg-black text-white">
      <form
        onSubmit={handleLogin}
        className="bg-neutral-900 p-8 rounded-xl w-80 space-y-4 shadow-lg"
      >
        <h1 className="text-2xl font-bold text-center tracking-wide">
          VANTRO
        </h1>

        <input
          className="w-full p-2 rounded bg-neutral-800 outline-none"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />

        <input
          className="w-full p-2 rounded bg-neutral-800 outline-none"
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />

        {error && <p className="text-red-400 text-sm">{error}</p>}

        <button
          disabled={loading}
          className="w-full bg-white text-black p-2 rounded font-medium disabled:opacity-60"
        >
          {loading ? "Logging in..." : "Login"}
        </button>
      </form>
    </div>
  );
}
