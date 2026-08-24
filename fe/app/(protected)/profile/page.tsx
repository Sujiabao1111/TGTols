"use client"

import { useRouter } from "next/navigation"
import ProfilePage from "@/components/ProfilePage"
import { useAuth } from "@/app/contexts/AuthContext"
import { View, WalletTab } from "@/app/mocks/types"

export default function ProfileRoute() {
  const router = useRouter()
  const { logout } = useAuth()

  const handleNavigate = (view: View) => {
    const routeMap: Record<View, string> = {
      home: "/home",
      wallet: "/wallet",
      profile: "/profile",
      invite: "/invite",
      activity: "/activity",
      vip: "/vip",
      favorites: "/favorites",
      game: "/game",
      "all-games": "/home",
      login: "/login",
      register: "/register",
      "help-center": "/help-center",
      "fairness-policy": "/fairness-policy",
      "privacy-policy": "/privacy-policy",
      "contact-us": "/contact-us",
      "404": "/404",
      adddesktop: "/adddesktop",
      bettingRank: "/bettingRank",
      weeklySpinWheel: "/weeklySpinWheel",
      dailyWeeklyChallenge: "/dailyWeeklyChallenge",
      sevenDayTopup: "/sevenDayTopup",
      newUserRecharge: "/newUserRecharge",

    }
    router.push(routeMap[view] || "/home")
  }

  const handleSetWalletTab = (tab: WalletTab) => {
    router.push(`/wallet?tab=${tab}`)
  }

  const handleLogout = () => {
    logout()
    router.push("/login")
  }

  return (
    <ProfilePage
      onNavigate={handleNavigate}
      onSetWalletTab={handleSetWalletTab}
      onLogout={handleLogout}
    />
  )
}
