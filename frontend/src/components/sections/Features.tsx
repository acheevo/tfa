import { useRef, useEffect } from 'react';
import { gsap } from 'gsap';

export default function Features() {
  const card1Ref = useRef<HTMLDivElement>(null);
  const card2Ref = useRef<HTMLDivElement>(null);
  const card3Ref = useRef<HTMLDivElement>(null);

  const cards = [
    {
      id: 'magician',
      frontImage: '/src/assets/img/magician.png',
      backImage: '/src/assets/img/card_background.png',
      ref: card1Ref
    },
    {
      id: 'strength',
      frontImage: '/src/assets/img/strenght.png',
      backImage: '/src/assets/img/card_background.png',
      ref: card2Ref
    },
    {
      id: 'highpriestess',
      frontImage: '/src/assets/img/highpriestess.png',
      backImage: '/src/assets/img/card_background.png',
      ref: card3Ref
    }
  ];

  useEffect(() => {
    // Set initial card states
    cards.forEach((card) => {
      if (card.ref.current) {
        gsap.set(card.ref.current.querySelector('.card-front'), {
          rotationY: 180,
          opacity: 0
        });
        gsap.set(card.ref.current.querySelector('.card-back'), {
          rotationY: 0,
          opacity: 1
        });
      }
    });
  }, []);

  const handleCardHover = (cardRef: React.RefObject<HTMLDivElement>, isHover: boolean) => {
    if (!cardRef.current) return;

    const front = cardRef.current.querySelector('.card-front');
    const back = cardRef.current.querySelector('.card-back');

    if (isHover) {
      // Flip to front
      gsap.to(back, {
        rotationY: 180,
        opacity: 0,
        duration: 0.3,
        ease: "power2.inOut"
      });
      gsap.to(front, {
        rotationY: 0,
        opacity: 1,
        duration: 0.3,
        ease: "power2.inOut",
        delay: 0.1
      });
    } else {
      // Flip to back
      gsap.to(front, {
        rotationY: 180,
        opacity: 0,
        duration: 0.3,
        ease: "power2.inOut"
      });
      gsap.to(back, {
        rotationY: 0,
        opacity: 1,
        duration: 0.3,
        ease: "power2.inOut",
        delay: 0.1
      });
    }
  };

  return (
    <section className="py-24 bg-white">
      <div className="container mx-auto px-4">
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
          <p className="text-lg md:text-xl text-gray-700 font-cinzel max-w-4xl mx-auto leading-relaxed">
            Embark on a mystical journey where ancient wisdom meets modern gameplay. Choose your arcane path, 
            master powerful spells, and uncover the secrets hidden within the cards of fate. Each decision shapes 
            your destiny in this immersive world of magic and mystery.
          </p>
        </div>

        {/* Card Gallery */}
        <div className="flex justify-center items-center relative mb-20">
          <div className="flex items-center" style={{ gap: '-120px' }}> {/* Overlapping cards */}
            {cards.map((card, index) => (
              <div
                key={card.id}
                ref={card.ref}
                className="relative cursor-pointer transform transition-transform duration-300 hover:scale-105 hover:z-10"
                style={{
                  zIndex: 3 - index, // First card on top, decreasing z-index
                  transform: `rotate(${(index - 1) * 8}deg)`, // Slight rotation for organic feel
                  marginLeft: index > 0 ? '-120px' : '0' // Manual overlap
                }}
                onMouseEnter={() => handleCardHover(card.ref, true)}
                onMouseLeave={() => handleCardHover(card.ref, false)}
              >
                {/* Card Container */}
                <div className="relative w-48 h-72 md:w-56 md:h-84 lg:w-64 lg:h-96" style={{ perspective: '1000px' }}>
                  {/* Card Back */}
                  <div className="card-back absolute inset-0" style={{ backfaceVisibility: 'hidden' }}>
                    <img
                      src={card.backImage}
                      alt="Card Back"
                      className="w-full h-full object-cover rounded-lg shadow-xl"
                    />
                  </div>
                  
                  {/* Card Front */}
                  <div className="card-front absolute inset-0" style={{ backfaceVisibility: 'hidden' }}>
                    <img
                      src={card.frontImage}
                      alt="Tarot Card"
                      className="w-full h-full object-cover rounded-lg shadow-xl"
                    />
                  </div>
                </div>

                {/* Magical glow effect on hover */}
                <div className="absolute inset-0 rounded-lg bg-gradient-to-r from-teal-400/20 to-cyan-400/20 opacity-0 hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>
              </div>
            ))}
          </div>
        </div>

        {/* Feature highlights */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          <div className="text-center p-6">
            <div className="text-3xl mb-4 text-teal-600">🔮</div>
            <h3 className="text-xl font-cinzel font-bold text-gray-800 mb-3">Mystical Arts</h3>
            <p className="text-gray-600 font-cinzel">Master ancient spells and unlock the secrets of arcane knowledge through intuitive gameplay.</p>
          </div>
          
          <div className="text-center p-6">
            <div className="text-3xl mb-4 text-teal-600">🏰</div>
            <h3 className="text-xl font-cinzel font-bold text-gray-800 mb-3">Epic Dungeons</h3>
            <p className="text-gray-600 font-cinzel">Explore mysterious realms filled with challenging puzzles and hidden treasures.</p>
          </div>
          
          <div className="text-center p-6">
            <div className="text-3xl mb-4 text-teal-600">⚔️</div>
            <h3 className="text-xl font-cinzel font-bold text-gray-800 mb-3">Forge Your Destiny</h3>
            <p className="text-gray-600 font-cinzel">Every choice influences your path in this dynamic world of consequence and adventure.</p>
          </div>
        </div>
      </div>
    </section>
  );
}