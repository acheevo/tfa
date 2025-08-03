import React from 'react';
import { Header, Hero } from '../components';

/**
 * Landing page component that shows the marketing/landing content
 * This will be displayed when users visit the root "/" path
 */
const Landing: React.FC = () => {
  return (
    <>
      <Header />
      <Hero />
    </>
  );
};

export default Landing;