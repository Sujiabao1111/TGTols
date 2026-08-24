"use client"

import type React from "react"
import { Shield, CheckCircle, Lock, Eye, Award, BarChart } from "lucide-react"

const FairnessPolicy: React.FC = () => {
  return (
    <div className="min-h-screen bg-lucky-dark">
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-green-500/10 via-black to-blue-500/10 border-b border-lucky-gold/20 py-12 md:py-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-green-400 to-blue-500 mb-4">
            <Shield className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-green-400 via-blue-400 to-purple-500 bg-clip-text text-transparent">
            Fairness & Transparency
          </h1>
          <p className="text-gray-400 text-lg">
            We are committed to providing a fair and transparent gaming experience
          </p>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Intro */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <h2 className="text-2xl font-bold mb-4 text-lucky-gold">Our Commitment</h2>
          <p className="text-gray-300 leading-relaxed mb-4">
            At LuckyBear, we believe that trust is the foundation of every great gaming experience. We are committed to
            maintaining the highest standards of fairness, transparency, and integrity in all our operations.
          </p>
          <p className="text-gray-300 leading-relaxed">
            All games on our platform are certified and regularly audited by independent third-party testing agencies to
            ensure random outcomes and fair play.
          </p>
        </div>

        {/* Key Principles */}
        <div className="grid md:grid-cols-2 gap-6 mb-8">
          <div className="bg-gradient-to-br from-zinc-900 to-black border border-green-500/30 rounded-xl p-6">
            <CheckCircle className="w-10 h-10 text-green-400 mb-3" />
            <h3 className="text-xl font-bold mb-2">Certified Games</h3>
            <p className="text-gray-400 text-sm">
              All games are provided by licensed providers and certified by independent testing labs like eCOGRA and
              iTech Labs.
            </p>
          </div>

          <div className="bg-gradient-to-br from-zinc-900 to-black border border-blue-500/30 rounded-xl p-6">
            <Lock className="w-10 h-10 text-blue-400 mb-3" />
            <h3 className="text-xl font-bold mb-2">Random Number Generation</h3>
            <p className="text-gray-400 text-sm">
              We use certified RNG (Random Number Generator) systems to ensure completely random and unpredictable game
              outcomes.
            </p>
          </div>

          <div className="bg-gradient-to-br from-zinc-900 to-black border border-purple-500/30 rounded-xl p-6">
            <Eye className="w-10 h-10 text-purple-400 mb-3" />
            <h3 className="text-xl font-bold mb-2">Transparent Operations</h3>
            <p className="text-gray-400 text-sm">
              We maintain complete transparency in our terms, payout rates (RTP), and game rules. No hidden conditions.
            </p>
          </div>

          <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/30 rounded-xl p-6">
            <Award className="w-10 h-10 text-lucky-gold mb-3" />
            <h3 className="text-xl font-bold mb-2">Licensed & Regulated</h3>
            <p className="text-gray-400 text-sm">
              We operate under strict gaming licenses and comply with all international gaming regulations and
              standards.
            </p>
          </div>
        </div>

        {/* RTP Information */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <BarChart className="w-8 h-8 text-lucky-gold" />
            <h2 className="text-2xl font-bold">Return to Player (RTP)</h2>
          </div>
          <p className="text-gray-300 leading-relaxed mb-4">
            RTP is the percentage of all wagered money that a game will pay back to players over time. Our games have
            RTPs ranging from 94% to 99%, which are clearly displayed on each game page.
          </p>
          <p className="text-gray-300 leading-relaxed">
            For example, a game with 96% RTP means that for every $100 wagered, $96 will be returned to players on
            average. Remember, RTP is calculated over millions of game rounds and individual sessions can vary
            significantly.
          </p>
        </div>

        {/* Provably Fair */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8">
          <h2 className="text-2xl font-bold mb-4 text-lucky-gold">Provably Fair Gaming</h2>
          <p className="text-gray-300 leading-relaxed mb-4">
            Selected games on our platform use "Provably Fair" technology, which allows you to verify that each game
            outcome was completely random and not manipulated.
          </p>
          <p className="text-gray-300 leading-relaxed mb-4">
            After each round, you can check the cryptographic hash and seed values to mathematically prove the fairness
            of the result.
          </p>
          <div className="bg-black/40 border border-lucky-gold/30 rounded-lg p-4 mt-4">
            <p className="text-sm text-gray-400">
              <span className="text-lucky-gold font-bold">Note:</span> Look for the "Provably Fair" badge on game cards
              to find games with this feature.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default FairnessPolicy
