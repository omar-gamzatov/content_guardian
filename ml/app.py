from fastapi import FastAPI
from pydantic import BaseModel
from detoxify import Detoxify
from presidio_analyzer import AnalyzerEngine
from presidio_anonymizer import AnonymizerEngine

app = FastAPI()
tox = Detoxify('multilingual')  # CPU ok, можно dockerize
analyzer = AnalyzerEngine()
anonymizer = AnonymizerEngine()

class In(BaseModel):
    text: str
    lang: str | None = None
    pii_redact: bool = False

@app.post("/classify")
def classify(inp: In):
    text = inp.text
    pii_spans = []
    if inp.pii_redact:
        results = analyzer.analyze(text=text, language='en')  # для RU добавь custom recognizers
        pii_spans = [{"entity_type": r.entity_type, "start": r.start, "end": r.end, "score": r.score} for r in results]
        text = anonymizer.anonymize(text, analyzer_results=results).text

    scores = tox.predict(text)
    # Detoxify keys: toxicity, severe_toxicity, obscene, identity_attack, insult, threat, sexual_explicit
    mapped = {
        "toxicity": float(scores.get("toxicity", 0.0)),
        "identity_attack": float(scores.get("identity_attack", 0.0)),
        "violence_threat": float(scores.get("threat", 0.0)),
        "sexual_explicit": float(scores.get("sexual_explicit", 0.0)),
        "profanity": float(scores.get("obscene", 0.0)),
        "insult": float(scores.get("insult", 0.0)),
    }
    return {
        "categories": [{"name": k, "score": v, "source": "model"} for k, v in mapped.items()],
        "explain": {"model": "detoxify-multilingual", "pii_spans": pii_spans}
    }
