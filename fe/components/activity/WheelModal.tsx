"use client";

import React, { useState, useEffect } from "react";
import { X, Gift, Sparkles, Copy, CheckCircle2 } from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";
import { Portal } from "../ui/portal";

interface WheelReward {
  value: number;
  label: string;
  color: string;
}

interface SpinResult {
  coupon_code: string;
  value: number;
  label: string;
  min_deposit: number;
  valid_until: string;
}

interface CurrentCoupon {
  coupon_code: string;
  coupon_value: number;
  min_deposit_to_activate: number;
  activated: boolean;
  expires_at: string;
}

interface WheelModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSpin: () => Promise<SpinResult | null>;
  spinning: boolean;
  result: SpinResult | null;
  currentCoupon?: CurrentCoupon | null;
  canSpin: boolean;
  rewards?: WheelReward[];
}

const DEFAULT_REWARDS: WheelReward[] = [
  { value: 1, label: "$1", color: "#FF6B6B" },
  { value: 2, label: "$2", color: "#4ECDC4" },
  { value: 5, label: "$5", color: "#45B7D1" },
  { value: 10, label: "$10", color: "#96CEB4" },
  { value: 50, label: "$50", color: "#FECA57" },
  { value: 100, label: "$100", color: "#FF9FF3" },
];

export const WheelModal: React.FC<WheelModalProps> = ({
  isOpen,
  onClose,
  onSpin,
  spinning,
  result,
  currentCoupon,
  canSpin,
  rewards = DEFAULT_REWARDS,
}) => {
  const { t } = useLanguage();
  const [rotation, setRotation] = useState(0);
  const [showResult, setShowResult] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (result) {
      setShowResult(true);
    }
  }, [result]);

  useEffect(() => {
    if (!isOpen) {
      setRotation(0);
      setShowResult(false);
      setCopied(false);
    }
  }, [isOpen]);

  const handleSpin = async () => {
    if (spinning || !canSpin) return;

    setShowResult(false);
    // Random rotation between 5-10 full spins (1800-3600 degrees) plus random offset
    const spins = 1800 + Math.random() * 1800;
    const newRotation = rotation + spins;
    setRotation(newRotation);

    await onSpin();
  };

  const handleCopyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  // Show today's coupon if already spun and no new result
  const hasAlreadySpun = !canSpin && currentCoupon && !result;

  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="relative bg-gradient-to-br from-gray-900 to-gray-800 w-full sm:max-w-md sm:w-full border border-white/10 shadow-2xl rounded-t-3xl sm:rounded-3xl max-h-[85vh] overflow-y-auto sm:!mb-0 p-6" style={{ marginBottom: '60px' }}>
        {/* Close button */}
        <button
          onClick={onClose}
          className="absolute right-4 top-4 text-white/60 hover:text-white transition-colors z-10"
        >
          <X size={24} />
        </button>

        {/* Header */}
        <div className="text-center mb-6">
          <h3 className="text-2xl font-bold text-white">
            {t("activity.wheel_title")}
          </h3>
          <p className="text-white/60 text-sm mt-1">
            {hasAlreadySpun
              ? t("activity.your_coupon_today") || "Your coupon for today"
              : t("activity.spin_to_win")}
          </p>
        </div>

        {/* Already spun - show today's coupon */}
        {hasAlreadySpun && (
          <div className="text-center py-4">
            <div className="bg-gradient-to-br from-yellow-400/20 to-orange-500/20 rounded-3xl p-6 border border-yellow-400/30 mb-6">
              <div className="text-5xl mb-4">🎁</div>
              <p className="text-white/80 text-sm mb-2">
                {t("activity.you_won") || "You won"}
              </p>
              <p className="text-white font-bold text-3xl mb-4">
                ${currentCoupon.coupon_value}
              </p>

              {/* Coupon code with copy */}
              <div className="bg-white/10 rounded-xl p-4">
                <p className="text-white/60 text-xs mb-2">
                  {t("activity.promo_code")}
                </p>
                <div className="flex items-center gap-2 justify-center">
                  <code className="text-white font-mono text-lg bg-black/20 px-3 py-1.5 rounded-lg">
                    {currentCoupon.coupon_code}
                  </code>
                  <button
                    onClick={() => handleCopyCode(currentCoupon.coupon_code)}
                    className="p-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
                    title={t("activity.copy") || "Copy"}
                  >
                    {copied ? (
                      <CheckCircle2 size={18} className="text-green-400" />
                    ) : (
                      <Copy size={18} className="text-white" />
                    )}
                  </button>
                </div>
              </div>

              {/* Valid until */}
              <p className="text-white/50 text-xs mt-4">
                {t("activity.valid_until") || "Valid until"}: {" "}
                {new Date(currentCoupon.expires_at).toLocaleDateString()}
              </p>
            </div>

            <button
              onClick={onClose}
              className="w-full bg-white text-gray-900 font-semibold py-3 rounded-xl hover:bg-white/90 transition-colors"
            >
              {t("activity.close")}
            </button>
          </div>
        )}

        {/* Result display after spin */}
        {showResult && result && (
          <div className="text-center py-4">
            <div className="bg-gradient-to-br from-yellow-400/20 to-orange-500/20 rounded-3xl p-6 border border-yellow-400/30 mb-6">
              <div className="text-5xl mb-4">🎉</div>
              <p className="text-white/80 text-sm mb-2">
                {t("activity.congratulations")}
              </p>
              <p className="text-white font-bold text-3xl mb-4">
                {result.label}
              </p>

              {/* Coupon code with copy */}
              <div className="bg-white/10 rounded-xl p-4">
                <p className="text-white/60 text-xs mb-2">
                  {t("activity.promo_code")}
                </p>
                <div className="flex items-center gap-2 justify-center">
                  <code className="text-white font-mono text-lg bg-black/20 px-3 py-1.5 rounded-lg">
                    {result.coupon_code}
                  </code>
                  <button
                    onClick={() => handleCopyCode(result.coupon_code)}
                    className="p-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
                    title={t("activity.copy") || "Copy"}
                  >
                    {copied ? (
                      <CheckCircle2 size={18} className="text-green-400" />
                    ) : (
                      <Copy size={18} className="text-white" />
                    )}
                  </button>
                </div>
              </div>
            </div>

            <button
              onClick={onClose}
              className="w-full bg-white text-gray-900 font-semibold py-3 rounded-xl hover:bg-white/90 transition-colors"
            >
              {t("activity.claim_prize")}
            </button>
          </div>
        )}

        {/* Spinning wheel - only show if can spin and no result */}
        {canSpin && !showResult && (
          <>
            {/* Wheel */}
            <div className="relative flex justify-center items-center py-4">
              {/* Outer ring */}
              <div className="absolute w-64 h-64 rounded-full border-4 border-white/10" />

              {/* Spinning wheel */}
              <div
                className="relative w-56 h-56 rounded-full overflow-hidden transition-transform duration-[5000ms] ease-out"
                style={{
                  transform: `rotate(${rotation}deg)`,
                }}
              >
                {rewards.map((reward, index) => {
                  const angle = (360 / rewards.length) * index;
                  return (
                    <div
                      key={index}
                      className="absolute w-full h-full"
                      style={{
                        transform: `rotate(${angle}deg)`,
                        clipPath: `polygon(50% 50%, 50% 0%, ${
                          50 + 50 * Math.cos((2 * Math.PI) / rewards.length)
                        }% ${50 - 50 * Math.sin((2 * Math.PI) / rewards.length)}%)`,
                        backgroundColor: reward.color,
                      }}
                    >
                      <span
                        className="absolute text-white font-bold text-sm"
                        style={{
                          top: "20%",
                          left: "50%",
                          transform: `translateX(-50%) rotate(${
                            360 / rewards.length / 2
                          }deg)`,
                        }}
                      >
                        {reward.label}
                      </span>
                    </div>
                  );
                })}
              </div>

              {/* Center circle */}
              <div className="absolute w-12 h-12 bg-white rounded-full flex items-center justify-center shadow-lg">
                <Sparkles size={20} className="text-pink-500" />
              </div>

              {/* Pointer */}
              <div className="absolute -top-2 left-1/2 -translate-x-1/2">
                <div className="w-0 h-0 border-l-[12px] border-r-[12px] border-t-[20px] border-l-transparent border-r-transparent border-t-yellow-400" />
              </div>
            </div>

            {/* Spin button */}
            <button
              onClick={handleSpin}
              disabled={spinning}
              className={`w-full mt-6 py-4 rounded-xl font-semibold text-lg flex items-center justify-center gap-2 transition-colors ${
                spinning
                  ? "bg-white/30 text-white/70 cursor-not-allowed"
                  : "bg-gradient-to-r from-pink-500 to-purple-500 text-white hover:from-pink-400 hover:to-purple-400 shadow-lg"
              }`}
            >
              {spinning ? (
                <>
                  <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  {t("activity.spinning")}...
                </>
              ) : (
                <>
                  <Gift size={20} />
                  {t("activity.spin_now")}
                </>
              )}
            </button>
          </>
        )}
      </div>
    </div>
    </Portal>
  );
};
