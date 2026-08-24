# Next.js App Router 路由组 Layout 重复渲染问题

## 问题描述

在 Activity 页面刷新时，出现两个相同的 `min-h-screen bg-lucky-dark` 包裹层，导致页面内容重复渲染。

## 现象

- 第一次登录正常
- 登录成功后在其他页面刷新正常
- **在 Activity 页面刷新时，会多出一个 Activity Page 块**
- 刷新后其他页面内容也带有这个 Activity Page

## 根本原因

Next.js App Router 的**路由组（Route Groups）机制**导致 `(public)` 和 `(protected)` 两个 layout 都被渲染。

### 路由组结构

```
app/
├── (public)/
│   ├── layout.tsx      # 总是渲染 Layout
│   └── home/
│       └── page.tsx
├── (protected)/
│   ├── layout.tsx      # 也渲染 Layout
│   └── activity/
│       └── page.tsx
```

### 问题分析

当访问 `/activity` 时：

1. **URL 匹配**：`/activity` 匹配 `(protected)/activity/page.tsx`
2. **Layout 继承**：Next.js 会查找并渲染所有匹配的 layout
3. **认证状态变化**：登录后 `isLoggedIn` 从 `false` → `true`
4. **重复渲染**：
   - `(public)/layout.tsx` 没有路径检查，总是返回 `Layout` 组件
   - `(protected)/layout.tsx` 也返回 `Layout` 组件
   - 两个 Layout 都被挂载到 DOM 中

### 问题代码（修复前）

```tsx
// (public)/layout.tsx - 问题：没有路径检查
export default function PublicLayout({ children }) {
  // ...hooks
  return (
    <Layout {...props}>
      {children}  {/* 这里也会渲染 activity 内容 */}
    </Layout>
  )
}
```

## 解决方案

### 1. Public Layout 添加路径保护

```tsx
"use client"

import { usePathname } from "next/navigation"
import { useEffect, useState } from "react"

// 需要登录才能访问的路由前缀
const PROTECTED_PATHS = ['/activity', '/wallet', '/profile', '/favorites', '/game/', '/invite']

function isProtectedPath(path: string): boolean {
  return PROTECTED_PATHS.some(prefix => path.startsWith(prefix))
}

export default function PublicLayout({ children }) {
  const pathname = usePathname()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  // 如果当前路径是 protected 路由，不渲染 Layout
  if (isProtectedPath(pathname)) {
    return null
  }

  // 避免 SSR/水合不匹配
  if (!mounted) {
    return <div className="min-h-screen bg-lucky-dark" />
  }

  return (
    <Layout {...props}>
      {children}
    </Layout>
  )
}
```

### 2. Protected Layout 优化挂载

```tsx
"use client"

import { useEffect, useState } from "react"

export default function ProtectedLayout({ children }) {
  const [mounted, setMounted] = useState(false)
  const { isLoggedIn, checkAuth } = useAuth()
  // ...其他 hooks

  useEffect(() => {
    setMounted(true)
    // 认证检查...
  }, [isLoggedIn, checkAuth, router])

  // 等待客户端挂载
  if (!mounted) {
    return <div className="min-h-screen bg-lucky-dark" />
  }

  if (!isLoggedIn) {
    return null
  }

  return (
    <Layout {...props}>
      {children}
    </Layout>
  )
}
```

## 关键修复点

| 修复点 | 说明 |
|--------|------|
| `usePathname()` | 获取当前路径，判断是否在 protected 路由 |
| `isProtectedPath()` | 检查路径是否需要登录 |
| 返回 `null` | Public layout 在 protected 路径上不渲染 |
| `mounted` 状态 | 等待客户端挂载，避免 SSR/水合不匹配 |

## 修复效果

**之前**：
- `/activity` → `(public)/layout.tsx` 渲染 Layout → `(protected)/layout.tsx` 渲染 Layout → 两个 Layout 嵌套

**之后**：
- `/activity` → `(public)/layout.tsx` 返回 `null` → 只有 `(protected)/layout.tsx` 渲染 Layout → 单个 Layout

## 最佳实践

### 1. 路由组命名规范

使用有意义的命名：
- `(public)` - 公开访问
- `(protected)` - 需要认证
- `(marketing)` - 营销页面

### 2. Layout 互斥处理

当使用路由组时，确保 layout 之间有明确的互斥逻辑：

```tsx
// 在公共 layout 中检查是否应由自己处理
if (shouldBeHandledByOtherLayout(pathname)) {
  return null
}
```

### 3. 避免嵌套 Layout

如果两个路由组都包裹同一个 Layout 组件，确保只有一个会被渲染。

### 4. 客户端挂载检查

使用 `useEffect` + `useState` 确保组件在客户端挂载后再渲染：

```tsx
const [mounted, setMounted] = useState(false)

useEffect(() => {
  setMounted(true)
}, [])

if (!mounted) {
  return <LoadingScreen />
}
```

## 相关文件

- `fe/app/(public)/layout.tsx`
- `fe/app/(protected)/layout.tsx`
- `fe/components/Layout.tsx`

## 参考

- [Next.js Route Groups](https://nextjs.org/docs/app/building-your-application/routing/route-groups)
- [Next.js Layouts](https://nextjs.org/docs/app/building-your-application/routing/pages-and-layouts)
