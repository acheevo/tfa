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
      for (let i = 0; i < 12; i++) { // Reduced from 25 to 12 particles (50% reduction)
        const particle = document.createElement('div');
        particle.className = 'absolute pointer-events-none';
        const size = Math.random() * 8 + 4; // Larger, more visible particles (4-12px)
        
        // Create different particle shapes
        const shapeType = Math.random();
        const isFromFirstLine = i < 7; // Adjusted for reduced particle count (was 15, now 7)
        
        if (shapeType < 0.4) {
          // Ethereal wisps - more mystical than stars
          particle.innerHTML = '~';
          particle.style.fontSize = (size * 1.2) + 'px';
          particle.style.textAlign = 'center';
          particle.style.lineHeight = '1';
          particle.style.color = isFromFirstLine 
            ? `rgba(20, 184, 166, ${0.3 + Math.random() * 0.2})` // Teal/cyan for first line
            : `rgba(51, 65, 85, ${0.3 + Math.random() * 0.2})`; // Slate for second line
          particle.style.textShadow = isFromFirstLine
            ? '0 0 8px rgba(20, 184, 166, 0.4), 0 0 12px rgba(20, 184, 166, 0.2)' // Teal/cyan glow
            : '0 0 8px rgba(51, 65, 85, 0.4), 0 0 12px rgba(51, 65, 85, 0.2)'; // Slate glow
        } else if (shapeType < 0.7) {
          // Mystical runes - fantasy symbols
          const runeSymbols = ['◊', '◇', '○', '●', '◐', '◑'];
          particle.innerHTML = runeSymbols[Math.floor(Math.random() * runeSymbols.length)];
          particle.style.fontSize = (size * 0.8) + 'px';
          particle.style.textAlign = 'center';
          particle.style.lineHeight = '1';
          particle.style.color = isFromFirstLine 
            ? `rgba(20, 184, 166, ${0.25 + Math.random() * 0.15})` // Teal/cyan for first line
            : `rgba(51, 65, 85, ${0.25 + Math.random() * 0.15})`; // Slate for second line
          particle.style.textShadow = isFromFirstLine
            ? '0 0 6px rgba(20, 184, 166, 0.3), 0 0 10px rgba(20, 184, 166, 0.15)' // Teal/cyan glow
            : '0 0 6px rgba(51, 65, 85, 0.3), 0 0 10px rgba(51, 65, 85, 0.15)'; // Slate glow
        } else {
          // Ethereal orbs - more mystical than bright
          particle.style.width = size + 'px';
          particle.style.height = size + 'px';
          particle.style.borderRadius = '50%';
          particle.style.background = isFromFirstLine
            ? `radial-gradient(circle at 40% 40%, rgba(255, 255, 255, 0.2) 0%, rgba(20, 184, 166, 0.3) 40%, rgba(20, 184, 166, 0.15) 80%, transparent 100%)` // Teal/cyan gradient
            : `radial-gradient(circle at 40% 40%, rgba(255, 255, 255, 0.15) 0%, rgba(51, 65, 85, 0.3) 40%, rgba(51, 65, 85, 0.15) 80%, transparent 100%)`; // Slate gradient
          particle.style.boxShadow = isFromFirstLine
            ? '0 0 8px rgba(20, 184, 166, 0.3), inset 0 0 4px rgba(255, 255, 255, 0.1)' // Teal/cyan glow
            : '0 0 8px rgba(51, 65, 85, 0.3), inset 0 0 4px rgba(255, 255, 255, 0.1)'; // Slate glow
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

      // Create magical reveal timeline - immediate start
      const tl = gsap.timeline({ delay: 0 });

      // Magical particle entrance - starts immediately
      tl.to(particles, {
        duration: 0.6,
        opacity: 0.7, // More visible
        scale: 1,
        rotation: 360,
        ease: "power2.out",
        stagger: {
          amount: 0.4,
          from: "random"
        }
      })
      
      // First line - magical materialization (slower and more magical)
      .to(line1Ref.current, {
        duration: 1.3,
        opacity: 1,
        scale: 1,
        rotationY: 0,
        ease: "power3.out",
        onStart: () => {
          // Ethereal magical glow - more mystical than bright
          gsap.fromTo(line1Ref.current,
            { filter: "drop-shadow(0 0 20px rgba(20, 184, 166, 0.4)) drop-shadow(0 0 30px rgba(20, 184, 166, 0.2))" },
            { 
              filter: "drop-shadow(0 0 3px rgba(20, 184, 166, 0.2))",
              duration: 2.0,
              ease: "power2.out"
            }
          );
        }
      }, "-=0.3")
      
      // Second line - enchanted manifestation (starts 0.2s after first line)
      .to(line2Ref.current, {
        duration: 0.7,
        opacity: 1,
        scale: 1,
        rotationY: 0,
        ease: "back.out(1.4)",
        onStart: () => {
          // Ethereal magical glow - more mystical than bright
          gsap.fromTo(line2Ref.current,
            { filter: "drop-shadow(0 0 25px rgba(71, 85, 105, 0.4)) drop-shadow(0 0 35px rgba(71, 85, 105, 0.2))" },
            { 
              filter: "drop-shadow(0 0 4px rgba(71, 85, 105, 0.2))",
              duration: 1.8,
              ease: "power2.out"
            }
          );
        }
      }, "+=0.2");

      // Enhanced particles emanating from text animation
      particles.forEach((particle, i) => {
        const isFromFirstLine = i < 7; // Adjusted for reduced particle count (was 15, now 7)
        
        // Ethereal floating movement - more mystical than mechanical
        gsap.to(particle, {
          y: isFromFirstLine ? `+=${Math.random() * 15 + 8}` : `+=${Math.random() * 12 + 6}`, // Gentler floating
          x: `+=${(Math.random() - 0.5) * 25}`, // Subtle horizontal drift
          rotation: `+=${Math.random() * 180 + 90}`, // Gentle rotation
          duration: Math.random() * 10 + 8, // Slower, more ethereal
          repeat: -1,
          yoyo: true,
          ease: "sine.inOut",
          delay: Math.random() * 4
        });
        
        // Ethereal breathing effect - more mystical than pulsing
        const baseOpacity = isFromFirstLine ? 0.3 : 0.25; // Much more subtle
        gsap.to(particle, {
          opacity: baseOpacity + Math.random() * 0.15, // Very gentle variation
          scale: 0.8 + Math.random() * 0.4, // Subtle breathing
          duration: Math.random() * 6 + 4, // Slower, more mystical
          repeat: -1,
          yoyo: true,
          ease: "sine.inOut",
          delay: Math.random() * 5
        });
        
        // Mystical shimmer effect - very subtle and ethereal
        if (Math.random() > 0.7) { // Rare mystical shimmer
          if (particle.innerHTML === '~') {
            // Ethereal wisp shimmer
            gsap.to(particle, {
              textShadow: isFromFirstLine 
                ? '0 0 12px rgba(20, 184, 166, 0.3), 0 0 20px rgba(20, 184, 166, 0.15)' // Teal/cyan shimmer
                : '0 0 12px rgba(51, 65, 85, 0.3), 0 0 20px rgba(51, 65, 85, 0.15)', // Slate shimmer
              duration: 2.0,
              repeat: -1,
              yoyo: true,
              ease: "sine.inOut",
              delay: Math.random() * 8
            });
          } else if (particle.style.background) {
            // Ethereal orb shimmer
            gsap.to(particle, {
              filter: isFromFirstLine 
                ? "brightness(1.1) drop-shadow(0 0 8px rgba(20, 184, 166, 0.2))" // Teal/cyan shimmer
                : "brightness(1.05) drop-shadow(0 0 8px rgba(51, 65, 85, 0.2))", // Slate shimmer
              duration: 3.0,
              repeat: -1,
              yoyo: true,
              ease: "sine.inOut",
              delay: Math.random() * 6
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
            style={{fontSize: 'clamp(48px, 8vw, 80px)', paddingTop: '24px', paddingBottom: '24px'}}
          >
          <span 
            ref={line1Ref}
            className="bg-gradient-to-r from-teal-600 via-cyan-700 to-teal-800 bg-clip-text text-transparent inline-block"
          >
            Ascend Beyond
          </span>
          <br />
          <span 
            ref={line2Ref}
            className="bg-gradient-to-r from-slate-700 via-slate-800 to-slate-900 bg-clip-text text-transparent inline-block"
          >
            Fate
          </span>
        </h1>
        
        <div className="mb-8">
          <div className="flex items-center justify-center mb-4">
            <div className="h-px bg-gray-400 w-24"></div>
            <div className="mx-4 text-gray-600 text-2xl">❦</div>
            <div className="h-px bg-gray-400 w-24"></div>
          </div>
          <p className="text-lg md:text-xl text-gray-700 font-cormorant max-w-3xl mx-auto">
            Master ancient arts, explore mystical dungeons, and forge your destiny in a world where every choice shapes your arcane path
          </p>
        </div>

        <div className="flex justify-center">
          <button 
            className="relative font-cormorant font-bold text-2xl"
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
            Join Beta
          </button>
        </div>
        </div>
      </div>
    </section>
  );
}