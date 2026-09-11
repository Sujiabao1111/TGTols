"use client"
import React, { createContext, useContext, useState, ReactNode, useEffect, useMemo } from 'react';
import { Language } from '../mocks/types';
import { translations } from '../mocks/translations';

interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: (key: string) => string;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

const DEFAULT_LANGUAGE: Language = "en"
const BROWSER_LANGUAGE_KEY = "language"
const MANUAL_LANGUAGE_KEY = "language-manually-selected"
const SUPPORTED_LANGUAGES: Language[] = ["en", "ru"]

type TelegramWebApp = {
  initData?: string
  initDataUnsafe?: {
    user?: {
      language_code?: string
    }
  }
}

function normalizeLanguage(lang?: string | null): Language {
  if (lang === "ru") return "ru"
  if (lang === "en") return "en"
  return DEFAULT_LANGUAGE
}

function getTelegramLanguage(): Language {
  if (typeof window === "undefined") return DEFAULT_LANGUAGE

  const webApp = (window as typeof window & {
    Telegram?: { WebApp?: TelegramWebApp }
  }).Telegram?.WebApp

  let languageCode = webApp?.initDataUnsafe?.user?.language_code

  if (!languageCode && webApp?.initData) {
    try {
      const user = JSON.parse(new URLSearchParams(webApp.initData).get("user") || "{}") as {
        language_code?: string
      }
      languageCode = user.language_code
    } catch {
      languageCode = undefined
    }
  }

  return languageCode?.toLowerCase().trim().startsWith("ru") ? "ru" : "en"
}

function getStoredLanguage(): Language | null {
  if (typeof window === "undefined") {
    return null
  }

  if (localStorage.getItem(MANUAL_LANGUAGE_KEY) !== "true") return null

  const storedLang = localStorage.getItem(BROWSER_LANGUAGE_KEY) as Language | null
  if (storedLang && SUPPORTED_LANGUAGES.includes(storedLang)) {
    return normalizeLanguage(storedLang)
  }

  return null
}

function getInitialLanguage(): Language {
  return getStoredLanguage() || getTelegramLanguage()
}


export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Language>(getInitialLanguage)

  const setLanguage = (lang: Language) => {
    const normalizedLanguage = normalizeLanguage(lang)
    localStorage.setItem(BROWSER_LANGUAGE_KEY, normalizedLanguage)
    localStorage.setItem(MANUAL_LANGUAGE_KEY, "true")
    setLanguageState(normalizedLanguage)
  }

  const t = useMemo(() => {
    return (key: string): string => {
      const langTranslations = translations[language]
      return langTranslations?.[key] || translations.en[key] || key
    }
  }, [language])

  useEffect(() => {
    document.documentElement.lang = language
  }, [language])

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  )
}

export function useLanguage() {
  const context = useContext(LanguageContext)
  if (context === undefined) {
    throw new Error("useLanguage must be used within a LanguageProvider")
  }
  return context
}
