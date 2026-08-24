"use client"

import React from "react"
import {
  Clock3,
  Coins,
  Gift,
  Info,
  Medal,
  TrendingUp,
  Trophy,
  Users,
} from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useBettingRankValue } from "@/hooks/useBettingRankValue"

type RuleItem = {
  place: number
  percent: string
  detail: string
}

const RULES_EN: RuleItem[] = [
  { place: 1, percent: "50%", detail: "Bet - 50% of the weekly contest prize pool" },
  { place: 2, percent: "25%", detail: "Bet - 25% of the weekly contest prize pool" },
  { place: 3, percent: "12%", detail: "Bet - 12% of the weekly contest prize pool" },
  { place: 4, percent: "6%", detail: "Bet - 6% of the weekly contest prize pool" },
  { place: 5, percent: "3%", detail: "Bet - 3% of the weekly contest prize pool" },
  { place: 6, percent: "1.5%", detail: "Bet - 1.5% of the weekly contest prize pool" },
  { place: 7, percent: "0.9%", detail: "Bet - 0.9% of the weekly contest prize pool" },
  { place: 8, percent: "0.7%", detail: "Bet - 0.7% of the weekly contest prize pool" },
  { place: 9, percent: "0.5%", detail: "Bet - 0.5% of the weekly contest prize pool" },
  { place: 10, percent: "0.4%", detail: "Bet - 0.4% of the weekly contest prize pool" },
]

const RULES_ID: RuleItem[] = [
  { place: 1, percent: "50%", detail: "Taruhan - 50% dari total hadiah mingguan" },
  { place: 2, percent: "25%", detail: "Taruhan - 25% dari total hadiah mingguan" },
  { place: 3, percent: "12%", detail: "Taruhan - 12% dari total hadiah mingguan" },
  { place: 4, percent: "6%", detail: "Taruhan - 6% dari total hadiah mingguan" },
  { place: 5, percent: "3%", detail: "Taruhan - 3% dari total hadiah mingguan" },
  { place: 6, percent: "1.5%", detail: "Taruhan - 1.5% dari total hadiah mingguan" },
  { place: 7, percent: "0.9%", detail: "Taruhan - 0.9% dari total hadiah mingguan" },
  { place: 8, percent: "0.7%", detail: "Taruhan - 0.7% dari total hadiah mingguan" },
  { place: 9, percent: "0.5%", detail: "Taruhan - 0.5% dari total hadiah mingguan" },
  { place: 10, percent: "0.4%", detail: "Taruhan - 0.4% dari total hadiah mingguan" },
]

const medalStyles: Record<number, string> = {
  1: "border-yellow-300/80 bg-gradient-to-br from-yellow-200 via-yellow-400 to-amber-600 text-[#2d1a00]",
  2: "border-slate-200/70 bg-gradient-to-br from-slate-100 via-slate-300 to-slate-500 text-[#111827]",
  3: "border-orange-200/70 bg-gradient-to-br from-orange-200 via-orange-400 to-amber-700 text-[#2d1400]",
}

const BettingRankPage: React.FC = () => {
  const { language } = useLanguage()
  const { value, loading, error } = useBettingRankValue()
  const isIndonesian = language === "id"

  const copy = isIndonesian
    ? {
      title: "Leaderboard Grand Prizes",
      subtitle: "Papan peringkat mingguan - bagi total hadiah",
      stat1Title: "Peringkat Mingguan",
      stat1Desc: "Setiap taruhan valid akan menambah total hadiah mingguan.",
      stat2Title: "Top 10 Pemain",
      stat2Desc: "10 pemain dengan total taruhan tertinggi berbagi prize pool.",
      stat3Title: "Waktu Undian",
      stat3Desc: "Pengumuman dilakukan setiap minggu sesuai jadwal reset UTC.",
      poolTitle: "Total Hadiah Saat Ini",
      poolHint: "Prize pool akan bertambah secara real time saat pemain memasang taruhan.",
      top10: "Top 10",
      top10Title: "Top 10 Bonus Window",
      top10Desc: "Finish the week inside the top 10 and you will share the live prize pool based on your final ranking.",
      rulesTitle: "Aturan Pembagian Hadiah",
      rulesSubtitle: "10 pemain dengan total taruhan mingguan tertinggi akan berbagi hadiah.",
      notesTitle: "Catatan Event",
      notes: [
        "Peringkat dihitung berdasarkan total taruhan valid selama periode event.",
        "Prize pool berasal dari akumulasi seluruh taruhan valid pemain."],
      loading: "Memuat...",
      fallbackValue: "--",
    }
    : {
      title: "Leaderboard Grand Prizes",
      subtitle: "Weekly leaderboard - share the prize pool",
      stat1Title: "Weekly Leaderboard",
      stat1Desc: "Every valid bet you place adds to the weekly prize pool.",
      stat2Title: "Top 10 Players",
      stat2Desc: "The top 10 players by total weekly bets share the pool.",
      stat3Title: "Draw Time",
      stat3Desc: "Results are settled weekly based on the UTC reset schedule.",
      poolTitle: "Current Prize Pool",
      poolHint: "The prize pool increases in real time as players place bets.",
      top10: "Top 10",
      top10Title: "Top 10 Bonus Window",
      top10Desc: "Finish the week inside the top 10 and you will share the live prize pool based on your final ranking.",
      rulesTitle: "Prize Distribution Rules",
      rulesSubtitle: "The top 10 players by total weekly bets will share the prize pool.",
      notesTitle: "Event Notes",
      notes: [
        "Leaderboard rankings are based on the total amount of valid bets during the event period.",
        "The prize pool includes all valid bets from all players."],
      loading: "Loading...",
      fallbackValue: "--",
    }

  const formatter = new Intl.NumberFormat(isIndonesian ? "id-ID" : "en-US")
  const displayValue = loading
    ? copy.loading
    : value
      ? formatter.format(value.current_value)
      : copy.fallbackValue
  const drawTime = isIndonesian
    ? "Setiap Sabtu pukul 12 siang (UTC+8)"
    : "Every Saturday 12:00 PM (UTC+8)"
  const rules = isIndonesian ? RULES_ID : RULES_EN

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(255,191,0,0.12),_transparent_28%),linear-gradient(180deg,_#161005_0%,_#090909_38%,_#0e0a05_100%)] px-3 py-4 sm:px-6 sm:py-6">
      <div className="mx-auto w-full max-w-md sm:max-w-xl lg:max-w-5xl [@media(orientation:landscape)]:max-w-7xl">
        <div className="overflow-hidden rounded-[28px] border border-[#7d5607] bg-[linear-gradient(180deg,_rgba(19,16,12,0.98)_0%,_rgba(9,9,9,0.99)_100%)] p-3 shadow-[0_24px_80px_rgba(0,0,0,0.55)] sm:p-6 lg:p-8">
          <div className="mb-3 sm:mb-4">
            <div className="rounded-[20px] border border-[#6e4b08] bg-[radial-gradient(circle_at_top_left,_rgba(255,204,92,0.16),_transparent_34%),linear-gradient(180deg,_rgba(21,18,13,0.98)_0%,_rgba(8,8,8,1)_100%)] p-3 sm:p-4 [@media(orientation:landscape)]:p-5">
              <div className="grid gap-3 [@media(orientation:landscape)]:grid-cols-[0.76fr_1.24fr] [@media(orientation:landscape)]:items-end">
                <div className="min-w-0">
                  <div className="inline-flex max-w-full items-center rounded-full border border-[#8b620c] bg-[#18120a] px-2.5 py-0.5 text-[9px] font-black uppercase tracking-[0.12em] text-[#ffd56c] sm:text-[10px] sm:tracking-[0.14em]">
                    <span className="truncate">{copy.subtitle}</span>
                  </div>
                  <h1 className="mt-2 max-w-[11ch] text-[1.55rem] font-black uppercase leading-[0.9] tracking-tight text-[#ffe8af] drop-shadow-[0_4px_10px_rgba(255,193,71,0.18)] sm:text-[2.25rem] lg:text-[2.85rem] [@media(orientation:landscape)]:max-w-none [@media(orientation:landscape)]:text-[clamp(2rem,3.9vw,3.25rem)]">
                    {copy.title}
                  </h1>
                </div>

                <div className="grid grid-cols-1 gap-2 sm:grid-cols-3 [@media(orientation:landscape)]:gap-2.5">
                  <FeatureCard
                    icon={<Medal className="h-4 w-4" />}
                    title={copy.stat1Title}
                    description={copy.stat1Desc}
                  />
                  <FeatureCard
                    icon={<Users className="h-4 w-4" />}
                    title={copy.stat2Title}
                    description={copy.stat2Desc}
                  />
                  <FeatureCard
                    icon={<Clock3 className="h-4 w-4" />}
                    title={copy.stat3Title}
                    description={drawTime}
                  />
                </div>
              </div>

            </div>
          </div>

          <section className="mb-4 overflow-hidden rounded-[22px] border border-[#6e4b08] bg-[radial-gradient(circle_at_left,_rgba(255,191,0,0.12),_transparent_22%),linear-gradient(180deg,_rgba(18,14,9,0.98)_0%,_rgba(8,8,8,1)_100%)] p-3 sm:mb-5 sm:p-6">
            <div className="grid grid-cols-[52px_minmax(0,1fr)_52px] items-center gap-3 sm:grid-cols-[80px_minmax(0,1fr)_120px] sm:gap-5">
              <div className="flex h-14 w-14 items-center justify-center rounded-xl border border-[#8d6515] bg-[#171109] text-[#ffcf68] shadow-[0_0_30px_rgba(255,188,0,0.08)] sm:h-20 sm:w-20 sm:rounded-2xl">
                <Coins className="h-6 w-6 sm:h-9 sm:w-9" strokeWidth={1.8} />
              </div>
              <div className="min-w-0">
                <div className="text-[11px] font-bold uppercase tracking-[0.16em] text-[#e2bc62] sm:text-sm sm:tracking-[0.18em]">
                  {copy.poolTitle}
                </div>
                <div className="mt-1.5 flex items-end gap-1 sm:mt-2 sm:gap-2">
                  <span className="text-lg font-black text-[#ffe9b2] sm:text-5xl">U</span>
                  <span className="truncate text-[1.65rem] font-black leading-none text-[#ffd56c] sm:text-5xl">
                    {displayValue}
                  </span>
                </div>
                <p className="mt-1.5 text-[11px] leading-4 text-[#dbcda9]/88 sm:mt-3 sm:text-sm sm:leading-5">
                  {copy.poolHint}
                </p>
              </div>

              <div className="flex items-end justify-end gap-1 self-end sm:gap-1.5">
                {[16, 26, 38, 54].map((height) => (
                  <div
                    key={height}
                    className="w-2.5 rounded-t-sm border border-[#b77c10] bg-gradient-to-t from-[#8f5f05] via-[#ffca4b] to-[#fff1a9] sm:w-4"
                    style={{ height }}
                  />
                ))}
                <TrendingUp className="mb-2 h-4 w-4 text-[#ffd56c] sm:mb-4 sm:h-6 sm:w-6" />
              </div>
            </div>
          </section>

          <section className="mb-4 rounded-[22px] border border-[#6e4b08] bg-[linear-gradient(180deg,_rgba(18,14,9,0.98)_0%,_rgba(8,8,8,1)_100%)] p-3 sm:mb-5 sm:p-5">
            <div className="flex flex-col items-center text-center">
              <div className="rounded-full bg-[#2b1d08] p-1.5 text-[#ffcf68] sm:p-2">
                <Trophy className="h-4 w-4 sm:h-5 sm:w-5" />
              </div>
              <div className="mt-2">
                <h2 className="text-lg font-black uppercase leading-none tracking-tight text-[#ffe9b2] sm:text-3xl">
                  {copy.rulesTitle}
                </h2>
                <p className="mt-1 text-xs leading-4 text-[#dbcda9]/86 sm:text-base">
                  {copy.rulesSubtitle}
                </p>
              </div>
            </div>

            <div className="mt-3 overflow-hidden rounded-[16px] border border-[#5f450f] sm:mt-4 sm:rounded-[20px]">
              {rules.map((rule, index) => (
                <div
                  key={rule.place}
                  className={`grid grid-cols-[32px_minmax(0,1fr)_56px] items-center gap-x-2 border-t border-[#463511] px-2.5 py-2 sm:grid-cols-[32px_minmax(0,1fr)_56px] sm:px-3 sm:py-2 ${index === 0 ? "border-t-0" : ""
                    }`}
                >
                  <div className="flex items-center justify-center">
                    {rule.place <= 3 ? (
                      <span
                        className={`inline-flex h-7 w-7 items-center justify-center rounded-full border text-[11px] font-black shadow-[0_0_18px_rgba(255,191,0,0.14)] sm:h-11 sm:w-11 sm:text-base ${medalStyles[rule.place]
                          }`}
                      >
                        {rule.place}
                      </span>
                    ) : (
                      <span className="inline-flex h-7 w-7 items-center justify-center rounded-full border border-[#71500e] bg-[#191209] text-[11px] font-black text-[#ffd56c] sm:h-10 sm:w-10 sm:text-base">
                        {rule.place}
                      </span>
                    )}
                  </div>

                  <div className="min-w-0 text-center text-[10px] leading-4 text-[#e8d8af]/88 sm:text-sm">
                    {rule.detail}
                  </div>

                  <div className="text-center text-lg font-black leading-none text-[#ffd56c] sm:text-2xl">
                    {rule.percent}
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="rounded-[22px] border border-[#6e4b08] bg-[radial-gradient(circle_at_right,_rgba(255,191,0,0.1),_transparent_22%),linear-gradient(180deg,_rgba(18,14,9,0.98)_0%,_rgba(8,8,8,1)_100%)] p-4 sm:p-6">
            <div className="grid gap-5 [@media(orientation:landscape)]:grid-cols-[1.2fr_0.8fr] [@media(orientation:landscape)]:items-center">
              <div>
                <div className="flex items-center gap-3">
                  <div className="rounded-full bg-[#2b1d08] p-2 text-[#ffcf68]">
                    <Info className="h-5 w-5" />
                  </div>
                  <h2 className="text-2xl font-black uppercase tracking-tight text-[#ffe9b2]">
                    {copy.notesTitle}
                  </h2>
                </div>

                <ol className="mt-4 grid gap-2 text-sm leading-6 text-[#e8d8af]/88 sm:text-base">
                  {copy.notes.map((note, index) => (
                    <li key={note} className="flex gap-3">
                      <span className="w-5 shrink-0 font-black text-[#ffd56c]">{index + 1}.</span>
                      <span>{note}</span>
                    </li>
                  ))}
                </ol>
              </div>

              <div className="hidden justify-center [@media(orientation:landscape)]:flex [@media(orientation:landscape)]:justify-end">
                <div className="relative flex h-48 w-48 items-center justify-center">
                  <div className="absolute bottom-5 left-4 h-6 w-6 rounded-full border border-[#9a6b0b] bg-gradient-to-br from-[#ffde87] to-[#a76a00]" />
                  <div className="absolute bottom-0 right-3 h-7 w-7 rounded-full border border-[#9a6b0b] bg-gradient-to-br from-[#ffde87] to-[#a76a00]" />
                  <div className="absolute bottom-4 right-10 h-4 w-4 rounded-full border border-[#9a6b0b] bg-gradient-to-br from-[#ffde87] to-[#a76a00]" />
                  <div className="absolute inset-x-6 bottom-2 h-6 rounded-full bg-yellow-500/10 blur-xl" />
                  <div className="relative flex h-36 w-32 items-center justify-center rounded-[28px] border border-[#8a6310] bg-[linear-gradient(180deg,_#2b1a07_0%,_#120d06_100%)] shadow-[0_18px_40px_rgba(0,0,0,0.45)]">
                    <Gift className="h-16 w-16 text-[#ffd56c]" />
                    <div className="absolute inset-x-0 top-1/2 h-3 -translate-y-1/2 bg-gradient-to-r from-[#8a6310] via-[#ffd56c] to-[#8a6310]" />
                    <div className="absolute inset-y-0 left-1/2 w-3 -translate-x-1/2 bg-gradient-to-b from-[#8a6310] via-[#ffd56c] to-[#8a6310]" />
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>

        {error && (
          <div className="mt-4 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">
            {error}
          </div>
        )}
      </div>
    </div>
  )
}

const FeatureCard: React.FC<{
  icon: React.ReactNode
  title: string
  description: string
  className?: string
}> = ({ icon, title, description, className }) => {
  return (
    <div className={`${className ?? ""} flex items-start gap-2.5 rounded-xl border border-[#5e430d] bg-[#100d09]/90 p-2.5 shadow-[inset_0_1px_0_rgba(255,215,112,0.05)] sm:gap-3 sm:p-3 [@media(orientation:landscape)]:min-h-[84px]`}>
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#231708] text-[#ffcf68]">
        {icon}
      </div>
      <div className="min-w-0">
        <div className="text-[10px] font-black uppercase leading-4 tracking-[0.07em] text-[#f6d987] sm:text-xs sm:tracking-[0.08em]">
          {title}
        </div>
        <p className="mt-0.5 text-[11px] leading-4 text-[#d8caa6]/85 sm:text-xs sm:leading-5">
          {description}
        </p>
      </div>
    </div>
  )
}

export default BettingRankPage
