"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";

export default function SignupPage() {
  const router = useRouter();
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSignup(e: FormEvent) {
    e.preventDefault();
    setError("");

    try {
      const data = await apiFetch<{ token: string }>("/api/auth/signup", {
        method: "POST",
        body: JSON.stringify({ full_name: fullName, email, password }),
      });

      localStorage.setItem("token", data.token);
      router.push("/dashboard");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Signup failed");
    }
  }

  return (
    <div className="flex items-center justify-center h-screen bg-black text-white">
      <form
        onSubmit={handleSignup}
        className="bg-neutral-900 p-8 rounded-lg w-96 space-y-4"
      >
        <h1 className="text-2xl font-bold text-center">VANTRO Sign up</h1>

        <input
          className="w-full p-2 rounded bg-neutral-800"
          placeholder="Full name"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
        />

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
          Create account
        </button>

        <p className="text-sm text-center opacity-80">
          Already have an account?{" "}
          <a className="underline" href="/login">
            Login
          </a>
        </p>
      </form>
    </div>
  );
}
