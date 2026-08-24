import React from 'react';
import { Home, Gift, User, Wallet, Share2 } from 'lucide-react';
import { View } from '../app/mocks/types';
import { useLanguage } from '../app/contexts/LanguageContext';

interface MobileNavProps {
  currentView: View;
  onNavigate: (view: View) => void;
}

const MobileNav: React.FC<MobileNavProps> = ({ currentView, onNavigate }) => {
  const { t } = useLanguage();
  const activityViews: View[] = ['activity', 'vip', 'adddesktop', 'bettingRank', 'weeklySpinWheel', 'dailyWeeklyChallenge', 'sevenDayTopup', 'newUserRecharge'];
  
  const getIconClass = (view: View) => 
    `flex flex-col items-center gap-1 transition-colors ${currentView === view ? 'text-lucky-gold' : 'text-gray-400 hover:text-white'}`;

  return (
    <div className="md:hidden fixed bottom-0 left-0 right-0 bg-lucky-dark border-t border-white/10 z-40 pb-safe shadow-[0_-5px_20px_rgba(0,0,0,0.5)]">
      <div className="flex justify-around items-center h-16 px-2">
        
        <button onClick={() => onNavigate('home')} className={getIconClass('home')}>
          <Home size={22} />
          <span className="text-[10px] font-medium">{t('nav.home')}</span>
        </button>

        <button onClick={() => onNavigate('invite')} className={getIconClass('invite')}>
          <Share2 size={22} />
          <span className="text-[10px] font-medium">{t('nav.invite')}</span>
        </button>
        
        {/* Center Floating Button - Promotion */}
        <div className="relative -top-6">
            <button 
                onClick={() => onNavigate('activity')}
                className={`w-14 h-14 rounded-full flex items-center justify-center shadow-lg border-4 border-lucky-dark transition-transform active:scale-95 ${activityViews.includes(currentView) ? 'bg-lucky-gold text-lucky-dark' : 'bg-gradient-to-tr from-lucky-pink to-purple-600 text-white'}`}
            >
                <Gift size={24} className={!activityViews.includes(currentView) ? "animate-pulse" : ""} />
            </button>
        </div>

        <button onClick={() => onNavigate('wallet')} className={getIconClass('wallet')}>
          <Wallet size={22} />
          <span className="text-[10px] font-medium">{t('nav.recharge')}</span>
        </button>

        <button onClick={() => onNavigate('profile')} className={getIconClass('profile')}>
          <User size={22} />
          <span className="text-[10px] font-medium">{t('nav.mine')}</span>
        </button>
      </div>
    </div>
  );
};

export default MobileNav;
