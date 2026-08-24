"use client"
import type React from "react"
import { useState, useEffect, useCallback, useRef } from "react"
import type { Game, ServerGame, GameLaunchResponse, GameLobbyResponse, DoFavoriteResponse } from "../app/mocks/types"
import { Maximize2, Minimize2, Heart, ArrowLeft } from "lucide-react"
import { gamePlatformState, gameService } from "../services/api"
import { useLanguage } from "../app/contexts/LanguageContext"
import { useUser } from "../app/contexts/UserContext"
import { showAddDesktopInsuranceDialog } from "./AddDesktopInsuranceDialog"

interface GamePlayPageProps {
  game: Game | ServerGame
  onBackToHome?: () => void
  favoriteGameIds?: Set<number>
}

const GamePlayPage: React.FC<GamePlayPageProps> = ({ game, onBackToHome, favoriteGameIds }) => {
  const [loading, setLoading] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [gameUrl, setGameUrl] = useState<string | null>(null)
  const [gameHtml, setGameHtml] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLaunching, setIsLaunching] = useState(false)
  const [isReturning, setIsReturning] = useState(false)
  const [isFavorite, setIsFavorite] = useState(false)
  const [isFavoriting, setIsFavoriting] = useState(false)
  const [iframeLoaded, setIframeLoaded] = useState(false)
  const [iframeTimeout, setIframeTimeout] = useState(false)
  const [iframeKey, setIframeKey] = useState(0)
  const [launchedPlatformCode, setLaunchedPlatformCode] = useState<string | null>(null)
  const iframeTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const launchErrorReportedRef = useRef(false)
  const recycleRequestedRef = useRef(false)
  const { language } = useLanguage()
  const { updateBalance, fetchUserBalance } = useUser()
  const serverGame = "game_code" in game ? (game as ServerGame) : null

  const reportLaunchError = useCallback(
    async (errorType: "launch_request_failed" | "iframe_load_timeout", errorMessage: string, launchUrl?: string) => {
      if (!serverGame || launchErrorReportedRef.current) {
        return
      }

      launchErrorReportedRef.current = true

      try {
        await gameService.reportLaunchError({
          game_code: serverGame.game_code,
          game_name: serverGame.name?.EN || "",
          provider_code: serverGame.provider?.code || "",
          provider_name: serverGame.provider?.name || "",
          is_lobby: !!serverGame.is_lobby,
          is_mobile: window.innerWidth < 768,
          language,
          error_type: errorType,
          error_message: errorMessage,
          page_url: window.location.href,
          game_url: launchUrl,
        })
      } catch (reportError) {
        launchErrorReportedRef.current = false
        console.error("Game launch error report failed:", reportError)
      }
    },
    [language, serverGame],
  )

  const handleIframeLoad = useCallback(() => {
    if (iframeTimeoutRef.current) {
      clearTimeout(iframeTimeoutRef.current)
      iframeTimeoutRef.current = null
    }
    setIframeLoaded(true)
    setIframeTimeout(false)
  }, [])

  useEffect(() => {
    const timer = setTimeout(() => {
      setLoading(false)
    }, 1000)
    return () => clearTimeout(timer)
  }, [])

  useEffect(() => {
    if ("game_code" in game) {
      const serverGame = game as ServerGame
      const gameId = typeof serverGame.id === "string" ? Number.parseInt(serverGame.id, 10) : serverGame.id
      const timer = window.setTimeout(() => {
        if (serverGame.is_favorite !== undefined) {
          setIsFavorite(serverGame.is_favorite)
        } else if (favoriteGameIds) {
          setIsFavorite(favoriteGameIds.has(gameId))
        }
      }, 0)
      return () => window.clearTimeout(timer)
    }
  }, [game, favoriteGameIds])

  // 当 gameUrl 变化时，重置 iframe 加载状态并设置超时检测
  useEffect(() => {
    if (gameUrl || gameHtml) {
      const resetTimer = window.setTimeout(() => {
        setIframeLoaded(false)
        setIframeTimeout(false)
      }, 0)
      
      // 30秒超时检测（考虑到大型游戏在慢网速下的加载时间）
      iframeTimeoutRef.current = setTimeout(() => {
        setIframeTimeout(true)
      }, 30000)
      
      return () => {
        window.clearTimeout(resetTimer)
        if (iframeTimeoutRef.current) {
          clearTimeout(iframeTimeoutRef.current)
          iframeTimeoutRef.current = null
        }
      }
    }
  }, [gameUrl, gameHtml, iframeKey])

  const handleRetry = useCallback(() => {
    if (iframeTimeoutRef.current) {
      clearTimeout(iframeTimeoutRef.current)
      iframeTimeoutRef.current = null
    }
    launchErrorReportedRef.current = false
    setIframeTimeout(false)
    setIframeLoaded(false)
    // 通过改变 key 强制 iframe 重新加载
    setIframeKey(prev => prev + 1)
  }, [])

  const handlePlayNow = async () => {
    if (!("game_code" in game)) {
      console.error("Cannot launch: game_code not found")
      return
    }

    launchErrorReportedRef.current = false
    setIsLaunching(true)
    setError(null)
    setGameUrl(null)
    setGameHtml(null)

    try {
      const serverGame = game as ServerGame
      const isMobile = window.innerWidth < 768

      if (serverGame.is_lobby) {
        const platformCode = serverGame.platform_code || serverGame.provider.platform_code || "HEDOC"
        const response: GameLobbyResponse = await gameService.getGameLobby(
          serverGame.game_type || "LIVE",
          serverGame.provider.code,
          isMobile,
          language,
          platformCode,
        )

        if (response.code === 0 && (response.data?.url || response.data?.html)) {
          const actualPlatformCode = response.data.platform_code || platformCode
          if (response.data.launch_content_type === "html" && response.data.html) {
            setGameHtml(response.data.html)
            setGameUrl(null)
          } else {
            setGameUrl(response.data.url)
            setGameHtml(null)
          }
          setLaunchedPlatformCode(actualPlatformCode)
          gamePlatformState.markActive(actualPlatformCode)
          // 余额已带入游戏，更新本地余额为 0
          updateBalance(response.data.local_balance ?? 0)
        } else {
          const message = response.message || "Failed to load game lobby"
          void reportLaunchError("launch_request_failed", message)
          setError(message)
        }
      } else {
        const platformCode = serverGame.platform_code || serverGame.provider.platform_code || "HEDOC"
        const response: GameLaunchResponse = await gameService.launchGame(
          game.game_code,
          isMobile,
          language,
          serverGame.id,
          platformCode,
        )

        if (response.code === 0 && (response.data?.url || response.data?.html)) {
          const actualPlatformCode = response.data.platform_code || platformCode
          if (response.data.launch_content_type === "html" && response.data.html) {
            setGameHtml(response.data.html)
            setGameUrl(null)
          } else {
            setGameUrl(response.data.url)
            setGameHtml(null)
          }
          setLaunchedPlatformCode(actualPlatformCode)
          gamePlatformState.markActive(actualPlatformCode)
          // 余额已带入游戏，更新本地余额为 0
          updateBalance(response.data.local_balance ?? 0)
          const gameId = typeof serverGame.id === "string" ? Number.parseInt(serverGame.id, 10) : serverGame.id
          if (response.data.is_favorite !== undefined) {
            setIsFavorite(response.data.is_favorite)
          } else if (favoriteGameIds) {
            setIsFavorite(favoriteGameIds.has(gameId))
          }
        } else {
          const message = response.message || "Failed to launch game"
          void reportLaunchError("launch_request_failed", message)
          setError(message)
        }
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : "Network error. Please try again."
      void reportLaunchError("launch_request_failed", message)
      setError(message)
      console.error("Game launch error:", err)
    } finally {
      setIsLaunching(false)
    }
  }

  useEffect(() => {
    const isServerGame = "game_code" in game
    const isLobbyGame = isServerGame && (game as ServerGame).is_lobby

    if (isLobbyGame && !gameUrl && !gameHtml && !error && !isLaunching) {
      const timer = window.setTimeout(() => {
        void handlePlayNow()
      }, 0)
      return () => window.clearTimeout(timer)
    }
  }, [game])

  useEffect(() => {
    if (!iframeTimeout || iframeLoaded || (!gameUrl && !gameHtml)) {
      return
    }

    void reportLaunchError("iframe_load_timeout", "iframe did not finish loading within 30 seconds", gameUrl || undefined)
  }, [gameUrl, gameHtml, iframeLoaded, iframeTimeout, reportLaunchError])

  const toggleFavorite = async () => {
    if (!("game_code" in game)) {
      console.error("Cannot toggle favorite: not a ServerGame")
      return
    }

    const serverGame = game as ServerGame
    setIsFavoriting(true)

    try {
      const newFavoriteState = !isFavorite
      const gameId = typeof serverGame.id === "string" ? Number.parseInt(serverGame.id, 10) : serverGame.id
      const response: DoFavoriteResponse = await gameService.doFavorite(gameId, newFavoriteState)

      if (response.code === 0) {
        setIsFavorite(newFavoriteState)
        window.dispatchEvent(new Event("favorites-updated"))
      } else {
        console.error("Toggle favorite failed:", response.message)
      }
    } catch (err) {
      console.error("Toggle favorite error:", err)
    } finally {
      setIsFavoriting(false)
    }
  }

  const handleReturnToHome = async () => {
    if (recycleRequestedRef.current) {
      return
    }

    setIsReturning(true)
    try {
      const platformCode = launchedPlatformCode || serverGame?.platform_code || serverGame?.provider?.platform_code || "HEDOC"
      recycleRequestedRef.current = true
      const recycleResult = await gameService.autoRecycleBalance({ force: true, cooldownMs: 0, platformCode })
      gamePlatformState.clearIfCurrent(platformCode)
      if (recycleResult?.data?.current_balance !== undefined) {
        updateBalance(recycleResult.data.current_balance)
      }
      if (recycleResult?.data?.insurance?.triggered) {
        const amount = `${(recycleResult.data.insurance.compensation_amount_u || 0).toFixed(2)}U`
        showAddDesktopInsuranceDialog(amount)
      }
      if (recycleResult?.data?.summary_pending) {
        void gameService.refreshPlayerTransactionsAfterPendingSync()
      }
      await fetchUserBalance()
    } catch (err) {
      console.error("Recycle balance error:", err)
    } finally {
      setIsReturning(false)
      if (onBackToHome) {
        onBackToHome()
      }
    }
  }

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().then(() => setIsFullscreen(true))
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen().then(() => setIsFullscreen(false))
      }
    }
  }

  const getGameInfo = () => {
    if ("title" in game) {
      return {
        title: game.title,
        provider: game.provider,
        image: game.image,
      }
    } else {
      return {
        title: game.name.EN,
        provider: game.provider.name,
        image: game.img_url,
      }
    }
  }

  const gameInfo = getGameInfo()

  return (
    <div
      className="relative flex w-full flex-col overflow-hidden bg-black"
      style={{ height: "calc(100dvh - var(--app-header-height) - var(--app-header-clearance))" }}
    >
      {/* Game Toolbar */}
      <div className="h-12 bg-lucky-dark border-b border-white/10 flex items-center justify-between px-4">
        <div className="flex items-center gap-3">
          <button
            onClick={handleReturnToHome}
            disabled={isReturning}
            className="text-gray-400 hover:text-lucky-gold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="Return to Homepage"
          >
            {isReturning ? (
              <div className="w-5 h-5 border-2 border-lucky-gold border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <ArrowLeft size={20} />
            )}
          </button>
          <span className="font-bold text-white">{gameInfo.title}</span>
          <span className="text-xs px-2 py-0.5 rounded bg-white/10 text-gray-400">{gameInfo.provider}</span>
        </div>
        <div className="flex items-center gap-4">
          <button
            onClick={toggleFavorite}
            disabled={isFavoriting}
            className={`transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${
              isFavorite ? "text-lucky-pink" : "text-gray-400 hover:text-lucky-pink"
            }`}
            title={isFavorite ? "Remove from favorites" : "Add to favorites"}
          >
            {isFavoriting ? (
              <div className="w-5 h-5 border-2 border-lucky-pink border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <Heart size={20} fill={isFavorite ? "currentColor" : "none"} />
            )}
          </button>
          <button className="text-gray-400 hover:text-white transition-colors" onClick={toggleFullscreen}>
            {isFullscreen ? <Minimize2 size={20} /> : <Maximize2 size={20} />}
          </button>
        </div>
      </div>

      {/* Game Container */}
      <div className="flex-1 relative bg-[#0a0a0a] flex items-center justify-center overflow-hidden">
        {loading ? (
          <div className="flex flex-col items-center gap-4">
            <div className="w-16 h-16 border-4 border-lucky-gold border-t-transparent rounded-full animate-spin"></div>
            <div className="text-lucky-gold font-bold animate-pulse">Loading {gameInfo.title}...</div>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center gap-6 p-8 text-center">
            <div className="w-20 h-20 rounded-full bg-red-500/20 flex items-center justify-center">
              <span className="text-4xl">⚠️</span>
            </div>
            <div>
              <h2 className="text-2xl font-bold text-white mb-2">Game Launch Failed</h2>
              <p className="text-red-400 max-w-md">{error}</p>
            </div>
            <button
              onClick={handleReturnToHome}
              disabled={isReturning}
              className="flex items-center gap-2 px-6 py-3 bg-lucky-gold hover:bg-lucky-gold/90 text-lucky-dark font-bold rounded-full transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isReturning ? (
                <>
                  <div className="w-5 h-5 border-2 border-lucky-dark border-t-transparent rounded-full animate-spin"></div>
                  Returning...
                </>
              ) : (
                <>
                  <ArrowLeft size={20} />
                  Return to Homepage
                </>
              )}
            </button>
          </div>
        ) : gameUrl || gameHtml ? (
          <div className="relative w-full h-full bg-[#0a0a0a]">
            {/* iframe */}
            <iframe
              key={iframeKey}
              src={gameHtml ? undefined : gameUrl || undefined}
              srcDoc={gameHtml || undefined}
              className="w-full h-full border-0"
              title={gameInfo.title}
              allow="autoplay; fullscreen; payment"
              onLoad={handleIframeLoad}
            />
            
            {/* 加载中遮罩 - 使用游戏封面作为背景 */}
            {!iframeLoaded && !iframeTimeout && (
              <div className="absolute inset-0 z-10 flex flex-col items-center justify-center overflow-hidden">
                {/* 封面背景 */}
                {gameInfo.image && (
                  <img 
                    src={gameInfo.image} 
                    alt="" 
                    className="absolute inset-0 w-full h-full object-cover opacity-20 blur-md scale-110"
                  />
                )}
                <div className="absolute inset-0 bg-black/70" />
                
                {/* 加载动画 */}
                <div className="relative z-20 flex flex-col items-center">
                  <div className="w-16 h-16 border-4 border-lucky-gold border-t-transparent rounded-full animate-spin mb-4" />
                  <div className="text-lucky-gold font-bold text-lg">Loading game...</div>
                  <div className="text-gray-400 text-sm mt-2">Please wait a moment</div>
                </div>
              </div>
            )}
            
            {/* 超时提示 */}
            {iframeTimeout && !iframeLoaded && (
              <div className="absolute inset-0 z-20 bg-[#0a0a0a] flex flex-col items-center justify-center p-8">
                <div className="text-5xl mb-4">⏱️</div>
                <h3 className="text-xl font-bold text-white mb-2">Loading is taking longer than expected</h3>
                <p className="text-gray-400 mb-2">Please check your network connection</p>
                <div className="flex gap-4 mt-6">
                  <button 
                    onClick={handleRetry} 
                    className="px-6 py-2 bg-lucky-gold text-lucky-dark font-bold rounded-full hover:bg-lucky-gold/90 transition-colors"
                  >
                    Retry
                  </button>
                  <button 
                    onClick={handleReturnToHome} 
                    disabled={isReturning}
                    className="px-6 py-2 border border-gray-600 text-white rounded-full hover:bg-white/10 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isReturning ? (
                      <span className="flex items-center gap-2">
                        <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                        Returning...
                      </span>
                    ) : (
                      "Back to Home"
                    )}
                  </button>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="relative w-full h-full">
            <div className="absolute inset-0 flex items-center justify-center bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] bg-repeat opacity-20"></div>
            <div className="absolute inset-0 flex flex-col items-center justify-center z-10 text-center p-8">
              <img
                src={gameInfo.image || "/placeholder.svg"}
                alt={gameInfo.title}
                className="w-32 h-32 rounded-2xl shadow-2xl mb-6 opacity-80"
              />
              <h2 className="text-3xl font-display font-bold text-white mb-4">GAME SCREEN</h2>
              <p className="text-gray-400 max-w-md mb-8">
                Ready to play {gameInfo.title} by {gameInfo.provider}
              </p>
              <button
                onClick={handlePlayNow}
                disabled={isLaunching}
                className="px-8 py-4 bg-gradient-to-r from-lucky-gold to-yellow-500 hover:from-yellow-500 hover:to-lucky-gold text-lucky-dark font-bold rounded-full transition-all transform hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
              >
                {isLaunching ? (
                  <span className="flex items-center gap-2">
                    <div className="w-5 h-5 border-2 border-lucky-dark border-t-transparent rounded-full animate-spin"></div>
                    Launching...
                  </span>
                ) : (
                  "PLAY NOW"
                )}
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Game Footer Info - Hidden
      <div className="bg-lucky-dark p-4 border-t border-white/10">
        <div className="flex items-start gap-2 text-xs text-gray-500">
          <Info size={14} className="mt-0.5 shrink-0" />
          <p>
            Malfunction voids all plays and pays. In case of disconnection, wait a few minutes and reload the game to
            resume your session.
          </p>
        </div>
      </div>
      */}
    </div>
  )
}

export default GamePlayPage
