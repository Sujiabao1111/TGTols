"use client"

import HomePage from "@/components/HomePage"
import { useGame } from "@/app/contexts/GameContext"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect } from "react"
import { Game, GameType, ServerGame } from "@/app/mocks/types"

export default function HomeRoute() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { gameTypes, selectedGameType, sortType, setSelectedGameType, setSortType } = useGame()
  const category = searchParams.get("category")

  useEffect(() => {
    if (category !== "sports" || gameTypes.length === 0) {
      return
    }

    const sportsGameType = gameTypes.find((gameType) => gameType.code === "SPORTBOOK")
      || gameTypes.find((gameType) => gameType.code.includes("SPORT"))

    if (sportsGameType) {
      setSortType(undefined)
      setSelectedGameType(sportsGameType)
    }
  }, [category, gameTypes, setSelectedGameType, setSortType])

  const handleGameSelect = (game: Game | ServerGame) => {
    // Store game in session storage for retrieval in game page
    const gameId = game.id
    sessionStorage.setItem(`game_${gameId}`, JSON.stringify(game))
    sessionStorage.setItem("game_return_to_lobby", "1")
    router.push(`/game/${gameId}`)
  }

  const handleFilterChange = (gameType: GameType | null) => {
    setSortType(undefined)
    setSelectedGameType(gameType)
  }

  const shouldShowHero = !sortType && (!selectedGameType || selectedGameType.code === "SLOT")

  return (
    <HomePage
      onGameSelect={handleGameSelect}
      externalGameType={selectedGameType}
      onFilterChange={handleFilterChange}
      sortType={sortType}
      showHero={shouldShowHero}
    />
  )
}
