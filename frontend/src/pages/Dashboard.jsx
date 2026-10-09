import { useStateData } from "../hooks/useStateData";
import Navbar from "../components/layout/Navbar";
import StatTiles from "../components/ui/StatTiles";
import EventForm from "../components/ui/EventForm";
import PendingTable from "../components/ui/PendingTable";
import ExceptionsTable from "../components/ui/ExceptionsTable";
import MqttPanel from "../components/ui/MqttPanel";

export default function Dashboard() {
  const { summary, pending, exceptions, mqtt, loading, refresh } = useStateData();

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <Navbar mqtt={mqtt} />

      <main className="max-w-7xl mx-auto px-4 py-6 space-y-6">

        {/* Section 1 — Production Supervisor */}
        <section>
          <p className="text-xs text-gray-500 uppercase tracking-widest mb-3 font-semibold">
            Production Supervisor
          </p>
          <StatTiles summary={summary} loading={loading} />
          <div className="mt-4 grid grid-cols-1 lg:grid-cols-2 gap-4">
            <PendingTable events={pending} onRefresh={refresh} />
            <ExceptionsTable events={exceptions} />
          </div>
        </section>

        {/* Section 2 — Factory Floor Operator */}
        <section>
          <p className="text-xs text-gray-500 uppercase tracking-widest mb-3 font-semibold">
            Factory Floor Operator
          </p>
          <EventForm onSuccess={refresh} />
        </section>

        {/* Section 3 — Support / Engineering */}
        <section>
          <p className="text-xs text-gray-500 uppercase tracking-widest mb-3 font-semibold">
            Support / Engineering
          </p>
          <MqttPanel mqtt={mqtt} />
        </section>

      </main>
    </div>
  );
}
