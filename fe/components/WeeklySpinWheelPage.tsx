"use client"

import React, { useEffect, useMemo, useState } from "react"
import { Loader2, RefreshCcw, Ticket, Trophy, X } from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { activityService, GAME_TRANSACTIONS_UPDATED_EVENT } from "@/services/api"

interface WeeklySpinWheelTicketItem {
  ticket_code: string
  awarded_at: string
  bet_amount_u: number
  is_daily_qualified: boolean
  is_today: boolean
  is_winner: boolean
  prize_tier?: string
  prize_rate?: number
  prize_amount_u?: number
}

interface WeeklySpinWheelDrawWinner {
  ticket_code: string
  prize_tier: string
  prize_rate: number
  prize_amount_u: number
}

interface WeeklySpinWheelLatestDraw {
  cycle_start: string
  cycle_end: string
  drawn_at: string
  total_tickets: number
  winners: WeeklySpinWheelDrawWinner[]
}

interface WeeklySpinWheelStatus {
  activity_type: string
  server_time_utc: string
  current_cycle_start: string
  current_cycle_end: string
  next_daily_reset_at: string
  next_weekly_draw_at: string
  display_currency: "IDR" | "PHP"
  today_bet_amount_u: number
  today_bet_amount_idr: number
  today_bet_amount_php: number
  today_bet_amount: number
  base_ticket_threshold_u: number
  extra_ticket_threshold_u: number
  base_ticket_threshold: number
  extra_ticket_threshold: number
  base_ticket_qualified: boolean
  extra_tickets_today: number
  current_cycle_tickets: number
  total_prize_pool_u: number
  total_prize_pool: number
  total_prize_pool_idr: number
  total_prize_pool_php: number
  hundred_u_idr: number
  hundred_u_php: number
  thousand_u_idr: number
  thousand_u_php: number
  tickets: WeeklySpinWheelTicketItem[]
  latest_draw?: WeeklySpinWheelLatestDraw | null
}

const BASE_TICKET_THRESHOLD_U = 100
const EXTRA_TICKET_THRESHOLD_U = 1000
const TOTAL_PRIZE_POOL_U = 20000
const USE_LOCAL_PREVIEW_DATA = false

const createPreviewStatus = (displayCurrency: "IDR" | "PHP"): WeeklySpinWheelStatus => ({
  activity_type: "weekly_spin_wheel",
  server_time_utc: "2026-06-12 06:37:56",
  current_cycle_start: "2026-06-07 16:00:00",
  current_cycle_end: "2026-06-14 16:00:00",
  next_daily_reset_at: "2026-06-12 16:00:00",
  next_weekly_draw_at: "2026-06-14 16:00:00",
  display_currency: displayCurrency,
  today_bet_amount_u: 0.4,
  today_bet_amount_idr: 6400,
  today_bet_amount_php: 22.4,
  today_bet_amount: displayCurrency === "IDR" ? 6400 : 22.4,
  base_ticket_threshold_u: BASE_TICKET_THRESHOLD_U,
  extra_ticket_threshold_u: EXTRA_TICKET_THRESHOLD_U,
  base_ticket_threshold: displayCurrency === "IDR" ? 1600000 : 5600,
  extra_ticket_threshold: displayCurrency === "IDR" ? 16000000 : 56000,
  base_ticket_qualified: true,
  extra_tickets_today: 0,
  current_cycle_tickets: 0,
  total_prize_pool_u: TOTAL_PRIZE_POOL_U,
  total_prize_pool: displayCurrency === "IDR" ? 320000000 : 1120000,
  total_prize_pool_idr: 320000000,
  total_prize_pool_php: 1120000,
  hundred_u_idr: 1600000,
  hundred_u_php: 5600,
  thousand_u_idr: 16000000,
  thousand_u_php: 56000,
  tickets: [
    {
      ticket_code: "WSW24061201",
      awarded_at: "2026-06-10 09:21:00",
      bet_amount_u: 100,
      is_daily_qualified: true,
      is_today: true,
      is_winner: false,
    },
    {
      ticket_code: "WSW24061202",
      awarded_at: "2026-06-11 12:08:00",
      bet_amount_u: 1000,
      is_daily_qualified: false,
      is_today: false,
      is_winner: false,
    },
  ],
  latest_draw: {
    cycle_start: "2026-06-01 16:00:00",
    cycle_end: "2026-06-08 16:00:00",
    drawn_at: "2026-06-08 16:05:00",
    total_tickets: 100,
    winners: [
      {
        ticket_code: "WSW24060888",
        prize_tier: "1st",
        prize_rate: 30,
        prize_amount_u: 6000,
      },
      {
        ticket_code: "WSW24060852",
        prize_tier: "2nd",
        prize_rate: 20,
        prize_amount_u: 4000,
      },
      {
        ticket_code: "WSW24060819",
        prize_tier: "3rd",
        prize_rate: 15,
        prize_amount_u: 3000,
      },
    ],
  },
})

const WeeklySpinWheelPage: React.FC = () => {
  const { language } = useLanguage()
  const isIndonesian = language === "id"
  const currencyCode = isIndonesian ? "IDR" : "PHP"

  const copy = useMemo(
    () =>
      isIndonesian
        ? {
          loading: "Memuat aktivitas...",
          viewTickets: "Lihat Kode",
          ticketTitle: "Kode tiket yang didapat",
          drawTitle: "Daftar pemenang putaran ini",
          drawButton: "Lihat pemenang terbaru",
          emptyTickets: "Belum ada tiket minggu ini.",
          emptyDraw: "Belum ada pemenang untuk putaran sebelumnya.",
          code: "Kode",
          awardedAt: "Waktu didapat",
          prizeTier: "Peringkat",
          prizeAmount: "Hadiah",
          close: "Tutup",
          todayBet: "Taruhan valid hari ini",
          currentTickets: "Tiket minggu ini",
          baseStatus: "Status 100U",
          qualified: "Sudah memenuhi",
          notQualified: "Belum memenuhi",
          extraTickets: "Tiket tambahan hari ini",
          nextReset: "Reset harian UTC+8",
          nextDraw: "Undian mingguan UTC+8",
          cycleRange: "Siklus saat ini",
          ticketHint: "Ikon cek hanya muncul untuk tiket ambang 100U yang didapat hari ini.",
          drawHint: "Setelah undian selesai, seluruh kode lain dari putaran tersebut otomatis tidak berlaku.",
          drawTotal: "Total tiket putaran",
          refreshedAt: "Diperbarui",
          refresh: "Refresh",
          weeklyEvent: "Acara mingguan",
          earnTicketsTitle: "Kumpulkan tiket",
          earnTicketsDesc: "Semakin besar taruhan validmu, semakin banyak tiket yang bisa dikumpulkan untuk putaran minggu ini.",
          weeklyDrawTitle: "Undian mingguan",
          weeklyDrawDesc: "Setiap Senin pukul 12:00 (UTC), semua tiket yang valid akan ikut dalam pengundian hadiah bertingkat.",
          howToEarnTitle: "Cara mendapatkan tiket",
          firstBetTitle: "Taruhan harian pertama",
          firstBetDesc: "Capai taruhan valid harian pertama untuk mendapatkan 1 tiket dasar.",
          firstBetSubdesc: "Kamu bisa memperoleh hingga 7 tiket dasar setiap minggu.",
          everyBetTitle: "Taruhan tambahan",
          everyBetDesc: "Setelah ambang pertama tercapai, setiap taruhan valid tambahan memberi 1 tiket lagi.",
          everyBetSubdesc: "Tidak ada batas jumlah tiket tambahan selama siklus berjalan.",
          rewardText: "+1 tiket",
          totalPrizePoolTitle: "Total prize pool",
          totalPrizePoolHint: "Kolam hadiah mingguan tetap",
          winnersPerDrawTitle: "Pemenang per undian",
          winnersPerDrawHint: "Hadiah dibagikan berdasarkan tier",
          prizeDistributionTitle: "Distribusi hadiah berdasarkan tier",
          winnersColumn: "Jumlah pemenang",
          percentageColumn: "Persentase hadiah",
          eventNotesTitle: "Catatan acara",
        }
        : {
          loading: "Loading activity...",
          viewTickets: "View Tickets",
          ticketTitle: "Current earned ticket codes",
          drawTitle: "Latest draw winning tickets",
          drawButton: "View Latest Winners",
          emptyTickets: "No tickets earned this week yet.",
          emptyDraw: "No completed draw from the previous round yet.",
          code: "Code",
          awardedAt: "Awarded At",
          prizeTier: "Tier",
          prizeAmount: "Prize",
          close: "Close",
          todayBet: "Today's valid bet",
          currentTickets: "Weekly tickets",
          baseStatus: "100U status",
          qualified: "Qualified",
          notQualified: "Not yet",
          extraTickets: "Extra tickets today",
          nextReset: "Daily reset UTC+8",
          nextDraw: "Weekly draw UTC+8",
          cycleRange: "Current cycle",
          ticketHint: "The check icon only appears on today's 100U threshold ticket.",
          drawHint: "After the draw, all remaining ticket codes from that round are automatically voided.",
          drawTotal: "Total tickets in round",
          refreshedAt: "Refreshed",
          refresh: "Refresh",
          weeklyEvent: "Weekly event",
          earnTicketsTitle: "Earn tickets",
          earnTicketsDesc: "The more valid bets you place, the more entries you unlock for this week's lucky draw.",
          weeklyDrawTitle: "Weekly draw",
          weeklyDrawDesc: "Every Monday at 12:00 (UTC), all valid tickets join a tiered prize draw for the current cycle.",
          howToEarnTitle: "How to earn tickets",
          firstBetTitle: "First daily bet target",
          firstBetDesc: "Hit the first daily valid-bet milestone to receive your base ticket.",
          firstBetSubdesc: "You can earn up to 7 base tickets during a weekly cycle.",
          everyBetTitle: "Additional betting",
          everyBetDesc: "After the first milestone, each extra valid-bet threshold grants another ticket.",
          everyBetSubdesc: "There is no cap on extra tickets while the round remains active.",
          rewardText: "+1 ticket",
          totalPrizePoolTitle: "Total prize pool",
          totalPrizePoolHint: "Guaranteed weekly pool",
          winnersPerDrawTitle: "Winners per draw",
          winnersPerDrawHint: "Prizes distributed by tiers",
          prizeDistributionTitle: "Prize distribution by tiers",
          winnersColumn: "Number of winners",
          percentageColumn: "Prize percentage",
          eventNotesTitle: "Event notes",
        },
    [isIndonesian],
  )

  const [status, setStatus] = useState<WeeklySpinWheelStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [ticketModalOpen, setTicketModalOpen] = useState(false)
  const [drawModalOpen, setDrawModalOpen] = useState(false)

  useEffect(() => {
    let cancelled = false

    const loadStatus = async (showInitialLoader = false) => {
      if (USE_LOCAL_PREVIEW_DATA) {
        setStatus(createPreviewStatus(currencyCode))
        setError(null)
        setLoading(false)
        setRefreshing(false)
        return
      }

      if (showInitialLoader) {
        setLoading(true)
      } else {
        setRefreshing(true)
      }

      try {
        const nextStatus = await activityService.getWeeklySpinWheelStatus(currencyCode)
        if (!cancelled) {
          setStatus(nextStatus)
          setError(null)

          if (nextStatus?.latest_draw?.cycle_end) {
            const storageKey = `weekly-spin-wheel-draw-dismissed:${nextStatus.latest_draw.cycle_end}`
            if (window.localStorage.getItem(storageKey) !== "1") {
              setDrawModalOpen(true)
            }
          }
        }

        void activityService.getPlayerTransactionOverview(true, currencyCode)
          .then(() => activityService.getWeeklySpinWheelStatus(currencyCode))
          .then((syncedStatus) => {
            if (!cancelled) {
              setStatus(syncedStatus)
              setError(null)
            }
          })
          .catch((syncError) => {
            console.error("Weekly spin wheel source sync failed:", syncError)
          })
      } catch (fetchError) {
        if (!cancelled) {
          setError(fetchError instanceof Error ? fetchError.message : "Failed to load activity")
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    }

    void loadStatus(true)

    const handleGameTransactionsUpdated = () => {
      void loadStatus(false)
    }
    window.addEventListener(GAME_TRANSACTIONS_UPDATED_EVENT, handleGameTransactionsUpdated)

    return () => {
      cancelled = true
      window.removeEventListener(GAME_TRANSACTIONS_UPDATED_EVENT, handleGameTransactionsUpdated)
    }
  }, [currencyCode])

  useEffect(() => {
    if (!ticketModalOpen && !drawModalOpen) {
      return
    }

    const previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = "hidden"

    return () => {
      document.body.style.overflow = previousBodyOverflow
    }
  }, [ticketModalOpen, drawModalOpen])

  const dismissDrawModal = () => {
    if (status?.latest_draw?.cycle_end) {
      window.localStorage.setItem(`weekly-spin-wheel-draw-dismissed:${status.latest_draw.cycle_end}`, "1")
    }
    setDrawModalOpen(false)
  }

  const handleRefresh = async () => {
    if (refreshing) {
      return
    }

    if (USE_LOCAL_PREVIEW_DATA) {
      setStatus(createPreviewStatus(currencyCode))
      setError(null)
      return
    }

    setRefreshing(true)
    try {
      const currentStatus = await activityService.getWeeklySpinWheelStatus(currencyCode)
      setStatus(currentStatus)
      setError(null)

      try {
        await activityService.getPlayerTransactionOverview(true, currencyCode)
      } catch (syncError) {
        console.error("Weekly spin wheel source sync failed:", syncError)
      }

      const nextStatus = await activityService.getWeeklySpinWheelStatus(currencyCode)
      setStatus(nextStatus)
      setError(null)
    } catch (fetchError) {
      setError(fetchError instanceof Error ? fetchError.message : "Failed to load activity")
    } finally {
      setRefreshing(false)
    }
  }

  const formatU = (amount: number) =>
    amount.toLocaleString(isIndonesian ? "id-ID" : "en-US", {
      minimumFractionDigits: amount % 1 === 0 ? 0 : 2,
      maximumFractionDigits: 2,
    })

  const formatUtc8DateTime = (value: string) => {
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

  const formatUtcDateParts = (value: string) => {
    const formatted = formatUtc8DateTime(value)
    const separators = [", ", " pukul "]
    for (const separator of separators) {
      if (formatted.includes(separator)) {
        const [datePart, timePart] = formatted.split(separator)
        return [datePart, timePart || ""] as const
      }
    }

    return [formatted, ""] as const
  }

  const showTodayBaseTicketBadge = !!status && status.base_ticket_qualified

  const tierRows = [
    { rank: "1", winners: "1", percentage: "30%" },
    { rank: "2", winners: "2", percentage: "20%" },
    { rank: "3", winners: "3", percentage: "15%" },
    { rank: "4-10", winners: "7", percentage: "10%" },
    { rank: "11-30", winners: "20", percentage: "15%" },
    { rank: "31-100", winners: "70", percentage: "10%" },
  ]

  const eventNotes = isIndonesian
    ? [
      "Tiket hanya berlaku untuk putaran mingguan yang sedang berjalan.",
      "Semua progres tiket akan dibersihkan setelah undian mingguan selesai.",
      "Tiket diberikan berdasarkan jumlah taruhan valid yang memenuhi syarat.",
      "Penyalahgunaan, kecurangan, atau multi-akun dapat membatalkan hadiah.",
    ]
    : [
      "Tickets are valid only for the current weekly draw cycle.",
      "All ticket progress resets automatically after the weekly draw is completed.",
      "Tickets are granted based on qualified valid-bet thresholds.", 
      "Fraud, abuse, or multi-account activity may result in prize cancellation.",
    ]

  const panelClass =
    "rounded-[20px] border border-lucky-gold/15 bg-[linear-gradient(180deg,rgba(14,18,28,0.96),rgba(8,11,19,0.94))] p-3 text-white shadow-[0_16px_40px_rgba(0,0,0,0.28)] backdrop-blur-sm sm:rounded-[24px] sm:p-4"
  const sectionCardClass =
    "rounded-[22px] border border-[#7b5a1e] bg-[linear-gradient(180deg,rgba(16,11,4,0.98),rgba(7,7,7,0.98))] shadow-[0_18px_54px_rgba(0,0,0,0.36)] sm:rounded-[26px]"
  if (loading && !status) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-gray-900 via-gray-800 to-gray-900 flex items-center justify-center px-4">
        <div className="flex items-center gap-3 rounded-2xl border border-lucky-gold/20 bg-black/40 px-5 py-4 text-lucky-gold">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>{copy.loading}</span>
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(255,205,84,0.12),_transparent_22%),linear-gradient(180deg,#192438_0%,#0d121d_40%,#07090d_100%)] px-3 py-3 sm:px-4 sm:py-4 lg:px-5 xl:px-6">
        <div className="mx-auto w-full max-w-[1180px]">
          <div className={`${sectionCardClass} relative overflow-hidden`}>
            <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_right,_rgba(255,201,93,0.16),_transparent_28%),radial-gradient(circle_at_bottom_left,_rgba(255,185,52,0.1),_transparent_24%)]" />
            <div className="pointer-events-none absolute inset-y-0 left-0 w-full bg-[linear-gradient(90deg,rgba(255,255,255,0.035)_1px,transparent_1px)] bg-[length:32px_100%] opacity-20" />

            <div className="relative p-3 sm:p-4 lg:p-5 xl:p-6">
              <div className="grid gap-3 xl:gap-4">
                <section className="relative overflow-hidden rounded-[20px] border border-lucky-gold/15 bg-[linear-gradient(180deg,rgba(8,8,8,0.92),rgba(15,10,3,0.92))] p-2.5 pb-3 shadow-[inset_0_1px_0_rgba(255,232,178,0.14)] sm:rounded-[22px] sm:p-4 xl:p-5">
                  <div className="grid min-h-[210px] items-center gap-3 sm:min-h-0 sm:grid-cols-[minmax(0,1fr)_minmax(220px,280px)] sm:gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,340px)] xl:gap-5">
                    <div className="relative z-10 pr-[82px] sm:pr-0">
                      <div className="max-w-[225px] sm:max-w-[320px]">
                        <img
                          src={isIndonesian ? "/images/activity2/img/yinni1.png" : "/images/activity2/img/yingyu1.png"}
                          alt={isIndonesian ? "Hadiah Spin Mingguan" : "Weekly Spin Rewards"}
                          className="block h-auto w-full object-contain"
                        />
                      </div>
                      <div className="mt-1 max-w-[285px] sm:max-w-[360px]">
                        <img
                          src={isIndonesian ? "/images/activity2/img/yinni2.png" : "/images/activity2/img/yingyu2.png"}
                          alt={isIndonesian ? "Pasang taruhan, kumpulkan tiket, menangkan hadiah" : "Bet to earn tickets, spin to win big"}
                          className="block h-auto w-full object-contain opacity-95"
                        />
                      </div>

                      <div className="mt-2 grid grid-cols-2 gap-2">
                        <div className="rounded-[12px] border border-lucky-gold/20 bg-black/35 p-2 shadow-[inset_0_1px_0_rgba(255,236,194,0.08)] sm:rounded-[16px] sm:p-3">
                          <div className="flex items-start gap-2 sm:gap-2.5">
                            <img src="/images/activity2/img/2.png" alt="" className="mt-0.5 h-6 w-6 shrink-0 object-contain sm:h-9 sm:w-9" />
                            <div>
                              <p className="text-[10px] font-black uppercase tracking-[0.1em] text-lucky-gold sm:text-sm sm:tracking-[0.12em]">{copy.earnTicketsTitle}</p>
                              <p className="mt-1.5 hidden text-sm leading-relaxed text-white/76 sm:block">{copy.earnTicketsDesc}</p>
                            </div>
                          </div>
                        </div>

                        <div className="rounded-[12px] border border-lucky-gold/20 bg-black/35 p-2 shadow-[inset_0_1px_0_rgba(255,236,194,0.08)] sm:rounded-[16px] sm:p-3">
                          <div className="flex items-start gap-2 sm:gap-2.5">
                            <img src="/images/activity2/img/6.png" alt="" className="mt-0.5 h-6 w-6 shrink-0 object-contain sm:h-9 sm:w-9" />
                            <div>
                              <p className="text-[10px] font-black uppercase tracking-[0.1em] text-lucky-gold sm:text-sm sm:tracking-[0.12em]">{copy.weeklyDrawTitle}</p>
                              <p className="mt-1.5 hidden text-sm leading-relaxed text-white/76 sm:block">{copy.weeklyDrawDesc}</p>
                            </div>
                          </div>
                        </div>
                      </div>

                      <div className="mt-2 flex flex-wrap items-center gap-2">
                        <button
                          type="button"
                          onClick={() => setTicketModalOpen(true)}
                          className="inline-flex items-center gap-1.5 rounded-full border border-lucky-gold/40 bg-[linear-gradient(180deg,rgba(255,206,104,0.22),rgba(186,118,9,0.14))] px-2.5 py-1 text-[11px] font-bold text-lucky-gold shadow-[0_10px_24px_rgba(251,191,36,0.16)] transition-transform hover:scale-[1.02] sm:px-3 sm:py-1.5 sm:text-sm"
                        >
                          <Ticket className="h-3 w-3 sm:h-4 sm:w-4" />
                          {copy.viewTickets}
                        </button>
                        {status?.latest_draw && (
                          <button
                            type="button"
                            onClick={() => setDrawModalOpen(true)}
                            className="inline-flex items-center gap-1.5 rounded-full border border-white/12 bg-white/6 px-2.5 py-1 text-[11px] font-semibold text-white/80 transition-colors hover:bg-white/10 sm:px-3 sm:py-1.5 sm:text-sm"
                          >
                            <Trophy className="h-3 w-3 text-lucky-gold sm:h-4 sm:w-4" />
                            {copy.drawButton}
                          </button>
                        )}
                      
                      </div>
                    </div>

                    <div className="pointer-events-none absolute bottom-[30px] right-[-8px] z-0 flex w-[176px] items-center justify-center opacity-50 sm:relative sm:bottom-auto sm:right-auto sm:mx-auto sm:mt-0 sm:w-full sm:max-w-[270px] sm:opacity-100 lg:max-w-[330px]">
                      <div className="absolute inset-x-[18%] top-[14%] h-[20%] rounded-full bg-lucky-gold/18 blur-3xl" />
                      <img
                        src="/images/activity2/img/1.png"
                        alt="Lucky spin wheel"
                        className="relative z-10 block h-auto w-full object-contain drop-shadow-[0_22px_34px_rgba(0,0,0,0.52)]"
                      />
                    </div>
                  </div>
                </section>
              </div>

              <div className={`mt-3 ${sectionCardClass} overflow-hidden`}>
                <div className="grid gap-2.5 p-2.5 sm:gap-3 sm:p-4">
                  <div className="rounded-[18px] border border-lucky-gold/14 bg-[linear-gradient(180deg,rgba(11,9,6,0.96),rgba(4,4,4,0.96))] px-3 py-2.5 sm:rounded-[22px] sm:px-4 sm:py-3">
                    <div className="flex items-center gap-2.5">
                      <img src="/images/activity2/img/2.png" alt="" className="h-6 w-6 object-contain sm:h-8 sm:w-8" />
                      <h2 className="text-base font-black uppercase tracking-[0.16em] text-lucky-gold sm:text-xl">
                        {copy.howToEarnTitle}
                      </h2>
                    </div>

                    <div className="mt-2 grid gap-1.5">
                      <div className="border-t border-lucky-gold/10 pt-2 first:border-t-0 first:pt-0">
                        <div className="grid grid-cols-[34px_minmax(0,1fr)_42px_64px_76px] items-center gap-2 sm:grid-cols-[44px_minmax(0,1fr)_52px_92px_120px] sm:gap-3">
                          <img src="/images/activity2/img/7.png" alt="" className="h-8 w-8 object-contain sm:h-11 sm:w-11" />
                          <div className="flex min-w-0 items-center gap-1.5">
                            <p className="min-w-0 text-sm font-bold leading-snug text-white sm:text-base">{copy.firstBetTitle}</p>
                          </div>
                          <div className="flex justify-center">
                            {showTodayBaseTicketBadge && (
                              <img
                                src="/images/activity2/finish.png"
                                alt="Qualified today"
                                className="h-7 w-7 shrink-0 object-contain sm:h-9 sm:w-9"
                              />
                            )}
                          </div>
                          <div className="text-center text-lg font-black leading-none text-lucky-gold drop-shadow-[0_0_14px_rgba(251,191,36,0.28)] sm:text-2xl">
                            {formatU(status?.base_ticket_threshold_u ?? BASE_TICKET_THRESHOLD_U)}U
                          </div>
                          <div className="flex items-center justify-center gap-1 rounded-full border border-lucky-gold/22 bg-lucky-gold/8 px-1.5 py-1 text-center text-[10px] font-black uppercase leading-tight tracking-[0.04em] text-lucky-gold sm:gap-1.5 sm:px-3 sm:py-1.5 sm:text-sm sm:tracking-[0.12em]">
                            <img src="/images/activity2/img/2.png" alt="" className="h-3.5 w-3.5 object-contain sm:h-4 sm:w-4" />
                            {copy.rewardText}
                          </div>
                        </div>
                        <div className="mt-1 pl-[42px] pr-1 sm:pl-[56px]">
                          <p className="text-xs leading-relaxed text-white/72 sm:text-sm">{copy.firstBetDesc}</p>
                          <p className="mt-0.5 text-xs leading-relaxed text-white/48 sm:text-sm">{copy.firstBetSubdesc}</p>
                        </div>
                      </div>

                      <div className="border-t border-lucky-gold/10 pt-2">
                        <div className="grid grid-cols-[34px_minmax(0,1fr)_42px_64px_76px] items-center gap-2 sm:grid-cols-[44px_minmax(0,1fr)_52px_92px_120px] sm:gap-3">
                          <img src="/images/activity2/img/4.png" alt="" className="h-8 w-8 object-contain sm:h-11 sm:w-11" />
                          <p className="min-w-0 text-sm font-bold leading-snug text-white sm:text-base">{copy.everyBetTitle}</p>
                          <div />
                          <div className="text-center text-lg font-black leading-none text-lucky-gold drop-shadow-[0_0_14px_rgba(251,191,36,0.28)] sm:text-2xl">
                            {formatU(status?.extra_ticket_threshold_u ?? EXTRA_TICKET_THRESHOLD_U)}U
                          </div>
                          <div className="flex items-center justify-center gap-1 rounded-full border border-lucky-gold/22 bg-lucky-gold/8 px-1.5 py-1 text-center text-[10px] font-black uppercase leading-tight tracking-[0.04em] text-lucky-gold sm:gap-1.5 sm:px-3 sm:py-1.5 sm:text-sm sm:tracking-[0.12em]">
                            <img src="/images/activity2/img/2.png" alt="" className="h-3.5 w-3.5 object-contain sm:h-4 sm:w-4" />
                            {copy.rewardText}
                          </div>
                        </div>
                        <div className="mt-1 pl-[42px] pr-1 sm:pl-[56px]">
                          <p className="text-xs leading-relaxed text-white/72 sm:text-sm">{copy.everyBetDesc}</p>
                          <p className="mt-0.5 text-xs leading-relaxed text-white/48 sm:text-sm">{copy.everyBetSubdesc}</p>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="rounded-[18px] border border-lucky-gold/14 bg-[linear-gradient(180deg,rgba(11,9,6,0.96),rgba(4,4,4,0.96))] px-3 py-2.5 sm:rounded-[22px] sm:px-4 sm:py-3">
                    <div className="grid grid-cols-2 divide-x divide-lucky-gold/10">
                      <div className="flex items-center gap-2 pr-2 sm:gap-3 sm:pr-4">
                        <img src="/images/activity2/img/qiandai.png" alt="" className="h-9 w-9 shrink-0 object-contain sm:h-12 sm:w-12" />
                        <div className="min-w-0">
                          <p className="text-[10px] font-bold uppercase leading-tight tracking-[0.1em] text-white/45 sm:text-xs sm:tracking-[0.18em]">{copy.totalPrizePoolTitle}</p>
                          <p className="mt-0.5 text-xl font-black leading-none text-lucky-gold sm:text-[clamp(2rem,3vw,2.6rem)]">
                              {formatU(status?.total_prize_pool_u ?? TOTAL_PRIZE_POOL_U)}U
                            </p>
                          <p className="mt-1 hidden text-xs text-white/50 sm:block">{copy.totalPrizePoolHint}</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-2 pl-2 sm:gap-3 sm:pl-4">
                        <img src="/images/activity2/img/3.png" alt="" className="h-9 w-9 shrink-0 object-contain sm:h-12 sm:w-12" />
                        <div className="min-w-0">
                          <p className="text-[10px] font-bold uppercase leading-tight tracking-[0.1em] text-white/45 sm:text-xs sm:tracking-[0.18em]">{copy.winnersPerDrawTitle}</p>
                          <p className="mt-0.5 text-xl font-black leading-none text-lucky-gold sm:text-[clamp(2rem,3vw,2.6rem)]">100</p>
                          <p className="mt-1 hidden text-xs text-white/50 sm:block">{copy.winnersPerDrawHint}</p>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="overflow-hidden rounded-[18px] border border-lucky-gold/14 bg-[linear-gradient(180deg,rgba(11,9,6,0.96),rgba(4,4,4,0.96))] sm:rounded-[22px]">
                    <div className="border-b border-lucky-gold/12 px-3 py-2.5 sm:px-4 sm:py-3">
                      <h2 className="text-base font-black uppercase tracking-[0.14em] text-lucky-gold sm:text-xl sm:tracking-[0.18em]">
                        {copy.prizeDistributionTitle}
                      </h2>
                    </div>

                    <div className="p-2.5 sm:p-4">
                      <div className="overflow-hidden rounded-[16px] border border-lucky-gold/14 bg-black/30 sm:rounded-[20px]">
                        <div className="grid grid-cols-[0.8fr_1fr_0.9fr] border-b border-lucky-gold/14 bg-lucky-gold/8 px-2 py-2 text-center text-[10px] font-bold uppercase tracking-[0.1em] text-lucky-gold sm:px-4 sm:py-2.5 sm:text-xs">
                          <div>{copy.prizeTier}</div>
                          <div>{copy.winnersColumn}</div>
                          <div>{copy.percentageColumn}</div>
                        </div>
                        <div>
                          {tierRows.map((row) => (
                            <div
                              key={row.rank}
                              className="grid grid-cols-[0.8fr_1fr_0.9fr] border-t border-white/6 px-2 py-2 text-center text-sm text-white/82 first:border-t-0 sm:px-4 sm:py-2.5"
                            >
                              <div className="font-semibold text-lucky-gold">{row.rank}</div>
                              <div>{row.winners}</div>
                              <div>{row.percentage}</div>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div className={`mt-3 ${sectionCardClass} p-2.5 sm:p-4`}>
                <div className="grid gap-2.5 lg:grid-cols-[minmax(0,1fr)_200px]">
                  <div>
                    <div className="flex items-center gap-2.5">
                      <img src="/images/activity2/img/6.png" alt="" className="h-7 w-7 object-contain sm:h-8 sm:w-8" />
                      <h2 className="text-base font-black uppercase tracking-[0.14em] text-lucky-gold sm:text-xl sm:tracking-[0.16em]">
                        {copy.eventNotesTitle}
                      </h2>
                    </div>

                    <div className="mt-2.5 grid gap-2">
                      {eventNotes.map((note, index) => (
                        <div key={note} className="flex items-start gap-2.5 rounded-[14px] border border-white/6 bg-black/26 px-3 py-2 text-xs leading-relaxed text-white/72 sm:rounded-[18px] sm:px-4 sm:py-2.5 sm:text-sm">
                          <span className="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-lucky-gold/24 bg-lucky-gold/10 text-[10px] font-black text-lucky-gold sm:h-6 sm:w-6 sm:text-xs">
                            {index + 1}
                          </span>
                          <span>{note}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="hidden items-end justify-center lg:flex">
                    <img src="/images/activity2/img/5.png" alt="" className="h-auto w-full max-w-[160px] object-contain opacity-95" />
                  </div>
                </div>
              </div>

              {status && (
                <aside className="mt-3 grid gap-2.5 sm:grid-cols-2">
                  <div className={panelClass}>
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-[10px] uppercase tracking-[0.18em] text-white/45 sm:text-xs sm:tracking-[0.22em]">{copy.todayBet}</p>
                        <p className="mt-1.5 text-3xl font-black leading-none text-lucky-gold sm:text-[clamp(2rem,4vw,3rem)]">
                          {formatU(status.today_bet_amount_u)}U
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setTicketModalOpen(true)}
                        className="inline-flex shrink-0 items-center gap-2 rounded-full border border-lucky-gold/25 bg-lucky-gold/10 px-3 py-1.5 text-xs font-semibold text-lucky-gold transition-colors hover:bg-lucky-gold/15"
                      >
                        <Ticket className="h-3.5 w-3.5" />
                        {status.current_cycle_tickets}
                      </button>
                    </div>

                    <div className="mt-3 grid gap-1.5 text-sm text-white/80 sm:mt-5 sm:gap-2">
                      <div className="flex items-start justify-between gap-4">
                        <span className="text-white/65">{copy.baseStatus}</span>
                        <span className={status.base_ticket_qualified ? "font-semibold text-emerald-300" : "font-semibold text-white/55"}>
                          {status.base_ticket_qualified ? copy.qualified : copy.notQualified}
                        </span>
                      </div>
                      <div className="flex items-start justify-between gap-4">
                        <span className="text-white/65">{copy.extraTickets}</span>
                        <span className="font-semibold text-white">{status.extra_tickets_today}</span>
                      </div>
                      <p className="pt-1 text-xs leading-relaxed text-white/50">{copy.ticketHint}</p>
                    </div>
                  </div>

                  <div className={panelClass}>
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-[10px] uppercase tracking-[0.18em] text-white/45 sm:text-xs sm:tracking-[0.22em]">{copy.currentTickets}</p>
                        <p className="mt-1.5 text-3xl font-black leading-none text-lucky-gold sm:text-[clamp(2rem,4vw,3rem)]">
                          {status.current_cycle_tickets}
                        </p>
                      </div>
                      <div className="flex items-center rounded-full border border-lucky-gold/20 bg-lucky-gold/10 px-3 py-1.5 text-xs font-bold uppercase tracking-[0.18em] text-lucky-gold">
                        <img src="/images/activity2/img/3.png" alt="" className="mr-2 h-4 w-4 object-contain" />
                        100
                      </div>
                    </div>

                    <div className="mt-3 grid gap-1.5 text-sm text-white/80 sm:mt-5 sm:gap-2">
                      <div className="flex items-start justify-between gap-4">
                        <span className="max-w-[8rem] text-white/65">{copy.nextReset}</span>
                        <span className="text-right font-medium text-white">{formatUtc8DateTime(status.next_daily_reset_at)}</span>
                      </div>
                      <div className="flex items-start justify-between gap-4">
                        <span className="max-w-[8rem] text-white/65">{copy.nextDraw}</span>
                        <span className="text-right font-medium text-white">{formatUtc8DateTime(status.next_weekly_draw_at)}</span>
                      </div>
                      <div className="flex items-start justify-between gap-4">
                        <span className="max-w-[8rem] text-white/65">{copy.refreshedAt}</span>
                        <span className="text-right font-medium text-white/80">{formatUtc8DateTime(status.server_time_utc)}</span>
                      </div>
                    </div>
                  </div>

                  <div className={`${panelClass} sm:col-span-2`}>
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="text-[10px] uppercase tracking-[0.18em] text-white/45 sm:text-xs sm:tracking-[0.22em]">{copy.cycleRange}</p>
                        <p className="mt-1.5 text-sm font-semibold leading-relaxed text-white sm:text-lg">
                          {formatUtc8DateTime(status.current_cycle_start)} - {formatUtc8DateTime(status.current_cycle_end)}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={handleRefresh}
                        disabled={refreshing}
                        className="inline-flex shrink-0 items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-semibold text-white/75 transition-colors hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        <RefreshCcw className={`h-3.5 w-3.5 ${refreshing ? "animate-spin" : ""}`} />
                        {copy.refresh}
                      </button>
                    </div>
                    <p className="mt-2 max-w-3xl text-sm leading-relaxed text-white/55 sm:mt-3 sm:text-base">{copy.drawHint}</p>
                  </div>

                  {error && (
                    <div className="sm:col-span-2 rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
                      {error}
                    </div>
                  )}
                </aside>
              )}

              {error && !status && (
                <div className="mt-4 rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
                  {error}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {ticketModalOpen && (
        <div className="fixed inset-0 z-[100001] flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm">
          <div className="relative w-full max-w-xl min-h-[460px] rounded-[26px] border-4 border-[#ff5b5b] bg-white p-5 shadow-2xl sm:min-h-[520px] sm:p-7">
            <button
              type="button"
              onClick={() => setTicketModalOpen(false)}
              className="absolute right-3 top-3 rounded-full p-1 text-[#ff5b5b] transition-colors hover:bg-[#fff1f1]"
            >
              <X className="h-5 w-5" />
            </button>

            <h3 className="text-center text-2xl font-bold text-[#ff4e4e] sm:text-3xl">{copy.ticketTitle}</h3>

            <div className="mt-6 overflow-hidden rounded-[20px] border-4 border-[#ff5b5b]">
              <div className="grid grid-cols-[1.1fr_1fr] bg-[#fff7f7] px-5 py-4 text-center text-xl font-semibold text-[#ff4e4e] sm:text-2xl">
                <div>{copy.code}</div>
                <div>{copy.awardedAt}</div>
              </div>

              <div className="max-h-[420px] min-h-[280px] overflow-y-auto bg-white px-5 py-5 sm:max-h-[460px] sm:min-h-[320px]">
                {status?.tickets?.length ? (
                  <div className="space-y-5">
                    {status.tickets.map((ticket) => (
                      <div key={`${ticket.ticket_code}-${ticket.awarded_at}`} className="grid grid-cols-[1.1fr_1fr] items-center gap-3 text-center text-[#ff5b5b]">
                        <div className="flex items-center justify-center gap-2 text-lg sm:text-2xl">
                          {showTodayBaseTicketBadge && ticket.is_daily_qualified && ticket.is_today && (
                            <img src="/images/activity2/finish.png" alt="Qualified" className="h-5 w-5 shrink-0 sm:h-6 sm:w-6" />
                          )}
                          <span className="font-medium tracking-wide">{ticket.ticket_code}</span>
                        </div>
                        <div className="text-base font-medium leading-snug sm:text-2xl">
                          <div>{formatUtcDateParts(ticket.awarded_at)[0]}</div>
                          <div>{formatUtcDateParts(ticket.awarded_at)[1]}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="py-10 text-center text-base text-[#ff7a7a] sm:text-lg">{copy.emptyTickets}</div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {drawModalOpen && (
        <div className="fixed inset-0 z-[100002] flex items-center justify-center bg-black/80 p-4 backdrop-blur-sm">
          <div className="relative w-full max-w-2xl rounded-[28px] border border-lucky-gold/30 bg-[#090909] p-5 shadow-2xl sm:p-6">
            <button
              type="button"
              onClick={dismissDrawModal}
              className="absolute right-4 top-4 rounded-full bg-white/5 p-2 text-white/60 transition-colors hover:bg-white/10 hover:text-white"
            >
              <X className="h-4 w-4" />
            </button>

            <div className="flex items-start gap-3">
              <div className="rounded-2xl bg-lucky-gold/15 p-3 text-lucky-gold">
                <Trophy className="h-5 w-5" />
              </div>
              <div>
                <p className="text-xl font-bold text-white sm:text-2xl">{copy.drawTitle}</p>
                {status?.latest_draw && (
                  <p className="mt-1 text-sm text-white/55">
                    {formatUtc8DateTime(status.latest_draw.cycle_start)} - {formatUtc8DateTime(status.latest_draw.cycle_end)}
                  </p>
                )}
              </div>
            </div>

            {status?.latest_draw ? (
              <>
                <div className="mt-5 grid gap-3 rounded-2xl border border-lucky-gold/15 bg-white/5 p-4 text-sm text-white/75 sm:grid-cols-2">
                  <div>
                    <p className="text-white/45">{copy.drawTotal}</p>
                    <p className="mt-1 text-lg font-bold text-white">{status.latest_draw.total_tickets}</p>
                  </div>
                  <div>
                    <p className="text-white/45">{copy.refreshedAt}</p>
                    <p className="mt-1 text-lg font-bold text-white">{formatUtc8DateTime(status.latest_draw.drawn_at)}</p>
                  </div>
                </div>

                <div className="mt-5 overflow-hidden rounded-2xl border border-lucky-gold/15">
                  <div className="grid grid-cols-[1.2fr_0.8fr_0.8fr] bg-lucky-gold/10 px-4 py-3 text-sm font-semibold uppercase tracking-[0.16em] text-lucky-gold">
                    <div>{copy.code}</div>
                    <div>{copy.prizeTier}</div>
                    <div>{copy.prizeAmount}</div>
                  </div>
                  <div className="max-h-[340px] overflow-y-auto">
                    {status.latest_draw.winners.length ? (
                      status.latest_draw.winners.map((winner) => (
                        <div
                          key={`${winner.ticket_code}-${winner.prize_tier}`}
                          className="grid grid-cols-[1.2fr_0.8fr_0.8fr] items-center border-t border-white/5 px-4 py-3 text-sm text-white/85"
                        >
                          <div className="font-semibold tracking-wide text-lucky-gold">{winner.ticket_code}</div>
                          <div>{winner.prize_tier}</div>
                          <div>{formatU(winner.prize_amount_u)}U</div>
                        </div>
                      ))
                    ) : (
                      <div className="px-4 py-10 text-center text-sm text-white/55">{copy.emptyDraw}</div>
                    )}
                  </div>
                </div>
              </>
            ) : (
              <div className="mt-6 rounded-2xl border border-white/10 bg-white/5 px-4 py-6 text-center text-sm text-white/55">
                {copy.emptyDraw}
              </div>
            )}

            <button
              type="button"
              onClick={dismissDrawModal}
              className="mt-6 w-full rounded-2xl bg-gradient-to-r from-lucky-gold to-orange-500 px-4 py-3 text-sm font-bold text-black transition-transform hover:scale-[1.01]"
            >
              {copy.close}
            </button>
          </div>
        </div>
      )}
    </>
  )
}

export default WeeklySpinWheelPage
