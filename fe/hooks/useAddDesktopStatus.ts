"use client"

import { useEffect, useState, useCallback } from "react"
import { activityService, authService } from "@/services/api"

export interface SingleClaimActivityStatus {
  activity_type: string
  claimed: boolean
  claimed_at?: string
}

interface UseSingleClaimActivityStatusResult {
  status: SingleClaimActivityStatus | null
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

export function useAddDesktopStatus(enabled = true): UseSingleClaimActivityStatusResult {
  const [status, setStatus] = useState<SingleClaimActivityStatus | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  const fetchStatus = useCallback(async () => {
    if (!enabled || !authService.isAuthenticated()) {
      setStatus(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)
    try {
      const data = await activityService.getAddDesktopStatus()
      setStatus(data as SingleClaimActivityStatus)
    } catch (fetchError) {
      setError(getErrorMessage(fetchError, "failed to load add desktop status"))
      setStatus(null)
    } finally {
      setLoading(false)
    }
  }, [enabled])

  useEffect(() => {
    queueMicrotask(() => {
      void fetchStatus()
    })
  }, [fetchStatus])

  return { status, loading, error, refresh: fetchStatus }
}
