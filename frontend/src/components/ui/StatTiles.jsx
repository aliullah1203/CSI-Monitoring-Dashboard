import { TrendingUp, CheckCircle, Clock, AlertTriangle, Copy, Zap, XCircle } from "lucide-react";

const tiles = [
  {
    key: "net_total", label: "Net Total", icon: TrendingUp,
    color: "#93c5fd", bg: "#1e3a5f", border: "#2d5a96", iconBg: "#254e82",
  },
  {
    key: "processed_events", label: "Processed", icon: CheckCircle,
    color: "#6ee7b7", bg: "#14532d", border: "#166534", iconBg: "#15803d",
  },
  {
    key: "pending_ack", label: "Pending ACK", icon: Clock,
    color: "#fcd34d", bg: "#451a03", border: "#78350f", iconBg: "#92400e",
  },
  {
    key: "unresolved", label: "Unresolved", icon: AlertTriangle,
    color: "#fdba74", bg: "#431407", border: "#7c2d12", iconBg: "#9a3412",
  },
  {
    key: "duplicates", label: "Duplicates", icon: Copy,
    color: "#d8b4fe", bg: "#3b0764", border: "#6b21a8", iconBg: "#7e22ce",
  },
  {
    key: "conflicts", label: "Conflicts", icon: Zap,
    color: "#fca5a5", bg: "#450a0a", border: "#991b1b", iconBg: "#b91c1c",
  },
  {
    key: "rejected_submissions", label: "Rejected", icon: XCircle,
    color: "#fb923c", bg: "#431407", border: "#9a3412", iconBg: "#7c2d12",
  },
];

export default function StatTiles({ summary, loading }) {
  return (
    <div style={{
      display: "grid",
      gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))",
      gap: 12,
    }}>
      {tiles.map((t) => {
        const Icon = t.icon;
        const val = loading ? null : (summary?.[t.key] ?? 0);
        const isAlert = !loading && val > 0 && (t.key === "conflicts" || t.key === "unresolved");

        return (
          <div key={t.key} style={{
            background: t.bg,
            border: `1px solid ${t.border}`,
            borderRadius: 10,
            padding: "18px",
            display: "flex",
            flexDirection: "column",
            gap: 14,
            boxShadow: isAlert ? `0 0 16px rgba(248,81,73,.2)` : `0 1px 8px rgba(0,0,0,.4)`,
            transition: "box-shadow .3s",
          }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
              <span style={{
                fontSize: 10, fontWeight: 700, letterSpacing: "0.1em",
                textTransform: "uppercase", color: t.color, opacity: 0.8,
              }}>
                {t.label}
              </span>
              <div style={{
                width: 30, height: 30, borderRadius: 8,
                background: t.iconBg,
                display: "flex", alignItems: "center", justifyContent: "center",
              }}>
                <Icon size={15} color={t.color} />
              </div>
            </div>
            <div style={{
              fontSize: 36,
              fontWeight: 800,
              color: loading ? "#30363d" : t.color,
              lineHeight: 1,
              fontVariantNumeric: "tabular-nums",
            }}>
              {loading ? "—" : val.toLocaleString()}
            </div>
          </div>
        );
      })}
    </div>
  );
}
