import React from "react"

const LoadingScreen: React.FC = () => {
  return (
    <div className="fixed inset-0 bg-lucky-dark z-50 flex flex-col items-center justify-center animate-fade-in">
      <div className="relative mb-6">
        <div className="w-24 h-24 rounded-full border-4 border-white/5 border-t-lucky-gold animate-spin absolute inset-0"></div>

        <div className="w-24 h-24 flex items-center justify-center relative">
          <img
            src="/images/PPLogo.png"
            alt="PPNET"
            className="h-12 w-12 animate-bounce object-contain"
          />
        </div>
      </div>

      <h2 className="text-3xl font-display font-bold text-white tracking-wider">
        PP<span className="text-lucky-gold">NET</span>
      </h2>

      <div className="mt-4 flex gap-1">
        <div className="w-2 h-2 bg-lucky-pink rounded-full animate-bounce" style={{ animationDelay: "0s" }}></div>
        <div className="w-2 h-2 bg-lucky-gold rounded-full animate-bounce" style={{ animationDelay: "0.2s" }}></div>
        <div className="w-2 h-2 bg-lucky-purple rounded-full animate-bounce" style={{ animationDelay: "0.4s" }}></div>
      </div>
    </div>
  )
}

export default LoadingScreen
