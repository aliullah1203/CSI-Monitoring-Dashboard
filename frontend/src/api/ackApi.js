const BASE = "/api";

export async function ackEvents(eventIDs) {
  const res = await fetch(`${BASE}/ack`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ event_ids: eventIDs }),
  });
  return res.json();
}
