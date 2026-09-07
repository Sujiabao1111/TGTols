"use client"

import { useState, useEffect, useCallback } from "react"
import { WalletRecord, paymentService } from "@/services/payment"
import { useLanguage } from "@/app/contexts/LanguageContext"
import { Loader2, ArrowDownLeft, ArrowUpRight, Gift, Search } from "lucide-react"

type WalletRecordType = "deposit" | "withdraw" | "bonus"

export function PaymentHistory() {
  const { t, language } = useLanguage()
  const [records, setRecords] = useState<WalletRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState("")

  const loadOrders = useCallback(async () => {
    try {
      const data = await paymentService.getUserPayments(20)
      setRecords(data)
    } catch (error) {
      console.error("Failed to load payment history:", error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    queueMicrotask(() => {
      void loadOrders()
    })
  }, [loadOrders])

  const getStatusText = (status: number) => {
    switch (status) {
      case 0:
        return t("wallet.deposit_pending")
      case 1:
        return t("wallet.status_completed")
      case 2:
        return t("wallet.deposit_failed")
      case 3:
        return t("wallet.status_processing")
      default:
        return t("wallet.status_unknown")
    }
  }

  const getStatusColor = (status: number) => {
    switch (status) {
      case 0:
        return "text-orange-500"
      case 1:
        return "text-green-500"
      case 2:
        return "text-red-500"
      case 3:
        return "text-blue-500"
      default:
        return "text-gray-500"
    }
  }

  const getNormalizedRecordType = (record: WalletRecord): WalletRecordType => {
    if (record.record_type === "deposit" || record.record_type === "withdraw" || record.record_type === "bonus") {
      return record.record_type
    }

    const text = [
      record.id,
      record.title,
      record.remark,
      record.reference_id,
      record.order_id,
    ].filter(Boolean).join(" ").toLowerCase()
    if (
      record.type === 6 ||
      text.includes("desktop") ||
      text.includes("add_desktop") ||
      text.includes("daily-weekly-challenge") ||
      text.includes("new-user-recharge") ||
      text.includes("recharge-rebate") ||
      text.includes("vip-monthly-bonus") ||
      text.includes("manual_reward") ||
      text.includes("bonus") ||
      text.includes("reward")
    ) {
      return "bonus"
    }

    if (record.amount < 0) {
      return "withdraw"
    }

    return "deposit"
  }

  const getRecordTypeText = (record: WalletRecord) => {
    switch (getNormalizedRecordType(record)) {
      case "withdraw":
        return t("wallet.type_withdraw")
      case "bonus":
        return t("wallet.type_bonus")
      default:
        return t("wallet.type_deposit")
    }
  }

  const getRecordIcon = (recordType: WalletRecordType) => {
    if (recordType === "withdraw") {
      return {
        wrapper: "bg-red-500/20",
        icon: <ArrowUpRight className="text-red-400" size={20} />,
      }
    }
    if (recordType === "bonus") {
      return {
        wrapper: "bg-lucky-gold/20",
        icon: <Gift className="text-lucky-gold" size={20} />,
      }
    }

    return {
      wrapper: "bg-green-500/20",
      icon: <ArrowDownLeft className="text-green-500" size={20} />,
    }
  }

  const getAmountClass = (recordType: WalletRecordType) => {
    if (recordType === "withdraw") {
      return "text-red-400"
    }
    if (recordType === "bonus") {
      return "text-lucky-gold"
    }
    return "text-green-400"
  }

  const getAmountPrefix = (recordType: WalletRecordType) => {
    return recordType === "withdraw" ? "-" : "+"
  }

  const getRecordCurrency = () => {
    // Platform wallet amounts are displayed in U regardless of the funding channel.
    return "U"
  }

  const getDateLocale = () => {
    switch (language) {
      case "ru":
        return "ru-RU"
      case "es":
        return "es-ES"
      case "cn":
        return "zh-CN"
      case "id":
        return "id-ID"
      case "ph":
        return "en-PH"
      default:
        return "en-US"
    }
  }

  const getSearchableRecordText = (record: WalletRecord) => {
    return [
      record.id,
      record.title,
      record.remark,
      record.reference_id,
      record.order_id,
      record.dst_code,
    ].filter(Boolean).join(" ").toLowerCase()
  }

  const hasCjkText = (value?: string) => {
    return /[\u3400-\u9fff]/.test(value || "")
  }

  const getBonusTitle = (record: WalletRecord) => {
    const text = getSearchableRecordText(record)
    if (text.includes("add-desktop-insurance") || text.includes("desktop insurance")) {
      return t("wallet.bonus_add_desktop_insurance")
    }
    if (text.includes("add_desktop") || text.includes("desktop")) {
      return t("wallet.bonus_add_desktop")
    }
    if (text.includes("daily-weekly-challenge") || text.includes("daily weekly challenge")) {
      return t("wallet.bonus_activity_task")
    }
    if (text.includes("new-user-recharge") || text.includes("new user recharge")) {
      return t("wallet.bonus_new_user_recharge")
    }
    if (text.includes("recharge-rebate") || text.includes("recharge rebate")) {
      return t("wallet.bonus_recharge_rebate")
    }
    if (text.includes("vip-monthly-bonus") || text.includes("vip monthly")) {
      return t("wallet.bonus_vip_monthly")
    }
    if (text.includes("mail")) {
      return t("wallet.bonus_mail")
    }
    if (text.includes("manual_reward") || text.includes("manual reward")) {
      return t("wallet.bonus_manual")
    }
    if (text.includes("withdraw refund")) {
      return t("wallet.bonus_withdraw_refund")
    }
    if (hasCjkText(record.title) || hasCjkText(record.remark)) {
      return t("wallet.bonus_generic")
    }
    return record.title || record.remark || t("wallet.bonus_generic")
  }

  const getRecordSubtitle = (record: WalletRecord) => {
    if (getNormalizedRecordType(record) === "bonus") {
      return getBonusTitle(record)
    }
    return record.dst_code || record.title
  }

  const filteredRecords = records.filter((record) => {
    const query = searchQuery.toLowerCase()
    return getSearchableRecordText(record).includes(query)
  })

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="animate-spin text-lucky-gold" size={24} />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="flex items-center bg-white/5 rounded-xl px-4 py-3 border border-white/10">
        <Search size={18} className="text-gray-400 mr-3" />
        <input
          type="text"
          placeholder={t("wallet.search_order")}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="bg-transparent border-none focus:outline-none text-sm text-white w-full"
        />
      </div>

      {/* Records */}
      <div className="space-y-3">
        {filteredRecords.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            <p>{t("wallet.no_records")}</p>
          </div>
        ) : (
          filteredRecords.map((record) => (
            (() => {
              const recordType = getNormalizedRecordType(record)
              const currency = getRecordCurrency()
              const icon = getRecordIcon(recordType)
              const reference = record.order_id || record.reference_id

              return (
                <div
                  key={record.id}
                  className="bg-lucky-dark border border-white/10 rounded-xl p-4 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-full flex items-center justify-center ${icon.wrapper}`}>
                      {icon.icon}
                    </div>
                    <div>
                      <div className="font-bold text-white text-sm">
                        {getRecordTypeText(record)} - {getRecordSubtitle(record)}
                      </div>
                      <div className="text-xs text-gray-500">
                        {new Date(record.created_at).toLocaleString(getDateLocale())}
                      </div>
                      {reference && <div className="text-xs text-gray-600 mt-0.5">{reference}</div>}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className={`font-bold text-sm ${getAmountClass(recordType)}`}>
                      {getAmountPrefix(recordType)}{Math.abs(record.amount).toLocaleString()} {currency}
                    </div>
                    <div className={`text-[10px] ${getStatusColor(record.status)}`}>
                      {getStatusText(record.status)}
                    </div>
                  </div>
                </div>
              )
            })()
          ))
        )}
      </div>
    </div>
  )
}
