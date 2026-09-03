"use client"

import { createContext, useContext, useState, useEffect, ReactNode } from "react"
import { authService } from "../../services/api"

interface AuthContextType {
  isLoggedIn: boolean
  isAuthReady: boolean
  setIsLoggedIn: (value: boolean) => void
  login: () => void
  logout: () => void
  checkAuth: () => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [isAuthReady, setIsAuthReady] = useState(false)

  useEffect(() => {
    // Initialize auth state from token
    const initializeAuth = async () => {
      const hasToken = authService.isAuthenticated()
      console.info("[auth] initialize", {
        hasToken,
        href: window.location.href,
        userAgent: navigator.userAgent,
        telegramPresent: Boolean((window as typeof window & { Telegram?: unknown }).Telegram),
      })
      // Telegram injects WebApp/initData asynchronously in some clients. Retry
      // briefly instead of checking only once during the initial render.
      let initData = ""
      let startParam = ""
      for (let attempt = 0; attempt < 30 && !initData; attempt += 1) {
        const webApp = (window as typeof window & { Telegram?: { WebApp?: { initData?: string; initDataUnsafe?: { start_param?: string }; ready?: () => void; expand?: () => void } } }).Telegram?.WebApp
        if (attempt === 0 || attempt % 5 === 0) {
          console.info("[auth] initData probe", {
            attempt: attempt + 1,
            webAppPresent: Boolean(webApp),
            initDataLength: webApp?.initData?.length || 0,
            initDataUnsafe: webApp?.initDataUnsafe,
          })
        }
        if (webApp?.initData) {
          initData = webApp.initData
          startParam = webApp.initDataUnsafe?.start_param || ""
          webApp.ready?.()
          webApp.expand?.()
          break
        }
        // Fallback for clients that expose Telegram data only in the URL.
        try {
          const params = new URLSearchParams(window.location.search)
          initData = params.get("tgWebAppData") || ""
          startParam = params.get("tgWebAppStartParam") || ""
          if (!initData && window.location.hash) {
            const hashParams = new URLSearchParams(window.location.hash.replace(/^#/, ""))
            initData = hashParams.get("tgWebAppData") || ""
            startParam = hashParams.get("tgWebAppStartParam") || startParam
          }
          if (attempt === 0) {
            console.info("[auth] URL fallback", {
              searchKeys: Array.from(params.keys()),
              hash: window.location.hash ? "present" : "empty",
              initDataLength: initData.length,
            })
          }
        } catch { /* ignore malformed URL */ }
        if (!initData) await new Promise((resolve) => setTimeout(resolve, 100))
      }
      if (initData) {
        try {
          console.info("[auth] Telegram initData found", { length: initData.length, startParam })
          await authService.telegramLogin(initData, startParam)
          setIsLoggedIn(true)
        } catch (error) {
          console.error("Telegram automatic login failed:", error instanceof Error ? error.message : JSON.stringify(error))
          // Keep an existing session if Telegram re-authentication fails.
          if (hasToken) setIsLoggedIn(true)
        }
      } else {
        if (hasToken) setIsLoggedIn(true)
        console.warn("[auth] No Telegram initData after probes", {
          telegram: Boolean((window as typeof window & { Telegram?: unknown }).Telegram),
          href: window.location.href,
          search: window.location.search,
          hash: window.location.hash,
        })
      }
      setIsAuthReady(true)
    }
    void initializeAuth()

    // Listen for auth events
    const handleUnauthorized = () => {
      setIsLoggedIn(false)
    }

    window.addEventListener("auth-unauthorized", handleUnauthorized)
    return () => window.removeEventListener("auth-unauthorized", handleUnauthorized)
  }, [])

  const login = () => {
    setIsLoggedIn(true)
  }

  const logout = () => {
    authService.logout()
    setIsLoggedIn(false)
  }

  const checkAuth = () => {
    return authService.isAuthenticated()
  }

  return (
    <AuthContext.Provider value={{ isLoggedIn, isAuthReady, setIsLoggedIn, login, logout, checkAuth }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return context
}
