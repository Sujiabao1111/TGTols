"use client"

import { ReactNode, useEffect, useState } from "react"
import { usePathname } from "next/navigation"
import Layout from "@/components/Layout"
import { useNavigation } from "@/hooks/useNavigation"
import { useAuth } from "@/app/contexts/AuthContext"
import { useUser } from "@/app/contexts/UserContext"
import { useGame } from "@/app/contexts/GameContext"

// 需要登录才能访问的路由前缀
const PROTECTED_PATHS = ['/activity', '/wallet', '/profile', '/favorites', '/game/', '/invite']

function isProtectedPath(path: string): boolean {
  return PROTECTED_PATHS.some(prefix => path.startsWith(prefix))
}

export default function PublicLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const { navigate, getCurrentView } = useNavigation()
  const { isLoggedIn } = useAuth()
  const { userBalance, fetchUserBalance } = useUser()
  const { gameTypes, selectedGameType, sortType, setSelectedGameType, setSortType, clearFilters } = useGame()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  // 如果当前路径是 protected 路由，不渲染 Layout（让 ProtectedLayout 处理）
  if (isProtectedPath(pathname)) {
    return null
  }

  // 避免 SSR/水合不匹配，等待客户端挂载
  if (!mounted) {
    return <div className="min-h-screen bg-lucky-dark" />
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
