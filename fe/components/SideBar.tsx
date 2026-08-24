"use client"

import type React from "react"
import { useState, useEffect, useRef } from "react"
import {
  Home,
  Gift,
  Gamepad2,
  ChevronLeft,
  ChevronRight,
  Globe,
  MessageSquare,
  ChevronDown,
} from "lucide-react"
import { useLanguage } from "../app/contexts/LanguageContext"
import type { View, Language, GameType } from "../app/mocks/types"
import { useAddDesktopStatus } from "@/hooks/useAddDesktopStatus"
import { useNewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"

interface SidebarProps {
  currentView: View
  onNavigate: (view: View) => void
  isLoggedIn: boolean
  onSidebarToggle?: (isExpanded: boolean) => void
  gameTypes?: GameType[]
  onGameTypeSelect?: (gameType: GameType) => void
  onClearFilters?: () => void
  onSortTypeSelect?: (sortType?: "POPULAR" | "RECOMMEND" | "NEW") => void
  sortType?: "POPULAR" | "RECOMMEND" | "NEW"
  activeFilter?: boolean
  externalGameType?: GameType | null
}

const Sidebar: React.FC<SidebarProps> = ({
  currentView,
  onNavigate,
  isLoggedIn,
  onSidebarToggle,
  onClearFilters,
  sortType,
  activeFilter = false,
}) => {
  const [isExpanded, setIsExpanded] = useState(true)
  const [isBonusOpen, setIsBonusOpen] = useState(true)
  const [showLangDropdown, setShowLangDropdown] = useState(false)
  const { language, setLanguage, t } = useLanguage()
  const langDropdownRef = useRef<HTMLDivElement>(null)
  const { status: addDesktopStatus, loading: addDesktopStatusLoading } = useAddDesktopStatus(isLoggedIn)
  const { status: newUserRechargeStatus, loading: newUserRechargeLoading } = useNewUserRechargeStatus(isLoggedIn)

  const toggleSidebar = () => {
    const newState = !isExpanded
    setIsExpanded(newState)
    onSidebarToggle?.(newState)
  }

  useEffect(() => {
    onSidebarToggle?.(isExpanded)
  }, [])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (langDropdownRef.current && !langDropdownRef.current.contains(event.target as Node)) {
        setShowLangDropdown(false)
      }
    }

    if (showLangDropdown) {
      document.addEventListener("mousedown", handleClickOutside)
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside)
    }
  }, [showLangDropdown])

  const languages: { code: Language; label: string; flag: string }[] = [
    { code: "en", label: "English", flag: "🇺🇸" },
    { code: "id", label: "Bahasa Indonesia", flag: "🇮🇩" },
  ]

  const currentLang = languages.find((l) => l.code === language) || languages[0]
  const activityViews: View[] = [
    "vip",
    "adddesktop",
    "bettingRank",
    "weeklySpinWheel",
    "dailyWeeklyChallenge",
    "sevenDayTopup",
    "newUserRecharge",
  ]
  const isActivityView = activityViews.includes(currentView)
  const showAddDesktopEntry = !isLoggedIn || addDesktopStatusLoading || !addDesktopStatus?.claimed
  const showNewUserRechargeEntry = !isLoggedIn || newUserRechargeLoading || !newUserRechargeStatus?.hidden
  const bonusItems: Array<{
    label: string
    tag?: "HOT" | "NEW"
    view: View
  }> = [
      ...(showAddDesktopEntry ? [{ label: t("nav.adddesktop"), tag: "HOT" as const, view: "adddesktop" as View }] : []),
      ...(showNewUserRechargeEntry ? [{ label: t("nav.newUserRecharge"), tag: "HOT" as const, view: "newUserRecharge" as View }] : []),
      { label: t("nav.sevenDayTopup"), tag: "HOT", view: "sevenDayTopup" },
      { label: t("nav.vip"), tag: "HOT", view: "vip" },
      { label: t("nav.weeklySpinWheel"), tag: "NEW", view: "weeklySpinWheel" },
      { label: t("nav.dailyWeeklyChallenge"), tag: "HOT", view: "dailyWeeklyChallenge" },
      { label: t("nav.bettingRank"), tag: "NEW", view: "bettingRank" },
    ]

  return (
    <div
      className={`fixed left-0 bottom-0 bg-black/95 border-r border-lucky-gold/20 shadow-2xl z-40 transition-all duration-300 flex flex-col ${isExpanded ? "w-64" : "w-16"
      }`}
      style={{ top: "calc(var(--app-header-height) + var(--app-header-clearance))" }}
    >
      {/* Header */}
      <div className="flex border-b border-lucky-gold/20 flex-shrink-0">
        {isExpanded ? (
          <div className="flex-1 border-b-2 border-lucky-gold bg-lucky-gold/10 py-3 text-center text-sm font-bold text-lucky-gold">
            CASINO
          </div>
        ) : (
          <div className="w-full py-3 flex justify-center">
            <div className="w-8 h-8 rounded-full bg-lucky-gold/20 flex items-center justify-center">
              <Gamepad2 size={18} className="text-lucky-gold" />
            </div>
          </div>
        )}
      </div>

      {/* Scrollable Content */}
      <div className="flex-1 overflow-y-auto overflow-x-hidden scrollbar-thin">
        <div className="p-2 space-y-4">
          {/* Main Navigation */}
          <div className="space-y-1">
            <NavItem
              icon={Home}
              label="Home"
              active={currentView === "home" && !sortType && !activeFilter}
              onClick={() => {
                onClearFilters?.()
                onNavigate("home")
              }}
              isExpanded={isExpanded}
            />
            {isExpanded ? (
              <div
                className={`rounded-lg transition-colors ${isBonusOpen || isActivityView
                    ? "border border-lucky-gold/30 bg-lucky-gold/10"
                    : "border border-white/10 bg-transparent"
                  }`}
              >
                <button
                  onClick={() => setIsBonusOpen((prev) => !prev)}
                  className={`flex w-full items-center justify-between gap-3 px-3 py-2.5 text-left transition-colors ${isActivityView ? "text-lucky-gold" : "text-gray-300 hover:text-white"
                    }`}
                >
                  <div className="flex items-center gap-3">
                    <Gift size={20} />
                    <span className="text-sm font-medium">Bonus</span>
                  </div>
                  <ChevronDown
                    size={16}
                    className={`text-gray-500 transition-transform ${isBonusOpen ? "rotate-180" : ""}`}
                  />
                </button>
                {isBonusOpen && (
                  <div className="space-y-1 px-3 pb-3">
                    {bonusItems.map((item) => (
                      <button
                        key={item.view}
                        onClick={() => onNavigate(item.view)}
                        className={`flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors ${currentView === item.view
                            ? "bg-gray-700/60 text-white"
                            : "text-gray-400 hover:bg-gray-700/50 hover:text-white"
                          }`}
                      >
                        <Gift size={14} className="shrink-0 text-lucky-gold" />
                        <span className="flex-1">{item.label}</span>
                        {item.tag && (
                          <span
                            className={`rounded px-1.5 py-0.5 text-xs font-bold text-white ${item.tag === "HOT" ? "bg-red-500" : "bg-green-500"
                              }`}
                          >
                            {item.tag}
                          </span>
                        )}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <NavItem
                icon={Gift}
                label="Bonus"
                active={isActivityView}
                onClick={() => setIsBonusOpen((prev) => !prev)}
                isExpanded={isExpanded}
              />
            )}
          </div>
        </div>
      </div>

      {/* Bottom Section - Social & Settings */}
      {isExpanded ? (
        <div className="flex-shrink-0 border-t border-lucky-gold/20 bg-black/95 p-3 space-y-3">
          {/* Social Media */}
          <div>
            <div className="text-xs text-gray-500 font-bold mb-2">SOCIAL</div>
            <div className="flex gap-2">
              <SocialButton
                icon="https://cdn.simpleicons.org/telegram/26A5E4"
                label="Telegram"
                onClick={() => window.open("https://t.me/", "_blank")}
              />
              <SocialButton
                icon="https://cdn.simpleicons.org/x/000000"
                label="Twitter"
                onClick={() => window.open("https://twitter.com/", "_blank")}
              />
              <SocialButton
                icon="https://cdn.simpleicons.org/instagram/E4405F"
                label="Instagram"
                onClick={() => window.open("https://instagram.com/", "_blank")}
              />
            </div>
          </div>

          {/* Language Selector */}
          <div>
            <div className="text-xs text-gray-500 font-bold mb-2">LANGUAGE</div>
            <div className="relative group">
              <button className="w-full flex items-center gap-2 px-3 py-2 rounded-lg bg-zinc-900 hover:bg-zinc-800 text-white text-sm transition-colors">
                <Globe size={16} />
                <span className="flex-1 text-left">{currentLang.label}</span>
                <ChevronDown size={14} />
              </button>
              <div className="absolute bottom-full left-0 right-0 mb-1 bg-zinc-900 border border-zinc-700 rounded-lg overflow-hidden opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
                {languages.map((lang) => (
                  <button
                    key={lang.code}
                    onClick={() => setLanguage(lang.code)}
                    className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-zinc-800 transition-colors ${language === lang.code ? "text-lucky-gold" : "text-gray-300"
                      }`}
                  >
                    <span>{lang.flag}</span>
                    <span>{lang.label}</span>
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Contact Support */}
          <button className="w-full flex items-center gap-2 px-3 py-2 rounded-lg bg-gradient-to-r from-lucky-gold/20 to-orange-500/20 border border-lucky-gold/30 hover:border-lucky-gold/60 text-white text-sm font-medium transition-colors">
            <MessageSquare size={16} />
            <span>Contact Support</span>
          </button>
        </div>
      ) : (
        <div className="flex-shrink-0 border-t border-lucky-gold/20 bg-black/95 p-2 space-y-2">
          {/* Language Selector - Icon Only */}
          <div className="relative" ref={langDropdownRef}>
            <button
              onClick={() => setShowLangDropdown(!showLangDropdown)}
              className="w-full p-2.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 text-white transition-colors flex items-center justify-center"
              title="Language"
            >
              <Globe size={20} />
            </button>
            {showLangDropdown && (
              <div className="absolute bottom-full left-full ml-2 mb-0 bg-zinc-900 border border-zinc-700 rounded-lg overflow-hidden shadow-lg">
                {languages.map((lang) => (
                  <button
                    key={lang.code}
                    onClick={() => {
                      setLanguage(lang.code)
                      setShowLangDropdown(false)
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-zinc-800 transition-colors whitespace-nowrap ${language === lang.code ? "text-lucky-gold" : "text-gray-300"
                      }`}
                  >
                    <span>{lang.flag}</span>
                    <span>{lang.label}</span>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Contact Support - Icon Only */}
          <button
            className="w-full p-2.5 rounded-lg bg-gradient-to-r from-lucky-gold/20 to-orange-500/20 border border-lucky-gold/30 hover:border-lucky-gold/60 text-white transition-colors flex items-center justify-center"
            title="Contact Support"
          >
            <MessageSquare size={20} />
          </button>
        </div>
      )}

      {/* Toggle Button */}
      <button
        onClick={toggleSidebar}
        className="absolute -right-3 top-4 w-6 h-6 rounded-full bg-lucky-gold text-black flex items-center justify-center hover:bg-yellow-500 transition-colors shadow-lg z-50"
      >
        {isExpanded ? <ChevronLeft size={14} /> : <ChevronRight size={14} />}
      </button>
    </div>
  )
}

// Helper Components
const NavItem: React.FC<{
  icon: React.ElementType
  label: string
  active: boolean
  onClick: () => void
  isExpanded: boolean
  badge?: string
}> = ({ icon: Icon, label, active, onClick, isExpanded, badge }) => (
  <button
    onClick={onClick}
    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all relative ${active
      ? "bg-gradient-to-r from-lucky-gold/20 to-orange-500/20 text-lucky-gold border border-lucky-gold/30"
      : "text-gray-300 hover:bg-zinc-900 hover:text-white"
      } ${!isExpanded && "justify-center"}`}
    title={!isExpanded ? label : undefined}
  >
    <Icon size={20} />
    {isExpanded && (
      <>
        <span className="flex-1 text-left text-sm font-medium">{label}</span>
        {badge && (
          <span className="text-[10px] px-2 py-0.5 rounded-full bg-lucky-gold/20 text-lucky-gold border border-lucky-gold/30">
            {badge}
          </span>
        )}
      </>
    )}
  </button>
)

const SocialButton: React.FC<{
  icon: string
  label: string
  onClick: () => void
}> = ({ icon, label, onClick }) => (
  <button
    onClick={onClick}
    className="flex-1 p-2 rounded-lg bg-zinc-900 hover:bg-zinc-800 transition-colors"
    title={label}
  >
    <img src={icon || "/placeholder.svg"} alt={label} className="w-5 h-5 mx-auto" />
  </button>
)

export default Sidebar
