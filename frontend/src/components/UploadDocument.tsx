"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

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
      const formData = new FormData();
      formData.append("file", file);

      const response = await fetch(
        "http://localhost:8080/api/documents/upload",
        {
          method: "POST",
          body: formData,
        }
      );

      if (!response.ok) {
        const data = await response.json().catch(() => null);

        throw new Error(
          data?.error ?? "Failed to upload document"
        );
      }

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
      <label className="inline-flex cursor-pointer items-center rounded-lg bg-black px-4 py-2 text-sm text-white">
        {uploading ? "Processing..." : "Upload PDF"}

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