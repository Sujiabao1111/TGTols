"use client";

import React from "react";
import { X, Info, CheckCircle } from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";
import { Portal } from "../ui/portal";

interface LossRebateTermsModalProps {
  isOpen: boolean;
  onClose: () => void;
  activityName: string;
  rebateRate: number;
  minLossThreshold: number;
  startTime: string;
  endTime: string;
}

export const LossRebateTermsModal: React.FC<LossRebateTermsModalProps> = ({
  isOpen,
  onClose,
  activityName,
  rebateRate,
  minLossThreshold,
  startTime,
  endTime,
}) => {
  const { t } = useLanguage();

  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="relative bg-white w-full sm:max-w-sm sm:w-full shadow-2xl overflow-hidden rounded-t-3xl sm:rounded-3xl max-h-[85vh] overflow-y-auto sm:!mb-0" style={{ marginBottom: '60px' }}>
        {/* Close button */}
        <button
          onClick={onClose}
          className="absolute right-3 top-3 z-10 h-8 w-8 rounded-full bg-black/20 hover:bg-black/30 flex items-center justify-center text-white transition-colors"
        >
          <X size={18} />
        </button>

        {/* Header image */}
        <div className="relative h-48 bg-gradient-to-br from-pink-400 via-fuchsia-500 to-purple-600">
          <img
            src="/images/activity-3-mobile.webp"
            alt="Cashback"
            className="w-full h-full object-cover"
          />
        </div>

        {/* Content */}
        <div className="p-6">
          {/* Title */}
          <h3 className="text-xl font-bold text-gray-900 mb-2">
            {activityName || t("activity.cashback_title")}
          </h3>

          {/* Description */}
          <p className="text-gray-500 text-sm mb-6">
            {t("activity.cashback_desc")}
          </p>

          {/* Terms list */}
          <div className="space-y-3">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-5 h-5 rounded-full bg-purple-100 flex items-center justify-center mt-0.5">
                <CheckCircle size={12} className="text-purple-600" />
              </div>
              <div>
                <p className="text-gray-700 text-sm font-medium">
                  {t("activity.rebate_rate_label") || "Rebate Rate"}
                </p>
                <p className="text-gray-500 text-xs">
                  {(rebateRate * 100).toFixed(0)}% {t("activity.of_net_loss") || "of net loss"}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-5 h-5 rounded-full bg-purple-100 flex items-center justify-center mt-0.5">
                <CheckCircle size={12} className="text-purple-600" />
              </div>
              <div>
                <p className="text-gray-700 text-sm font-medium">
                  {t("activity.min_loss_threshold") || "Minimum Loss Threshold"}
                </p>
                <p className="text-gray-500 text-xs">
                  ${minLossThreshold}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-5 h-5 rounded-full bg-purple-100 flex items-center justify-center mt-0.5">
                <CheckCircle size={12} className="text-purple-600" />
              </div>
              <div>
                <p className="text-gray-700 text-sm font-medium">
                  {t("activity.activity_period") || "Activity Period"}
                </p>
                <p className="text-gray-500 text-xs">
                  {new Date(startTime).toLocaleDateString()} - {new Date(endTime).toLocaleDateString()}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-5 h-5 rounded-full bg-purple-100 flex items-center justify-center mt-0.5">
                <Info size={12} className="text-purple-600" />
              </div>
              <div>
                <p className="text-gray-700 text-sm font-medium">
                  {t("activity.notes") || "Notes"}
                </p>
                <p className="text-gray-500 text-xs">
                  {t("activity.loss_rebate_note") || "Rebate is calculated daily and can be claimed the next day."}
                </p>
              </div>
            </div>
          </div>

          {/* Close button */}
          <button
            onClick={onClose}
            className="w-full mt-6 bg-gradient-to-r from-pink-500 to-purple-500 hover:from-pink-400 hover:to-purple-400 text-white font-semibold py-3.5 rounded-xl transition-colors shadow-lg"
          >
            {t("activity.got_it") || "Got it"}
          </button>
        </div>
      </div>
    </div>
    </Portal>
  );
};
