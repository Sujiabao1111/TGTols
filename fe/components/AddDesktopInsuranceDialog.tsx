"use client"

import { CheckCircle2, X } from "lucide-react"
import { useEffect, useState } from "react"
import { useLanguage } from "@/app/contexts/LanguageContext"

const ADD_DESKTOP_INSURANCE_DIALOG_EVENT = "add-desktop-insurance-dialog"

type InsuranceDialogPayload = {
  amount: string
}

export const showAddDesktopInsuranceDialog = (amount: string) => {
  if (typeof window === "undefined") {
    return
  }

  window.dispatchEvent(
    new CustomEvent<InsuranceDialogPayload>(ADD_DESKTOP_INSURANCE_DIALOG_EVENT, {
      detail: { amount },
    }),
  )
}

export function AddDesktopInsuranceDialog() {
  const { t } = useLanguage()
  const [amount, setAmount] = useState<string | null>(null)

  useEffect(() => {
    const handleDialog = (event: Event) => {
      const customEvent = event as CustomEvent<InsuranceDialogPayload>
      setAmount(customEvent.detail?.amount || "0.00U")
    }

    window.addEventListener(ADD_DESKTOP_INSURANCE_DIALOG_EVENT, handleDialog)
    return () => {
      window.removeEventListener(ADD_DESKTOP_INSURANCE_DIALOG_EVENT, handleDialog)
    }
  }, [])

  if (!amount) {
    return null
  }

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center bg-black/55 px-4">
      <div className="w-full max-w-sm rounded-lg border border-green-400/30 bg-[#0f1b14] p-5 text-white shadow-2xl">
        <div className="flex items-start gap-3">
          <CheckCircle2 className="mt-0.5 h-6 w-6 flex-none text-green-400" />
          <div className="min-w-0 flex-1">
            <p className="text-sm font-semibold leading-6 text-green-100">
              {t("activity.add_desktop_insurance_triggered").replace("{amount}", amount)}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setAmount(null)}
            className="flex h-8 w-8 flex-none items-center justify-center rounded-full border border-white/10 text-gray-300 transition-colors hover:bg-white/10 hover:text-white"
            aria-label={t("common.close")}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <button
          type="button"
          onClick={() => setAmount(null)}
          className="mt-5 w-full rounded-md bg-green-500 px-4 py-2 text-sm font-bold text-white transition-colors hover:bg-green-400"
        >
          {t("common.close")}
        </button>
      </div>
    </div>
  )
}
