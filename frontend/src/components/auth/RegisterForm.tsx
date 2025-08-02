import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { ApiError } from '../../lib/api';
import { Button } from '../ui/Button';

interface RegisterFormProps {
  onSuccess?: () => void;
  onSwitchToLogin?: () => void;
}

export const RegisterForm: React.FC<RegisterFormProps> = ({
  onSuccess,
  onSwitchToLogin,
}) => {
  const { register } = useAuth();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    first_name: '',
    last_name: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);

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

    // Email validation
    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    // First name validation
    if (!formData.first_name.trim()) {
      newErrors.first_name = 'First name is required';
    }

    // Last name validation
    if (!formData.last_name.trim()) {
      newErrors.last_name = 'Last name is required';
    }

    // Password validation
    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters long';
    }

    // Confirm password validation
    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password';
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
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
      const { confirmPassword, ...registerData } = formData;
      await register(registerData);
      onSuccess?.();
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
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white shadow-md rounded-lg px-8 pt-6 pb-8 mb-4">
        <div className="mb-6 text-center">
          <h2 className="text-2xl font-bold text-gray-900">Create Account</h2>
          <p className="text-gray-600 mt-2">Sign up to get started with your account.</p>
        </div>

        {errors.general && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {errors.general}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="flex gap-4 mb-4">
            <div className="flex-1">
              <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="first_name">
                First Name
              </label>
              <input
                className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                  errors.first_name ? 'border-red-500' : 'border-gray-300'
                }`}
                id="first_name"
                name="first_name"
                type="text"
                placeholder="First name"
                value={formData.first_name}
                onChange={handleChange}
                required
              />
              {errors.first_name && <p className="text-red-500 text-xs italic mt-1">{errors.first_name}</p>}
            </div>

            <div className="flex-1">
              <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="last_name">
                Last Name
              </label>
              <input
                className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                  errors.last_name ? 'border-red-500' : 'border-gray-300'
                }`}
                id="last_name"
                name="last_name"
                type="text"
                placeholder="Last name"
                value={formData.last_name}
                onChange={handleChange}
                required
              />
              {errors.last_name && <p className="text-red-500 text-xs italic mt-1">{errors.last_name}</p>}
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="email">
              Email Address
            </label>
            <input
              className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                errors.email ? 'border-red-500' : 'border-gray-300'
              }`}
              id="email"
              name="email"
              type="email"
              placeholder="Enter your email"
              value={formData.email}
              onChange={handleChange}
              required
            />
            {errors.email && <p className="text-red-500 text-xs italic mt-1">{errors.email}</p>}
          </div>

          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="password">
              Password
            </label>
            <input
              className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                errors.password ? 'border-red-500' : 'border-gray-300'
              }`}
              id="password"
              name="password"
              type="password"
              placeholder="Enter your password"
              value={formData.password}
              onChange={handleChange}
              required
            />
            {errors.password && <p className="text-red-500 text-xs italic mt-1">{errors.password}</p>}
          </div>

          <div className="mb-6">
            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="confirmPassword">
              Confirm Password
            </label>
            <input
              className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                errors.confirmPassword ? 'border-red-500' : 'border-gray-300'
              }`}
              id="confirmPassword"
              name="confirmPassword"
              type="password"
              placeholder="Confirm your password"
              value={formData.confirmPassword}
              onChange={handleChange}
              required
            />
            {errors.confirmPassword && <p className="text-red-500 text-xs italic mt-1">{errors.confirmPassword}</p>}
          </div>

          <div className="flex items-center justify-between mb-6">
            <Button
              type="submit"
              className="w-full"
              disabled={isLoading}
            >
              {isLoading ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Creating Account...
                </div>
              ) : (
                'Create Account'
              )}
            </Button>
          </div>
        </form>

        {onSwitchToLogin && (
          <div className="mt-6 text-center">
            <p className="text-gray-600">
              Already have an account?{' '}
              <button
                type="button"
                className="text-blue-600 hover:text-blue-800 font-medium"
                onClick={onSwitchToLogin}
              >
                Sign in
              </button>
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default RegisterForm;