import React, { useState } from 'react';
import { Navigate, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { ApiError } from '../lib/api';
import { Button } from '../components/ui/Button';
import { Container } from '../components/ui/Container';
import { Card } from '../components/ui/Card';
import { Eye, EyeOff, ArrowLeft } from 'lucide-react';

const Login: React.FC = () => {
  const { login, isAuthenticated, isLoading: authLoading } = useAuth();
  const location = useLocation();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Get the return URL from location state (passed by ProtectedRoute)
  const returnTo = location.state?.returnTo || '/dashboard';

  // Redirect if already authenticated
  if (isAuthenticated && !authLoading) {
    return <Navigate to={returnTo} replace />;
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    
    // Clear errors when user types
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setErrors({});

    try {
      await login(formData);
      // Navigation will be handled by the Navigate component above
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.details) {
          setErrors(error.details);
        } else {
          setErrors({ general: error.message });
        }
      } else {
        setErrors({ general: 'An unexpected error occurred' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleDemoLogin = async (role: 'admin' | 'user') => {
    const demoCredentials = {
      admin: { email: 'admin@example.com', password: 'admin123' },
      user: { email: 'user@example.com', password: 'user1234' }
    };

    const credentials = demoCredentials[role];
    setFormData(credentials);
    
    setIsLoading(true);
    setErrors({});

    try {
      await login(credentials);
      // Navigation will be handled by the Navigate component above
    } catch (error) {
      if (error instanceof ApiError) {
        setErrors({ general: error.message });
      } else {
        setErrors({ general: 'Demo login failed. You may need to register first.' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-accent-50">
      <Container>
        <div className="flex min-h-screen">
          {/* Left Panel - Branding */}
          <div className="hidden lg:flex lg:w-1/2 flex-col justify-center px-12">
            <div className="max-w-md">
              <Link 
                to="/" 
                className="inline-flex items-center text-primary-600 hover:text-primary-700 font-medium mb-8 transition-colors"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to home
              </Link>
              
              <div className="flex items-center gap-3 mb-8">
                <div className="w-12 h-12 bg-gradient-to-r from-primary-600 to-accent-600 rounded-xl flex items-center justify-center">
                  <span className="text-white font-bold text-lg">FT</span>
                </div>
                <h1 className="text-2xl font-bold text-secondary-900">
                  Fullstack Template
                </h1>
              </div>
              
              <h2 className="text-4xl font-bold text-secondary-900 mb-6">
                Welcome back
              </h2>
              
              <p className="text-lg text-secondary-600 mb-8">
                Sign in to access your dashboard and manage your account. 
                Experience the power of our full-stack authentication system.
              </p>

              <div className="space-y-4">
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Secure JWT-based authentication</span>
                </div>
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Role-based access control (RBAC)</span>
                </div>
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Modern React + TypeScript frontend</span>
                </div>
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Go backend with clean architecture</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Panel - Login Form */}
          <div className="w-full lg:w-1/2 flex items-center justify-center px-6 py-12">
            <div className="w-full max-w-md">
              {/* Mobile back button */}
              <div className="lg:hidden mb-8">
                <Link 
                  to="/" 
                  className="inline-flex items-center text-primary-600 hover:text-primary-700 font-medium transition-colors"
                >
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  Back to home
                </Link>
              </div>

              <Card variant="elevated" className="rounded-2xl shadow-xl animate-fade-in-up">
                <div className="text-center mb-8">
                  <h3 className="text-2xl font-bold text-secondary-900 mb-2">
                    Sign in to your account
                  </h3>
                  <p className="text-secondary-600">
                    Enter your credentials to access the dashboard
                  </p>
                </div>

                {/* Demo Login Buttons */}
                <div className="mb-6">
                  <p className="text-sm text-secondary-500 text-center mb-3">Quick demo access:</p>
                  <div className="grid grid-cols-2 gap-3">
                    <Button
                      variant="secondary"
                      onClick={() => handleDemoLogin('admin')}
                      disabled={isLoading}
                      className="text-sm"
                    >
                      Demo Admin
                    </Button>
                    <Button
                      variant="secondary"
                      onClick={() => handleDemoLogin('user')}
                      disabled={isLoading}
                      className="text-sm"
                    >
                      Demo User
                    </Button>
                  </div>
                  
                  <div className="relative my-6">
                    <div className="absolute inset-0 flex items-center">
                      <div className="w-full border-t border-secondary-200" />
                    </div>
                    <div className="relative flex justify-center text-sm">
                      <span className="px-2 bg-white text-secondary-500">Or continue with email</span>
                    </div>
                  </div>
                </div>

                {errors.general && (
                  <div className="mb-6 p-4 bg-error-50 border border-error-200 text-error-700 rounded-lg text-sm animate-slide-down">
                    {errors.general}
                  </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                  <div>
                    <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="email">
                      Email address
                    </label>
                    <input
                      className={`w-full px-4 py-3 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                        errors.email ? 'border-error-300 bg-error-50 animate-bounce-gentle' : 'border-secondary-200 bg-white hover:border-secondary-300 focus:shadow-glow'
                      }`}
                      id="email"
                      name="email"
                      type="email"
                      placeholder="Enter your email address"
                      value={formData.email}
                      onChange={handleChange}
                      required
                    />
                    {errors.email && <p className="text-error-600 text-sm mt-1 animate-slide-down">{errors.email}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="password">
                      Password
                    </label>
                    <div className="relative">
                      <input
                        className={`w-full px-4 py-3 pr-12 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                          errors.password ? 'border-error-300 bg-error-50 animate-bounce-gentle' : 'border-secondary-200 bg-white hover:border-secondary-300 focus:shadow-glow'
                        }`}
                        id="password"
                        name="password"
                        type={showPassword ? 'text' : 'password'}
                        placeholder="Enter your password"
                        value={formData.password}
                        onChange={handleChange}
                        required
                      />
                      <button
                        type="button"
                        className="absolute inset-y-0 right-0 pr-3 flex items-center text-secondary-400 hover:text-secondary-600 transition-colors duration-200 hover:scale-110"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                      </button>
                    </div>
                    {errors.password && <p className="text-error-600 text-sm mt-1 animate-slide-down">{errors.password}</p>}
                  </div>

                  <div className="flex items-center justify-between">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        className="rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
                      />
                      <span className="ml-2 text-sm text-secondary-600">Remember me</span>
                    </label>
                    <Link
                      to="/forgot-password"
                      className="text-sm text-primary-600 hover:text-primary-700 font-medium"
                    >
                      Forgot password?
                    </Link>
                  </div>

                  <Button
                    type="submit"
                    size="lg"
                    fullWidth
                    loading={isLoading}
                    loadingText="Signing in..."
                    disabled={isLoading}
                  >
                    Sign in
                  </Button>
                </form>

                <div className="mt-8 text-center">
                  <p className="text-secondary-600">
                    Don't have an account?{' '}
                    <Link
                      to="/register"
                      className="text-primary-600 hover:text-primary-700 font-semibold"
                    >
                      Sign up for free
                    </Link>
                  </p>
                </div>

                <div className="mt-6 text-center">
                  <p className="text-xs text-secondary-500">
                    By signing in, you agree to our{' '}
                    <Link to="/terms" className="text-primary-600 hover:text-primary-700">
                      Terms of Service
                    </Link>{' '}
                    and{' '}
                    <Link to="/privacy" className="text-primary-600 hover:text-primary-700">
                      Privacy Policy
                    </Link>
                  </p>
                </div>
              </Card>
            </div>
          </div>
        </div>
      </Container>
    </div>
  );
};

export default Login;