"use client"

import type React from "react"
import { useState } from "react"
import { Mail, MessageCircle, Phone, MapPin, Send, Clock, Globe } from "lucide-react"

const ContactUs: React.FC = () => {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    subject: "",
    message: "",
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // Handle form submission
    console.log("Form submitted:", formData)
    alert("Thank you for contacting us! We'll get back to you soon.")
    setFormData({ name: "", email: "", subject: "", message: "" })
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value })
  }

  return (
    <div className="min-h-screen bg-lucky-dark">
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-lucky-gold/10 via-black to-orange-500/10 border-b border-lucky-gold/20 py-12 md:py-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-lucky-gold to-orange-500 mb-4">
            <MessageCircle className="w-8 h-8 text-black" />
          </div>
          <h1 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-lucky-gold via-orange-400 to-red-500 bg-clip-text text-transparent">
            Contact Us
          </h1>
          <p className="text-gray-400 text-lg">We're here to help! Reach out to us anytime</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid md:grid-cols-2 gap-8">
          {/* Contact Form */}
          <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8">
            <h2 className="text-2xl font-bold mb-6 text-lucky-gold">Send us a Message</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">Your Name</label>
                <input
                  type="text"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  required
                  className="w-full bg-black/40 border border-lucky-gold/30 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                  placeholder="John Doe"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">Email Address</label>
                <input
                  type="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  required
                  className="w-full bg-black/40 border border-lucky-gold/30 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                  placeholder="john@example.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">Subject</label>
                <select
                  name="subject"
                  value={formData.subject}
                  onChange={handleChange}
                  required
                  className="w-full bg-black/40 border border-lucky-gold/30 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-lucky-gold transition-colors"
                >
                  <option value="">Select a subject</option>
                  <option value="general">General Inquiry</option>
                  <option value="technical">Technical Support</option>
                  <option value="account">Account Issues</option>
                  <option value="payment">Payment & Withdrawals</option>
                  <option value="feedback">Feedback & Suggestions</option>
                  <option value="partnership">Business Partnership</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">Message</label>
                <textarea
                  name="message"
                  value={formData.message}
                  onChange={handleChange}
                  required
                  rows={6}
                  className="w-full bg-black/40 border border-lucky-gold/30 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-lucky-gold transition-colors resize-none"
                  placeholder="Tell us how we can help..."
                />
              </div>

              <button
                type="submit"
                className="w-full bg-gradient-to-r from-lucky-gold to-orange-500 text-black font-bold py-3 rounded-lg hover:shadow-lg hover:shadow-lucky-gold/50 transition-all flex items-center justify-center gap-2"
              >
                <Send className="w-5 h-5" />
                Send Message
              </button>
            </form>
          </div>

          {/* Contact Information */}
          <div className="space-y-6">
            {/* Quick Contact */}
            <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8">
              <h2 className="text-2xl font-bold mb-6 text-lucky-gold">Get in Touch</h2>
              <div className="space-y-4">
                <div className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-lucky-gold/20 to-orange-500/20 flex items-center justify-center border border-lucky-gold/30 flex-shrink-0">
                    <Mail className="w-5 h-5 text-lucky-gold" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white mb-1">Email</h3>
                    <a
                      href="mailto:support@luckybear.com"
                      className="text-gray-400 hover:text-lucky-gold transition-colors text-sm"
                    >
                      support@luckybear.com
                    </a>
                  </div>
                </div>

                <div className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center border border-blue-500/30 flex-shrink-0">
                    <MessageCircle className="w-5 h-5 text-blue-400" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white mb-1">Live Chat</h3>
                    <p className="text-gray-400 text-sm">Available 24/7 on our website</p>
                  </div>
                </div>

                <div className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center border border-green-500/30 flex-shrink-0">
                    <Phone className="w-5 h-5 text-green-400" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white mb-1">Phone</h3>
                    <p className="text-gray-400 text-sm">+1 (800) 123-4567</p>
                  </div>
                </div>

                <div className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-red-500/20 to-pink-500/20 flex items-center justify-center border border-red-500/30 flex-shrink-0">
                    <MapPin className="w-5 h-5 text-red-400" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white mb-1">Address</h3>
                    <p className="text-gray-400 text-sm">
                      123 Gaming Street
                      <br />
                      Las Vegas, NV 89101
                      <br />
                      United States
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Business Hours */}
            <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8">
              <div className="flex items-center gap-3 mb-4">
                <Clock className="w-6 h-6 text-lucky-gold" />
                <h2 className="text-xl font-bold">Support Hours</h2>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-400">Live Chat:</span>
                  <span className="text-white font-semibold">24/7</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Email Support:</span>
                  <span className="text-white font-semibold">24/7</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Phone Support:</span>
                  <span className="text-white font-semibold">9 AM - 12 AM EST</span>
                </div>
              </div>
            </div>

            {/* Social Media */}
            <div className="bg-gradient-to-br from-zinc-900 to-black border border-lucky-gold/20 rounded-xl p-8">
              <div className="flex items-center gap-3 mb-4">
                <Globe className="w-6 h-6 text-lucky-gold" />
                <h2 className="text-xl font-bold">Follow Us</h2>
              </div>
              <div className="flex gap-3">
                <button className="w-10 h-10 rounded-lg bg-blue-600 hover:bg-blue-700 flex items-center justify-center transition-colors">
                  <span className="text-white font-bold text-sm">f</span>
                </button>
                <button className="w-10 h-10 rounded-lg bg-sky-500 hover:bg-sky-600 flex items-center justify-center transition-colors">
                  <span className="text-white font-bold text-sm">𝕏</span>
                </button>
                <button className="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-600 to-pink-600 hover:opacity-90 flex items-center justify-center transition-opacity">
                  <span className="text-white font-bold text-sm">IG</span>
                </button>
                <button className="w-10 h-10 rounded-lg bg-red-600 hover:bg-red-700 flex items-center justify-center transition-colors">
                  <span className="text-white font-bold text-sm">YT</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ContactUs
