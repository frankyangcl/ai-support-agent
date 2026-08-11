const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export type DocumentItem = {
  id: number;
  filename: string;
  created_at: string;
};

export type Source = {
  document_id: number;
  filename: string;
  chunk_index: number;
  distance: number;
  preview: string;
};

export type ChatResponse = {
  answer: string;
  sources: Source[];
};

export async function getDocuments(): Promise<DocumentItem[]> {
  const response = await fetch(`${API_BASE_URL}/api/documents`, {
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error("Failed to load documents");
  }

  const data = await response.json();

  return data.documents ?? [];
}

export async function askQuestion(
  question: string
): Promise<ChatResponse> {
  const response = await fetch(`${API_BASE_URL}/api/chat`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      question,
    }),
  });

  if (!response.ok) {
    const data = await response.json().catch(() => null);

    throw new Error(
      data?.error ?? "Failed to get answer"
    );
  }

  return response.json();
}

export async function uploadDocument(
  file: File
): Promise<void> {
  const formData = new FormData();

  formData.append("file", file);

  const response = await fetch(
    `${API_BASE_URL}/api/documents/upload`,
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
}

export async function deleteDocument(
  id: number
): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/api/documents/${id}`,
    {
      method: "DELETE",
    }
  );

  if (!response.ok) {
    const data = await response.json().catch(() => null);

    throw new Error(
      data?.error ?? "Failed to delete document"
    );
  }
}