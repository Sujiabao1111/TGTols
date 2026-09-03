"use client"

import React from "react"
import {
  Copy,
  RefreshCcw,
  LogOut,
  Wallet,
} from "lucide-react"
import type { View, WalletTab } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { useExchangeRate } from "@/hooks/useExchangeRate"
import { useUser } from "@/app/contexts/UserContext"
import { gameService } from "../services/api"
import { showAddDesktopInsuranceDialog } from "./AddDesktopInsuranceDialog"

interface ProfilePageProps {
  onNavigate: (view: View) => void
  onSetWalletTab: (tab: WalletTab) => void
  onLogout: () => void
}

const ProfilePage: React.FC<ProfilePageProps> = ({ onNavigate, onSetWalletTab, onLogout }) => {
  const { t } = useLanguage()
  const { formatUSD } = useExchangeRate()
  const { userBalance, withdrawableBalance, fetchUserBalance, userProfile, isUserInfoLoaded } = useUser()
  const [isRecycling, setIsRecycling] = React.useState(false)
  const [recycleMessage, setRecycleMessage] = React.useState<string | null>(null)

  const handleRecycleBalance = async () => {
    setIsRecycling(true)
    setRecycleMessage(null)

    try {
      const response = await gameService.recycleBalance()

      if (response.code === 0) {
        const recycledAmount = response.data?.recycled_amount || 0
        setRecycleMessage(`${t("profile.recycle_success")}: $${recycledAmount.toFixed(2)}`)
        if (response.data?.insurance?.triggered) {
          const amount = `${(response.data.insurance.compensation_amount_u || 0).toFixed(2)}U`
          showAddDesktopInsuranceDialog(amount)
        }
        await fetchUserBalance()
      } else {
        setRecycleMessage(response.message || t("profile.recycle_error"))
      }
    } catch (error) {
      setRecycleMessage(t("profile.recycle_error"))
      console.error("Recycle balance error:", error)
    } finally {
      setIsRecycling(false)

      setTimeout(() => {
        setRecycleMessage(null)
      }, 3000)
    }
  }

  const handleNavToWallet = (tab: WalletTab) => {
    onSetWalletTab(tab)
    onNavigate("wallet")
  }

  const handleCopyUID = () => {
    navigator.clipboard.writeText(userProfile.uid.toString())
  }

  if (!isUserInfoLoaded) {
    return (
      <div className="pt-20 pb-24 px-4 max-w-2xl mx-auto animate-fade-in flex items-center justify-center">
        <div className="text-gray-400">{t("common.loading")}</div>
      </div>
    )
  }

  return (
    <div className="pt-20 pb-24 px-4 max-w-2xl mx-auto animate-fade-in">
      {/* Header Profile */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <div className="w-16 h-16 rounded-full bg-gradient-to-tr from-lucky-gold to-lucky-pink p-1">
            <img
              src="https://picsum.photos/seed/bearlucky/200/200"
              alt={userProfile.username}
              className="w-full h-full object-cover rounded-full border-2 border-lucky-dark"
            />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-white">{userProfile.username}</h2>
            <div className="flex items-center gap-2 mt-1">
              <span className="text-xs text-gray-400">UID: {userProfile.uid}</span>
              <Copy size={12} className="text-gray-500 cursor-pointer" onClick={handleCopyUID} />
            </div>
          </div>
        </div>
        <div className="flex flex-col items-end">
          <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-lucky-gold/70">{t("nav.vip")}</div>
          <div className="mt-1 text-4xl font-display font-bold italic text-lucky-gold">VIP {userProfile.vipLevel}</div>
          <button
            type="button"
            onClick={() => onNavigate("vip")}
            className="mt-2 rounded-full border border-[#ffcf66] bg-[linear-gradient(180deg,#ffda75_0%,#f2b93a_52%,#b97816_100%)] px-3 py-1 text-[11px] font-black text-[#241300] shadow-[0_10px_22px_rgba(255,193,74,0.22)] transition-transform hover:scale-[1.03] active:scale-[0.98]"
          >
            {t("profile.vip_details")} &gt;
          </button>
        </div>
      </div>

      {/* Wallet Summary */}
      <div className="bg-lucky-dark border border-white/10 rounded-2xl p-5 mb-6 relative overflow-hidden">
        <div className="mb-4">
          <div>
            <div className="text-lucky-gold text-2xl font-bold font-display">{formatUSD(userBalance)}</div>
            <div className="text-xs text-gray-400">{t("profile.total_asset")}</div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-6">
          <div className="bg-black/20 rounded-lg p-2">
            <div className="text-xs text-gray-400 mb-1">{t("wallet.withdrawable")}</div>
            <div className="text-green-400 font-bold">{formatUSD(withdrawableBalance)}</div>
          </div>
          {/* <div className="bg-black/20 rounded-lg p-2">
            <div className="text-xs text-gray-400 mb-1">{t("wallet.bonus")}</div>
            <div className="text-blue-400 font-bold">{formatFromUSD(100)}</div>
          </div> */}
        </div>

        <div className="flex gap-3">
          <button
            onClick={() => handleNavToWallet("deposit")}
            className="flex-1 py-3 rounded-xl bg-blue-500/20 text-blue-400 font-bold border border-blue-500/30 flex items-center justify-center gap-2 hover:bg-blue-500/30 transition-colors"
          >
            <RefreshCcw size={16} /> {t("wallet.tab.deposit")}
          </button>
          <button
            onClick={() => handleNavToWallet("withdraw")}
            className="flex-1 py-3 rounded-xl bg-green-500/20 text-green-400 font-bold border border-green-500/30 flex items-center justify-center gap-2 hover:bg-green-500/30 transition-colors"
          >
            <Wallet size={16} /> {t("wallet.tab.withdraw")}
          </button>
        </div>

        <div className="mt-3">
          <button
            onClick={handleRecycleBalance}
            disabled={isRecycling}
            className="w-full py-3 rounded-xl bg-yellow-500/20 text-yellow-400 font-bold border border-yellow-500/30 flex items-center justify-center gap-2 hover:bg-yellow-500/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <RefreshCcw size={16} className={isRecycling ? "animate-spin" : ""} />
            {isRecycling ? t("profile.recycling") : t("profile.recycle_balance")}
          </button>

          {recycleMessage && (
            <div
              className={`mt-2 text-xs text-center ${recycleMessage.includes("success") || recycleMessage.includes("$") ? "text-green-400" : "text-red-400"}`}
            >
              {recycleMessage}
            </div>
          )}
        </div>
      </div>
      <button
        onClick={onLogout}
        className="hidden w-full mt-8 py-3 rounded-xl border border-white/10 text-gray-400 hover:text-white hover:border-white/30 transition-colors items-center justify-center gap-2"
      >
        <LogOut size={16} /> {t("profile.logout")}
      </button>
    </div>
  )
}

export default ProfilePage
