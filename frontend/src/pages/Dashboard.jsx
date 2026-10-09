import { useState } from "react";
import { useStateData } from "../hooks/useStateData";
import Navbar from "../components/layout/Navbar";
import StatTiles from "../components/ui/StatTiles";
import EventForm from "../components/ui/EventForm";
import PendingTable from "../components/ui/PendingTable";
import ExceptionsTable from "../components/ui/ExceptionsTable";
import MqttPanel from "../components/ui/MqttPanel";
import { RefreshCw, Filter, X } from "lucide-react";

export default function Dashboard() {
  const [sourceFilter, setSourceFilter] = useState("");
  const [inputVal, setInputVal] = useState("");
  const { summary, pending, exceptions, mqtt, loading, refresh } = useStateData(sourceFilter);

  const applyFilter = () => setSourceFilter(inputVal.trim());
  const clearFilter = () => { setSourceFilter(""); setInputVal(""); };

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

        {/* Source Filter */}
        <div style={{
          display: "flex", alignItems: "center", gap: 8, marginBottom: 14,
          background: "var(--bg-card)", border: "1px solid var(--border)",
          borderRadius: 8, padding: "10px 14px",
        }}>
          <Filter size={13} color="var(--text-muted)" />
          <span style={{ fontSize: 11, fontWeight: 700, color: "var(--text-muted)", letterSpacing: "0.08em", textTransform: "uppercase", whiteSpace: "nowrap" }}>
            Source Filter
          </span>
          <input
            value={inputVal}
            onChange={(e) => setInputVal(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && applyFilter()}
            placeholder="e.g. LINE-01"
            style={{
              flex: 1, background: "var(--bg-input)", border: "1px solid var(--border)",
              borderRadius: 5, color: "var(--text-primary)", fontSize: 12,
              padding: "5px 10px", outline: "none", maxWidth: 200,
            }}
          />
          <button onClick={applyFilter} className="btn btn-primary" style={{ padding: "5px 14px", fontSize: 11 }}>Apply</button>
          {sourceFilter && (
            <button onClick={clearFilter} style={{
              display: "flex", alignItems: "center", gap: 4, background: "none",
              border: "1px solid var(--border)", borderRadius: 5, cursor: "pointer",
              color: "var(--text-secondary)", fontSize: 11, padding: "5px 10px",
            }}>
              <X size={11} /> Clear
            </button>
          )}
          {sourceFilter && (
            <span style={{ fontSize: 11, color: "var(--blue)", fontFamily: "monospace" }}>
              Viewing: {sourceFilter}
            </span>
          )}
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
