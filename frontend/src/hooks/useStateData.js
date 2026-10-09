import { useState, useEffect, useCallback } from "react";
import { getMqttStatus } from "../api/stateApi";

export function useStateData() {
  const [summary, setSummary] = useState(null);
  const [pending, setPending] = useState([]);
  const [exceptions, setExceptions] = useState([]);
  const [mqtt, setMqtt] = useState(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const [allRes, mqttRes] = await Promise.all([
        fetch("/api/state/all").then((r) => r.json()),
        getMqttStatus(),
      ]);
      setSummary(allRes.summary);
      setPending(allRes.pending || []);
      setExceptions(allRes.exceptions || []);
      setMqtt(mqttRes);
    } catch (err) {
      console.error("State fetch error:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
    const id = setInterval(refresh, 5000);
    return () => clearInterval(id);
  }, [refresh]);

  return { summary, pending, exceptions, mqtt, loading, refresh };
}
