const tiles = [
  { key: "net_total", label: "Net Total", color: "text-blue-400" },
  { key: "processed_events", label: "Processed", color: "text-green-400" },
  { key: "pending_ack", label: "Pending ACK", color: "text-yellow-400" },
  { key: "unresolved", label: "Unresolved", color: "text-orange-400" },
  { key: "duplicates", label: "Duplicates", color: "text-purple-400" },
  { key: "conflicts", label: "Conflicts", color: "text-red-400" },
];

export default function StatTiles({ summary, loading }) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      {tiles.map((t) => (
        <div key={t.key} className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-gray-400 text-xs uppercase tracking-wide">{t.label}</p>
          <p className={`text-3xl font-bold mt-1 ${t.color}`}>
            {loading ? "—" : (summary?.[t.key] ?? 0)}
          </p>
        </div>
      ))}
    </div>
  );
}
