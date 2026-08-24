"use client"

import React, { useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2 } from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useUser } from "@/app/contexts/UserContext"
import { activityService, ActivityClaimRecordResponse, GAME_TRANSACTIONS_UPDATED_EVENT, PlayerTransactionOverview, VipMonthlyBonusStatusResponse } from "@/services/api"

interface VipTierConfig {
  level: number
  recharge: number
  wager: number
  rebateRate: string
  vipImage: string
  privilegeIcon: string
  privilegeKey: string
}

const VIP_TIERS: VipTierConfig[] = [
  { level: 1, recharge: 100, wager: 1000, rebateRate: "5%", vipImage: "/images/activity8/vip1.png", privilegeIcon: "/images/activity8/zu1.png", privilegeKey: "vip.privilege.1" },
  { level: 2, recharge: 1000, wager: 10000, rebateRate: "15%", vipImage: "/images/activity8/vip2.png", privilegeIcon: "/images/activity8/zu2.png", privilegeKey: "vip.privilege.2" },
  { level: 3, recharge: 5000, wager: 50000, rebateRate: "30%", vipImage: "/images/activity8/vip3.png", privilegeIcon: "/images/activity8/zu3.png", privilegeKey: "vip.privilege.3" },
  { level: 4, recharge: 10000, wager: 100000, rebateRate: "45%", vipImage: "/images/activity8/vip4.png", privilegeIcon: "/images/activity8/zu4.png", privilegeKey: "vip.privilege.4" },
  { level: 5, recharge: 50000, wager: 500000, rebateRate: "60%", vipImage: "/images/activity8/vip5.png", privilegeIcon: "/images/activity8/zu5.png", privilegeKey: "vip.privilege.5" },
  { level: 6, recharge: 100000, wager: 1000000, rebateRate: "100%", vipImage: "/images/activity8/vip6.png", privilegeIcon: "/images/activity8/zu6.png", privilegeKey: "vip.privilege.6" },
]

const panelClass =
  "overflow-hidden rounded-[28px] border border-[#7a5921] bg-[radial-gradient(circle_at_top,rgba(70,53,25,0.55),rgba(10,10,10,0.98)_55%)] shadow-[0_22px_70px_rgba(0,0,0,0.42)]"
const VIP_ZERO_IMAGE = "/images/activity8/vip0.png"

const clampVipLevel = (level: number) => {
  if (level <= 0) {
    return 0
  }

  if (level >= 6) {
    return 6
  }

  return level
}

const VipPage: React.FC = () => {
  const router = useRouter()
  const { language, t } = useLanguage()
  const { userProfile, isUserInfoLoaded, totalDeposit, fetchUserBalance } = useUser()
  const [overview, setOverview] = useState<PlayerTransactionOverview | null>(null)
  const [monthlyBonusStatus, setMonthlyBonusStatus] = useState<VipMonthlyBonusStatusResponse | null>(null)
  const [isClaimingMonthlyBonus, setIsClaimingMonthlyBonus] = useState(false)

  const isIndonesian = language === "id"
  const locale = isIndonesian ? "id-ID" : "en-US"
  const titleImage = isIndonesian ? "/images/activity8/yinnititle.png" : "/images/activity8/yingyutitle.png"
  const actualVipLevel = userProfile.vipLevel ?? 0
  const currentVipLevel = clampVipLevel(actualVipLevel)
  const currentTier = VIP_TIERS.find((tier) => tier.level === Math.max(currentVipLevel, 1)) ?? VIP_TIERS[0]
  const currentLevelImage = currentVipLevel === 0 ? VIP_ZERO_IMAGE : currentTier.vipImage
  const nextTier =
    VIP_TIERS.find((tier) => tier.level === (currentVipLevel <= 0 ? 1 : Math.min(currentVipLevel + 1, 6))) ??
    VIP_TIERS[VIP_TIERS.length - 1]
  const progressTargetTier = currentVipLevel >= 6 ? currentTier : nextTier

  const loadOverview = React.useCallback(async (sync = false) => {
    try {
      const nextOverview = await activityService.getPlayerTransactionOverview(sync)
      setOverview(nextOverview)
    } catch (error) {
      console.error("Failed to load VIP overview:", error)
      setOverview(null)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    void loadOverview()

    const handleGameTransactionsUpdated = () => {
      if (!cancelled) {
        void loadOverview()
      }
    }
    window.addEventListener(GAME_TRANSACTIONS_UPDATED_EVENT, handleGameTransactionsUpdated)

    return () => {
      cancelled = true
      window.removeEventListener(GAME_TRANSACTIONS_UPDATED_EVENT, handleGameTransactionsUpdated)
    }
  }, [loadOverview])

  useEffect(() => {
    if (!isUserInfoLoaded) {
      return
    }

    let cancelled = false

    const loadMonthlyBonusStatus = async () => {
      try {
        const nextStatus = await activityService.getVipMonthlyBonusStatus()
        if (!cancelled) {
          setMonthlyBonusStatus(nextStatus)
        }
      } catch (error) {
        console.error("Failed to load VIP monthly bonus status:", error)
        if (!cancelled) {
          setMonthlyBonusStatus(null)
        }
      }
    }

    void loadMonthlyBonusStatus()

    return () => {
      cancelled = true
    }
  }, [isUserInfoLoaded])

  const handleClaimMonthlyBonus = async () => {
    if (!monthlyBonusStatus?.can_claim || isClaimingMonthlyBonus) {
      return
    }

    try {
      setIsClaimingMonthlyBonus(true)
      const claimResult: ActivityClaimRecordResponse = await activityService.claimVipMonthlyBonus()

      setMonthlyBonusStatus((currentStatus) => ({
        activity_type: currentStatus?.activity_type ?? "vip_monthly_bonus",
        current_vip_level: currentStatus?.current_vip_level ?? actualVipLevel,
        claim_month: currentStatus?.claim_month ?? "",
        reward_amount: claimResult.reward_amount,
        can_claim: false,
        claimed: true,
        claimed_at: claimResult.claimed_at,
        balance: claimResult.balance,
      }))

      await fetchUserBalance()
    } catch (error) {
      console.error("Failed to claim VIP monthly bonus:", error)
    } finally {
      setIsClaimingMonthlyBonus(false)
    }
  }

  const formatU = (amount: number) =>
    amount.toLocaleString(locale, {
      minimumFractionDigits: Number.isInteger(amount) ? 0 : 2,
      maximumFractionDigits: 2,
    })

  const totalWagerU = useMemo(() => {
    const turnover = overview?.total?.summary?.turnover_u ?? 0
    const bet = overview?.total?.summary?.bet_u ?? 0
    return turnover > 0 ? turnover : bet
  }, [overview])

  const resolvedDepositU = Math.max(totalDeposit, actualVipLevel > 0 ? currentTier.recharge : 0)
  const resolvedWagerU = Math.max(totalWagerU, actualVipLevel > 0 ? currentTier.wager : 0)

  const rechargePercent = Math.min(
    100,
    progressTargetTier.recharge > 0 ? Math.round((resolvedDepositU / progressTargetTier.recharge) * 100) : 100,
  )
  const wagerPercent = Math.min(
    100,
    progressTargetTier.wager > 0 ? Math.round((resolvedWagerU / progressTargetTier.wager) * 100) : 100,
  )

  const monthlyBonusU = actualVipLevel >= 2 ? currentTier.recharge * 0.1 : 0
  const isMonthlyClaimed = monthlyBonusStatus?.claimed ?? false
  const canClaimMonthlyBonus = monthlyBonusStatus?.can_claim ?? false
  const displayedMonthlyBonusU = monthlyBonusStatus?.reward_amount ?? monthlyBonusU
  const monthlyButtonImage = canClaimMonthlyBonus ? "/images/activity8/btn1.png" : "/images/activity8/btn2.png"

  if (!isUserInfoLoaded) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-[#080808] via-[#121212] to-[#050505] px-4">
        <div className="flex items-center gap-3 rounded-2xl border border-lucky-gold/20 bg-black/45 px-5 py-4 text-lucky-gold">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>{t("common.loading")}</span>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,rgba(72,53,17,0.14),rgba(7,7,7,1)_38%),linear-gradient(180deg,#030303_0%,#090909_100%)] px-2.5 py-3 sm:px-5 sm:py-6">
      <div className="mx-auto flex w-full max-w-[980px] flex-col gap-4">
        <section className={`${panelClass} relative px-3 pb-4 pt-4 sm:px-5 sm:pb-5 sm:pt-5`}>
          <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,210,95,0.18),transparent_34%),linear-gradient(135deg,rgba(255,215,116,0.08),transparent_35%,transparent_65%,rgba(255,215,116,0.05))]" />
          <div className="pointer-events-none absolute inset-x-[8%] top-0 h-px bg-gradient-to-r from-transparent via-[#f5cf71] to-transparent opacity-80" />
          <div className="relative z-10 grid grid-cols-[64px_minmax(0,1fr)_78px] items-center gap-1.5 sm:grid-cols-[120px_minmax(0,1fr)_160px] sm:gap-4">
            <div className="flex justify-start">
              <img
                src="/images/activity8/top1.png"
                alt=""
                className="w-[74px] max-w-none object-contain drop-shadow-[0_16px_26px_rgba(0,0,0,0.42)] sm:w-[132px]"
              />
            </div>

            <div className="min-w-0 text-center">
              <img
                src={titleImage}
                alt={t("nav.vip")}
                className="mx-auto w-full max-w-[900px] object-contain drop-shadow-[0_14px_28px_rgba(255,202,92,0.2)]"
              />
              <div className="mx-auto mt-2 h-px w-[92%] bg-gradient-to-r from-transparent via-[#f1c861]/65 to-transparent sm:mt-3" />
              <p className="mx-auto mt-2 max-w-[680px] text-[10px] font-bold leading-relaxed text-[#f6d57e] sm:mt-3 sm:text-lg">
                {t("vip.hero_subtitle")}
              </p>
            </div>

            <div className="flex justify-end">
              <img
                src="/images/activity8/top2.png"
                alt=""
                className="w-[88px] max-w-none object-contain drop-shadow-[0_18px_28px_rgba(0,0,0,0.45)] sm:w-[180px]"
              />
            </div>
          </div>
        </section>

        <section className={`${panelClass} p-2 sm:p-3`}>
          <div className="grid grid-cols-[124px_minmax(0,1fr)] gap-1.5 max-[390px]:grid-cols-[116px_minmax(0,1fr)] sm:grid-cols-[220px_minmax(0,1fr)] sm:gap-3">
            <div className="rounded-[18px] border border-[#6d501d] bg-[linear-gradient(180deg,rgba(42,31,14,0.96),rgba(10,10,10,0.96))] p-2 text-center sm:rounded-[22px] sm:p-3">
              <div className="mx-auto inline-flex rounded-full bg-[linear-gradient(180deg,#d9b968,#8b6322)] px-2.5 py-1 text-[10px] font-black text-[#221304] sm:px-4 sm:py-1.5 sm:text-sm">
                {t("vip.current_level")}
              </div>

              <div className="mt-2 rounded-[18px] border border-[#815e20] bg-[radial-gradient(circle_at_top,rgba(255,220,132,0.16),rgba(0,0,0,0.92)_68%)] p-2 sm:mt-3 sm:rounded-[24px] sm:p-3">
                <img
                  src={currentLevelImage}
                  alt={`VIP ${Math.max(currentVipLevel, 0)}`}
                  className="mx-auto h-auto w-full max-w-[90px] object-contain sm:max-w-[152px]"
                />
                <p className="mt-1 text-sm font-black text-[#ffd96f] sm:mt-1.5 sm:text-xl">VIP {currentVipLevel}</p>
              </div>

              <p className="mt-2 rounded-[14px] border border-[#5a4319] bg-black/35 px-2 py-1.5 text-[10px] font-semibold leading-snug text-[#f4d37a] sm:mt-3 sm:rounded-full sm:px-3 sm:text-xs">
                {currentVipLevel >= 6 ? t("vip.max_level") : t("vip.next_level_gap")}
              </p>
            </div>

            <div className="min-w-0 rounded-[18px] border border-[#6d501d] bg-[linear-gradient(180deg,rgba(17,17,17,0.98),rgba(9,9,9,0.98))] p-2 sm:rounded-[22px] sm:p-4">
              <ProgressBlock
                title={t("vip.recharge_progress")}
                icon="/images/activity8/cash.png"
                value={resolvedDepositU}
                target={progressTargetTier.recharge}
                percent={rechargePercent}
                locale={locale}
              />

              <ProgressBlock
                title={t("vip.wager_progress")}
                icon="/images/activity8/couma1.png"
                value={resolvedWagerU}
                target={progressTargetTier.wager}
                percent={wagerPercent}
                locale={locale}
              />

              <button
                type="button"
                onClick={() => router.push("/wallet?tab=deposit")}
                className="relative mx-auto mt-3 block w-full max-w-[176px] transition-transform hover:scale-[1.01] active:scale-[0.985] sm:mt-5 sm:max-w-[270px]"
              >
                <img src="/images/activity8/btn1.png" alt={t("vip.deposit_now")} className="block h-auto w-full" />
                <span className="absolute inset-0 flex items-center justify-center text-xs font-black text-[#2b1800] sm:text-xl">
                  {t("vip.deposit_now")}
                </span>
              </button>
            </div>
          </div>
        </section>

        <section className={`${panelClass} hidden overflow-x-auto lg:block`}>
          <table className="min-w-[760px] w-full border-collapse">
            <thead>
              <tr className="border-b border-[#4b3815] bg-[linear-gradient(180deg,rgba(78,59,24,0.5),rgba(20,17,10,0.92))] text-[#f5d77d]">
                <th className="border-r border-[#302413] px-3 py-3 text-center text-xs font-black xl:px-4 xl:text-sm">{t("nav.vip")}</th>
                <th className="border-r border-[#302413] px-3 py-3 text-left text-xs font-black xl:px-4 xl:text-sm">{t("vip.recharge_amount")}</th>
                <th className="border-r border-[#302413] px-3 py-3 text-left text-xs font-black xl:px-4 xl:text-sm">{t("vip.wager_amount")}</th>
                <th className="border-r border-[#302413] px-3 py-3 text-left text-xs font-black xl:px-4 xl:text-sm">{t("vip.rebate_rate")}</th>
                <th className="px-3 py-3 text-left text-xs font-black xl:px-4 xl:text-sm">{t("vip.exclusive_privileges")}</th>
              </tr>
            </thead>
            <tbody>
              {VIP_TIERS.map((tier) => {
                const isCurrent = currentVipLevel > 0 && tier.level === currentVipLevel
                return (
                  <tr
                    key={tier.level}
                    className={`border-b border-[#302413] ${isCurrent ? "bg-[linear-gradient(90deg,rgba(255,217,112,0.08),rgba(255,217,112,0.01))]" : "bg-[rgba(7,7,7,0.28)]"}`}
                  >
                    <td className="border-r border-[#302413] px-3 py-2.5 xl:px-4 xl:py-3">
                      <div className="flex items-center justify-center gap-2">
                        <img src={tier.vipImage} alt={`VIP ${tier.level}`} className="h-10 w-10 object-contain xl:h-11 xl:w-11" />
                        <span className={`text-xs font-black xl:text-sm ${isCurrent ? "text-[#ffd96e]" : "text-white"}`}>VIP {tier.level}</span>
                      </div>
                    </td>
                    <td className="border-r border-[#302413] px-3 py-2.5 xl:px-4 xl:py-3">
                      <ValueCell icon="/images/activity8/cash.png" value={`${formatU(tier.recharge)}U`} />
                    </td>
                    <td className="border-r border-[#302413] px-3 py-2.5 xl:px-4 xl:py-3">
                      <ValueCell icon="/images/activity8/couma1.png" value={`${formatU(tier.wager)}U`} />
                    </td>
                    <td className={`border-r border-[#302413] px-3 py-2.5 text-base font-black xl:px-4 xl:py-3 xl:text-lg ${getRebateColor(tier.level)}`}>{tier.rebateRate}</td>
                    <td className="px-3 py-2.5 xl:px-4 xl:py-3">
                      <div className="flex items-center gap-2">
                        <img src={tier.privilegeIcon} alt="" className="h-6 w-6 rounded-full object-contain xl:h-7 xl:w-7" />
                        <span className="text-xs font-semibold text-white/90 xl:text-sm">{t(tier.privilegeKey)}</span>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </section>

        <section className={`${panelClass} overflow-hidden lg:hidden`}>
          <div className="border-b border-[#4b3815] bg-[linear-gradient(180deg,rgba(78,59,24,0.5),rgba(20,17,10,0.92))] px-2.5 py-2">
            <div className="grid grid-cols-[58px_minmax(0,1fr)_68px] gap-2 text-[10px] font-black text-[#f5d77d]">
              <span>{t("nav.vip")}</span>
              <span>{t("vip.exclusive_privileges")}</span>
              <span className="text-right">{t("vip.rebate_rate")}</span>
            </div>
          </div>

          <div>
            {VIP_TIERS.map((tier) => {
              const isCurrent = currentVipLevel > 0 && tier.level === currentVipLevel
              return (
                <div
                  key={`mobile-tier-${tier.level}`}
                  className={`border-b border-[#302413] px-2.5 py-2.5 ${
                    isCurrent ? "bg-[linear-gradient(90deg,rgba(255,217,112,0.07),rgba(255,217,112,0.01))]" : "bg-[rgba(7,7,7,0.24)]"
                  }`}
                >
                  <div className="grid grid-cols-[58px_minmax(0,1fr)_68px] gap-2">
                    <div className="flex flex-col items-center justify-center">
                      <img src={tier.vipImage} alt={`VIP ${tier.level}`} className="h-9 w-9 object-contain" />
                      <span className={`mt-0.5 text-[10px] font-black ${isCurrent ? "text-[#ffd96e]" : "text-white"}`}>VIP {tier.level}</span>
                    </div>

                    <div className="min-w-0 py-0.5">
                      <div className="grid grid-cols-2 gap-x-2 gap-y-1">
                        <div className="min-w-0">
                          <p className="text-[8px] font-semibold text-[#cfae67]">{t("vip.recharge_amount")}</p>
                          <div className="mt-0.5 flex items-center gap-1">
                            <img src="/images/activity8/cash.png" alt="" className="h-3.5 w-3.5 object-contain" />
                            <span className="truncate text-[10px] font-black text-white">{formatU(tier.recharge)}U</span>
                          </div>
                        </div>

                        <div className="min-w-0">
                          <p className="text-[8px] font-semibold text-[#cfae67]">{t("vip.wager_amount")}</p>
                          <div className="mt-0.5 flex items-center gap-1">
                            <img src="/images/activity8/couma1.png" alt="" className="h-3.5 w-3.5 object-contain" />
                            <span className="truncate text-[10px] font-black text-white">{formatU(tier.wager)}U</span>
                          </div>
                        </div>
                      </div>

                      <div className="mt-1.5 flex min-w-0 items-center gap-1.5">
                        <img src={tier.privilegeIcon} alt="" className="h-5 w-5 object-contain" />
                        <span className="truncate text-[10px] font-semibold text-white/88">{t(tier.privilegeKey)}</span>
                      </div>
                    </div>

                    <div className="flex items-center justify-end">
                      <span className={`text-[13px] font-black ${getRebateColor(tier.level)}`}>{tier.rebateRate}</span>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </section>

        <section className={`${panelClass} p-3 sm:p-5`}>
          <div className="grid grid-cols-[66px_minmax(0,1fr)_124px] items-center gap-2.5 px-1 py-1 sm:grid-cols-[106px_minmax(0,1fr)_220px] sm:gap-5">
            <div className="flex justify-center sm:justify-start">
              <img src="/images/activity8/top1.png" alt="" className="h-[54px] w-[54px] shrink-0 object-contain sm:h-[92px] sm:w-[92px]" />
            </div>

            <div className="min-w-0">
              <h2 className="text-base font-black leading-tight text-[#f7d06c] sm:text-[1.75rem]">{t("vip.monthly_bonus")}</h2>
              <p className="mt-1 text-[10px] leading-relaxed text-white/72 sm:mt-1.5 sm:text-[13px]">{t("vip.monthly_bonus_desc")}</p>
            </div>

            <div className="text-center">
              <p className="text-[10px] font-semibold text-[#d7b866] sm:text-sm">{t("vip.claimable_amount")}</p>
              <p className="mt-1 text-lg font-black text-[#ffd76f] sm:mt-1.5 sm:text-[2rem]">
                {displayedMonthlyBonusU > 0 ? `${formatU(displayedMonthlyBonusU)}U` : "--"}
              </p>

              <button
                type="button"
                onClick={handleClaimMonthlyBonus}
                disabled={!canClaimMonthlyBonus || isClaimingMonthlyBonus}
                className={`relative mx-auto mt-2 block w-full max-w-[112px] transition-transform sm:mt-2.5 sm:max-w-[190px] ${
                  !canClaimMonthlyBonus || isClaimingMonthlyBonus ? "cursor-not-allowed" : "hover:scale-[1.01] active:scale-[0.985]"
                }`}
              >
                <img
                  src={monthlyButtonImage}
                  alt={isMonthlyClaimed ? t("vip.claimed_bonus") : t("vip.claim_bonus")}
                  className={`block h-auto w-full ${!canClaimMonthlyBonus ? "grayscale opacity-70" : ""}`}
                />
                <span className="absolute inset-0 flex items-center justify-center px-2 text-[9px] font-black text-[#2b1800] sm:px-4 sm:text-base">
                  {actualVipLevel < 2 ? t("vip.bonus_locked") : isClaimingMonthlyBonus ? t("common.loading") : isMonthlyClaimed ? t("vip.claimed_bonus") : t("vip.claim_bonus")}
                </span>
              </button>
            </div>
          </div>
        </section>

      </div>
    </div>
  )
}

interface ProgressBlockProps {
  title: string
  icon: string
  value: number
  target: number
  percent: number
  locale: string
}

const ProgressBlock: React.FC<ProgressBlockProps> = ({ title, icon, value, target, percent, locale }) => {
  const formatValue = (amount: number) =>
    amount.toLocaleString(locale, {
      minimumFractionDigits: Number.isInteger(amount) ? 0 : 2,
      maximumFractionDigits: 2,
    })

  return (
    <div className="mb-3.5 last:mb-0 sm:mb-5">
      <div className="flex items-start gap-2 sm:gap-3">
        <div>
          <p className="text-sm font-black text-[#f7d06c] sm:text-2xl">{title}</p>
          <div className="mt-2 flex items-center gap-2 sm:mt-2.5 sm:gap-2.5">
            <img src={icon} alt="" className="h-7 w-7 object-contain sm:h-11 sm:w-11" />
            <p className="text-[clamp(0.95rem,4vw,2.35rem)] font-black leading-none text-[#ffd96f]">
              {formatValue(value)}
              <span className="ml-1 text-[clamp(0.68rem,2.6vw,1.3rem)] text-white/85 sm:ml-1.5">/ {formatValue(target)}U</span>
            </p>
          </div>
        </div>
      </div>

      <div className="mt-2.5 flex items-center gap-3 sm:mt-3">
        <div className="relative h-3 flex-1 overflow-hidden rounded-full border border-[#553d17] bg-[#1d160c] sm:h-4">
          <div
            className="absolute inset-y-0 left-0 rounded-full bg-[linear-gradient(90deg,#a7680c_0%,#f2c95f_35%,#ffd970_100%)] shadow-[0_0_18px_rgba(255,215,112,0.36)]"
            style={{ width: `${percent}%` }}
          />
        </div>
      </div>
    </div>
  )
}

const ValueCell = ({ icon, value }: { icon: string; value: string }) => (
  <div className="flex items-center gap-1.5 xl:gap-2">
    <img src={icon} alt="" className="h-5 w-5 object-contain xl:h-6 xl:w-6" />
    <span className="text-xs font-bold text-white/92 xl:text-sm">{value}</span>
  </div>
)

const getRebateColor = (level: number) => {
  switch (level) {
    case 1:
      return "text-[#55ff4a]"
    case 2:
      return "text-[#39a9ff]"
    case 3:
      return "text-[#c45cff]"
    case 4:
      return "text-[#ffbf3d]"
    case 5:
      return "text-[#ff4c42]"
    default:
      return "text-[#ffe08d]"
  }
}

export default VipPage
