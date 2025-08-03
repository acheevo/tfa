import React from 'react';
import { Header, Hero, Features } from '../components';

/**
 * Landing page component that shows the marketing/landing content
 * This will be displayed when users visit the root "/" path
 */
const Landing: React.FC = () => {
  return (
    <>
      <Header />
      <Hero />
      <Features />
    </>
  );
};

export default Landing;