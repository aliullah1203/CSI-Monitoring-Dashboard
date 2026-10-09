const badgeMap = {
  DUPLICATE: "badge-purple",
  CONFLICT:  "badge-red",
  REJECTED:  "badge-gray",
};

export default function ExceptionsTable({ events }) {
  return (
    <div className="card" style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      <div>
        <div style={{ fontWeight: 700, fontSize: 13, color: "var(--text-primary)" }}>Exceptions</div>
        <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 2 }}>
          {events.length} record{events.length !== 1 ? "s" : ""}
        </div>
      </div>

      {events.length === 0 ? (
        <div style={{ textAlign: "center", padding: "24px 0", color: "var(--text-muted)", fontSize: 13 }}>
          ✓ No exceptions
        </div>
      ) : (
        <div style={{ overflowX: "auto" }}>
          <table className="data-table">
            <thead>
              <tr>
                <th>Event ID</th>
                <th>Source</th>
                <th>Status</th>
                <th>Reason</th>
              </tr>
            </thead>
            <tbody>
              {events.map((ev, i) => (
                <tr key={i}>
                  <td style={{ fontFamily: "monospace", color: "#60a5fa", fontSize: 12 }}>{ev.event_id}</td>
                  <td style={{ color: "var(--text-secondary)" }}>{ev.source_id}</td>
                  <td><span className={`badge ${badgeMap[ev.status] || "badge-gray"}`}>{ev.status}</span></td>
                  <td style={{ color: "var(--text-muted)", fontSize: 12, maxWidth: 180 }}>{ev.reason || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
