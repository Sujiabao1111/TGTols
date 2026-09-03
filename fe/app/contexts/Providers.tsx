"use client"

import { ReactNode, useEffect, useRef } from "react"
import { Toaster } from "sonner"
import { LanguageProvider } from "./LanguageContext"
import { AuthProvider, useAuth } from "./AuthContext"
import { UserProvider } from "./UserContext"
import { GameProvider } from "./GameContext"
import { activityService, authService } from "@/services/api"
import { DESKTOP_APP_LAUNCH_SOURCE, getLaunchSource, registerPwaServiceWorker } from "@/lib/desktop-launch"
import { AddDesktopInsuranceDialog } from "@/components/AddDesktopInsuranceDialog"
import { TonConnectUIProvider } from "@tonconnect/ui-react"

function DesktopLaunchRewardClaimer() {
  const { isLoggedIn } = useAuth()
  const attemptedRef = useRef(false)

  useEffect(() => {
    if (!isLoggedIn || attemptedRef.current) {
      return
    }

    if (getLaunchSource() !== DESKTOP_APP_LAUNCH_SOURCE) {
      return
    }

    attemptedRef.current = true
    void activityService
      .claimAddDesktop()
      .catch((error) => {
        console.error("Desktop launch insurance coupon claim failed:", error)
      })
  }, [isLoggedIn])

  return null
}

export function Providers({ children }: { children: ReactNode }) {
  useEffect(() => {
    registerPwaServiceWorker()
    getLaunchSource()
    void authService.recordDomainClick()
  }, [])

  return (
    <TonConnectUIProvider manifestUrl="https://gg.ppnet55.com/tonconnect-manifest.json">
    <LanguageProvider>
      <AuthProvider>
        <UserProvider>
          <GameProvider>
            <DesktopLaunchRewardClaimer />
            {children}
            <AddDesktopInsuranceDialog />
            <Toaster position="top-center" richColors />
          </GameProvider>
        </UserProvider>
      </AuthProvider>
    </LanguageProvider>
    </TonConnectUIProvider>
  )
}
