"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { activityService, authService } from "@/services/api"

const REFRESH_INTERVAL_MS = 5 * 1000

export interface BettingRankValue {
  current_value: number
  week_start: string
  next_reset_at: string
  refreshed_at: string
  completed_intervals: number
}

interface UseBettingRankValueResult {
  value: BettingRankValue | null
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

export function useBettingRankValue(enabled = true): UseBettingRankValueResult {
  const [value, setValue] = useState<BettingRankValue | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)
  const isInitialLoadRef = useRef(true)

  const fetchValue = useCallback(async () => {
    if (!enabled || !authService.isAuthenticated()) {
      setValue(null)
      setLoading(false)
      isInitialLoadRef.current = true
      return
    }

    setLoading(isInitialLoadRef.current)
    setError(null)
    try {
      const data = await activityService.getBettingRankValue()
      setValue(data as BettingRankValue)
    } catch (fetchError) {
      setError(getErrorMessage(fetchError, "failed to load betting rank value"))
      setValue(null)
    } finally {
      isInitialLoadRef.current = false
      setLoading(false)
    }
  }, [enabled])

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void fetchValue()
    }, 0)

    return () => {
      window.clearTimeout(timeoutId)
    }
  }, [fetchValue])

  useEffect(() => {
    if (!enabled || !authService.isAuthenticated()) {
      return
    }

    const intervalId = window.setInterval(() => {
      void fetchValue()
    }, REFRESH_INTERVAL_MS)

    return () => {
      window.clearInterval(intervalId)
    }
  }, [enabled, fetchValue])

  return { value, loading, error, refresh: fetchValue }
}
