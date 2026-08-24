"use client"

import { paymentRegionOptions, type PaymentRegion } from "./paymentRegion"

interface PaymentRegionSelectorProps {
  selectedRegion: PaymentRegion
  allowedRegions?: PaymentRegion[]
  onChange: (region: PaymentRegion) => void
}

export function PaymentRegionSelector({ selectedRegion, allowedRegions, onChange }: PaymentRegionSelectorProps) {
  const visibleRegions = allowedRegions
    ? paymentRegionOptions.filter((region) => allowedRegions.includes(region.code))
    : paymentRegionOptions
  const isSingleRegion = visibleRegions.length === 1

  return (
    <div className="mb-6 rounded-2xl border border-lucky-gold/25 bg-gradient-to-r from-[#171717] via-[#121212] to-[#1a1a1a] p-2 shadow-[0_12px_30px_rgba(0,0,0,0.28)]">
      <div className={`grid gap-2 ${isSingleRegion ? "grid-cols-1" : "grid-cols-2"}`}>
        {visibleRegions.map((region) => {
          const isActive = selectedRegion === region.code

          return (
            <button
              key={region.code}
              type="button"
              disabled={isSingleRegion}
              onClick={() => onChange(region.code)}
              className={`rounded-xl border px-4 py-3 text-left transition-all ${
                isActive
                  ? "border-lucky-gold bg-lucky-gold text-lucky-dark shadow-[0_10px_24px_rgba(255,193,7,0.24)]"
                  : "border-white/10 bg-white/[0.03] text-white hover:border-lucky-gold/40 hover:bg-white/[0.06]"
              } ${isSingleRegion ? "cursor-default" : ""}`}
            >
              <div className="flex items-center gap-3">
                <span
                  className={`inline-flex h-9 w-9 items-center justify-center rounded-full text-xs font-black ${
                    isActive ? "bg-lucky-dark/10 text-lucky-dark" : "bg-white/10 text-lucky-gold"
                  }`}
                >
                  {region.code}
                </span>
                <div>
                  <div className="text-sm font-bold">{region.label}</div>
                  <div className={`text-[11px] ${isActive ? "text-lucky-dark/75" : "text-gray-400"}`}>
                    {region.code === "ID" ? "IDR methods" : "PHP methods"}
                  </div>
                </div>
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}
