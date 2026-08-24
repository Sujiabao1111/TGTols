"use client"

import type React from "react"
import { Lock, Shield, Eye, Database, UserCheck, FileText } from "lucide-react"

const PrivacyPolicy: React.FC = () => {
  return (
    <div className="min-h-screen bg-lucky-dark">
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-blue-500/10 via-black to-purple-500/10 border-b border-lucky-gold/20 py-12 md:py-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 mb-4">
            <Lock className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-blue-400 via-purple-400 to-pink-500 bg-clip-text text-transparent">
            Privacy Policy
          </h1>
          <p className="text-gray-400 text-lg">Your privacy and data security are our top priorities</p>
          <p className="text-gray-500 text-sm mt-2">Last updated: December 2024</p>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Overview */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <h2 className="text-2xl font-bold mb-4 text-lucky-gold">Overview</h2>
          <p className="text-gray-300 leading-relaxed mb-4">
            This Privacy Policy describes how LuckyBear ("we", "us", or "our") collects, uses, and protects your
            personal information when you use our gaming platform and services.
          </p>
          <p className="text-gray-300 leading-relaxed">
            By using our services, you agree to the collection and use of information in accordance with this policy. We
            are committed to protecting your privacy and handling your data in an open and transparent manner.
          </p>
        </div>

        {/* Information We Collect */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Database className="w-8 h-8 text-blue-400" />
            <h2 className="text-2xl font-bold">Information We Collect</h2>
          </div>

          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-lucky-gold mb-2">Personal Information</h3>
              <p className="text-gray-300 text-sm mb-2">When you register, we collect:</p>
              <ul className="list-disc list-inside text-gray-400 text-sm space-y-1 ml-4">
                <li>Full name, email address, phone number</li>
                <li>Date of birth and address (for age and identity verification)</li>
                <li>Payment information (processed securely through third-party providers)</li>
                <li>Government-issued ID for verification purposes</li>
              </ul>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-lucky-gold mb-2">Usage Information</h3>
              <p className="text-gray-300 text-sm mb-2">We automatically collect:</p>
              <ul className="list-disc list-inside text-gray-400 text-sm space-y-1 ml-4">
                <li>IP address, browser type, and device information</li>
                <li>Gaming activity, bet history, and transaction records</li>
                <li>Cookies and similar tracking technologies</li>
                <li>Customer support interactions</li>
              </ul>
            </div>
          </div>
        </div>

        {/* How We Use Your Information */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Eye className="w-8 h-8 text-purple-400" />
            <h2 className="text-2xl font-bold">How We Use Your Information</h2>
          </div>

          <ul className="space-y-3 text-gray-300 text-sm">
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">•</span>
              <span>To provide and maintain our services, process transactions, and manage your account</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">•</span>
              <span>To verify your identity and comply with legal requirements (KYC/AML)</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">•</span>
              <span>To detect and prevent fraud, money laundering, and underage gambling</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">•</span>
              <span>To send important updates, promotional offers, and personalized recommendations</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">•</span>
              <span>To improve our services, analyze usage patterns, and enhance user experience</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">•</span>
              <span>To provide customer support and respond to inquiries</span>
            </li>
          </ul>
        </div>

        {/* Data Protection */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Shield className="w-8 h-8 text-green-400" />
            <h2 className="text-2xl font-bold">Data Protection & Security</h2>
          </div>

          <p className="text-gray-300 leading-relaxed mb-4">
            We implement industry-standard security measures to protect your personal information:
          </p>

          <div className="grid md:grid-cols-2 gap-4">
            <div className="bg-black/40 border border-green-500/30 rounded-lg p-4">
              <h4 className="font-semibold text-green-400 mb-2">SSL Encryption</h4>
              <p className="text-gray-400 text-sm">
                All data transmitted between you and our servers is encrypted using 256-bit SSL.
              </p>
            </div>
            <div className="bg-black/40 border border-blue-500/30 rounded-lg p-4">
              <h4 className="font-semibold text-blue-400 mb-2">Secure Storage</h4>
              <p className="text-gray-400 text-sm">
                Personal data is stored on secure servers with restricted access and regular backups.
              </p>
            </div>
            <div className="bg-black/40 border border-purple-500/30 rounded-lg p-4">
              <h4 className="font-semibold text-purple-400 mb-2">Access Controls</h4>
              <p className="text-gray-400 text-sm">
                Only authorized personnel can access personal data on a need-to-know basis.
              </p>
            </div>
            <div className="bg-black/40 border border-lucky-gold/30 rounded-lg p-4">
              <h4 className="font-semibold text-lucky-gold mb-2">Regular Audits</h4>
              <p className="text-gray-400 text-sm">
                Our security practices are regularly reviewed and audited by third parties.
              </p>
            </div>
          </div>
        </div>

        {/* Your Rights */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <UserCheck className="w-8 h-8 text-lucky-gold" />
            <h2 className="text-2xl font-bold">Your Rights</h2>
          </div>

          <p className="text-gray-300 leading-relaxed mb-4">You have the right to:</p>

          <ul className="space-y-2 text-gray-300 text-sm">
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">✓</span>
              <span>
                <strong>Access:</strong> Request a copy of your personal data we hold
              </span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">✓</span>
              <span>
                <strong>Correction:</strong> Update or correct inaccurate information
              </span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">✓</span>
              <span>
                <strong>Deletion:</strong> Request deletion of your data (subject to legal obligations)
              </span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">✓</span>
              <span>
                <strong>Portability:</strong> Receive your data in a structured, machine-readable format
              </span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-lucky-gold mt-1">✓</span>
              <span>
                <strong>Opt-out:</strong> Unsubscribe from marketing communications at any time
              </span>
            </li>
          </ul>

          <div className="bg-black/40 border border-lucky-gold/30 rounded-lg p-4 mt-6">
            <p className="text-sm text-gray-400">
              To exercise any of these rights, please contact us at{" "}
              <span className="text-lucky-gold font-semibold">privacy@luckybear.com</span>
            </p>
          </div>
        </div>

        {/* Cookies */}
        <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <FileText className="w-8 h-8 text-orange-400" />
            <h2 className="text-2xl font-bold">Cookies & Tracking</h2>
          </div>

          <p className="text-gray-300 leading-relaxed mb-4">
            We use cookies and similar technologies to enhance your experience, remember your preferences, and analyze
            site usage. You can control cookies through your browser settings, but disabling them may affect site
            functionality.
          </p>
        </div>

        {/* Contact */}
        <div className="bg-gradient-to-r from-lucky-gold/10 to-orange-500/10 border border-lucky-gold/30 rounded-xl p-6 text-center">
          <h3 className="text-xl font-bold mb-2">Questions About Privacy?</h3>
          <p className="text-gray-400 mb-4">Contact our Data Protection Officer</p>
          <a
            href="mailto:privacy@luckybear.com"
            className="text-lucky-gold font-semibold hover:text-orange-400 transition-colors"
          >
            privacy@luckybear.com
          </a>
        </div>
      </div>
    </div>
  )
}

export default PrivacyPolicy
