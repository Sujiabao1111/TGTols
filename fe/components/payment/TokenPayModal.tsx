"use client"

import { QRCodeSVG } from "qrcode.react"
import { Copy, Check, X } from "lucide-react"
import { useState, useEffect } from "react"
import { useLanguage } from "@/app/contexts/LanguageContext"

interface TokenPayModalProps {
  isOpen: boolean
  onClose: () => void
  payAddress: string
  payAmount: string
  chainType: "ETH" | "TRX"
  expiredAt: number
}

export function TokenPayModal({
  isOpen,
  onClose,
  payAddress,
  payAmount,
  chainType,
  expiredAt,
}: TokenPayModalProps) {
  const { t } = useLanguage()
  const [copied, setCopied] = useState(false)
  const [timeLeft, setTimeLeft] = useState(0)

  useEffect(() => {
    if (!isOpen) return

    const updateTimer = () => {
      const seconds = Math.max(0, Math.floor((expiredAt * 1000 - Date.now()) / 1000))
      setTimeLeft(seconds)
    }

    updateTimer()
    const timer = setInterval(updateTimer, 1000)

    return () => clearInterval(timer)
  }, [isOpen, expiredAt])

  if (!isOpen) return null

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(payAddress)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // 复制失败静默处理
    }
  }

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60

    if (hours > 0) {
      return `${hours}小时 ${mins.toString().padStart(2, "0")}分 ${secs.toString().padStart(2, "0")}秒`
    }
    return `${mins}:${secs.toString().padStart(2, "0")}`
  }

  const chainNames: Record<string, string> = {
    ETH: "Ethereum (ERC20)",
    TRX: "Tron (TRC20)",
  }

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-lucky-dark rounded-2xl p-6 max-w-md w-full relative">
        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 text-gray-400 hover:text-white"
        >
          <X size={20} />
        </button>

        <h3 className="text-xl font-bold text-white mb-2 text-center">
          {t("wallet.usdt_title")}
        </h3>
        <p className="text-gray-400 text-sm text-center mb-4">
          {chainNames[chainType]}
        </p>

        {/* 倒计时 */}
        <div className="bg-yellow-500/20 rounded-lg p-3 mb-4 text-center">
          <p className="text-yellow-500 text-sm">
            {t("wallet.usdt_time_remaining")}: {formatTime(timeLeft)}
          </p>
        </div>

        {/* 二维码 */}
        <div className="bg-white p-4 rounded-xl flex justify-center mb-4">
          <QRCodeSVG value={payAddress} size={180} />
        </div>

        {/* 支付金额 */}
        <div className="bg-lucky-gold/20 rounded-lg p-3 mb-4 text-center">
          <p className="text-gray-400 text-xs mb-1">{t("wallet.usdt_pay_amount")}</p>
          <p className="text-2xl font-bold text-lucky-gold">{payAmount} USDT</p>
        </div>

        {/* 支付地址 */}
        <div className="bg-white/10 rounded-lg p-3 mb-4">
          <p className="text-gray-400 text-xs mb-1">{t("wallet.usdt_pay_address")}</p>
          <div className="flex items-center gap-2">
            <p className="text-white text-xs font-mono break-all flex-1">
              {payAddress}
            </p>
            <button
              onClick={handleCopy}
              className="p-2 bg-lucky-gold/20 rounded-lg hover:bg-lucky-gold/30 transition-colors shrink-0"
            >
              {copied ? (
                <Check size={16} className="text-green-400" />
              ) : (
                <Copy size={16} className="text-lucky-gold" />
              )}
              <span className="sr-only">{copied ? t("wallet.usdt_copied") : t("wallet.usdt_copy")}</span>
            </button>
          </div>
        </div>

        {/* 提示 */}
        <p className="text-gray-400 text-xs text-center">
          {t("wallet.usdt_hint")}
        </p>
      </div>
    </div>
  )
}
