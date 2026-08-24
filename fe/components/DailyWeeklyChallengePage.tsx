"use client"

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2, RefreshCcw, Target, Trophy } from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { activityService, GAME_TRANSACTIONS_UPDATED_EVENT } from "@/services/api"

interface DailyWeeklyChallengeTask {
  cycle_type: "daily" | "weekly"
  task_index: number
  required_bet_u: number
  reward_amount_u: number
  required_bet_idr: number
  reward_amount_idr: number
  required_bet_php: number
  reward_amount_php: number
  current_bet_u: number
  current_bet_idr: number
  current_bet_php: number
  display_currency: "IDR" | "PHP"
  required_bet: number
  reward_amount: number
  current_bet: number
  progress_ratio: number
  claimed: boolean
  claimable: boolean
  claimed_at?: string
  period_start_utc: string
  period_end_utc: string
}

interface DailyWeeklyChallengeStatus {
  activity_type: string
  server_time_utc: string
  daily_period_start: string
  daily_period_end: string
  weekly_period_start: string
  weekly_period_end: string
  next_daily_reset_at: string
  next_weekly_reset_at: string
  daily_tasks: DailyWeeklyChallengeTask[]
  weekly_tasks: DailyWeeklyChallengeTask[]
}

type TaskTone = "green" | "blue"

const DailyWeeklyChallengePage: React.FC = () => {
  const router = useRouter()
  const { language } = useLanguage()
  const isIndonesian = language === "id"
  const currencyCode = isIndonesian ? "IDR" : "PHP"

  const text = useMemo(
    () =>
      isIndonesian
        ? {
            loading: "Memuat aktivitas...",
            dailyTitle: "Tantangan Harian",
            weeklyTitle: "Tantangan Mingguan",
            dailyTaskLabel: "Harian",
            weeklyTaskLabel: "Mingguan",
            dailyReset: "Reset setiap hari 00:00 UTC+8",
            weeklyReset: "Reset setiap Senin 00:00 UTC+8",
            betLabel: "Taruhan",
            rewardLabel: "Hadiah",
            refresh: "Refresh",
            rulesTitle: "Penjelasan aktivitas",
            rules: [
              "Hanya taruhan valid olahraga dan e-sports yang dihitung.",
              "Setelah target tercapai, hadiah dikirim sebagai saldo bonus terikat.",
              "Saldo bonus terikat harus memenuhi syarat turnover sebelum bisa ditarik.",
              "Jika terdeteksi perilaku tidak normal, platform berhak membatalkan hadiah.",
            ],
            goAlt: "Selesaikan sekarang",
            claimAlt: "Klaim hadiah",
            claimedAlt: "Sudah diklaim",
            footerHint: "Semua target dan hadiah pada halaman ini ditampilkan dalam U.",
          }
        : {
            loading: "Loading activity...",
            dailyTitle: "Daily Challenge",
            weeklyTitle: "Weekly Challenge",
            dailyTaskLabel: "Daily",
            weeklyTaskLabel: "Weekly",
            dailyReset: "Resets daily at 00:00 UTC+8",
            weeklyReset: "Resets every Monday at 00:00 UTC+8",
            betLabel: "Bet",
            rewardLabel: "Reward",
            refresh: "Refresh",
            rulesTitle: "Activity rules",
            rules: [
              "Only valid sports and e-sports bets are counted.",
              "After completion, rewards are issued as locked bonus balance.",
              "Locked bonus balance must clear the required turnover before withdrawal.",
              "The platform may cancel rewards in case of abnormal activity.",
            ],
            goAlt: "Go complete",
            claimAlt: "Claim reward",
            claimedAlt: "Claimed",
            footerHint: "All target and reward amounts on this page are displayed in U.",
          },
    [isIndonesian],
  )

  const [status, setStatus] = useState<DailyWeeklyChallengeStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [submittingKey, setSubmittingKey] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const loadSequenceRef = useRef(0)

  const syncChallengeSourceData = useCallback(async () => {
    await activityService.getPlayerTransactionOverview(true, currencyCode)
  }, [currencyCode])

  const loadStatus = useCallback(async (showLoader = false) => {
    const sequence = ++loadSequenceRef.current

    if (showLoader) {
      setLoading(true)
    } else {
      setRefreshing(true)
    }

    try {
      const nextStatus = await activityService.getDailyWeeklyChallengeStatus(currencyCode)
      if (loadSequenceRef.current === sequence) {
        setStatus(nextStatus)
        setError(null)
      }
    } catch (fetchError) {
      if (loadSequenceRef.current === sequence) {
        setError(fetchError instanceof Error ? fetchError.message : "Failed to load activity")
      }
    } finally {
      if (loadSequenceRef.current === sequence) {
        setLoading(false)
        setRefreshing(false)
      }
    }

    void (async() => {
      try {
        await syncChallengeSourceData()
        const refreshedStatus = await activityService.getDailyWeeklyChallengeStatus(currencyCode)
        if (loadSequenceRef.current === sequence) {
          setStatus(refreshedStatus)
          setError(null)
        }
      } catch (syncError) {
        console.error("Daily weekly challenge source sync failed:", syncError)
      }
    })()
  }, [currencyCode, syncChallengeSourceData])

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadStatus(true)
    }, 0)

    const handleGameTransactionsUpdated = () => {
      void loadStatus(false)
    }
    window.addEventListener(GAME_TRANSACTIONS_UPDATED_EVENT, handleGameTransactionsUpdated)

    return () => {
      window.clearTimeout(timer)
      window.removeEventListener(GAME_TRANSACTIONS_UPDATED_EVENT, handleGameTransactionsUpdated)
      loadSequenceRef.current += 1
    }
  }, [loadStatus])

  const formatUTC8 = (value: string) => {
    const parsed = new Date(value.replace(" ", "T") + "Z")
    if (Number.isNaN(parsed.getTime())) {
      return value
    }

    return parsed.toLocaleString(isIndonesian ? "id-ID" : "en-US", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
      timeZone: "Asia/Shanghai",
    })
  }

  const formatU = (amount: number) =>
    `${amount.toLocaleString(isIndonesian ? "id-ID" : "en-US", {
      minimumFractionDigits: amount % 1 === 0 ? 0 : 2,
      maximumFractionDigits: 2,
    })}U`

  const getTaskTone = (taskIndex: number): TaskTone => (taskIndex === 1 ? "green" : "blue")

  const getGoButtonImage = (tone: TaskTone) => {
    if (isIndonesian) {
      return tone === "green" ? "/images/activity4/yin_go.png" : "/images/activity4/yin_go2.png"
    }

    return tone === "green" ? "/images/activity4/ying_go.png" : "/images/activity4/ying_go2.png"
  }

  const getClaimButtonImage = () =>
    isIndonesian ? "/images/activity4/yin_lingqu.png" : "/images/activity4/ying_lingqu.png"

  const getClaimedButtonImage = () =>
    isIndonesian ? "/images/activity4/yin_yilingqu.png" : "/images/activity4/ying_yilingqu.png"

  const handleTaskAction = async (task: DailyWeeklyChallengeTask) => {
    const taskKey = `${task.cycle_type}-${task.task_index}`
    if (submittingKey === taskKey) {
      return
    }

    if (!task.claimable) {
      router.push("/home")
      return
    }

    setSubmittingKey(taskKey)
    setError(null)

    try {
      try {
        await syncChallengeSourceData()
      } catch (syncError) {
        console.error("Daily weekly challenge source sync failed:", syncError)
      }

      await activityService.claimDailyWeeklyChallenge(task.cycle_type, task.task_index)
      await loadStatus(false)
    } catch (claimError) {
      setError(claimError instanceof Error ? claimError.message : "Failed to claim reward")
    } finally {
      setSubmittingKey(null)
    }
  }

  const renderProgressBar = (task: DailyWeeklyChallengeTask) => {
    const tone = getTaskTone(task.task_index)
    const fillImage = tone === "green" ? "/images/activity4/pro1.png" : "/images/activity4/pro2.png"
    const progressRatio = Math.max(0, Math.min(1, task.progress_ratio))
    const percentage = Math.round(progressRatio * 100)
    const requiredAmount = task.required_bet_u
    const currentAmount = Math.min(task.current_bet_u, requiredAmount)

    return (
      <div className="space-y-1.5">
        <div className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2 text-[10px] font-semibold text-white sm:text-[11px]">
          <span className="min-w-0 whitespace-nowrap leading-tight">
            {formatU(currentAmount)} / {formatU(requiredAmount)}
          </span>
          <span className="shrink-0 leading-tight">{percentage}%</span>
        </div>
        <div className="relative h-[10px] w-full overflow-hidden rounded-full">
          <img src="/images/activity4/probg.png" alt="" className="absolute inset-0 h-full w-full" />
          <div className="absolute inset-y-0 left-0 overflow-hidden" style={{ width: `${percentage}%` }}>
            <img src={fillImage} alt="" className="h-full w-full object-fill" />
          </div>
        </div>
      </div>
    )
  }

  const renderTaskBadge = (task: DailyWeeklyChallengeTask) => {
    const tone = getTaskTone(task.task_index)
    const isDaily = task.cycle_type === "daily"
    const Icon = isDaily ? Target : Trophy
    const frameClass =
      tone === "green"
        ? "border-[#8d6b1a] bg-[radial-gradient(circle_at_top,_rgba(255,211,96,0.18),_rgba(37,27,15,0.94)_68%)]"
        : "border-[#7d35df] bg-[radial-gradient(circle_at_top,_rgba(199,125,255,0.22),_rgba(33,17,40,0.94)_68%)]"
    const orbClass =
      tone === "green"
        ? "border-[#f0b54c] bg-[radial-gradient(circle,_rgba(255,233,154,0.95)_0%,_rgba(249,186,57,0.68)_45%,_rgba(118,71,8,0.88)_100%)] shadow-[0_0_24px_rgba(240,181,76,0.35)]"
        : "border-[#be7cff] bg-[radial-gradient(circle,_rgba(240,205,255,0.95)_0%,_rgba(190,124,255,0.68)_45%,_rgba(73,24,112,0.92)_100%)] shadow-[0_0_24px_rgba(190,124,255,0.35)]"
    const ringClass = tone === "green" ? "border-white/18" : "border-white/16"
    const iconClass = tone === "green" ? "text-[#4f2f00]" : "text-[#34104f]"

    return (
      <div className={`relative flex min-h-[72px] items-center justify-center overflow-hidden rounded-[18px] border ${frameClass}`}>
        <div
          className={`pointer-events-none absolute -right-3 -top-3 h-10 w-10 rounded-full blur-xl ${
            tone === "green" ? "bg-[#ffca54]/25" : "bg-[#b25cff]/25"
          }`}
        />
        <div
          className={`pointer-events-none absolute -bottom-3 -left-3 h-10 w-10 rounded-full blur-xl ${
            tone === "green" ? "bg-[#8a5806]/30" : "bg-[#5e1aa8]/30"
          }`}
        />
        <div className={`relative flex h-12 w-12 items-center justify-center rounded-full border-[3px] ${orbClass}`}>
          <div className={`absolute inset-[5px] rounded-full border ${ringClass}`} />
          <Icon className={`relative z-10 h-5 w-5 ${iconClass}`} strokeWidth={2.4} />
        </div>
      </div>
    )
  }

  const renderTaskCard = (task: DailyWeeklyChallengeTask) => {
    const tone = getTaskTone(task.task_index)
    const toneStyles =
      tone === "green"
        ? {
            border: "border-[#8d6b1a]",
            badge: "from-[#ffcf5f] to-[#c98a00]",
            reward: "text-[#a7f35e]",
            shadow: "shadow-[0_0_22px_rgba(245,190,74,0.12)]",
          }
        : {
            border: "border-[#7d35df]",
            badge: "from-[#c07cff] to-[#7f2dff]",
            reward: "text-[#a7f35e]",
            shadow: "shadow-[0_0_24px_rgba(153,92,255,0.14)]",
          }

    const buttonImage = task.claimed
      ? getClaimedButtonImage()
      : task.claimable
        ? getClaimButtonImage()
        : getGoButtonImage(tone)

    return (
      <div
        key={`${task.cycle_type}-${task.task_index}`}
        className={`rounded-[22px] border bg-[#15110d] p-2.5 sm:p-3 [@media(orientation:landscape)]:p-3.5 ${toneStyles.border} ${toneStyles.shadow} ${
          task.claimed ? "opacity-60 grayscale-[0.3]" : ""
        }`}
      >
        <div className="grid grid-cols-[104px_1fr] gap-2.5 sm:grid-cols-[116px_1fr] sm:gap-3 [@media(orientation:landscape)]:grid-cols-[124px_1fr] [@media(orientation:landscape)]:gap-4">
          <div className="flex flex-col gap-2">
            <div
              className={`inline-flex w-fit rounded-xl bg-gradient-to-r px-2.5 py-1.5 text-[11px] font-black text-black sm:text-xs [@media(orientation:landscape)]:px-3 [@media(orientation:landscape)]:text-[13px] ${toneStyles.badge}`}
            >
              {task.cycle_type === "daily" ? text.dailyTaskLabel : text.weeklyTaskLabel} {task.task_index}
            </div>
            {renderTaskBadge(task)}
          </div>

          <div className="flex min-w-0 flex-col justify-between gap-2">
            <div className="space-y-2">
              <div>
                <p className="text-[11px] text-white/70 sm:text-xs">{text.betLabel}</p>
                <p className="flex flex-wrap items-baseline gap-x-1 gap-y-0.5 text-[#f4c257]">
                  <span className="text-[15px] font-black leading-tight sm:text-[19px]">{formatU(task.required_bet_u)}</span>
                </p>
              </div>

              <div>
                <p className="text-[11px] text-white/70 sm:text-xs">{text.rewardLabel}</p>
                <p className={`flex flex-wrap items-baseline gap-x-1 gap-y-0.5 ${toneStyles.reward}`}>
                  <span className="text-[15px] font-black leading-tight sm:text-[19px]">{formatU(task.reward_amount_u)}</span>
                </p>
              </div>
            </div>

            {renderProgressBar(task)}

            <button
              type="button"
              onClick={() => void handleTaskAction(task)}
              disabled={submittingKey === `${task.cycle_type}-${task.task_index}`}
              className="transition-transform hover:scale-[1.01] active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-70 [@media(orientation:landscape)]:mx-auto [@media(orientation:landscape)]:w-full [@media(orientation:landscape)]:max-w-[320px]"
            >
              <img
                src={buttonImage}
                alt={task.claimed ? text.claimedAlt : task.claimable ? text.claimAlt : text.goAlt}
                className="h-auto w-full"
              />
            </button>
          </div>
        </div>
      </div>
    )
  }

  if (loading && !status) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-gray-950 via-black to-gray-900 px-4">
        <div className="flex items-center gap-3 rounded-2xl border border-lucky-gold/20 bg-black/40 px-5 py-4 text-lucky-gold">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>{text.loading}</span>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#050505] via-[#111111] to-[#050505] px-3 py-3 sm:px-5 sm:py-4 [@media(orientation:landscape)]:px-4 [@media(orientation:landscape)]:py-4">
      <div className="mx-auto w-full max-w-md sm:max-w-lg md:max-w-xl [@media(orientation:landscape)]:max-w-6xl">
        <div className="mt-2 space-y-3.5 [@media(orientation:landscape)]:grid [@media(orientation:landscape)]:grid-cols-2 [@media(orientation:landscape)]:items-start [@media(orientation:landscape)]:gap-4 [@media(orientation:landscape)]:space-y-0">
          <section className="rounded-[24px] border border-[#73531b] bg-[#0b0b0b] p-3 shadow-[0_18px_50px_rgba(0,0,0,0.4)] [@media(orientation:landscape)]:h-full [@media(orientation:landscape)]:p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <h2 className="text-xl font-black text-[#f4c257] sm:text-2xl">{text.dailyTitle}</h2>
                <p className="mt-0.5 text-xs text-white/55 sm:text-sm">{text.dailyReset}</p>
              </div>
              <button
                type="button"
                onClick={() => void loadStatus(false)}
                disabled={refreshing}
                className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/5 px-2.5 py-1 text-[11px] font-semibold text-white/75 transition-colors hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
              >
                <RefreshCcw className={`h-3 w-3 ${refreshing ? "animate-spin" : ""}`} />
                {text.refresh}
              </button>
            </div>
            <div className="space-y-2.5">{status?.daily_tasks.map(renderTaskCard)}</div>
          </section>

          <section className="rounded-[24px] border border-[#5c2bb0] bg-[#0b0b0b] p-3 shadow-[0_18px_50px_rgba(0,0,0,0.4)] [@media(orientation:landscape)]:h-full [@media(orientation:landscape)]:p-4">
            <div className="mb-3">
              <h2 className="text-xl font-black text-[#c184ff] sm:text-2xl">{text.weeklyTitle}</h2>
              <p className="mt-0.5 text-xs text-white/55 sm:text-sm">{text.weeklyReset}</p>
            </div>
            <div className="space-y-2.5">{status?.weekly_tasks.map(renderTaskCard)}</div>
          </section>

          <section className="rounded-[22px] border border-[#73531b] bg-[#121212] p-3 text-white/80 [@media(orientation:landscape)]:col-span-2 [@media(orientation:landscape)]:p-4">
            <h3 className="text-lg font-black text-[#f4c257] sm:text-xl">{text.rulesTitle}</h3>
            <ol className="mt-2.5 space-y-1.5 text-xs leading-relaxed sm:text-sm">
              {text.rules.map((rule, index) => (
                <li key={rule}>
                  {index + 1}. {rule}
                </li>
              ))}
            </ol>
            <p className="mt-3 text-[11px] leading-relaxed text-white/45 sm:text-xs">{text.footerHint}</p>
            {status && (
              <div className="mt-3 grid gap-1.5 rounded-2xl border border-white/10 bg-white/5 p-2.5 text-[11px] text-white/60 sm:grid-cols-2 sm:text-xs">
                <p>Daily UTC+8: {formatUTC8(status.next_daily_reset_at)}</p>
                <p>Weekly UTC+8: {formatUTC8(status.next_weekly_reset_at)}</p>
              </div>
            )}
          </section>

          {error && (
            <div className="rounded-2xl border border-red-500/30 bg-red-500/10 px-3 py-2.5 text-xs text-red-200 sm:text-sm [@media(orientation:landscape)]:col-span-2">
              {error}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default DailyWeeklyChallengePage
