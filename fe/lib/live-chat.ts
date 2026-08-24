export type LiveChatVisibility = "maximized" | "minimized" | "hidden"

export type LiveChatWidgetApi = {
  _q?: unknown[]
  call: (method: string, payload?: unknown) => void
  init: () => void
  on: (event: string, callback: (payload: { visibility?: LiveChatVisibility }) => void) => void
  once: (event: string, callback: () => void) => void
}

declare global {
  interface Window {
    __lc?: {
      license?: number
      integration_name?: string
      product_name?: string
      asyncInit?: boolean
    }
    __liveChatInitialized?: boolean
    __liveChatHandlersBound?: boolean
    LiveChatWidget?: LiveChatWidgetApi
  }
}

function bindLiveChatHandlers() {
  if (typeof window === "undefined" || window.__liveChatHandlersBound || !window.LiveChatWidget) {
    return
  }

  window.LiveChatWidget.on("visibility_changed", ({ visibility }) => {
    if (visibility === "minimized") {
      window.LiveChatWidget?.call("hide")
    }
  })

  window.__liveChatHandlersBound = true
}

export function openLiveChat(fallbackUrl?: string) {
  if (typeof window === "undefined") {
    return
  }

  const widget = window.LiveChatWidget

  if (!widget) {
    if (fallbackUrl) {
      window.location.assign(fallbackUrl)
    }
    return
  }

  bindLiveChatHandlers()

  if (!window.__liveChatInitialized) {
    window.__liveChatInitialized = true
    widget.once("ready", () => {
      bindLiveChatHandlers()
      widget.call("maximize")
    })
    widget.init()
    return
  }

  widget.call("maximize")
}
