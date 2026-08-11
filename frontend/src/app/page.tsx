import { getDocuments } from "@/lib/api";
import ChatPanel from "@/components/ChatPanel";
import UploadDocument from "@/components/UploadDocument";

export default async function Home() {
  const documents = await getDocuments();

  return (
    <main className="min-h-screen bg-gray-50 p-8">
      <div className="mx-auto max-w-6xl">
        <h1 className="mb-6 text-3xl font-semibold">
          AI Support Agent
        </h1>

        <div className="grid gap-6 md:grid-cols-[320px_1fr]">
          <section className="rounded-xl border bg-white p-5">
            <h2 className="mb-4 text-lg font-medium">
              Knowledge Base
            </h2>
<UploadDocument />
            {documents.length === 0 ? (
              <p className="text-sm text-gray-500">
                No documents uploaded yet.
              </p>
            ) : (
              <div className="space-y-3">
                {documents.map((document) => (
                  <div
                    key={document.id}
                    className="rounded-lg border p-3"
                  >
                    <div className="font-medium">
                      {document.filename}
                    </div>

                    <div className="mt-1 text-xs text-gray-500">
                      Document #{document.id}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          <section className="rounded-xl border bg-white p-5">
            <h2 className="mb-4 text-lg font-medium">
              Ask your knowledge base
            </h2>

            <ChatPanel />
          </section>
        </div>
      </div>
    </main>
  );
}