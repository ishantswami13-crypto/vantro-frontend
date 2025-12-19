export function getActiveBusinessId(): number | null {
  if (typeof window === "undefined") return null;
  const v = localStorage.getItem("active_business_id");
  if (!v) return null;
  const n = Number(v);
  return Number.isFinite(n) ? n : null;
}

export function setActiveBusinessId(id: number) {
  if (typeof window === "undefined") return;
  localStorage.setItem("active_business_id", String(id));
}
