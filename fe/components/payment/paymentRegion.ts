import type { PaymentMethod, WithdrawMethod } from "@/services/payment"

export type PaymentRegion = "ID" | "PH"

export const paymentRegionOptions: Array<{
  code: PaymentRegion
  label: string
}> = [
  { code: "ID", label: "Indonesia" },
  { code: "PH", label: "Philippines" },
]

const PH_PAYMENT_CODES = new Set(["GCASH", "GCASH_QR", "GCASH_APP"])
const USDT_PAYMENT_CODES = new Set(["USDT"])
const ID_DEPOSIT_CODES = new Set(["DANA", "USDT"])
const FORCED_PAYMENT_REGIONS: Record<string, PaymentRegion> = {
  "ppnetpp.net": "PH",
  "ppbetpp.tech": "PH",
  "ppnet11.com": "PH",
  "ppnet22.com": "PH",
  "ppnet33.com": "PH",
  "ppnet44.com": "PH",
  "ppnet55.com": "PH",
  "ppnetpp.com": "ID",
}

function normalizePaymentHostname(hostname: string) {
  const normalized = hostname.trim().toLowerCase().replace(/\.$/, "")
  const withoutPort = normalized.replace(/:\d+$/, "")
  return withoutPort.replace(/^www\./, "")
}

export function getForcedPaymentRegionByHostname(hostname: string): PaymentRegion | null {
  return FORCED_PAYMENT_REGIONS[normalizePaymentHostname(hostname)] ?? null
}

export function getCurrentForcedPaymentRegion(): PaymentRegion | null {
  if (typeof window === "undefined") {
    return null
  }

  return getForcedPaymentRegionByHostname(window.location.hostname)
}

export function matchesDepositRegion(method: PaymentMethod, region: PaymentRegion, strictRegionOnly = false) {
	if (method.code.toUpperCase() === "TON") return true
	// Telegram Stars is rendered in its own payment section, never inside
	// Indonesia/Philippines regional payment methods.
	if (method.code.toUpperCase() === "TG_STARS") {
		return false
	}
	if (method.code.toUpperCase() === "TG_STARS") {
		return true
	}
  // USDT is supported globally, including the region-locked Indonesia and Philippines sites.
  if (method.code.toUpperCase() === "USDT") {
    return true
  }

  if (strictRegionOnly) {
    return region === "ID" ? method.currency === "IDR" : method.currency === "PHP"
  }

  if (region === "ID") {
    return ID_DEPOSIT_CODES.has(method.code.toUpperCase()) && method.currency === "IDR"
  }

  return method.currency === "PHP"
}

export function getCurrencyByPaymentCode(code: string) {
  const normalizedCode = code.toUpperCase()
  if (USDT_PAYMENT_CODES.has(normalizedCode) || normalizedCode.startsWith("USDT")) {
    return "USDT"
  }
  if (PH_PAYMENT_CODES.has(normalizedCode)) {
    return "PHP"
  }
  return "IDR"
}

export function matchesWithdrawRegion(method: WithdrawMethod, region: PaymentRegion) {
  if (method.code.toUpperCase() === "TON" || method.currency === "TON") return true
  if (region === "ID") {
    return method.currency === "IDR"
  }

  return method.currency === "PHP"
}
