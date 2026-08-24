"use client"

import { createContext, useContext, useState, useEffect, ReactNode } from "react"
import { authService } from "../../services/api"

interface AuthContextType {
  isLoggedIn: boolean
  setIsLoggedIn: (value: boolean) => void
  login: () => void
  logout: () => void
  checkAuth: () => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false)

  useEffect(() => {
    // Initialize auth state from token
    const hasToken = authService.isAuthenticated()
    setIsLoggedIn(hasToken)

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
    <AuthContext.Provider value={{ isLoggedIn, setIsLoggedIn, login, logout, checkAuth }}>
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
