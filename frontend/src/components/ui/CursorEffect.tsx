import { useEffect, useRef } from 'react';

interface Spark {
  x: number;
  y: number;
  vx: number;
  vy: number;
  life: number;
  maxLife: number;
  size: number;
  type: 'spark' | 'star';
  rotation: number;
  rotationSpeed: number;
}



export const CursorEffect: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const sparksRef = useRef<Spark[]>([]);
  const mouseRef = useRef({ x: 0, y: 0 });
  const animationRef = useRef<number>();

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

         const ctx = canvas.getContext('2d');
     if (!ctx) {
       console.error('Failed to get 2D context');
       return;
     }
     console.log('Canvas initialized, size:', canvas.width, 'x', canvas.height); // Debug

    // Set canvas size
    const resizeCanvas = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };

    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);

               // Mouse move handler
      const handleMouseMove = (e: MouseEvent) => {
        mouseRef.current.x = e.clientX;
        mouseRef.current.y = e.clientY;
      };

      // Create new spark
      const createSpark = (x: number, y: number): Spark => {
        const angle = Math.random() * Math.PI * 2;
        const speed = Math.random() * 1.5 + 0.8;
        const type = Math.random() > 0.9 ? 'star' : 'spark';
        
        return {
          x,
          y,
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          life: 1,
          maxLife: Math.random() * 0.4 + 0.6,
          size: type === 'star' ? Math.random() * 2 + 1.5 : Math.random() * 1.5 + 0.8,
          type,
          rotation: Math.random() * Math.PI * 2,
          rotationSpeed: (Math.random() - 0.5) * 0.15
        };
      };

               // Animation loop
      const animate = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Create new sparks occasionally
        if (Math.random() > 0.92) {
          const spark = createSpark(mouseRef.current.x, mouseRef.current.y);
          sparksRef.current.push(spark);
        }

      // Update and draw sparks
      sparksRef.current = sparksRef.current.filter(spark => {
        // Update position
        spark.x += spark.vx;
        spark.y += spark.vy;
        spark.life -= 0.02;
        spark.rotation += spark.rotationSpeed;

        // Remove dead sparks
        if (spark.life <= 0) return false;

        // Draw spark
        const alpha = spark.life / spark.maxLife;
        ctx.save();
        ctx.globalAlpha = alpha;

                 if (spark.type === 'star') {
           // Draw star
           ctx.fillStyle = `rgba(20, 184, 166, ${alpha})`;
           ctx.shadowColor = 'rgba(20, 184, 166, 0.8)';
           ctx.shadowBlur = 8;
          
          ctx.translate(spark.x, spark.y);
          ctx.rotate(spark.rotation);
          
          // Draw 5-pointed star
          ctx.beginPath();
          for (let i = 0; i < 5; i++) {
            const angle = (i * 2 * Math.PI) / 5 - Math.PI / 2;
            const x = Math.cos(angle) * spark.size;
            const y = Math.sin(angle) * spark.size;
            
            if (i === 0) {
              ctx.moveTo(x, y);
            } else {
              ctx.lineTo(x, y);
            }
            
            const innerAngle = angle + Math.PI / 5;
            const innerX = Math.cos(innerAngle) * (spark.size * 0.5);
            const innerY = Math.sin(innerAngle) * (spark.size * 0.5);
            ctx.lineTo(innerX, innerY);
          }
          ctx.closePath();
          ctx.fill();
                 } else {
           // Draw spark
           const gradient = ctx.createRadialGradient(
             spark.x, spark.y, 0,
             spark.x, spark.y, spark.size
           );
           gradient.addColorStop(0, `rgba(6, 182, 212, ${alpha})`);
           gradient.addColorStop(0.5, `rgba(20, 184, 166, ${alpha * 0.7})`);
           gradient.addColorStop(1, 'transparent');
           
           ctx.fillStyle = gradient;
           ctx.shadowColor = 'rgba(6, 182, 212, 0.8)';
           ctx.shadowBlur = 6;
          
          ctx.beginPath();
          ctx.arc(spark.x, spark.y, spark.size, 0, Math.PI * 2);
          ctx.fill();
        }

                         ctx.restore();
        return true;
      });

      animationRef.current = requestAnimationFrame(animate);
    };

    // Start animation
    animate();

    // Add event listeners
    window.addEventListener('mousemove', handleMouseMove);

    // Cleanup
    return () => {
      window.removeEventListener('resize', resizeCanvas);
      window.removeEventListener('mousemove', handleMouseMove);
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 pointer-events-none z-[9999]"
      style={{ background: 'transparent' }}
    />
  );
}; 