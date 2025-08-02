import React, { useState } from 'react';
import { api, ApiError } from '../../lib/api';
import { Button } from '../ui/Button';

interface ForgotPasswordFormProps {
  onBack?: () => void;
  onSuccess?: () => void;
}

export const ForgotPasswordForm: React.FC<ForgotPasswordFormProps> = ({
  onBack,
  onSuccess,
}) => {
  const [email, setEmail] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    
    // Clear errors when user types
    if (errors.email) {
      setErrors(prev => ({ ...prev, email: '' }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setErrors({});

    try {
      await api.forgotPassword({ email });
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
            
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Check Your Email</h2>
            <p className="text-gray-600 mb-6">
              If an account with that email exists, we've sent you a password reset link.
            </p>
            
            <div className="space-y-3">
              <Button
                type="button"
                className="w-full"
                onClick={onBack}
              >
                Back to Sign In
              </Button>
              
              <button
                type="button"
                className="w-full text-blue-600 hover:text-blue-800 text-sm font-medium"
                onClick={() => {
                  setIsSuccess(false);
                  setEmail('');
                }}
              >
                Send another email
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white shadow-md rounded-lg px-8 pt-6 pb-8 mb-4">
        <div className="mb-6 text-center">
          <h2 className="text-2xl font-bold text-gray-900">Reset Password</h2>
          <p className="text-gray-600 mt-2">
            Enter your email address and we'll send you a link to reset your password.
          </p>
        </div>

        {errors.general && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {errors.general}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-6">
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
              value={email}
              onChange={handleChange}
              required
            />
            {errors.email && <p className="text-red-500 text-xs italic mt-1">{errors.email}</p>}
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
                  Sending Reset Link...
                </div>
              ) : (
                'Send Reset Link'
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

export default ForgotPasswordForm;