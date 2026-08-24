"use client"

import React from 'react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'
import { useLanguage } from '../app/contexts/LanguageContext'
import { useAddDesktopStatus } from '../hooks/useAddDesktopStatus'
import { useNewUserRechargeStatus } from '../hooks/useNewUserRechargeStatus'

interface PromoNavigationCardConfig {
  href: string
  image: string
  title: string
  tag?: string
  loading?: boolean
  statusText?: string
}

const ActivityPage: React.FC = () => {
  const { t, language } = useLanguage()
  const router = useRouter()
  const isIndonesian = language === 'id'
  const { status: addDesktopActivityStatus, loading: loadingAddDesktop } = useAddDesktopStatus(true)
  const { status: newUserRechargeStatus, loading: loadingNewUserRecharge } = useNewUserRechargeStatus(true)

  const openActivityPage = (href: string) => {
    router.push(href)
  }

  const renderPromoNavigationCard = ({ href, image, title, tag, loading, statusText }: PromoNavigationCardConfig) => {
    return (
      <button
        type="button"
        onClick={() => openActivityPage(href)}
        className="group relative min-h-[360px] overflow-hidden rounded-3xl border border-white/15 bg-white/5 shadow-2xl transition-all hover:-translate-y-1 hover:border-white/25"
      >
        <Image
          src={image}
          alt={title}
          fill
          sizes="(max-width: 768px) 100vw, 50vw"
          className="absolute inset-0 h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent" />

        {tag && (
          <span className="absolute left-4 top-4 rounded-full bg-red-500/90 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white">
            {tag}
          </span>
        )}

        {loading && (
          <div className="absolute right-4 top-4 rounded-full bg-black/35 p-2 text-white">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        )}

        <div className="relative flex h-full min-h-[360px] flex-col justify-end p-5 text-left">
          <div className="max-w-xs">
            <p className="text-xs font-medium uppercase tracking-[0.24em] text-white/70">{t('nav.activity')}</p>
            <h3 className="mt-2 text-2xl font-bold text-white drop-shadow-lg">{title}</h3>
            {statusText && (
              <p className="mt-2 inline-flex rounded-full bg-white/15 px-3 py-1 text-xs font-medium text-white/85 backdrop-blur-sm">
                {statusText}
              </p>
            )}
          </div>

          <div className="mt-4 inline-flex w-fit items-center rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-transform group-hover:translate-x-1">
            {t('activity.details')}
          </div>
        </div>
      </button>
    )
  }

  const showAddDesktopCard = loadingAddDesktop || !addDesktopActivityStatus?.claimed
  const showNewUserRechargeCard = loadingNewUserRecharge || !newUserRechargeStatus?.hidden
  const promoNavigationCards: PromoNavigationCardConfig[] = [
    ...(showAddDesktopCard
      ? [{
        href: '/adddesktop',
        image: isIndonesian ? '/images/activity6/act_addDesktop_yinni.png' : '/images/activity6/act_addDesktop_yingyu.png',
        title: t('nav.adddesktop'),
        tag: 'HOT',
        loading: loadingAddDesktop,
      }]
      : []),
    ...(showNewUserRechargeCard
      ? [{
        href: '/newUserRecharge',
        image: isIndonesian ? '/images/activity1/act_yinni.jpg' : '/images/activity1/act_yingyu.jpg',
        title: t('nav.newUserRecharge'),
        tag: 'HOT',
        loading: loadingNewUserRecharge,
      }]
      : []),
    {
      href: '/sevenDayTopup',
      image: isIndonesian ? '/images/activity5/act_yinni.jpg' : '/images/activity5/act_yingyu.jpg',
      title: t('nav.sevenDayTopup'),
      tag: 'HOT',
    },
    {
      href: '/weeklySpinWheel',
      image: isIndonesian ? '/images/activity2/act_yinni.jpg' : '/images/activity2/act_yingyu.jpg',
      title: t('nav.weeklySpinWheel'),
      tag: 'NEW',
    },
    {
      href: '/dailyWeeklyChallenge',
      image: isIndonesian ? '/images/activity4/act_yinni.jpg' : '/images/activity4/act_yingyu.jpg',
      title: t('nav.dailyWeeklyChallenge'),
      tag: 'HOT',
    },
    {
      href: '/bettingRank',
      image: isIndonesian ? '/images/activity3/act_yinni.png' : '/images/activity3/act_yingyu.png',
      title: t('nav.bettingRank'),
      tag: 'NEW',
    },
  ]

  return (
    <div className="mx-auto max-w-7xl animate-fade-in px-4 pb-24 pt-16 sm:px-6 lg:px-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold text-white sm:text-4xl">{t('activity.title')}</h2>
          <p className="mt-1 text-sm text-gray-400">{t('activity.subtitle')}</p>
        </div>
        {/* <button className="rounded-full bg-white/10 px-3 py-1.5 text-xs text-gray-300 transition-colors hover:bg-white/20">
          {t('activity.history')}
        </button> */}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {promoNavigationCards.map((card) => (
          <React.Fragment key={card.href}>
            {renderPromoNavigationCard(card)}
          </React.Fragment>
        ))}
      </div>
    </div>
  )
}

export default ActivityPage
