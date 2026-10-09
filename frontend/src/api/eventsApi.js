const BASE = "/api";

export async function submitEvents(events) {
  const payload = Array.isArray(events) ? events : [events];
  const res = await fetch(`${BASE}/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return res.json();
}
