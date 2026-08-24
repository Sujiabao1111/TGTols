"use client"

import { createContext, useContext, useState, useEffect, ReactNode } from "react"
import { gameService } from "../../services/api"
import { GameType } from "../mocks/types"

interface GameContextType {
  gameTypes: GameType[]
  selectedGameType: GameType | null
  sortType: "POPULAR" | "RECOMMEND" | "NEW" | undefined
  setSelectedGameType: (gameType: GameType | null) => void
  setSortType: (sortType: "POPULAR" | "RECOMMEND" | "NEW" | undefined) => void
  clearFilters: () => void
}

const GameContext = createContext<GameContextType | undefined>(undefined)

export const GameProvider = ({ children }: { children: ReactNode }) => {
  const [gameTypes, setGameTypes] = useState<GameType[]>([])
  const [selectedGameType, setSelectedGameType] = useState<GameType | null>(null)
  const [sortType, setSortType] = useState<"POPULAR" | "RECOMMEND" | "NEW" | undefined>(undefined)

  useEffect(() => {
    const fetchGameTypes = async () => {
      try {
        const response = await gameService.getGameOptions()
        if (response.code === 0 && response.data?.game_types) {
          setGameTypes(response.data.game_types)
        }
      } catch (error) {
        console.error("Failed to fetch game types:", error)
      }
    }
    fetchGameTypes()
  }, [])

  const clearFilters = () => {
    setSelectedGameType(null)
    setSortType(undefined)
  }

  return (
    <GameContext.Provider
      value={{
        gameTypes,
        selectedGameType,
        sortType,
        setSelectedGameType,
        setSortType,
        clearFilters,
      }}
    >
      {children}
    </GameContext.Provider>
  )
}

export const useGame = () => {
  const context = useContext(GameContext)
  if (!context) {
    throw new Error("useGame must be used within GameProvider")
  }
  return context
}
