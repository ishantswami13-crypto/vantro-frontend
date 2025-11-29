"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleLogin(e: any) {
    e.preventDefault();

    const res = await fetch("https://YOUR-BACKEND-URL/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      setError("Invalid credentials");
      return;
    }

    const data = await res.json();

    // save token
    localStorage.setItem("token", data.token);

    router.push("/dashboard");
  }

  return (
    <div className="flex items-center justify-center h-screen bg-black text-white">
      <form
        onSubmit={handleLogin}
        className="bg-neutral-900 p-8 rounded-lg w-80 space-y-4"
      >
        <h1 className="text-2xl font-bold text-center">VANTRO Login</h1>

        <input
          className="w-full p-2 rounded bg-neutral-800"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />

        <input
          className="w-full p-2 rounded bg-neutral-800"
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />

        {error && <p className="text-red-400 text-sm">{error}</p>}

        <button className="w-full bg-white text-black p-2 rounded">
          Login
        </button>
      </form>
    </div>
  );
}
