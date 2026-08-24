"use client"

import { useEffect, useState } from "react"
import { apiUrl } from "@/lib/api-base-url"

const DEFAULT_SUPPORT_LINK = "https://wa.me/message/G4CV6NJKBOKAM1"

export function useSupportLink() {
  const [supportLink, setSupportLink] = useState(DEFAULT_SUPPORT_LINK)

  useEffect(() => {
    let cancelled = false

    const loadSupportLink = async () => {
      try {
        const res = await fetch(apiUrl("/getAppcfgs"), {
          cache: "no-store",
        })
        if (!res.ok) {
          return
        }

        const data = await res.json()
        const nextLink = data?.data?.whatsapp_url
        if (!cancelled && typeof nextLink === "string" && nextLink.trim()) {
          setSupportLink(nextLink.trim())
        }
      } catch (error) {
        console.error("[useSupportLink] Failed to load support link:", error)
      }
    }

    void loadSupportLink()

    return () => {
      cancelled = true
    }
  }, [])

  return supportLink
}
