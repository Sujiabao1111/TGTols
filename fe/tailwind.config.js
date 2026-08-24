// tailwind.config.js
   module.exports = {
     content: [
       './pages/**/*.{js,ts,jsx,tsx,mdx}',
       './components/**/*.{js,ts,jsx,tsx,mdx}',
       './app/**/*.{js,ts,jsx,tsx,mdx}',
           "*.{js,ts,jsx,tsx,mdx}"
    ],
     theme: {
       extend: {
            colors: {
              'lucky-purple': '#2e0249',
              'lucky-dark': '#1a002b',
              'lucky-gold': '#ffd700',
              'lucky-pink': '#ff007f',
              'lucky-accent': '#570a85'
            },
            fontFamily: {
              sans: ['Inter', 'sans-serif'],
              display: ['Fredoka', 'sans-serif'],
            },
            animation: {
              'fade-in': 'fadeIn 0.5s ease-out forwards',
            },
            keyframes: {
              fadeIn: {
                '0%': { opacity: '0', transform: 'translateY(10px)' },
                '100%': { opacity: '1', transform: 'translateY(0)' },
              }
            }
        }
     },
     plugins: [],
   };
