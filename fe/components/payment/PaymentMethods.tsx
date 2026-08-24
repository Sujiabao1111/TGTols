"use client"

import { useEffect, useMemo, useState } from "react"
import { paymentService, PaymentMethod } from "@/services/payment"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { Loader2 } from "lucide-react"
import type { PaymentRegion } from "./paymentRegion"
import { matchesDepositRegion } from "./paymentRegion"

interface PaymentMethodsProps {
  region: PaymentRegion
  strictRegionOnly?: boolean
  selectedMethod: PaymentMethod | null
  onSelect: (method: PaymentMethod) => void
}

export function PaymentMethods({ region, strictRegionOnly = false, selectedMethod, onSelect }: PaymentMethodsProps) {
  const { t } = useLanguage()
  const [methods, setMethods] = useState<PaymentMethod[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const sortMethods = (data: PaymentMethod[]) =>
      [...data].sort((a, b) => {
        if (a.enabled === b.enabled) return 0
        return a.enabled ? -1 : 1
      })

    const loadPaymentMethods = async () => {
      try {
        const data = await paymentService.getPaymentMethods()
        const availableMethods = data.filter((method) => method.code !== "199VOUCHER")
        const initialRegionMethods = sortMethods(
          availableMethods.filter((method) => matchesDepositRegion(method, region, strictRegionOnly)),
        )

        setMethods(availableMethods)
        if (initialRegionMethods.length > 0) {
          onSelect(initialRegionMethods.find((method) => method.enabled) || initialRegionMethods[0])
        }
      } catch (error) {
        console.error("Failed to load payment methods:", error)
      } finally {
        setLoading(false)
      }
    }

    loadPaymentMethods()
  }, [onSelect, region, strictRegionOnly])

  const visibleMethods = useMemo(() => {
    const filteredMethods = methods.filter((method) => matchesDepositRegion(method, region, strictRegionOnly))

    return [...filteredMethods].sort((a, b) => {
      if (a.enabled === b.enabled) return 0
      return a.enabled ? -1 : 1
    })
  }, [methods, region, strictRegionOnly])

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="animate-spin text-lucky-gold" size={24} />
      </div>
    )
  }

  if (visibleMethods.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm text-gray-400">
        No payment methods available for this country yet.
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {visibleMethods.map((method) => {
        const isDisabled = !method.enabled
        const isSelected = selectedMethod?.code === method.code
        const isPhilippinesUSDT = method.code === "USDT" && region === "PH"
        const amountCurrency = isPhilippinesUSDT ? "PHP" : method.code === "USDT" ? "IDR" : method.currency || "IDR"
        const minAmount = isPhilippinesUSDT ? 500 : method.min_amount
        const maxAmount = isPhilippinesUSDT ? 57000 : method.max_amount

        return (
          <div
            key={method.code}
            onClick={() => !isDisabled && onSelect(method)}
            className={`p-4 rounded-xl border transition-all flex items-center justify-between ${
              isDisabled
                ? "bg-white/5 border-white/5 opacity-50 cursor-not-allowed"
                : isSelected
                  ? "bg-lucky-purple border-lucky-gold shadow-lucky-gold/10 shadow-lg cursor-pointer"
                  : "bg-white/5 border-white/10 hover:bg-white/10 cursor-pointer"
            }`}
          >
            <div className="flex items-center gap-4">
              <div
                className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold ${
                  isDisabled ? "bg-gray-600 text-gray-400" : "bg-white/10 text-lucky-gold"
                }`}
              >
                {method.name.slice(0, 2).toUpperCase()}
              </div>
              <div>
                <div className={`font-bold ${isDisabled ? "text-gray-500" : "text-white"}`}>
                  {method.name}
                  {isDisabled && (
                    <span className="ml-2 text-xs px-2 py-0.5 rounded-full bg-gray-700 text-gray-400">
                      {t("wallet.coming_soon_tag")}
                    </span>
                  )}
                </div>
                <div className="text-xs text-gray-500">
                  {minAmount.toLocaleString()} - {maxAmount.toLocaleString()} {amountCurrency}
                </div>
                {method.description && <div className="text-[11px] text-gray-500 mt-1">{method.description}</div>}
              </div>
            </div>
            {isSelected && !isDisabled && (
              <div className="w-5 h-5 rounded-full bg-lucky-gold flex items-center justify-center">
                <svg className="w-3 h-3 text-lucky-dark" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                </svg>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
