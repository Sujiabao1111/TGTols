import React from 'react';
import { Home, AlertTriangle } from 'lucide-react';

interface NotFoundPageProps {
    onGoHome: () => void;
}

const NotFoundPage: React.FC<NotFoundPageProps> = ({ onGoHome }) => {
  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center p-4 text-center animate-fade-in">
        <div className="relative mb-8">
            <div className="w-32 h-32 bg-white/5 rounded-full flex items-center justify-center backdrop-blur-sm border border-white/10">
                <AlertTriangle size={64} className="text-lucky-pink opacity-80" />
            </div>
            <div className="absolute -bottom-2 -right-2 text-6xl rotate-12 filter drop-shadow-lg">🐻</div>
        </div>
        
        <h1 className="text-8xl font-display font-bold text-transparent bg-clip-text bg-gradient-to-r from-lucky-gold to-orange-500 mb-2 drop-shadow-sm">
            404
        </h1>
        
        <h2 className="text-2xl text-white font-bold mb-4">Page Not Found</h2>
        
        <p className="text-gray-400 max-w-md mb-8 leading-relaxed">
            Oops! It seems you've wandered into a part of the forest that doesn't exist. 
            Don't worry, the Lucky Bear can guide you back.
        </p>
        
        <button 
            onClick={onGoHome}
            className="px-8 py-3 rounded-full bg-lucky-gold text-lucky-dark font-bold hover:scale-105 hover:shadow-lg hover:shadow-lucky-gold/20 transition-all flex items-center gap-2"
        >
            <Home size={20} /> 
            <span>Back to Home</span>
        </button>
    </div>
  );
};

export default NotFoundPage;
