"use client"
import React from "react"
import { useState, useRef, useEffect } from "react"
import { Menu, X, Search, Bell, ChevronLeft, ChevronDown, Gift } from "lucide-react"
import type { View, Language } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { useUser } from "@/app/contexts/UserContext"
import { useAddDesktopStatus } from "@/hooks/useAddDesktopStatus"
import { useNewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"

interface NavbarProps {
  currentView: View
  onNavigate: (view: View) => void
  isLoggedIn: boolean
  userBalance: number
}




const Navbar: React.FC<NavbarProps> = ({ currentView, onNavigate, isLoggedIn, userBalance }) => {

  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [isLangOpen, setIsLangOpen] = useState(false)
  const { language, setLanguage, t } = useLanguage()
  const [isActivityOpen, setIsActivityOpen] = useState(true)  // Activity menu state

  const { status: addDesktopStatus, loading: addDesktopStatusLoading } = useAddDesktopStatus(isLoggedIn)
  const { status: newUserRechargeStatus, loading: newUserRechargeLoading } = useNewUserRechargeStatus(isLoggedIn)
  const { formatUSD } = useExchangeRate()
  const { userProfile } = useUser()

  const activityViews: View[] = ["activity", "vip", "adddesktop", "bettingRank", "weeklySpinWheel", "dailyWeeklyChallenge", "sevenDayTopup", "newUserRecharge"]
  const isActivityView = activityViews.includes(currentView)


  // Refs for click-outside detection
  const langMenuRef = useRef<HTMLDivElement>(null)
  const mobileLangMenuRef = useRef<HTMLDivElement>(null)

  // Helper to get mobile title based on view
  const getMobileTitle = () => {
    switch (currentView) {
      case "wallet":
        return t("nav.deposit")
      case "invite":
        return t("nav.invite")
      case "activity":
        return t("nav.promo")
      case "vip":
        return t("nav.vip")
      case "sevenDayTopup":
        return t("nav.sevenDayTopup")
      case "newUserRecharge":
        return t("nav.newUserRecharge")
      case "weeklySpinWheel":
        return t("nav.weeklySpinWheel")
      case "dailyWeeklyChallenge":
        return t("nav.dailyWeeklyChallenge")
      case "profile":
        return t("nav.profile")
      case "game":
        return "Playing"
      case "all-games":
        return "Game Lobby"
      case "login":
        return "Welcome"
      default:
        return null
    }
  }

  const mobileTitle = getMobileTitle()

  const handleLangChange = (lang: Language) => {
    setLanguage(lang)
    setIsLangOpen(false)
  }

  // Close dropdown when clicking outside (Desktop or Mobile)
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node

      const isDesktopInside = langMenuRef.current && langMenuRef.current.contains(target)
      const isMobileInside = mobileLangMenuRef.current && mobileLangMenuRef.current.contains(target)

      if (!isDesktopInside && !isMobileInside) {
        setIsLangOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const languages: { code: Language; label: string; flag: string }[] = [
    { code: "en", label: "English", flag: "🇺🇸" },
    { code: "ru", label: "Русский", flag: "🇷🇺" },
  ]

  const currentLangObj = languages.find((l) => l.code === language) || languages[0]
  const showAddDesktopEntry = !isLoggedIn || addDesktopStatusLoading || !addDesktopStatus?.claimed
  const showNewUserRechargeEntry = !isLoggedIn || newUserRechargeLoading || !newUserRechargeStatus?.hidden

  // Hide Navbar completely on Login page if on mobile, or just simplify it
  if (currentView === "login") {
    return (
      <nav className="bg-transparent p-4">
        <button
          onClick={() => onNavigate("home")}
          className="bg-black/40 backdrop-blur-md p-2 rounded-full text-white hover:bg-black/60 transition-colors"
        >
          <ChevronLeft size={24} />
        </button>
      </nav>
    )
  }

  return (
    <>
      <nav className="relative bg-gradient-to-b from-black via-black/95 to-black/90 backdrop-blur-xl border-t-2 border-t-lucky-red shadow-[0_8px_32px_rgba(0,0,0,0.8)] before:absolute before:inset-0 before:bg-gradient-to-b before:from-lucky-gold/5 before:via-transparent before:to-transparent before:pointer-events-none after:absolute after:bottom-0 after:left-0 after:right-0 after:h-px after:bg-gradient-to-r after:from-transparent after:via-lucky-gold/30 after:to-transparent">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="flex items-center justify-between h-16">
            {/* --- Desktop View & Mobile Home View (Logo) --- */}
            <div className="flex items-center gap-2">
              {currentView === "all-games" ? (
                <button
                  onClick={() => onNavigate("home")}
                  className="flex items-center gap-1 text-gray-300 hover:text-white"
                >
                  <ChevronLeft size={24} />
                </button>
              ) : mobileTitle ? (
                <button
                  onClick={() => onNavigate("home")}
                  className="flex items-center gap-2 text-gray-300 hover:text-white"
                >
                  <ChevronLeft size={20} />
                  <span className="font-semibold text-white">{mobileTitle}</span>
                </button>
              ) : (
                <button
                  type="button"
                  onClick={() => onNavigate("home")}
                  className="flex h-18items-center transition-opacity hover:opacity-90"
                  aria-label="LuckyBear Home"
                >
                  <img
                    src="/images/logo.png"
                    alt="LuckyBear"
                    className="h-15 w-auto object-contain"
                  />
                </button>
              )}
            </div>

            {/* --- Desktop Navigation --- */}
            <div
              className={`hidden md:flex items-center space-x-8 ${currentView === "game" ? "opacity-0 pointer-events-none" : ""}`}
            >
              <a
                onClick={() => onNavigate("home")}
                className={`cursor-pointer transition-colors font-medium ${currentView === "home" ? "text-lucky-red" : "text-gray-300 hover:text-lucky-gold"}`}
              >
                {t("nav.home")}
              </a>
              <a
                onClick={() => onNavigate("activity")}
                className={`cursor-pointer transition-colors font-medium ${isActivityView ? "text-lucky-gold" : "text-gray-300 hover:text-lucky-gold"}`}
              >
                {t("nav.promo")}
              </a>
              <a
                onClick={() => onNavigate("invite")}
                className={`cursor-pointer transition-colors font-medium ${currentView === "invite" ? "text-lucky-gold" : "text-gray-300 hover:text-lucky-gold"}`}
              >
                {t("nav.invite")}
              </a>
            </div>

            {/* --- Right Actions (Desktop) --- */}
            <div className="hidden md:flex items-center gap-4">
              {/* Logic for Logged In vs Guest */}
              {isLoggedIn ? (
                <>
                  {currentView !== "game" && (
                    <>
                      <button className="text-gray-400 hover:text-white transition-colors">
                        <Search size={20} />
                      </button>
                      <button className="text-gray-400 hover:text-white transition-colors relative">
                        <Bell size={20} />
                        <span className="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full"></span>
                      </button>
                    </>
                  )}

                  <button
                    onClick={() => onNavigate("wallet")}
                    className="flex items-center gap-2 bg-gradient-to-r from-lucky-gold/20 to-orange-500/20 px-4 py-1.5 rounded-full border border-lucky-gold/30 hover:border-lucky-gold transition-colors"
                  >
                    <span className="text-lucky-gold font-bold text-sm">{formatUSD(userBalance)}</span>
                    <div className="w-6 h-6 rounded-full bg-gradient-to-br from-lucky-gold to-orange-500 flex items-center justify-center text-black font-bold text-xs shadow-lg">
                      +
                    </div>
                  </button>

                  <button
                    onClick={() => onNavigate("profile")}
                    className="w-8 h-8 rounded-full bg-gradient-to-br from-lucky-gold via-orange-500 to-red-500 flex items-center justify-center text-black font-bold border-2 border-white/20 shadow-lg"
                  >
                    U
                  </button>
                </>
              ) : (
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => onNavigate("login")}
                    className="px-4 py-2 text-sm font-bold text-white hover:text-lucky-gold transition-colors"
                  >
                    {t("nav.login")}
                  </button>
                  <button
                    onClick={() => onNavigate("login")}
                    className="px-5 py-2 text-sm font-bold bg-gradient-to-r from-lucky-gold to-orange-500 text-black rounded-full hover:shadow-lg hover:shadow-lucky-gold/50 transition-all"
                  >
                    {t("nav.register")}
                  </button>
                </div>
              )}
            </div>

            {/* --- Mobile Right Actions (Dynamic) --- */}
            <div className="md:hidden flex items-center gap-3">
              {currentView === "home" && (
                <>
                  {/* Mobile Language Dropdown */}
                  <div className="relative" ref={mobileLangMenuRef}>
                    <button
                      onClick={() => setIsLangOpen(!isLangOpen)}
                      className="flex items-center gap-1 bg-white/5 border border-white/20 rounded-full px-2 py-1"
                    >
                      <span className="text-lg leading-none">{currentLangObj.flag}</span>
                      <span className="text-xs font-bold text-gray-300 uppercase">{language}</span>
                      <ChevronDown size={10} className="text-gray-400" />
                    </button>
                    {isLangOpen && (
                      <>
                        <div className="fixed inset-0 z-40" onClick={() => setIsLangOpen(false)}></div>
                        <div className="absolute top-full left-0 mt-2 w-32 bg-black border border-white/10 rounded-lg shadow-2xl py-1 z-50 animate-slide-in-left">
                          {languages.map((lang) => (
                            <button
                              key={lang.code}
                              onClick={() => handleLangChange(lang.code)}
                              className={`block w-full text-left px-3 py-2 text-sm hover:bg-white/10 flex items-center gap-2 ${language === lang.code ? "text-lucky-gold" : "text-gray-300"}`}
                            >
                              <span className="text-base">{lang.flag}</span>
                              <span>{lang.label}</span>
                            </button>
                          ))}
                        </div>
                      </>
                    )}
                  </div>

                  {isLoggedIn ? (
                    <button
                      onClick={() => onNavigate("wallet")}
                      className="bg-gradient-to-r from-lucky-gold/20 to-orange-500/20 px-3 py-1 rounded-full text-lucky-gold font-bold text-sm border border-lucky-gold/30"
                    >
                      {formatUSD(userBalance)}
                    </button>
                  ) : (
                    <button
                      onClick={() => onNavigate("login")}
                      className="bg-gradient-to-r from-lucky-gold to-orange-500 text-black px-4 py-1 rounded-full font-bold text-sm shadow-lg"
                    >
                      {t("nav.login")}
                    </button>
                  )}
                </>
              )}

              {(currentView === "game" || currentView === "all-games") && (
                <button onClick={() => onNavigate("home")} className="bg-white/10 p-1.5 rounded-full text-white">
                  <X size={18} />
                </button>
              )}

              {(currentView === "invite" || isActivityView) && (
                <button className="text-gray-300 relative">
                  <Bell size={20} />
                  <span className="absolute top-0 right-0 w-2 h-2 bg-red-500 rounded-full"></span>
                </button>
              )}

              <button onClick={() => setIsMenuOpen(!isMenuOpen)} className="text-gray-300 hover:text-white p-1">
                <Menu size={20} />
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Mobile Menu Dropdown - Outside nav to avoid z-index constraint */}
      {isMenuOpen && (
        <>
          {/* Backdrop */}
          <div
            className="md:hidden fixed inset-0 bg-black/60 z-[99999]"
            onClick={() => setIsMenuOpen(false)}
          ></div>
          {/* Side Menu */}
          <div className="md:hidden fixed top-0 left-0 h-full w-80 bg-gradient-to-b from-gray-900 via-gray-800 to-gray-900 z-[100000] shadow-2xl shadow-black/50 border-r border-lucky-gold/30 animate-slide-in-left overflow-y-auto">
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-4 border-b border-gray-700">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-lucky-gold to-yellow-600 flex items-center justify-center">
                  <span className="text-black font-bold text-lg"></span>
                </div>
                <div>
                  <div className="text-white font-bold">{userProfile.username}</div>
                  <div className="text-gray-400 text-xs">UID: {userProfile.uid}</div>
                </div>
              </div>
              <button
                onClick={() => setIsMenuOpen(false)}
                className="w-8 h-8 rounded-full bg-gray-700/50 flex items-center justify-center text-gray-400 hover:text-white hover:bg-gray-600 transition-colors"
              >
                <X size={18} />
              </button>
            </div>

            {/* Balance Section */}
            {/* {isLoggedIn && (
              <div className="mx-4 mt-4 p-4 bg-lucky-gold/10 rounded-xl border border-lucky-gold/20">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-gray-400 text-xs">{t("nav.home")}</div>
                    <div className="text-lucky-gold font-bold text-xl mt-1">{formatUSD(userBalance)}</div>
                  </div>
                  <ChevronDown size={18} className="text-lucky-gold rotate-[-90deg]" />
                </div>
              </div>
            )} */}

            {/* Menu Items */}
            <div className="px-4 py-4 space-y-1">
              {/* Home */}
              <a
                onClick={() => {
                  onNavigate("home")
                  setIsMenuOpen(false)
                }}
                className={`flex items-center px-3 py-2.5 rounded-lg transition-colors ${currentView === "home" ? 'bg-lucky-gold/20 text-lucky-gold' : 'text-gray-300 hover:bg-gray-700/50'}`}
              >
                <div className="flex items-center gap-3">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                  </svg>
                  <span>{t("nav.home")}</span>
                </div>
              </a>


              {/* Activity */}
              <div className="bg-lucky-gold/10 rounded-lg border border-lucky-gold/30">
                <button
                  onClick={() => setIsActivityOpen(!isActivityOpen)}
                  className={`w-full flex items-center justify-between px-3 py-2.5 transition-colors ${isActivityView ? 'text-lucky-gold' : 'text-gray-300'}`}
                >
                  <div className="flex items-center gap-3">
                    <Gift size={20} />
                    <span>{t("nav.activity")}</span>
                  </div>
                  <ChevronDown
                    className={`transition-transform ${isActivityOpen ? 'rotate-180' : ''} ${isActivityView ? 'text-lucky-gold' : 'text-gray-500'}`}
                  />
                </button>
                {/* Show or hide submenu based on state */}
                {isActivityOpen && (
                  <div className="px-3 pb-2 space-y-1">
                    {[
                      ...(showAddDesktopEntry ? [{ label: <span>{t("nav.adddesktop")}</span>, tag: 'HOT', tagColor: 'bg-red-500', view: 'adddesktop' as View }] : []),
                      ...(showNewUserRechargeEntry ? [{ label: <span>{t("nav.newUserRecharge")}</span>, tag: 'HOT', tagColor: 'bg-red-500', view: 'newUserRecharge' as View }] : []),
                      { label: <span>{t("nav.sevenDayTopup")}</span>, tag: 'HOT', tagColor: 'bg-red-500', view: 'sevenDayTopup' as View },
                      { label: <span>{t("nav.vip")}</span>, tag: 'HOT', tagColor: 'bg-red-500', view: 'vip' as View },
                      { label: <span>{t("nav.weeklySpinWheel")}</span>, tag: 'NEW', tagColor: 'bg-green-500', view: 'weeklySpinWheel' as View },
                      { label: <span>{t("nav.dailyWeeklyChallenge")}</span>, tag: 'HOT', tagColor: 'bg-red-500', view: 'dailyWeeklyChallenge' as View },
                      { label: <span>{t("nav.bettingRank")}</span>, tag: 'NEW', tagColor: 'bg-green-500', view: 'bettingRank' as View },
                    ].map((item) => (
                      <a
                        key={item.view}
                        onClick={() => {
                          onNavigate(item.view)  // Navigate to the selected activity page
                          setIsMenuOpen(false)
                        }}
                        className={`flex items-center px-3 py-2 rounded text-sm transition-colors ${currentView === item.view ? "bg-gray-700/60 text-white" : "text-gray-400 hover:text-white hover:bg-gray-700/50"}`}
                      >
                        <div className="flex items-center gap-2">
                          <Gift size={16} className="text-lucky-gold" />
                          <span>{item.label}</span>
                          {item.tag && (
                            <span className={`${item.tagColor} text-white text-xs px-1.5 py-0.5 rounded`}>
                              {item.tag}
                            </span>
                          )}
                        </div>
                        {/* <ChevronDown size={14} className="text-gray-500 rotate-[-90deg]" /> */}
                      </a>
                    ))}
                  </div>
                )}
              </div>

              {/* Promotions */}
              <a
                onClick={() => {
                  onNavigate("activity")
                  setIsMenuOpen(false)
                }}
                className={`flex items-center px-3 py-2.5 rounded-lg transition-colors ${currentView === "activity" ? 'bg-lucky-gold/20 text-lucky-gold' : 'text-gray-300 hover:bg-gray-700/50'}`}
              >
                <div className="flex items-center gap-3">
                  <Gift size={20} />
                  <span>{t("nav.promo")}</span>
                </div>
              </a>

              {/* Invite */}
              <a
                onClick={() => {
                  onNavigate("invite")
                  setIsMenuOpen(false)
                }}
                className={`flex items-center px-3 py-2.5 rounded-lg transition-colors ${currentView === "invite" ? 'bg-lucky-gold/20 text-lucky-gold' : 'text-gray-300 hover:bg-gray-700/50'}`}
              >
                <div className="flex items-center gap-3">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
                  </svg>
                  <span>{t("nav.invite")}</span>
                </div>
              </a>

              {/* Wallet */}
              <a
                onClick={() => {
                  onNavigate("wallet")
                  setIsMenuOpen(false)
                }}
                className={`flex items-center px-3 py-2.5 rounded-lg transition-colors ${currentView === "wallet" ? 'bg-lucky-gold/20 text-lucky-gold' : 'text-gray-300 hover:bg-gray-700/50'}`}
              >
                <div className="flex items-center gap-3">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{t("nav.deposit")}</span>
                </div>
              </a>


              {/* Login */}
              {!isLoggedIn && (
                <a
                  onClick={() => {
                    onNavigate("login")
                    setIsMenuOpen(false)
                  }}
                  className="flex items-center justify-center px-3 py-3 mt-4 rounded-lg bg-gradient-to-r from-lucky-gold to-yellow-600 text-black font-bold hover:opacity-90 transition-opacity"
                >
                  {t("nav.login")} / {t("nav.register")}
                </a>
              )}
            </div>
          </div>
        </>
      )}
    </>
  )
}

export default Navbar
