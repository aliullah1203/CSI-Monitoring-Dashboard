import { useState } from "react";
import { ackEvents } from "../../api/ackApi";
import { CheckSquare } from "lucide-react";

export default function PendingTable({ events, onRefresh }) {
  const [selected, setSelected] = useState([]);
  const [loading, setLoading] = useState(false);
  const [ackResult, setAckResult] = useState(null);

  const toggle = (id) =>
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );

  const toggleAll = () =>
    setSelected(selected.length === events.length ? [] : events.map((e) => e.event_id));

  const handleAck = async () => {
    if (!selected.length) return;
    setLoading(true);
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
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-white font-semibold text-sm uppercase tracking-wide">
          Pending ACK — Supervisor
        </h2>
        <button onClick={handleAck} disabled={!selected.length || loading}
          className="flex items-center gap-1.5 bg-green-700 hover:bg-green-600 disabled:opacity-40 text-white px-3 py-1.5 rounded text-xs font-medium">
          <CheckSquare size={13} /> Acknowledge ({selected.length})
        </button>
      </div>

      {events.length === 0 ? (
        <p className="text-gray-500 text-sm">No pending events.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-gray-400 text-xs border-b border-gray-700">
                <th className="pb-2 pr-3 text-left">
                  <input type="checkbox"
                    checked={selected.length === events.length && events.length > 0}
                    onChange={toggleAll} className="accent-green-500" />
                </th>
                <th className="pb-2 pr-3 text-left">Event ID</th>
                <th className="pb-2 pr-3 text-left">Source</th>
                <th className="pb-2 pr-3 text-left">Type</th>
                <th className="pb-2 pr-3 text-left">Qty</th>
                <th className="pb-2 text-left">Time</th>
              </tr>
            </thead>
            <tbody>
              {events.map((ev) => (
                <tr key={ev.event_id} className="border-b border-gray-750 hover:bg-gray-750">
                  <td className="py-1.5 pr-3">
                    <input type="checkbox" checked={selected.includes(ev.event_id)}
                      onChange={() => toggle(ev.event_id)} className="accent-green-500" />
                  </td>
                  <td className="py-1.5 pr-3 font-mono text-blue-300">{ev.event_id}</td>
                  <td className="py-1.5 pr-3 text-gray-300">{ev.source_id}</td>
                  <td className="py-1.5 pr-3">
                    <span className={`text-xs font-medium ${ev.type === "COUNT" ? "text-green-400" : "text-orange-400"}`}>
                      {ev.type}
                    </span>
                  </td>
                  <td className="py-1.5 pr-3 text-gray-300">{ev.quantity ?? "—"}</td>
                  <td className="py-1.5 text-gray-400 text-xs">
                    {new Date(ev.event_time).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {ackResult && (
        <div className="mt-3 bg-gray-900 rounded p-2 text-xs space-y-1">
          {ackResult.results?.map((r, i) => (
            <div key={i}>
              <span className="text-gray-400 font-mono">{r.event_id}: </span>
              <span className={r.status === "ACKED" ? "text-green-400" : "text-yellow-400"}>{r.status}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
