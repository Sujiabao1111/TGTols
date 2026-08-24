import type React from "react"
import { useEffect, useRef } from "react"
import { ArrowUp } from "lucide-react"

const ScrollToTop: React.FC = () => {
  const buttonRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    const toggleVisibility = () => {
      const button = buttonRef.current
      if (!button) return

      const isVisible = window.scrollY > 300
      button.classList.toggle("opacity-100", isVisible)
      button.classList.toggle("translate-y-0", isVisible)
      button.classList.toggle("opacity-0", !isVisible)
      button.classList.toggle("translate-y-10", !isVisible)
      button.classList.toggle("pointer-events-none", !isVisible)
      button.setAttribute("aria-hidden", String(!isVisible))
    }

    toggleVisibility()

    window.addEventListener("scroll", toggleVisibility)
    return () => window.removeEventListener("scroll", toggleVisibility)
  }, [])

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: "smooth",
    })
  }

  return (
    <button
      ref={buttonRef}
      onClick={scrollToTop}
      className="pointer-events-none fixed bottom-24 right-4 z-[70] translate-y-10 rounded-full bg-gradient-to-r from-lucky-gold to-yellow-500 p-3 text-black opacity-0 shadow-2xl transition-all duration-300 hover:scale-110 hover:shadow-lucky-gold/50 lg:bottom-8 lg:right-8"
      aria-label="Scroll to top"
      aria-hidden="true"
    >
      <ArrowUp size={24} />
    </button>
  )
}

export default ScrollToTop
