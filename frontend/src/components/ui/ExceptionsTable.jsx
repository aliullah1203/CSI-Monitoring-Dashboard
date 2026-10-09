const statusStyle = {
  DUPLICATE: "bg-purple-900 text-purple-300",
  CONFLICT: "bg-red-900 text-red-300",
  REJECTED: "bg-gray-700 text-gray-300",
};

export default function ExceptionsTable({ events }) {
  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
      <h2 className="text-white font-semibold text-sm uppercase tracking-wide mb-3">
        Exceptions
      </h2>
      {events.length === 0 ? (
        <p className="text-gray-500 text-sm">No exceptions.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-gray-400 text-xs border-b border-gray-700">
                <th className="pb-2 pr-3 text-left">Event ID</th>
                <th className="pb-2 pr-3 text-left">Source</th>
                <th className="pb-2 pr-3 text-left">Status</th>
                <th className="pb-2 text-left">Reason</th>
              </tr>
            </thead>
            <tbody>
              {events.map((ev, i) => (
                <tr key={i} className="border-b border-gray-750 hover:bg-gray-750">
                  <td className="py-1.5 pr-3 font-mono text-blue-300">{ev.event_id}</td>
                  <td className="py-1.5 pr-3 text-gray-300">{ev.source_id}</td>
                  <td className="py-1.5 pr-3">
                    <span className={`text-xs font-medium px-1.5 py-0.5 rounded ${statusStyle[ev.status] || "bg-gray-700 text-gray-300"}`}>
                      {ev.status}
                    </span>
                  </td>
                  <td className="py-1.5 text-gray-400 text-xs">{ev.reason || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
