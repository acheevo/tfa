import { Container } from '@/components/ui';
import { Code2, Server, Rocket, Database, Shield, Zap } from 'lucide-react';

const features = [
  {
    title: 'React Frontend',
    description: 'Modern React 18 with TypeScript, Tailwind CSS, and Vite for lightning-fast development and builds.',
    icon: Code2,
    color: 'from-blue-500 to-cyan-500',
  },
  {
    title: 'Go API Backend',
    description: 'Robust Go API with Gin framework, clean architecture, and production-grade patterns.',
    icon: Server,
    color: 'from-emerald-500 to-teal-500',
  },
  {
    title: 'Production Ready',
    description: 'Docker support, PostgreSQL integration, health checks, and structured logging out of the box.',
    icon: Rocket,
    color: 'from-purple-500 to-pink-500',
  },
  {
    title: 'Database Integration',
    description: 'PostgreSQL with GORM, connection pooling, and migration support for scalable data management.',
    icon: Database,
    color: 'from-orange-500 to-red-500',
  },
  {
    title: 'Security First',
    description: 'JWT authentication, CORS middleware, and security best practices built into the foundation.',
    icon: Shield,
    color: 'from-indigo-500 to-purple-500',
  },
  {
    title: 'Developer Experience',
    description: 'Hot reload, TypeScript support, linting, and comprehensive tooling for productive development.',
    icon: Zap,
    color: 'from-yellow-500 to-orange-500',
  },
];

export default function Features() {
  return (
    <section id="features" className="py-24 bg-gradient-to-b from-secondary-50 to-white">
      <Container>
        <div className="text-center mb-20">
          <div className="inline-flex items-center gap-2 bg-primary-100 text-primary-700 rounded-full px-4 py-2 mb-6">
            <Zap className="h-4 w-4" />
            <span className="text-sm font-medium">Powerful Features</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-secondary-900 mb-6">
            Everything you need to build
            <br />
            <span className="bg-gradient-to-r from-primary-600 to-accent-600 bg-clip-text text-transparent">
              amazing applications
            </span>
          </h2>
          <p className="text-xl text-secondary-600 max-w-3xl mx-auto leading-relaxed">
            A comprehensive fullstack template with modern tools, best practices, 
            and production-ready features to accelerate your development.
          </p>
        </div>
        
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div key={index} className="group relative">
              <div className="bg-white rounded-2xl p-8 shadow-soft hover:shadow-medium transition-all duration-300 border border-secondary-100 hover:border-secondary-200 h-full">
                {/* Icon with gradient background */}
                <div className={`inline-flex items-center justify-center w-12 h-12 rounded-xl bg-gradient-to-r ${feature.color} mb-6 group-hover:scale-110 transition-transform duration-300`}>
                  <feature.icon className="h-6 w-6 text-white" />
                </div>
                
                {/* Content */}
                <h3 className="text-xl font-bold text-secondary-900 mb-4 group-hover:text-primary-600 transition-colors">
                  {feature.title}
                </h3>
                <p className="text-secondary-600 leading-relaxed">
                  {feature.description}
                </p>
                
                {/* Hover effect border */}
                <div className="absolute inset-0 rounded-2xl border-2 border-transparent group-hover:border-primary-200 transition-colors duration-300 pointer-events-none" />
              </div>
            </div>
          ))}
        </div>
        
        {/* Bottom section */}
        <div className="text-center mt-20">
          <div className="bg-gradient-to-r from-primary-600 to-accent-600 rounded-2xl p-8 md:p-12 text-white">
            <h3 className="text-2xl md:text-3xl font-bold mb-4">
              Ready to start building?
            </h3>
            <p className="text-xl mb-8 text-white/90 max-w-2xl mx-auto">
              Get up and running in minutes with our comprehensive template
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <button className="bg-white text-primary-600 hover:bg-secondary-50 px-8 py-3 rounded-lg font-semibold transition-colors shadow-lg">
                Quick Start Guide
              </button>
              <button className="border-2 border-white/30 hover:bg-white/10 px-8 py-3 rounded-lg font-semibold transition-colors backdrop-blur-sm">
                Documentation
              </button>
            </div>
          </div>
        </div>
      </Container>
    </section>
  );
}