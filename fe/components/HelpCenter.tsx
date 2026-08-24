"use client"

import type React from "react"
import { Search, Book, MessageCircle, FileText, Shield, CreditCard, HelpCircle } from "lucide-react"

const HelpCenter: React.FC = () => {
  const faqs = [
    {
      category: "Getting Started",
      icon: Book,
      questions: [
        {
          q: "How do I create an account?",
          a: "Click the Register button in the top right corner and fill in the required information.",
        },
        {
          q: "How do I verify my account?",
          a: "Go to Profile > Settings > Verification and upload the required documents.",
        },
        {
          q: "What payment methods do you accept?",
          a: "We accept credit cards, e-wallets, cryptocurrencies, and bank transfers.",
        },
      ],
    },
    {
      category: "Deposits & Withdrawals",
      icon: CreditCard,
      questions: [
        {
          q: "How do I make a deposit?",
          a: "Go to Wallet > Deposit, select your payment method and amount, then follow the instructions.",
        },
        {
          q: "How long do withdrawals take?",
          a: "Withdrawals are typically processed within 24-48 hours, depending on the payment method.",
        },
        {
          q: "Are there any fees?",
          a: "Deposits are free. Withdrawal fees vary by payment method, check the Wallet page for details.",
        },
      ],
    },
    {
      category: "Games & Betting",
      icon: FileText,
      questions: [
        {
          q: "How do I start playing?",
          a: "Browse our game library, click on any game, set your bet amount, and start playing.",
        },
        {
          q: "What are min/max buy-ins?",
          a: "Min buy-in is the minimum amount required to play. Max buy-in is the maximum allowed per session.",
        },
        { q: "Can I play for free?", a: "Many games offer demo mode. Look for the 'Try Free' option on game cards." },
      ],
    },
    {
      category: "Security & Safety",
      icon: Shield,
      questions: [
        {
          q: "Is my data secure?",
          a: "Yes, we use industry-standard SSL encryption and follow strict data protection protocols.",
        },
        {
          q: "How can I set deposit limits?",
          a: "Go to Profile > Responsible Gaming to set daily, weekly, or monthly limits.",
        },
        {
          q: "What if I forget my password?",
          a: "Click 'Forgot Password' on the login page and follow the reset instructions.",
        },
      ],
    },
  ]

  return (
    <div className="min-h-screen bg-lucky-dark">
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-lucky-gold/10 via-black to-lucky-red/10 border-b border-lucky-gold/20 py-12 md:py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-lucky-gold to-orange-500 mb-4">
              <HelpCircle className="w-8 h-8 text-black" />
            </div>
            <h1 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-lucky-gold via-yellow-400 to-orange-500 bg-clip-text text-transparent">
              Help Center
            </h1>
            <p className="text-gray-400 text-lg max-w-2xl mx-auto">
              Find answers to common questions and get the support you need
            </p>
          </div>

          {/* Search Bar */}
          <div className="max-w-2xl mx-auto">
            <div className="relative">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 w-5 h-5" />
              <input
                type="text"
                placeholder="Search for help..."
                className="w-full bg-black/40 border border-lucky-gold/30 rounded-full py-3 pl-12 pr-4 text-white placeholder-gray-500 focus:outline-none focus:border-lucky-gold transition-colors"
              />
            </div>
          </div>
        </div>
      </div>

      {/* FAQ Sections */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid gap-8">
          {faqs.map((section, idx) => (
            <div
              key={idx}
              className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-6 md:p-8"
            >
              <div className="flex items-center gap-3 mb-6">
                <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-lucky-gold/20 to-orange-500/20 flex items-center justify-center border border-lucky-gold/30">
                  <section.icon className="w-6 h-6 text-lucky-gold" />
                </div>
                <h2 className="text-2xl font-bold text-white">{section.category}</h2>
              </div>

              <div className="space-y-4">
                {section.questions.map((item, qIdx) => (
                  <details key={qIdx} className="group">
                    <summary className="flex items-center justify-between cursor-pointer p-4 bg-black/40 hover:bg-black/60 rounded-lg transition-colors border border-transparent hover:border-lucky-gold/30">
                      <span className="font-medium text-white">{item.q}</span>
                      <span className="text-lucky-gold group-open:rotate-180 transition-transform">▼</span>
                    </summary>
                    <div className="mt-2 p-4 bg-black/20 rounded-lg border-l-2 border-lucky-gold">
                      <p className="text-gray-300">{item.a}</p>
                    </div>
                  </details>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Contact Support */}
        <div className="mt-12 bg-gradient-to-r from-lucky-gold/10 to-orange-500/10 border border-lucky-gold/30 rounded-xl p-8 text-center">
          <MessageCircle className="w-12 h-12 text-lucky-gold mx-auto mb-4" />
          <h3 className="text-2xl font-bold mb-2">Still need help?</h3>
          <p className="text-gray-400 mb-6">Our support team is available 24/7 to assist you</p>
          <button className="px-8 py-3 bg-gradient-to-r from-lucky-gold to-orange-500 text-black font-bold rounded-full hover:shadow-lg hover:shadow-lucky-gold/50 transition-all">
            Contact Support
          </button>
        </div>
      </div>
    </div>
  )
}

export default HelpCenter
