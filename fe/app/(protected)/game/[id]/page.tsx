"use client"

import { useState, useEffect } from "react"
import { useParams, useRouter } from "next/navigation"
import GamePlayPage from "@/components/GamePlayPage"
import { ServerGame, Game } from "@/app/mocks/types"
import { useUser } from "@/app/contexts/UserContext"

export default function GamePage() {
  const params = useParams()
  const router = useRouter()
  const { favoriteGameIds } = useUser()
  const [game, setGame] = useState<ServerGame | Game | null>(null)

  useEffect(() => {
    // Fetch game details by ID from session storage
    const gameId = params.id as string
    const storedGame = sessionStorage.getItem(`game_${gameId}`)

    if (storedGame) {
      setGame(JSON.parse(storedGame))
    } else {
      // If game not found in storage, redirect to home
      router.push("/home")
    }
  }, [params.id, router])

  const handleBackToHome = () => {
    // Return to the existing lobby entry so its loaded state is preserved.
    // Fall back to /home when this page was opened directly without history.
    const cameFromLobby = typeof window !== "undefined"
      && sessionStorage.getItem("game_return_to_lobby") === "1"
    if (cameFromLobby) {
      sessionStorage.removeItem("game_return_to_lobby")
      router.back()
    } else {
      router.push("/home")
    }
  }

  if (!game) {
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>
  }

  return <GamePlayPage game={game} onBackToHome={handleBackToHome} favoriteGameIds={favoriteGameIds} />
}
