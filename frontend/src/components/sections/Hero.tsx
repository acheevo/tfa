import { useEffect, useState, useRef } from 'react';
import { gsap } from 'gsap';

export default function Hero() {
  const [scrollY, setScrollY] = useState(0);
  const titleRef = useRef<HTMLHeadingElement>(null);
  const line1Ref = useRef<HTMLSpanElement>(null);
  const line2Ref = useRef<HTMLSpanElement>(null);
  const particlesRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    if (line1Ref.current && line2Ref.current && particlesRef.current) {
      // Create magical particles emanating from title
      const particles: HTMLDivElement[] = [];
      for (let i = 0; i < 25; i++) {
        const particle = document.createElement('div');
        particle.className = 'absolute pointer-events-none';
        const size = Math.random() * 8 + 4; // Larger, more visible particles (4-12px)
        
        // Create different particle shapes
        const shapeType = Math.random();
        const isFromFirstLine = i < 15;
        
        if (shapeType < 0.4) {
          // Large star/sparkle shape - 30% less intense
          particle.innerHTML = '✦';
          particle.style.fontSize = size + 'px';
          particle.style.textAlign = 'center';
          particle.style.lineHeight = '1';
          particle.style.color = isFromFirstLine 
            ? `rgba(20, 184, 166, ${0.56 + Math.random() * 0.14})` // Reduced from 0.8-1.0 to 0.56-0.7
            : `rgba(71, 85, 105, ${0.56 + Math.random() * 0.14})`;
          particle.style.textShadow = isFromFirstLine
            ? '0 0 6px rgba(20, 184, 166, 0.63), 0 0 8px rgba(20, 184, 166, 0.42)' // Reduced by 30%
            : '0 0 6px rgba(71, 85, 105, 0.63), 0 0 8px rgba(71, 85, 105, 0.42)';
        } else if (shapeType < 0.7) {
          // Tiny star shape - different from large stars
          const starSymbols = ['✧', '✩', '✪', '⋆', '☆'];
          particle.innerHTML = starSymbols[Math.floor(Math.random() * starSymbols.length)];
          particle.style.fontSize = (size * 0.7) + 'px'; // Slightly smaller than main stars
          particle.style.textAlign = 'center';
          particle.style.lineHeight = '1';
          particle.style.color = isFromFirstLine 
            ? `rgba(20, 184, 166, ${0.56 + Math.random() * 0.14})`
            : `rgba(71, 85, 105, ${0.56 + Math.random() * 0.14})`;
          particle.style.textShadow = isFromFirstLine
            ? '0 0 4px rgba(20, 184, 166, 0.63), 0 0 6px rgba(20, 184, 166, 0.42)'
            : '0 0 4px rgba(71, 85, 105, 0.63), 0 0 6px rgba(71, 85, 105, 0.42)';
        } else {
          // Enhanced orb shape - 30% less intense
          particle.style.width = size + 'px';
          particle.style.height = size + 'px';
          particle.style.borderRadius = '50%';
          particle.style.background = isFromFirstLine
            ? `radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.63) 0%, rgba(20, 184, 166, 0.63) 30%, rgba(20, 184, 166, 0.42) 70%, transparent 100%)` // Reduced by 30%
            : `radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.56) 0%, rgba(71, 85, 105, 0.63) 30%, rgba(71, 85, 105, 0.42) 70%, transparent 100%)`;
          particle.style.boxShadow = isFromFirstLine
            ? '0 0 10px rgba(20, 184, 166, 0.63), inset 0 0 6px rgba(255, 255, 255, 0.21)' // Reduced by 30%
            : '0 0 10px rgba(71, 85, 105, 0.63), inset 0 0 6px rgba(255, 255, 255, 0.21)';
        }
        
        // Position particles closer to title area - constrained to avoid subtitle
        if (isFromFirstLine) {
          // Particles from "ASCEND BEYOND" area - tighter bounds
          particle.style.left = (Math.random() * 40 + 30) + '%';
          particle.style.top = (Math.random() * 8 + 32) + '%'; // Constrained vertically
        } else {
          // Particles from "FATE" area - even tighter bounds
          particle.style.left = (Math.random() * 30 + 35) + '%';
          particle.style.top = (Math.random() * 8 + 45) + '%'; // Constrained vertically
        }
        
        particlesRef.current?.appendChild(particle);
        particles.push(particle);
      }

      // Set initial state
      gsap.set([line1Ref.current, line2Ref.current], {
        opacity: 0,
        scale: 0.7,
        rotationY: 60,
        transformOrigin: "center"
      });

      gsap.set(particles, {
        opacity: 0,
        scale: 0,
        rotation: 0
      });

      // Create magical reveal timeline - slower and more dramatic
      const tl = gsap.timeline({ delay: 0.8 });

      // Magical particle entrance
      tl.to(particles, {
        duration: 2.0,
        opacity: 0.7, // More visible
        scale: 1,
        rotation: 360,
        ease: "power2.out",
        stagger: {
          amount: 1.5,
          from: "random"
        }
      })
      
      // First line - magical materialization (slower)
      .to(line1Ref.current, {
        duration: 2.0,
        opacity: 1,
        scale: 1,
        rotationY: 0,
        ease: "power3.out",
        onStart: () => {
          // Enhanced magical glow
          gsap.fromTo(line1Ref.current,
            { filter: "drop-shadow(0 0 25px rgba(20, 184, 166, 0.8)) drop-shadow(0 0 40px rgba(20, 184, 166, 0.4))" },
            { 
              filter: "drop-shadow(0 0 5px rgba(20, 184, 166, 0.3))",
              duration: 2.5,
              ease: "power2.out"
            }
          );
        }
      }, "-=1.0")
      
      // Second line - enchanted manifestation (slower)
      .to(line2Ref.current, {
        duration: 1.8,
        opacity: 1,
        scale: 1,
        rotationY: 0,
        ease: "back.out(1.4)",
        onStart: () => {
          // Enhanced magical glow
          gsap.fromTo(line2Ref.current,
            { filter: "drop-shadow(0 0 30px rgba(71, 85, 105, 0.9)) drop-shadow(0 0 50px rgba(71, 85, 105, 0.5))" },
            { 
              filter: "drop-shadow(0 0 8px rgba(71, 85, 105, 0.4))",
              duration: 3.0,
              ease: "power2.out"
            }
          );
        }
      }, "-=0.8");

      // Enhanced particles emanating from text animation
      particles.forEach((particle, i) => {
        const isFromFirstLine = i < 15;
        const isLargeStar = particle.innerHTML === '✦';
        const isTinyStar = particle.innerHTML && particle.innerHTML !== '✦' && particle.style.fontSize;
        
        // Constrained emanating movement - stays around title area
        gsap.to(particle, {
          y: isFromFirstLine ? `+=${Math.random() * 20 + 10}` : `+=${Math.random() * 15 + 8}`, // Reduced upward movement
          x: `+=${(Math.random() - 0.5) * 40}`, // Reduced horizontal spread
          rotation: (isLargeStar || isTinyStar) ? `+=${Math.random() * 720 + 360}` : `+=${Math.random() * 360}`, // All stars spin more
          duration: Math.random() * 8 + 5, // Even slower, more majestic
          repeat: -1,
          yoyo: true,
          ease: "sine.inOut",
          delay: Math.random() * 3
        });
        
        // Toned down magical pulsing - 30% reduction
        const baseOpacity = isFromFirstLine ? 0.56 : 0.49; // Reduced base opacity by 30%
        gsap.to(particle, {
          opacity: baseOpacity + Math.random() * 0.21, // Reduced from 0.3 to 0.21
          scale: (isLargeStar || isTinyStar) ? 0.6 + Math.random() * 0.8 : 0.7 + Math.random() * 0.6, // All stars pulse more dramatically
          duration: Math.random() * 4 + 2,
          repeat: -1,
          yoyo: true,
          ease: "sine.inOut",
          delay: Math.random() * 4
        });
        
        // Reduced sparkle effect - 30% less intense
        if (Math.random() > 0.6) { // Reduced frequency from 50% to 40%
          if (isLargeStar || isTinyStar) {
            // Toned down twinkling for all stars
            gsap.to(particle, {
              textShadow: isFromFirstLine 
                ? '0 0 10px rgba(20, 184, 166, 0.7), 0 0 18px rgba(20, 184, 166, 0.56), 0 0 25px rgba(20, 184, 166, 0.42)' // Reduced by 30%
                : '0 0 10px rgba(71, 85, 105, 0.7), 0 0 18px rgba(71, 85, 105, 0.56), 0 0 25px rgba(71, 85, 105, 0.42)',
              duration: 0.3,
              repeat: -1,
              yoyo: true,
              ease: "power2.inOut",
              delay: Math.random() * 6
            });
          } else {
            // Reduced glow for orbs
            gsap.to(particle, {
              filter: isFromFirstLine 
                ? "brightness(1.26) drop-shadow(0 0 6px rgba(20, 184, 166, 0.7)) drop-shadow(0 0 11px rgba(20, 184, 166, 0.42))" // Reduced by 30%
                : "brightness(1.12) drop-shadow(0 0 6px rgba(71, 85, 105, 0.7)) drop-shadow(0 0 11px rgba(71, 85, 105, 0.42))",
              duration: 0.4,
              repeat: -1,
              yoyo: true,
              ease: "power2.inOut",
              delay: Math.random() * 5
            });
          }
        }
      });
    }
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

      {/* Magical Particles */}
      <div 
        ref={particlesRef}
        className="absolute inset-0 z-5 pointer-events-none"
        style={{ transform: `translateY(${scrollY * 0.1}px)` }}
      />

      {/* Content Overlay */}
      <div className="relative z-10 h-full flex flex-col items-center justify-center text-center px-4" style={{paddingTop: 'calc(180px + 40px)'}}>
        <div className="relative">
          <h1 
            ref={titleRef}
            className="font-cinzel font-bold text-gray-800 mb-6 md:mb-8 tracking-wide drop-shadow-sm text-4xl md:text-6xl lg:text-7xl xl:text-8xl"
            style={{fontSize: 'clamp(48px, 8vw, 80px)'}}
          >
          <span 
            ref={line1Ref}
            className="bg-gradient-to-r from-teal-600 via-cyan-700 to-teal-800 bg-clip-text text-transparent inline-block"
          >
            ASCEND BEYOND
          </span>
          <br />
          <span 
            ref={line2Ref}
            className="bg-gradient-to-r from-slate-700 via-slate-800 to-slate-900 bg-clip-text text-transparent inline-block"
          >
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

        <div className="flex justify-center">
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
        </div>
      </div>
    </section>
  );
}