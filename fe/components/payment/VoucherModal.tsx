"use client"

import { useState, useEffect } from "react"
import { X, Loader2, Gift, ExternalLink, Info } from "lucide-react"
import { paymentService } from "@/services/payment"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { Portal } from "@/components/ui/portal"

interface VoucherModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (voucherCode: string) => Promise<void>
  amount: number
}

export function VoucherModal({ isOpen, onClose, onSubmit, amount }: VoucherModalProps) {
  const { t } = useLanguage()
  const [voucherCode, setVoucherCode] = useState("")
  const [loading, setLoading] = useState(false)
  const [purchaseUrl, setPurchaseUrl] = useState("")
  const [configLoading, setConfigLoading] = useState(true)

  // 获取卡密购买网站配置
  useEffect(() => {
    if (isOpen) {
      loadConfig()
    }
  }, [isOpen])

  const loadConfig = async () => {
    setConfigLoading(true)
    try {
      const config = await paymentService.getVoucherConfig()
      setPurchaseUrl(config.purchase_url)
    } catch (error) {
      console.error("获取配置失败:", error)
      // 使用默认链接
      setPurchaseUrl("https://www.188topup.com")
    } finally {
      setConfigLoading(false)
    }
  }

  if (!isOpen) return null

  const handleSubmit = async () => {
    if (!voucherCode.trim()) return
    setLoading(true)
    try {
      await onSubmit(voucherCode.trim())
      setVoucherCode("")
    } finally {
      setLoading(false)
    }
  }

  // 格式化卡密输入：XXXX-XXXX-XXXX-XXXX
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let value = e.target.value.replace(/[^a-zA-Z0-9]/g, "").toUpperCase()
    if (value.length > 16) value = value.slice(0, 16)

    // 添加分隔符
    const parts = []
    for (let i = 0; i < value.length; i += 4) {
      parts.push(value.slice(i, i + 4))
    }
    setVoucherCode(parts.join("-"))
  }

  // 打开购买网站
  const openPurchaseSite = () => {
    if (purchaseUrl) {
      window.open(purchaseUrl, "_blank")
    }
  }

  return (
    <Portal>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
        <div className="bg-lucky-dark border border-white/10 rounded-2xl w-full max-w-md mx-4 p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-lucky-gold/20 flex items-center justify-center">
                <Gift className="text-lucky-gold" size={20} />
              </div>
              <div>
                <h3 className="font-bold text-white">{t("wallet.voucher_title")}</h3>
                <p className="text-xs text-gray-400">{t("wallet.voucher_subtitle")}</p>
              </div>
            </div>
            <button onClick={onClose} className="text-gray-400 hover:text-white">
              <X size={20} />
            </button>
          </div>

          {/* 购买提示区域 */}
          <div className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/20 rounded-xl p-4 mb-6">
            <div className="flex items-start gap-3">
              <Info className="text-blue-400 shrink-0 mt-0.5" size={18} />
              <div className="flex-1">
                <p className="text-sm text-gray-300 mb-2">
                  {t("wallet.voucher_no_voucher_hint")}
                </p>
                <button
                  onClick={openPurchaseSite}
                  disabled={configLoading || !purchaseUrl}
                  className="flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300 transition-colors"
                >
                  <ExternalLink size={14} />
                  {configLoading ? t("wallet.voucher_loading") : t("wallet.voucher_go_purchase")}
                </button>
              </div>
            </div>
          </div>

          {/* Amount Display */}
          <div className="bg-white/5 rounded-xl p-4 mb-6 text-center">
            <div className="text-sm text-gray-400 mb-1">{t("wallet.voucher_amount_label")}</div>
            <div className="text-2xl font-bold text-lucky-gold">
              {amount.toLocaleString()} IDR
            </div>
          </div>

          {/* Voucher Code Input */}
          <div className="mb-6">
            <label className="block text-sm text-gray-400 mb-2">{t("wallet.voucher_code_label")}</label>
            <input
              type="text"
              value={voucherCode}
              onChange={handleInputChange}
              placeholder={t("wallet.voucher_code_placeholder")}
              className="w-full bg-black/30 border border-white/20 rounded-xl px-4 py-4 text-white text-center text-lg tracking-wider font-mono focus:border-lucky-gold focus:outline-none uppercase"
              maxLength={19}
              disabled={loading}
            />
            <p className="text-xs text-gray-500 mt-2 text-center">
              {t("wallet.voucher_code_hint")}
            </p>
          </div>

          {/* Submit Button */}
          <button
            onClick={handleSubmit}
            disabled={loading || voucherCode.length < 19}
            className="w-full py-4 rounded-full bg-gradient-to-r from-lucky-gold to-orange-500 text-lucky-dark font-bold text-lg shadow-xl shadow-lucky-gold/20 hover:scale-[1.02] transition-transform disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {loading ? (
              <>
                <Loader2 className="animate-spin" size={20} />
                {t("wallet.voucher_verifying")}
              </>
            ) : (
              t("wallet.voucher_redeem_now")
            )}
          </button>

          {/* Demo Notice */}
          <div className="mt-4 p-3 bg-green-500/10 border border-green-500/20 rounded-lg">
            <p className="text-xs text-green-400 text-center">
              ✅ {t("wallet.voucher_demo_notice")}
            </p>
          </div>
        </div>
      </div>
    </Portal>
  )
}
