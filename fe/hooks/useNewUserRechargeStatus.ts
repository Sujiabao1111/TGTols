"use client"

import { useCallback, useEffect, useState } from "react"
import { activityService, authService } from "@/services/api"
import { useExchangeRate } from "@/hooks/useExchangeRate"

export interface NewUserRechargeTierStatus {
  day: number
  rate: number
  status: "locked" | "current" | "unlocked"
  deposit_amount: number
  reward_amount: number
  min_deposit: number
  wager_required: number
  wager_completed: number
  deposit_amount_u?: number
  reward_amount_u?: number
  min_deposit_u?: number
  wager_required_u?: number
  wager_completed_u?: number
  wager_unlocked: boolean
  reward_granted_at?: string
}

export interface NewUserRechargeStatus {
  activity_type: string
  min_deposit: number
  wager_multiplier: number
  currency?: string
  progress_count: number
  next_tier: number
  hidden: boolean
  all_wager_unlocked: boolean
  remaining_wager: number
  remaining_wager_u?: number
  latest_reward_amount: number
  latest_reward_amount_u?: number
  tiers: NewUserRechargeTierStatus[]
}

interface UseNewUserRechargeStatusResult {
  status: NewUserRechargeStatus | null
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

export function useNewUserRechargeStatus(enabled = true): UseNewUserRechargeStatusResult {
  const { currencyCode } = useExchangeRate()
  const [status, setStatus] = useState<NewUserRechargeStatus | null>(null)
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
      const data = await activityService.getNewUserRechargeStatus(currencyCode)
      setStatus(data as NewUserRechargeStatus)
    } catch (fetchError) {
      setError(getErrorMessage(fetchError, "failed to load new user recharge status"))
      setStatus(null)
    } finally {
      setLoading(false)
    }
  }, [currencyCode, enabled])

  useEffect(() => {
    queueMicrotask(() => {
      void fetchStatus()
    })
  }, [fetchStatus])

  return { status, loading, error, refresh: fetchStatus }
}
