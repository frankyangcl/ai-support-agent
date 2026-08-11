"use client";

import { useState } from "react";

import {
  askQuestion,
  type Source,
} from "@/lib/api";

type Message = {
  id: number;
  question: string;
  answer: string;
  sources: Source[];
};

export default function ChatPanel() {
  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    const trimmedQuestion = question.trim();

    if (!trimmedQuestion || loading) {
      return;
    }

    setLoading(true);
    setError("");

    try {
      const data = await askQuestion(trimmedQuestion);
      
      setMessages((current) => [
        ...current,
        {
          id: Date.now(),
          question: trimmedQuestion,
          answer: data.answer,
          sources: data.sources ?? [],
        },
      ]);

      setQuestion("");
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
    <div className="flex min-h-[570px] flex-col">
      <div className="flex-1">
        {messages.length === 0 ? (
          <div className="flex min-h-[360px] items-center justify-center">
            <div className="max-w-md text-center">
              <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-slate-100 text-lg">
                AI
              </div>

              <h3 className="font-medium text-slate-900">
                Ask your first question
              </h3>

              <p className="mt-2 text-sm leading-6 text-slate-500">
                I&apos;ll search your uploaded documents and answer
                using the most relevant information.
              </p>
            </div>
          </div>
        ) : (
          <div className="space-y-8 pb-8">
            {messages.map((message) => (
              <div key={message.id}>
                <div className="flex justify-end">
                  <div className="max-w-[75%]">
                    <div className="mb-2 text-right text-xs font-medium text-slate-500">
                      You
                    </div>

                    <div className="rounded-2xl rounded-tr-md bg-slate-900 px-4 py-3 text-sm leading-6 text-white">
                      {message.question}
                    </div>
                  </div>
                </div>

                <div className="mt-6 max-w-[85%]">
                  <div className="mb-2 text-xs font-medium text-slate-500">
                    AI Support Agent
                  </div>

                  <div className="rounded-2xl rounded-tl-md bg-slate-100 px-4 py-3 text-sm leading-6 text-slate-800">
                    {message.answer}
                  </div>

                  {message.sources.length > 0 && (
                    <div className="mt-3">
                      <p className="mb-2 text-xs font-medium text-slate-500">
                        Sources
                      </p>

                      <div className="space-y-2">
                        {message.sources.map((source, index) => (
                          <details
                            key={`${source.document_id}-${source.chunk_index}-${index}`}
                            className="group rounded-xl border border-slate-200 bg-white"
                            >
                            <summary className="cursor-pointer list-none p-3">
                                <div className="flex items-center justify-between gap-4">
                                <div className="min-w-0">
                                    <span className="block truncate text-xs font-medium text-slate-700">
                                    [{index + 1}] {source.filename}
                                    </span>

                                    <span className="mt-1 block text-xs text-slate-400">
                                    Chunk {source.chunk_index}
                                    </span>
                                </div>

                                <span className="shrink-0 text-xs text-slate-400 transition group-open:rotate-180">
                                    ▼
                                </span>
                                </div>
                            </summary>

                            <div className="border-t border-slate-100 px-3 py-3">
                                <p className="whitespace-pre-wrap text-xs leading-5 text-slate-600">
                                {source.preview}
                                </p>
                            </div>
                            </details>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}

            {loading && (
              <div className="max-w-[85%]">
                <div className="mb-2 text-xs font-medium text-slate-500">
                  AI Support Agent
                </div>

                <div className="inline-block rounded-2xl rounded-tl-md bg-slate-100 px-4 py-3 text-sm text-slate-500">
                  Searching knowledge base...
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="sticky bottom-0 border-t border-slate-100 bg-white pt-5">
        <form onSubmit={handleSubmit}>
          <div className="rounded-xl border border-slate-200 bg-white p-2 shadow-sm transition focus-within:border-slate-400 focus-within:ring-4 focus-within:ring-slate-100">
            <textarea
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder={
                messages.length === 0
                  ? "Ask a question about your documents..."
                  : "Ask another question..."
              }
              className="min-h-20 w-full resize-none border-0 p-2 text-sm text-slate-900 outline-none placeholder:text-slate-400"
            />

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={loading || !question.trim()}
                className="rounded-lg bg-slate-900 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {loading ? "Thinking..." : "Send"}
              </button>
            </div>
          </div>

          {error && (
            <p className="mt-2 text-sm text-red-600">
              {error}
            </p>
          )}
        </form>
      </div>
    </div>
  );
}