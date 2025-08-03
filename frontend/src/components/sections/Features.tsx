import { useRef, useEffect, useState, useLayoutEffect } from 'react';
import { gsap } from 'gsap';

export default function Features() {
  const card1Ref = useRef<HTMLDivElement>(null);
  const card2Ref = useRef<HTMLDivElement>(null);
  const card3Ref = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [selectedCard, setSelectedCard] = useState<string | null>(null);
  const [isAnimating, setIsAnimating] = useState(false);
  const animationTimelineRef = useRef<gsap.core.Timeline | null>(null);
  const [isDesktop, setIsDesktop] = useState(false);
  const [scrollY, setScrollY] = useState(0);
  const forestRef = useRef<HTMLDivElement>(null);

  const cards = [
    {
      id: 'magician',
      title: 'The Magician',
      frontImage: '/src/assets/img/magician.png',
      backImage: '/src/assets/img/card_background.png',
      ref: card1Ref,
      description: {
        story: "In the ancient halls of the Arcane Academy, the Magician stands as the master of manifestation. With the power to transform thought into reality, this card represents the bridge between the spiritual and material worlds. When drawn, it signifies the awakening of your latent magical abilities and the beginning of your journey into the mystical arts.",
        abilities: "The Magician grants you the power to manipulate the elements, cast powerful spells, and bend reality to your will. Your magical prowess grows with each challenge overcome, unlocking new abilities and deeper understanding of the arcane forces that shape our world."
      }
    },
    {
      id: 'strength',
      title: 'Strength',
      frontImage: '/src/assets/img/strenght.png',
      backImage: '/src/assets/img/card_background.png',
      ref: card2Ref,
      description: {
        story: "The Strength card embodies the courage that lies within every warrior's heart. Not merely physical might, but the inner fortitude to face impossible odds and emerge victorious. This card tells the tale of those who have mastered their fears and channeled their determination into unstoppable force.",
        abilities: "Strength enhances your combat abilities, grants resistance to magical attacks, and allows you to perform feats of legendary prowess. Your willpower becomes your greatest weapon, enabling you to overcome any obstacle that stands between you and your destiny."
      }
    },
    {
      id: 'highpriestess',
      title: 'The High Priestess',
      frontImage: '/src/assets/img/highpriestess.png',
      backImage: '/src/assets/img/card_background.png',
      ref: card3Ref,
      description: {
        story: "The High Priestess guards the sacred knowledge of the ancients, her wisdom spanning countless generations. She represents the intuitive understanding that transcends logic, the deep connection to the mystical forces that govern the universe. Her presence heralds the revelation of hidden truths and the awakening of your spiritual sight.",
        abilities: "The High Priestess grants you the gift of foresight, allowing you to sense danger before it strikes and understand the deeper meaning behind events. Your intuition becomes a powerful tool, guiding you through the darkest of times and revealing the path to enlightenment."
      }
    }
  ];

  // Initialize cards with proper 3D setup
  useLayoutEffect(() => {
    cards.forEach((card, index) => {
      if (card.ref.current) {
        const cardElement = card.ref.current;
        const cardContainer = cardElement.querySelector('.card-container') as HTMLElement;
        const front = cardElement.querySelector('.card-front') as HTMLElement;
        const back = cardElement.querySelector('.card-back') as HTMLElement;
        
        if (cardContainer && front && back) {
          // Set up 3D container
          gsap.set(cardElement, {
            perspective: 1200,
            perspectiveOrigin: "50% 50%"
          });
          
          gsap.set(cardContainer, {
            transformStyle: "preserve-3d",
            transformOrigin: "50% 50%",
            rotationY: 0
          });
          
          // Set up card faces
          gsap.set([front, back], {
            backfaceVisibility: "hidden",
            position: "absolute",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%"
          });
          
          // Initial state: showing back (container rotated to show back face)
          gsap.set(front, { rotationY: 0 });      // Front face normal
          gsap.set(back, { rotationY: 180 });     // Back face flipped 
          gsap.set(cardContainer, { rotationY: 180 }); // Container showing back initially
          
          // Initial positioning with constrained fan spacing
          gsap.set(cardElement, {
            rotation: (index - 1) * 12, // Reduced rotation for tighter grouping
            x: (index - 1) * 35,       // Reduced spacing for better containment
            zIndex: 3 - index
          });
        }
      }
    });
  }, []);

  // Track screen size for responsive animations
  useEffect(() => {
    const updateScreenSize = () => {
      setIsDesktop(window.innerWidth >= 1024);
    };

    updateScreenSize();
    window.addEventListener('resize', updateScreenSize);

    return () => {
      window.removeEventListener('resize', updateScreenSize);
    };
  }, []);

  // Parallax scroll effect for forest background
  useEffect(() => {
    const handleScroll = () => {
      if (forestRef.current) {
        const rect = forestRef.current.getBoundingClientRect();
        const windowHeight = window.innerHeight;
        const elementTop = rect.top;
        const elementBottom = rect.bottom;
        
        // Check if element is in viewport
        if (elementBottom >= 0 && elementTop <= windowHeight) {
          // Calculate relative scroll position
          const scrollProgress = (windowHeight - elementTop) / (windowHeight + rect.height);
          const parallaxOffset = scrollProgress * 200 - 100; // Adjust multiplier for effect strength
          setScrollY(parallaxOffset);
        }
      }
    };

    window.addEventListener('scroll', handleScroll);
    handleScroll(); // Initial calculation

    return () => {
      window.removeEventListener('scroll', handleScroll);
    };
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (animationTimelineRef.current) {
        animationTimelineRef.current.kill();
      }
      gsap.killTweensOf("*");
    };
  }, []);

  const handleCardHover = (cardRef: React.RefObject<HTMLDivElement>, isHover: boolean) => {
    if (!cardRef.current || selectedCard || isAnimating) return;

    const cardElement = cardRef.current;
    const cardIndex = cards.findIndex(card => card.ref === cardRef);

    if (isHover) {
      gsap.to(cardElement, {
        scale: 1.05, // Reduced hover scale to prevent boundary overflow
        y: -8,       // Reduced hover lift
        zIndex: 100,
        duration: 0.3,
        ease: "power2.out"
      });
    } else {
      gsap.to(cardElement, {
        scale: 1,
        y: 0,
        zIndex: 3 - cardIndex,
        duration: 0.3,
        ease: "power2.out"
      });
    }
  };

  const handleCardClick = (cardId: string) => {
    if (isAnimating) return;

    if (selectedCard === cardId) {
      setSelectedCard(null);
      resetToOriginalLayout();
    } else {
      setSelectedCard(cardId);
      animateToSelectedLayout(cardId);
    }
  };

  const handleCardKeyDown = (event: React.KeyboardEvent, cardId: string) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleCardClick(cardId);
    }
  };

  const animateToSelectedLayout = (selectedCardId: string) => {
    if (animationTimelineRef.current) {
      animationTimelineRef.current.kill();
    }

    setIsAnimating(true);
    const tl = gsap.timeline({
      onComplete: () => setIsAnimating(false)
    });
    animationTimelineRef.current = tl;

    cards.forEach((card) => {
      if (!card.ref.current) return;
      
      const cardElement = card.ref.current;
      const cardContainer = cardElement.querySelector('.card-container') as HTMLElement;
      
      if (card.id === selectedCardId) {
        // Selected card moves to right and flips to front - constrained positioning
        tl.to(cardElement, {
          x: isDesktop ? 150 : 30, // Reduced movement to stay within bounds
          y: isDesktop ? 0 : -20,   // Reduced vertical movement
          scale: isDesktop ? 1.05 : 1.1, // Slightly smaller scale
          rotation: 0,
          zIndex: 1000,
          duration: 0.8,
          ease: "power2.out"
        }, 0)
        .to(cardContainer, {
          rotationY: 0,
          duration: 0.6,
          ease: "power2.inOut"
        }, 0.2);
      } else {
        // Other cards move to left and stay as backs - constrained positioning
        const stackIndex = cards.filter(c => c.id !== selectedCardId).findIndex(c => c.id === card.id);
        tl.to(cardElement, {
          x: isDesktop ? -150 - (stackIndex * 25) : -80 - (stackIndex * 20), // Reduced movement
          y: isDesktop ? 30 + (stackIndex * 6) : 10 + (stackIndex * 6),      // Reduced vertical offset
          scale: isDesktop ? 0.8 : 0.75,  // Slightly larger to remain visible
          rotation: stackIndex * -3,      // Reduced rotation
          zIndex: 100 - stackIndex,
          duration: 0.8,
          ease: "power2.out"
        }, 0);
        
        // If card is showing front, flip it back
        const currentRotation = gsap.getProperty(cardContainer, "rotationY") as number || 0;
        if (Math.abs(currentRotation) > 90) {
          tl.to(cardContainer, {
            rotationY: 0,
            duration: 0.6,
            ease: "power2.inOut"
          }, 0);
        }
      }
    });
  };

  const resetToOriginalLayout = () => {
    if (animationTimelineRef.current) {
      animationTimelineRef.current.kill();
    }

    setIsAnimating(true);
    const tl = gsap.timeline({
      onComplete: () => setIsAnimating(false)
    });
    animationTimelineRef.current = tl;

    cards.forEach((card, index) => {
      if (!card.ref.current) return;
      
      const cardElement = card.ref.current;
      const cardContainer = cardElement.querySelector('.card-container') as HTMLElement;
      
      // Return to original position - using same constrained values as initial setup
      tl.to(cardElement, {
        x: (index - 1) * 35,  // Match initial positioning
        y: 0,
        scale: 1,
        rotation: (index - 1) * 12, // Match initial rotation
        zIndex: 3 - index,
        duration: 0.8,
        ease: "power2.out"
      }, 0);
      
      // Flip all cards back to showing backs
      tl.to(cardContainer, {
        rotationY: 180,
        duration: 0.6,
        ease: "power2.inOut"
      }, 0);
    });
  };

  return (
    <section className="pt-24 pb-0 bg-white relative" style={{ contain: 'layout' }}>
      <div className="container mx-auto px-4 relative">
        {/* Divider - same as under H1 */}
        <div className="flex items-center justify-center mb-16">
          <div className="h-px bg-gray-400 w-24"></div>
          <div className="mx-4 text-gray-600 text-2xl">❦</div>
          <div className="h-px bg-gray-400 w-24"></div>
        </div>

        {/* Introductory Text */}
        <div className="text-center mb-20">
          <h2 className="text-4xl md:text-5xl font-cinzel font-bold text-gray-800 mb-8 bg-gradient-to-r from-teal-600 to-cyan-700 bg-clip-text text-transparent">
            Discover Your Path
          </h2>
          <p className="text-lg md:text-xl text-gray-700 font-cormorant max-w-4xl mx-auto leading-relaxed">
            Embark on a mystical journey where ancient wisdom meets modern gameplay. Choose your arcane path, 
            master powerful spells, and uncover the secrets hidden within the cards of fate. Each decision shapes 
            your destiny in this immersive world of magic and mystery.
          </p>
        </div>

        {/* Card Gallery and Description Container */}
        <div 
          className="relative mb-20"
          role="main"
          aria-label="Interactive tarot card gallery"
          style={{ minHeight: '600px', overflow: 'visible' }}
        >
          {/* Screen reader announcement for dynamic content */}
          <div 
            aria-live="polite" 
            aria-atomic="true" 
            className="sr-only"
          >
            {selectedCard && `${cards.find(card => card.id === selectedCard)?.title} card selected. Description is now visible.`}
          </div>

          {/* Main Content Container - Side by Side Layout */}
          <div className={`flex ${selectedCard ? 'flex-col lg:flex-row lg:justify-between lg:items-start gap-8 lg:gap-12' : 'justify-center'} transition-all duration-800 relative`} style={{ minHeight: selectedCard ? '600px' : '500px', overflow: 'visible' }}>
            {/* Card Gallery */}
            <div 
              ref={containerRef}
              className={`${selectedCard ? 'flex-shrink-0 order-1 lg:order-none' : 'flex justify-center items-center'} relative`}
              style={{ 
                minHeight: '500px', 
                width: selectedCard ? (isDesktop ? '600px' : '100%') : '100%',
                maxWidth: '100%',
                overflow: 'visible'
              }}
              onClick={(e) => {
                // If clicking on the container (not on a card), reset to original state
                if (e.target === e.currentTarget && selectedCard) {
                  setSelectedCard(null);
                  resetToOriginalLayout();
                }
              }}
            >
              <div className="relative flex items-center justify-center" style={{ minHeight: '450px', paddingTop: '24px', overflow: 'visible' }}>
                {cards.map((card, index) => (
                  <div
                    key={card.id}
                    ref={card.ref}
                    className={`absolute cursor-pointer focus:outline-none rounded-lg ${
                      selectedCard === card.id 
                        ? 'ring-2 ring-teal-400/30' 
                        : 'focus:ring-2 focus:ring-teal-400/40'
                    } ${isAnimating ? 'pointer-events-none' : ''}`}
                    style={{
                      left: '50%',
                      top: selectedCard ? '40%' : '50%',
                      transform: 'translate(-50%, -50%)',
                      zIndex: selectedCard === card.id ? 1000 : (3 - index),
                      overflow: 'visible'
                    }}
                    role="button"
                    tabIndex={isAnimating ? -1 : 0}
                    aria-label={`${card.title} tarot card. ${selectedCard === card.id ? 'Currently selected' : 'Click or press Enter to select'}`}
                    aria-describedby={selectedCard === card.id ? `${card.id}-description` : undefined}
                    aria-expanded={selectedCard === card.id}
                    onMouseEnter={() => handleCardHover(card.ref, true)}
                    onMouseLeave={() => handleCardHover(card.ref, false)}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCardClick(card.id);
                    }}
                    onKeyDown={(e) => handleCardKeyDown(e, card.id)}
                  >
                    {/* Card Container - Original sizing restored */}
                    <div className="relative w-48 h-72 md:w-56 md:h-84 lg:w-64 lg:h-96" style={{ overflow: 'visible' }}>
                      {/* Inner card container for 3D flip */}
                      <div 
                        className="card-container relative w-full h-full rounded-lg"
                        style={{ 
                          transformStyle: 'preserve-3d',
                          perspective: '1200px',
                          perspectiveOrigin: '50% 50%',
                          overflow: 'visible'
                        }}
                      >
                        {/* Card Back */}
                        <div 
                          className="card-back absolute inset-0"
                          style={{ backfaceVisibility: 'hidden' }}
                        >
                          <img
                            src={card.backImage}
                            alt={`${card.title} card back - mysterious tarot card design`}
                            className="w-full h-full object-cover rounded-lg shadow-xl"
                            draggable={false}
                          />
                        </div>
                        
                        {/* Card Front */}
                        <div 
                          className="card-front absolute inset-0"
                          style={{ backfaceVisibility: 'hidden' }}
                        >
                          <img
                            src={card.frontImage}
                            alt={`${card.title} tarot card - ${card.description.story.substring(0, 100)}...`}
                            className="w-full h-full object-cover rounded-lg shadow-xl"
                            draggable={false}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Card Description Section - Side by Side */}
            {selectedCard && (
              <div className="flex-1 max-w-full lg:max-w-2xl order-2 lg:order-none mt-8 lg:mt-0">
                <div className="p-6 lg:p-8" id={`${selectedCard}-description`} role="region" aria-label="Card description">
                  <h3 className="text-3xl md:text-4xl font-cinzel font-bold mb-6 bg-gradient-to-r from-teal-600 to-cyan-700 bg-clip-text text-transparent">
                    {cards.find(card => card.id === selectedCard)?.title}
                  </h3>
                  
                  <div className="space-y-6">
                    {/* Story Section */}
                    <div>
                      <h4 className="text-xl font-cinzel font-bold mb-4 flex items-center" id={`${selectedCard}-story-heading`}>
                        <span className="text-teal-600 mr-3 text-xl" aria-hidden="true">❦</span>
                        <span className="text-gray-800">The Tale</span>
                      </h4>
                      <p className="text-gray-700 font-cormorant leading-relaxed text-lg" aria-labelledby={`${selectedCard}-story-heading`}>
                        {cards.find(card => card.id === selectedCard)?.description.story}
                      </p>
                    </div>

                    {/* Abilities Section */}
                    <div>
                      <h4 className="text-xl font-cinzel font-bold mb-4 flex items-center" id={`${selectedCard}-abilities-heading`}>
                        <span className="text-teal-600 mr-3 text-xl" aria-hidden="true">⚡</span>
                        <span className="text-gray-800">Arcane Powers</span>
                      </h4>
                      <p className="text-gray-700 font-cormorant leading-relaxed text-lg" aria-labelledby={`${selectedCard}-abilities-heading`}>
                        {cards.find(card => card.id === selectedCard)?.description.abilities}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>

        </div>
      </div>

      {/* Arcane Mastery Section - Full Width */}
      <div 
        ref={forestRef}
        className="relative mt-24 mb-0 overflow-hidden"
        style={{
          minHeight: '800px'
        }}
      >
        {/* Parallax Background */}
        <div 
          className="absolute inset-0 w-full"
          style={{
            backgroundImage: 'url(/src/assets/img/forest2.png)',
            backgroundSize: 'cover',
            backgroundPosition: 'center top',
            backgroundRepeat: 'no-repeat',
            transform: `translateY(${scrollY * 0.5}px)`,
            height: '120%',
            top: '-10%'
          }}
        />
        {/* Content container - centered without overlay */}
        <div className="relative min-h-[800px] flex items-center justify-center py-16">
          <div className="container mx-auto px-4 text-center">
              {/* Section title */}
              <div className="text-center mb-12">
                <h3 className="text-3xl md:text-4xl font-cinzel font-bold mb-4 bg-gradient-to-r from-teal-600 to-cyan-700 bg-clip-text text-transparent">
                  Arcane Mastery
                </h3>
                <div className="flex items-center justify-center mb-6">
                  <div className="h-px bg-gray-400 w-24"></div>
                  <div className="mx-4 text-gray-600 text-2xl">❦</div>
                  <div className="h-px bg-gray-400 w-24"></div>
                </div>
                <p className="text-lg md:text-xl text-gray-700 font-cormorant max-w-4xl mx-auto leading-relaxed">
                  Unlock the ancient secrets and master the mystical arts that await your discovery
                </p>
              </div>

                             {/* Feature grid */}
                 <div className="grid grid-cols-1 md:grid-cols-3 py-8" style={{gap: '0px'}}>
                   <div className="text-center group p-0">
                     <div className="mb-6">
                       <img 
                         src="/src/assets/img/spellbook.png" 
                         alt="Spellbook" 
                         className="w-16 h-16 mx-auto mb-4 group-hover:scale-110 transition-transform duration-300"
                       />
                     </div>
                     <h4 className="text-xl font-cinzel font-bold text-gray-900 mb-4 group-hover:text-teal-600 transition-colors duration-300 drop-shadow-sm">
                       Mystical Arts
                     </h4>
                     <p className="text-gray-800 font-cormorant leading-relaxed drop-shadow-sm max-w-xs mx-auto">
                       Master ancient spells and unlock the secrets of arcane knowledge through intuitive gameplay.
                     </p>
                   </div>
                   
                   <div className="text-center group p-0">
                     <div className="mb-6">
                       <img 
                         src="/src/assets/img/dungeon.png" 
                         alt="Dungeon" 
                         className="w-16 h-16 mx-auto mb-4 group-hover:scale-110 transition-transform duration-300"
                       />
                     </div>
                     <h4 className="text-xl font-cinzel font-bold text-gray-900 mb-4 group-hover:text-teal-600 transition-colors duration-300 drop-shadow-sm">
                       Epic Dungeons
                     </h4>
                     <p className="text-gray-800 font-cormorant leading-relaxed drop-shadow-sm max-w-xs mx-auto">
                       Explore mysterious realms filled with challenging puzzles and hidden treasures.
                     </p>
                   </div>
                   
                   <div className="text-center group p-0">
                     <div className="mb-6">
                       <img 
                         src="/src/assets/img/destiny.png" 
                         alt="Destiny" 
                         className="w-16 h-16 mx-auto mb-4 group-hover:scale-110 transition-transform duration-300"
                       />
                     </div>
                     <h4 className="text-xl font-cinzel font-bold text-gray-900 mb-4 group-hover:text-teal-600 transition-colors duration-300 drop-shadow-sm">
                       Forge Your Destiny
                     </h4>
                     <p className="text-gray-800 font-cormorant leading-relaxed drop-shadow-sm max-w-xs mx-auto">
                       Every choice influences your path in this dynamic world of consequence and adventure.
                     </p>
                   </div>
                 </div>
          </div>
        </div>
      </div>
    </section>
  );
}