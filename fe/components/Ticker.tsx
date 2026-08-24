'use client';

import React, { useState, useEffect } from 'react';
import { Trophy } from 'lucide-react';

const Ticker: React.FC = () => {
  const [amounts, setAmounts] = useState<number[]>([]);

  useEffect(() => {
    // Generate random amounts only on client side after hydration
    setAmounts(Array.from({ length: 10 }, () => Math.random() * 1000));
  }, []);

  return (
    <div className="bg-lucky-purple border-y border-white/5 py-2 overflow-hidden relative">
      <div className="flex animate-marquee whitespace-nowrap gap-12">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="flex items-center gap-2 text-sm text-gray-300">
            <Trophy size={14} className="text-lucky-gold" />
            <span>User<span className="text-white font-mono">88***{i}</span> just won <span className="text-lucky-gold font-bold">${amounts[i - 1]?.toFixed(2) || '0.00'}</span> in Slots</span>
          </div>
        ))}
         {[1, 2, 3, 4, 5].map((i) => (
          <div key={`d-${i}`} className="flex items-center gap-2 text-sm text-gray-300">
            <Trophy size={14} className="text-lucky-gold" />
            <span>User<span className="text-white font-mono">88***{i}</span> just won <span className="text-lucky-gold font-bold">${amounts[i + 4]?.toFixed(2) || '0.00'}</span> in Slots</span>
          </div>
        ))}
      </div>
      <style>{`
        @keyframes marquee {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
        .animate-marquee {
          animation: marquee 20s linear infinite;
        }
      `}</style>
    </div>
  );
};

export default Ticker;
