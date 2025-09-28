// apps/rules-engine/src/index.ts
import express from "express";
import jsonLogic from "json-logic-js";

const app = express();
app.use(express.json({ limit: "128kb" }));

// Простейшая проверка правил: клиент присылает policy + signals
app.post("/v1/rules/evaluate", (req, res) => {
  const { policy, signals } = req.body || {};
  if (!policy || typeof policy !== "object") {
    return res.status(400).json({ error: "policy required" });
  }
  if (!signals || typeof signals !== "object") {
    return res.status(400).json({ error: "signals required" });
  }
  try {
    const result = jsonLogic.apply(policy, signals);
    return res.json({ decision: result });
  } catch (e: any) {
    return res.status(400).json({ error: e?.message || "invalid policy" });
  }
});

app.get("/healthz", (_req, res) => res.send("ok"));

const port = process.env.PORT || 3000;
app.listen(port, () => console.log(`rules-engine listening on :${port}`));
