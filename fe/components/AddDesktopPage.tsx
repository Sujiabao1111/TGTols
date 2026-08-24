"use client"

import React from "react"
import { Download } from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useAddDesktopStatus } from "@/hooks/useAddDesktopStatus"
import { usePwaInstallPrompt } from "@/hooks/usePwaInstallPrompt"

const AddDesktopPage: React.FC = () => {
  const { language, t } = useLanguage()
  const { status } = useAddDesktopStatus(true)
  const {
    install,
    isInstalled,
    isInstalling,
    lastOutcome,
    manualInstallPlatform,
    showManualInstructions,
  } = usePwaInstallPrompt()

  const isIndonesian = language === "id"
  const portraitBackgroundImage = isIndonesian
    ? "/images/activity6/addDesktop_yinni.png"
    : "/images/activity6/addDesktop_yingyu.png"
 
  const landscapeBackgroundImage = isIndonesian
    ? "/images/activity6/yinnihengbg.png"
    : "/images/activity6/yingyuhengbg.png"
  const landscapeRewardTop = isIndonesian ? "28.3%" : "calc(28.3% + 12px)"
  const installButtonLabel = isInstalled
    ? t("pwa.install_button.installed")
    : isInstalling
      ? t("pwa.install_button.installing")
      : t("pwa.install_banner.button")
  const manualInstructionKey =
    manualInstallPlatform === "ios"
      ? "pwa.manual.ios"
      : manualInstallPlatform === "android"
        ? "pwa.manual.android"
        : manualInstallPlatform === "desktop"
          ? "pwa.manual.desktop"
          : "pwa.manual.generic"
  const renderInstallButton = (className = "") => (
    <button
      type="button"
      onClick={install}
      disabled={isInstalled || isInstalling}
      aria-label={isInstalled ? "PPNET is already added to your desktop" : "Download and add PPNET to desktop"}
      className={`mx-auto flex w-[86%] max-w-[30rem] items-center justify-center gap-2 rounded-lg border border-yellow-100/70 bg-gradient-to-b from-yellow-200 via-yellow-400 to-orange-500 px-4 py-2 text-sm font-black text-black shadow-[0_8px_24px_rgba(0,0,0,0.35)] transition-transform hover:scale-[1.03] active:scale-95 disabled:cursor-default disabled:opacity-70 disabled:hover:scale-100 sm:py-2.5 sm:text-base [@media(orientation:landscape)]:w-[42%] [@media(orientation:landscape)]:max-w-[32rem] ${className}`}
    >
      <Download size={18} strokeWidth={2.5} />
      <span className="leading-none">{installButtonLabel}</span>
    </button>
  )

  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-900 via-gray-800 to-gray-900 flex justify-center px-3 py-4 sm:px-6 sm:py-6">
      <div className="w-full max-w-md sm:max-w-lg md:max-w-xl [@media(orientation:landscape)]:max-w-6xl">
        <div className="relative [@media(orientation:landscape)]:hidden">
          <img
            src={portraitBackgroundImage}
            alt="Add Desktop Activity"
            className="w-full h-auto rounded-xl shadow-2xl"
          />

          <div
            className="text-lucky-gold font-bold text-3xl drop-shadow-lg whitespace-nowrap"
            style={{
              position: "absolute",
              top: "39%",
              left: "60.5%",
              transform: "translateX(-50%)",
              zIndex: 10
            }}
          >
          </div>
        </div>

        <div className="relative hidden [@media(orientation:landscape)]:block">
          <img
            src={landscapeBackgroundImage}
            alt="Add Desktop Activity"
            className="w-full h-auto rounded-xl shadow-2xl"
          />

          <div
            className="text-lucky-gold font-bold text-[clamp(2rem,4vw,4rem)] drop-shadow-lg whitespace-nowrap"
            style={{
              position: "absolute",
              top: landscapeRewardTop,
              left: "22.8%",
              transform: "translateX(-50%)",
              zIndex: 10,
            }}
          >
          </div>
        </div>

        {renderInstallButton("mt-3 [@media(orientation:landscape)]:mt-2")}

        {isInstalling && (
          <div className="mt-4 rounded-lg border border-lucky-gold/30 bg-lucky-gold/10 px-4 py-3 text-sm text-lucky-gold">
            {t("pwa.install_status.opening")}
          </div>
        )}

        {isInstalled && (
          <div className="mt-4 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300">
            {t("pwa.install_status.installed")}
          </div>
        )}

        {showManualInstructions && !isInstalled && (
          <div className="mt-4 rounded-lg border border-white/15 bg-white/10 px-4 py-3 text-sm text-white/90">
            {t(manualInstructionKey)}
          </div>
        )}

        {lastOutcome === "dismissed" && !isInstalled && (
          <div className="mt-4 rounded-lg border border-yellow-500/30 bg-yellow-500/10 px-4 py-3 text-sm text-yellow-200">
            {t("pwa.install_status.dismissed")}
          </div>
        )}

        {status?.claimed && (
          <div className="mt-4 rounded-lg border border-lucky-gold/30 bg-lucky-gold/10 px-4 py-3 text-sm text-lucky-gold">
            {t("pwa.reward.saved")}
            {status.claimed_at ? ` ${status.claimed_at}` : ""}.
          </div>
        )}
      </div>
    </div>
  )
}

export default AddDesktopPage
