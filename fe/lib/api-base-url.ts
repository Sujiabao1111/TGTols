const LOCAL_API_BASE_URL = "http://localhost:3555"

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "")

const isLocalHostname = (hostname: string) =>
  hostname === "localhost" || hostname === "127.0.0.1" || hostname === "::1"

export function getApiBaseUrl(): string {
  const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL?.trim()
  if (configuredBaseUrl) {
    return trimTrailingSlash(configuredBaseUrl)
  }

  if (typeof window !== "undefined") {
    const { protocol, hostname } = window.location
    if (isLocalHostname(hostname)) {
      return LOCAL_API_BASE_URL
    }

    return `${protocol}//${hostname}:3555`
  }

  return LOCAL_API_BASE_URL
}

export function apiUrl(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`
  return `${getApiBaseUrl()}${normalizedPath}`
}
