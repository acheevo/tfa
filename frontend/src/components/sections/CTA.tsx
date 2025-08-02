import { Container, Button } from '@/components/ui';
import { ArrowRight, Download, Github } from 'lucide-react';

export default function CTA() {
  return (
    <section className="py-24 bg-gradient-to-r from-secondary-900 via-secondary-800 to-secondary-900 text-white relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 bg-gradient-to-r from-primary-600/20 to-accent-600/20"></div>
      <div className="absolute top-0 left-0 w-72 h-72 bg-primary-500 rounded-full blur-3xl opacity-10 -translate-x-36 -translate-y-36"></div>
      <div className="absolute bottom-0 right-0 w-72 h-72 bg-accent-500 rounded-full blur-3xl opacity-10 translate-x-36 translate-y-36"></div>
      
      <Container>
        <div className="relative text-center">
          <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm border border-white/20 rounded-full px-4 py-2 mb-8">
            <span className="w-2 h-2 bg-success-400 rounded-full animate-pulse"></span>
            <span className="text-sm font-medium text-white/90">Template Ready</span>
          </div>
          
          <h2 className="text-4xl md:text-5xl font-bold mb-6 leading-tight">
            Ready to build something
            <br />
            <span className="bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">
              extraordinary?
            </span>
          </h2>
          
          <p className="text-xl mb-12 max-w-3xl mx-auto text-white/80 leading-relaxed">
            Get your fullstack application up and running in minutes. No complex setup, 
            no configuration headaches—just pure development productivity.
          </p>
          
          <div className="flex flex-col sm:flex-row gap-6 justify-center items-center mb-16">
            <Button 
              size="xl" 
              variant="secondary"
              icon={ArrowRight}
              iconPosition="right"
              className="shadow-xl hover:shadow-2xl transform hover:-translate-y-0.5 transition-all"
            >
              Get Started Now
            </Button>
            
            <Button 
              size="xl" 
              variant="ghost"
              icon={Github}
              className="backdrop-blur-sm bg-white/5 hover:bg-white/10 border border-white/20 text-white shadow-lg"
            >
              Clone Repository
            </Button>
            
            <Button 
              size="xl" 
              variant="ghost"
              icon={Download}
              className="backdrop-blur-sm bg-white/5 hover:bg-white/10 border border-white/20 text-white shadow-lg"
            >
              Download ZIP
            </Button>
          </div>
          
          {/* Quick stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-4xl mx-auto">
            <div className="text-center">
              <div className="text-3xl font-bold text-primary-400 mb-2">5 min</div>
              <div className="text-white/70">Setup Time</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-accent-400 mb-2">Production</div>
              <div className="text-white/70">Ready Code</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-success-400 mb-2">Best</div>
              <div className="text-white/70">Practices</div>
            </div>
          </div>
        </div>
      </Container>
    </section>
  );
}