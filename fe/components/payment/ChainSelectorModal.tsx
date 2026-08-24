"use client"

import { X } from "lucide-react"
import { useLanguage } from "@/app/contexts/LanguageContext"

interface ChainSelectorModalProps {
  isOpen: boolean
  onClose: () => void
  onSelect: (chain: "ETH" | "TRX") => void
  amount: number
  localRate: number
}

export function ChainSelectorModal({
  isOpen,
  onClose,
  onSelect,
  amount,
  localRate,
}: ChainSelectorModalProps) {
  const { t } = useLanguage()
  console.log("[ChainSelectorModal] Rendering, isOpen:", isOpen, "amount:", amount)
  if (!isOpen) return null

  const usdtAmount = (amount / localRate).toFixed(2)

  const chains = [
    {
      code: "ETH" as const,
      name: "Ethereum",
      standard: "ERC20",
      minAmount: 5,
      color: "from-blue-500 to-purple-500",
      icon: "⟠",
    },
    {
      code: "TRX" as const,
      name: "Tron",
      standard: "TRC20",
      minAmount: 10,
      color: "from-red-500 to-orange-500",
      icon: "🔴",
    },
  ]

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-lucky-dark rounded-2xl p-6 max-w-sm w-full relative">
        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 text-gray-400 hover:text-white"
        >
          <X size={20} />
        </button>

        <h3 className="text-xl font-bold text-white mb-2 text-center">
          {t("wallet.chain_title")}
        </h3>
        <p className="text-gray-400 text-sm text-center mb-6">
          ≈ {usdtAmount} USDT
        </p>

        <div className="space-y-3">
          {chains.map((chain) => (
            <button
              key={chain.code}
              onClick={() => onSelect(chain.code)}
              className={`w-full py-4 bg-gradient-to-r ${chain.color} rounded-xl flex items-center justify-between px-4 hover:opacity-90 transition-opacity`}
            >
              <div className="flex items-center gap-3">
                <span className="text-2xl">{chain.icon}</span>
                <div className="text-left">
                  <span className="text-white font-bold block">
                    {chain.code === "ETH" ? t("wallet.chain_eth") : t("wallet.chain_trx")}
                  </span>
                  <span className="text-white/70 text-xs">
                    {chain.standard}
                  </span>
                </div>
              </div>
              <span className="text-white/80 text-xs">
                {t("wallet.chain_min")} {chain.minAmount} USDT
              </span>
            </button>
          ))}
        </div>

        <button
          onClick={onClose}
          className="w-full mt-4 py-3 text-gray-400 hover:text-white transition-colors"
        >
          {t("wallet.chain_cancel")}
        </button>
      </div>
    </div>
  )
}
