"use client"

import { useState } from "react"
import { PaymentMethods } from "./PaymentMethods"
import { VoucherModal } from "./VoucherModal"
import { TokenPayModal } from "./TokenPayModal"
import { ChainSelectorModal } from "./ChainSelectorModal"
import type { PaymentRegion } from "./paymentRegion"
import { PaymentMethod, paymentService, TokenPayCreateOrderResponse } from "@/services/payment"
import { useToast } from "@/hooks/use-toast"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { useUser } from "@/app/contexts/UserContext"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { useNewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"
import { ChevronRight, Loader2 } from "lucide-react"
import { useRouter } from "next/navigation"

interface DepositFormProps {
  selectedRegion: PaymentRegion
  strictRegionOnly?: boolean
  onSuccess?: () => void
}

const QUICK_AMOUNTS_IDR = [50000, 100000, 200000, 500000, 1000000]
const QUICK_AMOUNTS_USDT_IDR = [10, 20, 50, 100, 200, 500]
const QUICK_AMOUNTS_PHP = [100, 300, 500, 1000, 3000]
const QUICK_AMOUNTS_USDT_PHP = [10, 20, 50, 100, 200, 500]

export function DepositForm({ selectedRegion, strictRegionOnly = false, onSuccess }: DepositFormProps) {
  const { t } = useLanguage()
  const router = useRouter()
  const { updateBalance } = useUser()
  const { getRate } = useExchangeRate()
  const { status: newUserRechargeStatus } = useNewUserRechargeStatus(true)
  const [amount, setAmount] = useState<string>("")
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod | null>(null)
  const [loading, setLoading] = useState(false)
  const [showVoucherModal, setShowVoucherModal] = useState(false)
  const [showChainSelector, setShowChainSelector] = useState(false)
  const [showTokenPayModal, setShowTokenPayModal] = useState(false)
  const [tokenPayData, setTokenPayData] = useState<TokenPayCreateOrderResponse | null>(null)
  const [selectedChain, setSelectedChain] = useState<"ETH" | "TRX">("ETH")
  const { toast } = useToast()

  const isUSDT = selectedMethod?.code === "USDT"
  const isTON = selectedMethod?.code === "TON"
  const isTelegramStars = selectedMethod?.code === "TG_STARS"
  const isPHP = selectedMethod?.currency === "PHP" || (isUSDT && selectedRegion === "PH")
  const quickAmounts = isUSDT
    ? isPHP
      ? QUICK_AMOUNTS_USDT_PHP
      : QUICK_AMOUNTS_USDT_IDR
    : isPHP
      ? QUICK_AMOUNTS_PHP
      : QUICK_AMOUNTS_IDR
  const amountPrefix = isTON ? "USD" : isPHP ? "PHP" : "Rp"
  const amountLabel = isTON ? "Deposit Amount (USD)" : isPHP ? "Deposit Amount (PHP)" : t("wallet.deposit_amount")
  const amountCurrency = isTON ? "USD" : isPHP ? "PHP" : "IDR"
  const usdRate = getRate("USD") || 0.000061
  const phpRate = getRate("PHP") || 0.0035
  const usdtLocalRate = isPHP ? phpRate / usdRate : 1 / usdRate
  const selectedMinAmount = isUSDT
    ? isPHP
      ? Math.max(500, Math.ceil(5 * usdtLocalRate))
      : Math.ceil(5 * usdtLocalRate)
    : selectedMethod?.min_amount || 0
  const selectedMaxAmount = isUSDT ? Math.floor(1000 * usdtLocalRate) : selectedMethod?.max_amount || 0
  const newUserRechargeTotal = newUserRechargeStatus?.tiers.length || 4
  const newUserRechargeProgress = Math.min(newUserRechargeStatus?.progress_count || 0, newUserRechargeTotal)
  const amountInU = amount && parseFloat(amount) > 0
    ? isPHP
      ? parseFloat(amount) / usdtLocalRate
      : parseFloat(amount) * usdRate
    : 0

  const handleVoucherSubmit = async (voucherCode: string) => {
    const numAmount = parseFloat(amount)
    try {
      const response = await paymentService.redeemVoucher({
        amount: numAmount,
        voucher_code: voucherCode,
        dst_code: "199VOUCHER",
      })

      toast({
        title: "Voucher Redeemed",
        description: `Credited $${response.usd_amount.toFixed(2)} USD`,
      })

      updateBalance(response.balance)
      setShowVoucherModal(false)
      onSuccess?.()
    } catch (error: unknown) {
      toast({
        title: "Redeem Failed",
        description: error instanceof Error ? error.message : t("common.error"),
        variant: "destructive",
      })
    }
  }

  const handleSubmit = async () => {
    if (!selectedMethod) {
      toast({
        title: t("wallet.select_method_first"),
        variant: "destructive",
      })
      return
    }

    if (!selectedMethod.enabled) {
      toast({
        title: t("wallet.method_unavailable"),
        description: t("wallet.coming_soon"),
        variant: "destructive",
      })
      return
    }

    if (!amount || parseFloat(amount) <= 0) {
      toast({
        title: t("wallet.enter_amount"),
        variant: "destructive",
      })
      return
    }

    const numAmount = parseFloat(amount)

    if (numAmount < selectedMinAmount) {
      toast({
        title: `${selectedMinAmount.toLocaleString()} ${amountCurrency} min`,
        variant: "destructive",
      })
      return
    }

    if (numAmount > selectedMaxAmount) {
      toast({
        title: `${selectedMaxAmount.toLocaleString()} ${amountCurrency} max`,
        variant: "destructive",
      })
      return
    }

    if (selectedMethod.code === "199VOUCHER") {
      setShowVoucherModal(true)
      return
    }

    if (selectedMethod.code === "USDT") {
      setShowChainSelector(true)
      return
    }

    if (selectedMethod.code === "TON") {
      toast({ title: "Use the TON wallet panel", description: "TON payments must be sent through TonConnect." })
      return
      /* Legacy deep-link flow intentionally disabled.
      setLoading(true)
      try {
        toast({ title: "TON order created", description: `${invoice.amount_ton.toFixed(4)} TON · ${invoice.comment}` })
        onSuccess?.()
      } catch (error: unknown) {
        toast({ title: "TON order failed", description: error instanceof Error ? error.message : t("common.error"), variant: "destructive" })
      } finally { setLoading(false) }
      */
      return
    }

    if (selectedMethod.code === "TG_STARS") {
      setLoading(true)
      try {
        const initData = (window as any).Telegram?.WebApp?.initData || ""
        const response = await paymentService.createTelegramStarsOrder(numAmount, initData)
        if (response.invoice_url) window.location.href = response.invoice_url
        toast({ title: "Order Created", description: "Please complete payment in Telegram" })
        onSuccess?.()
      } catch (error: unknown) {
        toast({ title: "Payment failed", description: error instanceof Error ? error.message : t("common.error"), variant: "destructive" })
      } finally { setLoading(false) }
      return
    }

    setLoading(true)

    try {
      const currentUrl = window.location.origin + "/wallet"

      const response = await paymentService.createPaymentOrder({
        amount: numAmount,
        type: "channel",
        dst_code: selectedMethod.code,
        callback_url: currentUrl,
        currency: selectedMethod.currency,
        channel: selectedMethod.channel,
      })

      toast({
        title: t("wallet.order_success"),
        description: t("wallet.redirecting"),
      })

      if (response.pay_url) {
        window.location.href = response.pay_url
      }

      onSuccess?.()
    } catch (error: unknown) {
      toast({
        title: t("wallet.order_failed"),
        description: error instanceof Error ? error.message : t("common.error"),
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleChainSelect = async (chain: "ETH" | "TRX") => {
    setSelectedChain(chain)
    setShowChainSelector(false)

    const numAmount = parseFloat(amount)
    setLoading(true)

    try {
      const currentUrl = window.location.origin + "/wallet"

      const response = await paymentService.createTokenPayOrder({
        amount: numAmount,
        type: "channel",
        dst_code: "USDT",
        currency: amountCurrency,
        callback_url: currentUrl,
        chain_type: chain,
      })

      setTokenPayData(response)
      setShowTokenPayModal(true)

      toast({
        title: "Order Created",
        description: "Please complete the payment",
      })
    } catch (error: unknown) {
      toast({
        title: "Create Order Failed",
        description: error instanceof Error ? error.message : t("common.error"),
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm text-gray-400 mb-3 pl-1">{t("wallet.select_payment_method")}</h3>
        <PaymentMethods
          region={selectedRegion}
          strictRegionOnly={strictRegionOnly}
          selectedMethod={selectedMethod}
          onSelect={setSelectedMethod}
        />
        {isTelegramStars && (
          <div className="mt-3 rounded-xl border border-[#229ED9]/40 bg-[#229ED9]/10 p-4 text-sm text-[#8ed8ff]">
            <div className="font-semibold text-white">Telegram Stars</div>
            <div className="mt-1">将在 Telegram 内打开独立支付界面。当前 Bot 尚未配置，功能暂不可用。</div>
          </div>
        )}
      </div>

      <div>
        <h3 className="text-sm text-gray-400 mb-1 pl-1">{amountLabel}</h3>
        {newUserRechargeStatus && !newUserRechargeStatus.hidden && (
          <button
            type="button"
            onClick={() => router.push("/newUserRecharge")}
            className="mb-3 flex w-full items-center gap-1 pl-1 text-left text-xs font-medium text-lucky-gold transition-colors hover:text-yellow-300"
          >
            <span>{t("wallet.new_user_recharge_hint")} ({newUserRechargeProgress}/{newUserRechargeTotal})</span>
            <ChevronRight className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
          </button>
        )}
        <div className="grid grid-cols-3 gap-3 mb-4">
          {quickAmounts.map((amt) => {
            const quickAmountValue = isUSDT ? Math.ceil(amt * usdtLocalRate) : amt
            return (
            <button
              key={amt}
              onClick={() => setAmount(quickAmountValue.toString())}
              className={`py-3 rounded-xl font-bold border transition-all ${
                amount === quickAmountValue.toString()
                  ? "bg-lucky-gold border-lucky-gold text-lucky-dark"
                  : "bg-transparent border-white/20 text-white hover:border-lucky-gold"
              }`}
            >
              {isUSDT ? `${amt}U` : isPHP ? amt.toLocaleString() : `${(amt / 1000).toFixed(0)}K`}
            </button>
            )
          })}
        </div>

        <div className="relative">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">{amountPrefix}</span>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder={t("wallet.input_amount")}
            className="w-full bg-lucky-dark border border-white/20 rounded-xl pl-16 pr-4 py-4 text-white focus:border-lucky-gold focus:outline-none text-lg"
          />
        </div>

        {selectedMethod && (
          <p className="text-xs text-gray-500 mt-2">
            {selectedMinAmount.toLocaleString()} - {selectedMaxAmount.toLocaleString()} {amountCurrency}
          </p>
        )}

        {amountInU > 0 && (
          <p className="mt-2 text-xs text-lucky-gold">≈ {amountInU.toFixed(2)} U</p>
        )}
      </div>

      <button
        onClick={handleSubmit}
        disabled={loading || !amount || !selectedMethod || !selectedMethod.enabled}
        className="w-full py-4 rounded-full bg-gradient-to-r from-lucky-gold to-orange-500 text-lucky-dark font-bold text-lg shadow-xl shadow-lucky-gold/20 hover:scale-[1.02] transition-transform disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {loading ? (
          <>
            <Loader2 className="animate-spin" size={20} />
            {t("wallet.btn_processing")}
          </>
        ) : !selectedMethod?.enabled ? (
          <>{t("wallet.coming_soon_tag")}</>
        ) : (
          <>{t("wallet.deposit_now")}</>
        )}
      </button>

      {selectedMethod?.code === "199VOUCHER" && (
        <VoucherModal
          isOpen={showVoucherModal}
          onClose={() => setShowVoucherModal(false)}
          onSubmit={handleVoucherSubmit}
          amount={parseFloat(amount) || 0}
        />
      )}

      <ChainSelectorModal
        isOpen={showChainSelector}
        onClose={() => setShowChainSelector(false)}
        onSelect={handleChainSelect}
        amount={parseFloat(amount) || 0}
        localRate={usdtLocalRate}
      />

      {tokenPayData && (
        <TokenPayModal
          isOpen={showTokenPayModal}
          onClose={() => {
            setShowTokenPayModal(false)
            onSuccess?.()
          }}
          payAddress={tokenPayData.pay_address}
          payAmount={tokenPayData.pay_amount}
          chainType={selectedChain}
          expiredAt={tokenPayData.expired_at}
        />
      )}
    </div>
  )
}
