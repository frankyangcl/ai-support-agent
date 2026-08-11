"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { uploadDocument } from "@/lib/api";

export default function UploadDocument() {
  const router = useRouter();

  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  async function handleChange(
    event: React.ChangeEvent<HTMLInputElement>
  ) {
    const file = event.target.files?.[0];

    if (!file) {
      return;
    }

    setUploading(true);
    setError("");

    try {
      await uploadDocument(file);

      router.refresh();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Something went wrong"
      );
    } finally {
      setUploading(false);
      event.target.value = "";
    }
  }

  return (
    <div className="mb-5">
     <label className="flex w-full cursor-pointer items-center justify-center rounded-lg bg-slate-900 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-slate-800">
  {uploading ? "Processing document..." : "+ Upload PDF"}

  <input
    type="file"
    accept="application/pdf,.pdf"
    disabled={uploading}
    onChange={handleChange}
    className="hidden"
  />
</label>

      {uploading && (
        <p className="mt-2 text-xs text-gray-500">
          Parsing document and generating embeddings...
        </p>
      )}

      {error && (
        <p className="mt-2 text-sm text-red-600">
          {error}
        </p>
      )}
    </div>
  );
}