"use client"

import type React from "react"
import { useState } from "react"
import { Play, Flame, Zap, Star, Coins } from "lucide-react"
import type { Game, ServerGame } from "../app/mocks/types"

interface GameCardProps {
  game: Game | ServerGame
  onSelect: (game: Game | ServerGame) => void
  language?: string
}

const isServerGame = (game: Game | ServerGame): game is ServerGame => {
  return "game_code" in game
}

const GameCard: React.FC<GameCardProps> = ({ game, onSelect, language = "EN" }) => {
  const [isMobileActive, setIsMobileActive] = useState(false)
  const [failedImageSrc, setFailedImageSrc] = useState<string | null>(null)

  const handleMobileTap = () => {
    setIsMobileActive(!isMobileActive)
  }

  const displayData = isServerGame(game)
    ? {
        title: game.name[language as keyof typeof game.name] || game.name.EN,
        image: game.img_url,
        provider: game.provider.name,
        isHot: false,
        isNew: false,
        isRecommended: false,
        minBuy: 0,
        maxBuy: 0,
      }
    : {
        title: game.title,
        image: game.image,
        provider: game.provider,
        isHot: game.isHot || false,
        isNew: game.isNew || false,
        isRecommended: game.isRecommended || false,
        minBuy: game.minBuy,
        maxBuy: game.maxBuy,
      }

  return (
    <div
      className="group relative rounded-2xl overflow-hidden bg-gradient-to-b from-zinc-900 to-black border-2 border-zinc-800 hover:border-lucky-gold/50 shadow-xl transition-all duration-300 md:hover:-translate-y-1 md:hover:shadow-2xl md:hover:shadow-lucky-gold/20 cursor-pointer"
      onClick={handleMobileTap}
    >
      <div className="absolute top-2 left-2 z-20 flex gap-1.5">
        {displayData.isRecommended && (
          <span className="px-2 py-1 rounded-md bg-gradient-to-r from-lucky-gold to-yellow-500 text-black text-[10px] font-bold uppercase tracking-wider shadow-lg flex items-center gap-1">
            <Star size={10} fill="currentColor" /> TOP
          </span>
        )}
        {displayData.isHot && (
          <span className="px-2 py-1 rounded-md bg-gradient-to-r from-red-600 to-red-500 text-white text-[10px] font-bold uppercase tracking-wider shadow-lg flex items-center gap-1">
            <Flame size={10} fill="white" /> HOT
          </span>
        )}
        {displayData.isNew && (
          <span className="px-2 py-1 rounded-md bg-gradient-to-r from-green-600 to-emerald-500 text-white text-[10px] font-bold uppercase tracking-wider shadow-lg flex items-center gap-1">
            <Zap size={10} fill="white" /> NEW
          </span>
        )}
      </div>

      {!isServerGame(game) && displayData.minBuy > 0 && (
        <div className="absolute bottom-2 left-2 right-2 z-20 flex items-center justify-between px-2 py-1.5 bg-black/80 backdrop-blur-md rounded-lg border border-lucky-gold/30">
          <div className="flex items-center gap-1">
            <Coins size={12} className="text-lucky-gold" />
            <span className="text-[10px] font-bold text-gray-300">Min:</span>
            <span className="text-[10px] font-bold text-white">${displayData.minBuy}</span>
          </div>
          <div className="w-px h-3 bg-gray-600" />
          <div className="flex items-center gap-1">
            <Coins size={12} className="text-lucky-gold" />
            <span className="text-[10px] font-bold text-gray-300">Max:</span>
            <span className="text-[10px] font-bold text-white">${displayData.maxBuy.toLocaleString()}</span>
          </div>
        </div>
      )}

        {/* <div className="absolute top-2 right-2 z-20">
          <span className="px-2 py-1 rounded-md bg-black/80 backdrop-blur-md text-gray-200 text-[9px] font-bold border border-white/20 shadow-lg">
            {displayData.provider}
          </span>
        </div> */}

      {/* Image */}
      <div className="aspect-[3/4] overflow-hidden bg-zinc-900">
        <img
          src={displayData.image && failedImageSrc !== displayData.image ? displayData.image : "/placeholder.svg"}
          alt={displayData.title}
          loading="lazy"
          onError={() => setFailedImageSrc(displayData.image || null)}
          className="w-full h-full object-cover md:group-hover:scale-110 transition-transform duration-500"
        />
      </div>

      <div
        className={`absolute inset-0 bg-gradient-to-t from-black via-black/80 to-transparent backdrop-blur-sm flex flex-col items-center justify-center gap-3 transition-opacity duration-300 ${
          isMobileActive ? "opacity-100 z-30" : "opacity-0 md:group-hover:opacity-100"
        }`}
      >
        <button
          onClick={(e) => {
            e.stopPropagation()
            onSelect(game)
          }}
          className={`w-14 h-14 rounded-full bg-gradient-to-br from-lucky-gold via-yellow-500 to-orange-500 text-black flex items-center justify-center shadow-2xl shadow-lucky-gold/50 transition-transform duration-300 ${isMobileActive ? "scale-100" : "scale-0 md:group-hover:scale-100"} hover:scale-110`}
        >
          <Play fill="currentColor" size={22} className="ml-1" />
        </button>
        <div className="text-center px-2">
          <span className="text-white font-bold tracking-wide block text-sm drop-shadow-lg">{displayData.title}</span>
          <span className="text-xs text-lucky-gold font-medium">{displayData.provider}</span>
        </div>
      </div>

      <div
        className={`md:hidden absolute bottom-0 left-0 right-0 p-2.5 bg-gradient-to-t from-black via-black/95 to-transparent transition-opacity ${isMobileActive ? "opacity-0" : "opacity-100"}`}
      >
        <p className="text-xs font-bold text-white truncate drop-shadow-md">{displayData.title}</p>
      </div>
    </div>
  )
}

export default GameCard
