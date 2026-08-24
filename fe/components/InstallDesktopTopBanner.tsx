"use client"

import { Download, X } from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"

interface InstallDesktopTopBannerProps {
  isInstalling: boolean
  onDismiss: () => void
  onInstall: () => void
  onOpenActivity: () => void
  visible: boolean
}

export function InstallDesktopTopBanner({
  isInstalling,
  onDismiss,
  onInstall,
  onOpenActivity,
  visible,
}: InstallDesktopTopBannerProps) {
  const { t } = useLanguage()

  if (!visible) {
    return null
  }

  return (
    <div
      className="h-[var(--install-banner-height)] cursor-pointer bg-[#11983a] text-white shadow-lg"
      onClick={onOpenActivity}
      role="button"
      tabIndex={0}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault()
          onOpenActivity()
        }
      }}
    >
      <div className="mx-auto flex h-full max-w-7xl items-center gap-2 px-3 sm:gap-3 sm:px-6 lg:px-8">
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation()
            onDismiss()
          }}
          aria-label="Close install banner"
          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-white/85 transition-colors hover:bg-white/15 hover:text-white active:bg-white/20"
        >
          <X size={18} />
        </button>
        <img
          src="/images/PPLogo.png"
          alt="PPNET"
          className="ml-1 h-7 w-7 shrink-0 rounded-sm object-contain sm:ml-2 sm:h-8 sm:w-8"
        />
        <div className="min-w-0 flex-1 overflow-hidden text-center text-[11px] font-extrabold leading-tight tracking-normal [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2] sm:text-sm md:text-base">
          {t("pwa.install_banner.title")}
        </div>
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation()
            onInstall()
          }}
          disabled={isInstalling}
          className="flex h-8 shrink-0 items-center gap-1 rounded-md bg-white px-3 text-xs font-bold text-[#0c7d2f] shadow-md transition-transform hover:scale-[1.03] active:scale-95 disabled:cursor-wait disabled:opacity-70 sm:px-4 sm:text-sm"
        >
          <Download size={14} />
          {isInstalling ? t("pwa.install_button.installing") : t("pwa.install_banner.button")}
        </button>
      </div>
    </div>
  )
}
