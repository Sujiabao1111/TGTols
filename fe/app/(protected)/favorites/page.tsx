"use client"

import { useRouter } from "next/navigation"
import FavoritesPage from "@/components/FavoritesPage"
import { Game, ServerGame } from "@/app/mocks/types"

export default function FavoritesRoute() {
  const router = useRouter()

  const handleGameSelect = (game: Game | ServerGame) => {
    // Store game in session storage for retrieval in game page
    const gameId = game.id
    sessionStorage.setItem(`game_${gameId}`, JSON.stringify(game))
    router.push(`/game/${gameId}`)
  }

  return <FavoritesPage onGameSelect={handleGameSelect} />
}
