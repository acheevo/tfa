import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Container, Button } from '@/components/ui';
import { Github, Settings, LogOut, Shield, Bell } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import { useRBAC } from '../../contexts/RBACContext';

export default function Header() {
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const { isAuthenticated, isLoading, user, logout } = useAuth();
  const { isAdmin } = useRBAC();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/');
      setIsUserMenuOpen(false);
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleGetStarted = () => {
    if (isAuthenticated) {
      navigate('/dashboard');
    } else {
      navigate('/login');
    }
  };

  // Show loading skeleton while authentication is being checked
  const AuthSection = () => {
    if (isLoading) {
      return (
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-secondary-200 rounded-full animate-pulse"></div>
        </div>
      );
    }

    if (isAuthenticated && user) {
      return (
        <div className="flex items-center space-x-3">
          {/* Notifications */}
          <button className="p-2 text-secondary-500 hover:text-secondary-700 hover:bg-secondary-100 rounded-lg">
            <Bell className="w-5 h-5" />
          </button>

          {/* User Menu */}
          <div className="relative">
            <button
              onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
              className="flex items-center space-x-2 p-1 rounded-lg hover:bg-secondary-100"
            >
              <div className="w-8 h-8 bg-gradient-to-r from-primary-600 to-accent-600 rounded-full flex items-center justify-center">
                <span className="text-white font-semibold text-sm">
                  {user.first_name?.charAt(0)?.toUpperCase() || 'U'}
                </span>
              </div>
            </button>

            {/* User Menu Dropdown */}
            {isUserMenuOpen && (
              <div className="absolute right-0 mt-2 w-56 bg-white rounded-lg shadow-lg border border-secondary-200 py-1 z-50">
                <div className="px-4 py-3 border-b border-secondary-200">
                  <div className="text-sm font-medium text-secondary-900">
                    {user.first_name} {user.last_name}
                  </div>
                  <div className="text-xs text-secondary-500">
                    {user.email}
                  </div>
                </div>
                
                <Link
                  to="/settings"
                  className="flex items-center px-4 py-2 text-sm text-secondary-700 hover:bg-secondary-50"
                  onClick={() => setIsUserMenuOpen(false)}
                >
                  <Settings className="w-4 h-4 mr-3" />
                  Settings
                </Link>
                
                {isAdmin() && (
                  <Link
                    to="/admin"
                    className="flex items-center px-4 py-2 text-sm text-secondary-700 hover:bg-secondary-50"
                    onClick={() => setIsUserMenuOpen(false)}
                  >
                    <Shield className="w-4 h-4 mr-3" />
                    Admin Panel
                  </Link>
                )}
                
                <div className="border-t border-secondary-200 my-1"></div>
                <button
                  onClick={handleLogout}
                  className="flex items-center w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                >
                  <LogOut className="w-4 h-4 mr-3" />
                  Sign out
                </button>
              </div>
            )}
          </div>
        </div>
      );
    }

    return (
      <div className="flex items-center space-x-4">
        <Button variant="ghost" size="sm" icon={Github}>
          <a href="https://github.com" target="_blank" rel="noopener noreferrer" className="flex items-center">
            GitHub
          </a>
        </Button>
        <Button size="sm" onClick={handleGetStarted}>
          Get Started
        </Button>
      </div>
    );
  };

  return (
    <header className="bg-white border-b border-secondary-200 sticky top-0 z-50">
      <Container>
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link 
            to={isAuthenticated ? "/dashboard" : "/"} 
            className="flex items-center gap-3 hover:opacity-80"
          >
            <div className="w-8 h-8 bg-gradient-to-r from-primary-600 to-accent-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">FT</span>
            </div>
            <h1 className="text-xl font-bold text-secondary-900">
              Fullstack Template
            </h1>
          </Link>
          
          {/* Right Side - Auth Section */}
          <AuthSection />
        </div>
      </Container>
      
      {/* Backdrop for user menu */}
      {isUserMenuOpen && (
        <div 
          className="fixed inset-0 z-40"
          onClick={() => setIsUserMenuOpen(false)}
        />
      )}
    </header>
  );
}