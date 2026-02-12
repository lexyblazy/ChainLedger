export async function makeRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  // simulate slow network
  //   await delay(500);

  const url = `${process.env.NEXT_PUBLIC_API_BASE_URL}${path}`;
  const res = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
    },
    ...options,
  });

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    const message = data?.error || data?.message || "Request failed";

    throw new Error(message);
  }

  return data;
}

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));
