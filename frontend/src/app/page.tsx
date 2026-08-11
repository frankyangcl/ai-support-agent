import ChatPanel from "@/components/ChatPanel";
import UploadDocument from "@/components/UploadDocument";
import { getDocuments } from "@/lib/api";
import DeleteDocumentButton from "@/components/DeleteDocumentButton";
export default async function Home() {
  const documents = await getDocuments();

  return (
    <main className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-5">
          <div>
            <h1 className="text-xl font-semibold tracking-tight text-slate-900">
              AI Support Agent
            </h1>
            <p className="mt-1 text-sm text-slate-500">
              Ask questions and get grounded answers from your documents.
            </p>
          </div>

          <div className="rounded-full bg-emerald-50 px-3 py-1.5 text-xs font-medium text-emerald-700">
            Knowledge base ready
          </div>
        </div>
      </header>

      <div className="mx-auto grid max-w-7xl gap-6 px-6 py-8 lg:grid-cols-[320px_minmax(0,1fr)]">
        <aside className="self-start rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-100 p-5">
            <h2 className="font-semibold text-slate-900">
              Knowledge Base
            </h2>

            <p className="mt-1 text-sm text-slate-500">
              {documents.length}{" "}
              {documents.length === 1 ? "document" : "documents"}
            </p>

            <div className="mt-4">
              <UploadDocument />
            </div>
          </div>

          <div className="p-3">
            {documents.length === 0 ? (
              <div className="rounded-xl border border-dashed border-slate-200 p-6 text-center">
                <p className="text-sm text-slate-500">
                  Upload a PDF to create your knowledge base.
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {documents.map((document) => (
                  <div
                    key={document.id}
                    className="rounded-xl border border-slate-200 p-3 transition hover:bg-slate-50"
                  >
                    <div className="flex items-start gap-3">
                      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-xs font-semibold text-slate-600">
                        PDF
                      </div>

                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-slate-900">
                          {document.filename}
                        </p>

                        <p className="mt-1 text-xs text-slate-500">
                          Ready · Document #{document.id}
                        </p>

                          <DeleteDocumentButton
                            documentId={document.id}
                            filename={document.filename}
                          />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </aside>

        <section className="min-h-[700px] rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-100 px-6 py-5">
            <h2 className="font-semibold text-slate-900">
              Ask your knowledge base
            </h2>

            <p className="mt-1 text-sm text-slate-500">
              Answers are generated only from relevant uploaded content.
            </p>
          </div>

          <div className="p-6">
            <ChatPanel />
          </div>
        </section>
      </div>
    </main>
  );
}