"use client"

import type React from "react"
import Image from "next/image"
import { createPortal } from "react-dom"
import { useState, useMemo, useEffect, useRef } from "react"
import Hero from "./Hero"
import Ticker from "./Ticker"
import GameCard from "./GameCard"
import type { Game, ServerGame, GameType, Provider } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { useUser } from "../app/contexts/UserContext"
import { Search, ChevronLeft, ChevronRight } from "lucide-react"
import { gamePlatformState, gameService } from "../services/api"
import { cn } from "../lib/utils"
import { showAddDesktopInsuranceDialog } from "./AddDesktopInsuranceDialog"
import { useNewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"
import { useNavigation } from "@/hooks/useNavigation"
import { consumeInviteActivityPopupPending } from "@/lib/invite-code"

interface HomePageProps {
  onGameSelect: (game: Game | ServerGame) => void
  externalGameType?: GameType | null
  onFilterChange?: (gameType: GameType | null, provider: Provider | null) => void
  sortType?: "POPULAR" | "RECOMMEND" | "NEW"
  showHero?: boolean
}

type CategoryMeta = {
  defaultImage: string
  activeImage: string
}

const CATEGORY_META: Record<string, CategoryMeta> = {
  SLOT: {
    defaultImage: "/images/logo/SLOT.png",
    activeImage: "/images/logo/SLOT-1.png",
  },
  CASINO: {
    defaultImage: "/images/logo/CASINO.png",
    activeImage: "/images/logo/CASINO-1.png",
  },
  SPORTBOOK: {
    defaultImage: "/images/logo/SPORTBOOK.png",
    activeImage: "/images/logo/SPORTBOOK-1.png",
  },
}

const GAME_PLATFORMS = ["HEDOC", "M7", "M7PP"] as const
const HOME_ACTIVITY_POPUP_SHOWN_KEY = "home-activity-popup-shown"
const LAST_PROVIDER_STORAGE_KEY = "home-last-provider-by-game-type"
const PAGE_LOAD_ID = `${Date.now()}-${Math.random().toString(36).slice(2)}`
const GAME_LIST_CACHE_TTL_MS = 5 * 60 * 1000
const RECOMMENDED_GAMES_STORAGE_KEY = "home-recommended-games"

interface GameListCacheEntry {
  games: ServerGame[]
  hasMore: boolean
  cachedAt: number
}

const gameListCache = new Map<string, GameListCacheEntry>()

const providerPlatform = (provider?: Provider | null) => provider?.platform_code || "HEDOC"

const providerDisplayName = (provider: Provider) => provider.name

const providerKey = (provider?: Provider | null) =>
  provider ? `${providerPlatform(provider).toUpperCase()}-${provider.id}-${provider.code}` : ""

const isSameProvider = (a?: Provider | null, b?: Provider | null) => Boolean(a && b && providerKey(a) === providerKey(b))

const getLastProviderKey = (gameTypeCode: string) => {
  try {
    const storedProviders = JSON.parse(window.localStorage.getItem(LAST_PROVIDER_STORAGE_KEY) || "{}")
    return typeof storedProviders?.[gameTypeCode] === "string" ? storedProviders[gameTypeCode] : ""
  } catch {
    return ""
  }
}

const saveLastProviderKey = (gameTypeCode: string, selectedProviderKey: string) => {
  try {
    const storedProviders = JSON.parse(window.localStorage.getItem(LAST_PROVIDER_STORAGE_KEY) || "{}")
    window.localStorage.setItem(
      LAST_PROVIDER_STORAGE_KEY,
      JSON.stringify({ ...storedProviders, [gameTypeCode]: selectedProviderKey }),
    )
  } catch {
    // Continue with the in-memory selection when storage is unavailable.
  }
}

const normalizeProviderLogoKey = (value = "") => value.trim().toLowerCase().replace(/[\s_+.-]/g, "")

const PROVIDER_LOGO_FILES: Record<string, Record<string, string>> = {
  HEDOC: {
    ae: "AE Sexy.png",
    aesexy: "AE Sexy.png",
    evolution: "Evolution Gaming.png",
    evolutiongaming: "Evolution Gaming.png",
    m8: "M8 Sports.png",
    m8sports: "M8 Sports.png",
    netent: "NetEnt.png",
    nolimit: "NoLimitCity.png",
    nolimitcity: "NoLimitCity.png",
    playtech: "Playtech Slots.png",
    playtechcasino: "Playtech Casino.png",
    playtechslots: "Playtech Slots.png",
    pp: "Pragmatic Play Slots.png",
    pragmatic: "Pragmatic Play Slots.png",
    pragmaticplay: "Pragmatic Play Slots.png",
    pragmaticplaycasino: "Pragmatic Play Casino.png",
    pragmaticplayslots: "Pragmatic Play Slots.png",
    redtiger: "RedTiger.png",
    sbo: "SBO Sports.png",
    sbosports: "SBO Sports.png",
    uu: "UU Slots.png",
    uuslots: "UU Slots.png",
  },
  M7: {
    evolution: "Evolution.png",
    evolutiongaming: "Evolution.png",
    hacksaw: "Hacksaw.png",
    jdb: "JDB.png",
    microgaming: "Microgaming+.png",
    pg: "PG.png",
    pgsoft: "PG.png",
    pgslot: "PG.png",
    playtech: "Playtech.png",
    pp: "PP.png",
    pp\u771f\u4eba: "PP\u771f\u4eba.png",
    pragmatic: "PP.png",
    pragmaticplay: "PP.png",
    slotmill: "Slotmill.png",
  },
  M7PP: {
    pp: "PP.png",
    pp真人: "PP真人.png",
    pragmatic: "PP.png",
    pragmaticplay: "PP.png",
  },
}

const PROVIDER_CATEGORY_LOGO_FILES: Record<string, Record<string, Record<string, string>>> = {
  HEDOC: {
    CASINO: {
      playtech: "Playtech Casino.png",
      pp: "Pragmatic Play Casino.png",
      pragmatic: "Pragmatic Play Casino.png",
      pragmaticplay: "Pragmatic Play Casino.png",
    },
    SLOT: {
      playtech: "Playtech Slots.png",
      pp: "Pragmatic Play Slots.png",
      pragmatic: "Pragmatic Play Slots.png",
      pragmaticplay: "Pragmatic Play Slots.png",
    },
  },
  M7: {
    CASINO: {
      pp: "PP\u771f\u4eba.png",
      pragmatic: "PP\u771f\u4eba.png",
      pragmaticplay: "PP\u771f\u4eba.png",
    },
  },
  M7PP: {
    CASINO: {
      pp: "PP真人.png",
      pp真人: "PP真人.png",
      pragmatic: "PP真人.png",
      pragmaticplay: "PP真人.png",
    },
  },
}

const providerLogoSrc = (provider: Provider, gameTypeCode?: string) => {
  const platform = providerPlatform(provider).toUpperCase()
  const assetPlatform = platform === "M7PP" ? "M7" : platform
  const normalizedGameType = gameTypeCode?.toUpperCase() || ""
  const keys = [provider.name, provider.code].map(normalizeProviderLogoKey)
  const categoryLookup = PROVIDER_CATEGORY_LOGO_FILES[platform]?.[normalizedGameType] || {}
  const platformLookup = PROVIDER_LOGO_FILES[platform] || {}
  const fileName =
    keys.map((key) => categoryLookup[key]).find(Boolean) ||
    keys.map((key) => platformLookup[key]).find(Boolean) ||
    `${providerDisplayName(provider)}.png`

  return `/images/logo/${assetPlatform}/${encodeURIComponent(fileName)}`
}

interface ProviderBrandCardProps {
  provider: Provider
  gameTypeCode?: string
  active: boolean
  onSelect: (provider: Provider) => void
}

const ProviderBrandCard: React.FC<ProviderBrandCardProps> = ({ provider, gameTypeCode, active, onSelect }) => {
  const [imageFailed, setImageFailed] = useState(false)
  const name = providerDisplayName(provider)

  return (
    <button
      type="button"
      onClick={() => onSelect(provider)}
      aria-pressed={active}
      className={cn(
        "group relative h-[62px] w-[29.44vw] shrink-0 overflow-hidden rounded-lg border bg-zinc-950 text-left shadow-lg transition-all duration-200 sm:h-[72px] sm:w-[143px] md:h-[82px] md:w-[164px]",
        active
          ? "border-lucky-gold shadow-lucky-gold/25"
          : "border-white/10 hover:border-lucky-gold/60 hover:shadow-lucky-gold/15",
      )}
    >
      {!imageFailed ? (
        <Image
          src={providerLogoSrc(provider, gameTypeCode)}
          alt={name}
          fill
          sizes="(min-width: 768px) 164px, (min-width: 640px) 143px, 29.44vw"
          onError={() => setImageFailed(true)}
          className="pointer-events-none object-cover transition-transform duration-300 group-hover:scale-105"
        />
      ) : (
        <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-zinc-900 via-black to-zinc-950 px-4">
          <span className="text-center text-base font-bold text-white sm:text-lg">{name}</span>
        </div>
      )}
      <span className={cn("pointer-events-none absolute inset-0 bg-black/0 transition-colors", active && "bg-black/15")} />
      <span className="sr-only">{name}</span>
    </button>
  )
}

const isPgProvider = (provider?: Provider | null) => {
  if (!provider) return false

  const code = provider.code.trim().toUpperCase()
  const name = provider.name.trim().toUpperCase().replace(/\s+/g, "")

  return code === "PG" || code === "PG_SLOT" || code === "PGSOFT" || name === "PG" || name === "PGSOFT"
}

const providerPriority = (provider?: Provider | null) => {
  const platform = providerPlatform(provider).toUpperCase()
  const isM7Provider = platform === "M7"
  if (platform === "M7PP") return 0
  if (isM7Provider && isPgProvider(provider)) return 1
  if (isM7Provider) return 2
  if (isPgProvider(provider)) return 3
  return 4
}

const gameProviderPriority = (provider?: Provider | null) => {
  const isM7Provider = providerPlatform(provider).toUpperCase() === "M7"
  if (isM7Provider && isPgProvider(provider)) return 0
  if (isPgProvider(provider)) return 1
  if (isM7Provider) return 2
  return 3
}

const sortByProviderPriority = <T,>(
  items: T[],
  getProvider: (item: T) => Provider | null | undefined,
  getPriority = providerPriority,
) =>
  [...items]
    .map((item, index) => ({ item, index }))
    .sort((a, b) => {
      const providerDiff = getPriority(getProvider(a.item)) - getPriority(getProvider(b.item))
      return providerDiff || a.index - b.index
    })
    .map(({ item }) => item)

const HomePage: React.FC<HomePageProps> = ({
  onGameSelect,
  externalGameType,
  onFilterChange,
  sortType,
  showHero = true,
}) => {
  const [gameTypes, setGameTypes] = useState<GameType[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [activeGameType, setActiveGameType] = useState<GameType | null>(null)
  const [activeProvider, setActiveProvider] = useState<Provider | null>(null)
  const [searchQuery, setSearchQuery] = useState("")
  const [games, setGames] = useState<ServerGame[]>([])
  const [recommendedGames, setRecommendedGames] = useState<ServerGame[]>([])
  const [showRecommendations, setShowRecommendations] = useState(true)
  const [lobbyUrl, setLobbyUrl] = useState("")
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [isBrandsDragging, setIsBrandsDragging] = useState(false)
  const brandsScrollRef = useRef<HTMLDivElement>(null)
  const brandDragStartXRef = useRef(0)
  const brandDragStartScrollLeftRef = useRef(0)
  const brandDragMovedRef = useRef(false)
  const gameListRequestIdRef = useRef(0)
  const { language, t } = useLanguage() // Declare the language variable
  const { updateBalance } = useUser()
  const { navigate } = useNavigation()
  const { status: newUserRechargeStatus, loading: loadingNewUserRechargeStatus } = useNewUserRechargeStatus(true)
  const [isPromoPopupOpen, setIsPromoPopupOpen] = useState(false)
  const [shouldShowInvitePromoPopup, setShouldShowInvitePromoPopup] = useState(false)
  const isIndonesian = language === "id"
  const shouldShowNewUserRechargePopup = Boolean(newUserRechargeStatus && newUserRechargeStatus.hidden === false)
  const shouldShowPromoPopup = shouldShowNewUserRechargePopup || shouldShowInvitePromoPopup

  useEffect(() => {
    try {
      const stored = JSON.parse(window.localStorage.getItem(RECOMMENDED_GAMES_STORAGE_KEY) || "[]")
      if (Array.isArray(stored)) queueMicrotask(() => setRecommendedGames(stored))
    } catch {
      queueMicrotask(() => setRecommendedGames([]))
    }
  }, [])

  const handleRecommendedGameSelect = (game: ServerGame) => {
    setRecommendedGames((current) => {
      const key = String(game.id)
      const next = [game, ...current.filter((item) => String(item.id) !== key)].slice(0, 10)
      try { window.localStorage.setItem(RECOMMENDED_GAMES_STORAGE_KEY, JSON.stringify(next)) } catch {}
      return next
    })
    onGameSelect(game)
  }

  const handleGameSelect = (game: Game | ServerGame) => {
    if ("game_code" in game) handleRecommendedGameSelect(game)
    else onGameSelect(game)
  }

  useEffect(() => {
    if (!activeGameType || activeGameType.code !== "SLOT" || providers.length === 0) return
    if (recommendedGames.length > 0) return
    let cancelled = false
    const seedRecommendations = async () => {
      const candidates = await Promise.all(
        providers.filter((provider) => provider.type_id === activeGameType.id).map(async (provider) => {
          try {
            const response = await gameService.getGameList(1, 20, "SLOT", provider.code, providerPlatform(provider))
            const list = response.code === 0 && response.data?.list ? response.data.list : []
            return list.length ? list[Math.floor(Math.random() * list.length)] : null
          } catch { return null }
        }),
      )
      if (cancelled) return
      const seeded = candidates.filter(Boolean) as ServerGame[]
      setRecommendedGames(seeded.slice(0, 10))
      try { window.localStorage.setItem(RECOMMENDED_GAMES_STORAGE_KEY, JSON.stringify(seeded.slice(0, 10))) } catch {}
    }
    seedRecommendations()
    return () => { cancelled = true }
  }, [activeGameType, providers, recommendedGames.length])

  useEffect(() => {
    queueMicrotask(() => {
      setShouldShowInvitePromoPopup(consumeInviteActivityPopupPending())
    })
  }, [])

  useEffect(() => {
    if ((!shouldShowInvitePromoPopup && loadingNewUserRechargeStatus) || !shouldShowPromoPopup) {
      return
    }

    const currentLoadId = PAGE_LOAD_ID
    const shownLoadId = window.sessionStorage.getItem(HOME_ACTIVITY_POPUP_SHOWN_KEY)
    if (shownLoadId === currentLoadId) {
      return
    }

    const frameId = window.requestAnimationFrame(() => {
      setIsPromoPopupOpen(true)
      window.sessionStorage.setItem(HOME_ACTIVITY_POPUP_SHOWN_KEY, currentLoadId)
    })

    return () => window.cancelAnimationFrame(frameId)
  }, [loadingNewUserRechargeStatus, shouldShowInvitePromoPopup, shouldShowPromoPopup])

  useEffect(() => {
    if (!isPromoPopupOpen) {
      return
    }

    const previousBodyOverflow = document.body.style.overflow
    const previousHtmlOverflow = document.documentElement.style.overflow

    document.body.style.overflow = "hidden"
    document.documentElement.style.overflow = "hidden"

    return () => {
      document.body.style.overflow = previousBodyOverflow
      document.documentElement.style.overflow = previousHtmlOverflow
    }
  }, [isPromoPopupOpen])

  const handlePromoPopupClick = () => {
    setIsPromoPopupOpen(false)
    navigate("newUserRecharge")
  }

  // Home 页加载后，自动同步三方余额（兜底机制）
  useEffect(() => {
    if (providers.length === 0) return

    const lastActivePlatform = gamePlatformState.getLast()
    const enabledPlatformCodes = Array.from(
      new Set(providers.map((provider) => providerPlatform(provider).toUpperCase())),
    ).filter((platformCode) => platformCode !== "M7PP" || lastActivePlatform === "M7PP")

    const syncBalance = async () => {
      try {
        for (const platformCode of enabledPlatformCodes) {
          const res = await gameService.autoRecycleBalance({ platformCode })
        // 如果回收了金额，刷新用户余额显示
          if (res?.data?.current_balance !== undefined) {
            updateBalance(res.data.current_balance)
          }
          if (res?.data?.insurance?.triggered) {
            const amount = `${(res.data.insurance.compensation_amount_u || 0).toFixed(2)}U`
            showAddDesktopInsuranceDialog(amount)
            continue
          // 可选：显示提示
          // console.log(`Balance synced: +$${res.data.recycled_amount}`)
          }
        }
      } catch (err) {
        // 静默失败，不打扰用户
        console.error("Balance sync failed:", err)
      }
    }

    // 延迟 2 秒执行，让页面先渲染完成
    const timer = setTimeout(syncBalance, 2000)

    return () => {
      clearTimeout(timer)
    }
  }, [providers, updateBalance])

  useEffect(() => {
    const fetchOptions = async () => {
      try {
        const results = await Promise.allSettled(
          GAME_PLATFORMS.map((platformCode) => gameService.getGameOptions(platformCode)),
        )
        const successResponses = results
          .filter((result) => result.status === "fulfilled")
          .map((result) => result.value)
          .filter((response) => response.code === 0 && response.data)

        if (successResponses.length > 0) {
          const mergedGameTypes = successResponses[0].data.game_types
          const mergedProviders = successResponses.flatMap((response) => response.data.providers)

          setGameTypes(mergedGameTypes)
          setProviders(mergedProviders)
          const slotType = mergedGameTypes.find((type: GameType) => type.code === "SLOT")
          if (slotType) {
            setActiveGameType((currentType) => currentType || slotType)
          }
        }
      } catch (error) {
        console.error("Failed to fetch game options:", error)
      }
    }
    fetchOptions()
  }, [])

  useEffect(() => {
    if (!externalGameType) {
      return
    }

    const isSameGameType = activeGameType?.id === externalGameType.id
    setActiveGameType(externalGameType)

    if (!isSameGameType) {
      setActiveProvider(null)
      setSearchQuery("")
      setPage(1)
      setHasMore(true)
    }
  }, [activeGameType?.id, externalGameType])

  // Restore the last selected entry for each game category. New users start on recommendations.
  useEffect(() => {
    if (!activeGameType) return
    const lastProvider = getLastProviderKey(activeGameType.code)
    setShowRecommendations(activeGameType.code === "SLOT" && !lastProvider)
  }, [activeGameType?.code])

  const handleGameTypeSelect = (type: GameType) => {
    setActiveGameType(type)
    setActiveProvider(null)
    setSearchQuery("")
    if (type.code !== "SLOT") setShowRecommendations(false)
    onFilterChange?.(type, null)
  }

  const handleProviderSelect = (provider: Provider) => {
    setShowRecommendations(false)
    setActiveProvider(provider)
    if (activeGameType) {
      saveLastProviderKey(activeGameType.code, providerKey(provider))
    }
    onFilterChange?.(activeGameType, provider)
  }

  const scrollBrands = (direction: "left" | "right") => {
    brandsScrollRef.current?.scrollBy({
      left: direction === "left" ? -320 : 320,
      behavior: "smooth",
    })
  }

  const handleBrandsMouseDown = (event: React.MouseEvent<HTMLDivElement>) => {
    if (event.button !== 0) return

    const target = event.currentTarget
    setIsBrandsDragging(true)
    brandDragMovedRef.current = false
    brandDragStartXRef.current = event.clientX
    brandDragStartScrollLeftRef.current = target.scrollLeft
  }

  const handleBrandsMouseMove = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!isBrandsDragging) return

    const deltaX = event.clientX - brandDragStartXRef.current
    if (Math.abs(deltaX) > 8) {
      brandDragMovedRef.current = true
      event.preventDefault()
    }
    event.currentTarget.scrollLeft = brandDragStartScrollLeftRef.current - deltaX
  }

  const handleBrandsMouseEnd = () => {
    setIsBrandsDragging(false)
  }

  const handleBrandsClickCapture = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!brandDragMovedRef.current) return

    event.preventDefault()
    event.stopPropagation()
    window.setTimeout(() => {
      brandDragMovedRef.current = false
    }, 0)
  }

  const availableProviders = useMemo(() => {
    if (!activeGameType) return []
    return sortByProviderPriority(
      providers.filter((p) => p.type_id === activeGameType.id),
      (provider) => provider,
    )
  }, [activeGameType, providers])

  const selectedProviderIsAvailable = Boolean(
    activeProvider && availableProviders.some((provider) => isSameProvider(provider, activeProvider)),
  )

  const effectiveProvider = useMemo(() => {
    if (selectedProviderIsAvailable) {
      return activeProvider
    }
    if (!activeGameType || availableProviders.length === 0) {
      return null
    }

    const lastProviderKey = getLastProviderKey(activeGameType.code)
    return availableProviders.find((provider) => providerKey(provider) === lastProviderKey) || availableProviders[0]
  }, [activeGameType, activeProvider, availableProviders, selectedProviderIsAvailable])

  useEffect(() => {
    if (!activeGameType || !effectiveProvider) {
      setActiveProvider(null)
      return
    }

    if (selectedProviderIsAvailable) {
      return
    }

    setActiveProvider(effectiveProvider)
    onFilterChange?.(activeGameType, effectiveProvider)
  }, [activeGameType, effectiveProvider, onFilterChange, selectedProviderIsAvailable])

  useEffect(() => {
    if (!effectiveProvider) return

    const frameId = window.requestAnimationFrame(() => {
      const container = brandsScrollRef.current
      const selectedBrand = container?.querySelector<HTMLElement>('[aria-pressed="true"]')
      if (!container || !selectedBrand) return

      const centeredScrollLeft = selectedBrand.offsetLeft - (container.clientWidth - selectedBrand.clientWidth) / 2
      container.scrollTo({ left: Math.max(0, centeredScrollLeft), behavior: "auto" })
    })

    return () => window.cancelAnimationFrame(frameId)
  }, [effectiveProvider])

  const providerGames = useMemo(() => {
    if (!activeGameType || activeGameType.code === "SLOT") return []

    return availableProviders.filter((provider) => {
      if (providerPlatform(provider) !== "HEDOC") return false
      return !effectiveProvider || isSameProvider(provider, effectiveProvider)
    }).flatMap((provider) => {
      const baseGame: ServerGame = {
        id: `${providerPlatform(provider)}-${provider.code}-lobby`,
        game_code: `${provider.code}-lobby`,
        platform_code: providerPlatform(provider),
        name: {
          EN: `${provider.name} Lobby`,
          CN: `${provider.name} 大厅`,
          ZH: `${provider.name} 大廳`,
          TH: `${provider.name} ล็อบบี้`,
        },
        img_url: `https://gg.ppnet55.com/img/providers/${provider.code}.png`,
        provider,
        is_lobby: true,
      }

      // const mockGames: ServerGame[] = Array.from({ length: 5 }, (_, index) => ({
      //   ...baseGame,
      //   id:`${provider.code}-lobby-${index}`,
      //   img_url: `https://gg.ppnet55.com/img/providers/mock/${provider.code}${index + 1}.webp`,
      // }))

      // return [baseGame, ...mockGames]
      return [baseGame]
    }) as ServerGame[]
  }, [activeGameType, availableProviders, effectiveProvider])

  useEffect(() => {
    const requestId = ++gameListRequestIdRef.current

    if (!activeGameType) {
      setGames([])
      setLobbyUrl("")
      setPage(1)
      setHasMore(true)
      return
    }

    if (activeGameType.code !== "SLOT" && activeGameType.code !== "SPORTBOOK" && activeGameType.code !== "CASINO") {
      setLoading(false)
      setGames([])
      setPage(1)
      setHasMore(false)
      return
    }

    const cacheKey = [
      activeGameType.code,
      effectiveProvider ? providerKey(effectiveProvider) : "ALL",
      sortType || "DEFAULT",
    ].join(":")
    const cached = gameListCache.get(cacheKey)

    if (cached && Date.now() - cached.cachedAt < GAME_LIST_CACHE_TTL_MS) {
      setGames(cached.games)
      setLobbyUrl("")
      setPage(1)
      setHasMore(cached.hasMore)
      setLoading(false)
      return
    }

    if (cached) {
      gameListCache.delete(cacheKey)
    }

    const fetchData = async () => {
      setLoading(true)
      setLobbyUrl("")
      setPage(1)
      setHasMore(true)

      try {
        if (sortType) {
          const response = await gameService.getSlotGames(1, 20, sortType)
          if (requestId === gameListRequestIdRef.current && response.code === 0 && response.data) {
            setGames(response.data.list)
            setHasMore(response.data.page < response.data.total_pages)
            gameListCache.set(cacheKey, {
              games: response.data.list,
              hasMore: response.data.page < response.data.total_pages,
              cachedAt: Date.now(),
            })
          }
        } else if (!effectiveProvider) {
          const results = await Promise.allSettled(
            GAME_PLATFORMS.map((platformCode) =>
              gameService.getGameList(1, 20, activeGameType.code, undefined, platformCode),
            ),
          )
          const successResponses = results
            .filter((result) => result.status === "fulfilled")
            .map((result) => result.value)
            .filter((response) => response.code === 0 && response.data)
          const mergedGames = successResponses.flatMap((response) => response.data.list)

          if (requestId === gameListRequestIdRef.current) {
            const hasMoreResults = successResponses.some((response) => response.data.page < response.data.total_pages)
            setGames(mergedGames)
            setHasMore(hasMoreResults)
            gameListCache.set(cacheKey, { games: mergedGames, hasMore: hasMoreResults, cachedAt: Date.now() })
          }
        } else {
          const response = await gameService.getGameList(
            1,
            20,
            activeGameType.code,
            effectiveProvider.code,
            effectiveProvider.platform_code || "HEDOC",
          )
          if (requestId === gameListRequestIdRef.current && response.code === 0 && response.data) {
            setGames(response.data.list)
            setHasMore(response.data.page < response.data.total_pages)
            gameListCache.set(cacheKey, {
              games: response.data.list,
              hasMore: response.data.page < response.data.total_pages,
              cachedAt: Date.now(),
            })
          }
        }
      } catch (error) {
        console.error("Failed to fetch games:", error)
      } finally {
        if (requestId === gameListRequestIdRef.current) {
          setLoading(false)
        }
      }
    }

    fetchData()
  }, [activeGameType, effectiveProvider, language, sortType])

  const loadMore = async () => {
    if (!activeGameType || loading || !hasMore || activeGameType.code !== "SLOT") return

    setLoading(true)
    try {
      const nextPage = page + 1
      if (activeGameType.code === "SLOT" && sortType) {
        const response = await gameService.getSlotGames(nextPage, 20, sortType)
        if (response.code === 0 && response.data) {
          setGames((prev) => [...prev, ...response.data.list])
          setPage(nextPage)
          setHasMore(response.data.page < response.data.total_pages)
        }
      } else if (!effectiveProvider) {
        const results = await Promise.allSettled(
          GAME_PLATFORMS.map((platformCode) =>
            gameService.getGameList(nextPage, 20, activeGameType.code, undefined, platformCode),
          ),
        )
        const successResponses = results
          .filter((result) => result.status === "fulfilled")
          .map((result) => result.value)
          .filter((response) => response.code === 0 && response.data)
        const mergedGames = successResponses.flatMap((response) => response.data.list)

        setGames((prev) => [...prev, ...mergedGames])
        setPage(nextPage)
        setHasMore(successResponses.some((response) => response.data.page < response.data.total_pages))
      } else {
        const response = await gameService.getGameList(
          nextPage,
          20,
          activeGameType.code,
          effectiveProvider.code,
          providerPlatform(effectiveProvider),
        )
        if (response.code === 0 && response.data) {
          setGames((prev) => [...prev, ...response.data.list])
          setPage(nextPage)
          setHasMore(response.data.page < response.data.total_pages)
        }
      }
    } catch (error) {
      console.error("Failed to load more games:", error)
    } finally {
      setLoading(false)
    }
  }

  const filteredGames = useMemo(() => {
    if (!searchQuery) return games
    const query = searchQuery.toLowerCase()
    return games.filter(
      (game) =>
        game.name.EN.toLowerCase().includes(query) ||
        game.name.CN.toLowerCase().includes(query) ||
        game.name.ZH.toLowerCase().includes(query),
    )
  }, [games, searchQuery])

  const additionalPlatformGames = filteredGames.filter((game) =>
    ["M7", "M7PP"].includes(providerPlatform(game.provider).toUpperCase()),
  )
  const prioritizedFilteredGames = effectiveProvider
    ? filteredGames
    : sortByProviderPriority(filteredGames, (game) => game.provider, gameProviderPriority)
  const displayGames =
    activeGameType?.code === "SLOT"
      ? prioritizedFilteredGames
      : sortByProviderPriority([...providerGames, ...additionalPlatformGames], (game) => game.provider, gameProviderPriority)

  const promoPopup =
    isPromoPopupOpen && shouldShowPromoPopup && typeof document !== "undefined"
      ? createPortal(
        <div className="fixed inset-0 z-[1000] flex items-center justify-center overflow-hidden bg-black/75 p-4 backdrop-blur-sm">
          <div className="flex w-full max-w-[420px] flex-col items-center justify-center">
            <button
              type="button"
              onClick={handlePromoPopupClick}
              className="block w-full overflow-hidden rounded-[28px] border border-white/10 bg-black shadow-[0_24px_80px_rgba(0,0,0,0.55)] transition-transform hover:scale-[1.01] active:scale-[0.99]"
            >
              <Image
                src={isIndonesian ? "/images/activity1/act_yinni.jpg" : "/images/activity1/act_yingyu.jpg"}
                alt={t("nav.newUserRecharge")}
                width={600}
                height={600}
                className="block h-auto w-full"
              />
            </button>

            <button
              type="button"
              onClick={() => setIsPromoPopupOpen(false)}
              className="mt-3 w-full rounded-2xl border border-white/15 bg-white/10 px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-white/15"
            >
              {t("activity.close")}
            </button>
          </div>
        </div>,
        document.body,
      )
      : null

  return (
    <div className="animate-fade-in pb-24">
      {promoPopup}
      {showHero && (
        <Hero
          newUserRechargeStatus={newUserRechargeStatus}
          loadingNewUserRechargeStatus={loadingNewUserRechargeStatus}
        />
      )}
      <Ticker />
      <div
        className="sticky z-30 bg-black/95 px-2 py-2 backdrop-blur-md sm:px-4"
        style={{ top: "calc(var(--app-header-height) + var(--app-header-clearance))" }}
      >
        <div className="mx-auto flex max-w-7xl flex-col gap-3">
          <div className="overflow-x-auto scrollbar-hide">
            {/* Category Filter Buttons */}
            <div className="grid min-w-full grid-cols-3 gap-2 [@media(orientation:landscape)]:gap-2">
              {/* <span className="text-xs text-gray-400 font-bold whitespace-nowrap min-w-fit">CATEGORY:</span> */}
              {gameTypes.map((type) => {
                const isActive = activeGameType?.id === type.id
                const meta = CATEGORY_META[type.code]
                if (!meta) return null
                return (
                  <button
                    key={type.id}
                    onClick={() => handleGameTypeSelect(type)}
                    aria-pressed={isActive}
                    className={cn(
                      "group relative min-w-0 overflow-hidden rounded-[14px] bg-transparent p-0 text-left transition-all duration-200",
                      "aspect-[1.42/1] [@media(orientation:landscape)]:aspect-[2.25/1]",
                      !isActive && "hover:opacity-95",
                    )}
                  >
                    <div className="relative h-full w-full">
                      <div className="absolute inset-0 rounded-[14px] transition-transform duration-200">
                        <Image
                          src={isActive ? meta.activeImage : meta.defaultImage}
                          alt={type.name}
                          fill
                          sizes="(orientation: landscape) 22vw, 30vw"
                          className="object-contain object-center"
                        />
                      </div>
                      <span className="sr-only">{type.name}</span>
                    </div>
                  </button>
                )
              })}
            </div>
          </div>

          {availableProviders.length > 0 && (
            <section className="space-y-2">
              <div className="flex items-center justify-between">
                <h2 className="px-1 text-lg font-bold text-white">Brands</h2>
                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => scrollBrands("left")}
                    className="flex h-8 w-8 items-center justify-center rounded-full text-gray-400 transition-colors hover:bg-white/10 hover:text-white"
                    aria-label="Scroll brands left"
                  >
                    <ChevronLeft size={24} />
                  </button>
                  <button
                    type="button"
                    onClick={() => scrollBrands("right")}
                    className="flex h-8 w-8 items-center justify-center rounded-full text-white transition-colors hover:bg-white/10"
                    aria-label="Scroll brands right"
                  >
                    <ChevronRight size={24} />
                  </button>
                </div>
              </div>
              <div
                ref={brandsScrollRef}
                onMouseDown={handleBrandsMouseDown}
                onMouseMove={handleBrandsMouseMove}
                onMouseUp={handleBrandsMouseEnd}
                onMouseLeave={handleBrandsMouseEnd}
                onClickCapture={handleBrandsClickCapture}
                className={cn(
                  "flex touch-pan-x select-none gap-3 overflow-x-auto pb-1 scrollbar-hide",
                  isBrandsDragging ? "cursor-grabbing" : "cursor-grab",
                )}
              >
                {activeGameType?.code === "SLOT" && <button
                  type="button"
                  onClick={() => {
                    setShowRecommendations(true)
                    setActiveProvider(null)
                    if (activeGameType) saveLastProviderKey(activeGameType.code, "")
                  }}
                  aria-pressed={showRecommendations}
                  className={cn(
                    "relative h-[62px] w-[29.44vw] shrink-0 overflow-hidden rounded-[10px] border bg-black/30 p-0 transition-all hover:opacity-90 sm:h-[72px] sm:w-[143px] md:h-[82px] md:w-[164px]",
                    showRecommendations ? "border-lucky-gold ring-2 ring-lucky-gold/60" : "border-white/10",
                  )}
                >
                  <Image src="/images/logo/tuijian.png" alt="推荐玩法" fill sizes="164px" className="object-cover" />
                  <span className="sr-only">推荐玩法</span>
                </button>}
                {availableProviders.map((provider) => (
                  <ProviderBrandCard
                    key={`${activeGameType?.code || "ALL"}-${providerKey(provider)}`}
                    provider={provider}
                    gameTypeCode={activeGameType?.code}
                    active={!showRecommendations && isSameProvider(effectiveProvider, provider)}
                    onSelect={handleProviderSelect}
                  />
                ))}
              </div>
            </section>
          )}

          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            {!showRecommendations && activeGameType?.code === "SLOT" && (
              <div className="relative w-full sm:flex-1 md:max-w-xs">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                <input
                  type="text"
                  placeholder="Search games..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full rounded-2xl border border-white/10 bg-black/35 py-2.5 pl-10 pr-4 text-white placeholder-gray-500 transition-colors focus:border-lucky-gold/50 focus:outline-none [@media(orientation:landscape)]:py-2"
                />
              </div>
            )}
          </div>
        </div>
      </div>
      <div className="max-w-7xl mx-auto px-4 py-6">
        {/* Loading State */}
        {loading && games.length === 0 && (
          <div className="text-center py-12">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-lucky-gold border-t-transparent"></div>
            <p className="text-gray-400 mt-4">Loading games...</p>
          </div>
        )}

        {showRecommendations && activeGameType?.code === "SLOT" && recommendedGames.length > 0 && (
          <div className="grid grid-cols-3 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
            {recommendedGames.map((game) => <GameCard key={game.id} game={game} onSelect={handleGameSelect} language={language.toUpperCase()} />)}
          </div>
        )}

        {!showRecommendations && !loading && displayGames.length > 0 && (
          <>
            <div className="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {displayGames.map((game) => (
                <GameCard key={game.id} game={game} onSelect={handleGameSelect} language={language.toUpperCase()} />
              ))}
            </div>

            {hasMore && activeGameType?.code === "SLOT" && (
              <div className="text-center mt-8">
                <button
                  onClick={loadMore}
                  disabled={loading}
                  className="px-8 py-3 rounded-full bg-gradient-to-r from-lucky-gold to-yellow-500 text-black font-bold hover:shadow-lg hover:shadow-lucky-gold/30 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? "Loading..." : "Load More"}
                </button>
              </div>
            )}
          </>
        )}

        {/* Empty State */}
        {!loading && !lobbyUrl && displayGames.length === 0 && activeGameType && (
          <div className="text-center py-12 text-gray-500">
            <p>{searchQuery ? "No games found matching your search" : "No games available"}</p>
          </div>
        )}

        {/* Initial State */}
        {!activeGameType && (
          <div className="text-center py-12 text-gray-400">
            <p>Select a game category to start</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default HomePage
