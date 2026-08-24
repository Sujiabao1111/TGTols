"use client"

import { useState, useEffect, useCallback } from "react"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { apiUrl } from "@/lib/api-base-url"

const localeCurrencyMap: Record<string, string> = {
  en: "PHP",
  cn: "CNY",
  ru: "RUB",
  es: "EUR",
  id: "IDR",
  ph: "PHP",
}

const currencyConfig: Record<
  string,
  {
    code: string
    symbol: string
    locale: string
    decimals: number
  }
> = {
  USD: { code: "USD", symbol: "$", locale: "en-US", decimals: 2 },
  CNY: { code: "CNY", symbol: "CNY", locale: "zh-CN", decimals: 2 },
  RUB: { code: "RUB", symbol: "RUB", locale: "ru-RU", decimals: 2 },
  EUR: { code: "EUR", symbol: "EUR", locale: "es-ES", decimals: 2 },
  IDR: { code: "IDR", symbol: "Rp", locale: "id-ID", decimals: 0 },
  PHP: { code: "PHP", symbol: "PHP", locale: "en-PH", decimals: 2 },
}

const defaultRates: Record<string, number> = {
  IDR: 1,
  USD: 0.000061,
  CNY: 0.00044,
  RUB: 0.0052,
  EUR: 0.000057,
  PHP: 0.0035,
}

interface ExchangeRateData {
  base: string
  rates: Record<string, number>
}

export function useExchangeRate() {
  const { language } = useLanguage()
  const [rates, setRates] = useState<Record<string, number>>(defaultRates)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const currencyCode = localeCurrencyMap[language] || "PHP"
  const config = currencyConfig[currencyCode] || currencyConfig.PHP

  const convertFromIDRTo = useCallback(
    (amountIDR: number, targetCurrency: string): number => {
      const rate = rates[targetCurrency]
      if (!rate) return amountIDR
      return amountIDR * rate
    },
    [rates],
  )

  const convertFromUSDTo = useCallback(
    (amountUSD: number, targetCurrency: string): number => {
      const usdRate = rates.USD
      const targetRate = rates[targetCurrency]
      if (!usdRate || !targetRate) return amountUSD

      const idrAmount = amountUSD / usdRate
      return idrAmount * targetRate
    },
    [rates],
  )

  const fetchRates = useCallback(async () => {
    try {
      setLoading(true)
      const res = await fetch(apiUrl("/exchange-rates"))
      if (!res.ok) {
        throw new Error(`Failed to fetch exchange rates: ${res.status}`)
      }

      const data: ExchangeRateData = await res.json()
      setRates({ ...defaultRates, ...data.rates })
      setError(null)
    } catch (err) {
      console.error("[useExchangeRate] Fetch failed:", err)
      setError("Failed to load exchange rates")
      setRates(defaultRates)
    } finally {
      setLoading(false)
    }
  }, [])

  const convertFromIDR = useCallback(
    (amountIDR: number): number => {
      return convertFromIDRTo(amountIDR, currencyCode)
    },
    [convertFromIDRTo, currencyCode],
  )

  const convertFromUSD = useCallback(
    (amountUSD: number): number => {
      return convertFromUSDTo(amountUSD, currencyCode)
    },
    [convertFromUSDTo, currencyCode],
  )

  const getUSToIDRRate = useCallback((): number => {
    const usdRate = rates.USD || defaultRates.USD
    if (!usdRate) return 17450
    return Math.round(1 / usdRate)
  }, [rates])

  const convertUSToIDR = useCallback(
    (amountU: number): number => {
      return amountU * getUSToIDRRate()
    },
    [getUSToIDRRate],
  )

  const formatUSToIDR = useCallback(
    (amountU = 1): string => {
      const idrAmount = convertUSToIDR(amountU)
      return `${idrAmount}${currencyConfig.IDR.symbol} (${amountU}U)`
    },
    [convertUSToIDR],
  )

  const formatCurrency = useCallback(
    (amount: number, currency?: string): string => {
      const targetCurrency = currency || currencyCode
      const targetConfig = currencyConfig[targetCurrency]

      if (!targetConfig) {
        return amount.toFixed(2)
      }

      if (targetCurrency === "IDR") {
        return `${targetConfig.symbol}${Math.round(amount).toLocaleString(targetConfig.locale)}`
      }

      return new Intl.NumberFormat(targetConfig.locale, {
        style: "currency",
        currency: targetCurrency,
        minimumFractionDigits: targetConfig.decimals,
        maximumFractionDigits: targetConfig.decimals,
      }).format(amount)
    },
    [currencyCode],
  )

  const formatUSD = useCallback(
    (amount: number): string => {
      return formatCurrency(amount, "USD")
    },
    [formatCurrency],
  )

  const formatFromIDR = useCallback(
    (amountIDR: number): string => {
      return formatCurrency(convertFromIDR(amountIDR))
    },
    [convertFromIDR, formatCurrency],
  )

  const formatFromIDRTo = useCallback(
    (amountIDR: number, targetCurrency: string): string => {
      return formatCurrency(convertFromIDRTo(amountIDR, targetCurrency), targetCurrency)
    },
    [convertFromIDRTo, formatCurrency],
  )

  const formatFromUSD = useCallback(
    (amountUSD: number): string => {
      return formatCurrency(convertFromUSD(amountUSD))
    },
    [convertFromUSD, formatCurrency],
  )

  const formatFromUSDTo = useCallback(
    (amountUSD: number, targetCurrency: string): string => {
      return formatCurrency(convertFromUSDTo(amountUSD, targetCurrency), targetCurrency)
    },
    [convertFromUSDTo, formatCurrency],
  )

  const getRate = useCallback(
    (currency: string): number => {
      return rates[currency] || 0
    },
    [rates],
  )

  useEffect(() => {
    let cancelled = false

    const run = async () => {
      try {
        const res = await fetch(apiUrl("/exchange-rates"))
        if (!res.ok) {
          throw new Error(`Failed to fetch exchange rates: ${res.status}`)
        }

        const data: ExchangeRateData = await res.json()
        if (!cancelled) {
          setRates({ ...defaultRates, ...data.rates })
          setError(null)
        }
      } catch (err) {
        console.error("[useExchangeRate] Fetch failed:", err)
        if (!cancelled) {
          setError("Failed to load exchange rates")
          setRates(defaultRates)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void run()

    return () => {
      cancelled = true
    }
  }, [])

  return {
    rates,
    currencyCode,
    currencySymbol: config.symbol,
    loading,
    error,
    convertFromIDRTo,
    convertFromIDR,
    convertFromUSDTo,
    convertFromUSD,
    convertUSToIDR,
    getUSToIDRRate,
    formatUSToIDR,
    formatCurrency,
    formatUSD,
    formatFromIDRTo,
    formatFromIDR,
    formatFromUSDTo,
    formatFromUSD,
    getRate,
    refresh: fetchRates,
  }
}

export function useBalanceWithCurrency(balanceIDR: number) {
  const { convertFromIDR, formatFromIDR, currencyCode, currencySymbol, loading } = useExchangeRate()

  const formatted = formatFromIDR(balanceIDR)
  const numericValue = convertFromIDR(balanceIDR)

  return {
    idrBalance: balanceIDR,
    displayBalance: numericValue,
    formattedBalance: formatted,
    currencyCode,
    currencySymbol,
    loading,
  }
}

export function useDepositConversion(amountIDR: number) {
  const { convertFromIDR, formatCurrency, currencyCode, getRate, loading } = useExchangeRate()

  const convertedAmount = convertFromIDR(amountIDR)
  const rate = getRate(currencyCode)
  const formatted = amountIDR > 0 ? formatCurrency(convertedAmount) : ""

  return {
    idrAmount: amountIDR,
    convertedAmount,
    formatted,
    rate,
    currencyCode,
    loading,
  }
}
