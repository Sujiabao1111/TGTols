"use client"

import type React from "react"
import { ChevronRight } from "lucide-react"
import type { Game, ServerGame, GameCategory } from "../app/mocks/types"
import GameCard from "./GameCard"

interface GameCategorySectionProps {
  title: string
  category: GameCategory
  games: (Game | ServerGame)[]
  onGameSelect: (game: Game | ServerGame) => void
  onSeeAll: (category: GameCategory) => void
}

const GameCategorySection: React.FC<GameCategorySectionProps> = ({
  title,
  category,
  games,
  onGameSelect,
  onSeeAll,
}) => {
  // Take only the first 4 games for preview
  const displayGames = games.slice(0, 4)

  const getGameId = (game: Game | ServerGame): string => {
    if ("game_code" in game) {
      // ServerGame type
      return `server-${game.id}`
    } else {
      // Game type
      return game.id
    }
  }

  return (
    <section className="py-6 px-4 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-display font-bold text-white flex items-center gap-2">{title}</h2>
        <button
          onClick={() => onSeeAll(category)}
          className="flex items-center gap-1 text-sm text-gray-400 hover:text-lucky-gold transition-colors"
        >
          All <ChevronRight size={16} />
        </button>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-6">
        {displayGames.map((game) => (
          <GameCard key={getGameId(game)} game={game} onSelect={onGameSelect} />
        ))}
      </div>
    </section>
  )
}

export default GameCategorySection
