"use client"

import type React from "react"
import { useState } from "react"
import Image from "next/image"
import { User, Lock, ArrowRight, Eye, EyeOff, AlertCircle, Mail } from "lucide-react"
import { useLanguage } from "../app/contexts/LanguageContext"
import { authService } from "../services/api"
import { clearStoredInviteCode, resolveInviteCode } from "@/lib/invite-code"

interface RegisterScreenProps {
  onRegister: () => void
  onBackToLogin: () => void
  inviteCode?: string // Added inviteCode prop
}

const RegisterScreen: React.FC<RegisterScreenProps> = ({ onRegister, onBackToLogin, inviteCode }) => {
  const { t } = useLanguage()
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    // Validation
    if (!username || !password || !confirmPassword) {
      setError(t("auth.register_missing_fields"))
      return
    }

    if (password !== confirmPassword) {
      setError(t("auth.register_password_mismatch"))
      return
    }

    if (password.length < 6) {
      setError(t("auth.register_password_min"))
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const data = await authService.register(username, password, email, resolveInviteCode(inviteCode))

      if (data.success && data.token) {
        // Save token
        localStorage.setItem("token", data.token)
        clearStoredInviteCode()
        // Trigger app state update
        onRegister()
      } else {
        setError(t("auth.register_failed"))
      }
    } catch (err) {
      console.error(err)
      setError(t("auth.connection_failed"))
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[url('https://images.unsplash.com/photo-1511193311914-0346f16efe90?q=80&w=2073&auto=format&fit=crop')] bg-cover bg-center flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-lucky-dark/80 backdrop-blur-sm"></div>

      <div className="relative w-full max-w-md bg-lucky-dark/90 border border-white/10 rounded-3xl p-8 shadow-2xl animate-fade-in">
        {/* Logo Area */}
        <div className="flex flex-col items-center mb-8">
          <div className="relative mb-3 h-16 w-16">
            <Image
              src="/images/PPLogo.png"
              alt="PP Logo"
              fill
              sizes="64px"
              className="object-contain"
            />
          </div>
          <h1 className="text-3xl font-display font-bold text-white">{t("auth.create")}</h1>
          <p className="text-gray-400 text-sm mt-1">{t("auth.register_subtitle")}</p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/50 rounded-xl p-3 mb-4 flex items-center gap-2 text-red-400 text-sm">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-1">
            <label className="text-xs text-gray-400 font-bold ml-1">{t("auth.username")} *</label>
            <div className="relative">
              <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                <User size={18} />
              </div>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl py-3 pl-11 pr-4 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                placeholder={t("auth.username_placeholder")}
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs text-gray-400 font-bold ml-1">{t("auth.email_optional")}</label>
            <div className="relative">
              <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                <Mail size={18} />
              </div>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl py-3 pl-11 pr-4 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                placeholder="your@email.com"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs text-gray-400 font-bold ml-1">{t("auth.password")} *</label>
            <div className="relative">
              <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                <Lock size={18} />
              </div>
              <input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl py-3 pl-11 pr-12 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                placeholder="••••••••"
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

          <div className="space-y-1">
            <label className="text-xs text-gray-400 font-bold ml-1">{t("auth.confirm_password")} *</label>
            <div className="relative">
              <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                <Lock size={18} />
              </div>
              <input
                type={showConfirmPassword ? "text" : "password"}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl py-3 pl-11 pr-12 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                placeholder="••••••••"
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white"
              >
                {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-gradient-to-r from-lucky-gold to-orange-500 text-lucky-dark font-bold py-4 rounded-xl shadow-lg shadow-lucky-gold/20 hover:scale-[1.02] transition-all flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed"
          >
            {isLoading ? (
              <div className="w-5 h-5 border-2 border-lucky-dark border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <>
                {t("auth.create")} <ArrowRight size={20} />
              </>
            )}
          </button>
        </form>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-gray-400 text-sm">
            {t("auth.already_have_account")}{" "}
            <button onClick={onBackToLogin} className="text-lucky-gold font-bold hover:underline">
              {t("auth.sign_in")}
            </button>
          </p>
        </div>
      </div>
    </div>
  )
}

export default RegisterScreen
