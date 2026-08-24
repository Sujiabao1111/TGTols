"use client"

import { useRouter, useSearchParams } from "next/navigation"
import RegisterScreen from "@/components/RegisterScreen"
import { useAuth } from "@/app/contexts/AuthContext"
import { Suspense } from "react"

// 内部组件使用 useSearchParams
function RegisterContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { login } = useAuth()
  const inviteCode = searchParams.get("code")

  const handleRegister = () => {
    login()
    router.push("/home")
  }

  const handleBackToLogin = () => {
    router.push("/login")
  }

  return (
    <RegisterScreen
      onRegister={handleRegister}
      onBackToLogin={handleBackToLogin}
      inviteCode={inviteCode || undefined}
    />
  )
}

// 主页面组件
export default function RegisterPage() {
  return (
    <Suspense fallback={<div className="flex items-center justify-center min-h-screen">Loading...</div>}>
      <RegisterContent />
    </Suspense>
  )
}
