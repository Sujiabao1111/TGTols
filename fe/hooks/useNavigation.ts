"use client"

import { useRouter, usePathname } from "next/navigation"
import { View } from "@/app/mocks/types"

export const useNavigation = () => {
  const router = useRouter()
  const pathname = usePathname()

  // Detect if we're using new routing (URL-based) or old routing (state-based)
  const isNewRouting = pathname !== "/" || typeof window === "undefined"

  const navigate = (view: View) => {
    if (isNewRouting) {
      // Use Next.js router for new routing
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
    } else {
      // Fallback to old state-based navigation (handled by parent)
      console.warn("Old routing detected, navigation handled by parent")
    }
  }

  const getCurrentView = (): View => {
    const viewMap: Record<string, View> = {
      "/home": "home",
      "/wallet": "wallet",
      "/profile": "profile",
      "/invite": "invite",
      "/activity": "activity",
      "/vip": "vip",
      "/favorites": "favorites",
      "/game": "game",
      "/login": "login",
      "/register": "register",
      "/help-center": "help-center",
      "/fairness-policy": "fairness-policy",
      "/privacy-policy": "privacy-policy",
      "/contact-us": "contact-us",
      "/adddesktop": "adddesktop",
      "/bettingRank": "bettingRank",
      "/weeklySpinWheel": "weeklySpinWheel",
      "/dailyWeeklyChallenge": "dailyWeeklyChallenge",
      "/sevenDayTopup": "sevenDayTopup",
      "/newUserRecharge": "newUserRecharge",
    }

    // Check for dynamic routes
    if (pathname.startsWith("/game/")) return "game"

    return viewMap[pathname] || "home"
  }

  return {
    navigate,
    getCurrentView,
    isNewRouting,
    pathname,
  }
}
