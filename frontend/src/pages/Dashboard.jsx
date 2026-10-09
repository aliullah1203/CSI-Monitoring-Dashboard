import { useStateData } from "../hooks/useStateData";
import Navbar from "../components/layout/Navbar";
import StatTiles from "../components/ui/StatTiles";
import EventForm from "../components/ui/EventForm";
import PendingTable from "../components/ui/PendingTable";
import ExceptionsTable from "../components/ui/ExceptionsTable";
import MqttPanel from "../components/ui/MqttPanel";
import { RefreshCw } from "lucide-react";

export default function Dashboard() {
  const { summary, pending, exceptions, mqtt, loading, refresh } = useStateData();

  return (
    <div style={{ minHeight: "100vh", background: "var(--bg-base)" }}>
      <Navbar mqtt={mqtt} />

      <main style={{ maxWidth: 1280, margin: "0 auto", padding: "24px 20px 48px" }}>

        {/* ── Production Supervisor ── */}
        <div className="section-label" style={{ marginBottom: 14 }}>
          Production Supervisor
          <button onClick={refresh} style={{
            background: "none", border: "none", cursor: "pointer",
            color: "var(--text-muted)", display: "flex", alignItems: "center", gap: 4,
            fontSize: 10, fontWeight: 600, letterSpacing: "0.06em",
          }}>
            <RefreshCw size={10} /> REFRESH
          </button>
        </div>

        <StatTiles summary={summary} loading={loading} />

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14, marginTop: 14 }}>
          <PendingTable events={pending} onRefresh={refresh} />
          <ExceptionsTable events={exceptions} />
        </div>

        {/* ── Factory Floor Operator ── */}
        <div className="section-label" style={{ margin: "28px 0 14px" }}>
          Factory Floor Operator
        </div>
        <EventForm onSuccess={refresh} />

        {/* ── Support / Engineering ── */}
        <div className="section-label" style={{ margin: "28px 0 14px" }}>
          Support / Engineering
        </div>
        <MqttPanel mqtt={mqtt} />

      </main>
    </div>
  );
}
