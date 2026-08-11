"use client";

import { useState } from "react";

type Source = {
  document_id: number;
  filename: string;
  chunk_index: number;
  distance: number;
  preview: string;
};

type ChatResponse = {
  answer: string;
  sources: Source[];
};

export default function ChatPanel() {
  const [question, setQuestion] = useState("");
  const [answer, setAnswer] = useState("");
  const [sources, setSources] = useState<Source[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!question.trim()) {
      return;
    }

    setLoading(true);
    setError("");
    setAnswer("");
    setSources([]);

    try {
      const response = await fetch(
        "http://localhost:8080/api/chat",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            question: question.trim(),
          }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to get answer");
      }

      const data: ChatResponse = await response.json();

      setAnswer(data.answer);
      setSources(data.sources ?? []);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Something went wrong"
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <form
        onSubmit={handleSubmit}
        className="space-y-4"
      >
        <textarea
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          placeholder="Ask a question about your documents..."
          className="min-h-28 w-full rounded-lg border p-3"
        />

        <button
          type="submit"
          disabled={loading}
          className="rounded-lg bg-black px-4 py-2 text-white disabled:opacity-50"
        >
          {loading ? "Thinking..." : "Ask"}
        </button>
      </form>

      {error && (
        <p className="mt-4 text-sm text-red-600">
          {error}
        </p>
      )}

      {answer && (
        <div className="mt-6">
          <h3 className="mb-2 font-medium">
            Answer
          </h3>

          <p className="whitespace-pre-wrap text-sm leading-6">
            {answer}
          </p>
        </div>
      )}

      {sources.length > 0 && (
        <div className="mt-6">
          <h3 className="mb-3 font-medium">
            Sources
          </h3>

          <div className="space-y-3">
            {sources.map((source, index) => (
              <div
                key={`${source.document_id}-${source.chunk_index}-${index}`}
                className="rounded-lg border p-3"
              >
                <div className="text-sm font-medium">
                  {source.filename}
                </div>

                <div className="mt-1 text-xs text-gray-500">
                  Chunk {source.chunk_index}
                </div>

                <p className="mt-2 text-xs leading-5 text-gray-600">
                  {source.preview}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}