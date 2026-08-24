import type { BannerResponse, HeroSlide, Language } from "@/app/mocks/types"
import { apiUrl } from "@/lib/api-base-url"
import { getLaunchSource } from "@/lib/desktop-launch"

interface LoginResponse {
  success: boolean
  token: string
  message?: string
}

interface RegisterResponse {
  success: boolean
  token: string
}

export interface PlayerTransactionSummaryView {
  currency: string
  count: number
  turnover: number
  bet: number
  win: number
  winlose: number
  jp_share: number
  jp_win: number
  turnover_u: number
  bet_u: number
  win_u: number
  winlose_u: number
  jp_share_u: number
  jp_win_u: number
}

export interface PlayerTransactionOverviewBlock {
  period_type: "daily" | "weekly" | "total"
  period_key: string
  from_date: string
  to_date: string
  summary: PlayerTransactionSummaryView
}

export interface PlayerTransactionOverview {
  currency: string
  login_id: string
  daily: PlayerTransactionOverviewBlock
  weekly: PlayerTransactionOverviewBlock
  total: PlayerTransactionOverviewBlock
  refreshed_at: string
}

export interface VipMonthlyBonusStatusResponse {
  activity_type: string
  current_vip_level: number
  claim_month: string
  reward_amount: number
  can_claim: boolean
  claimed: boolean
  claimed_at?: string
  balance: number
}

export interface ActivityClaimRecordResponse {
  success: boolean
  activity_type: string
  claimed: boolean
  already_claimed: boolean
  claimed_at: string
  reward_amount: number
  balance: number
  message: string
}

export interface UserCoupon {
  id: number
  user_id: number
  coupon_code: string
  activity_type: string
  coupon_value: number
  min_deposit: number
  status: number
  valid_start: string
  valid_end: string
  activated_at?: string
  used_at?: string
}

interface RecycleBalanceApiResponse {
  code: number
  message: string
  data?: {
    recycled_amount?: number
    current_balance?: number
    game_balance?: number
    platform_code?: string
    summary_pending?: boolean
    summary_error?: string
    provider_unavailable?: boolean
    provider_error_code?: string
    insurance?: {
      triggered: boolean
      entry_amount_u: number
      exit_amount_u: number
      loss_amount_u: number
      compensation_amount_u: number
      current_balance: number
      coupon_code?: string
      message?: string
    } | null
  }
}

export interface GameLaunchErrorReportPayload {
  game_code?: string
  game_name?: string
  provider_code?: string
  provider_name?: string
  is_lobby?: boolean
  is_mobile?: boolean
  language?: string
  error_type: "launch_request_failed" | "iframe_load_timeout"
  error_message: string
  page_url?: string
  game_url?: string
}

const AUTO_RECYCLE_LAST_RUN_KEY = "auto-recycle:last-run-at"
const AUTO_RECYCLE_DEFAULT_COOLDOWN_MS = 30 * 1000
const LAST_GAME_PLATFORM_KEY = "game:last-platform-code"
export const GAME_TRANSACTIONS_UPDATED_EVENT = "game-transactions-updated"
const recycleBalanceInFlightByPlatform = new Map<string, Promise<RecycleBalanceApiResponse>>()

const normalizePlatformCode = (platformCode?: string) => (platformCode || "HEDOC").trim().toUpperCase()
const autoRecycleLastRunKey = (platformCode: string) => `${AUTO_RECYCLE_LAST_RUN_KEY}:${normalizePlatformCode(platformCode)}`

export const gamePlatformState = {
  markActive(platformCode?: string) {
    if (typeof window === "undefined") {
      return
    }
    window.localStorage.setItem(LAST_GAME_PLATFORM_KEY, normalizePlatformCode(platformCode))
  },

  getLast() {
    if (typeof window === "undefined") {
      return "HEDOC"
    }
    return normalizePlatformCode(window.localStorage.getItem(LAST_GAME_PLATFORM_KEY) || "HEDOC")
  },

  clearIfCurrent(platformCode?: string) {
    if (typeof window === "undefined") {
      return
    }
    const normalizedPlatformCode = normalizePlatformCode(platformCode)
    if (this.getLast() === normalizedPlatformCode) {
      window.localStorage.removeItem(LAST_GAME_PLATFORM_KEY)
    }
  },
}

export const notifyGameTransactionsUpdated = () => {
  if (typeof window === "undefined") {
    return
  }
  window.dispatchEvent(new Event(GAME_TRANSACTIONS_UPDATED_EVENT))
}

export const authService = {
  async recordDomainClick(): Promise<void> {
    if (typeof window === "undefined") return
    const day = new Date().toLocaleDateString("en-CA")
    const key = `domain-visit:${window.location.hostname}:${day}`
    if (window.localStorage.getItem(key)) return
    try {
      const response = await fetch(apiUrl("/api/domain-click"), { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" }, body: new URLSearchParams({ domain: window.location.hostname }), keepalive: true })
      if (response.ok) window.localStorage.setItem(key, "1")
    } catch { /* analytics must not affect page access */ }
  },
  async login(username: string, password: string): Promise<LoginResponse> {
    try {
      const response = await fetch(apiUrl("/api/login"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username,
          password,
          site_domain: typeof window !== "undefined" ? window.location.hostname : "",
        }),
      })

      if (!response.ok) {
        const rawText = await response.text()
        let detail = rawText.trim()

        if (detail) {
          try {
            const parsed = JSON.parse(detail) as { message?: string; error?: string }
            detail = parsed.message || parsed.error || detail
          } catch {
            // Keep raw response text when it is not JSON.
          }
        }

        if (response.status === 401) {
          throw new Error("AUTH_INVALID_CREDENTIALS")
        }
        if (response.status === 403) {
          throw new Error("AUTH_ACCOUNT_DISABLED")
        }
        if (response.status === 400) {
          throw new Error("AUTH_MISSING_CREDENTIALS")
        }

        throw new Error(detail ? `HTTP ${response.status}: ${detail}` : `HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Login request failed:", error)
      throw error
    }
  },

  async register(username: string, password: string, email?: string, inviteCode?: string): Promise<RegisterResponse> {
    try {
      const body: Record<string, unknown> = { username, password }
      if (typeof window !== "undefined") {
        body.site_domain = window.location.hostname
      }
      if (email) {
        body.email = email
      }
      if (inviteCode) {
        body.invite_code = inviteCode
      }

      const response = await fetch(apiUrl("/api/reg"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Registration request failed:", error)
      throw error
    }
  },

  logout() {
    localStorage.removeItem("token")
    // Dispatch a custom event so App.tsx can react to it
    window.dispatchEvent(new Event("auth-unauthorized"))
  },

  getToken() {
    if (typeof window !== "undefined") {
      return localStorage.getItem("token")
    }
    return null
  },

  isAuthenticated() {
    return !!this.getToken()
  },
}

// Wrapper for authenticated requests
export const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
  const token = authService.getToken()

  const headers = {
    "Content-Type": "application/json",
    ...options.headers,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  } as HeadersInit

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    })

    // Handle token expiration or forbidden access
    if (response.status === 403 || response.status === 401) {
      authService.logout()
      return response // Return response so caller can handle it if needed, but app will redirect
    }

    return response
  } catch (error) {
    throw error
  }
}

const buildHttpError = async (response: Response) => {
  const rawText = await response.text()
  let detail = rawText.trim()

  if (detail) {
    try {
      const parsed = JSON.parse(detail) as { message?: string; error?: string }
      detail = parsed.message || parsed.error || detail
    } catch {
      // Keep raw response text when it's not JSON.
    }
  }

  return new Error(detail ? `HTTP ${response.status}: ${detail}` : `HTTP error! status: ${response.status}`)
}

export const gameService = {
  async getGameOptions(platformCode = "HEDOC") {
    try {
      const response = await fetchWithAuth(apiUrl("/games/options"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ platform_code: platformCode }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get game options failed:", error)
      throw error
    }
  },

  async getGameList(page: number, pageSize: number, gameTypeId: string, gameProvider?: string, platformCode = "HEDOC") {
    try {
      const body: Record<string, unknown> = {
        page,
        page_size: pageSize,
        platform_code: platformCode,
        game_type_id: gameTypeId,
      }

      if (gameProvider) {
        body.game_provider = gameProvider
      }

      const response = await fetchWithAuth(apiUrl("/games/list"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get game list failed:", error)
      throw error
    }
  },

  async getGameLobby(gameTypeId: string, providerCode: string, isMobile = false, lang = "en", platformCode = "HEDOC") {
    try {
      const body = {
        game_type_id: gameTypeId,
        platform_code: platformCode,
        provider_code: providerCode,
        is_mobile: isMobile,
        language: lang,
        launch_source: getLaunchSource(),
      }

      const response = await fetchWithAuth(apiUrl("/games/lobby2"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get game lobby failed:", error)
      throw error
    }
  },

  async launchGame(gameCode: string, isMobile = false, lang = "en", gameId?: number | string, platformCode = "HEDOC") {
    try {
      const body = {
        game_id: gameId,
        platform_code: platformCode,
        game_code: gameCode,
        is_mobile: isMobile,
        language: lang,
        launch_source: getLaunchSource(),
      }

      const response = await fetchWithAuth(apiUrl("/games/launch2"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Launch game failed:", error)
      throw error
    }
  },
  async reportLaunchError(payload: GameLaunchErrorReportPayload) {
    try {
      const response = await fetchWithAuth(apiUrl("/games/launch-error/report"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      return await response.json()
    } catch (error) {
      console.error("Report launch error failed:", error)
      throw error
    }
  },
  async recycleBalance(platformCode = "HEDOC", options?: { forceSummarySync?: boolean }): Promise<RecycleBalanceApiResponse> {
    const normalizedPlatformCode = normalizePlatformCode(platformCode)
    try {
      const response = await fetchWithAuth(apiUrl("/games/recyle"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          platform_code: normalizedPlatformCode,
          force_summary_sync: options?.forceSummarySync ?? false,
        }),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Recycle balance failed:", error)
      throw error
    }
  },
  async autoRecycleBalance(options?: { force?: boolean; cooldownMs?: number; platformCode?: string }): Promise<RecycleBalanceApiResponse | null> {
    const force = options?.force ?? false
    const cooldownMs = options?.cooldownMs ?? AUTO_RECYCLE_DEFAULT_COOLDOWN_MS
    const platformCode = normalizePlatformCode(options?.platformCode)

    if (typeof window !== "undefined" && !force) {
      const lastRunRaw = window.localStorage.getItem(autoRecycleLastRunKey(platformCode))
      const lastRunAt = lastRunRaw ? Number.parseInt(lastRunRaw, 10) : 0
      if (lastRunAt > 0 && Date.now() - lastRunAt < cooldownMs) {
        return null
      }
    }

    const inFlight = recycleBalanceInFlightByPlatform.get(platformCode)
    if (inFlight) {
      return inFlight
    }

    const recyclePromise = this.recycleBalance(platformCode)
      .then((result) => {
        if (typeof window !== "undefined") {
          window.localStorage.setItem(autoRecycleLastRunKey(platformCode), Date.now().toString())
        }
        return result
      })
      .finally(() => {
        recycleBalanceInFlightByPlatform.delete(platformCode)
      })

    recycleBalanceInFlightByPlatform.set(platformCode, recyclePromise)
    return recyclePromise
  },
  async refreshPlayerTransactionsAfterPendingSync(options?: { attempts?: number; delayMs?: number }) {
    const attempts = options?.attempts ?? 6
    const delayMs = options?.delayMs ?? 5000

    for (let index = 0; index < attempts; index += 1) {
      await new Promise((resolve) => window.setTimeout(resolve, delayMs))
      try {
        await activityService.getPlayerTransactionOverview(false)
        notifyGameTransactionsUpdated()
      } catch (error) {
        console.error("Refresh player transactions after pending sync failed:", error)
      }
    }
  },
  async getUserInfo() {
    try {
      const response = await fetchWithAuth(apiUrl("/games/info"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get user info failed:", error)
      throw error
    }
  },
  async getInviteInfo() {
    try {
      const response = await fetchWithAuth(apiUrl("/games/inviteinfo"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get invite info failed:", error)
      throw error
    }
  },
  async getBanners(language: Language = "en"): Promise<BannerResponse> {
    try {
      const isIndonesian = language === "id"
      const banners: HeroSlide[] = [
        {
          id: 1,
          title: "",
          image: isIndonesian
            ? "/images/banners/addDesktopBanner-yinni.jpg"
            : "/images/banners/addDesktopBanner-yingyu.jpg",
          color: "from-lucky-gold to-orange-500",
          jumpLink: "/adddesktop",
        },
        {
          id: 2,
          title: "",
          image: isIndonesian
            ? "/images/activity5/banners_yinni.jpg"
            : "/images/activity5/banners_yingyu.jpg",
          color: "from-lucky-gold to-orange-500",
          jumpLink: "/sevenDayTopup",
        },
        {
          id: 3,
          title: "",
          image: isIndonesian
            ? "/images/activity2/banners_yinni.jpg"
            : "/images/activity2/banners_yingyu.jpg",
          color: "from-lucky-gold to-orange-500",
          jumpLink: "/weeklySpinWheel",
        },
        {
          id: 4,
          title: "",
          image: isIndonesian
            ? "/images/activity4/banners_yinni.jpg"
            : "/images/activity4/banners_yingyu.jpg",
          color: "from-lucky-gold to-orange-500",
          jumpLink: "/dailyWeeklyChallenge",
        },
        {
          id: 5,
          title: "",
          image: isIndonesian
            ? "/images/banners/bettingRankBanner_yinni.jpg"
            : "/images/banners/bettingRankBanner_yingyu.jpg",
          color: "from-lucky-gold to-orange-500",
          jumpLink: "/bettingRank",
        },
        {
          id: 6,
          title: "",
          image: isIndonesian
            ? "/images/activity1/banners_yinni.jpg"
            : "/images/activity1/banners_yingyu.jpg",
          color: "from-lucky-gold to-orange-500",
          jumpLink: '/newUserRecharge',
        },
      ]

      return {
        code: 0,
        data: banners,
        message: "ok",
      }
    } catch (error) {
      console.error("Get banners failed:", error)
      throw error
    }
  },
  async getFavoriteGames(page = 1, pageSize = 20) {
    try {
      // Added pagination parameters to request body
      const response = await fetchWithAuth(apiUrl("/games/favlist"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          page,
          page_size: pageSize,
        }),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get favorite games failed:", error)
      throw error
    }
  },
  async doFavorite(gameId: number | string, isFavorite: boolean) {
    try {
      const response = await fetchWithAuth(apiUrl("/games/dofav"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          game_id: gameId,
          is_favorite: isFavorite,
        }),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Toggle favorite failed:", error)
      throw error
    }
  },
  async getSlotGames(page = 1, pageSize = 20, sortType: "POPULAR" | "RECOMMEND" | "NEW" = "POPULAR") {
    try {
      const response = await fetchWithAuth(apiUrl("/games/reorderlist"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          page,
          page_size: pageSize,
          sort_type: sortType,
        }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get slot games failed:", error)
      throw error
    }
  },
}

// 活动相关 API
export async function getDailyWeeklyChallengeStatus(currency?: string) {
  try {
    const params = new URLSearchParams()
    if (currency) {
      params.set("currency", currency)
    }

    const response = await fetchWithAuth(apiUrl(`/activities/daily-weekly-challenge${params.toString() ? `?${params.toString()}` : ""}`), {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    })

    if (!response.ok) {
      throw await buildHttpError(response)
    }

    return await response.json()
  } catch (error) {
    console.error("Get daily weekly challenge status failed:", error)
    throw error
  }
}

export async function claimDailyWeeklyChallenge(cycleType: string, taskIndex: number) {
  try {
    const response = await fetchWithAuth(apiUrl("/activities/daily-weekly-challenge/claim"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        cycle_type: cycleType,
        task_index: taskIndex,
      }),
    })

    if (!response.ok) {
      throw await buildHttpError(response)
    }

    return await response.json()
  } catch (error) {
    console.error("Claim daily weekly challenge failed:", error)
    throw error
  }
}

export const activityService = {
  // 储值返利活动
  async getRechargeRebateStatus(currency?: string) {
    try {
      const params = new URLSearchParams()
      if (currency) {
        params.set("currency", currency)
      }

      const response = await fetchWithAuth(apiUrl(`/activities/recharge-rebate${params.toString() ? `?${params.toString()}` : ""}`), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get recharge rebate status failed:", error)
      throw error
    }
  },

  async claimRechargeRebate(dayNumber: number) {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/recharge-rebate/claim"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ day_number: dayNumber }),
      })

      const data = await response.json()

      if (!response.ok) {
        // Extract error message from backend response
        const errorMessage = data.error || data.message || `HTTP error! status: ${response.status}`
        throw new Error(errorMessage)
      }

      return data
    } catch (error) {
      console.error("Claim recharge rebate failed:", error)
      throw error
    }
  },

  // 轮盘活动
  async getWheelStatus() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/wheel"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get wheel status failed:", error)
      throw error
    }
  },

  async spinWheel() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/wheel/spin"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Spin wheel failed:", error)
      throw error
    }
  },

  // 输返活动
  async getLossRebateStatus() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/loss-rebate"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get loss rebate status failed:", error)
      throw error
    }
  },

  async claimLossRebate() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/loss-rebate/claim"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Claim loss rebate failed:", error)
      throw error
    }
  },

  async getAddDesktopStatus() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/add-desktop"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error("Get add desktop status failed:", error)
      throw error
    }
  },

  async claimAddDesktop() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/add-desktop/claim"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })

      const data = await response.json()
      if (!response.ok) {
        const errorMessage = data.error || data.message || `HTTP error! status: ${response.status}`
        throw new Error(errorMessage)
      }

      return data
    } catch (error) {
      console.error("Claim add desktop failed:", error)
      throw error
    }
  },

  async getBettingRankValue() {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/betting-rank/value"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error("Get betting rank value failed:", error)
      throw error
    }
  },

  async getWeeklySpinWheelStatus(currency?: string) {
    try {
      const params = new URLSearchParams()
      if (currency) {
        params.set("currency", currency)
      }

      const response = await fetchWithAuth(apiUrl(`/activities/weekly-spin-wheel${params.toString() ? `?${params.toString()}` : ""}`), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      return await response.json()
    } catch (error) {
      console.error("Get weekly spin wheel status failed:", error)
      throw error
    }
  },

  async getDailyWeeklyChallengeStatus(currency?: string) {
    return await getDailyWeeklyChallengeStatus(currency)
  },

  async claimDailyWeeklyChallenge(cycleType: string, taskIndex: number) {
    return await claimDailyWeeklyChallenge(cycleType, taskIndex)
  },

  async getPlayerTransactionOverview(sync = false, currency?: string): Promise<PlayerTransactionOverview> {
    try {
      const params = new URLSearchParams()
      if (sync) {
        params.set("sync", "1")
      }
      if (currency) {
        params.set("currency", currency)
      }
      const query = params.toString() ? `?${params.toString()}` : ""
      const response = await fetchWithAuth(apiUrl(`/games/transactions/overview${query}`), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const payload = await response.json()
      return payload.data as PlayerTransactionOverview
    } catch (error) {
      console.error("Get player transaction overview failed:", error)
      throw error
    }
  },

  async syncWeeklySpinWheelSource(date?: string) {
    try {
      if (!date) {
        return await this.getPlayerTransactionOverview(true)
      }

      const response = await fetchWithAuth(apiUrl("/games/transactions/sync"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ date }),
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      const payload = await response.json()
      return payload.data
    } catch (error) {
      console.error("Sync weekly spin wheel source failed:", error)
      throw error
    }
  },

  async getSevenDayTopupStatus(currency: string) {
    try {
      const response = await fetchWithAuth(
        apiUrl(`/activities/seven-day-topup?currency=${encodeURIComponent(currency)}`),
        {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
          },
        },
      )

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      return await response.json()
    } catch (error) {
      console.error("Get seven day topup status failed:", error)
      throw error
    }
  },

  async getNewUserRechargeStatus(currency?: string) {
    try {
      const params = new URLSearchParams()
      if (currency) {
        params.set("currency", currency)
      }

      const response = await fetchWithAuth(apiUrl(`/activities/new-user-recharge${params.toString() ? `?${params.toString()}` : ""}`), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error("Get new user recharge status failed:", error)
      throw error
    }
  },

  async getVipMonthlyBonusStatus(): Promise<VipMonthlyBonusStatusResponse> {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/vip-monthly-bonus"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw await buildHttpError(response)
      }

      return await response.json()
    } catch (error) {
      console.error("Get VIP monthly bonus status failed:", error)
      throw error
    }
  },

  async claimVipMonthlyBonus(): Promise<ActivityClaimRecordResponse> {
    try {
      const response = await fetchWithAuth(apiUrl("/activities/vip-monthly-bonus/claim"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })

      const data = await response.json()
      if (!response.ok) {
        const errorMessage = data.error || data.message || `HTTP error! status: ${response.status}`
        throw new Error(errorMessage)
      }

      return data as ActivityClaimRecordResponse
    } catch (error) {
      console.error("Claim VIP monthly bonus failed:", error)
      throw error
    }
  },
}

// 优惠券相关 API
export const couponService = {
  async getUserCoupons(status = -1): Promise<UserCoupon[]> {
    try {
      const url = status >= 0
        ? apiUrl(`/coupons?status=${status}`)
        : apiUrl("/coupons")

      const response = await fetchWithAuth(url, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data as UserCoupon[]
    } catch (error) {
      console.error("Get user coupons failed:", error)
      throw error
    }
  },

  async activateCoupon(couponCode: string, depositAmount: number) {
    try {
      const response = await fetchWithAuth(apiUrl("/coupons/activate"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          coupon_code: couponCode,
          deposit_amount: depositAmount,
        }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Activate coupon failed:", error)
      throw error
    }
  },

  async getCouponStats() {
    try {
      const response = await fetchWithAuth(apiUrl("/coupons/stats"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error("Get coupon stats failed:", error)
      throw error
    }
  },
}
