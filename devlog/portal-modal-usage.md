# React Portal 与弹框组件使用指南

## 概述

本项目使用 React Portal 技术实现弹框（Modal）组件，确保弹框能够正确渲染在页面最上层，不受父组件样式限制。

## 什么是 Portal

React Portal 允许将子组件渲染到父组件 DOM 树之外的 DOM 节点中。

### 传统方式的问题

```jsx
// 传统方式：Modal 渲染在组件树内部
<div className="activity-page">
  <div className="modal">...</div>  {/* 受父组件样式影响 */}
</div>
```

**问题：**
- 受父级 `z-index` 限制
- 可能被父组件的 `overflow: hidden` 裁剪
- 移动端底部导航栏（通常 `z-index: 50`）会遮挡弹框

### Portal 解决方案

```jsx
// Portal 方式：Modal 渲染到 document.body
<div className="activity-page">...</div>
<!-- Modal 渲染在这里，与 activity-page 同级 -->
<div className="modal z-[100]">...</div>
```

## Portal 组件实现

### 基础 Portal 组件

```tsx
// fe/components/ui/portal.tsx
"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";

interface PortalProps {
  children: React.ReactNode;
}

export function Portal({ children }: PortalProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  if (!mounted) return null;

  return createPortal(children, document.body);
}
```

### 关键说明

| 特性 | 说明 |
|------|------|
| `mounted` 状态 | 防止服务端渲染（SSR）不匹配 |
| `createPortal` | React 提供的 API，将内容挂载到指定容器 |
| `document.body` | 目标容器，确保在所有其他元素之上 |

## 弹框组件实现

### 基础结构

```tsx
"use client";

import { Portal } from "../ui/portal";

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
}

export function Modal({ isOpen, onClose, children }: ModalProps) {
  if (!isOpen) return null;

  return (
    <Portal>
      {/* 遮罩层 */}
      <div
        className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm"
        onClick={onClose}
      >
        {/* 弹框内容 */}
        <div
          className="relative bg-white w-full sm:max-w-sm rounded-t-3xl sm:rounded-3xl max-h-[85vh] overflow-y-auto"
          style={{ marginBottom: '60px' }}
          onClick={e => e.stopPropagation()}
        >
          {children}
        </div>
      </div>
    </Portal>
  );
}
```

### 样式说明

| 类名 | 作用 |
|------|------|
| `fixed inset-0` | 全屏定位 |
| `z-[100]` | 高于导航栏（通常 z-50） |
| `items-end sm:items-center` | 移动端底部固定，桌面端垂直居中 |
| `marginBottom: '60px'` | 为底部导航栏留出空间 |
| `max-h-[85vh]` | 限制最大高度，避免超出屏幕 |
| `overflow-y-auto` | 内容过多时可滚动 |

## 项目中的弹框组件

### 1. CouponInputModal - 优惠券输入

```tsx
// fe/components/activity/CouponInputModal.tsx
"use client";

import { Portal } from "../ui/portal";
import { X, Wallet } from "lucide-react";

type InputState = "input" | "failed" | "success";

interface CouponInputModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (code: string) => Promise<{ success: boolean; message?: string; extraBalance?: number }>;
  onNavigateToWallet: () => void;
}

export const CouponInputModal: React.FC<CouponInputModalProps> = ({
  isOpen, onClose, onSubmit, onNavigateToWallet
}) => {
  const [state, setState] = useState<InputState>("input");
  const [code, setCode] = useState("");
  // ...

  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
        <div className="relative bg-white sm:rounded-3xl w-full sm:max-w-sm shadow-2xl overflow-hidden rounded-t-3xl max-h-[85vh] overflow-y-auto" style={{ marginBottom: '60px' }}>
          {/* Header image with close button */}
          <div className="relative h-48 bg-gradient-to-br from-pink-400 via-fuchsia-500 to-purple-600">
            <button onClick={handleClose} className="absolute right-2 top-2 z-20 h-10 w-10 rounded-full bg-black/50 hover:bg-black/60 flex items-center justify-center text-white">
              <X size={20} />
            </button>
            <img src="/images/promocode-active.webp" alt="Promo" className="w-full h-full object-cover" />
          </div>

          {/* Content */}
          <div className="p-6 text-center">
            {/* 三种状态：input / failed / success */}
            {state === "input" && <InputStateContent />}
            {state === "failed" && <FailedStateContent />}
            {state === "success" && <SuccessStateContent />}
          </div>
        </div>
      </div>
    </Portal>
  );
};
```

**特点：**
- 三种状态切换：输入中、验证失败、验证成功
- 头部图片 + 关闭按钮
- 优惠券码输入框（自动转大写）
- 成功后可跳转到钱包页面

### 2. WheelModal - 幸运转盘

```tsx
// fe/components/activity/WheelModal.tsx
"use client";

import { Portal } from "../ui/portal";
import { X, Gift, Sparkles, Copy, CheckCircle2 } from "lucide-react";

interface WheelModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSpin: () => Promise<SpinResult | null>;
  spinning: boolean;
  result: SpinResult | null;
  currentCoupon?: CurrentCoupon | null;
  canSpin: boolean;
}

export const WheelModal: React.FC<WheelModalProps> = ({
  isOpen, onClose, onSpin, spinning, result, currentCoupon, canSpin
}) => {
  const [rotation, setRotation] = useState(0);
  const [showResult, setShowResult] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleSpin = async () => {
    if (spinning || !canSpin) return;

    // 随机旋转 5-10 圈
    const spins = 1800 + Math.random() * 1800;
    setRotation(prev => prev + spins);
    setShowResult(false);

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

  // 今日已抽奖，显示优惠券
  const hasAlreadySpun = !canSpin && currentCoupon && !result;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
        <div className="relative bg-gradient-to-br from-gray-900 to-gray-800 w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl p-6" style={{ marginBottom: '60px' }}>
          <button onClick={onClose} className="absolute right-4 top-4 text-white/60 hover:text-white">
            <X size={24} />
          </button>

          {/* 已抽奖状态 - 显示优惠券 */}
          {hasAlreadySpun && (
            <div className="text-center py-4">
              <div className="bg-gradient-to-br from-yellow-400/20 to-orange-500/20 rounded-3xl p-6 border border-yellow-400/30">
                <div className="text-5xl mb-4">🎁</div>
                <p className="text-white/80">You won</p>
                <p className="text-white font-bold text-3xl">${currentCoupon.coupon_value}</p>

                {/* 优惠券码复制 */}
                <div className="bg-white/10 rounded-xl p-4 mt-4">
                  <div className="flex items-center gap-2 justify-center">
                    <code className="text-white font-mono text-lg">{currentCoupon.coupon_code}</code>
                    <button onClick={() => handleCopyCode(currentCoupon.coupon_code)}>
                      {copied ? <CheckCircle2 className="text-green-400" /> : <Copy className="text-white" />}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* 抽奖结果 */}
          {showResult && result && <ResultDisplay result={result} />}

          {/* 转盘 */}
          {canSpin && !showResult && (
            <>
              <div className="relative flex justify-center items-center py-4">
                <div className="absolute w-64 h-64 rounded-full border-4 border-white/10" />
                <div
                  className="relative w-56 h-56 rounded-full overflow-hidden transition-transform duration-[5000ms] ease-out"
                  style={{ transform: `rotate(${rotation}deg)` }}
                >
                  {/* 转盘扇区 */}
                  {rewards.map((reward, index) => (
                    <WheelSegment key={index} reward={reward} index={index} total={rewards.length} />
                  ))}
                </div>
                <div className="absolute w-12 h-12 bg-white rounded-full flex items-center justify-center">
                  <Sparkles className="text-pink-500" />
                </div>
              </div>
              <button onClick={handleSpin} disabled={spinning} className="w-full mt-6 py-4 rounded-xl bg-gradient-to-r from-pink-500 to-purple-500">
                {spinning ? "Spinning..." : "Spin Now"}
              </button>
            </>
          )}
        </div>
      </div>
    </Portal>
  );
};
```

**特点：**
- CSS 动画转盘，5秒旋转效果
- 三种显示状态：已抽奖/抽奖结果/转盘
- 优惠券码一键复制
- 每日限制检查

### 3. LossRebateTermsModal - 活动规则

```tsx
// fe/components/activity/LossRebateTermsModal.tsx
"use client";

import { Portal } from "../ui/portal";
import { X, Info, CheckCircle } from "lucide-react";

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
  isOpen, onClose, activityName, rebateRate, minLossThreshold, startTime, endTime
}) => {
  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
        <div className="relative bg-white w-full sm:max-w-sm rounded-t-3xl sm:rounded-3xl max-h-[85vh] overflow-y-auto" style={{ marginBottom: '60px' }}>
          <button onClick={onClose} className="absolute right-3 top-3 h-8 w-8 rounded-full bg-black/20 hover:bg-black/30 flex items-center justify-center text-white">
            <X size={18} />
          </button>

          {/* Header image */}
          <div className="relative h-48 bg-gradient-to-br from-pink-400 via-fuchsia-500 to-purple-600">
            <img src="/images/activity-3-mobile.webp" alt="Cashback" className="w-full h-full object-cover" />
          </div>

          {/* Content */}
          <div className="p-6">
            <h3 className="text-xl font-bold text-gray-900 mb-2">{activityName}</h3>

            {/* Terms list */}
            <div className="space-y-3">
              <TermItem
                icon={<CheckCircle size={12} className="text-purple-600" />}
                label="Rebate Rate"
                value={`${(rebateRate * 100).toFixed(0)}% of net loss`}
              />
              <TermItem
                icon={<CheckCircle size={12} className="text-purple-600" />}
                label="Minimum Loss Threshold"
                value={`$${minLossThreshold}`}
              />
              <TermItem
                icon={<Info size={12} className="text-purple-600" />}
                label="Notes"
                value="Rebate is calculated daily and can be claimed the next day."
              />
            </div>

            <button onClick={onClose} className="w-full mt-6 bg-gradient-to-r from-pink-500 to-purple-500 text-white font-semibold py-3.5 rounded-xl">
              Got it
            </button>
          </div>
        </div>
      </div>
    </Portal>
  );
};
```

**特点：**
- 活动规则展示
- 响应式图片头部
- 条款列表带图标
- 统一的关闭按钮样式

### 4. ActivityTermsModal - 充值返利规则

```tsx
// fe/components/activity/ActivityTermsModal.tsx
"use client";

import { Portal } from "../ui/portal";
import { X, AlertCircle, Gift } from "lucide-react";

interface ActivityTermsModalProps {
  isOpen: boolean;
  onClose: () => void;
  activityName: string;
  tiers: TierInfo[];
  minDeposit: number;
}

export const ActivityTermsModal: React.FC<ActivityTermsModalProps> = ({
  isOpen, onClose, activityName, tiers, minDeposit
}) => {
  if (!isOpen) return null;

  return (
    <Portal>
      <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
        <div className="relative bg-gradient-to-br from-fuchsia-600 via-violet-600 to-blue-600 w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl p-1" style={{ marginBottom: '60px' }}>
          <div className="bg-gray-900/95 rounded-[22px] p-6">
            {/* Header */}
            <div className="text-center mb-6">
              <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-gradient-to-br from-fuchsia-500 to-violet-500 mb-3">
                <Gift size={24} className="text-white" />
              </div>
              <h3 className="text-xl font-bold text-white">Bonus Terms +{tiers[0]?.rate * 100 || 50}%</h3>
            </div>

            {/* Terms content */}
            <div className="space-y-4 text-sm">
              <div className="flex justify-between text-white/80">
                <span>Min. deposit</span>
                <span className="text-white font-semibold">${minDeposit}</span>
              </div>

              {/* Bonus tiers */}
              <div className="bg-white/5 rounded-xl p-4">
                <h4 className="text-white font-semibold mb-3">Bonus Tiers</h4>
                {tiers.map((tier) => (
                  <div key={tier.day} className="flex justify-between text-sm">
                    <span className="text-white/70">{tier.label}</span>
                    <span className="text-white font-medium">+{tier.rate * 100}%</span>
                  </div>
                ))}
              </div>

              {/* Warning section */}
              <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-xl p-4">
                <div className="flex items-start gap-3">
                  <AlertCircle size={20} className="text-yellow-400 flex-shrink-0" />
                  <div>
                    <h4 className="text-yellow-400 font-semibold mb-2">Be Careful</h4>
                    <ul className="space-y-1 text-white/70 text-xs">
                      <li>• Available only for certain games</li>
                      <li>• When withdrawing funds, all uncancelled bonuses will be cancelled</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Portal>
  );
};
```

**特点：**
- 渐变边框效果
- 多级奖励展示
- 警告提示区域
- 暗色主题内容区

## 弹框使用最佳实践

### 1. 状态管理

```tsx
const [modalOpen, setModalOpen] = useState(false);

// 打开弹框
const openModal = () => setModalOpen(true);

// 关闭弹框
const closeModal = () => setModalOpen(false);

return (
  <>
    <button onClick={openModal}>Open Modal</button>
    <MyModal isOpen={modalOpen} onClose={closeModal} />
  </>
);
```

### 2. 关闭方式

支持多种关闭方式：
- 点击关闭按钮
- 点击遮罩层
- 按 ESC 键
- 操作完成后自动关闭

```tsx
// 点击遮罩层关闭
<div onClick={onClose} className="...">
  {/* 阻止事件冒泡，点击内容区不关闭 */}
  <div onClick={e => e.stopPropagation()}>
    {children}
  </div>
</div>
```

### 3. 动画效果

```tsx
// 使用 Tailwind 过渡
<div className="transition-all duration-300 ease-in-out">

// 旋转动画
<div className="transition-transform duration-[5000ms] ease-out">

// 淡入淡出
<div className="animate-fade-in">
```

### 4. 响应式设计

```tsx
// 移动端底部弹出，桌面端居中
className="flex items-end sm:items-center ..."

// 移动端全宽，桌面端限制宽度
className="w-full sm:max-w-sm ..."

// 移动端圆角只在顶部
className="rounded-t-3xl sm:rounded-3xl"
```

### 5. 滚动处理

```tsx
// 限制最大高度，允许内容滚动
className="max-h-[85vh] overflow-y-auto"
```

### 6. z-index 管理

```tsx
// 使用 z-[100] 确保在导航栏之上（导航栏通常是 z-50）
className="z-[100] ..."
```

## 相关文件

| 文件 | 说明 |
|------|------|
| `fe/components/ui/portal.tsx` | Portal 基础组件 |
| `fe/components/activity/CouponInputModal.tsx` | 优惠券输入弹框 |
| `fe/components/activity/WheelModal.tsx` | 幸运转盘弹框 |
| `fe/components/activity/LossRebateTermsModal.tsx` | 输返活动规则弹框 |
| `fe/components/activity/ActivityTermsModal.tsx` | 充值返利规则弹框 |

## 参考

- [React Portals](https://react.dev/reference/react-dom/createPortal)
- [Next.js Client Components](https://nextjs.org/docs/app/building-your-application/rendering/client-components)
- [Tailwind CSS Transitions](https://tailwindcss.com/docs/transition-property)
