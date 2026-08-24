"use client"

import type React from "react"
import { useState } from "react"
import Image from "next/image"
import { User, Lock, ArrowRight, Eye, EyeOff, AlertCircle, X } from "lucide-react"
import { useRouter } from "next/navigation"
import { useLanguage } from "../app/contexts/LanguageContext"
import { authService } from "../services/api"

interface LoginPageProps {
  onLogin: () => void
  onGoToRegister: () => void
}

const LoginPage: React.FC<LoginPageProps> = ({ onLogin, onGoToRegister }) => {
  const { t } = useLanguage()
  const router = useRouter()
  const [showPassword, setShowPassword] = useState(false)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const translateAuthMessage = (message: string) => {
    const normalized = message
      .replace(/^http\s+\d+\s*:\s*/i, "")
      .replace(/^http error!\s*status:\s*\d+\s*/i, "")
      .trim()

    if (normalized === "AUTH_INVALID_CREDENTIALS") return t("auth.invalid_credentials")
    if (normalized === "AUTH_ACCOUNT_DISABLED") return t("auth.account_disabled")
    if (normalized === "AUTH_MISSING_CREDENTIALS") return t("auth.missing_credentials")

    return normalized || t("auth.connection_failed")
  }

  const resolveLoginError = (err: unknown) => {
    if (!(err instanceof Error)) {
      return t("auth.connection_failed")
    }

    const message = err.message || ""
    const lower = message.toLowerCase()
    if (
      message === "AUTH_INVALID_CREDENTIALS" ||
      message === "AUTH_ACCOUNT_DISABLED" ||
      message === "AUTH_MISSING_CREDENTIALS" ||
      lower.includes("http 401") ||
      lower.includes("status: 401") ||
      lower.includes("http 403") ||
      lower.includes("status: 403") ||
      lower.includes("http 400") ||
      lower.includes("status: 400")
    ) {
      return translateAuthMessage(message)
    }

    return t("auth.connection_failed")
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username || !password) return

    setIsLoading(true)
    setError(null)

    try {
      const data = await authService.login(username, password)

      if (data.success && data.token) {
        localStorage.setItem("token", data.token)
        onLogin()
      } else {
        setError(translateAuthMessage(data.message || t("auth.invalid_credentials")))
      }
    } catch (err) {
      console.error(err)
      setError(resolveLoginError(err))
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[url('https://images.unsplash.com/photo-1511193311914-0346f16efe90?q=80&w=2073&auto=format&fit=crop')] bg-cover bg-center flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-lucky-dark/80 backdrop-blur-sm"></div>

      <div className="relative w-full max-w-md bg-lucky-dark/90 border border-white/10 rounded-3xl p-8 shadow-2xl animate-fade-in">
        <button
          type="button"
          onClick={() => router.push("/home")}
          className="absolute right-4 top-4 inline-flex h-10 w-10 items-center justify-center rounded-full bg-white/5 text-white/70 transition-colors hover:bg-white/10 hover:text-white"
          aria-label="Close login and return home"
        >
          <X size={20} />
        </button>

        <div className="mb-8 flex flex-col items-center">
          <div className="relative mb-3 h-16 w-16">
            <Image
              src="/images/PPLogo.png"
              alt="PP Logo"
              fill
              sizes="64px"
              className="object-contain"
            />
          </div>
          <h1 className="text-3xl font-display font-bold text-white">{t("auth.welcome")}</h1>
          <p className="mt-1 text-sm text-gray-400">{t("auth.welcome_subtitle")}</p>
        </div>

        {error && (
          <div className="mb-4 flex items-center gap-2 rounded-xl border border-red-500/50 bg-red-500/10 p-3 text-sm text-red-400">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-1">
            <label className="ml-1 text-xs font-bold text-gray-400">{t("auth.username")}</label>
            <div className="relative">
              <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                <User size={18} />
              </div>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full rounded-xl border border-white/10 bg-white/5 py-3 pl-11 pr-4 text-white transition-colors focus:border-lucky-gold focus:outline-none"
                placeholder="User123"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="ml-1 text-xs font-bold text-gray-400">{t("auth.password")}</label>
            <div className="relative">
              <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                <Lock size={18} />
              </div>
              <input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-xl border border-white/10 bg-white/5 py-3 pl-11 pr-12 text-white transition-colors focus:border-lucky-gold focus:outline-none"
                placeholder="********"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white"
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>

          <div className="flex justify-end">
            <a href="#" className="text-xs text-lucky-gold hover:underline">
              {t("auth.forgot_password")}
            </a>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-lucky-gold to-orange-500 py-4 font-bold text-lucky-dark shadow-lg shadow-lucky-gold/20 transition-all hover:scale-[1.02] disabled:cursor-not-allowed disabled:opacity-70"
          >
            {isLoading ? (
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-lucky-dark border-t-transparent"></div>
            ) : (
              <>
                {t("auth.submit")} <ArrowRight size={20} />
              </>
            )}
          </button>
        </form>

        <div className="mt-8 text-center">
          <p className="text-sm text-gray-400">
            {t("auth.no_account")}{" "}
            <button onClick={onGoToRegister} className="font-bold text-lucky-gold hover:underline">
              {t("auth.create")}
            </button>
          </p>
        </div>
      </div>
    </div>
  )
}

export default LoginPage
