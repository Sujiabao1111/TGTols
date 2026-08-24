"use client"
import type React from "react"
import { useState, useEffect, useRef } from "react"
import { Flame, MonitorPlay, Fish, Trophy, ChevronDown, Loader2 } from "lucide-react"
import { type Game, type ServerGame, GameCategory } from "../app/mocks/types"
import { SAMPLE_GAMES, PROVIDERS } from "../app/mocks/games"
import GameCard from "./GameCard"
import { useLanguage } from "../app/contexts/LanguageContext"

interface GameGridProps {
  initialCategory: GameCategory
  onGameSelect: (game: Game | ServerGame) => void
}

const PAGE_SIZE = 8

const GameGrid: React.FC<GameGridProps> = ({ initialCategory, onGameSelect }) => {
  const { t } = useLanguage()
  const [activeCategory, setActiveCategory] = useState<GameCategory>(initialCategory)
  const [activeProvider, setActiveProvider] = useState<string>("All")
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [showProviderDropdown, setShowProviderDropdown] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setVisibleCount(PAGE_SIZE)
    setActiveCategory(initialCategory)
  }, [initialCategory])

  useEffect(() => {
    setVisibleCount(PAGE_SIZE)
  }, [activeCategory, activeProvider])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowProviderDropdown(false)
      }
    }

    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const categories = [
    { label: t("cat.slots"), value: GameCategory.SLOTS, icon: <Flame size={16} /> },
    { label: t("cat.live"), value: GameCategory.LIVE, icon: <MonitorPlay size={16} /> },
    { label: t("cat.fishing"), value: GameCategory.FISHING, icon: <Fish size={16} /> },
    { label: t("cat.sports"), value: GameCategory.SPORTS, icon: <Trophy size={16} /> },
  ]

  const allCategories = [
    { label: t("btn.all_games"), value: GameCategory.ALL, icon: <Trophy size={16} /> },
    ...categories,
  ]

  const filteredGames = SAMPLE_GAMES.filter((g) => {
    const categoryMatch = activeCategory === GameCategory.ALL || g.category === activeCategory
    const providerMatch = activeProvider === "All" || g.provider === activeProvider
    return categoryMatch && providerMatch
  })

  const displayGames = filteredGames.slice(0, visibleCount)
  const hasMore = visibleCount < filteredGames.length

  const handleLoadMore = () => {
    setIsLoadingMore(true)
    setTimeout(() => {
      setVisibleCount((prev) => prev + PAGE_SIZE)
      setIsLoadingMore(false)
    }, 800)
  }

  useEffect(() => {
    const handleScroll = () => {
      if (window.innerHeight + document.documentElement.scrollTop >= document.documentElement.offsetHeight - 100) {
        if (hasMore && !isLoadingMore) {
          handleLoadMore()
        }
      }
    }

    window.addEventListener("scroll", handleScroll)
    return () => window.removeEventListener("scroll", handleScroll)
  }, [hasMore, isLoadingMore])

  return (
    <div className="min-h-screen bg-lucky-dark animate-fade-in pb-24">
      <div className="sticky top-16 z-30 bg-lucky-dark/95 backdrop-blur-md border-b border-white/10 shadow-lg pt-4 pb-2 px-4">
        <div className="max-w-7xl mx-auto flex flex-col gap-4">
          <div className="flex justify-between items-center gap-4">
            <div className="flex overflow-x-auto gap-2 scrollbar-hide w-full">
              {allCategories.map((cat) => (
                <button
                  key={cat.value}
                  onClick={() => setActiveCategory(cat.value)}
                  className={`flex items-center gap-1.5 px-4 py-2 rounded-full font-bold whitespace-nowrap text-sm transition-all ${
                    activeCategory === cat.value
                      ? "bg-lucky-gold text-lucky-dark shadow-lg shadow-lucky-gold/20"
                      : "bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white"
                  }`}
                >
                  {cat.icon}
                  {cat.label}
                </button>
              ))}
            </div>
            <div ref={dropdownRef} className="hidden md:block relative min-w-[150px]">
              <button
                onClick={() => setShowProviderDropdown(!showProviderDropdown)}
                className="w-full px-4 py-2 rounded-lg bg-zinc-900 border border-zinc-700 text-white hover:border-lucky-gold/50 transition-colors flex items-center justify-between gap-2 whitespace-nowrap"
              >
                <span className="text-sm font-medium">
                  {activeProvider === "All" ? t("filter.all_providers") : activeProvider}
                </span>
                <ChevronDown size={16} className={`transition-transform ${showProviderDropdown ? "rotate-180" : ""}`} />
              </button>
              {showProviderDropdown && (
                <div className="absolute top-full mt-2 w-full min-w-[180px] bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl z-40 max-h-60 overflow-y-auto">
                  {PROVIDERS.map((provider) => (
                    <button
                      key={provider}
                      onClick={() => {
                        setActiveProvider(provider)
                        setShowProviderDropdown(false)
                      }}
                      className={`w-full px-4 py-2 text-left text-sm hover:bg-zinc-800 transition-colors ${
                        activeProvider === provider ? "text-lucky-gold font-bold" : "text-gray-300"
                      }`}
                    >
                      {provider === "All" ? t("filter.all_providers") : provider}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
          <div ref={dropdownRef} className="md:hidden relative w-full">
            <button
              onClick={() => setShowProviderDropdown(!showProviderDropdown)}
              className="w-full px-4 py-3 rounded-xl bg-zinc-900 border border-zinc-700 text-white hover:border-lucky-gold/50 transition-colors flex items-center justify-between gap-2"
            >
              <span className="text-sm font-medium">
                {activeProvider === "All" ? t("filter.all_providers") : `${t("filter.provider")}: ${activeProvider}`}
              </span>
              <ChevronDown size={16} className={`transition-transform ${showProviderDropdown ? "rotate-180" : ""}`} />
            </button>
            {showProviderDropdown && (
              <div className="absolute top-full mt-2 w-full bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl z-40 max-h-60 overflow-y-auto">
                {PROVIDERS.map((provider) => (
                  <button
                    key={provider}
                    onClick={() => {
                      setActiveProvider(provider)
                      setShowProviderDropdown(false)
                    }}
                    className={`w-full px-4 py-2 text-left text-sm hover:bg-zinc-800 transition-colors ${
                      activeProvider === provider ? "text-lucky-gold font-bold" : "text-gray-300"
                    }`}
                  >
                    {provider === "All" ? t("filter.all_providers") : `${t("filter.provider")}: ${provider}`}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {filteredGames.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-gray-500">
            <Trophy size={48} className="mb-4 opacity-50" />
            <p>{t("msg.no_games")}</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-3 md:gap-6">
            {displayGames.map((game) => (
              <GameCard key={game.id} game={game} onSelect={onGameSelect} />
            ))}
          </div>
        )}
        <div className="py-8 flex justify-center">
          {isLoadingMore && (
            <div className="flex items-center gap-2 text-lucky-gold">
              <Loader2 className="animate-spin" size={24} />
              <span className="font-bold text-sm">{t("grid.loading")}</span>
            </div>
          )}
          {!hasMore && filteredGames.length > 0 && (
            <div className="text-gray-500 text-xs uppercase tracking-widest font-bold">{t("grid.end")}</div>
          )}
        </div>
      </div>
    </div>
  )
}

export default GameGrid
