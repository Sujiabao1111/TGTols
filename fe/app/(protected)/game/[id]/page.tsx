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
    router.push("/home")
  }

  if (!game) {
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>
  }

  return <GamePlayPage game={game} onBackToHome={handleBackToHome} favoriteGameIds={favoriteGameIds} />
}
