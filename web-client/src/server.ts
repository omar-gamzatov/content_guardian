import express from "express";
// import morgan from "morgan";
import path from "node:path";
import fetch from "node-fetch";

const app = express();
const PORT = process.env.PORT ? Number(process.env.PORT) : 3000;
// В docker-compose будем использовать внутренний адрес API
// Пример: http://api:8080
const API_BASE = process.env.API_BASE || "http://api:8080";

app.disable("x-powered-by");
// app.use(morgan("dev"));
app.use(express.json({ limit: "256kb" }));
app.use(express.static(path.join(process.cwd(), "dist/public"), { maxAge: "1h" }));

// Health
app.get("/healthz", (_req, res) => res.status(200).send("ok"));

// Прокси в API: POST /api/moderations -> API_BASE/v1/moderations
app.post("/api/moderations", async (req, res) => {
  try {
    const url = `${API_BASE}/v1/moderations`;
    const headers: Record<string, string> = {
      "content-type": "application/json"
    };
    // Прокидываем Authorization, если фронт его отправил (не обязательно)
    const auth = req.header("authorization");
    if (auth) headers["authorization"] = auth;

    const r = await fetch(url, {
      method: "POST",
      headers,
      body: JSON.stringify(req.body),
      // таймауты можно реализовать через AbortController при необходимости
    });

    const text = await r.text();
    res.status(r.status).set({
      "content-type": r.headers.get("content-type") || "application/json"
    }).send(text);
  } catch (e: any) {
    console.error("Proxy error:", e?.message || e);
    res.status(502).json({ error: "Bad gateway", detail: e?.message || String(e) });
  }
});

// Отдаём index.html для корня и неизвестных путей (SPA)
app.get("*", (_req, res) => {
  res.sendFile(path.join(process.cwd(), "dist/public/index.html"));
});

app.listen(PORT, "0.0.0.0", () => {
  console.log(`[webclient] listening on http://0.0.0.0:${PORT} (API=${API_BASE})`);
});
