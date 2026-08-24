"use client";

import React from "react";
import { X, AlertCircle, Gift } from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";
import { Portal } from "../ui/portal";

interface TierInfo {
  day: number;
  rate: number;
  label: string;
  description: string;
}

interface ActivityTermsModalProps {
  isOpen: boolean;
  onClose: () => void;
  activityName: string;
  tiers: TierInfo[];
  minDeposit: number;
}

export const ActivityTermsModal: React.FC<ActivityTermsModalProps> = ({
  isOpen,
  onClose,
  activityName,
  tiers,
  minDeposit,
}) => {
  const { t } = useLanguage();

  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="relative bg-gradient-to-br from-fuchsia-600 via-violet-600 to-blue-600 w-full sm:max-w-md sm:w-full shadow-2xl rounded-t-3xl sm:rounded-3xl max-h-[85vh] overflow-y-auto sm:!mb-0 p-1" style={{ marginBottom: '60px' }}>
        {/* Inner content */}
        <div className="bg-gray-900/95 rounded-[22px] p-6">
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute right-4 top-4 text-white/60 hover:text-white transition-colors z-10"
          >
            <X size={24} />
          </button>

          {/* Header */}
          <div className="text-center mb-6">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-gradient-to-br from-fuchsia-500 to-violet-500 mb-3">
              <Gift size={24} className="text-white" />
            </div>
            <h3 className="text-xl font-bold text-white">
              {t("activity.bonus_terms")} +{tiers[0]?.rate * 100 || 50}%
            </h3>
          </div>

          {/* Terms content */}
          <div className="space-y-4 text-sm">
            {/* Min deposit */}
            <div className="flex justify-between text-white/80">
              <span>{t("activity.min_deposit_label") || "Min. deposit"}</span>
              <span className="text-white font-semibold">${minDeposit}</span>
            </div>

            {/* Max bonus */}
            <div className="flex justify-between text-white/80">
              <span>{t("activity.max_bonus_label") || "Max. bonus amount"}</span>
              <span className="text-white font-semibold">10x</span>
            </div>

            {/* Wagering */}
            <div className="flex justify-between text-white/80">
              <span>{t("activity.wagering_label") || "Wager"}</span>
              <span className="text-white font-semibold">x40</span>
            </div>

            {/* Validity */}
            <div className="flex justify-between text-white/80">
              <span>{t("activity.validity_label") || "Validity"}</span>
              <span className="text-white font-semibold">30 {t("activity.days") || "days"}</span>
            </div>

            {/* Bonus tiers */}
            <div className="bg-white/5 rounded-xl p-4 mt-4">
              <h4 className="text-white font-semibold mb-3">{t("activity.bonus_tiers") || "Bonus Tiers"}</h4>
              <div className="space-y-2">
                {tiers.map((tier) => (
                  <div key={tier.day} className="flex justify-between text-sm">
                    <span className="text-white/70">{tier.label}</span>
                    <span className="text-white font-medium">+{tier.rate * 100}%</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Warning section */}
            <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-xl p-4 mt-4">
              <div className="flex items-start gap-3">
                <AlertCircle size={20} className="text-yellow-400 flex-shrink-0 mt-0.5" />
                <div>
                  <h4 className="text-yellow-400 font-semibold mb-2">
                    {t("activity.be_careful")}
                  </h4>
                  <ul className="space-y-1 text-white/70 text-xs">
                    <li>• {t("activity.warning_1") || "Available only for certain games"}</li>
                    <li>• {t("activity.warning_2") || "When withdrawing funds from the real balance, all uncancelled bonuses will be cancelled"}</li>
                    <li>• {t("activity.warning_3") || "Wagering is completed first from the real balance, then from the bonus balance"}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          {/* Close button */}
          <button
            onClick={onClose}
            className="w-full mt-6 bg-white/20 hover:bg-white/30 text-white font-semibold py-3 rounded-xl transition-colors"
          >
            {t("activity.close")}
          </button>
        </div>
      </div>
    </div>
    </Portal>
  );
};
