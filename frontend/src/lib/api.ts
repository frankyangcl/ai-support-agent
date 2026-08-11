const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export type DocumentItem = {
  id: number;
  filename: string;
  created_at: string;
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