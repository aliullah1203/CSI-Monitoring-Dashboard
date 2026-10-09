import { useState } from "react";
import { submitEvents } from "../../api/eventsApi";
import { Send } from "lucide-react";

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

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

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
      if (form.type === "COUNT") {
        payload.quantity = parseInt(form.quantity, 10);
      } else {
        payload.target_event_id = form.target_event_id;
      }
      const res = await submitEvents(payload);
      setResult(res);
      onSuccess?.();
    } catch (err) {
      setResult({ error: err.message });
    } finally {
      setLoading(false);
    }
  };

  const statusColor = (status) => {
    const map = {
      ACCEPTED: "text-green-400",
      DUPLICATE: "text-purple-400",
      CONFLICT: "text-red-400",
      REJECTED: "text-red-400",
      PENDING_REFERENCE: "text-yellow-400",
    };
    return map[status] || "text-gray-400";
  };

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
      <h2 className="text-white font-semibold mb-3 text-sm uppercase tracking-wide">
        Submit Event — Factory Floor
      </h2>
      <form onSubmit={handleSubmit} className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-xs text-gray-400">Source ID</label>
            <input name="source_id" value={form.source_id} onChange={handleChange}
              className="w-full bg-gray-900 border border-gray-600 text-white rounded px-2 py-1.5 text-sm mt-1" />
          </div>
          <div>
            <label className="text-xs text-gray-400">Event ID</label>
            <input name="event_id" value={form.event_id} onChange={handleChange} required
              className="w-full bg-gray-900 border border-gray-600 text-white rounded px-2 py-1.5 text-sm mt-1"
              placeholder="EV-101" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-xs text-gray-400">Type</label>
            <select name="type" value={form.type} onChange={handleChange}
              className="w-full bg-gray-900 border border-gray-600 text-white rounded px-2 py-1.5 text-sm mt-1">
              <option value="COUNT">COUNT</option>
              <option value="VOID">VOID</option>
            </select>
          </div>
          {form.type === "COUNT" ? (
            <div>
              <label className="text-xs text-gray-400">Quantity</label>
              <input name="quantity" type="number" min="1" value={form.quantity} onChange={handleChange} required
                className="w-full bg-gray-900 border border-gray-600 text-white rounded px-2 py-1.5 text-sm mt-1"
                placeholder="5" />
            </div>
          ) : (
            <div>
              <label className="text-xs text-gray-400">Target Event ID</label>
              <input name="target_event_id" value={form.target_event_id} onChange={handleChange} required
                className="w-full bg-gray-900 border border-gray-600 text-white rounded px-2 py-1.5 text-sm mt-1"
                placeholder="EV-101" />
            </div>
          )}
        </div>

        <div>
          <label className="text-xs text-gray-400">Event Time</label>
          <input name="event_time" value={form.event_time} onChange={handleChange}
            className="w-full bg-gray-900 border border-gray-600 text-white rounded px-2 py-1.5 text-sm mt-1" />
        </div>

        <button type="submit" disabled={loading}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white px-4 py-2 rounded text-sm font-medium">
          <Send size={14} /> {loading ? "Submitting..." : "Submit Event"}
        </button>
      </form>

      {result && (
        <div className="mt-3 bg-gray-900 rounded p-3">
          {result.error ? (
            <p className="text-red-400 text-xs">{result.error}</p>
          ) : (
            result.results?.map((r, i) => (
              <div key={i} className="text-xs">
                <span className="text-gray-400">{r.event_id}: </span>
                <span className={`font-semibold ${statusColor(r.status)}`}>{r.status}</span>
                {r.message && <span className="text-gray-500 ml-2">— {r.message}</span>}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
