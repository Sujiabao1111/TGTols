"use client"

export function PaymentRegionSelector() {
  return (
    <div className="mb-6 rounded-2xl border border-lucky-gold/25 bg-gradient-to-r from-[#171717] via-[#121212] to-[#1a1a1a] p-2 shadow-[0_12px_30px_rgba(0,0,0,0.28)]">
      <div className="grid grid-cols-1 gap-2">
        <button
          type="button"
          disabled
          className="min-w-0 cursor-default rounded-xl border border-[#229ED9] bg-[#229ED9] px-2 py-2.5 text-left text-white"
        >
          <div className="flex min-w-0 items-center gap-2">
            <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/10 text-[10px] font-black">TG</span>
            <div className="min-w-0"><div className="truncate text-xs font-bold leading-tight">Telegram</div><div className="truncate text-[9px] leading-tight opacity-75">Stars</div></div>
          </div>
        </button>
      </div>
    </div>
  )
}
