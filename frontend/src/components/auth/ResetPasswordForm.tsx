import React, { useState, useEffect } from 'react';
import { api, ApiError } from '../../lib/api';
import { Button } from '../ui/Button';

interface ResetPasswordFormProps {
  token?: string;
  onSuccess?: () => void;
  onBack?: () => void;
}

export const ResetPasswordForm: React.FC<ResetPasswordFormProps> = ({
  token,
  onSuccess,
  onBack,
}) => {
  const [formData, setFormData] = useState({
    password: '',
    confirm_password: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  useEffect(() => {
    // If no token is provided, try to get it from URL params
    if (!token) {
      const urlParams = new URLSearchParams(window.location.search);
      const urlToken = urlParams.get('token');
      if (!urlToken) {
        setErrors({ general: 'Invalid or missing reset token' });
      }
    }
  }, [token]);

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

    // Password validation
    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters long';
    }

    // Confirm password validation
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

    const resetToken = token || new URLSearchParams(window.location.search).get('token');
    if (!resetToken) {
      setErrors({ general: 'Invalid or missing reset token' });
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      await api.resetPassword({
        token: resetToken,
        password: formData.password,
        confirm_password: formData.confirm_password,
      });
      setIsSuccess(true);
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

  if (isSuccess) {
    return (
      <div className="w-full max-w-md mx-auto">
        <div className="bg-white shadow-md rounded-lg px-8 pt-6 pb-8 mb-4">
          <div className="text-center">
            <div className="mb-4">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
                <svg
                  className="h-6 w-6 text-green-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
            </div>
            
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Password Reset Successful</h2>
            <p className="text-gray-600 mb-6">
              Your password has been successfully reset. You can now sign in with your new password.
            </p>
            
            <Button
              type="button"
              className="w-full"
              onClick={() => window.location.href = '/login'}
            >
              Go to Sign In
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white shadow-md rounded-lg px-8 pt-6 pb-8 mb-4">
        <div className="mb-6 text-center">
          <h2 className="text-2xl font-bold text-gray-900">Set New Password</h2>
          <p className="text-gray-600 mt-2">
            Enter your new password below.
          </p>
        </div>

        {errors.general && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {errors.general}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="password">
              New Password
            </label>
            <input
              className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                errors.password ? 'border-red-500' : 'border-gray-300'
              }`}
              id="password"
              name="password"
              type="password"
              placeholder="Enter your new password"
              value={formData.password}
              onChange={handleChange}
              required
            />
            {errors.password && <p className="text-red-500 text-xs italic mt-1">{errors.password}</p>}
          </div>

          <div className="mb-6">
            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="confirm_password">
              Confirm New Password
            </label>
            <input
              className={`shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline ${
                errors.confirm_password ? 'border-red-500' : 'border-gray-300'
              }`}
              id="confirm_password"
              name="confirm_password"
              type="password"
              placeholder="Confirm your new password"
              value={formData.confirm_password}
              onChange={handleChange}
              required
            />
            {errors.confirm_password && <p className="text-red-500 text-xs italic mt-1">{errors.confirm_password}</p>}
          </div>

          <div className="flex flex-col space-y-3">
            <Button
              type="submit"
              className="w-full"
              disabled={isLoading}
            >
              {isLoading ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Resetting Password...
                </div>
              ) : (
                'Reset Password'
              )}
            </Button>

            {onBack && (
              <button
                type="button"
                className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                onClick={onBack}
              >
                Back to Sign In
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
  );
};

export default ResetPasswordForm;