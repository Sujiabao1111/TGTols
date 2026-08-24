import React, { useState, useEffect, useRef } from 'react';
import { MessageSquare, X, Send, Bot, Minimize2, RefreshCcw } from 'lucide-react';
import { useSupportLink } from '@/hooks/useSupportLink';
import { sendMessageToGemini, initializeChat } from '../services/gemini';
import { ChatMessage } from '../app/mocks/types';


const AIChat: React.FC = () => {
  const supportLink = useSupportLink();
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Initialize chat when opened for the first time
  useEffect(() => {
    if (isOpen && messages.length === 0) {
      const init = async () => {
        setIsLoading(true);
        await initializeChat();
        setMessages([
          { role: 'model', text: "Hi there! I'm Lucky Bear 🐻. Need help finding a game or have a question about bonuses?", timestamp: new Date() }
        ]);
        setIsLoading(false);
      };
      init();
    }
  }, [isOpen]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;

    const userMsg: ChatMessage = { role: 'user', text: input, timestamp: new Date() };
    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setIsLoading(true);

    const responseText = await sendMessageToGemini(userMsg.text);

    const botMsg: ChatMessage = { role: 'model', text: responseText, timestamp: new Date() };
    setMessages(prev => [...prev, botMsg]);
    setIsLoading(false);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <>
      {/* Floating Toggle Button */}
      {!isOpen && (
        <button
          // onClick={() => setIsOpen(true)}
          onClick={() => window.location.href = supportLink}
          className="fixed bottom-24 right-4 md:bottom-8 md:right-8 z-50 w-14 h-14 rounded-full bg-gradient-to-r from-lucky-gold to-orange-500 shadow-xl shadow-lucky-gold/30 flex items-center justify-center hover:scale-110 transition-transform group"
        >
          <div className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 rounded-full animate-pulse"></div>
          <Bot size={28} className="text-lucky-dark group-hover:rotate-12 transition-transform" />
        </button>
      )}

      {/* Chat Window */}
      {isOpen && (
        <div className="fixed bottom-0 right-0 md:bottom-8 md:right-8 z-50 w-full md:w-80 h-[500px] md:h-[600px] flex flex-col bg-lucky-dark border border-lucky-gold/20 shadow-2xl rounded-t-3xl md:rounded-3xl overflow-hidden font-sans">

          {/* Header */}
          <div className="bg-gradient-to-r from-lucky-purple to-lucky-accent p-4 flex items-center justify-between border-b border-white/10">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center">
                <Bot size={24} className="text-lucky-gold" />
              </div>
              <div>
                <h3 className="font-bold text-white">Lucky Assistant</h3>
                <div className="flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                  <span className="text-xs text-gray-300">Online</span>
                </div>
              </div>
            </div>
            <div className="flex gap-2">
              <button onClick={() => setMessages([])} className="text-gray-400 hover:text-white p-1" title="Reset Chat">
                <RefreshCcw size={18} />
              </button>
              <button onClick={() => setIsOpen(false)} className="text-gray-400 hover:text-white p-1">
                <Minimize2 size={20} />
              </button>
            </div>
          </div>

          {/* Messages Area */}
          <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-black/20">
            {messages.map((msg, idx) => (
              <div
                key={idx}
                className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`max-w-[85%] rounded-2xl px-4 py-3 text-sm ${msg.role === 'user'
                    ? 'bg-lucky-purple text-white rounded-br-none border border-lucky-pink/30'
                    : 'bg-white/10 text-gray-200 rounded-bl-none'
                    }`}
                >
                  {msg.text}
                </div>
              </div>
            ))}
            {isLoading && (
              <div className="flex justify-start">
                <div className="bg-white/10 rounded-2xl px-4 py-3 rounded-bl-none flex gap-2 items-center">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce delay-100"></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce delay-200"></div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Input Area */}
          <div className="p-4 bg-lucky-dark border-t border-white/10">
            <div className="relative">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyPress}
                placeholder="Ask about bonuses, games..."
                className="w-full bg-white/5 border border-white/10 rounded-full pl-4 pr-12 py-3 text-sm text-white focus:outline-none focus:border-lucky-gold/50 transition-colors"
              />
              <button
                onClick={handleSend}
                disabled={!input.trim() || isLoading}
                className="absolute right-2 top-1/2 -translate-y-1/2 p-2 bg-lucky-gold text-lucky-dark rounded-full hover:bg-white disabled:opacity-50 disabled:hover:bg-lucky-gold transition-colors"
              >
                <Send size={16} />
              </button>
            </div>
            <div className="text-center mt-2">
              <p className="text-[10px] text-gray-500">AI can make mistakes. Check standard rules.</p>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default AIChat;
