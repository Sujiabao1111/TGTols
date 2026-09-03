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

const DEFAULT_LANGUAGE: Language = "ru"
const BROWSER_LANGUAGE_KEY = "language"
const SUPPORTED_LANGUAGES: Language[] = ["en", "ru"]

function normalizeLanguage(lang?: string | null): Language {
  if (lang === "ru") return "ru"
  if (lang === "en") return "en"
  return DEFAULT_LANGUAGE
}

function getBrowserLanguage(): Language {
  if (typeof navigator === "undefined") {
    return DEFAULT_LANGUAGE
  }

  const rawLang = navigator.languages?.[0] || navigator.language || DEFAULT_LANGUAGE
  const normalizedLang = rawLang.toLowerCase().trim()

  if (normalizedLang === "ru" || normalizedLang.startsWith("ru-")) return "ru"
  return DEFAULT_LANGUAGE
}

function getStoredLanguage(): Language | null {
  if (typeof window === "undefined") {
    return null
  }

  const storedLang = localStorage.getItem(BROWSER_LANGUAGE_KEY) as Language | null
  if (storedLang && SUPPORTED_LANGUAGES.includes(storedLang)) {
    return normalizeLanguage(storedLang)
  }

  return null
}

function getInitialLanguage(): Language {
  return getStoredLanguage() || getBrowserLanguage()
}


export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Language>(getInitialLanguage)

  const setLanguage = (lang: Language) => {
    setLanguageState(normalizeLanguage(lang))
  }

  const t = useMemo(() => {
    return (key: string): string => {
      const langTranslations = translations[language]
      return langTranslations?.[key] || translations.en[key] || key
    }
  }, [language])

  useEffect(() => {
    localStorage.setItem(BROWSER_LANGUAGE_KEY, language)
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
