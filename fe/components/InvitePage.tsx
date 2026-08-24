"use client"

import React from "react"
import { Share2, Copy } from "lucide-react"
import { useLanguage } from "../app/contexts/LanguageContext"
import { gameService } from "../services/api"
import type { InviteInfoResponse } from "../app/mocks/types"

const InvitePage: React.FC = () => {
  const { t } = useLanguage()

  const [inviteCode, setInviteCode] = React.useState("")
  const [isLoading, setIsLoading] = React.useState(true)
  const [showCopySuccess, setShowCopySuccess] = React.useState(false)

  React.useEffect(() => {
    const fetchInviteInfo = async () => {
      try {
        const response: InviteInfoResponse = await gameService.getInviteInfo()
        if (response.code === 0) {
          setInviteCode(response.data.invite_code)
        }
      } catch (error) {
        console.error("Failed to fetch invite info:", error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchInviteInfo()
  }, [])

  const inviteLink = inviteCode
    ? `${typeof window !== "undefined" ? window.location.origin : ""}?code=${inviteCode}`
    : "Loading..."

  const handleCopyLink = () => {
    navigator.clipboard.writeText(inviteLink)
    setShowCopySuccess(true)
    setTimeout(() => {
      setShowCopySuccess(false)
    }, 2000)
  }

  return (
    <div className="pt-8 pb-24 px-4 max-w-2xl mx-auto animate-fade-in">
      {/*
      Hidden for now:
      - Total Pool
      - User Stats
      - Commission Structure
      - How It Works Steps
      - Leaderboard
      */}

      <div className="bg-lucky-dark border border-white/10 rounded-2xl p-4 mb-6">
        <label className="text-xs text-gray-400 mb-2 block">{t("invite.link_label")}</label>
        <div className="flex gap-2">
          <div className="flex-1 bg-white/5 rounded-lg px-3 py-3 text-sm text-gray-300 truncate font-mono">
            {isLoading ? "Loading..." : inviteLink}
          </div>
          <button
            onClick={handleCopyLink}
            disabled={isLoading}
            className="bg-lucky-gold text-lucky-dark font-bold px-4 rounded-lg hover:bg-white transition-colors disabled:opacity-50"
          >
            <Copy size={18} />
          </button>
        </div>
        {showCopySuccess && (
          <div className="mt-2 text-sm text-green-400 flex items-center gap-2 animate-fade-in">
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                clipRule="evenodd"
              />
            </svg>
            {t("invite.copy_success")}
          </div>
        )}
        <button className="w-full mt-4 bg-gradient-to-r from-lucky-pink to-purple-600 py-3 rounded-xl font-bold text-white shadow-lg shadow-lucky-pink/20 flex items-center justify-center gap-2">
          <Share2 size={18} /> {t("invite.btn_invite")}
        </button>
      </div>
    </div>
  )
}

export default InvitePage
