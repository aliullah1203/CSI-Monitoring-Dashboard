import { Radio, AlertCircle } from "lucide-react";

export default function MqttPanel({ mqtt }) {
  if (!mqtt) return null;

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
      <h2 className="text-white font-semibold text-sm uppercase tracking-wide mb-3 flex items-center gap-2">
        <Radio size={14} /> MQTT — Support / Engineering
      </h2>
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-sm">
        <div>
          <p className="text-gray-400 text-xs">Status</p>
          <p className={`font-semibold mt-0.5 ${mqtt.connected ? "text-green-400" : "text-red-400"}`}>
            {mqtt.connected ? "ONLINE" : "OFFLINE"}
          </p>
        </div>
        <div>
          <p className="text-gray-400 text-xs">Candidate ID</p>
          <p className="text-white font-mono mt-0.5">{mqtt.candidate_id || "—"}</p>
        </div>
        <div>
          <p className="text-gray-400 text-xs">Last Challenge</p>
          <p className="text-blue-300 font-mono mt-0.5">{mqtt.last_challenge_id || "—"}</p>
        </div>
        <div>
          <p className="text-gray-400 text-xs">Last Response</p>
          <p className={`font-medium mt-0.5 ${mqtt.last_status?.includes("FAILED") ? "text-red-400" : "text-green-400"}`}>
            {mqtt.last_status || "—"}
          </p>
        </div>
        <div>
          <p className="text-gray-400 text-xs">Last Seen</p>
          <p className="text-gray-300 text-xs mt-0.5">
            {mqtt.last_seen_at && mqtt.last_seen_at !== "0001-01-01T00:00:00Z"
              ? new Date(mqtt.last_seen_at).toLocaleString()
              : "—"}
          </p>
        </div>
        {mqtt.last_error && (
          <div className="col-span-2 sm:col-span-1">
            <p className="text-gray-400 text-xs flex items-center gap-1">
              <AlertCircle size={11} /> Last Error
            </p>
            <p className="text-red-400 text-xs mt-0.5">{mqtt.last_error}</p>
          </div>
        )}
      </div>
    </div>
  );
}
