"use client"

import React, { useEffect, useState } from "react"
import { HelpCircle, History } from "lucide-react"
import { WalletTab } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { useUser } from "@/app/contexts/UserContext"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { couponService } from "@/services/api"
import { DepositForm } from "./payment/DepositForm"
import { PaymentHistory } from "./payment/PaymentHistory"
import { PaymentRegionSelector } from "./payment/PaymentRegionSelector"
import { WithdrawForm } from "./payment/WithdrawForm"
import { getCurrentForcedPaymentRegion, type PaymentRegion } from "./payment/paymentRegion"

interface WalletPageProps {
  activeTab: WalletTab
  onTabChange: (tab: WalletTab) => void
}

const WalletPage: React.FC<WalletPageProps> = ({ activeTab, onTabChange }) => {
  const { t, language } = useLanguage()
  const {
    userBalance,
    withdrawableBalance,
    remainingWager,
    depositWagerMultiplier,
    rewardWagerMultiplier,
    totalDeposit,
    fetchUserBalance,
    isUserInfoLoaded,
  } = useUser()
  const { formatFromUSDTo, formatUSD } = useExchangeRate()
  const [balanceLoading, setBalanceLoading] = useState(!isUserInfoLoaded)
  const [insuranceCouponCount, setInsuranceCouponCount] = useState(0)
  const [showInsuranceInfo, setShowInsuranceInfo] = useState(false)
  const [forcedRegion] = useState<PaymentRegion | null>(() => getCurrentForcedPaymentRegion())
  const [selectedRegion, setSelectedRegion] = useState<PaymentRegion>(() => forcedRegion ?? (language === "id" ? "ID" : "PH"))
  const strictRegionOnly = forcedRegion !== null
  const loading = !isUserInfoLoaded && balanceLoading
  const balanceCurrency = selectedRegion === "ID" ? "IDR" : "PHP"
  const primaryBalanceUSD = activeTab === "withdraw" ? withdrawableBalance : userBalance
  const primaryBalanceLabel = activeTab === "withdraw" ? t("wallet.available_withdraw") : t("wallet.balance_title")

  useEffect(() => {
    if (isUserInfoLoaded || !balanceLoading) {
      return
    }

    let cancelled = false

    const loadBalance = async () => {
      try {
        
        await fetchUserBalance()
      } catch (error) {
        console.error("Failed to load wallet balance:", error)
      } finally {
        if (!cancelled) {
          setBalanceLoading(false)
        }
      }
    }

    void loadBalance()

    return () => {
      cancelled = true
    }
  }, [balanceLoading, fetchUserBalance, isUserInfoLoaded])

  useEffect(() => {
    let cancelled = false

    const loadInsuranceCoupons = async () => {
      try {
        const coupons = await couponService.getUserCoupons(1)
        if (cancelled || !Array.isArray(coupons)) {
          return
        }

        setInsuranceCouponCount(
          coupons.filter((coupon) => coupon?.activity_type === "add_desktop_insurance").length,
        )
      } catch (error) {
        if (!cancelled) {
          console.error("Failed to load insurance coupon count:", error)
          setInsuranceCouponCount(0)
        }
      }
    }

    void loadInsuranceCoupons()

    return () => {
      cancelled = true
    }
  }, [])

  const refreshBalance = async () => {
    try {
      await fetchUserBalance()
    } catch (error) {
      console.error("Failed to refresh wallet balance:", error)
    } finally {
      setBalanceLoading(false)
    }
  }

  const handleDepositSuccess = () => {
    void refreshBalance()
  }

  const renderDepositContent = () => (
    <div className="animate-fade-in">
      <DepositForm
        key={`deposit-${selectedRegion}-${strictRegionOnly ? "locked" : "open"}`}
        selectedRegion={selectedRegion}
        strictRegionOnly={strictRegionOnly}
        onSuccess={handleDepositSuccess}
      />
    </div>
  )

  const renderWithdrawContent = () => (
    <div className="animate-fade-in">
      <WithdrawForm
        key={`withdraw-${selectedRegion}`}
        selectedRegion={selectedRegion}
        totalBalanceUSD={userBalance}
        withdrawableBalanceUSD={withdrawableBalance}
        totalDepositUSD={totalDeposit}
        remainingWager={remainingWager}
        depositWagerMultiplier={depositWagerMultiplier}
        rewardWagerMultiplier={rewardWagerMultiplier}
        onSuccess={refreshBalance}
      />
    </div>
  )

  const renderRecordsContent = () => (
    <div className="animate-fade-in">
      <PaymentHistory />
    </div>
  )

  return (
    <div className="pt-20 pb-24 px-4 max-w-2xl mx-auto animate-fade-in">
      {activeTab !== "records" && (
        !forcedRegion && <PaymentRegionSelector selectedRegion={selectedRegion} onChange={setSelectedRegion} />
      )}

      <div className="flex bg-white/5 rounded-full p-1 mb-6">
        <button
          onClick={() => onTabChange("deposit")}
          className={`flex-1 py-2 rounded-full font-bold text-sm transition-colors ${
            activeTab === "deposit" ? "bg-lucky-gold text-lucky-dark shadow-lg" : "text-gray-400 hover:text-white"
          }`}
        >
          {t("wallet.tab.deposit")}
        </button>
        <button
          onClick={() => onTabChange("withdraw")}
          className={`flex-1 py-2 rounded-full font-bold text-sm transition-colors ${
            activeTab === "withdraw" ? "bg-lucky-gold text-lucky-dark shadow-lg" : "text-gray-400 hover:text-white"
          }`}
        >
          {t("wallet.tab.withdraw")}
        </button>
        <button
          onClick={() => onTabChange("records")}
          className={`flex-1 py-2 rounded-full font-bold text-sm transition-colors flex items-center justify-center gap-1 ${
            activeTab === "records" ? "bg-lucky-gold text-lucky-dark shadow-lg" : "text-gray-400 hover:text-white"
          }`}
        >
          <History size={14} /> {t("wallet.tab.records")}
        </button>
      </div>

      <div className="mb-3 flex items-center justify-between rounded-lg border border-white/10 bg-lucky-dark/90 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-gray-200">
          <span>{t("wallet.insurance_coupon_label")}</span>
          <button
            type="button"
            onClick={() => setShowInsuranceInfo(true)}
            className="inline-flex h-5 w-5 items-center justify-center rounded-full border border-lucky-gold/50 text-lucky-gold transition-colors hover:bg-lucky-gold hover:text-lucky-dark focus:outline-none focus:ring-2 focus:ring-lucky-gold/60"
            aria-label={t("wallet.insurance_coupon_help")}
          >
            <HelpCircle size={13} />
          </button>
        </div>
        <span className="text-sm font-bold text-lucky-gold">
          {t("wallet.insurance_coupon_count").replace("{count}", insuranceCouponCount.toString())}
        </span>
      </div>

      <div className="bg-gradient-to-r from-lucky-dark to-lucky-purple border border-white/10 rounded-2xl p-6 mb-6 relative overflow-hidden">
        <div className="absolute right-0 top-0 h-full w-1/3 bg-white/5 skew-x-12 transform translate-x-8"></div>
        <div className="relative z-10">
          <div className="text-sm text-gray-400 mb-1">{primaryBalanceLabel}</div>
          <div className="text-3xl font-display font-bold text-white flex items-baseline gap-2">
            {loading ? (
              <span className="text-gray-400">--</span>
            ) : (
              <>{formatUSD(primaryBalanceUSD)}</>
            )}
          </div>
          <div className="text-xs text-gray-500 mt-1">
            {activeTab === "withdraw" ? `${t("wallet.balance_title")}: ${formatFromUSDTo(userBalance, balanceCurrency)}` : "USD"}
          </div>
        </div>
      </div>

      {activeTab === "deposit" && renderDepositContent()}
      {activeTab === "withdraw" && renderWithdrawContent()}
      {activeTab === "records" && renderRecordsContent()}

      {showInsuranceInfo && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4">
          <div className="w-full max-w-sm rounded-xl border border-lucky-gold/30 bg-lucky-dark p-5 shadow-2xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-lg font-bold text-white">{t("wallet.insurance_coupon_title")}</h3>
              <button
                type="button"
                onClick={() => setShowInsuranceInfo(false)}
                className="text-2xl leading-none text-gray-400 transition-colors hover:text-white"
                aria-label={t("common.close")}
              >
                ×
              </button>
            </div>
            <p className="text-sm leading-6 text-gray-300">{t("wallet.insurance_coupon_desc")}</p>
            <button
              type="button"
              onClick={() => setShowInsuranceInfo(false)}
              className="mt-5 w-full rounded-full bg-lucky-gold px-4 py-2 font-bold text-lucky-dark transition-transform hover:scale-[1.01]"
            >
              {t("common.confirm")}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default WalletPage
