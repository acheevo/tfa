import { useEffect, useState } from 'react';

export default function Hero() {
  const [scrollY, setScrollY] = useState(0);

  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <section className="relative h-screen overflow-hidden bg-white">
      {/* Parallax Background */}
      <div 
        className="absolute inset-0 w-full h-full"
        style={{
          transform: `translateY(${scrollY * 0.5}px)`,
        }}
      >
        <div className="absolute inset-0 bg-white" />
      </div>

      {/* Hero Image */}
      <div 
        className="absolute inset-0 flex items-center justify-center"
        style={{
          transform: `translateY(${scrollY * 0.3}px) translateX(330px)`, // Moved image 10% more to the right
        }}
      >
        <img 
          src="/src/assets/img/hero_woback.png" 
          alt="Hero Characters" 
          className="w-auto object-contain"
          style={{ height: '112.5%' }}
        />
      </div>

      {/* Content Overlay */}
      <div className="relative z-10 h-full flex flex-col items-center justify-center text-center px-4 pt-32 md:pt-40 lg:pt-48">
        <h1 className="font-cinzel font-bold text-gray-800 mb-6 md:mb-8 tracking-wide drop-shadow-sm text-4xl md:text-6xl lg:text-7xl xl:text-8xl" style={{fontSize: 'clamp(48px, 8vw, 80px)'}}>
          <span className="bg-gradient-to-r from-teal-600 via-cyan-700 to-teal-800 bg-clip-text text-transparent">
            ASCEND BEYOND
          </span>
          <br />
          <span className="bg-gradient-to-r from-slate-700 via-slate-800 to-slate-900 bg-clip-text text-transparent">
            FATE
          </span>
        </h1>
        
        <div className="mb-8">
          <div className="flex items-center justify-center mb-4">
            <div className="h-px bg-gray-400 w-24"></div>
            <div className="mx-4 text-gray-600 text-2xl">❦</div>
            <div className="h-px bg-gray-400 w-24"></div>
          </div>
          <p className="text-lg md:text-xl text-gray-700 font-cinzel max-w-3xl mx-auto">
            Master ancient arts, explore mystical dungeons, and forge your destiny in a world where every choice shapes your arcane path
          </p>
        </div>

        <button 
          className="relative font-cinzel font-bold text-2xl"
          style={{
            backgroundImage: 'url(/src/assets/img/button.png)',
            backgroundSize: 'contain',
            backgroundRepeat: 'no-repeat',
            backgroundPosition: 'center',
            padding: '52px 135px',
            border: 'none',
            backgroundColor: 'transparent',
            color: '#b7bd97',
            minWidth: '600px',
            minHeight: '180px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'color 0.3s ease'
          }}
          onMouseEnter={(e) => e.currentTarget.style.color = '#4a7d6b'}
          onMouseLeave={(e) => e.currentTarget.style.color = '#b7bd97'}
        >
          JOIN BETA
        </button>
      </div>
    </section>
  );
}