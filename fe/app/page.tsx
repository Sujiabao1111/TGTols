"use client"

import { useEffect, Suspense } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import LoadingScreen from "../components/LoadingScreen"
import { saveInviteCode } from "@/lib/invite-code"

function RootContent() {
  const router = useRouter()
  const searchParams = useSearchParams()

  useEffect(() => {
    saveInviteCode(searchParams.get("code"))
    router.replace("/home")
  }, [router, searchParams])

  // Show loading screen while redirecting
  return <LoadingScreen />
}

export default function RootPage() {
  return (
    <Suspense fallback={<LoadingScreen />}>
      <RootContent />
    </Suspense>
  )
}
