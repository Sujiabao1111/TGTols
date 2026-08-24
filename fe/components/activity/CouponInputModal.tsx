"use client";

import React, { useState } from "react";
import { X, Copy, CheckCircle2, Wallet } from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";
import { Portal } from "../ui/portal";

type InputState = "input" | "failed" | "success";

interface CouponInputModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (code: string) => Promise<{ success: boolean; message?: string; extraBalance?: number }>;
  onNavigateToWallet: () => void;
}

export const CouponInputModal: React.FC<CouponInputModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  onNavigateToWallet,
}) => {
  const { t } = useLanguage();
  const [state, setState] = useState<InputState>("input");
  const [code, setCode] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [extraBalance, setExtraBalance] = useState(0);

  if (!isOpen) return null;

  const handleSubmit = async () => {
    if (!code.trim()) return;
    setSubmitting(true);
    try {
      const result = await onSubmit(code.trim());
      if (result.success) {
        setState("success");
        setExtraBalance(result.extraBalance || 0);
      } else {
        setState("failed");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleTryAgain = () => {
    setState("input");
    setCode("");
  };

  const handleClose = () => {
    setState("input");
    setCode("");
    setExtraBalance(0);
    onClose();
  };

  const getHeaderImage = () => {
    switch (state) {
      case "input":
        return "/images/promocode-active.webp";
      case "failed":
        return "/images/error-cashierv1.webp";
      case "success":
        return "/images/promocode.webp";
    }
  };

  const getTitle = () => {
    switch (state) {
      case "input":
        return t("activity.activate_promo_code") || "Activate the promo code";
      case "failed":
        return t("activity.invalid_code") || "Invalid Code";
      case "success":
        return t("activity.code_activated") || "Code Activated!";
    }
  };

  const getSubtitle = () => {
    switch (state) {
      case "input":
        return t("activity.promo_input_desc") || "Earn promo codes by being active in TonPlay and community groups";
      case "failed":
        return t("activity.try_again_desc") || "The code you entered is invalid or expired";
      case "success":
        return t("activity.deposit_bonus_desc") || "Deposit now and get extra balance!";
    }
  };

  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
        <div className="relative bg-white sm:rounded-3xl w-full sm:max-w-sm sm:w-full shadow-2xl overflow-hidden rounded-t-3xl sm:rounded-3xl max-h-[85vh] overflow-y-auto sm:!mb-0" style={{ marginBottom: '60px' }}>
          {/* Header image with close button */}
          <div className="relative h-48 bg-gradient-to-br from-pink-400 via-fuchsia-500 to-purple-600">
            {/* Close button - positioned on top of header image */}
            <button
              onClick={handleClose}
              className="absolute right-2 top-2 z-20 h-10 w-10 rounded-full bg-black/50 hover:bg-black/60 flex items-center justify-center text-white transition-colors border border-white/30 shadow-lg"
            >
              <X size={20} />
            </button>
            <img
              src={getHeaderImage()}
              alt="Promo"
              className="w-full h-full object-cover"
            />
          </div>

          {/* Content */}
          <div className="p-6 text-center">
            {/* Title and subtitle */}
            <h3 className="text-xl font-bold text-gray-900 mb-1">
              {getTitle()}
            </h3>
            <p className="text-gray-500 text-sm mb-6">
              {getSubtitle()}
            </p>

            {/* State-specific content */}
            {state === "input" && (
              <>
                {/* Input field */}
                <div className="relative mb-4">
                  <input
                    type="text"
                    value={code}
                    onChange={(e) => setCode(e.target.value.toUpperCase())}
                    placeholder={t("activity.enter_promo_code") || "Enter the promo code"}
                    className="w-full bg-gray-100 border border-gray-200 rounded-xl py-4 px-4 pr-12 text-gray-900 placeholder:text-gray-400 focus:outline-none focus:border-pink-400 text-center font-semibold tracking-wider uppercase"
                    onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
                    disabled={submitting}
                  />
                <div className="absolute right-4 top-1/2 -translate-y-1/2">
                  <img
                    src="/images/ton-icon.svg"
                    alt="TON"
                    className="w-6 h-6 opacity-50"
                    onError={(e) => {
                      // Fallback if icon doesn't exist
                      e.currentTarget.style.display = 'none';
                    }}
                  />
                </div>
              </div>

              {/* Submit button */}
              <button
                onClick={handleSubmit}
                disabled={!code.trim() || submitting}
                className="w-full bg-gradient-to-r from-pink-500 to-purple-500 hover:from-pink-400 hover:to-purple-400 disabled:from-gray-300 disabled:to-gray-300 text-white font-semibold py-3.5 rounded-xl transition-colors shadow-lg mb-3"
              >
                {submitting ? (
                  <span className="flex items-center justify-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    {t("activity.processing")}
                  </span>
                ) : (
                  t("activity.activate") || "Activate"
                )}
              </button>

              {/* Close button */}
              <button
                onClick={handleClose}
                className="w-full bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold py-3.5 rounded-xl transition-colors"
              >
                {t("activity.close") || "Close"}
              </button>
            </>
          )}

          {state === "failed" && (
            <>
              {/* Error message */}
              <div className="mb-6 p-4 bg-red-50 rounded-xl border border-red-100">
                <p className="text-red-600 text-sm">
                  {t("activity.invalid_code_msg") || "This promo code is invalid or has already been used."}
                </p>
              </div>

              {/* Try again button */}
              <button
                onClick={handleTryAgain}
                className="w-full bg-gradient-to-r from-pink-500 to-purple-500 hover:from-pink-400 hover:to-purple-400 text-white font-semibold py-3.5 rounded-xl transition-colors shadow-lg mb-3"
              >
                {t("activity.try_again") || "Try Again"}
              </button>

              {/* Close button */}
              <button
                onClick={handleClose}
                className="w-full bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold py-3.5 rounded-xl transition-colors"
              >
                {t("activity.close") || "Close"}
              </button>
            </>
          )}

          {state === "success" && (
            <>
              {/* Extra balance info */}
              <div className="mb-6">
                <div className="text-gray-500 text-sm mb-2">
                  {t("activity.extra_balance_on_deposit") || "Extra balance on deposit"}
                </div>
                <div className="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-pink-500 to-purple-500">
                  +${extraBalance.toFixed(2)}
                </div>
              </div>

              {/* Deposit button */}
              <button
                onClick={onNavigateToWallet}
                className="w-full bg-gradient-to-r from-pink-500 to-purple-500 hover:from-pink-400 hover:to-purple-400 text-white font-semibold py-3.5 rounded-xl transition-colors shadow-lg flex items-center justify-center gap-2 mb-3"
              >
                <Wallet size={18} />
                {t("activity.deposit_now") || "Deposit Now"}
              </button>

              {/* Close button */}
              <button
                onClick={handleClose}
                className="w-full bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold py-3.5 rounded-xl transition-colors"
              >
                {t("activity.close") || "Close"}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  </Portal>
  );
};
