"use client"

import { useCallback, useEffect, useState } from "react"
import {
  hasInstalledRelatedPwa,
  isDesktopAppDisplayMode,
  isPwaInstallRemembered,
  rememberPwaInstalled,
} from "@/lib/desktop-launch"

type InstallOutcome = "accepted" | "dismissed"
export type PwaManualInstallPlatform = "ios" | "android" | "desktop" | "unknown"

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: InstallOutcome; platform: string }>
}

type WindowWithInstallPrompt = Window & {
  __pwaInstallPromptEvent?: BeforeInstallPromptEvent | null
}

function isIosDevice() {
  if (typeof window === "undefined") {
    return false
  }

  const userAgent = window.navigator.userAgent.toLowerCase()
  const isTouchMac = window.navigator.maxTouchPoints > 1 && userAgent.includes("macintosh")
  return /iphone|ipad|ipod/.test(userAgent) || isTouchMac
}

function getManualInstallPlatform(): PwaManualInstallPlatform {
  if (typeof window === "undefined") {
    return "unknown"
  }

  const userAgent = window.navigator.userAgent.toLowerCase()
  if (isIosDevice()) {
    return "ios"
  }
  if (userAgent.includes("android")) {
    return "android"
  }
  if (userAgent.includes("windows") || userAgent.includes("macintosh") || userAgent.includes("linux")) {
    return "desktop"
  }

  return "unknown"
}

function isTelegramWebView() {
  if (typeof window === "undefined") return false
  return Boolean((window as typeof window & { Telegram?: { WebApp?: unknown } }).Telegram?.WebApp)
}

export function usePwaInstallPrompt() {
  const [promptEvent, setPromptEvent] = useState<BeforeInstallPromptEvent | null>(null)
  const [isInstalled, setIsInstalled] = useState(false)
  const [isInstalling, setIsInstalling] = useState(false)
  const [ready, setReady] = useState(false)
  const [lastOutcome, setLastOutcome] = useState<InstallOutcome | null>(null)
  const [showManualInstructions, setShowManualInstructions] = useState(false)
  const [manualInstallPlatform, setManualInstallPlatform] = useState<PwaManualInstallPlatform>("unknown")
  const [isTelegram, setIsTelegram] = useState(false)

  useEffect(() => {
    const installPromptWindow = window as WindowWithInstallPrompt

    queueMicrotask(() => {
      const displayModeInstalled = isDesktopAppDisplayMode()
      const rememberedInstalled = isPwaInstallRemembered()
      if (displayModeInstalled || rememberedInstalled) {
        rememberPwaInstalled()
        setIsInstalled(true)
        setReady(true)
      } else {
        setIsInstalled(false)
        void hasInstalledRelatedPwa().then((installed) => {
          if (installed) {
            rememberPwaInstalled()
            setIsInstalled(true)
          }
          setReady(true)
        })
      }

      setPromptEvent(installPromptWindow.__pwaInstallPromptEvent ?? null)
      setManualInstallPlatform(getManualInstallPlatform())
      setShowManualInstructions(isIosDevice())
      setIsTelegram(isTelegramWebView())
    })

    const handleBeforeInstallPrompt = (event: Event) => {
      event.preventDefault()
      const installEvent = event as BeforeInstallPromptEvent
      installPromptWindow.__pwaInstallPromptEvent = installEvent
      setPromptEvent(installEvent)
      setShowManualInstructions(false)
    }

    const handleAppInstalled = () => {
      installPromptWindow.__pwaInstallPromptEvent = null
      rememberPwaInstalled()
      setPromptEvent(null)
      setIsInstalled(true)
      setReady(true)
      setLastOutcome("accepted")
      setShowManualInstructions(false)
    }

    window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt)
    window.addEventListener("appinstalled", handleAppInstalled)

    return () => {
      window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt)
      window.removeEventListener("appinstalled", handleAppInstalled)
    }
  }, [])

  const install = useCallback(async () => {
    if (isInstalled || isInstalling) {
      return
    }

    if (!promptEvent) {
      setShowManualInstructions(true)
      return
    }

    setIsInstalling(true)

    try {
      await promptEvent.prompt()
      const choice = await promptEvent.userChoice
      setLastOutcome(choice.outcome)
      setPromptEvent(null)
      ;(window as WindowWithInstallPrompt).__pwaInstallPromptEvent = null

      if (choice.outcome === "accepted") {
        setShowManualInstructions(false)
        rememberPwaInstalled()
        setIsInstalled(true)
      } else {
        setShowManualInstructions(true)
      }
    } finally {
      setIsInstalling(false)
    }
  }, [isInstalled, isInstalling, promptEvent])

  return {
    canPrompt: !!promptEvent,
    install,
    isInstalled,
    isInstalling,
    ready,
    lastOutcome,
    showManualInstructions,
    manualInstallPlatform,
    isTelegram,
  }
}
