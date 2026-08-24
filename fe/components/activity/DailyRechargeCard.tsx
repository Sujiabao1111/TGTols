"use client";

import React from "react";
import { Gift, CheckCircle2 } from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";

interface TierInfo {
  day: number;
  rate: number;
  label: string;
  description: string;
  status: "upcoming" | "active" | "completed" | "claimed";
  deposit_amount: number;
  reward_amount: number;
  min_deposit: number;
}

interface DailyRechargeCardProps {
  tiers: TierInfo[];
  onClaim: (day: number) => void;
  claimingDay: number | null;
  onDeposit?: () => void;
  onShowDetails?: () => void;
}

const getBoxImage = (status: string): string => {
  const images: Record<string, string> = {
    claimed: "/images/box_opened.png",
    completed: "/images/box_opened.png",
    active: "/images/box_available.png",
    upcoming: "/images/box_unavailable.png",
  };
  return images[status] || "/images/box_unavailable.png";
};

export const DailyRechargeCard: React.FC<DailyRechargeCardProps> = ({
  tiers,
  onClaim,
  claimingDay,
  onDeposit,
  onShowDetails,
}) => {
  const { t } = useLanguage();

  const activeTier = tiers.find((t) => t.status === "active");
  const completedTier = tiers.find((t) => t.status === "completed");

  return (
    <div className="relative rounded-3xl overflow-hidden bg-gradient-to-br from-fuchsia-500/90 via-violet-500/90 to-blue-500/90 p-6 shadow-2xl border border-white/15 min-h-[360px]">
      {/* Background effects */}
      <div className="absolute inset-0 opacity-30">
        <div className="absolute -top-10 -right-6 h-36 w-36 rounded-full bg-white/20 blur-2xl" />
        <div className="absolute -bottom-10 -left-6 h-40 w-40 rounded-full bg-black/20 blur-2xl" />
      </div>

      <div className="relative flex h-full flex-col">
        {/* Header */}
        <div className="mb-4">
          <h3 className="text-2xl sm:text-3xl font-bold text-white">
            {t("activity.welcome_bonus_title")}
          </h3>
          <p className="text-white/80 text-sm mt-1 max-w-sm">
            {t("activity.welcome_bonus_desc")}
          </p>
        </div>

        {/* Progress boxes */}
        <div className="mt-4">
          <div className="grid grid-cols-4 gap-2 text-center">
            {tiers.map((tier) => (
              <div key={tier.day} className="space-y-2">
                <div className="bg-white/20 rounded-full py-1 text-[10px] text-white/90">
                  Dep {tier.day}
                </div>
                <div
                  className={`h-16 rounded-2xl flex items-center justify-center border-2 overflow-hidden ${
                    tier.status === "claimed"
                      ? "bg-green-500/30 border-green-400/50"
                      : tier.status === "completed"
                      ? "bg-white/30 border-white/50"
                      : tier.status === "active"
                      ? "bg-white/30 border-white/50 animate-pulse"
                      : "bg-white/15 border-white/20"
                  }`}
                >
                  {tier.status === "claimed" ? (
                    <CheckCircle2 size={28} className="text-green-400" />
                  ) : (
                    <img
                      src={getBoxImage(tier.status)}
                      alt={`Dep ${tier.day}`}
                      className={`w-10 h-10 object-contain ${
                        tier.status === "upcoming" ? "opacity-50 grayscale" : ""
                      }`}
                    />
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* Progress dots */}
          <div className="mt-4 flex items-center px-1">
            {tiers.map((tier, index) => (
              <div key={`dot-${tier.day}`} className="flex items-center flex-1">
                <div
                  className={`h-3 w-3 rounded-full ${
                    tier.status === "claimed"
                      ? "bg-green-400"
                      : tier.status === "completed" || tier.status === "active"
                      ? "bg-white scale-125"
                      : "bg-white/30"
                  }`}
                />
                {index < tiers.length - 1 && (
                  <div
                    className={`h-1 flex-1 mx-2 rounded-full ${
                      tier.status === "claimed" ? "bg-green-400" : "bg-white/30"
                    }`}
                  />
                )}
              </div>
            ))}
          </div>

          {/* Rate labels */}
          <div className="mt-3 grid grid-cols-4 gap-2 text-center text-[11px] text-white/90">
            {tiers.map((tier) => (
              <div
                key={`rate-${tier.day}`}
                className={`rounded-xl py-1.5 ${
                  tier.status === "claimed"
                    ? "bg-green-500/30 text-green-300"
                    : tier.status === "active" || tier.status === "completed"
                    ? "bg-white/20 text-white"
                    : "bg-white/10 text-white/60"
                }`}
              >
                {tier.label}
              </div>
            ))}
          </div>
        </div>

        {/* Action buttons */}
        <div className="mt-auto pt-4 flex gap-3">
          <button
            onClick={onShowDetails}
            className="flex-1 bg-white/20 hover:bg-white/30 text-white font-semibold py-3 rounded-xl transition-colors"
          >
            {t("activity.details")}
          </button>
          {activeTier && onDeposit && (
            <button
              onClick={onDeposit}
              className="flex-1 bg-white text-purple-900 font-semibold py-3 rounded-xl hover:bg-white/90 transition-colors"
            >
              {t("activity.deposit_now")}
            </button>
          )}
          {!activeTier && (
            <button className="flex-1 bg-white/30 text-white/70 font-semibold py-3 rounded-xl cursor-not-allowed">
              {completedTier ? t("activity.claim") : t("activity.completed")}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
