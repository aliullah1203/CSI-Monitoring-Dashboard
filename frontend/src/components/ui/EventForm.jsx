import { useState } from "react";
import { submitEvents } from "../../api/eventsApi";
import { Send, ToggleLeft, ToggleRight, Loader2 } from "lucide-react";

const badgeMap = {
  ACCEPTED:          "badge-green",
  DUPLICATE:         "badge-purple",
  CONFLICT:          "badge-red",
  REJECTED:          "badge-red",
  PENDING_REFERENCE: "badge-yellow",
};

const defaultForm = {
  source_id: "LINE-01",
  event_id: "",
  type: "COUNT",
  quantity: "",
  target_event_id: "",
  event_time: new Date().toISOString().slice(0, 19) + "Z",
};

export default function EventForm({ onSuccess }) {
  const [form, setForm] = useState(defaultForm);
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);

  const set = (k, v) => setForm((p) => ({ ...p, [k]: v }));
  const isVoid = form.type === "VOID";

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setResult(null);
    try {
      const payload = {
        source_id: form.source_id,
        event_id: form.event_id,
        type: form.type,
        event_time: form.event_time,
      };
      if (!isVoid) payload.quantity = parseInt(form.quantity, 10);
      else payload.target_event_id = form.target_event_id;

      const res = await submitEvents(payload);
      setResult(res);
      if (res.results?.[0]?.status === "ACCEPTED") {
        setForm((p) => ({ ...p, event_id: "", quantity: "", target_event_id: "" }));
        onSuccess?.();
      }
    } catch (err) {
      setResult({ error: err.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <div style={{ marginBottom: 16 }}>
        <div style={{ fontWeight: 700, fontSize: 13, color: "var(--text-primary)" }}>Submit Event</div>
        <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 2 }}>Factory floor manual entry</div>
      </div>

      <form onSubmit={handleSubmit}>
        {/* Type toggle */}
        <div style={{ display: "flex", gap: 6, marginBottom: 14 }}>
          {["COUNT", "VOID"].map((t) => (
            <button key={t} type="button"
              onClick={() => set("type", t)}
              style={{
                flex: 1, padding: "8px 0", borderRadius: 7, border: "none", cursor: "pointer",
                fontWeight: 700, fontSize: 12, letterSpacing: "0.04em",
                background: form.type === t
                  ? (t === "COUNT" ? "rgba(16,185,129,.15)" : "rgba(245,158,11,.15)")
                  : "var(--bg-input)",
                color: form.type === t
                  ? (t === "COUNT" ? "#34d399" : "#fbbf24")
                  : "var(--text-muted)",
                border: form.type === t
                  ? `1px solid ${t === "COUNT" ? "rgba(16,185,129,.3)" : "rgba(245,158,11,.3)"}`
                  : "1px solid var(--border)",
                transition: "all .15s",
              }}>
              {t}
            </button>
          ))}
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10, marginBottom: 10 }}>
          <div>
            <label style={{ fontSize: 11, color: "var(--text-muted)", fontWeight: 600, letterSpacing: "0.06em" }}>
              SOURCE ID
            </label>
            <input className="input" style={{ marginTop: 4 }}
              value={form.source_id} onChange={(e) => set("source_id", e.target.value)} />
          </div>
          <div>
            <label style={{ fontSize: 11, color: "var(--text-muted)", fontWeight: 600, letterSpacing: "0.06em" }}>
              EVENT ID
            </label>
            <input className="input" style={{ marginTop: 4 }}
              value={form.event_id} onChange={(e) => set("event_id", e.target.value)}
              required placeholder="EV-101" />
          </div>
        </div>

        <div style={{ marginBottom: 10 }}>
          <label style={{ fontSize: 11, color: "var(--text-muted)", fontWeight: 600, letterSpacing: "0.06em" }}>
            {isVoid ? "TARGET EVENT ID" : "QUANTITY"}
          </label>
          {isVoid ? (
            <input className="input" style={{ marginTop: 4 }}
              value={form.target_event_id} onChange={(e) => set("target_event_id", e.target.value)}
              required placeholder="EV-101" />
          ) : (
            <input className="input" style={{ marginTop: 4 }}
              type="number" min="1" value={form.quantity}
              onChange={(e) => set("quantity", e.target.value)}
              required placeholder="5" />
          )}
        </div>

        <div style={{ marginBottom: 14 }}>
          <label style={{ fontSize: 11, color: "var(--text-muted)", fontWeight: 600, letterSpacing: "0.06em" }}>
            EVENT TIME
          </label>
          <input className="input" style={{ marginTop: 4 }}
            value={form.event_time} onChange={(e) => set("event_time", e.target.value)} />
        </div>

        <button type="submit" className="btn btn-primary" disabled={loading} style={{ width: "100%", justifyContent: "center" }}>
          {loading
            ? <Loader2 size={14} style={{ animation: "spin 1s linear infinite" }} />
            : <Send size={14} />}
          {loading ? "Submitting..." : `Submit ${form.type} Event`}
        </button>
      </form>

      {result && (
        <div style={{ marginTop: 12, background: "var(--bg-base)", borderRadius: 8, padding: "10px 12px" }}>
          {result.error ? (
            <span style={{ color: "#f87171", fontSize: 12 }}>{result.error}</span>
          ) : (
            result.results?.map((r, i) => (
              <div key={i} style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 12 }}>
                <span style={{ fontFamily: "monospace", color: "var(--text-muted)" }}>{r.event_id}</span>
                <span className={`badge ${badgeMap[r.status] || "badge-gray"}`}>{r.status}</span>
                {r.message && <span style={{ color: "var(--text-muted)", fontSize: 11 }}>— {r.message}</span>}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
