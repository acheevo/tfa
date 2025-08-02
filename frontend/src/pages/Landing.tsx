import React from 'react';
import { Hero, Features, CTA } from '../components';

/**
 * Landing page component that shows the marketing/landing content
 * This will be displayed when users visit the root "/" path
 */
const Landing: React.FC = () => {
  return (
    <main className="flex-grow">
      <Hero />
      <Features />
      <CTA />
    </main>
  );
};

export default Landing;