import { Wifi, WifiOff, Activity } from "lucide-react";

export default function Navbar({ mqtt }) {
  const connected = mqtt?.connected;
  return (
    <nav style={{
      background: "#161b22",
      borderBottom: "1px solid #30363d",
      padding: "0 24px",
      height: 58,
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      position: "sticky",
      top: 0,
      zIndex: 100,
    }}>
      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
        <div style={{
          width: 34, height: 34, borderRadius: 8,
          background: "#1f6feb",
          display: "flex", alignItems: "center", justifyContent: "center",
          flexShrink: 0,
        }}>
          <Activity size={17} color="#fff" />
        </div>
        <div>
          <div style={{ fontWeight: 700, fontSize: 14, color: "#e6edf3", lineHeight: 1.3 }}>
            NorthBridge Garments
          </div>
          <div style={{ fontSize: 10, color: "#484f58", letterSpacing: "0.1em", textTransform: "uppercase" }}>
            Production Monitor
          </div>
        </div>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        {mqtt?.candidate_id && (
          <div style={{ fontSize: 11, color: "#8b949e", fontFamily: "monospace" }}>
            <span style={{ color: "#484f58" }}>ID: </span>
            <span style={{ color: "#58a6ff" }}>{mqtt.candidate_id}</span>
          </div>
        )}
        <div style={{
          display: "flex", alignItems: "center", gap: 6,
          padding: "5px 12px", borderRadius: 999,
          background: connected ? "#1a4429" : "#3d1a1a",
          border: `1px solid ${connected ? "#2d6a3f" : "#6e2929"}`,
          fontSize: 11, fontWeight: 700, letterSpacing: "0.06em",
          color: connected ? "#3fb950" : "#f85149",
        }}>
          <div className="pulse-dot" style={{ background: connected ? "#3fb950" : "#f85149" }} />
          {connected ? <Wifi size={11} /> : <WifiOff size={11} />}
          {connected ? "ONLINE" : "OFFLINE"}
        </div>
      </div>
    </nav>
  );
}
