import { Container } from '@/components/ui';

export default function Footer() {
  return (
    <footer className="bg-gray-800 text-white">
      <Container>
        <div className="py-8">
          <div className="text-center">
            <p className="text-gray-400">
              © 2025 Fullstack Template. Built with React and Go.
            </p>
          </div>
        </div>
      </Container>
    </footer>
  );
}