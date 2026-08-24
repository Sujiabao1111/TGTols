# 前端组件设计文档 - 活动功能

## 概述

本文档描述活动功能相关的前端组件设计，基于Next.js + React + TypeScript + Tailwind CSS实现。

## 现有结构回顾

当前活动页面结构：
```
fe/
├── app/(protected)/activity/
│   └── page.tsx              # 活动页面路由
├── components/
│   └── ActivityPage.tsx      # 活动页面主组件
└── services/
    └── api.ts                # API服务
```

## 新增组件结构

```
fe/
├── app/(protected)/activity/
│   └── page.tsx                    # 页面路由（复用现有）
├── components/
│   ├── ActivityPage.tsx            # 活动页面主组件（改造）
│   ├── activity/
│   │   ├── DailyRechargeCard.tsx   # 储值返利活动卡片
│   │   ├── CouponWheelCard.tsx     # 轮盘活动卡片
│   │   ├── LossRebateCard.tsx      # 输返活动卡片（新增）
│   │   ├── WheelModal.tsx          # 轮盘抽奖弹窗
│   │   ├── CouponList.tsx          # 优惠券列表
│   │   ├── ActivityProgress.tsx    # 活动进度组件
│   │   └── RewardClaimModal.tsx    # 奖励领取弹窗
│   └── ui/                         # 通用UI组件（如有需要）
├── hooks/
│   └── useActivity.ts              # 活动相关自定义Hook
├── services/
│   ├── api.ts                      # 现有API服务
│   └── activityApi.ts              # 活动API服务（新增）
└── types/
    └── activity.ts                 # 活动类型定义（新增）
```

## 类型定义

### types/activity.ts

```typescript
// 活动类型
export type ActivityType = 'recharge_rebate' | 'coupon_wheel' | 'loss_rebate';

// 活动状态
export type ActivityStatus = 0 | 1; // 0:禁用 1:启用

// 用户活动进度状态
export type ProgressStatus = 0 | 1 | 2 | 3;
// 0:未开始 1:进行中 2:已完成 3:已领取

// 优惠券状态
export type CouponStatus = 0 | 1 | 2 | 3;
// 0:未激活 1:已激活 2:已使用 3:已过期

// 活动基础信息
export interface Activity {
  id: number;
  type: ActivityType;
  name: string;
  description?: string;
  config: ActivityConfig;
  start_time: string;
  end_time: string;
  status: ActivityStatus;
  user_progress?: UserActivityProgress;
}

// 活动配置
export interface ActivityConfig {
  min_deposit: number;
  currency: string;
  repeatable?: boolean;      // 是否可重复进行
  cycle_days?: number;       // 活动周期天数（如4天储值返利）
  days?: DayRewardConfig[];
  daily_limit?: number;
}

// 每日返利配置
export interface DayRewardConfig {
  day: number;
  rate: number;
  label: string;
}

// 用户活动进度
export interface UserActivityProgress {
  id: number;
  user_id: number;
  activity_type: ActivityType;
  day_number: number;
  progress_data: ProgressData;
  status: ProgressStatus;
  progress_date: string;
  created_at: string;
  updated_at: string;
}

// 储值返利进度数据（可重复活动）
export interface DailyRechargeProgressData {
  activity_id: number;       // 活动实例ID，支持多活动并行
  day_number: number;        // 当前第几天（1-4）
  deposit_amount: number;
  reward_amount: number;
  reward_rate: number;
  min_deposit: number;
  claimed: boolean;
  activity_start: string;    // 活动开始时间，用于区分不同周期
}

// 轮盘进度数据
export interface WheelProgressData {
  coupon_code: string;
  coupon_value: number;
  min_deposit_to_activate: number;
  spun_at: string;
  activated: boolean;
  activated_at?: string;
}

// 输返活动进度数据
export interface LossRebateProgressData {
  activity_id: number;
  bet_amount: number;
  win_amount: number;
  net_loss: number;
  rebate_rate: number;
  rebate_amount: number;
  min_loss_threshold: number;
  claimed: boolean;
  calculated_at: string;
  claimed_at?: string;
}

export type ProgressData = DailyRechargeProgressData | WheelProgressData | LossRebateProgressData;

// 优惠券
export interface Coupon {
  id: number;
  coupon_code: string;
  coupon_value: number;
  min_deposit: number;
  status: CouponStatus;
  status_text: string;
  valid_start: string;
  valid_end: string;
  activated_at?: string;
  used_at?: string;
}

// 轮盘状态
export interface WheelStatus {
  can_spin: boolean;
  last_spin_at?: string;
  next_spin_at?: string;
  today_coupon?: {
    code: string;
    value: number;
    status: string;
    valid_until: string;
  };
}

// 抽奖结果
export interface SpinResult {
  coupon_code: string;
  coupon_value: number;
  min_deposit: number;
  valid_until: string;
  wheel_position: number;
  result_label: string;
}

// 领取奖励响应
export interface ClaimRewardResponse {
  reward_amount: number;
  new_balance: number;
  transaction_id: string;
}

// 输返活动统计
export interface LossRebateStats {
  activity_id: number;
  activity_name: string;
  start_time: string;
  end_time: string;
  rebate_rate: number;
  min_loss_threshold: number;
  today_stats: {
    date: string;
    bet_amount: number;
    win_amount: number;
    net_loss: number;
    rebate_amount: number;
    claimed: boolean;
  };
  total_stats: {
    total_bet: number;
    total_win: number;
    total_loss: number;
    total_rebate: number;
    claimed_rebate: number;
    pending_rebate: number;
  };
  history: {
    date: string;
    bet_amount: number;
    win_amount: number;
    net_loss: number;
    rebate_amount: number;
    claimed: boolean;
    claimed_at?: string;
  }[];
}
```

## API服务

### services/activityApi.ts

```typescript
import { fetchWithAuth, API_BASE_URL } from './api';
import type {
  Activity,
  Coupon,
  WheelStatus,
  SpinResult,
  ClaimRewardResponse,
} from '@/types/activity';

export const activityApi = {
  // 获取活动列表
  async getActivities(): Promise<{ code: number; data: Activity[] }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/activities`);
    return response.json();
  },

  // 获取活动进度
  async getActivityProgress(type: string) {
    const response = await fetchWithAuth(`${API_BASE_URL}/activities/${type}/progress`);
    return response.json();
  },

  // 领取活动奖励
  async claimReward(type: string, date: string, dayNumber: number) {
    const response = await fetchWithAuth(`${API_BASE_URL}/activities/${type}/claim`, {
      method: 'POST',
      body: JSON.stringify({ date, day_number: dayNumber }),
    });
    return response.json();
  },

  // 轮盘抽奖
  async spinWheel(): Promise<{ code: number; data: SpinResult }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/wheel/spin`, {
      method: 'POST',
    });
    return response.json();
  },

  // 获取轮盘状态
  async getWheelStatus(): Promise<{ code: number; data: WheelStatus }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/wheel/status`);
    return response.json();
  },

  // 获取用户优惠券列表
  async getCoupons(status?: number, page = 1, pageSize = 10) {
    const params = new URLSearchParams();
    if (status !== undefined) params.append('status', status.toString());
    params.append('page', page.toString());
    params.append('page_size', pageSize.toString());

    const response = await fetchWithAuth(`${API_BASE_URL}/coupons?${params}`);
    return response.json();
  },

  // 激活优惠券
  async activateCoupon(code: string) {
    const response = await fetchWithAuth(`${API_BASE_URL}/coupons/${code}/activate`, {
      method: 'POST',
    });
    return response.json();
  },

  // 使用优惠券
  async useCoupon(code: string): Promise<{ code: number; data: ClaimRewardResponse }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/coupons/${code}/use`, {
      method: 'POST',
    });
    return response.json();
  },

  // 获取输返活动统计
  async getLossRebateStats(): Promise<{ code: number; data: LossRebateStats }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/activities/loss_rebate/stats`);
    return response.json();
  },
};
```

## 自定义Hook

### hooks/useActivity.ts

```typescript
import { useState, useEffect, useCallback } from 'react';
import { activityApi } from '@/services/activityApi';
import type { Activity, Coupon, WheelStatus } from '@/types/activity';

export function useActivities() {
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchActivities = useCallback(async () => {
    try {
      setLoading(true);
      const res = await activityApi.getActivities();
      if (res.code === 0) {
        setActivities(res.data);
      }
    } catch (err) {
      setError('获取活动列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchActivities();
  }, [fetchActivities]);

  return { activities, loading, error, refetch: fetchActivities };
}

export function useWheelStatus() {
  const [status, setStatus] = useState<WheelStatus | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchStatus = useCallback(async () => {
    try {
      const res = await activityApi.getWheelStatus();
      if (res.code === 0) {
        setStatus(res.data);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  return { status, loading, refetch: fetchStatus };
}

export function useCoupons(status?: number) {
  const [coupons, setCoupons] = useState<Coupon[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchCoupons = useCallback(async () => {
    try {
      const res = await activityApi.getCoupons(status);
      if (res.code === 0) {
        setCoupons(res.data.list);
      }
    } finally {
      setLoading(false);
    }
  }, [status]);

  useEffect(() => {
    fetchCoupons();
  }, [fetchCoupons]);

  return { coupons, loading, refetch: fetchCoupons };
}

export function useLossRebateStats() {
  const [stats, setStats] = useState<LossRebateStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await activityApi.getLossRebateStats();
      if (res.code === 0) {
        setStats(res.data);
      } else if (res.code === 404) {
        setStats(null);
      }
    } catch (err) {
      setError('获取输返活动统计失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  return { stats, loading, error, refetch: fetchStats };
}
```

## 组件详细设计

### 1. DailyRechargeCard - 储值返利活动卡片

**功能描述：**
展示储值返利活动，包含4天进度条和当前状态。

**Props：**
```typescript
interface DailyRechargeCardProps {
  activity: Activity;
  onClaim: (day: number) => void;
}
```

**UI结构：**
```
┌─────────────────────────────────────────────┐
│  [渐变背景卡片]                                │
│  ┌───────────────────────────────────────┐  │
│  │  储值返利活动标题                        │  │
│  │  连续4天充值返利活动说明                  │  │
│  └───────────────────────────────────────┘  │
│                                             │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐        │
│  │Day 1│  │Day 2│  │Day 3│  │Day 4│        │
│  │ +50%│  │ +75%│  │+100%│  │+150%│        │
│  └─────┘  └─────┘  └─────┘  └─────┘        │
│    ○────────○────────○────────○            │
│   已完成   进行中    锁定     锁定          │
│                                             │
│  ┌───────────────────────────────────────┐  │
│  │         [查看详情]  [去充值]            │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**代码实现：**
```typescript
// components/activity/DailyRechargeCard.tsx
'use client';

import React from 'react';
import { Gift, ChevronRight } from 'lucide-react';
import type { Activity, DayRewardConfig } from '@/types/activity';

interface DailyRechargeCardProps {
  activity: Activity;
  onClaim?: (day: number) => void;
}

export const DailyRechargeCard: React.FC<DailyRechargeCardProps> = ({
  activity,
  onClaim,
}) => {
  const days = activity.config.days || [];
  const currentDay = activity.user_progress?.day_number || 1;
  const progressData = activity.user_progress?.progress_data as {
    deposit_amount?: number;
    claimed?: boolean;
  };

  return (
    <div className="relative rounded-3xl overflow-hidden bg-gradient-to-br from-fuchsia-500/90 via-violet-500/90 to-blue-500/90 p-6 shadow-2xl border border-white/15 min-h-[360px]">
      {/* 背景光效 */}
      <div className="absolute inset-0 opacity-30">
        <div className="absolute -top-10 -right-6 h-36 w-36 rounded-full bg-white/20 blur-2xl" />
        <div className="absolute -bottom-10 -left-6 h-40 w-40 rounded-full bg-black/20 blur-2xl" />
      </div>

      <div className="relative flex h-full flex-col">
        <h3 className="text-2xl sm:text-3xl font-bold text-white mb-2">
          {activity.name}
        </h3>
        <p className="text-white/80 text-sm max-w-sm">
          {activity.description || '连续4天充值，享受递增返利'}
        </p>

        {/* 4天进度展示 */}
        <div className="mt-6">
          <div className="grid grid-cols-4 gap-3 text-center">
            {days.map((day: DayRewardConfig) => (
              <div key={day.day} className="space-y-3">
                <div className="bg-white/20 rounded-full py-1 text-xs text-white/90">
                  Day {day.day}
                </div>
                <div
                  className={`h-16 rounded-2xl flex items-center justify-center ${
                    day.day < currentDay
                      ? 'bg-green-500/30 border border-green-400/50'
                      : day.day === currentDay
                      ? 'bg-white/30 border border-white/50 animate-pulse'
                      : 'bg-white/15 border border-white/20'
                  }`}
                >
                  <Gift
                    size={28}
                    className={`${
                      day.day <= currentDay ? 'text-white' : 'text-white/40'
                    }`}
                  />
                </div>
              </div>
            ))}
          </div>

          {/* 进度指示器 */}
          <div className="mt-5 flex items-center">
            {days.map((day: DayRewardConfig, index: number) => (
              <div key={`dot-${day.day}`} className="flex items-center flex-1">
                <div
                  className={`h-3 w-3 rounded-full ${
                    day.day < currentDay
                      ? 'bg-green-400'
                      : day.day === currentDay
                      ? 'bg-white scale-125'
                      : 'bg-white/30'
                  }`}
                />
                {index < days.length - 1 && (
                  <div
                    className={`h-1 flex-1 mx-2 rounded-full ${
                      day.day < currentDay ? 'bg-green-400' : 'bg-white/30'
                    }`}
                  />
                )}
              </div>
            ))}
          </div>

          {/* 返利比例 */}
          <div className="mt-4 grid grid-cols-4 gap-3 text-center text-[11px] text-white/90">
            {days.map((day: DayRewardConfig) => (
              <div key={`rate-${day.day}`} className="bg-white/10 rounded-xl p-2">
                {day.label}
              </div>
            ))}
          </div>
        </div>

        {/* 今日状态 */}
        <div className="mt-4 p-3 bg-white/10 rounded-xl">
          <div className="flex justify-between items-center text-white text-sm">
            <span>今日充值: ${progressData?.deposit_amount || 0}</span>
            <span>最低要求: ${activity.config.min_deposit}</span>
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="mt-6 flex gap-3">
          <button className="flex-1 bg-white/20 text-white font-semibold py-3 rounded-xl hover:bg-white/30 transition">
            查看详情
          </button>
          <button
            onClick={() => onClaim?.(currentDay)}
            disabled={progressData?.claimed}
            className="flex-1 bg-white text-lucky-dark font-semibold py-3 rounded-xl hover:bg-white/90 transition disabled:opacity-50"
          >
            {progressData?.claimed ? '已领取' : '去充值'}
          </button>
        </div>
      </div>
    </div>
  );
};
```

### 2. CouponWheelCard - 轮盘活动卡片

**功能描述：**
展示轮盘抽奖活动入口和今日抽奖状态。

**Props：**
```typescript
interface CouponWheelCardProps {
  activity: Activity;
  wheelStatus: WheelStatus;
  onSpin: () => void;
}
```

**UI结构：**
```
┌─────────────────────────────────────────────┐
│  [渐变背景卡片 - 粉色/紫色渐变]               │
│  ┌───────────────────────────────────────┐  │
│  │  🎡 每日幸运轮盘                       │  │
│  │  每日抽奖赢取超值优惠券                 │  │
│  └───────────────────────────────────────┘  │
│                                             │
│        ┌───────────────┐                    │
│        │   [轮盘图示]   │                    │
│        │    或奖品     │                    │
│        │    展示区     │                    │
│        └───────────────┘                    │
│                                             │
│  ┌───────────────────────────────────────┐  │
│  │  🎁 今日优惠券: $10 (未激活)           │  │
│  │  充值 $3 即可激活                      │  │
│  └───────────────────────────────────────┘  │
│                                             │
│  ┌───────────────────────────────────────┐  │
│  │     [立即抽奖] 或 [已抽奖]             │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### 3. WheelModal - 轮盘抽奖弹窗

**功能描述：**
轮盘抽奖动画和结果展示弹窗。

**Props：**
```typescript
interface WheelModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSpin: () => Promise<SpinResult>;
  rewards: { value: number; label: string; color: string }[];
}
```

**状态流程：**
1. 初始状态 - 显示轮盘，可点击抽奖
2. 旋转中 - 轮盘旋转动画
3. 结果展示 - 显示中奖结果

### 4. LossRebateCard - 输返活动卡片

**功能描述：**
展示输返活动的实时统计，包括下注、输赢、可返还金额等信息。

**Props：**
```typescript
interface LossRebateCardProps {
  stats: LossRebateStats;
  onClaim: () => void;
}
```

**UI结构：**
```
┌─────────────────────────────────────────────┐
│  [渐变背景卡片 - 金色/橙色渐变]               │
│  ┌───────────────────────────────────────┐  │
│  │  💰 亏损返还活动                       │  │
│  │  活动期间净亏损按6%返还                 │  │
│  │  活动时间: 02/01 - 02/14               │  │
│  └───────────────────────────────────────┘  │
│                                             │
│  ┌───────────────────────────────────────┐  │
│  │           今日统计                      │  │
│  │  ┌──────────┬──────────┬──────────┐   │  │
│  │  │ 下注     │ 派彩     │ 亏损     │   │  │
│  │  │ $500     │ $300     │ $200     │   │  │
│  │  └──────────┴──────────┴──────────┘   │  │
│  │                                       │  │
│  │  可返还金额: $12 (6%)                  │  │
│  │  [领取返还] 或 [未达到门槛]            │  │
│  └───────────────────────────────────────┘  │
│                                             │
│  ┌───────────────────────────────────────┐  │
│  │           活动累计                      │  │
│  │  总下注: $1500                          │  │
│  │  总亏损: $300                           │  │
│  │  已返还: $6                             │  │
│  │  待返还: $12                            │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**代码实现：**
```typescript
// components/activity/LossRebateCard.tsx
'use client';

import React from 'react';
import { TrendingDown, TrendingUp, DollarSign, Calendar } from 'lucide-react';
import type { LossRebateStats } from '@/types/activity';

interface LossRebateCardProps {
  stats: LossRebateStats;
  onClaim?: () => void;
}

export const LossRebateCard: React.FC<LossRebateCardProps> = ({
  stats,
  onClaim,
}) => {
  const { today_stats, total_stats, rebate_rate, min_loss_threshold } = stats;

  const canClaim = today_stats.rebate_amount > 0 && !today_stats.claimed;
  const belowThreshold = today_stats.net_loss > 0 && today_stats.net_loss < min_loss_threshold;

  return (
    <div className="relative rounded-3xl overflow-hidden bg-gradient-to-br from-amber-500/90 via-orange-500/90 to-red-500/90 p-6 shadow-2xl border border-white/15 min-h-[360px]">
      {/* 背景光效 */}
      <div className="absolute inset-0 opacity-30">
        <div className="absolute -top-10 -right-6 h-36 w-36 rounded-full bg-yellow-300/20 blur-2xl" />
        <div className="absolute -bottom-10 -left-6 h-40 w-40 rounded-full bg-red-300/20 blur-2xl" />
      </div>

      <div className="relative flex h-full flex-col">
        {/* 标题区域 */}
        <div className="flex justify-between items-start">
          <div>
            <h3 className="text-2xl sm:text-3xl font-bold text-white mb-2">
              {stats.activity_name}
            </h3>
            <p className="text-white/80 text-sm">
              活动期间净亏损按 {(rebate_rate * 100).toFixed(0)}% 返还
            </p>
            <div className="flex items-center gap-1 text-white/60 text-xs mt-1">
              <Calendar size={12} />
              <span>
                {new Date(stats.start_time).toLocaleDateString()} - {new Date(stats.end_time).toLocaleDateString()}
              </span>
            </div>
          </div>
          <div className="bg-white/20 text-white text-xs px-2 py-1 rounded-full">
            返还比例 {(rebate_rate * 100).toFixed(0)}%
          </div>
        </div>

        {/* 今日统计 */}
        <div className="mt-6 p-4 bg-white/10 rounded-2xl">
          <div className="text-white/80 text-sm font-medium mb-3">今日统计</div>
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 text-white/60 text-xs mb-1">
                <TrendingUp size={12} />
                下注
              </div>
              <div className="text-white font-bold">${today_stats.bet_amount.toFixed(2)}</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 text-white/60 text-xs mb-1">
                <DollarSign size={12} />
                派彩
              </div>
              <div className="text-white font-bold">${today_stats.win_amount.toFixed(2)}</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 text-white/60 text-xs mb-1">
                <TrendingDown size={12} />
                亏损
              </div>
              <div className={`font-bold ${today_stats.net_loss > 0 ? 'text-red-300' : 'text-white'}`}>
                ${today_stats.net_loss.toFixed(2)}
              </div>
            </div>
          </div>

          {/* 可返还金额 */}
          {today_stats.net_loss > 0 && (
            <div className="mt-4 pt-4 border-t border-white/20">
              <div className="flex justify-between items-center">
                <span className="text-white/80 text-sm">可返还金额</span>
                <span className="text-yellow-300 font-bold text-xl">
                  ${today_stats.rebate_amount.toFixed(2)}
                </span>
              </div>
              {belowThreshold && (
                <div className="text-white/50 text-xs mt-1">
                  亏损未达到 ${min_loss_threshold} 返还门槛
                </div>
              )}
            </div>
          )}

          {/* 领取按钮 */}
          <button
            onClick={onClaim}
            disabled={!canClaim}
            className={`w-full mt-4 py-3 rounded-xl font-semibold transition ${
              canClaim
                ? 'bg-white text-orange-600 hover:bg-white/90'
                : 'bg-white/20 text-white/50 cursor-not-allowed'
            }`}
          >
            {today_stats.claimed
              ? '今日已领取'
              : belowThreshold
              ? '未达到门槛'
              : today_stats.net_loss <= 0
              ? '今日无亏损'
              : `领取 $${today_stats.rebate_amount.toFixed(2)}`}
          </button>
        </div>

        {/* 活动累计统计 */}
        <div className="mt-4 p-4 bg-white/5 rounded-2xl">
          <div className="text-white/60 text-xs mb-2">活动累计</div>
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="flex justify-between">
              <span className="text-white/60">总下注</span>
              <span className="text-white">${total_stats.total_bet.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-white/60">总亏损</span>
              <span className="text-red-300">${total_stats.total_loss.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-white/60">已返还</span>
              <span className="text-green-300">${total_stats.claimed_rebate.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-white/60">待返还</span>
              <span className="text-yellow-300">${total_stats.pending_rebate.toFixed(2)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
```

### 5. ActivityPage 改造

**改造后的主页面：**

```typescript
// components/ActivityPage.tsx
'use client';

import React, { useState } from 'react';
import { Gift, Clock, Sparkles, Ticket } from 'lucide-react';
import { useLanguage } from '../app/contexts/LanguageContext';
import { useActivities, useWheelStatus } from '@/hooks/useActivity';
import { DailyRechargeCard } from './activity/DailyRechargeCard';
import { CouponWheelCard } from './activity/CouponWheelCard';
import { WheelModal } from './activity/WheelModal';
import { CouponList } from './activity/CouponList';
import { activityApi } from '@/services/activityApi';
import type { Activity, SpinResult } from '@/types/activity';

const ActivityPage: React.FC = () => {
  const { t } = useLanguage();
  const { activities, loading, refetch } = useActivities();
  const { status: wheelStatus, refetch: refetchWheel } = useWheelStatus();

  const [wheelOpen, setWheelOpen] = useState(false);
  const [spinning, setSpinning] = useState(false);
  const [spinResult, setSpinResult] = useState<SpinResult | null>(null);

  // 获取特定类型的活动
  const rechargeActivity = activities.find(
    (a) => a.type === 'daily_recharge'
  );
  const wheelActivity = activities.find((a) => a.type === 'coupon_wheel');

  // 处理轮盘抽奖
  const handleSpin = async () => {
    setSpinning(true);
    try {
      const result = await activityApi.spinWheel();
      if (result.code === 0) {
        setSpinResult(result.data);
        await refetchWheel();
      }
    } finally {
      setSpinning(false);
    }
  };

  // 处理奖励领取
  const handleClaimReward = async (day: number) => {
    // 实现领取逻辑
  };

  if (loading) {
    return (
      <div className="pt-16 pb-24 px-4 flex justify-center">
        <div className="animate-spin h-8 w-8 border-2 border-white/30 border-t-white rounded-full" />
      </div>
    );
  }

  return (
    <div className="pt-16 pb-24 px-4 sm:px-6 lg:px-8 max-w-6xl mx-auto animate-fade-in">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-3xl sm:text-4xl font-display font-bold text-white">
            {t('activity.title')}
          </h2>
          <p className="text-sm text-gray-400 mt-1">
            Seasonal promos & weekly perks
          </p>
        </div>
        <button className="text-xs bg-white/10 px-3 py-1 rounded-full text-gray-300">
          {t('activity.history')}
        </button>
      </div>

      {/* 活动分类Tab */}
      <div className="inline-flex items-center gap-2 mb-8 rounded-full bg-white/10 p-1">
        <button className="px-4 py-2 rounded-full bg-white/20 text-white text-sm font-semibold">
          All bonuses
        </button>
        <span className="h-5 w-px bg-white/20"></span>
        <button className="px-4 py-2 rounded-full text-gray-300 text-sm font-semibold hover:bg-white/10">
          My Coupons
        </button>
      </div>

      {/* 活动卡片网格 */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* 储值返利活动 */}
        {rechargeActivity && (
          <DailyRechargeCard
            activity={rechargeActivity}
            onClaim={handleClaimReward}
          />
        )}

        {/* 轮盘活动 */}
        {wheelActivity && wheelStatus && (
          <CouponWheelCard
            activity={wheelActivity}
            wheelStatus={wheelStatus}
            onSpin={() => setWheelOpen(true)}
          />
        )}

        {/* 其他活动占位 - 例如每周返利 */}
        <div className="relative rounded-3xl overflow-hidden bg-gradient-to-br from-purple-600/90 via-fuchsia-600/90 to-sky-500/90 p-6 shadow-2xl border border-white/15 min-h-[360px]">
          <div className="absolute inset-0 opacity-20">
            <div className="absolute -bottom-8 right-4 h-44 w-44 rounded-full bg-white/20 blur-2xl" />
          </div>
          <div className="relative h-full flex flex-col justify-between">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="text-2xl font-bold text-white">Cashback every Friday</h3>
                <p className="text-white/80 text-sm mt-2">
                  {t('activity.cashback_desc')}
                </p>
              </div>
              <span className="bg-white/20 text-white text-xs px-2 py-1 rounded-full">
                Weekly
              </span>
            </div>
            <div className="mt-6">
              <div className="text-white/90 text-sm font-semibold">Updating soon</div>
              <div className="text-white/70 text-xs">We look forward to seeing you again</div>
              <div className="flex items-center gap-2 text-white/90 text-sm mt-3 bg-white/15 px-3 py-1 rounded-full w-fit">
                <Clock size={14} />
                03d:04h:16m:43s
              </div>
            </div>
            <button className="mt-4 bg-white text-lucky-dark font-semibold px-4 py-2 rounded-xl">
              {t('activity.check_status')}
            </button>
          </div>
        </div>
      </div>

      {/* 优惠券列表区域 */}
      <div className="mt-10">
        <h4 className="text-xl font-semibold text-white mb-4">My Coupons</h4>
        <CouponList />
      </div>

      {/* 轮盘抽奖弹窗 */}
      <WheelModal
        isOpen={wheelOpen}
        onClose={() => setWheelOpen(false)}
        onSpin={handleSpin}
        spinning={spinning}
        result={spinResult}
        rewards={[
          { value: 1, label: '$1', color: '#FF6B6B' },
          { value: 2, label: '$2', color: '#4ECDC4' },
          { value: 5, label: '$5', color: '#45B7D1' },
          { value: 10, label: '$10', color: '#96CEB4' },
          { value: 50, label: '$50', color: '#FECA57' },
          { value: 100, label: '$100', color: '#FF9FF3' },
        ]}
      />
    </div>
  );
};

export default ActivityPage;
```

## 页面布局说明

### 响应式设计

- **桌面端 (lg+)**: 3列网格布局
- **平板端 (md)**: 2列网格布局
- **移动端**: 单列堆叠布局

### 交互状态

| 组件 | 状态 | 样式 |
|------|------|------|
| DailyRechargeCard | 已完成 | 绿色边框/图标 |
| DailyRechargeCard | 进行中 | 脉冲动画/白色高亮 |
| DailyRechargeCard | 锁定 | 灰度/降低透明度 |
| CouponWheelCard | 可抽奖 | 渐变动画按钮 |
| CouponWheelCard | 已抽奖 | 禁用状态/倒计时 |

## 图片资源清单（design/resource/）

### 已提供的资源

| 图片名 | 用途 | 使用位置 |
|--------|------|----------|
| `box_available.png` | 可领奖盒子状态 | 活动1进度展示 |
| `box_opened.png` | 已领奖盒子状态 | 活动1进度展示 |
| `box_unavailable.png` | 未解锁盒子状态 | 活动1进度展示 |
| `gameconsole.png` | 活动1详情图1 | 1-activity-1-dialog.png |
| `depositepig.png` | 活动1详情图2 | 1-activity-1-dialog.png |
| `creditcard.png` | 活动1详情图3 | 1-activity-1-dialog.png |
| `activity-2-desktop.webp` | 活动2 PC背景 | CouponWheelCard组件 |
| `activity-2-mobile.webp` | 活动2 移动端背景 | CouponWheelCard组件 |
| `activity-3-desktop.webp` | 活动3 PC背景 | LossRebateCard组件 |
| `activity-3-mobile.webp` | 活动3 移动端背景 | LossRebateCard组件 |

### 待设计资源

以下UI元素仍需设计人员提供：

1. **轮盘图片** - 抽奖轮盘的视觉设计图
2. **活动1卡片背景** - 储值返利活动卡片背景（PC/移动端）
3. **图标资源** - 活动类型专属图标
4. **中奖动画** - 抽奖结果展示的动画效果
5. **优惠券样式** - 不同面值优惠券的视觉设计

### 图片使用示例

```typescript
// 响应式背景图使用
const getBackgroundImage = (activityType: string, isMobile: boolean) => {
  const images: Record<string, string> = {
    'recharge_rebate': isMobile ? '/images/activity-1-mobile.webp' : '/images/activity-1-desktop.webp',
    'coupon_wheel': isMobile ? '/images/activity-2-mobile.webp' : '/images/activity-2-desktop.webp',
    'loss_rebate': isMobile ? '/images/activity-3-mobile.webp' : '/images/activity-3-desktop.webp',
  };
  return images[activityType];
};

// 盒子状态图片
const getBoxImage = (status: 'available' | 'opened' | 'unavailable') => {
  const images = {
    'available': '/images/box_available.png',
    'opened': '/images/box_opened.png',
    'unavailable': '/images/box_unavailable.png',
  };
  return images[status];
};
```
