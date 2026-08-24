"use client"

import { useRouter } from "next/navigation"
import LoginScreen from "@/components/LoginScreen"
import { useAuth } from "@/app/contexts/AuthContext"

export default function LoginPage() {
  const router = useRouter()
  const { login } = useAuth()

  const handleLogin = () => {
    login()
    router.push("/home")
  }

  const handleGoToRegister = () => {
    router.push("/register")
  }

  return <LoginScreen onLogin={handleLogin} onGoToRegister={handleGoToRegister} />
}
