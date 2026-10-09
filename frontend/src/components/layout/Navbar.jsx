import { Wifi, WifiOff } from "lucide-react";

export default function Navbar({ mqtt }) {
  const connected = mqtt?.connected;
  return (
    <nav className="bg-gray-900 border-b border-gray-700 px-6 py-3 flex items-center justify-between">
      <div>
        <h1 className="text-white font-semibold text-lg">NorthBridge Garments</h1>
        <p className="text-gray-400 text-xs">Production Monitor</p>
      </div>
      <div className="flex items-center gap-3">
        {mqtt?.candidate_id && (
          <span className="text-xs text-gray-400 font-mono">
            ID: {mqtt.candidate_id}
          </span>
        )}
        <div className={`flex items-center gap-1.5 text-xs font-medium px-2 py-1 rounded-full ${connected ? "bg-green-900 text-green-300" : "bg-red-900 text-red-300"}`}>
          {connected ? <Wifi size={12} /> : <WifiOff size={12} />}
          MQTT {connected ? "ONLINE" : "OFFLINE"}
        </div>
      </div>
    </nav>
  );
}
