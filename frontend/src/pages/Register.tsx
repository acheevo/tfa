import React, { useState } from 'react';
import { Navigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { ApiError } from '../lib/api';
import { Button } from '../components/ui/Button';
import { Container } from '../components/ui/Container';
import { Eye, EyeOff, ArrowLeft } from 'lucide-react';

const Register: React.FC = () => {
  const { register, isAuthenticated, isLoading: authLoading } = useAuth();
  const [formData, setFormData] = useState({
    first_name: '',
    last_name: '',
    email: '',
    password: '',
    confirm_password: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // Redirect if already authenticated
  if (isAuthenticated && !authLoading) {
    return <Navigate to="/dashboard" replace />;
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    
    // Clear errors when user types
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.first_name.trim()) {
      newErrors.first_name = 'First name is required';
    }

    if (!formData.last_name.trim()) {
      newErrors.last_name = 'Last name is required';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters long';
    }

    if (!formData.confirm_password) {
      newErrors.confirm_password = 'Please confirm your password';
    } else if (formData.password !== formData.confirm_password) {
      newErrors.confirm_password = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      const { confirm_password, ...registerData } = formData;
      await register(registerData);
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
                Join our platform
              </h2>
              
              <p className="text-lg text-secondary-600 mb-8">
                Create your account to access powerful features and start building
                amazing applications with our full-stack template.
              </p>

              <div className="space-y-4">
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Complete authentication system</span>
                </div>
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Role-based permissions</span>
                </div>
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Modern React + TypeScript</span>
                </div>
                <div className="flex items-center text-secondary-600">
                  <div className="w-2 h-2 bg-primary-500 rounded-full mr-3"></div>
                  <span>Production-ready architecture</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Panel - Register Form */}
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

              <div className="bg-white rounded-2xl shadow-xl p-8">
                <div className="text-center mb-8">
                  <h3 className="text-2xl font-bold text-secondary-900 mb-2">
                    Create your account
                  </h3>
                  <p className="text-secondary-600">
                    Get started with your free account today
                  </p>
                </div>

                {errors.general && (
                  <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
                    {errors.general}
                  </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="first_name">
                        First Name
                      </label>
                      <input
                        className={`w-full px-4 py-3 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                          errors.first_name ? 'border-red-300 bg-red-50' : 'border-secondary-200 bg-white hover:border-secondary-300'
                        }`}
                        id="first_name"
                        name="first_name"
                        type="text"
                        placeholder="John"
                        value={formData.first_name}
                        onChange={handleChange}
                        required
                      />
                      {errors.first_name && <p className="text-red-600 text-sm mt-1">{errors.first_name}</p>}
                    </div>

                    <div>
                      <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="last_name">
                        Last Name
                      </label>
                      <input
                        className={`w-full px-4 py-3 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                          errors.last_name ? 'border-red-300 bg-red-50' : 'border-secondary-200 bg-white hover:border-secondary-300'
                        }`}
                        id="last_name"
                        name="last_name"
                        type="text"
                        placeholder="Doe"
                        value={formData.last_name}
                        onChange={handleChange}
                        required
                      />
                      {errors.last_name && <p className="text-red-600 text-sm mt-1">{errors.last_name}</p>}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="email">
                      Email address
                    </label>
                    <input
                      className={`w-full px-4 py-3 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                        errors.email ? 'border-red-300 bg-red-50' : 'border-secondary-200 bg-white hover:border-secondary-300'
                      }`}
                      id="email"
                      name="email"
                      type="email"
                      placeholder="john@example.com"
                      value={formData.email}
                      onChange={handleChange}
                      required
                    />
                    {errors.email && <p className="text-red-600 text-sm mt-1">{errors.email}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="password">
                      Password
                    </label>
                    <div className="relative">
                      <input
                        className={`w-full px-4 py-3 pr-12 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                          errors.password ? 'border-red-300 bg-red-50' : 'border-secondary-200 bg-white hover:border-secondary-300'
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
                        className="absolute inset-y-0 right-0 pr-3 flex items-center text-secondary-400 hover:text-secondary-600"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                      </button>
                    </div>
                    {errors.password && <p className="text-red-600 text-sm mt-1">{errors.password}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-secondary-900 mb-2" htmlFor="confirm_password">
                      Confirm Password
                    </label>
                    <div className="relative">
                      <input
                        className={`w-full px-4 py-3 pr-12 rounded-lg border text-secondary-900 placeholder-secondary-400 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
                          errors.confirm_password ? 'border-red-300 bg-red-50' : 'border-secondary-200 bg-white hover:border-secondary-300'
                        }`}
                        id="confirm_password"
                        name="confirm_password"
                        type={showConfirmPassword ? 'text' : 'password'}
                        placeholder="Confirm your password"
                        value={formData.confirm_password}
                        onChange={handleChange}
                        required
                      />
                      <button
                        type="button"
                        className="absolute inset-y-0 right-0 pr-3 flex items-center text-secondary-400 hover:text-secondary-600"
                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                      >
                        {showConfirmPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                      </button>
                    </div>
                    {errors.confirm_password && <p className="text-red-600 text-sm mt-1">{errors.confirm_password}</p>}
                  </div>

                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      id="terms"
                      className="rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
                      required
                    />
                    <label htmlFor="terms" className="ml-2 text-sm text-secondary-600">
                      I agree to the{' '}
                      <Link to="/terms" className="text-primary-600 hover:text-primary-700 font-medium">
                        Terms of Service
                      </Link>{' '}
                      and{' '}
                      <Link to="/privacy" className="text-primary-600 hover:text-primary-700 font-medium">
                        Privacy Policy
                      </Link>
                    </label>
                  </div>

                  <Button
                    type="submit"
                    className="w-full py-3"
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <div className="flex items-center justify-center">
                        <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                        Creating account...
                      </div>
                    ) : (
                      'Create Account'
                    )}
                  </Button>
                </form>

                <div className="mt-8 text-center">
                  <p className="text-secondary-600">
                    Already have an account?{' '}
                    <Link
                      to="/login"
                      className="text-primary-600 hover:text-primary-700 font-semibold"
                    >
                      Sign in instead
                    </Link>
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Container>
    </div>
  );
};

export default Register;