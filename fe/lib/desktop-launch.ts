export const DESKTOP_APP_LAUNCH_SOURCE = "desktop_app"

const DESKTOP_SOURCE_STORAGE_KEY = "launch_source"
const LEGACY_PWA_INSTALLED_STORAGE_KEY = "pwa_installed"
const PWA_INSTALLED_STORAGE_KEY = "pwa_installed_verified"

type NavigatorWithStandalone = Navigator & {
  standalone?: boolean
}

type NavigatorWithInstalledRelatedApps = Navigator & {
  getInstalledRelatedApps?: () => Promise<Array<Record<string, unknown>>>
}

export function isDesktopAppDisplayMode() {
  if (typeof window === "undefined") {
    return false
  }

  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    window.matchMedia("(display-mode: minimal-ui)").matches ||
    window.matchMedia("(display-mode: fullscreen)").matches ||
    window.matchMedia("(display-mode: window-controls-overlay)").matches ||
    (window.navigator as NavigatorWithStandalone).standalone === true
  )
}

export function rememberPwaInstalled() {
  if (typeof window === "undefined") {
    return
  }

  window.localStorage.setItem(PWA_INSTALLED_STORAGE_KEY, "true")
  window.localStorage.removeItem(LEGACY_PWA_INSTALLED_STORAGE_KEY)
}

export function isPwaInstallRemembered() {
  if (typeof window === "undefined") {
    return false
  }

  return window.localStorage.getItem(PWA_INSTALLED_STORAGE_KEY) === "true"
}

export async function hasInstalledRelatedPwa() {
  if (typeof navigator === "undefined") {
    return false
  }

  const getInstalledRelatedApps = (navigator as NavigatorWithInstalledRelatedApps).getInstalledRelatedApps
  if (!getInstalledRelatedApps) {
    return false
  }

  try {
    const relatedApps = await getInstalledRelatedApps.call(navigator)
    return relatedApps.length > 0
  } catch {
    return false
  }
}

export function registerPwaServiceWorker() {
  if (typeof window === "undefined" || !("serviceWorker" in navigator)) {
    return
  }

  const isLocalhost = ["localhost", "127.0.0.1", "[::1]"].includes(window.location.hostname)
  if (window.location.protocol !== "https:" && !isLocalhost) {
    return
  }

  window.addEventListener("load", () => {
    void navigator.serviceWorker.register("/sw.js").catch((error) => {
      console.error("PWA service worker registration failed:", error)
    })
  })
}

export function getLaunchSource() {
  if (typeof window === "undefined") {
    return "browser"
  }

  const params = new URLSearchParams(window.location.search)
  const source = params.get("source") || params.get("launch_source")
  if (source === DESKTOP_APP_LAUNCH_SOURCE) {
    window.sessionStorage.setItem(DESKTOP_SOURCE_STORAGE_KEY, DESKTOP_APP_LAUNCH_SOURCE)
    return DESKTOP_APP_LAUNCH_SOURCE
  }

  if (isDesktopAppDisplayMode()) {
    window.sessionStorage.setItem(DESKTOP_SOURCE_STORAGE_KEY, DESKTOP_APP_LAUNCH_SOURCE)
    rememberPwaInstalled()
    return DESKTOP_APP_LAUNCH_SOURCE
  }

  return window.sessionStorage.getItem(DESKTOP_SOURCE_STORAGE_KEY) || "browser"
}
