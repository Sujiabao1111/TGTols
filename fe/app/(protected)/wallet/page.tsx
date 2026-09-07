"use client"

import { useState, Suspense } from "react"
import { useSearchParams } from "next/navigation"
import WalletPage from "@/components/WalletPage"
import { WalletTab } from "@/app/mocks/types"

function WalletContent() {
  const searchParams = useSearchParams()
  const tab = (searchParams.get("tab") as WalletTab) || "deposit"
  const [activeTab, setActiveTab] = useState<WalletTab>(tab)

  const handleTabChange = (newTab: WalletTab) => {
    setActiveTab(newTab)
    const url = new URL(window.location.href)
    url.searchParams.set("tab", newTab)
    window.history.replaceState(window.history.state, "", url.toString())
  }

  return <WalletPage activeTab={activeTab} onTabChange={handleTabChange} />
}

export default function WalletRoute() {
  return (
    <Suspense fallback={<div className="flex items-center justify-center min-h-screen">Loading...</div>}>
      <WalletContent />
    </Suspense>
  )
}
