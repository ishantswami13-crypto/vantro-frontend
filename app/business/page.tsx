"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { setActiveBusinessId } from "@/lib/business";
import { getToken } from "@/lib/auth";

type Biz = { id: number; name: string; currency: string; created_at: string };

export default function BusinessPage() {
  const router = useRouter();
  const [list, setList] = useState<Biz[]>([]);
  const [name, setName] = useState("");
  const [err, setErr] = useState("");

  async function load() {
    const data = await apiFetch<Biz[]>("/api/businesses");
    setList(data);
  }

  useEffect(() => {
    const token = getToken();
    if (!token) {
      router.replace("/login");
      return;
    }
    load().catch((e: any) => setErr(e?.message || "failed"));
  }, [router]);

  async function create() {
    setErr("");
    try {
      const res = await apiFetch<{ id: number }>("/api/businesses", {
        method: "POST",
        body: JSON.stringify({ name, currency: "INR" }),
      });
      setName("");
      await load();
      setActiveBusinessId(res.id);
      router.push("/dashboard");
    } catch (e: any) {
      setErr(e?.message || "failed to create");
    }
  }

  return (
    <div className="min-h-screen bg-black text-white flex items-center justify-center p-6">
      <div className="w-full max-w-lg space-y-4 bg-neutral-900 p-6 rounded-2xl">
        <h1 className="text-2xl font-bold">Select Business</h1>

        {err ? <div className="text-red-400 text-sm">{err}</div> : null}

        <div className="space-y-2">
          {list.map((b) => (
            <button
              key={b.id}
              className="w-full text-left p-3 rounded bg-neutral-800 hover:bg-neutral-700"
              onClick={() => {
                setActiveBusinessId(b.id);
                router.push("/dashboard");
              }}
            >
              <div className="font-semibold">{b.name}</div>
              <div className="text-xs text-neutral-400">{b.currency}</div>
            </button>
          ))}
        </div>

        <div className="pt-4 border-t border-neutral-800 space-y-2">
          <input
            className="w-full p-2 rounded bg-neutral-800"
            placeholder="New business name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <button onClick={create} className="w-full bg-white text-black p-2 rounded">
            Create Business
          </button>
        </div>
      </div>
    </div>
  );
}
