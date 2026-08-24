"use client"

import type React from "react"
import { Facebook, Twitter, Instagram } from "lucide-react"
import type { View } from "../app/mocks/types"

interface FooterProps {
  onNavigate?: (view: View) => void
}

const Footer: React.FC<FooterProps> = ({ onNavigate }) => {
  const handleClick = (e: React.MouseEvent, view: View) => {
    e.preventDefault()
    if (onNavigate) {
      onNavigate(view)
    }
  }

  return (
    <footer className="bg-lucky-dark border-t border-white/10 pt-16 pb-24 md:pb-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-12">
          {/* Brand */}
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <img src="/images/PPLogo.png" alt="PPNET" className="h-8 w-8 object-contain" />
              <span className="font-display font-bold text-xl text-white">PPNET</span>
            </div>
            <p className="text-gray-400 text-sm">
              The premium destination for online entertainment. Licensed and regulated for fair play.
            </p>
            <div className="flex gap-4">
              <a
                href="#"
                className="w-8 h-8 rounded-full bg-white/5 flex items-center justify-center hover:bg-lucky-pink transition-colors"
              >
                <Facebook size={16} />
              </a>
              <a
                href="#"
                className="w-8 h-8 rounded-full bg-white/5 flex items-center justify-center hover:bg-lucky-pink transition-colors"
              >
                <Twitter size={16} />
              </a>
              <a
                href="#"
                className="w-8 h-8 rounded-full bg-white/5 flex items-center justify-center hover:bg-lucky-pink transition-colors"
              >
                <Instagram size={16} />
              </a>
            </div>
          </div>

          {/* Links */}
          <div>
            <h4 className="font-bold text-white mb-4">Games</h4>
            <ul className="space-y-2 text-sm text-gray-400">
              <li>
                <a href="#" className="hover:text-lucky-gold">
                  Slots
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-lucky-gold">
                  Live Casino
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-lucky-gold">
                  Sportsbook
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-lucky-gold">
                  Fishing
                </a>
              </li>
            </ul>
          </div>

          <div>
            <h4 className="font-bold text-white mb-4">Support</h4>
            <ul className="space-y-2 text-sm text-gray-400">
              <li>
                <a
                  href="#"
                  onClick={(e) => handleClick(e, "help-center")}
                  className="hover:text-lucky-gold cursor-pointer"
                >
                  Help Center
                </a>
              </li>
              <li>
                <a
                  href="#"
                  onClick={(e) => handleClick(e, "fairness-policy")}
                  className="hover:text-lucky-gold cursor-pointer"
                >
                  Fairness Policy
                </a>
              </li>
              <li>
                <a
                  href="#"
                  onClick={(e) => handleClick(e, "privacy-policy")}
                  className="hover:text-lucky-gold cursor-pointer"
                >
                  Privacy Policy
                </a>
              </li>
              <li>
                <a
                  href="#"
                  onClick={(e) => handleClick(e, "contact-us")}
                  className="hover:text-lucky-gold cursor-pointer"
                >
                  Contact Us
                </a>
              </li>
            </ul>
          </div>

          {/* Payment Providers (Visual Only) */}
          <div>
            <h4 className="font-bold text-white mb-4">Secure Payments</h4>
            <div className="grid grid-cols-3 gap-2">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <div
                  key={i}
                  className="h-8 bg-white/5 rounded border border-white/10 flex items-center justify-center text-[10px] text-gray-500"
                >
                  PROVIDER
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-12 pt-8 border-t border-white/5 text-center text-xs text-gray-600">
          &copy; 2026 PPNET Games. All rights reserved. 18+ Only. Gamble Responsibly.
        </div>
      </div>
    </footer>
  )
}

export default Footer
