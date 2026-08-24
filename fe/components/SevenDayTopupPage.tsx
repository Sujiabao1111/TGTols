"use client"

import React, { useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import {
  CalendarCheck,
  CheckCircle2,
  Crown,
  Gift,
  Loader2,
  LockKeyhole,
  RefreshCcw,
  TriangleAlert,
  X,
} from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { activityService } from "@/services/api"

interface PendingMakeupState {
  missed_date: string
  recharge_date: string
}

interface SevenDayTopupStatus {
  activity_type: string
  currency: string
  daily_threshold: number
  progress_dates: string[]
  progress_count: number
  used_makeup: boolean
  completed_round: boolean
  pending_makeup?: PendingMakeupState | null
  direct_invite_count: number
}

const ASSET_BASE = "/images/activity5"

const SevenDayTopupPage: React.FC = () => {
  const router = useRouter()
  const { language } = useLanguage()
  const { currencyCode } = useExchangeRate()

  const isIndonesian = language === "id"
  const titleImage = isIndonesian ? `${ASSET_BASE}/yinnititle.png` : `${ASSET_BASE}/yingyutitle.png`

  const text = useMemo(
    () =>
      isIndonesian
        ? {
            loading: "Memuat aktivitas...",
            heroAlt: "Isi ulang 7 hari menang iPhone",
            dailyTarget: "Target harian",
            dailyTargetHint: "Isi ulang 10U setiap hari untuk menyalakan 1 progres.",
            rewardHint: "Setiap isi ulang sukses akan mendapat hadiah sesuai aturan.",
            finalPrize: "Hari ke-7 ikut undian iPhone.",
            progressTitle: "Progres isi ulang 7 hari",
            loopHint: "Siklus berulang",
            progressHint: "Setelah hadiah tiap hari diklaim atau putaran terputus, progres akan masuk siklus baru.",
            completed: "Kesempatan undian iPhone sudah terbuka.",
            day: "Hari",
            unlocked: "Selesai",
            locked: "Terkunci",
            finalDay: "Undian iPhone",
            makeupTitle: "Fitur pengganti",
            makeupSubtitle: "Setiap putaran hanya dapat mengganti 1 hari terlewat",
            missedStep: "Progres terputus",
            missedDesc: "Dapat menggunakan pengganti",
            inviteStep: "Undang 1 pemain",
            inviteDesc: "Pastikan teman berhasil isi ulang",
            successStep: "Pengganti selesai",
            successDesc: "Progres tetap berlanjut",
            inviteNow: "Undang sekarang",
            pendingTitle: "Isi ulang hari ini sudah tercatat",
            pendingBody:
              "Undang 1 pemain dan pastikan mereka berhasil isi ulang hari ini. Jika tugas tidak selesai hari ini, isi ulang hari ini akan dihitung sebagai Hari 1 putaran baru.",
            modalTitle: "Pengingat pengganti",
            modalClose: "Nanti saja",
            inviteCount: "Undangan langsung",
            inviteHint: "Satu putaran hanya punya satu kesempatan pengganti.",
            rulesTitle: "Aturan aktivitas",
            rules: [
              "Isi ulang harian 10U akan menambah 1 hari progres.",
              "Isi ulang harus berurutan selama 7 hari untuk membuka kesempatan undian iPhone.",
              "Jika progres terputus, undang 1 pemain baru yang berhasil isi ulang untuk mengganti 1 hari.",
              "Setelah hadiah diklaim atau putaran gagal, sistem akan memulai siklus berikutnya.",
            ],
          }
        : {
            loading: "Loading activity...",
            heroAlt: "7-day recharge win iPhone",
            dailyTarget: "Daily target",
            dailyTargetHint: "Top up 10U daily to light up 1 progress day.",
            rewardHint: "Every successful recharge can unlock the matching daily reward.",
            finalPrize: "Day 7 enters the iPhone lucky draw.",
            progressTitle: "7-day topup progress",
            loopHint: "Looping cycle",
            progressHint: "After daily rewards are claimed or the streak is interrupted, progress starts a new cycle.",
            completed: "Lucky draw chance for the iPhone is unlocked.",
            day: "Day",
            unlocked: "Done",
            locked: "Locked",
            finalDay: "iPhone draw",
            makeupTitle: "Make-up feature",
            makeupSubtitle: "Each round can make up one missed day only",
            missedStep: "Progress interrupted",
            missedDesc: "Make-up can be used",
            inviteStep: "Invite 1 player",
            inviteDesc: "Friend completes a recharge",
            successStep: "Make-up complete",
            successDesc: "Progress continues",
            inviteNow: "Invite now",
            pendingTitle: "Today's recharge has been recorded",
            pendingBody:
              "Invite 1 player and make sure they complete a successful recharge today. If the task is not finished today, this recharge will be recorded as Day 1 of a new round.",
            modalTitle: "Make-up reminder",
            modalClose: "Later",
            inviteCount: "Direct invites",
            inviteHint: "Each round has only one make-up chance.",
            rulesTitle: "Activity rules",
            rules: [
              "Top up 10U daily to add one day of progress.",
              "Complete 7 consecutive topup days to unlock one iPhone lucky draw chance.",
              "If the streak is interrupted, invite one new player who completes a recharge to make up one missed day.",
              "After rewards are claimed or a round fails, the system starts the next cycle.",
            ],
          },
    [isIndonesian],
  )

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [status, setStatus] = useState<SevenDayTopupStatus | null>(null)
  const [dismissedPromptDate, setDismissedPromptDate] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    const fetchPageData = async () => {
      try {
        setLoading(true)
        setError(null)

        const nextStatus = await activityService.getSevenDayTopupStatus(currencyCode)
        if (!cancelled) {
          setStatus(nextStatus)
        }
      } catch (fetchError) {
        if (!cancelled) {
          setError(fetchError instanceof Error ? fetchError.message : "Failed to load activity data")
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void fetchPageData()

    return () => {
      cancelled = true
    }
  }, [currencyCode])

  const progressCount = Math.min(Math.max(status?.progress_count || 0, 0), 7)
  const isMakeupModalOpen =
    !!status?.pending_makeup?.recharge_date &&
    dismissedPromptDate !== status.pending_makeup.recharge_date

  const dayCards = Array.from({ length: 7 }, (_, index) => {
    const dayNumber = index + 1
    const done = index < progressCount

    return {
      dayNumber,
      done,
      final: dayNumber === 7,
      label: `${text.day} ${dayNumber}`,
    }
  })

  const featureCards = [
    {
      icon: `${ASSET_BASE}/2.png`,
      title: text.dailyTarget,
      desc: text.dailyTargetHint,
    },
    {
      icon: `${ASSET_BASE}/3.png`,
      title: isIndonesian ? "Hadiah harian" : "Daily rewards",
      desc: text.rewardHint,
    },
    {
      icon: `${ASSET_BASE}/4.png`,
      title: isIndonesian ? "Hadiah utama" : "Grand prize",
      desc: text.finalPrize,
    },
  ]

  const makeupSteps = [
    {
      icon: `${ASSET_BASE}/5.png`,
      title: text.missedStep,
      desc: text.missedDesc,
    },
    {
      icon: `${ASSET_BASE}/7.png`,
      title: text.inviteStep,
      desc: text.inviteDesc,
    },
    {
      icon: `${ASSET_BASE}/2.png`,
      title: text.successStep,
      desc: text.successDesc,
    },
  ]

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#050403] px-4">
        <div className="flex items-center gap-3 rounded-2xl border border-[#e5a83a]/30 bg-black/50 px-5 py-4 text-[#f8c961] shadow-[0_20px_80px_rgba(229,168,58,0.18)]">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>{text.loading}</span>
        </div>
      </div>
    )
  }

  return (
    <>
      <main className="min-h-screen overflow-hidden bg-[#050403] px-3 py-4 text-white sm:px-5 lg:px-8 lg:py-7">
        <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(circle_at_20%_8%,rgba(245,181,66,0.2),transparent_28%),radial-gradient(circle_at_82%_18%,rgba(255,236,188,0.12),transparent_26%),linear-gradient(180deg,#100a03_0%,#050403_48%,#080602_100%)]" />
        <div className="pointer-events-none fixed inset-0 opacity-[0.17] [background-image:linear-gradient(rgba(255,198,91,0.22)_1px,transparent_1px),linear-gradient(90deg,rgba(255,198,91,0.18)_1px,transparent_1px)] [background-size:42px_42px]" />

        <div className="relative mx-auto flex w-full max-w-[760px] flex-col gap-3">
          <section className="relative overflow-hidden rounded-[26px] border border-[#a66c18]/60 bg-[linear-gradient(145deg,rgba(21,17,11,0.98),rgba(5,4,3,0.97)_58%,rgba(42,25,4,0.96))] p-4 shadow-[0_22px_70px_rgba(0,0,0,0.58)] sm:p-6">
            <div className="absolute inset-0 opacity-80 [background-image:radial-gradient(circle_at_72%_23%,rgba(255,185,56,0.28),transparent_23%),radial-gradient(circle_at_46%_0%,rgba(255,226,155,0.16),transparent_19%)]" />

            <div className="relative z-10 flex flex-col gap-2 sm:gap-3">
              <div className="relative min-h-[150px] min-[420px]:min-h-[170px] sm:min-h-[235px]">
                <div className="relative z-10 max-w-[72%] sm:max-w-[58%]">
                  <img src={titleImage} alt={text.heroAlt} className="h-auto w-full max-w-[392px] drop-shadow-[0_8px_18px_rgba(0,0,0,0.75)]" />
                </div>
                <img
                  src={`${ASSET_BASE}/1.png`}
                  alt="iPhone lucky draw prize"
                  className="pointer-events-none absolute right-[-10px] top-[-6px] z-0 w-[44%] max-w-[300px] opacity-95 drop-shadow-[0_22px_36px_rgba(0,0,0,0.7)] min-[420px]:w-[48%] sm:right-1 sm:top-[-10px] sm:w-[40%]"
                />
                <div className="absolute right-[8%] top-8 h-24 w-24 rounded-full bg-[#f6b93e]/20 blur-3xl sm:top-12 sm:h-36 sm:w-36" />
              </div>

              <div className="grid gap-3 border-t border-[#815716]/60 pt-3 sm:grid-cols-3">
                {featureCards.map((feature) => (
                  <div
                    key={feature.title}
                    className="flex items-start gap-2.5 border-[#815716]/55 sm:border-r sm:pr-3 sm:last:border-r-0 sm:last:pr-0"
                  >
                    <img src={feature.icon} alt="" className="mt-0.5 h-9 w-9 shrink-0 object-contain" />
                    <div className="min-w-0">
                      <p className="text-[10px] font-bold uppercase leading-tight text-[#ffdc82]/85">{feature.title}</p>
                      <p className="mt-0.5 text-[10px] leading-snug text-white/72 sm:text-[11px]">{feature.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </section>

          <div className="flex flex-col gap-3">
            <section className="rounded-[22px] border border-[#8e611c]/65 bg-[linear-gradient(180deg,rgba(19,16,11,0.96),rgba(7,6,5,0.96))] p-4 shadow-[0_18px_58px_rgba(0,0,0,0.38)] sm:p-5">
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <h2 className="text-lg font-black text-[#ffce62] sm:text-xl">{text.progressTitle}</h2>
                  <span className="inline-flex items-center gap-1 rounded-full bg-[#f7b940]/10 px-2 py-1 text-[11px] font-bold text-[#f5bf45]">
                    <RefreshCcw className="h-3.5 w-3.5" />
                    {text.loopHint}
                  </span>
                </div>
                <p className="mt-1 text-xs text-white/55">{text.progressHint}</p>
              </div>

              {status?.completed_round && (
                <div className="mt-3 flex items-center gap-2 rounded-2xl border border-emerald-400/25 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-200">
                  <CheckCircle2 className="h-4 w-4 shrink-0" />
                  <span>{text.completed}</span>
                </div>
              )}

              {status?.pending_makeup && (
                <div className="mt-3 rounded-2xl border border-amber-400/25 bg-amber-500/10 px-3 py-3 text-xs text-amber-100">
                  <div className="flex items-start gap-2">
                    <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0 text-amber-300" />
                    <div>
                      <p className="font-bold text-amber-200">{text.pendingTitle}</p>
                      <p className="mt-1 leading-relaxed text-white/76">{text.pendingBody}</p>
                    </div>
                  </div>
                </div>
              )}

              <div className="mt-4 grid grid-cols-4 gap-2 sm:grid-cols-7 lg:gap-2.5">
                {dayCards.map((card) => (
                  <div
                    key={card.label}
                    className={`relative min-h-[122px] overflow-hidden rounded-[14px] border px-2 py-3 text-center shadow-[inset_0_1px_0_rgba(255,255,255,0.07)] transition-transform hover:-translate-y-0.5 ${
                      card.done
                        ? "border-[#e3a33a]/70 bg-[linear-gradient(180deg,#efb74a,#9a5c08)] text-white"
                        : "border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.025))] text-white/[0.46]"
                    } ${card.final ? "col-span-2 sm:col-span-1" : ""}`}
                  >
                    {card.done && <div className="absolute inset-x-2 bottom-0 h-12 bg-[#ffd875]/18 blur-xl" />}
                    <p className="relative text-xs font-black">{card.label}</p>
                    <div className="relative mx-auto mt-4 grid h-12 w-12 place-items-center rounded-full border border-current/25 bg-black/18">
                      {card.done ? (
                        <CheckCircle2 className="h-8 w-8 text-[#fff2b8]" />
                      ) : card.final ? (
                        <img src={`${ASSET_BASE}/6.png`} alt="iPhone prize" className="h-16 w-12 object-contain opacity-95" />
                      ) : (
                        <LockKeyhole className="h-7 w-7 text-white/[0.32]" />
                      )}
                    </div>
                    <p className="relative mt-3 text-[11px] font-bold leading-tight">
                      {card.final ? text.finalDay : card.done ? text.unlocked : text.locked}
                    </p>
                  </div>
                ))}
              </div>
            </section>

            <section className="rounded-[22px] border border-[#8e611c]/65 bg-[linear-gradient(180deg,rgba(18,15,10,0.96),rgba(8,7,5,0.96))] p-4 shadow-[0_18px_58px_rgba(0,0,0,0.34)] sm:p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h2 className="text-lg font-black text-[#ffce62]">{text.makeupTitle}</h2>
                  <p className="mt-1 text-xs text-[#ffd27a]/80">{text.makeupSubtitle}</p>
                </div>
                <button
                  type="button"
                  onClick={() => router.push("/invite")}
                  className="rounded-full bg-[linear-gradient(180deg,#ffe08a,#e49b27_55%,#a86108)] px-5 py-2.5 text-sm font-black text-[#1d1002] shadow-[0_10px_24px_rgba(226,158,44,0.32)] transition-transform hover:scale-[1.02] active:scale-[0.98]"
                >
                  {text.inviteNow}
                </button>
              </div>

              <div className="mt-4 grid grid-cols-3 gap-2">
                {makeupSteps.map((step, index) => (
                  <div key={step.title} className="relative rounded-[14px] border border-[#815716]/55 bg-black/30 p-2.5">
                    {index > 0 && (
                      <div className="absolute -left-3 top-1/2 hidden h-px w-3 bg-[#c5882a] sm:block" />
                    )}
                    <div className="grid h-11 place-items-center">
                      <img src={step.icon} alt="" className="max-h-11 max-w-12 object-contain" />
                    </div>
                    <p className="mt-2 text-center text-[11px] font-black leading-tight text-white sm:text-sm">{step.title}</p>
                    <p className="mt-1 text-center text-[10px] leading-snug text-white/55 sm:text-xs">{step.desc}</p>
                  </div>
                ))}
              </div>
            </section>

            <section className="relative overflow-hidden rounded-[22px] border border-[#8e611c]/65 bg-[linear-gradient(135deg,rgba(17,14,10,0.98),rgba(5,4,3,0.96))] p-4 shadow-[0_18px_58px_rgba(0,0,0,0.34)] sm:p-5">
              <div className="relative">
                <h2 className="inline-flex items-center gap-2 rounded-lg border border-[#9c6b21]/60 bg-[#1a1005]/80 px-3 py-1 text-base font-black text-[#ffce62]">
                  <Crown className="h-4 w-4" />
                  {text.rulesTitle}
                </h2>
                <ol className="mt-3 space-y-2 text-xs leading-relaxed text-white/66 sm:text-sm">
                  {text.rules.map((rule, index) => (
                    <li key={rule}>
                      <span className="mr-2 font-black text-[#f8c961]">{index + 1}.</span>
                      {rule}
                    </li>
                  ))}
                </ol>
              </div>
            </section>

            {error && (
              <div className="rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
                {error}
              </div>
            )}
          </div>
        </div>
      </main>

      {isMakeupModalOpen && status?.pending_makeup && (
        <div className="fixed inset-0 z-[100001] flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm">
          <div className="relative w-full max-w-md rounded-3xl border border-[#d99a32]/30 bg-[#090704] p-6 shadow-2xl">
            <button
              type="button"
              onClick={() => setDismissedPromptDate(status.pending_makeup!.recharge_date)}
              className="absolute right-4 top-4 rounded-full bg-white/5 p-2 text-white/60 transition-colors hover:bg-white/10 hover:text-white"
            >
              <X className="h-4 w-4" />
            </button>

            <div className="flex items-center gap-3">
              <div className="rounded-2xl bg-[#f6b93e]/15 p-3 text-[#f8c961]">
                <Gift className="h-5 w-5" />
              </div>
              <div>
                <p className="text-lg font-bold text-white">{text.modalTitle}</p>
                <p className="text-sm text-white/55">{status.pending_makeup.recharge_date}</p>
              </div>
            </div>

            <div className="mt-5 rounded-2xl border border-amber-400/20 bg-amber-500/10 px-4 py-3 text-sm text-white/80">
              <p className="font-semibold text-amber-200">{text.pendingTitle}</p>
              <p className="mt-2 leading-relaxed">{text.pendingBody}</p>
            </div>

            <div className="mt-4 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-white/75">
              <p>
                {text.inviteCount}: <span className="font-semibold text-white">{status.direct_invite_count}</span>
              </p>
              <p className="mt-2">{text.inviteHint}</p>
            </div>

            <div className="mt-6 flex flex-col gap-3">
              <button
                type="button"
                onClick={() => router.push("/invite")}
                className="rounded-2xl bg-[linear-gradient(180deg,#ffe08a,#e49b27_55%,#a86108)] px-4 py-3 text-sm font-black text-[#1d1002] transition-transform hover:scale-[1.01]"
              >
                {text.inviteNow}
              </button>
              <button
                type="button"
                onClick={() => setDismissedPromptDate(status.pending_makeup!.recharge_date)}
                className="text-sm text-white/45 transition-colors hover:text-white/70"
              >
                {text.modalClose}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default SevenDayTopupPage
