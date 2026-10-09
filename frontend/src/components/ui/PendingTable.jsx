import { useState } from "react";
import { ackEvents } from "../../api/ackApi";
import { CheckSquare, Square, Loader2 } from "lucide-react";

export default function PendingTable({ events, onRefresh }) {
  const [selected, setSelected] = useState([]);
  const [loading, setLoading] = useState(false);
  const [ackResult, setAckResult] = useState(null);

  const toggle = (id) =>
    setSelected((p) => p.includes(id) ? p.filter((x) => x !== id) : [...p, id]);
  const toggleAll = () =>
    setSelected(selected.length === events.length ? [] : events.map((e) => e.event_id));

  const handleAck = async () => {
    if (!selected.length) return;
    setLoading(true);
    setAckResult(null);
    try {
      const res = await ackEvents(selected);
      setAckResult(res);
      setSelected([]);
      onRefresh?.();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card" style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        <div>
          <div style={{ fontWeight: 700, fontSize: 13, color: "var(--text-primary)" }}>Pending ACK</div>
          <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 2 }}>
            {events.length} event{events.length !== 1 ? "s" : ""} awaiting review
          </div>
        </div>
        <button className="btn btn-success" onClick={handleAck} disabled={!selected.length || loading}>
          {loading ? <Loader2 size={13} style={{ animation: "spin 1s linear infinite" }} /> : <CheckSquare size={13} />}
          Acknowledge {selected.length > 0 ? `(${selected.length})` : ""}
        </button>
      </div>

      {events.length === 0 ? (
        <div style={{ textAlign: "center", padding: "24px 0", color: "var(--text-muted)", fontSize: 13 }}>
          ✓ No pending events
        </div>
      ) : (
        <div style={{ overflowX: "auto" }}>
          <table className="data-table">
            <thead>
              <tr>
                <th style={{ width: 32 }}>
                  <div onClick={toggleAll} style={{ cursor: "pointer", color: "var(--text-muted)", display: "flex" }}>
                    {selected.length === events.length && events.length > 0
                      ? <CheckSquare size={14} color="#3b82f6" />
                      : <Square size={14} />}
                  </div>
                </th>
                <th>Event ID</th>
                <th>Source</th>
                <th>Type</th>
                <th>Qty</th>
                <th>Event Time</th>
              </tr>
            </thead>
            <tbody>
              {events.map((ev) => {
                const sel = selected.includes(ev.event_id);
                return (
                  <tr key={ev.event_id} onClick={() => toggle(ev.event_id)} style={{ cursor: "pointer" }}>
                    <td>
                      {sel
                        ? <CheckSquare size={14} color="#3b82f6" />
                        : <Square size={14} color="var(--text-muted)" />}
                    </td>
                    <td style={{ fontFamily: "monospace", color: "#60a5fa", fontSize: 12 }}>{ev.event_id}</td>
                    <td style={{ color: "var(--text-secondary)" }}>{ev.source_id}</td>
                    <td>
                      <span className={`badge ${ev.type === "COUNT" ? "badge-green" : "badge-yellow"}`}>
                        {ev.type}
                      </span>
                    </td>
                    <td style={{ color: "var(--text-primary)", fontVariantNumeric: "tabular-nums" }}>
                      {ev.quantity ?? "—"}
                    </td>
                    <td style={{ color: "var(--text-muted)", fontSize: 11 }}>
                      {new Date(ev.event_time).toLocaleString()}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {ackResult?.results && (
        <div style={{ background: "var(--bg-base)", borderRadius: 6, padding: "8px 12px", display: "flex", flexDirection: "column", gap: 3 }}>
          {ackResult.results.map((r, i) => (
            <div key={i} style={{ display: "flex", gap: 8, fontSize: 12 }}>
              <span style={{ fontFamily: "monospace", color: "var(--text-muted)" }}>{r.event_id}</span>
              <span className={`badge ${r.status === "ACKED" ? "badge-green" : "badge-yellow"}`}>{r.status}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
