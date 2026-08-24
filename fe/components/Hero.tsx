import type React from "react"
import { useState, useEffect } from "react"
import { ChevronRight, Sparkles, ChevronLeft, ExternalLink } from "lucide-react"
import type { HeroSlide } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { gameService } from "../services/api"
import { useNavigation } from "@/hooks/useNavigation"
import { useAddDesktopStatus } from "@/hooks/useAddDesktopStatus"
import type { NewUserRechargeStatus } from "@/hooks/useNewUserRechargeStatus"

interface HeroProps {
  newUserRechargeStatus: NewUserRechargeStatus | null
  loadingNewUserRechargeStatus: boolean
}

const Hero: React.FC<HeroProps> = ({ newUserRechargeStatus, loadingNewUserRechargeStatus }) => {
  const [currentSlide, setCurrentSlide] = useState(0)
  const { t, language } = useLanguage()
  const { navigate } = useNavigation()
  const [rawSlides, setRawSlides] = useState<HeroSlide[]>([])
  const [loading, setLoading] = useState(true)
  const { status: addDesktopStatus, loading: loadingAddDesktopStatus } = useAddDesktopStatus(true)
  const slides = rawSlides.filter((slide) => {
    if (slide.jumpLink === "/adddesktop" && addDesktopStatus?.claimed) {
      return false
    }

    if (slide.jumpLink === "/newUserRecharge" && newUserRechargeStatus?.hidden) {
      return false
    }

    return true
  })
  const activeSlideIndex = slides.length === 0 ? 0 : currentSlide % slides.length

  useEffect(() => {
    const fetchBanners = async () => {
      try {
        setLoading(true)
        const response = await gameService.getBanners(language)
        if (response.code === 0 && response.data) {
          setRawSlides(response.data)
        }
      } catch (error) {
        console.error("Failed to fetch banners:", error)
      } finally {
        setLoading(false)
      }
    }

    fetchBanners()
  }, [language])

  const getTranslatedText = (key: string): string => {
    try {
      const translated = t(key)
      // If translation key not found, it returns the key itself or empty string
      // In that case, return the original key as display text
      if (translated === key || !translated) {
        return key
      }
      return translated
    } catch {
      return key
    }
  }

  useEffect(() => {
    if (slides.length === 0) return

    const timer = setInterval(() => {
      setCurrentSlide((prev) => (prev + 1) % slides.length)
    }, 5000)
    return () => clearInterval(timer)
  }, [slides.length])

  const nextSlide = () => setCurrentSlide((prev) => (prev + 1) % slides.length)
  const prevSlide = () => setCurrentSlide((prev) => (prev - 1 + slides.length) % slides.length)

  const navigateToJumpLink = (jumpLink: string) => {
    if (jumpLink === "/adddesktop") {
      navigate("adddesktop")
      return
    }

    if (jumpLink === "/sevenDayTopup") {
      navigate("sevenDayTopup")
      return
    }

    if (jumpLink === "/weeklySpinWheel") {
      navigate("weeklySpinWheel")
      return
    }

    if (jumpLink === "/dailyWeeklyChallenge") {
      navigate("dailyWeeklyChallenge")
      return
    }

    if (jumpLink === "/newUserRecharge") {
      navigate("newUserRecharge")
      return
    }

    if (jumpLink === "/bettingRank") {
      navigate("bettingRank")
      return
    }

    window.location.assign(jumpLink)
  }

  const renderButtons = (slide: HeroSlide) => {
    if (!slide.buttons || slide.buttons.length === 0) return null

    return (
      <div className="flex flex-col sm:flex-row flex-wrap gap-3 sm:gap-4 justify-center px-4">
        {slide.buttons.map((btn, idx) => (
          <button
            key={idx}
            className={`px-6 sm:px-8 py-3 sm:py-3.5 rounded-xl font-bold text-sm sm:text-base shadow-2xl hover:scale-105 transition-all duration-300 flex items-center justify-center gap-2 ${btn.primary
              ? "bg-gradient-to-r from-lucky-gold via-yellow-500 to-lucky-gold text-black shadow-lucky-gold/50"
              : "bg-white/5 text-white border-2 border-lucky-gold/30 hover:bg-lucky-gold/10 hover:border-lucky-gold/60"
              }`}
          >
            {getTranslatedText(btn.label)} {btn.primary && <ChevronRight size={18} />}
          </button>
        ))}
      </div>
    )
  }

  const HERO_HEIGHT = "h-[320px] sm:h-[360px] md:h-[400px] lg:h-[420px]"

  if (loading || loadingAddDesktopStatus || loadingNewUserRechargeStatus) {
    return (
      <div className="relative mt-2 sm:mt-5 pt-1 sm:pt-4 pb-3 sm:pb-4 overflow-hidden">
        <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8">
          <div
            className={`relative rounded-2xl sm:rounded-3xl overflow-hidden shadow-2xl border border-lucky-gold/20 bg-gradient-to-br from-gray-900 via-black to-gray-900 ${HERO_HEIGHT} flex items-center justify-center`}
          >
            <div className="text-lucky-gold text-lg">Loading...</div>
          </div>
        </div>
      </div>
    )
  }

  if (slides.length === 0) {
    return (
      <div className="relative mt-2 sm:mt-5 pt-1 sm:pt-4 pb-3 sm:pb-4 overflow-hidden">
        <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8">
          <div
            className={`relative rounded-2xl sm:rounded-3xl overflow-hidden shadow-2xl border border-lucky-gold/20 bg-gradient-to-br from-gray-900 via-black to-gray-900 ${HERO_HEIGHT} flex items-center justify-center`}
          >
            <div className="text-gray-400 text-lg">No banners available</div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="relative mt-2 sm:mt-5 pt-1 sm:pt-4 pb-3 sm:pb-4 overflow-hidden">
      <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8">
        <div className="relative rounded-2xl sm:rounded-3xl overflow-hidden shadow-2xl border border-lucky-gold/20 bg-black">
          {/* Slides Container */}
          <div className={`relative w-full ${HERO_HEIGHT}`}>
            {slides.map((slide, index) => {
              const isActive = index === activeSlideIndex
              const isLink = !!slide.jumpLink
              const hasOverlayContent = Boolean(
                slide.tag || slide.title || slide.subtitle || (slide.buttons && slide.buttons.length > 0),
              )

              const handleSlideClick = () => {
                if (!slide.jumpLink) {
                  return
                }

                navigateToJumpLink(slide.jumpLink)
              }

              // Wrapper Content
              const Content = (
                <div className="relative w-full h-full flex items-center justify-center">
                  {/* Background Image Layer with light overlay */}
                  <div className="absolute inset-0 overflow-hidden">
                    <img
                      src={slide.image || "/placeholder.svg"}
                      alt={getTranslatedText(slide.title)}
                      className="absolute inset-0 h-full w-full object-cover object-top"
                    />
                    {/* Subtle overlay for readability without darkening the image */}
                    <div className="absolute inset-0 bg-gradient-to-t from-black/35 via-black/10 to-transparent"></div>
                    <div className="absolute inset-0 bg-gradient-to-r from-black/10 via-transparent to-black/10"></div>
                  </div>

                  {/* Text Content Layer */}
                  {hasOverlayContent && (
                    <div className="relative z-10 text-center max-w-4xl mx-auto px-4 sm:px-6 flex flex-col items-center h-full py-6">
                      <div className="flex-shrink-0">
                        {slide.tag && (
                          <div className="inline-flex items-center gap-2 px-3 sm:px-4 py-1 sm:py-1.5 rounded-full bg-gradient-to-r from-lucky-red via-red-600 to-lucky-red text-white text-xs sm:text-sm font-bold border border-lucky-gold/30 shadow-2xl shadow-lucky-red/50 mb-3 sm:mb-4 backdrop-blur-sm">
                            <Sparkles size={12} className="text-lucky-gold animate-pulse" />
                            <span>{getTranslatedText(slide.tag)}</span>
                          </div>
                        )}
                      </div>

                      <h1 className="text-2xl sm:text-4xl md:text-5xl lg:text-6xl font-display font-black leading-tight mb-3 sm:mb-4 flex-shrink-0">
                        <span
                          className={`text-transparent bg-clip-text bg-gradient-to-r ${slide.color} drop-shadow-[0_2px_7px_rgba(0,0,0,0.2)] [text-shadow:_0_0_30px_rgb(0_0_0_/_80%)]`}
                          style={{
                            filter: "drop-shadow(0 4px 6px rgba(0,0,0,0.3)) drop-shadow(0 0 20px rgba(0,0,0,0.7))",
                          }}
                        >
                          {getTranslatedText(slide.title)}
                        </span>
                        {isLink && (
                          <ExternalLink
                            className="inline ml-2 mb-1 text-lucky-gold opacity-70 drop-shadow-lg"
                            size={20}
                          />
                        )}
                      </h1>

                      <div className="min-h-[34px] sm:min-h-[40px] flex items-center">
                        {slide.subtitle && (
                          <p className="text-gray-100 text-sm sm:text-base md:text-lg font-medium mb-2 px-2 [text-shadow:_0_2px_8px_rgb(0_0_0_/_90%)] bg-black/20 backdrop-blur-sm rounded-lg py-1.5 inline-block flex-shrink-0">
                            {getTranslatedText(slide.subtitle)}
                          </p>
                        )}
                      </div>

                      <div className="mt-auto h-[90px] sm:h-[80px] flex items-center justify-center flex-shrink-0">
                        {!isLink && renderButtons(slide)}
                      </div>
                    </div>
                  )}
                </div>
              )

              return (
                <div
                  key={slide.id}
                  className={`absolute inset-0 h-full w-full transition-all duration-700 ${isActive ? "opacity-100" : "opacity-0 pointer-events-none"
                    }`}
                >
                  {isLink ? (
                    <a
                      href={slide.jumpLink}
                      onClick={(e) => {
                        e.preventDefault()
                        handleSlideClick()
                      }}
                      className="block h-full w-full group cursor-pointer"
                    >
                      {Content}
                    </a>
                  ) : (
                    Content
                  )}
                </div>
              )
            })}
          </div>

          <div className="absolute bottom-4 sm:bottom-8 left-1/2 -translate-x-1/2 flex gap-2 sm:gap-3 z-20">
            {slides.map((_, idx) => (
              <button
                key={idx}
                onClick={() => setCurrentSlide(idx)}
                className={`h-2 sm:h-2.5 rounded-full transition-all duration-300 ${idx === activeSlideIndex
                  ? "bg-lucky-gold w-8 sm:w-12 shadow-lg shadow-lucky-gold/70"
                  : "bg-white/40 hover:bg-white/70 w-2 sm:w-2.5"
                  }`}
              />
            ))}
          </div>

          <button
            onClick={prevSlide}
            className="absolute left-2 sm:left-6 top-1/2 -translate-y-1/2 p-2 sm:p-3 rounded-full bg-black/60 hover:bg-black/80 border-2 border-lucky-gold/30 text-lucky-gold transition-all duration-300 z-20 hover:border-lucky-gold hover:shadow-lg hover:shadow-lucky-gold/50 hover:scale-110"
          >
            <ChevronLeft size={20} className="sm:hidden" />
            <ChevronLeft size={28} className="hidden sm:block" />
          </button>
          <button
            onClick={nextSlide}
            className="absolute right-2 sm:right-6 top-1/2 -translate-y-1/2 p-2 sm:p-3 rounded-full bg-black/60 hover:bg-black/80 border-2 border-lucky-gold/30 text-lucky-gold transition-all duration-300 z-20 hover:border-lucky-gold hover:shadow-lg hover:shadow-lucky-gold/50 hover:scale-110"
          >
            <ChevronRight size={20} className="sm:hidden" />
            <ChevronRight size={28} className="hidden sm:block" />
          </button>
        </div>
      </div>
    </div>
  )
}

export default Hero
