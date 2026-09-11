"use client"

import React, { useEffect, useState } from "react"
import { HelpCircle, History, ChevronRight } from "lucide-react"
import { useRouter } from "next/navigation"
import { WalletTab } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { useUser } from "@/app/contexts/UserContext"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { couponService } from "@/services/api"
import { PaymentHistory } from "./payment/PaymentHistory"
import { WithdrawForm } from "./payment/WithdrawForm"
import { getCurrentForcedPaymentRegion, type PaymentRegion } from "./payment/paymentRegion"
import { paymentService } from "@/services/payment"
import { useToast } from "@/hooks/use-toast"
import { TokenPayModal } from "./payment/TokenPayModal"
import { TonConnectButton, useTonConnectUI } from "@tonconnect/ui-react"
import { beginCell } from "@ton/core"
import { useNewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"

interface WalletPageProps {
  activeTab: WalletTab
  onTabChange: (tab: WalletTab) => void
}

interface TelegramWebApp {
  initData?: string
  platform?: string
  version?: string
  isVersionAtLeast?: (version: string) => boolean
  openInvoice?: (url: string, callback?: (status: string) => void) => void
  openTelegramLink?: (url: string) => void
}

const getTelegramWebApp = () =>
  (window as typeof window & { Telegram?: { WebApp?: TelegramWebApp } }).Telegram?.WebApp

const getTelegramInitData = () => {
  const webApp = getTelegramWebApp()
  if (webApp?.initData) return webApp.initData
  const read = (raw: string) => new URLSearchParams(raw).get("tgWebAppData") || ""
  return read(window.location.search) || read(window.location.hash.replace(/^#/, ""))
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
  const { formatUSD } = useExchangeRate()
  const [balanceLoading, setBalanceLoading] = useState(!isUserInfoLoaded)
  const [insuranceCouponCount, setInsuranceCouponCount] = useState(0)
  const [showInsuranceInfo, setShowInsuranceInfo] = useState(false)
  const [forcedRegion] = useState<PaymentRegion | null>(() => getCurrentForcedPaymentRegion())
  const [selectedRegion] = useState<PaymentRegion>(() => forcedRegion ?? (language === "id" ? "ID" : "PH"))
  const telegramMode = true
  const [telegramAmount, setTelegramAmount] = useState("5")
  const [telegramPayLoading, setTelegramPayLoading] = useState(false)
  const [depositMethod, setDepositMethod] = useState<"telegram" | "ton" | "usdt">("telegram")
  const [tonAmount, setTonAmount] = useState("5")
  const [tonPayLoading, setTonPayLoading] = useState(false)
  const [tonRate, setTonRate] = useState(5)
  const [tonWithdrawConditionsEnabled, setTonWithdrawConditionsEnabled] = useState(true)
  const [tonMinWithdrawAmount, setTonMinWithdrawAmount] = useState(10)
  const [usdtAmount, setUsdtAmount] = useState("10")
  const [usdtOrder, setUsdtOrder] = useState<{ pay_address: string; pay_amount: string; expired_at: number } | null>(null)
  const [usdtLoading, setUsdtLoading] = useState(false)
  const [tonAddress, setTonAddress] = useState("")
  const [tonWithdrawAmount, setTonWithdrawAmount] = useState("")
  const [tonWithdrawLoading, setTonWithdrawLoading] = useState(false)
  const { toast } = useToast()
  const router = useRouter()
  const { status: newUserRechargeStatus } = useNewUserRechargeStatus(activeTab === "deposit")
  const newUserRechargeTotal = newUserRechargeStatus?.tiers.length || 4
  const newUserRechargeProgress = Math.min(newUserRechargeStatus?.progress_count || 0, newUserRechargeTotal)
  const [tonConnectUI] = useTonConnectUI()
  useEffect(() => {
    if (depositMethod !== "ton" && activeTab !== "withdraw") return
    paymentService.getTonConfig().then((config) => {
      setTonRate(config.usdPerTon)
      setTonWithdrawConditionsEnabled(config.withdrawConditionsEnabled)
      setTonMinWithdrawAmount(config.minWithdrawAmountUSD)
    }).catch(() => undefined)
  }, [depositMethod, activeTab])
  const strictRegionOnly = forcedRegion !== null
  const loading = !isUserInfoLoaded && balanceLoading
  const primaryBalanceUSD = activeTab === "withdraw" ? withdrawableBalance : userBalance
  const primaryBalanceLabel = activeTab === "withdraw" ? t("wallet.available_withdraw") : t("wallet.balance_title")
  const tonAvailableBalance = tonWithdrawConditionsEnabled ? withdrawableBalance : userBalance

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

  const waitForTelegramPayment = async (orderID: string) => {
    for (let attempt = 0; attempt < 30; attempt += 1) {
      const order = await paymentService.getPaymentStatus(orderID)
      if (order.status === 1) { await refreshBalance(); toast({ title: t("wallet.order_success") }); return }
      if (order.status === 2) throw new Error("Telegram Stars payment failed")
      await new Promise((resolve) => window.setTimeout(resolve, 2000))
    }
    toast({ title: "Payment is processing", description: "The balance will update after Telegram confirms the payment." })
  }

  const waitForTonPayment = async (orderID: string) => {
    for (let attempt = 0; attempt < 40; attempt += 1) {
      const order = await paymentService.getPaymentStatus(orderID)
      if (order.status === 1) { await refreshBalance(); toast({ title: "TON payment successful" }); return }
      if (order.status === 2) { toast({ title: "TON payment failed", variant: "destructive" }); return }
      await new Promise((resolve) => window.setTimeout(resolve, 3000))
    }
  }

  const handleTelegramPay = async () => {
    const amount = Number(telegramAmount)
    if (!Number.isInteger(amount) || amount <= 0) return
    const telegram = getTelegramWebApp()
    const initData = getTelegramInitData()
    if (!initData) { toast({ title: "Open in Telegram", description: "Please open PPNetApp from @ppnetbet_bot.", variant: "destructive" }); return }
    setTelegramPayLoading(true)
    try {
      const order = await paymentService.createTelegramStarsOrder(amount, initData)
      if (!order.invoice_url || !/^https:\/\/t\.me\/\$/i.test(order.invoice_url)) {
        throw new Error("Telegram returned an invalid invoice URL")
      }

      const onInvoiceClosed = (status: string) => {
        if (status === "paid" || status === "pending") void waitForTelegramPayment(order.order_id)
      }

      const supportsInvoice = telegram.openInvoice && (!telegram.isVersionAtLeast || telegram.isVersionAtLeast("6.1"))
      if (supportsInvoice) {
        try {
          telegram.openInvoice!(order.invoice_url, onInvoiceClosed)
          return
        } catch (error) {
          console.error("Telegram openInvoice failed; opening the invoice link instead", {
            error,
            platform: telegram.platform,
            version: telegram.version,
          })
        }
      }

      // Older clients may not expose openInvoice; opening the invoice URL
      // still launches Telegram's native payment screen.
      if (telegram.openTelegramLink) telegram.openTelegramLink(order.invoice_url)
      else window.location.assign(order.invoice_url)
    } catch (error) {
      toast({ title: "Telegram payment failed", description: error instanceof Error ? error.message : "Please try again.", variant: "destructive" })
    } finally {
      setTelegramPayLoading(false)
    }
  }

  const handleTonPay = async () => {
    const amount = Number(tonAmount)
    if (!Number.isFinite(amount) || amount <= 0) return
    await tonConnectUI.connectionRestored
    if (!tonConnectUI.connected || !tonConnectUI.account?.address) {
      toast({ title: "Connect your TON wallet first", variant: "destructive" })
      return
    }
    setTonPayLoading(true)
    try {
      const order = await paymentService.createTonOrder(amount, tonConnectUI.account.address)
      const payload = beginCell().storeUint(0, 32).storeStringTail(order.comment).endCell().toBoc().toString("base64")
      const txResult = await tonConnectUI.sendTransaction({ validUntil: Math.min(Math.floor(Date.parse(order.expires_at) / 1000), Math.floor(Date.now() / 1000) + 600), messages: [{ address: order.wallet_addr, amount: String(order.nano_ton), payload }] })
      void txResult
      toast({ title: "Transaction sent", description: "Waiting for blockchain confirmation." })
      // TonConnect returns a signed BOC; the backend scanner resolves its chain hash.
      // Keep polling the order while the scanner validates sender, destination and amount.
      void waitForTonPayment(order.order_id)
    } catch (error) {
      toast({ title: "Gram (TON) payment failed", description: error instanceof Error ? error.message : "Please try again.", variant: "destructive" })
    } finally { setTonPayLoading(false) }
  }

  const handleTONWithdraw = async () => {
    const amount = Number(tonWithdrawAmount)
    if (!tonAddress.trim() || !Number.isFinite(amount) || amount < tonMinWithdrawAmount || amount > tonAvailableBalance) {
      const description = tonAvailableBalance <= 0
        ? `${t("wallet.insufficient_withdrawable_balance")} ${t("wallet.ton_available").replace("{amount}", formatUSD(tonAvailableBalance))}`
        : t("wallet.ton_invalid_desc")
      toast({ title: t("wallet.ton_invalid_title"), description, variant: "destructive" })
      return
    }

    setTonWithdrawLoading(true)
    try {
      await paymentService.createWithdrawOrder({
        amount,
        type: "crypto",
        dst_code: "TON",
        account: tonAddress.trim(),
        account_name: "TON Wallet",
        address: tonAddress.trim(),
        currency: "USD",
        channel: "manual",
      })
      toast({ title: t("wallet.ton_submitted"), description: t("wallet.ton_submitted_desc") })
      setTonAddress("")
      setTonWithdrawAmount("")
      await refreshBalance()
    } catch (error) {
      toast({ title: "Withdrawal failed", description: error instanceof Error ? error.message : "Please try again.", variant: "destructive" })
    } finally {
      setTonWithdrawLoading(false)
    }
  }

  const renderDepositContent = () => (
    <div className="animate-fade-in">
      {depositMethod === "usdt" ? (
        <div className="rounded-xl border border-[#229ED9]/50 bg-[#229ED9]/10 p-4">
          <h3 className="mb-3 font-bold text-white">USDT</h3>
          <p className="mb-3 text-xs text-[#8ed8ff]">Deposit Amount (U)</p>
          <div className="grid grid-cols-3 gap-2">
            {[10, 20, 50, 100, 200, 500].map((value) => (
              <button key={value} type="button" onClick={() => setUsdtAmount(String(value))} className={`rounded-lg border py-3 text-white ${usdtAmount === String(value) ? "border-[#229ED9] bg-[#229ED9]/30" : "border-white/20"}`}>
                {value}U
              </button>
            ))}
          </div>
          <input value={usdtAmount} onChange={(event) => setUsdtAmount(event.target.value)} className="mt-3 w-full rounded-lg bg-black/30 p-3 text-white" type="number" min="1" />
          <button type="button" disabled={usdtLoading || Number(usdtAmount) <= 0} onClick={async () => {
            setUsdtLoading(true)
            try {
              const response = await paymentService.createTokenPayOrder({ amount: Number(usdtAmount), type: "channel", dst_code: "USDT", currency: "USD", callback_url: window.location.origin + "/wallet", chain_type: "TRX" })
              setUsdtOrder(response)
            } catch (error) {
              toast({ title: "USDT payment failed", description: error instanceof Error ? error.message : "Please try again.", variant: "destructive" })
            } finally { setUsdtLoading(false) }
          }} className="mt-3 w-full rounded-lg bg-[#229ED9] py-3 font-bold text-white disabled:opacity-50">{usdtLoading ? "Creating order..." : "Pay with USDT"}</button>
        </div>
      ) : null}
    </div>
  )

  const renderWithdrawContent = () => (
    <div className="animate-fade-in">
      {telegramMode ? (
        <div className="space-y-4 rounded-xl border border-[#229ED9]/50 bg-[#229ED9]/10 p-4">
          <div><h3 className="font-bold text-white">{t("wallet.ton_withdraw_title")}</h3><p className="mt-1 text-xs text-[#8ed8ff]">{t("wallet.ton_available").replace("{amount}", formatUSD(tonAvailableBalance))}</p></div><p className="text-xs text-[#f6c945]">{t("wallet.ton_withdraw_fee")}</p>
          <div><label className="mb-2 block text-sm text-gray-300">{t("wallet.ton_address")}</label><input value={tonAddress} onChange={(event) => setTonAddress(event.target.value)} className="w-full rounded-xl border border-white/15 bg-black/30 px-4 py-3 text-white outline-none focus:border-[#229ED9]" placeholder={t("wallet.ton_address_placeholder")} /></div>
          <div><label className="mb-2 block text-sm text-gray-300">{t("wallet.ton_withdraw_amount")}</label><input value={tonWithdrawAmount} onChange={(event) => setTonWithdrawAmount(event.target.value)} type="number" min={tonMinWithdrawAmount} max={tonAvailableBalance} className="w-full rounded-xl border border-white/15 bg-black/30 px-4 py-3 text-white outline-none focus:border-[#229ED9]" placeholder={`Minimum ${tonMinWithdrawAmount} U`} />{Number(tonWithdrawAmount) > 0 && tonRate > 0 ? <div className="mt-2 text-xs text-[#8ed8ff]"><p>≈ {(Number(tonWithdrawAmount) / tonRate).toFixed(4)} TON</p><p>{t("wallet.ton_withdraw_fee_detail").replace("{ton}", (2 / tonRate).toFixed(4))}</p><p>{t("wallet.ton_withdraw_receive").replace("{ton}", (Math.max(Number(tonWithdrawAmount) - 2, 0) / tonRate).toFixed(4))}</p></div> : null}</div>
          <button type="button" onClick={handleTONWithdraw} disabled={tonWithdrawLoading || !tonAddress.trim() || !tonWithdrawAmount || tonAvailableBalance < tonMinWithdrawAmount} className="w-full rounded-xl bg-[#229ED9] py-3 font-bold text-white disabled:opacity-50">{tonWithdrawLoading ? t("wallet.ton_submitting") : t("wallet.ton_submit")}</button>
        </div>
      ) : <WithdrawForm
        key={`withdraw-${selectedRegion}`}
        selectedRegion={selectedRegion}
        totalBalanceUSD={userBalance}
        withdrawableBalanceUSD={withdrawableBalance}
        totalDepositUSD={totalDeposit}
        remainingWager={remainingWager}
        depositWagerMultiplier={depositWagerMultiplier}
        rewardWagerMultiplier={rewardWagerMultiplier}
        onSuccess={refreshBalance}
      />}
    </div>
  )

  const renderRecordsContent = () => (
    <div className="animate-fade-in">
      <PaymentHistory />
    </div>
  )

  return (
    <div className="pt-2 pb-24 px-4 max-w-2xl mx-auto animate-fade-in">
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

      <div className="bg-gradient-to-r from-lucky-dark to-lucky-purple border border-white/10 rounded-2xl p-3 mb-6 relative overflow-hidden">
        <div className="absolute right-0 top-0 h-full w-1/3 bg-white/5 skew-x-12 transform translate-x-8"></div>
        <div className="relative z-10">
          <div className="text-xs text-gray-400 mb-0.5">{primaryBalanceLabel}</div>
          <div className="text-2xl font-display font-bold text-white flex items-baseline gap-2 leading-tight">
            {loading ? (
              <span className="text-gray-400">--</span>
            ) : (
              <>{formatUSD(primaryBalanceUSD)}</>
            )}
          </div>
          <div className="text-[11px] text-gray-500 mt-0.5">
            {activeTab === "withdraw" ? `${t("wallet.balance_title")}: ${Number(userBalance || 0).toFixed(2)}U` : "USD"}
          </div>
        </div>
      </div>

      {activeTab === "deposit" && (
        <div className="mb-4 grid grid-cols-3 gap-2 rounded-xl border border-white/10 bg-white/5 p-1">
          <button
            type="button"
            onClick={() => setDepositMethod("telegram")}
            className={`rounded-lg py-2 text-sm font-semibold transition-colors ${depositMethod === "telegram" ? "bg-[#229ED9] text-white" : "text-gray-400 hover:text-white"}`}
          >
            Telegram Stars
          </button>
          <button type="button" onClick={() => setDepositMethod("ton")} className={`rounded-lg py-2 text-sm font-semibold transition-colors ${depositMethod === "ton" ? "bg-[#229ED9] text-white" : "text-gray-400 hover:text-white"}`}>Gram (TON)</button>
          <button
            type="button"
            onClick={() => setDepositMethod("usdt")}
            className={`rounded-lg py-2 text-sm font-semibold transition-colors ${depositMethod === "usdt" ? "bg-lucky-gold text-lucky-dark" : "text-gray-400 hover:text-white"}`}
          >
            USDT
          </button>
        </div>
      )}

      {activeTab === "deposit" && newUserRechargeStatus && !newUserRechargeStatus.hidden && (
        <button
          type="button"
          onClick={() => router.push("/newUserRecharge")}
          className="mb-4 flex w-full items-center justify-between rounded-lg border border-[#f2b93a]/70 bg-gradient-to-r from-[#f2b93a]/20 via-[#f2b93a]/10 to-transparent px-3 py-2 text-left text-sm font-semibold text-[#ffd36b] shadow-[0_0_14px_rgba(242,185,58,0.18)] transition hover:border-[#ffd36b] hover:bg-[#f2b93a]/25"
        >
          <span className="flex items-center gap-2"><span className="animate-pulse">🎁</span>{t("wallet.new_user_recharge_hint")} ({newUserRechargeProgress}/{newUserRechargeTotal})</span>
          <ChevronRight className="h-4 w-4 shrink-0" aria-hidden="true" />
        </button>
      )}

      {activeTab === "deposit" && depositMethod === "ton" && (
        <div className="mb-4 rounded-xl border border-[#229ED9]/50 bg-[#229ED9]/10 p-4">
          <h3 className="mb-3 font-bold text-white">Gram (TON)</h3>
          <div className="mb-3"><TonConnectButton /></div>
          <p className="mb-3 text-xs text-[#8ed8ff]">Deposit Amount (USD)</p>
          <div className="grid grid-cols-3 gap-2">{[5,20,50,100,200,500].map(v => <button key={v} type="button" className={`rounded-lg border py-3 text-white ${tonAmount === String(v) ? "border-[#229ED9] bg-[#229ED9]/30" : "border-white/20"}`} onClick={() => setTonAmount(String(v))}>{v}U</button>)}</div>
          <input value={tonAmount} onChange={(event) => setTonAmount(event.target.value)} className="mt-3 w-full rounded-lg bg-black/30 p-3 text-white" type="number" min="1" />
          <p className="mt-2 text-xs text-[#8ed8ff]">≈ {(Number(tonAmount) / tonRate || 0).toFixed(4)} TON</p>
          <button type="button" onClick={handleTonPay} disabled={tonPayLoading || Number(tonAmount) <= 0} className="mt-3 w-full rounded-lg bg-[#229ED9] py-3 font-bold text-white disabled:opacity-50">{tonPayLoading ? "Creating order..." : "Pay with Gram (TON)"}</button>
        </div>
      )}

      {telegramMode && activeTab === "deposit" && depositMethod === "telegram" && (
        <div className="mb-4 rounded-xl border border-[#229ED9]/50 bg-[#229ED9]/10 p-4">
          <h3 className="mb-3 font-bold text-white">{t("wallet.telegram_stars")}</h3>
          <p className="mb-3 text-xs text-[#8ed8ff]">{t("wallet.telegram_deposit_amount")}</p>
          <div className="grid grid-cols-3 gap-2">{[5,20,50,100,200,500].map(v => <button key={v} type="button" className={`rounded-lg border py-3 text-white ${telegramAmount === String(v) ? "border-[#229ED9] bg-[#229ED9]/30" : "border-white/20"}`} onClick={() => setTelegramAmount(String(v))}>{v}U</button>)}</div>
          <input value={telegramAmount} onChange={(event) => setTelegramAmount(event.target.value)} className="mt-3 w-full rounded-lg bg-black/30 p-3 text-white" type="number" min="1" placeholder={t("wallet.telegram_custom_amount")} />
          <button type="button" onClick={handleTelegramPay} disabled={telegramPayLoading || !telegramAmount || Number(telegramAmount) <= 0} className="mt-3 w-full rounded-lg bg-[#229ED9] py-3 font-bold text-white disabled:opacity-50">{telegramPayLoading ? t("wallet.ton_submitting") : t("wallet.telegram_pay")}</button>
        </div>
      )}

      {activeTab === "deposit" && renderDepositContent()}
      {activeTab === "withdraw" && renderWithdrawContent()}
      {activeTab === "records" && renderRecordsContent()}

      <TokenPayModal isOpen={Boolean(usdtOrder)} onClose={() => setUsdtOrder(null)} payAddress={usdtOrder?.pay_address || ""} payAmount={usdtOrder?.pay_amount || ""} chainType="TRX" expiredAt={usdtOrder?.expired_at || 0} />

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
