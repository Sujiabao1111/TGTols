"use client";

import React, { useState, useEffect } from "react";
import {
  TrendingDown,
  TrendingUp,
  DollarSign,
  Calendar,
  Clock,
  Info,
} from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";

interface LossStats {
  bet_amount: number;
  win_amount: number;
  net_loss: number;
  rebate_rate: number;
  rebate_amount: number;
  min_loss_threshold: number;
  claimed: boolean;
}

interface LossRebateCardProps {
  activityName: string;
  startTime: string;
  endTime: string;
  rebateRate: number;
  minLossThreshold: number;
  todayStats: LossStats;
  totalStats?: {
    total_bet: number;
    total_win: number;
    total_loss: number;
    total_rebate: number;
    claimed_rebate: number;
    pending_rebate: number;
  };
  canClaim: boolean;
  onClaim: () => void;
  claiming: boolean;
  onShowDetails?: () => void;
  countdown?: number;
  isActive?: boolean;
}

export const LossRebateCard: React.FC<LossRebateCardProps> = ({
  activityName,
  startTime,
  endTime,
  rebateRate,
  minLossThreshold,
  todayStats,
  totalStats,
  canClaim,
  onClaim,
  claiming,
  onShowDetails,
  countdown,
  isActive = true,
}) => {
  const { t } = useLanguage();
  const [timeLeft, setTimeLeft] = useState(countdown || 0);

  useEffect(() => {
    if (countdown && countdown > 0) {
      setTimeLeft(countdown);
    }
  }, [countdown]);

  useEffect(() => {
    if (timeLeft <= 0) return;

    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [timeLeft]);

  const formatCountdown = (seconds: number) => {
    const days = Math.floor(seconds / (24 * 60 * 60));
    const hours = Math.floor((seconds % (24 * 60 * 60)) / (60 * 60));
    const minutes = Math.floor((seconds % (60 * 60)) / 60);
    const secs = seconds % 60;

    return `${days}d ${hours.toString().padStart(2, "0")}h ${minutes
      .toString()
      .padStart(2, "0")}m ${secs.toString().padStart(2, "0")}s`;
  };

  const belowThreshold =
    todayStats.net_loss > 0 && todayStats.net_loss < minLossThreshold;

  return (
    <div className="relative rounded-3xl overflow-hidden shadow-2xl border border-white/15 min-h-[360px]"
         style={{ background: "linear-gradient(135deg, #c084fc 0%, #e879f9 50%, #a855f7 100%)" }}>
      {/* Background image */}
      <div
        className="absolute inset-0 bg-cover bg-center opacity-40"
        style={{
          backgroundImage: `url('/images/activity-3-desktop.webp')`,
        }}
      />

      {/* Gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-t from-purple-900/40 via-transparent to-transparent" />

      <div className="relative flex h-full flex-col min-h-[360px] p-6">
        {/* Header */}
        <div className="flex justify-between items-start">
          <div>
            <h3 className="text-2xl sm:text-3xl font-bold text-white mb-2 drop-shadow-lg">
              {activityName || t("activity.cashback_title")}
            </h3>
            <p className="text-white/90 text-sm max-w-[280px] drop-shadow-md">
              {t("activity.cashback_desc")}
            </p>
          </div>
          {/* Info button */}
          {onShowDetails && (
            <button
              onClick={onShowDetails}
              className="h-8 w-8 rounded-full bg-white/20 hover:bg-white/30 flex items-center justify-center text-white transition-colors backdrop-blur-sm"
            >
              <Info size={18} />
            </button>
          )}
        </div>

        {/* Spacer */}
        <div className="flex-1" />

        {/* Countdown or Stats */}
        {!isActive && timeLeft > 0 ? (
          /* Countdown section */
          <div className="mt-auto">
            <p className="text-white/80 text-sm mb-2">
              {t("activity.starting_soon") || "Starting soon"}
            </p>
            <div className="flex items-center gap-2 text-white text-xl font-bold font-mono">
              <Clock size={20} className="text-white/80" />
              {formatCountdown(timeLeft)}
            </div>
          </div>
        ) : (
          /* Stats section */
          <>
            {/* Today's stats */}
            <div className="mt-4 p-4 bg-white/10 backdrop-blur-sm rounded-2xl border border-white/10">
              <div className="text-white/80 text-sm font-medium mb-3">
                {t("activity.today_stats")}
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div className="text-center">
                  <div className="flex items-center justify-center gap-1 text-white/60 text-xs mb-1">
                    <TrendingUp size={12} />
                    {t("activity.bet")}
                  </div>
                  <div className="text-white font-bold text-sm">
                    ${todayStats.bet_amount.toFixed(2)}
                  </div>
                </div>
                <div className="text-center">
                  <div className="flex items-center justify-center gap-1 text-white/60 text-xs mb-1">
                    <DollarSign size={12} />
                    {t("activity.win")}
                  </div>
                  <div className="text-white font-bold text-sm">
                    ${todayStats.win_amount.toFixed(2)}
                  </div>
                </div>
                <div className="text-center">
                  <div className="flex items-center justify-center gap-1 text-white/60 text-xs mb-1">
                    <TrendingDown size={12} />
                    {t("activity.loss")}
                  </div>
                  <div
                    className={`font-bold text-sm ${
                      todayStats.net_loss > 0 ? "text-red-300" : "text-white"
                    }`}
                  >
                    ${todayStats.net_loss.toFixed(2)}
                  </div>
                </div>
              </div>

              {/* Rebate amount */}
              {todayStats.net_loss > 0 && (
                <div className="mt-3 pt-3 border-t border-white/20">
                  <div className="flex justify-between items-center">
                    <span className="text-white/80 text-sm">
                      {t("activity.rebate_amount")}
                    </span>
                    <span className="text-yellow-300 font-bold text-lg">
                      ${todayStats.rebate_amount.toFixed(2)}
                    </span>
                  </div>
                  {belowThreshold && (
                    <div className="text-white/50 text-xs mt-1">
                      {t("activity.below_threshold")}: ${minLossThreshold}
                    </div>
                  )}
                </div>
              )}

              {/* Claim button */}
              <button
                onClick={onClaim}
                disabled={!canClaim || claiming}
                className={`w-full mt-3 py-3 rounded-xl font-semibold transition flex items-center justify-center gap-2 ${
                  canClaim
                    ? "bg-white text-purple-600 hover:bg-white/90 shadow-lg"
                    : "bg-white/20 text-white/50 cursor-not-allowed"
                }`}
              >
                {claiming ? (
                  <>
                    <div className="w-4 h-4 border-2 border-purple-600/30 border-t-purple-600 rounded-full animate-spin" />
                    ...
                  </>
                ) : todayStats.claimed ? (
                  <>{t("activity.already_claimed")}</>
                ) : belowThreshold ? (
                  <>{t("activity.below_threshold_short")}</>
                ) : todayStats.net_loss <= 0 ? (
                  <>
                    <Clock size={16} />
                    {t("activity.no_loss")}
                  </>
                ) : (
                  <>
                    {t("activity.claim")} ${todayStats.rebate_amount.toFixed(2)}
                  </>
                )}
              </button>
            </div>

            {/* Total stats */}
            {totalStats && (
              <div className="mt-3 p-3 bg-white/5 backdrop-blur-sm rounded-2xl border border-white/10">
                <div className="text-white/60 text-xs mb-2">
                  {t("activity.total_stats")}
                </div>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-white/60">{t("activity.total_bet")}</span>
                    <span className="text-white">
                      ${totalStats.total_bet.toFixed(2)}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-white/60">
                      {t("activity.total_loss")}
                    </span>
                    <span className="text-red-300">
                      ${totalStats.total_loss.toFixed(2)}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-white/60">
                      {t("activity.claimed_rebate")}
                    </span>
                    <span className="text-green-300">
                      ${totalStats.claimed_rebate.toFixed(2)}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-white/60">
                      {t("activity.pending_rebate")}
                    </span>
                    <span className="text-yellow-300">
                      ${totalStats.pending_rebate.toFixed(2)}
                    </span>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};
