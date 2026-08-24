"use client"

import { createContext, useContext, useState, useEffect, ReactNode, useCallback, useRef } from "react"
import { authService, gameService } from "../../services/api"
import { useAuth } from "./AuthContext"

interface UserProfile {
  uid: number
  username: string
  inviteCode: string
  vipLevel: number
  level: number
}

interface UserContextType {
  userBalance: number
  withdrawableBalance: number
  remainingWager: number
  depositWagerMultiplier: number
  rewardWagerMultiplier: number
  totalDeposit: number
  userProfile: UserProfile
  isUserInfoLoaded: boolean
  favoriteGameIds: Set<number>
  fetchUserBalance: () => Promise<void>
  fetchFavoriteGameIds: () => Promise<void>
  updateBalance: (newBalance: number, nextWithdrawableBalance?: number) => void
  addFavorite: (gameId: number) => void
  removeFavorite: (gameId: number) => void
}

const UserContext = createContext<UserContextType | undefined>(undefined)

const defaultUserProfile: UserProfile = {
  uid: 0,
  username: "LuckyPlayer",
  inviteCode: "",
  vipLevel: 0,
  level: 0,
}

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const { isLoggedIn } = useAuth()
  const [userBalance, setUserBalance] = useState(0)
  const [withdrawableBalance, setWithdrawableBalance] = useState(0)
  const [remainingWager, setRemainingWager] = useState(0)
  const [depositWagerMultiplier, setDepositWagerMultiplier] = useState(2)
  const [rewardWagerMultiplier, setRewardWagerMultiplier] = useState(20)
  const [totalDeposit, setTotalDeposit] = useState(0)
  const [userProfile, setUserProfile] = useState<UserProfile>(defaultUserProfile)
  const [isUserInfoLoaded, setIsUserInfoLoaded] = useState(false)
  const [favoriteGameIds, setFavoriteGameIds] = useState<Set<number>>(new Set())
  const userInfoRequestRef = useRef<Promise<void> | null>(null)

  const fetchUserBalance = useCallback(async () => {
    if (!isLoggedIn) {
      return
    }

    if (userInfoRequestRef.current) {
      return userInfoRequestRef.current
    }

    const request = (async () => {
      try {
        const response = await gameService.getUserInfo()
        if (response.code === 0) {
          setUserBalance(response.data.balance)
          setWithdrawableBalance(response.data.withdrawable_balance ?? response.data.balance)
          setRemainingWager(response.data.remaining_wager ?? 0)
          setDepositWagerMultiplier(response.data.deposit_wager_multiplier ?? 2)
          setRewardWagerMultiplier(response.data.reward_wager_multiplier ?? 20)
          setTotalDeposit(response.data.total_deposit ?? 0)
          setUserProfile({
            uid: response.data.uid ?? 0,
            username: response.data.username ?? defaultUserProfile.username,
            inviteCode: response.data.invite_code ?? "",
            vipLevel: response.data.vip_level ?? 0,
            level: response.data.level ?? 0,
          })
          setIsUserInfoLoaded(true)
        }
      } catch (error) {
        console.error("Failed to fetch user balance:", error)
        if (
          error instanceof Error &&
          (error.message.includes("HTTP 401") ||
            error.message.includes("HTTP 403") ||
            error.message.includes("HTTP 404") ||
            error.message.includes("Unauthorized") ||
            error.message.includes("User not found"))
        ) {
          authService.logout()
        }
      } finally {
        setIsUserInfoLoaded(true)
        userInfoRequestRef.current = null
      }
    })()

    userInfoRequestRef.current = request
    return request
  }, [isLoggedIn])

  const fetchFavoriteGameIds = useCallback(async () => {
    if (!isLoggedIn) return
    try {
      const response = await gameService.getFavoriteGames(1, 100)
      if (response.code === 0 && response.data?.list) {
        const ids = new Set<number>(response.data.list.map((game: { id: number }) => game.id))
        setFavoriteGameIds(ids)
      }
    } catch (error) {
      console.error("Failed to fetch favorite game IDs:", error)
    }
  }, [isLoggedIn])

  const updateBalance = (newBalance: number, nextWithdrawableBalance?: number) => {
    setUserBalance(newBalance)
    setWithdrawableBalance((prev) => {
      if (typeof nextWithdrawableBalance === "number") {
        return nextWithdrawableBalance
      }

      return Math.min(prev, newBalance)
    })
  }

  const addFavorite = (gameId: number) => {
    setFavoriteGameIds((prev) => new Set([...prev, gameId]))
  }

  const removeFavorite = (gameId: number) => {
    setFavoriteGameIds((prev) => {
      const newSet = new Set(prev)
      newSet.delete(gameId)
      return newSet
    })
  }

  // Listen for balance update events
  useEffect(() => {
    const handleFavoritesUpdate = () => {
      fetchFavoriteGameIds()
    }
    const handleWindowFocus = () => {
      fetchUserBalance()
    }
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        fetchUserBalance()
      }
    }

    window.addEventListener("favorites-updated", handleFavoritesUpdate)
    window.addEventListener("focus", handleWindowFocus)
    document.addEventListener("visibilitychange", handleVisibilityChange)

    return () => {
      window.removeEventListener("favorites-updated", handleFavoritesUpdate)
      window.removeEventListener("focus", handleWindowFocus)
      document.removeEventListener("visibilitychange", handleVisibilityChange)
    }
  }, [fetchUserBalance, fetchFavoriteGameIds])

  // Fetch data when user logs in
  useEffect(() => {
    if (isLoggedIn) {
      const syncUserData = async () => {
        await Promise.all([fetchUserBalance(), fetchFavoriteGameIds()])
      }

      void syncUserData()
    } else {
      queueMicrotask(() => {
        setUserBalance(0)
        setWithdrawableBalance(0)
        setRemainingWager(0)
        setDepositWagerMultiplier(2)
        setRewardWagerMultiplier(20)
        setTotalDeposit(0)
        setUserProfile(defaultUserProfile)
        setIsUserInfoLoaded(false)
        setFavoriteGameIds(new Set())
      })
    }
  }, [isLoggedIn, fetchUserBalance, fetchFavoriteGameIds])

  return (
    <UserContext.Provider
      value={{
        userBalance,
        withdrawableBalance,
        remainingWager,
        depositWagerMultiplier,
        rewardWagerMultiplier,
        totalDeposit,
        userProfile,
        isUserInfoLoaded,
        favoriteGameIds,
        fetchUserBalance,
        fetchFavoriteGameIds,
        updateBalance,
        addFavorite,
        removeFavorite,
      }}
    >
      {children}
    </UserContext.Provider>
  )
}

export const useUser = () => {
  const context = useContext(UserContext)
  if (!context) {
    throw new Error("useUser must be used within UserProvider")
  }
  return context
}
