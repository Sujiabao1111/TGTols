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
const SUPPORTED_LANGUAGES: Language[] = ["en", "id"]
const ID_DEFAULT_HOSTNAMES = new Set(["ppnetpp.com", "www.ppnetpp.com"])
const EN_DEFAULT_HOSTNAMES = new Set([
  "ppnetpp.net",
  "www.ppnetpp.net",
  "ppbetpp.tech",
  "www.ppbetpp.tech",
  "ppnet11.com",
  "www.ppnet11.com",
  "ppnet22.com",
  "www.ppnet22.com",
  "ppnet33.com",
  "www.ppnet33.com",
  "ppnet44.com",
  "www.ppnet44.com",
  "ppnet55.com",
  "www.ppnet55.com",
])

function normalizeLanguage(lang?: string | null): Language {
  return lang === "id" ? "id" : DEFAULT_LANGUAGE
}

function getBrowserLanguage(): Language {
  if (typeof navigator === "undefined") {
    return DEFAULT_LANGUAGE
  }

  const rawLang = navigator.languages?.[0] || navigator.language || DEFAULT_LANGUAGE
  const normalizedLang = rawLang.toLowerCase().trim()

  if (normalizedLang === "id" || normalizedLang.startsWith("id-")) {
    return "id"
  }

  return DEFAULT_LANGUAGE
}

function getHostnameDefaultLanguage(): Language | null {
  if (typeof window === "undefined") {
    return null
  }

  const hostname = window.location.hostname.toLowerCase()
  if (ID_DEFAULT_HOSTNAMES.has(hostname)) {
    return "id"
  }
  if (EN_DEFAULT_HOSTNAMES.has(hostname)) {
    return "en"
  }

  return null
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
  return getHostnameDefaultLanguage() || getStoredLanguage() || getBrowserLanguage()
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
