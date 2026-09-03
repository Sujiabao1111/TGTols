"use client"

import { ReactNode, useEffect, useState } from "react"
import { usePathname, useRouter } from "next/navigation"
import Layout from "@/components/Layout"
import { useNavigation } from "@/hooks/useNavigation"
import { useAuth } from "@/app/contexts/AuthContext"
import { useUser } from "@/app/contexts/UserContext"
import { useGame } from "@/app/contexts/GameContext"

export default function ProtectedLayout({ children }: { children: ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const { isLoggedIn, isAuthReady, checkAuth } = useAuth()
  const { navigate, getCurrentView } = useNavigation()
  const { userBalance, fetchUserBalance } = useUser()
  const { gameTypes, selectedGameType, sortType, setSelectedGameType, setSortType, clearFilters } = useGame()
  const [mounted, setMounted] = useState(false)
  const isPublicActivityPath = pathname === "/newUserRecharge"

  useEffect(() => {
    queueMicrotask(() => {
      setMounted(true)
    })

    if (isPublicActivityPath) {
      return
    }

    // Client-side auth check
    if (typeof window !== "undefined") {
      const hasAuth = checkAuth()
      if (isAuthReady && !hasAuth && !isLoggedIn) {
        router.push("/login")
      }
    }
  }, [isLoggedIn, isAuthReady, checkAuth, isPublicActivityPath, router])

  // 等待客户端挂载，避免水合不匹配
  if (!mounted) {
    return <div className="min-h-screen bg-lucky-dark" />
  }

  if (!isAuthReady || (!isLoggedIn && !isPublicActivityPath)) {
    return null // or loading spinner
  }

  return (
    <Layout
      currentView={getCurrentView()}
      onNavigate={navigate}
      isLoggedIn={isLoggedIn}
      userBalance={userBalance}
      onBalanceUpdate={fetchUserBalance}
      gameTypes={gameTypes}
      onGameTypeSelect={setSelectedGameType}
      onClearFilters={clearFilters}
      sortType={sortType}
      onSortTypeSelect={setSortType}
      externalGameType={selectedGameType}
    >
      {children}
    </Layout>
  )
}
