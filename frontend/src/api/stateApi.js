const BASE = "/api";

export async function getSummary(sourceID = "") {
  const url = sourceID
    ? `${BASE}/state?view=summary&source_id=${sourceID}`
    : `${BASE}/state?view=summary`;
  const res = await fetch(url);
  return res.json();
}

export async function getPending(sourceID = "") {
  const url = sourceID
    ? `${BASE}/state?view=pending&source_id=${sourceID}`
    : `${BASE}/state?view=pending`;
  const res = await fetch(url);
  return res.json();
}

export async function getExceptions(sourceID = "") {
  const url = sourceID
    ? `${BASE}/state?view=exceptions&source_id=${sourceID}`
    : `${BASE}/state?view=exceptions`;
  const res = await fetch(url);
  return res.json();
}

export async function getMqttStatus() {
  const res = await fetch(`${BASE}/mqtt/status`);
  return res.json();
}
