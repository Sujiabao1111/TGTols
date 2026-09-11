"use client"

import type React from "react"
import Navbar from "./NavBar"
import MobileNav from "./MobileNav"
import LiveChatWidget from "./LiveChatWidget"
import Sidebar from "./SideBar"
import ScrollToTop from "./ScrollToTop"
import { InstallDesktopTopBanner } from "./InstallDesktopTopBanner"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useAddDesktopStatus } from "@/hooks/useAddDesktopStatus"
import { usePwaInstallPrompt } from "@/hooks/usePwaInstallPrompt"
import type { View, GameType } from "../app/mocks/types"
import { useCallback, useEffect, useMemo, useState } from "react"

interface LayoutProps {
  children: React.ReactNode
  currentView: View
  onNavigate: (view: View) => void
  isLoggedIn: boolean
  userBalance: number
  onBalanceUpdate: () => void
  gameTypes?: GameType[]
  onGameTypeSelect?: (gameType: GameType) => void
  onClearFilters?: () => void
  sortType?: "POPULAR" | "RECOMMEND" | "NEW" | undefined
  onSortTypeSelect?: (sortType: "POPULAR" | "RECOMMEND" | "NEW" | undefined) => void
  externalGameType?: GameType | null
}

const Layout: React.FC<LayoutProps> = ({
  children,
  currentView,
  onNavigate,
  isLoggedIn,
  userBalance,
  gameTypes = [],
  onGameTypeSelect,
  onClearFilters,
  sortType,
  onSortTypeSelect,
  externalGameType,
}) => {
  const { t } = useLanguage()
  const showSidebar = currentView !== "login" && currentView !== "game"
  const [isSidebarExpanded, setIsSidebarExpanded] = useState(true)
  const [isInstallBannerDismissed, setIsInstallBannerDismissed] = useState(false)
  const [showUnsupportedPwaNotice, setShowUnsupportedPwaNotice] = useState(false)
  const { status: addDesktopStatus, loading: addDesktopStatusLoading } = useAddDesktopStatus(isLoggedIn)
  const {
    canPrompt: canPromptPwaInstall,
    install: installPwa,
    isInstalled: isPwaInstalled,
    isInstalling: isPwaInstalling,
    ready: pwaInstallReady,
    isTelegram,
  } = usePwaInstallPrompt()
  useEffect(() => {
    const dismissedDate = window.localStorage.getItem("install-banner-dismissed-date")
    const today = new Date().toISOString().slice(0, 10)
    setIsInstallBannerDismissed(dismissedDate === today)
  }, [])
  const isInstallBannerVisible = useMemo(() => {
    if (!pwaInstallReady || isTelegram || currentView === "login" || isInstallBannerDismissed || isPwaInstalled) {
      return false
    }

    if (!isLoggedIn) {
      return true
    }

    if (addDesktopStatusLoading) {
      return false
    }

    return !addDesktopStatus?.claimed
  }, [
    addDesktopStatus?.claimed,
    addDesktopStatusLoading,
    currentView,
    isInstallBannerDismissed,
    isLoggedIn,
    isPwaInstalled,
    pwaInstallReady,
    isTelegram,
  ])

  const hasActiveFilter: boolean = Boolean(sortType) || Boolean(externalGameType && externalGameType.code !== "SLOT")
  const handleNavigate = useCallback(
    (view: View) => {
      if (view === "home") {
        onClearFilters?.()
      }

      onNavigate(view)
    },
    [onClearFilters, onNavigate],
  )
  const handleDismissInstallBanner = useCallback(() => {
    setIsInstallBannerDismissed(true)
    window.localStorage.setItem("install-banner-dismissed-date", new Date().toISOString().slice(0, 10))
  }, [])
  const handleInstallDesktop = useCallback(() => {
    if (canPromptPwaInstall) {
      void installPwa()
      return
    }

    setShowUnsupportedPwaNotice(true)
    handleNavigate("adddesktop")
  }, [canPromptPwaInstall, handleNavigate, installPwa])
  useEffect(() => {
    if (!showUnsupportedPwaNotice) {
      return
    }

    const timer = window.setTimeout(() => {
      setShowUnsupportedPwaNotice(false)
    }, 2200)

    return () => window.clearTimeout(timer)
  }, [showUnsupportedPwaNotice])
  const headerOffsetStyle = { height: "calc(var(--app-header-height) + var(--app-header-clearance))" }

  return (
      <div
        className="install-banner-shell min-h-screen bg-lucky-dark text-white font-sans selection:bg-lucky-gold selection:text-black flex flex-col relative overflow-hidden"
        data-install-banner={isInstallBannerVisible ? "true" : "false"}
      >
        <header className="fixed left-0 right-0 top-0 z-50">
          <InstallDesktopTopBanner
            isInstalling={isPwaInstalling}
            onDismiss={handleDismissInstallBanner}
            onInstall={handleInstallDesktop}
            onOpenActivity={() => handleNavigate("adddesktop")}
            visible={false}
          />

          {/* Header / Navbar */}
          <Navbar
            currentView={currentView}
            onNavigate={handleNavigate}
            isLoggedIn={isLoggedIn}
            userBalance={userBalance}
          />
        </header>

        {currentView !== "login" && <div aria-hidden="true" className="shrink-0" style={headerOffsetStyle} />}

        {/* Sidebar - Desktop only */}
        {showSidebar && (
          <div className="hidden lg:block">
            <Sidebar
              currentView={currentView}
              onNavigate={handleNavigate}
              isLoggedIn={isLoggedIn}
              onSidebarToggle={setIsSidebarExpanded}
              gameTypes={gameTypes}
              onGameTypeSelect={onGameTypeSelect}
              onClearFilters={onClearFilters}
              sortType={sortType}
              onSortTypeSelect={onSortTypeSelect}
              activeFilter={hasActiveFilter}
              externalGameType={externalGameType}
            />
          </div>
        )}

        {/* Main Content */}
        <main
          className={`flex-grow relative z-10 transition-all duration-300 ${
            currentView === "login"
              ? "pt-0 pb-0"
              : currentView === "game"
                ? "pb-0"
                : "pb-20 md:pb-8"
          } ${showSidebar ? (isSidebarExpanded ? "lg:ml-64" : "lg:ml-16") : ""}`}
          style={
            currentView === "login"
              ? undefined
              : { scrollPaddingTop: "calc(var(--app-header-height) + var(--app-header-clearance))" }
          }
        >
          {children}
        </main>

        {/* Footer temporarily hidden per product request. Keep the component for future re-enable. */}

        {/* Mobile Bottom Navigation - Games use all available screen space. */}
        {currentView !== "login" && currentView !== "game" && (
          <MobileNav currentView={currentView} onNavigate={handleNavigate} />
        )}

        {/* Global LiveChat launcher - Hide on login and third-party game views */}
        {currentView !== "login" && currentView !== "game" && <LiveChatWidget />}

        {showUnsupportedPwaNotice && (
          <div className="pointer-events-none fixed inset-0 z-[100000] flex items-center justify-center px-6">
            <div className="max-w-[22rem] rounded-xl border border-white/15 bg-black/85 px-5 py-3 text-center text-sm font-semibold text-white shadow-2xl backdrop-blur-md sm:text-base">
              {t("pwa.unsupported")}
            </div>
          </div>
        )}

        {/* Scroll to Top button - Show only on home and game list pages */}
        {(currentView === "home" || currentView === "all-games") && <ScrollToTop />}
      </div>
  )
}

export default Layout
