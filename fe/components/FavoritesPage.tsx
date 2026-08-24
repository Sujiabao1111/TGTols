import type React from "react"
import { useState, useEffect } from "react"
import { Heart } from "lucide-react"
import GameCard from "./GameCard"
import type { ServerGame, Game } from "../app/mocks/types"
import { useLanguage } from "../app/contexts/LanguageContext"
import { gameService } from "../services/api"

interface FavoritesPageProps {
  onGameSelect: (game: Game | ServerGame) => void
}

const FavoritesPage: React.FC<FavoritesPageProps> = ({ onGameSelect }) => {
  const [games, setGames] = useState<ServerGame[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const { language } = useLanguage()

  useEffect(() => {
    const fetchFavorites = async () => {
      setLoading(true)
      setError(false)
      try {
        const response = await gameService.getFavoriteGames()
        if (response.code === 0 && response.data) {
          setGames(response.data.list || [])
        } else {
          setError(true)
        }
      } catch (err) {
        console.error("Failed to fetch favorite games:", err)
        setError(true)
      } finally {
        setLoading(false)
      }
    }

    fetchFavorites()

    const handleFavoritesUpdate = () => {
      fetchFavorites()
    }
    window.addEventListener("favorites-updated", handleFavoritesUpdate)

    return () => {
      window.removeEventListener("favorites-updated", handleFavoritesUpdate)
    }
  }, [])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-lucky-gold border-t-transparent"></div>
          <p className="text-gray-400 mt-4">Loading favorites...</p>
        </div>
      </div>
    )
  }

  if (error || games.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center px-4">
        <div className="text-center max-w-md">
          <div className="mb-6">
            <Heart size={64} className="mx-auto text-lucky-gold/30" />
          </div>
          <h2 className="text-2xl font-bold text-white mb-2">No favorite games</h2>
          <p className="text-gray-400">Add your favorite games for quick access</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen py-8 px-4">
      <div className="max-w-7xl mx-auto">
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-white flex items-center gap-3">
            <Heart size={32} className="text-lucky-gold" />
            My Favorites
          </h1>
          <p className="text-gray-400 mt-2">{games.length} favorite games</p>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
          {games.map((game) => (
            <GameCard key={game.id} game={game} onSelect={onGameSelect} language={language.toUpperCase()} />
          ))}
        </div>
      </div>
    </div>
  )
}

export default FavoritesPage
