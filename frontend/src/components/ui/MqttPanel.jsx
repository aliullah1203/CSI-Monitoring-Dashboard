import { Radio, AlertCircle, Clock, Zap, Hash } from "lucide-react";

function Field({ icon: Icon, label, value, valueStyle }) {
  return (
    <div style={{
      background: "var(--bg-base)", borderRadius: 8, padding: "10px 14px",
      display: "flex", flexDirection: "column", gap: 5,
    }}>
      <div style={{ display: "flex", alignItems: "center", gap: 5, color: "var(--text-muted)", fontSize: 10, fontWeight: 700, letterSpacing: "0.1em", textTransform: "uppercase" }}>
        {Icon && <Icon size={10} />} {label}
      </div>
      <div style={{ fontSize: 13, fontWeight: 600, ...valueStyle }}>
        {value || "—"}
      </div>
    </div>
  );
}

export default function MqttPanel({ mqtt }) {
  if (!mqtt) return null;
  const connected = mqtt.connected;
  const hasChallenge = mqtt.last_seen_at && mqtt.last_seen_at !== "0001-01-01T00:00:00Z";

  return (
    <div className="card">
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 14 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <div style={{
            width: 32, height: 32, borderRadius: 8,
            background: connected ? "rgba(16,185,129,.1)" : "rgba(239,68,68,.1)",
            border: `1px solid ${connected ? "rgba(16,185,129,.2)" : "rgba(239,68,68,.2)"}`,
            display: "flex", alignItems: "center", justifyContent: "center",
          }}>
            <Radio size={15} color={connected ? "#34d399" : "#f87171"} />
          </div>
          <div>
            <div style={{ fontWeight: 700, fontSize: 13, color: "var(--text-primary)" }}>MQTT Device</div>
            <div style={{ fontSize: 11, color: "var(--text-muted)" }}>Support / Engineering view</div>
          </div>
        </div>
        <span className={`badge ${connected ? "badge-green" : "badge-red"}`}>
          <div className="pulse-dot" style={{ background: connected ? "#10b981" : "#ef4444" }} />
          {connected ? "ONLINE" : "OFFLINE"}
        </span>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))", gap: 8 }}>
        <Field icon={Hash} label="Candidate ID"
          value={<span style={{ fontFamily: "monospace", color: "#60a5fa" }}>{mqtt.candidate_id}</span>}
          valueStyle={{}} />

        <Field icon={Zap} label="Last Challenge"
          value={mqtt.last_challenge_id
            ? <span style={{ fontFamily: "monospace", color: "#a78bfa" }}>{mqtt.last_challenge_id}</span>
            : <span style={{ color: "var(--text-muted)", fontSize: 12 }}>Waiting...</span>}
          valueStyle={{}} />

        <Field icon={Radio} label="Last Response"
          value={mqtt.last_status
            ? <span style={{ color: mqtt.last_status.includes("FAILED") ? "#f87171" : "#34d399" }}>{mqtt.last_status}</span>
            : <span style={{ color: "var(--text-muted)", fontSize: 12 }}>None yet</span>}
          valueStyle={{}} />

        <Field icon={Clock} label="Last Seen"
          value={hasChallenge
            ? <span style={{ color: "var(--text-secondary)", fontSize: 12 }}>{new Date(mqtt.last_seen_at).toLocaleString()}</span>
            : <span style={{ color: "var(--text-muted)", fontSize: 12 }}>No challenge yet</span>}
          valueStyle={{}} />
      </div>

      {mqtt.last_error && (
        <div style={{
          marginTop: 10, background: "rgba(239,68,68,.06)", border: "1px solid rgba(239,68,68,.15)",
          borderRadius: 7, padding: "8px 12px",
          display: "flex", alignItems: "flex-start", gap: 8,
        }}>
          <AlertCircle size={13} color="#f87171" style={{ marginTop: 1, flexShrink: 0 }} />
          <span style={{ color: "#f87171", fontSize: 12 }}>{mqtt.last_error}</span>
        </div>
      )}

      {!mqtt.last_challenge_id && connected && (
        <div style={{
          marginTop: 10, background: "rgba(59,130,246,.05)", border: "1px solid rgba(59,130,246,.12)",
          borderRadius: 7, padding: "8px 12px", fontSize: 12, color: "#60a5fa",
          textAlign: "center",
        }}>
          Subscribed to <span style={{ fontFamily: "monospace" }}>fse-01/{mqtt.candidate_id}/challenge</span> — waiting for examiner
        </div>
      )}
    </div>
  );
}
