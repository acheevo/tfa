import { Container, Button } from '@/components/ui';
import { ArrowRight, Github, Zap } from 'lucide-react';

export default function Hero() {
  return (
    <section className="relative bg-gradient-to-br from-primary-600 via-primary-700 to-accent-600 text-white overflow-hidden">
      {/* Background decorations */}
      <div className="absolute inset-0 opacity-40" style={{
        backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.05'%3E%3Ccircle cx='30' cy='30' r='1'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`
      }}></div>
      <div className="absolute top-0 right-0 w-96 h-96 bg-accent-500 rounded-full blur-3xl opacity-20 -translate-y-32 translate-x-32"></div>
      <div className="absolute bottom-0 left-0 w-96 h-96 bg-primary-400 rounded-full blur-3xl opacity-20 translate-y-32 -translate-x-32"></div>
      
      <Container>
        <div className="relative py-24 lg:py-32 text-center animate-fade-in-up">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm border border-white/20 rounded-full px-4 py-2.5 mb-8 hover:bg-white/15 transition-all duration-300 animate-bounce-gentle">
            <Zap className="h-4 w-4 text-accent-300 animate-pulse" />
            <span className="text-sm font-medium text-white/95">Modern Fullstack Template</span>
          </div>
          
          {/* Main headline */}
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold mb-6 bg-gradient-to-r from-white to-white/80 bg-clip-text text-transparent leading-tight">
            Build Amazing Apps
            <br />
            <span className="bg-gradient-to-r from-accent-300 to-accent-100 bg-clip-text text-transparent">
              Lightning Fast
            </span>
          </h1>
          
          {/* Subtitle */}
          <p className="text-xl md:text-2xl mb-12 max-w-4xl mx-auto text-white/90 leading-relaxed">
            A production-ready fullstack template with React frontend and Go backend.
            <br className="hidden md:block" />
            Get started in minutes with modern best practices built-in.
          </p>
          
          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-16">
            <Button 
              size="xl" 
              variant="secondary"
              icon={ArrowRight}
              iconPosition="right"
              className="shadow-xl hover:shadow-2xl transform hover:-translate-y-1 hover:scale-105 transition-all duration-300 animate-bounce-gentle"
            >
              Get Started
            </Button>
            <Button 
              size="xl" 
              variant="ghost"
              icon={Github}
              className="backdrop-blur-sm bg-white/10 hover:bg-white/20 border border-white/30 text-white shadow-lg hover:shadow-xl transform hover:-translate-y-1 hover:scale-105 transition-all duration-300"
            >
              View on GitHub
            </Button>
          </div>
          
          {/* Feature highlights */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
            {[
              { label: 'React 18 + TypeScript', value: 'Modern Frontend' },
              { label: 'Go + Gin Framework', value: 'Robust Backend' },
              { label: 'Docker + PostgreSQL', value: 'Production Ready' },
            ].map((item, index) => (
              <div 
                key={index} 
                className="bg-white/10 backdrop-blur-sm rounded-xl p-6 border border-white/20 hover:bg-white/15 hover:border-white/30 transform hover:-translate-y-1 hover:scale-105 transition-all duration-300 shadow-soft hover:shadow-medium"
                style={{ animationDelay: `${index * 0.1}s` }}
              >
                <div className="text-sm text-white/80 mb-2 font-medium">{item.label}</div>
                <div className="font-bold text-white text-lg">{item.value}</div>
              </div>
            ))}
          </div>
        </div>
      </Container>
    </section>
  );
}