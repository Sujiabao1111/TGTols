"use client"

import { useEffect, useMemo, useState } from "react"
import { Loader2, Wallet } from "lucide-react"
import { paymentService, WithdrawMethod } from "@/services/payment"
import { useToast } from "@/hooks/use-toast"
import { Portal } from "@/components/ui/portal"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { useLanguage } from "@/app/contexts/LanguageContext"
import type { PaymentRegion } from "./paymentRegion"
import { matchesWithdrawRegion } from "./paymentRegion"

interface WithdrawFormProps {
  selectedRegion: PaymentRegion
  totalBalanceUSD: number
  withdrawableBalanceUSD: number
  totalDepositUSD: number
  remainingWager?: number
  depositWagerMultiplier?: number
  rewardWagerMultiplier?: number
  onSuccess?: () => void
}

const MIN_WITHDRAW_USD = 10
const QUICK_AMOUNTS_IDR = [50000, 100000, 200000, 500000, 1000000]
const QUICK_AMOUNTS_PHP = [100, 300, 500, 1000, 3000]

interface WithdrawErrorNotice {
  title: string
  description: string
  useDialog: boolean
}

function formatWithdrawAmount(amount: number, currency: string) {
  if (currency === "PHP") {
    return `PHP ${amount.toLocaleString(undefined, {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    })}`
  }

  return `Rp ${amount.toLocaleString(undefined, {
    maximumFractionDigits: 0,
  })}`
}

function buildWithdrawRangeNotice(
  amount: number,
  minAmount: number,
  maxAmount: number,
  currency: string,
  labels: {
    tooLowTitle: string
    tooHighTitle: string
    minDescription: (value: string) => string
    maxDescription: (value: string) => string
  },
): WithdrawErrorNotice {
  const formattedMin = formatWithdrawAmount(minAmount, currency)
  const formattedMax = formatWithdrawAmount(maxAmount, currency)

  if (amount < minAmount) {
    return {
      title: labels.tooLowTitle,
      description: labels.minDescription(formattedMin),
      useDialog: true,
    }
  }

  return {
    title: labels.tooHighTitle,
    description: labels.maxDescription(formattedMax),
    useDialog: true,
  }
}

function buildInsufficientBalanceNotice(
  availableAmount: number,
  currency: string,
  labels: {
    title: string
    description: (value: string) => string
  },
): WithdrawErrorNotice {
  return {
    title: labels.title,
    description: labels.description(formatWithdrawAmount(availableAmount, currency)),
    useDialog: true,
  }
}

function sortMethods(methods: WithdrawMethod[]) {
  return [...methods].sort((a, b) => a.name.localeCompare(b.name))
}

function formatMultiplier(value: number) {
  return Number.isInteger(value) ? value.toFixed(0) : value.toFixed(2)
}

function getWithdrawErrorNotice(
  error: unknown,
  labels: {
    failedTitle: string
    fallback: string
    unavailableTitle: string
    lockedDescription: (remaining: string) => string
    insufficientTitle: string
    insufficientDescription: string
    depositRequiredTitle: string
    depositRequiredDescription: (amount: string) => string
    dailyLimitTitle: string
    dailyLimitDescription: string
  },
): WithdrawErrorNotice {
  const message = error instanceof Error ? error.message : ""

  if (!message) {
    return {
      title: labels.failedTitle,
      description: labels.fallback,
      useDialog: false,
    }
  }

  const wagerMatch = message.match(/reward funds still require\s+(\d+(?:\.\d+)?)\s+wager turnover/i)
  if (wagerMatch) {
    const remaining = Number(wagerMatch[1]).toFixed(2)
    return {
      title: labels.unavailableTitle,
      description: labels.lockedDescription(remaining),
      useDialog: true,
    }
  }

  if (/insufficient balance/i.test(message)) {
    return {
      title: labels.insufficientTitle,
      description: labels.insufficientDescription,
      useDialog: true,
    }
  }

  const depositMatch = message.match(/withdraw requires matching deposit\s+(\d+(?:\.\d+)?)\s+USDT/i)
  if (depositMatch) {
    return {
      title: labels.depositRequiredTitle,
      description: labels.depositRequiredDescription(`USDT ${Number(depositMatch[1]).toFixed(2)}`),
      useDialog: true,
    }
  }

  if (/daily withdraw limit reached/i.test(message)) {
    return {
      title: labels.dailyLimitTitle,
      description: labels.dailyLimitDescription,
      useDialog: true,
    }
  }

  return {
    title: labels.failedTitle,
    description: message,
    useDialog: false,
  }
}

export function WithdrawForm({
  selectedRegion,
  totalBalanceUSD,
  withdrawableBalanceUSD,
  totalDepositUSD,
  remainingWager = 0,
  depositWagerMultiplier = 2,
  rewardWagerMultiplier = 20,
  onSuccess,
}: WithdrawFormProps) {
  const { toast } = useToast()
  const { t } = useLanguage()
  const { convertFromUSDTo, formatUSD, getRate } = useExchangeRate()
  const [methods, setMethods] = useState<WithdrawMethod[]>([])
  const [selectedMethod, setSelectedMethod] = useState<WithdrawMethod | null>(null)
  const [amount, setAmount] = useState("")
  const [account, setAccount] = useState("")
  const [accountName, setAccountName] = useState("")
  const [phone, setPhone] = useState("")
  const [email, setEmail] = useState("")
  const [loading, setLoading] = useState(false)
  const [loadingMethods, setLoadingMethods] = useState(true)
  const [errorNotice, setErrorNotice] = useState<WithdrawErrorNotice | null>(null)

  useEffect(() => {
    const loadMethods = async () => {
      try {
        const data = await paymentService.getWithdrawMethods()
        const enabledMethods = sortMethods(data.filter((method) => method.enabled))
        const initialRegionMethods = enabledMethods.filter((method) => matchesWithdrawRegion(method, selectedRegion))

        setMethods(enabledMethods)
        setSelectedMethod(initialRegionMethods[0] || null)
      } catch (error) {
        console.error("Failed to load withdraw methods:", error)
      } finally {
        setLoadingMethods(false)
      }
    }

    void loadMethods()
  }, [selectedRegion])

  const visibleMethods = useMemo(
    () => sortMethods(methods.filter((method) => matchesWithdrawRegion(method, selectedRegion))),
    [methods, selectedRegion],
  )

  const activeCurrency = selectedMethod?.currency || (selectedRegion === "ID" ? "IDR" : "PHP")
  const quickAmounts = activeCurrency === "PHP" ? QUICK_AMOUNTS_PHP : QUICK_AMOUNTS_IDR
  const effectiveWithdrawableBalanceUSD = withdrawableBalanceUSD
  const lockedBonusUSD = Math.max(totalBalanceUSD - withdrawableBalanceUSD, 0)
  const withdrawableAmountInCurrency = convertFromUSDTo(effectiveWithdrawableBalanceUSD, activeCurrency)
  const targetRate = getRate(activeCurrency)
  const usdRate = getRate("USD")
  const amountLabel = activeCurrency === "PHP" ? t("wallet.withdraw_amount_php") : t("wallet.withdraw_amount_idr")
  const lockedBonusLabel = formatUSD(lockedBonusUSD)
  const accountLabel = selectedMethod?.type === "ewallet" ? t("wallet.wallet_account") : t("wallet.bank_account_number")
  const accountPlaceholder =
    selectedMethod?.type === "ewallet" ? t("wallet.enter_wallet_account") : t("wallet.enter_bank_account_number")

  const handleSubmit = async () => {
    if (!selectedMethod) {
      toast({
        title: t("wallet.select_withdraw_method"),
        variant: "destructive",
      })
      return
    }

    const numericAmount = parseFloat(amount)
    if (!numericAmount || numericAmount <= 0) {
      toast({
        title: t("wallet.enter_amount"),
        variant: "destructive",
      })
      return
    }

    const withdrawCurrency = selectedMethod.currency || activeCurrency
    const requiredBalanceUSD = targetRate > 0 && usdRate > 0 ? (numericAmount / targetRate) * usdRate : 0
    const minimumWithdrawAmount = targetRate > 0 && usdRate > 0 ? (MIN_WITHDRAW_USD / usdRate) * targetRate : selectedMethod.min_amount
    const effectiveMinAmount = Math.max(selectedMethod.min_amount, minimumWithdrawAmount)

    if (numericAmount < effectiveMinAmount || numericAmount > selectedMethod.max_amount) {
      setErrorNotice(
        buildWithdrawRangeNotice(
          numericAmount,
          effectiveMinAmount,
          selectedMethod.max_amount,
          withdrawCurrency,
          {
            tooLowTitle: t("wallet.amount_too_low"),
            tooHighTitle: t("wallet.amount_too_high"),
            minDescription: (value) => t("wallet.min_withdraw_amount").replace("{amount}", value),
            maxDescription: (value) => t("wallet.max_withdraw_amount").replace("{amount}", value),
          },
        ),
      )
      return
    }

    if (totalDepositUSD <= 0) {
      setErrorNotice({
        title: t("wallet.withdraw_deposit_required_title"),
        description: t("wallet.withdraw_deposit_required_desc").replace(
          "{amount}",
          formatWithdrawAmount(numericAmount, withdrawCurrency),
        ),
        useDialog: true,
      })
      return
    }

    if (numericAmount > withdrawableAmountInCurrency) {
      setErrorNotice(
        buildInsufficientBalanceNotice(withdrawableAmountInCurrency, activeCurrency, {
          title: t("wallet.insufficient"),
          description: (value) => t("wallet.available_balance_hint").replace("{amount}", value),
        }),
      )
      return
    }

    if (!account.trim() || !accountName.trim()) {
      toast({
        title: t("wallet.complete_account_information"),
        variant: "destructive",
      })
      return
    }

    if (!phone.trim() || !email.trim()) {
      toast({
        title: t("wallet.complete_contact_information"),
        variant: "destructive",
      })
      return
    }

    if (requiredBalanceUSD > effectiveWithdrawableBalanceUSD) {
      setErrorNotice(
        buildInsufficientBalanceNotice(withdrawableAmountInCurrency, activeCurrency, {
          title: t("wallet.insufficient"),
          description: (value) => t("wallet.available_balance_hint").replace("{amount}", value),
        }),
      )
      return
    }

    setLoading(true)
    try {
      await paymentService.createWithdrawOrder({
        amount: numericAmount,
        type: selectedMethod.type as "bankcard" | "ewallet",
        dst_code: selectedMethod.code,
        account,
        account_name: accountName,
        phone,
        email,
        currency: selectedMethod.currency,
        channel: selectedMethod.channel,
        bank_code: selectedMethod.type === "bankcard" ? selectedMethod.code : undefined,
      })

      toast({
        title: t("wallet.withdrawal_submitted"),
        description: t("wallet.withdrawal_submitted_desc"),
      })

      setAmount("")
      setAccount("")
      setAccountName("")
      setPhone("")
      setEmail("")
      onSuccess?.()
    } catch (error: unknown) {
      const notice = getWithdrawErrorNotice(error, {
        failedTitle: t("wallet.withdrawal_failed"),
        fallback: t("wallet.withdrawal_failed_desc"),
        unavailableTitle: t("wallet.withdrawal_unavailable"),
        lockedDescription: (remaining) => t("wallet.locked_reward_balance_desc").replace("{amount}", remaining),
        insufficientTitle: t("wallet.insufficient"),
        insufficientDescription: t("wallet.insufficient_withdrawable_balance"),
        depositRequiredTitle: t("wallet.withdraw_deposit_required_title"),
        depositRequiredDescription: (depositAmount) =>
          t("wallet.withdraw_deposit_required_desc").replace("{amount}", depositAmount),
        dailyLimitTitle: t("wallet.withdraw_daily_limit_title"),
        dailyLimitDescription: t("wallet.withdraw_daily_limit_desc"),
      })
      if (notice.useDialog) {
        setErrorNotice(notice)
      } else {
        toast({
          title: notice.title,
          description: notice.description,
          variant: "destructive",
        })
      }
    } finally {
      setLoading(false)
    }
  }

  if (loadingMethods) {
    return (
      <div className="flex justify-center py-10">
        <Loader2 className="animate-spin text-lucky-gold" size={24} />
      </div>
    )
  }

  return (
    <>
      {errorNotice && (
        <Portal>
          <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/70 px-4 backdrop-blur-sm">
            <div className="w-full max-w-md rounded-2xl border border-red-500/30 bg-slate-950 p-6 text-white shadow-2xl">
              <h3 className="text-lg font-semibold">{errorNotice.title}</h3>
              <p className="mt-3 text-sm leading-6 text-gray-300">{errorNotice.description}</p>
              <button
                type="button"
                className="mt-6 w-full rounded-full bg-gradient-to-r from-lucky-gold to-orange-500 px-4 py-3 font-bold text-lucky-dark transition-opacity hover:opacity-90"
                onClick={() => setErrorNotice(null)}
              >
                {t("common.i_understand")}
              </button>
            </div>
          </div>
        </Portal>
      )}

      <div className="space-y-6">
        {lockedBonusUSD > 0 && (
          <div className="rounded-lg border border-amber-400/20 bg-amber-400/10 px-3 py-2 text-xs leading-5 text-amber-100">
            {t("wallet.locked_bonus")}: {lockedBonusLabel}.{" "}
            {t("wallet.locked_bonus_desc")
              .replace("{depositMultiplier}", formatMultiplier(depositWagerMultiplier))
              .replace("{rewardMultiplier}", formatMultiplier(rewardWagerMultiplier))}
            {remainingWager > 0 ? ` ${t("wallet.remaining_turnover")}: ${remainingWager.toFixed(2)}.` : ""}
          </div>
        )}

        <div>
          <h3 className="text-sm text-gray-400 mb-3 pl-1">{t("wallet.withdraw_method")}</h3>
          {visibleMethods.length === 0 ? (
            <div className="rounded-xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm text-gray-400">
              {t("wallet.no_withdraw_methods")}
            </div>
          ) : (
            <div className="space-y-3">
              {visibleMethods.map((method) => {
                const isSelected = selectedMethod?.code === method.code

                return (
                  <button
                    key={method.code}
                    type="button"
                    onClick={() => setSelectedMethod(method)}
                    className={`w-full text-left p-4 rounded-xl border transition-all ${
                      isSelected
                        ? "bg-lucky-purple border-lucky-gold shadow-lucky-gold/10 shadow-lg"
                        : "bg-white/5 border-white/10 hover:bg-white/10"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-white/10 text-lucky-gold flex items-center justify-center">
                          <Wallet size={18} />
                        </div>
                        <div>
                          <div className="font-bold text-white">{method.name}</div>
                          <div className="text-xs text-gray-400">
                            {method.min_amount.toLocaleString()} - {method.max_amount.toLocaleString()} {method.currency || activeCurrency}
                          </div>
                          {method.description && <div className="text-[11px] text-gray-500 mt-1">{method.description}</div>}
                        </div>
                      </div>
                      {isSelected && <div className="w-3 h-3 rounded-full bg-lucky-gold" />}
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </div>

        <div>
          <h3 className="text-sm text-gray-400 mb-3 pl-1">{amountLabel}</h3>
          <div className="grid grid-cols-3 gap-3 mb-4">
            {quickAmounts.map((amt) => (
              <button
                key={amt}
                type="button"
                onClick={() => setAmount(amt.toString())}
                className={`py-3 rounded-xl font-bold border transition-all ${
                  amount === amt.toString()
                    ? "bg-lucky-gold border-lucky-gold text-lucky-dark"
                    : "bg-transparent border-white/20 text-white hover:border-lucky-gold"
                }`}
              >
                {activeCurrency === "PHP" ? amt.toLocaleString() : `${(amt / 1000).toFixed(0)}K`}
              </button>
            ))}
          </div>

          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">
              {activeCurrency === "PHP" ? "PHP" : "Rp"}
            </span>
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder={t("wallet.enter_amount")}
              className={`w-full bg-lucky-dark border border-white/20 rounded-xl pr-4 py-4 text-white focus:border-lucky-gold focus:outline-none text-lg ${
                activeCurrency === "PHP" ? "pl-16" : "pl-14"
              }`}
            />
          </div>
        </div>

        <div className="space-y-4">
          <div>
            <h3 className="text-sm text-gray-400 mb-2 pl-1">{accountLabel}</h3>
            <input
              type="text"
              value={account}
              onChange={(e) => setAccount(e.target.value)}
              placeholder={accountPlaceholder}
              className="w-full bg-lucky-dark border border-white/20 rounded-xl px-4 py-4 text-white focus:border-lucky-gold focus:outline-none"
            />
          </div>

          <div>
            <h3 className="text-sm text-gray-400 mb-2 pl-1">{t("wallet.account_name")}</h3>
            <input
              type="text"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              placeholder={t("wallet.enter_account_name")}
              className="w-full bg-lucky-dark border border-white/20 rounded-xl px-4 py-4 text-white focus:border-lucky-gold focus:outline-none"
            />
          </div>

          <div>
            <h3 className="text-sm text-gray-400 mb-2 pl-1">{t("wallet.phone_number")}</h3>
            <input
              type="text"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder={t("wallet.enter_phone_number")}
              className="w-full bg-lucky-dark border border-white/20 rounded-xl px-4 py-4 text-white focus:border-lucky-gold focus:outline-none"
            />
          </div>

          <div>
            <h3 className="text-sm text-gray-400 mb-2 pl-1">{t("wallet.email")}</h3>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t("wallet.enter_email")}
              className="w-full bg-lucky-dark border border-white/20 rounded-xl px-4 py-4 text-white focus:border-lucky-gold focus:outline-none"
            />
          </div>
        </div>

        <button
          type="button"
          onClick={handleSubmit}
          disabled={loading || !selectedMethod || visibleMethods.length === 0}
          className="w-full py-4 rounded-full bg-gradient-to-r from-lucky-gold to-orange-500 text-lucky-dark font-bold text-lg shadow-xl shadow-lucky-gold/20 hover:scale-[1.02] transition-transform disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {loading ? (
            <>
              <Loader2 className="animate-spin" size={20} />
              {t("wallet.btn_processing")}
            </>
          ) : (
            <>{t("wallet.withdraw_now")}</>
          )}
        </button>
      </div>
    </>
  )
}
