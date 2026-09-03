"use client"

import React, { useMemo } from "react"
import { useRouter } from "next/navigation"
import {
  CheckCircle2,
  ChevronRight,
  Gift,
  Loader2,
  LockKeyhole,
  ShieldCheck,
} from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useNewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"

const DEPOSIT_STAGE_NAMES = ["First Deposit", "Second Deposit", "Third Deposit", "Fourth Deposit"]

const NewUserRechargePage: React.FC = () => {
  const router = useRouter()
  const { language } = useLanguage()
  const { status, loading } = useNewUserRechargeStatus()

  const isIndonesian = language === "id"
  // Chinese strings in this legacy activity were mojibake; fall back to English until translated.
  const isChinese = false
  const isRussian = language === "ru"

  const tiers = status?.tiers || [
    {
      day: 1,
      rate: 1,
      status: "current",
      deposit_amount: 0,
      reward_amount: 0,
      min_deposit: 5,
      wager_required: 0,
      wager_completed: 0,
      wager_unlocked: false,
    },
    {
      day: 2,
      rate: 1.75,
      status: "locked",
      deposit_amount: 0,
      reward_amount: 0,
      min_deposit: 5,
      wager_required: 0,
      wager_completed: 0,
      wager_unlocked: false,
    },
    {
      day: 3,
      rate: 2,
      status: "locked",
      deposit_amount: 0,
      reward_amount: 0,
      min_deposit: 5,
      wager_required: 0,
      wager_completed: 0,
      wager_unlocked: false,
    },
    {
      day: 4,
      rate: 2.5,
      status: "locked",
      deposit_amount: 0,
      reward_amount: 0,
      min_deposit: 5,
      wager_required: 0,
      wager_completed: 0,
      wager_unlocked: false,
    },
  ]


  const text = useMemo(
    () =>
      isChinese
        ? {
          titleTop: "充值",
          titleBottom: "赠送",
          summary: "您充值得越多赠送得越多",
          activityPeriod: "活动周期：新用户专属",
          rewardTitle: "前 4 次充值额外奖励",
          bonusLabel: "充值赠送",
          restrictionsTitle: "活动限制",
          restrictionFund: "额外获得的充值赠送金额不可直接提取。",
          restrictionMinDeposit: "单笔充值大于等于5U才有赠送。",
          restrictionUnlock: "每一档次需在完成上一档充值后解锁。",
          completed: "前 4 档奖励已全部解锁。",
          depositNow: "立即充值领取更多奖励",
          stageLabel: "第",
          stageNames: ["一档", "二档", "三档", "四档"],
          current: "进行中",
          unlocked: "已完成",
          locked: "未解锁",
        }
        : isRussian
          ? {
            titleTop: "Пополнение",
            titleBottom: "Бонус",
            summary: "Чем больше вы пополняете счёт, тем больше бонусов получаете.",
            activityPeriod: "Эксклюзив для новых пользователей",
            rewardTitle: "Дополнительные бонусы за первые 4 депозита",
            bonusLabel: "Бонус за депозит",
            restrictionsTitle: "Правила акции",
            restrictionFund: "Дополнительный бонус за депозит нельзя вывести напрямую.",
            restrictionMinDeposit: "Бонус доступен только для одного депозита от 5U.",
            restrictionUnlock: "Каждый этап открывается после завершения предыдущего депозита.",
            completed: "Все 4 бонусных этапа разблокированы.",
            depositNow: "Пополнить снова и получить больше наград",
            stageLabel: "Депозит",
            stageNames: ["Первый", "Второй", "Третий", "Четвёртый"],
            current: "В процессе",
            unlocked: "Завершено",
            locked: "Заблокировано",
          }
        : isIndonesian
          ? {
            titleTop: "Isi Ulang",
            titleBottom: "Bonus",
            summary: "Semakin banyak kamu isi ulang, semakin banyak bonus yang kamu dapatkan.",
            activityPeriod: "Eksklusif Pengguna Baru",
            rewardTitle: "Bonus ekstra untuk 4 deposit pertama",
            bonusLabel: "Bonus deposit",
            restrictionsTitle: "Syarat aktivitas",
            restrictionFund: "Bonus deposit yang didapat tidak bisa langsung ditarik.",
            restrictionMinDeposit: "Bonus hanya berlaku untuk sekali deposit minimal 5U.",
            restrictionUnlock: "Setiap tahap terbuka setelah tahap deposit sebelumnya selesai.",
            completed: "Semua 4 tahap bonus sudah berhasil dibuka.",
            depositNow: "Deposit lagi & dapatkan hadiah lebih banyak",
            stageLabel: "Deposit ke-",
            stageNames: ["Pertama", "Kedua", "Ketiga", "Keempat"],
            current: "Sedang berlangsung",
            unlocked: "Selesai",
            locked: "Terkunci",
          }
          : {
            titleTop: "Deposit",
            titleBottom: "Bonus",
            summary: "The more you deposit, the more bonus you receive.",
            activityPeriod: "New User Exclusive",
            rewardTitle: "Extra rewards for the first 4 deposits",
            bonusLabel: "Deposit bonus",
            restrictionsTitle: "Activity rules",
            restrictionFund: "The extra deposit bonus amount cannot be withdrawn directly.",
            restrictionMinDeposit: "Bonuses are only available for single deposits of 5U or more.",
            restrictionUnlock: "Each stage unlocks after the previous deposit stage is completed.",
            completed: "All 4 bonus stages have been unlocked.",
            depositNow: "Deposit again & get more rewards",
            stageLabel: "Stage",
            stageNames: ["First", "Second", "Third", "Final"],
            current: "In progress",
            unlocked: "Completed",
            locked: "Locked",
          },
    [isChinese, isIndonesian, isRussian],
  )

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,_rgba(255,193,59,0.12),_#060606_34%,_#020202_72%)] px-4">
        <div className="flex items-center gap-3 rounded-2xl border border-lucky-gold/20 bg-black/40 px-5 py-4 text-lucky-gold">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>{isRussian ? "Загрузка акции..." : isIndonesian ? "Memuat aktivitas..." : "Loading activity..."}</span>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(255,193,59,0.12),_#060606_34%,_#020202_72%)] px-3 py-4 sm:px-6 sm:py-6">
      <div className="mx-auto w-full max-w-md sm:max-w-2xl lg:max-w-none 2xl:max-w-[1600px]">
        <div className="overflow-hidden rounded-[30px] border border-[#6b4a0b] bg-[#070707] p-3 shadow-[0_26px_90px_rgba(0,0,0,0.58)] sm:p-5">
          <section className="relative overflow-hidden rounded-[22px] border border-[#7c560e] bg-[linear-gradient(135deg,#161004_0%,#090909_34%,#1b1204_100%)] px-4 py-3.5 sm:px-5 sm:py-4">
            <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(255,196,72,0.18),transparent_36%),linear-gradient(90deg,rgba(255,196,72,0.06)_0%,rgba(255,196,72,0.02)_24%,rgba(0,0,0,0)_50%,rgba(255,196,72,0.02)_76%,rgba(255,196,72,0.06)_100%)]" />

            <div className="relative mx-auto max-w-3xl text-center">
              <div className="flex items-center justify-center gap-2 text-[clamp(2rem,8vw,4.8rem)] font-black uppercase leading-[0.9] tracking-[-0.06em] sm:gap-4">
                <span className="text-transparent [text-shadow:0_0_18px_rgba(255,208,99,0.2)] bg-[linear-gradient(180deg,#fff9e7_0%,#ffe69b_18%,#ffcb4b_44%,#f29a00_74%,#7d4300_100%)] bg-clip-text">
                  {text.titleTop}
                </span>
                <span className="text-transparent [text-shadow:0_0_20px_rgba(255,189,58,0.22)] bg-[linear-gradient(180deg,#fff9ea_0%,#ffd86f_18%,#ffb300_46%,#d77a00_72%,#6c3400_100%)] bg-clip-text">
                  {text.titleBottom}
                </span>
              </div>

              <div className="mt-2 inline-flex max-w-full items-center justify-center rounded-full border border-[#b8831f]/40 bg-[linear-gradient(180deg,rgba(255,212,104,0.12),rgba(30,18,3,0.5))] px-4 py-1 shadow-[inset_0_1px_0_rgba(255,235,170,0.12)] sm:px-5">
                <p className="text-[11px] font-medium tracking-[0.18em] text-[#f0d18a]/90 sm:text-sm">
                  {text.rewardTitle}
                </p>
              </div>
            </div>
          </section>

          <section className="mt-4 overflow-hidden rounded-[24px] border border-[#5f4310] bg-[linear-gradient(180deg,#0d0d0d_0%,#090909_100%)]">
            <div className="border-b border-[#3a2a09] bg-[linear-gradient(180deg,rgba(255,196,72,0.12),rgba(255,196,72,0.03))] px-4 py-3 text-center sm:px-6">
              <h2 className="text-lg font-black tracking-wide text-[#f6c85e] sm:text-2xl">{text.activityPeriod}</h2>
            </div>

            <div className="grid gap-2.5 p-2.5 [@media(orientation:landscape)]:grid-cols-2 sm:gap-3 sm:p-4">
              {tiers.map((tier, index) => {
                const bonusPercent = Math.round(tier.rate * 100)
                const isLocked = tier.status === "locked"
                const isUnlocked = tier.status === "unlocked"
                const statusText = isUnlocked ? text.unlocked : isLocked ? text.locked : text.current

                return (
                  <div
                    key={tier.day}
                    className={`rounded-[20px] border px-3 py-2.5 transition-colors sm:px-4 sm:py-3 ${isLocked
                      ? "border-[#403219] bg-[linear-gradient(180deg,#101010_0%,#090909_100%)]"
                      : "border-[#8b6420] bg-[radial-gradient(circle_at_18%_10%,rgba(255,208,99,0.10),transparent_28%),radial-gradient(circle_at_80%_56%,rgba(255,179,0,0.12),transparent_24%),linear-gradient(135deg,#181104_0%,#0d0a05_46%,#181104_100%)]"
                      }`}
                  >
                    <div className="mb-1.5 flex justify-end">
                      <div
                        className={`inline-flex items-center gap-1.5 rounded-full px-3 py-0.5 text-[11px] font-bold sm:text-xs ${isUnlocked
                          ? "bg-emerald-500/18 text-emerald-300"
                          : isLocked
                            ? "bg-white/6 text-white/60"
                            : "bg-[#f2b93a]/18 text-[#ffd36b]"
                          }`}
                      >
                        {isUnlocked ? <CheckCircle2 className="h-4 w-4" /> : <LockKeyhole className="h-4 w-4" />}
                        {statusText}
                      </div>
                    </div>

                    <div className="relative overflow-hidden px-1 py-0 sm:px-1.5 sm:py-0.5">
                      <div
                        className={`pointer-events-none absolute inset-y-0 left-0 w-14 sm:w-16 ${
                          isLocked
                            ? "bg-[radial-gradient(circle,_rgba(255,255,255,0.16)_1.5px,transparent_1.6px)] [background-size:10px_10px] opacity-20"
                            : "bg-[radial-gradient(circle,_rgba(255,191,36,0.26)_1.5px,transparent_1.6px)] [background-size:10px_10px] opacity-35"
                        }`}
                      />
                      <div className="relative flex items-start justify-between gap-3">
                        <div className="min-w-0 pt-0">
                          <p
                            className={`text-[10px] font-bold uppercase tracking-[0.2em] sm:text-[12px] ${
                              isLocked ? "text-[#b9b9b9]" : "text-[#ffd36b]"
                            }`}
                          >
                            {isRussian ? `ЭТАП ${tier.day}` : `STAGE ${tier.day}`}
                          </p>
                          <div
                            className={`mt-0 text-[clamp(2.05rem,8.5vw,4.4rem)] font-black leading-[0.8] tracking-[-0.06em] ${
                              isLocked ? "text-[#bdbdbd]" : "text-[#ffd04f]"
                            }`}
                          >
                            {bonusPercent}%
                          </div>
                          <div className="text-[clamp(1.2rem,4.9vw,2.5rem)] font-black uppercase leading-[0.8] tracking-[-0.03em] text-white">
                            {isRussian ? "БОНУС" : "BONUS"}
                          </div>
                          <p
                            className={`mt-0 text-[0.95rem] font-medium leading-tight sm:text-base ${
                              isLocked ? "text-[#d0d0d0]" : "text-white"
                            }`}
                          >
                            {isRussian ? (text.stageNames[index] || `Депозит ${tier.day}`) : (DEPOSIT_STAGE_NAMES[index] || `Deposit ${tier.day}`)}
                          </p>
                        </div>

                        <div
                          className={`ml-2 mt-2 flex h-12 w-12 shrink-0 items-center justify-center self-center rounded-[1.35rem] border sm:mt-3 sm:h-16 sm:w-16 ${
                            isLocked ? "border-white/10 bg-white/[0.03]" : "border-[#a77716] bg-[#2a1b05]/70"
                          }`}
                        >
                          <Gift className={`h-6 w-6 sm:h-8 sm:w-8 ${isLocked ? "text-[#9f9f9f]" : "text-[#ffd36b]"}`} />
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </section>

          <section className="mt-4 overflow-hidden rounded-[24px] border border-[#5f4310] bg-[linear-gradient(180deg,#0d0d0d_0%,#090909_100%)] p-4 sm:p-5">
            <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_180px] lg:items-center">
              <div>
                <h3 className="text-xl font-black text-[#f6c85e]">{text.restrictionsTitle}</h3>
                <div className="mt-4 space-y-3">
                  <RestrictionRow icon={Gift} text={text.restrictionFund} />
                  <RestrictionRow icon={Gift} text={text.restrictionMinDeposit} />
                  <RestrictionRow icon={LockKeyhole} text={text.restrictionUnlock} />
                  {!!status?.hidden && (
                    <RestrictionRow icon={ShieldCheck} text={text.completed} highlight />
                  )}
                </div>
              </div>

              <div className="relative hidden h-[180px] lg:block">
                <div className="absolute inset-0 flex items-center justify-center rounded-[30px] bg-[radial-gradient(circle,_rgba(255,196,72,0.16),transparent_60%)]" />
                <ShieldCheck className="absolute left-6 top-4 h-28 w-28 text-[#f1be54]" strokeWidth={1.5} />
                <LockKeyhole className="absolute bottom-3 right-5 h-20 w-20 text-[#f1be54]" strokeWidth={1.5} />
                <Gift className="absolute bottom-10 left-10 h-16 w-16 text-[#f1be54]" strokeWidth={1.5} />
              </div>
            </div>
          </section>

          {!status?.hidden ? (
            <button
              type="button"
              onClick={() => router.push("/wallet")}
              className="mt-4 grid w-full grid-cols-[1fr_auto_1fr] items-center rounded-full border border-[#b98113] bg-[linear-gradient(180deg,#ffd34f_0%,#ffb700_58%,#d58c00_100%)] px-5 py-4 text-sm font-black uppercase tracking-[0.04em] text-[#231100] shadow-[0_18px_40px_rgba(232,169,27,0.22)] transition-transform hover:scale-[1.01] active:scale-[0.985] sm:px-7 sm:text-lg"
            >
              <span aria-hidden="true" />
              <span className="text-center">{text.depositNow}</span>
              <span className="ml-auto flex h-9 w-9 items-center justify-center rounded-full bg-[#231100] text-[#ffcf55]">
                <ChevronRight className="h-5 w-5" />
              </span>
            </button>
          ) : (
            <div className="mt-4 rounded-full border border-emerald-500/25 bg-emerald-500/10 px-5 py-4 text-center text-sm font-semibold text-emerald-300 sm:text-base">
              {text.completed}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

const RestrictionRow = ({
  icon: Icon,
  text,
  highlight = false,
}: {
  icon: React.ElementType
  text: string
  highlight?: boolean
}) => (
  <div className={`flex items-start gap-3 text-sm leading-6 ${highlight ? "text-emerald-300" : "text-[#e5c983]"}`}>
    <Icon className={`mt-0.5 h-5 w-5 shrink-0 ${highlight ? "text-emerald-300" : "text-[#efc560]"}`} />
    <span>{text}</span>
  </div>
)

export default NewUserRechargePage
