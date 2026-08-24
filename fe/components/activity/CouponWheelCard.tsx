"use client";

import React, { useState, useEffect } from "react";
import { Ticket, Gift } from "lucide-react";
import { useLanguage } from "../../app/contexts/LanguageContext";

interface CurrentCoupon {
  coupon_code: string;
  coupon_value: number;
  min_deposit_to_activate: number;
  activated: boolean;
  activated_at?: string;
  expires_at: string;
}

interface CouponWheelCardProps {
  canSpin: boolean;
  currentCoupon?: CurrentCoupon;
  todayCoupon?: {
    code: string;
    value: number;
    status: string;
    valid_until: string;
  };
  onOpenWheel: () => void;
  onOpenCouponInput: () => void;
}

export const CouponWheelCard: React.FC<CouponWheelCardProps> = ({
  canSpin,
  currentCoupon,
  todayCoupon,
  onOpenWheel,
  onOpenCouponInput,
}) => {
  const { t } = useLanguage();
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    checkMobile();
    window.addEventListener("resize", checkMobile);
    return () => window.removeEventListener("resize", checkMobile);
  }, []);

  const backgroundImage = isMobile
    ? "/images/activity-2-mobile.webp"
    : "/images/activity-2-desktop.webp";

  return (
    <div className="relative rounded-3xl overflow-hidden shadow-2xl border border-white/15 min-h-[360px]">
      {/* Background image - responsive */}
      <div
        className="absolute inset-0 bg-cover bg-center"
        style={{
          backgroundImage: `url('${backgroundImage}')`,
        }}
      />

      {/* Gradient overlay for text readability */}
      <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent" />

      {/* Content */}
      <div className="relative h-full flex flex-col min-h-[360px]">
        {/* Header - positioned at top */}
        <div className="p-6">
          <h3 className="text-2xl sm:text-3xl font-bold text-white drop-shadow-lg">
            {t("activity.wheel_title")}
          </h3>
          <p className="text-white/90 text-sm mt-1 max-w-[200px] drop-shadow-md">
            {t("activity.wheel_desc")}
          </p>
        </div>

        {/* Spacer to push buttons to bottom */}
        <div className="flex-1" />

        {/* Bottom buttons */}
        <div className="p-6 flex gap-3">
          <button
            onClick={onOpenCouponInput}
            className="flex-1 bg-white/20 hover:bg-white/30 backdrop-blur-sm text-white text-sm font-semibold py-3 px-2 rounded-xl transition-colors flex items-center justify-center gap-1.5 whitespace-nowrap"
          >
            <Ticket size={16} />
            <span className="truncate">{t("activity.input_coupon")}</span>
          </button>
          <button
            onClick={onOpenWheel}
            className="flex-1 bg-white text-pink-600 text-sm font-semibold py-3 px-2 rounded-xl hover:bg-white/90 transition-colors flex items-center justify-center gap-1.5 shadow-lg whitespace-nowrap"
          >
            <Gift size={16} />
            <span className="truncate">{t("activity.open_wheel")}</span>
          </button>
        </div>
      </div>
    </div>
  );
};
