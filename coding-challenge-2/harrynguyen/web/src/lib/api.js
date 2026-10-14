const base = import.meta.env.VITE_API_BASE ?? ""

/**
 * @param {string} path
 * @param {Record<string, unknown>} body
 */
export async function apiPost(path, body) {
  const res = await fetch(`${base}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    const err = new Error(
      typeof data.error === "string" ? data.error : res.statusText || "Request failed",
    )
    err.status = res.status
    err.body = data
    throw err
  }
  return data
}
