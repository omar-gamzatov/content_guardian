// web-client/src/client.ts
type CategoryScore = { name: string; score: number; source: "rule" | "model" };
type Verdict = {
  action: "allow" | "soft_block" | "block" | "escalate";
  severity?: "low" | "medium" | "high";
  categories: CategoryScore[];
  explain?: any;
};
type ModerationResponse = {
  request_id: string;
  verdict: Verdict;
  sla: { latency_ms: number; mode: string };
};

const $ = (sel: string) => document.querySelector(sel) as HTMLElement | null;

function setText(selector: string, val: string) {
  const el = $(selector);
  if (el) el.textContent = val;
}

function setJSON(selector: string, obj: any) {
  const el = $(selector);
  if (el) el.textContent = JSON.stringify(obj, null, 2);
}

function badge(text: string, kind: "action" | "severity") {
  const span = document.createElement("span");
  span.className = `badge ${kind}-${text.toLowerCase().replace(/\s+/g, "-")}`;
  span.textContent = text.toUpperCase();
  return span;
}

function progress(score: number) {
  const wrap = document.createElement("div");
  wrap.className = "progress";
  const bar = document.createElement("div");
  bar.className = "bar";
  bar.style.width = `${Math.round(score * 100)}%`;
  if (score >= 0.9) bar.classList.add("p90");
  else if (score >= 0.7) bar.classList.add("p70");
  else if (score >= 0.4) bar.classList.add("p40");
  wrap.appendChild(bar);
  return wrap;
}

function capitalize(s: string) {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function humanName(cat: string) {
  const map: Record<string, string> = {
    toxicity: "Toxicity",
    identity_attack: "Identity attack",
    violence_threat: "Threat/Violence",
    sexual_explicit: "Sexual explicit",
    profanity: "Profanity",
    insult: "Insult"
  };
  return map[cat] || cat.replace(/_/g, " ");
}

function renderVerdict(v: Verdict, sla?: { latency_ms: number; mode: string }) {
  const container = document.createElement("div");
  container.className = "verdict";

  // Top line: Action + Severity + SLA
  const header = document.createElement("div");
  header.className = "verdict-header";

  header.appendChild(badge(v.action, "action"));
  header.appendChild(badge(v.severity || "low", "severity"));

  if (sla) {
    const meta = document.createElement("span");
    meta.className = "meta";
    meta.textContent = `Latency: ${sla.latency_ms} ms · Mode: ${sla.mode}`;
    header.appendChild(meta);
  }
  container.appendChild(header);

  // Categories table
  const cats = [...(v.categories || [])].sort((a, b) => b.score - a.score);
  const list = document.createElement("div");
  list.className = "categories";

  for (const c of cats) {
    const row = document.createElement("div");
    row.className = "cat-row";

    const left = document.createElement("div");
    left.className = "cat-left";
    const name = document.createElement("div");
    name.className = "cat-name";
    name.textContent = humanName(c.name);
    const src = document.createElement("div");
    src.className = "cat-src";
    src.textContent = c.source === "rule" ? "rule" : "model";
    left.appendChild(name);
    left.appendChild(src);

    const right = document.createElement("div");
    right.className = "cat-right";
    const val = document.createElement("div");
    val.className = "cat-score";
    val.textContent = (c.score * 100).toFixed(1) + "%";
    right.appendChild(val);
    right.appendChild(progress(c.score));

    row.appendChild(left);
    row.appendChild(right);
    list.appendChild(row);
  }
  container.appendChild(list);

  // Explain (compact)
  if (v.explain) {
    const ex = document.createElement("div");
    ex.className = "explain";
    const rules = v.explain["rules_fired"] as string[] | undefined;
    const model = v.explain["model"];
    const unc = v.explain["uncertainty"];
    const policy = v.explain["policy_version"];

    const parts: string[] = [];
    if (policy) parts.push(`Policy: ${policy}`);
    if (Array.isArray(rules) && rules.length) parts.push(`Rules: ${rules.join(", ")}`);
    if (unc != null) parts.push(`Uncertainty: ${Number(unc).toFixed(2)}`);
    if (model) {
      const m = typeof model === "string" ? model : (model.model || "model");
      parts.push(`Model: ${m}`);
    }
    ex.textContent = parts.join(" · ");
    container.appendChild(ex);
  }

  return container;
}

function renderVerdictInto(selector: string, v: Verdict, sla?: { latency_ms: number; mode: string }) {
  const mount = $(selector);
  if (!mount) return;
  // Заменяем pre на div, если нужно
  if (mount.tagName.toLowerCase() === "pre") {
    const parent = mount.parentElement!;
    const replacement = document.createElement("div");
    replacement.id = (mount as HTMLElement).id;
    parent.replaceChild(replacement, mount);
  }
  const node = renderVerdict(v, sla);
  const target = $(selector)!;
  target.innerHTML = "";
  target.appendChild(node);
}

function buildBody() {
  const tenant = (document.getElementById("tenant") as HTMLInputElement).value.trim() || "default";
  const policy = (document.getElementById("policy") as HTMLInputElement).value.trim();
  const mode = (document.getElementById("mode") as HTMLSelectElement).value;
  const text = (document.getElementById("text") as HTMLTextAreaElement).value;

  const body: any = {
    tenant_id: tenant,
    request_id: crypto.randomUUID(),
    response_mode: mode,
    content: { type: "text", text },
    metadata: { client: "web-demo" }
  };
  if (policy) body.policy_version = policy;
  return body;
}

async function onSend() {
  const btn = document.getElementById("send") as HTMLButtonElement;
  const token = (document.getElementById("apiToken") as HTMLInputElement).value.trim();
  btn.disabled = true;
  setText("#status", "Отправка...");
  setJSON("#raw", {}); // сырые данные оставляем как есть

  try {
    const body = buildBody();
    const r = await fetch("/api/moderations", {
      method: "POST",
      headers: {
        "content-type": "application/json",
        ...(token ? { authorization: `Bearer ${token}` } : {})
      },
      body: JSON.stringify(body)
    });

    const data = (await r.json()) as ModerationResponse;
    setText("#status", r.ok ? `Готово (${data.sla.latency_ms} ms)` : "Ошибка");
    renderVerdictInto("#verdict", data.verdict, data.sla);
    setJSON("#raw", data);
  } catch (e: any) {
    setText("#status", `Ошибка: ${e?.message || String(e)}`);
  } finally {
    btn.disabled = false;
  }
}

function main() {
  const btn = document.getElementById("send") as HTMLButtonElement;
  btn?.addEventListener("click", onSend);
  const text = document.getElementById("text") as HTMLTextAreaElement;
  text?.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") onSend();
  });
}

document.addEventListener("DOMContentLoaded", main);
