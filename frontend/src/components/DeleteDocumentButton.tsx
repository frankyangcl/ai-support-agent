"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { deleteDocument } from "@/lib/api";

type Props = {
  documentId: number;
  filename: string;
};

export default function DeleteDocumentButton({
  documentId,
  filename,
}: Props) {
  const router = useRouter();

  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");

  async function handleDelete() {
    const confirmed = window.confirm(
      `Delete "${filename}" from the knowledge base?`
    );

    if (!confirmed) {
      return;
    }

    setDeleting(true);
    setError("");

    try {
      await deleteDocument(documentId);

      router.refresh();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to delete document"
      );
    } finally {
      setDeleting(false);
    }
  }

  return (
    <div>
      <button
        type="button"
        disabled={deleting}
        onClick={handleDelete}
        className="text-xs font-medium text-red-500 transition hover:text-red-700 disabled:opacity-50"
      >
        {deleting ? "Deleting..." : "Delete"}
      </button>

      {error && (
        <p className="mt-1 text-xs text-red-600">
          {error}
        </p>
      )}
    </div>
  );
}