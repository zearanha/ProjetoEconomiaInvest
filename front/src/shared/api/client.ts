const apiBaseURL = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseURL}${path}`, { ...init, headers: { 'Content-Type': 'application/json', ...init?.headers } })
  if (!response.ok) {
    let message = `Falha na API (${response.status})`
    try { message = (await response.json()).error ?? message } catch { /* resposta sem JSON */ }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}
