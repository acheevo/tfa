import React, { useState, useEffect } from 'react';
import { api, ApiError } from '../../lib/api';
import { useAuth } from '../../contexts/AuthContext';
import { Button } from '../ui/Button';

interface EmailVerificationFormProps {
  token?: string;
  onSuccess?: () => void;
  onBack?: () => void;
}

export const EmailVerificationForm: React.FC<EmailVerificationFormProps> = ({
  token,
  onSuccess,
  onBack,
}) => {
  const { refreshAuth, user } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [resendMessage, setResendMessage] = useState('');

  useEffect(() => {
    // Auto-verify if token is provided
    if (token) {
      handleVerify(token);
    } else {
      // Try to get token from URL params
      const urlParams = new URLSearchParams(window.location.search);
      const urlToken = urlParams.get('token');
      if (urlToken) {
        handleVerify(urlToken);
      }
    }
  }, [token]);

  const handleVerify = async (verificationToken: string) => {
    setIsLoading(true);
    setErrors({});

    try {
      await api.verifyEmail({ token: verificationToken });
      setIsSuccess(true);
      
      // Refresh auth state to update email verification status
      await refreshAuth();
      
      onSuccess?.();
    } catch (error) {
      if (error instanceof ApiError) {
        setErrors({ general: error.message });
      } else {
        setErrors({ general: 'An unexpected error occurred' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendVerification = async () => {
    setIsResending(true);
    setErrors({});
    setResendMessage('');

    try {
      await api.resendEmailVerification();
      setResendMessage('Verification email sent! Please check your inbox.');
    } catch (error) {
      if (error instanceof ApiError) {
        setErrors({ general: error.message });
      } else {
        setErrors({ general: 'Failed to resend verification email' });
      }
    } finally {
      setIsResending(false);
    }
  };

  if (isLoading) {
    return (
      <div className="w-full max-w-md mx-auto">
        <div className="bg-white shadow-md rounded-lg px-8 pt-6 pb-8 mb-4">
          <div className="text-center">
            <div className="mb-4">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            </div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">Verifying Email</h2>
            <p className="text-gray-600">Please wait while we verify your email address...</p>
          </div>
        </div>
      </div>
    );
  }

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
            
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Email Verified!</h2>
            <p className="text-gray-600 mb-6">
              Your email address has been successfully verified. You now have full access to your account.
            </p>
            
            <Button
              type="button"
              className="w-full"
              onClick={() => window.location.href = '/'}
            >
              Continue to Dashboard
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white shadow-md rounded-lg px-8 pt-6 pb-8 mb-4">
        <div className="text-center">
          <div className="mb-4">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-yellow-100">
              <svg
                className="h-6 w-6 text-yellow-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.314 16.5c-.77.833.192 2.5 1.732 2.5z"
                />
              </svg>
            </div>
          </div>
          
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Verify Your Email</h2>
          
          {user ? (
            <p className="text-gray-600 mb-6">
              We've sent a verification email to <strong>{user.email}</strong>. 
              Please check your inbox and click the verification link.
            </p>
          ) : (
            <p className="text-gray-600 mb-6">
              Please check your email for a verification link.
            </p>
          )}

          {errors.general && (
            <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded text-sm">
              {errors.general}
            </div>
          )}

          {resendMessage && (
            <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded text-sm">
              {resendMessage}
            </div>
          )}

          <div className="space-y-3">
            {user && (
              <Button
                type="button"
                className="w-full"
                onClick={handleResendVerification}
                disabled={isResending}
              >
                {isResending ? (
                  <div className="flex items-center justify-center">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Sending...
                  </div>
                ) : (
                  'Resend Verification Email'
                )}
              </Button>
            )}

            {onBack && (
              <button
                type="button"
                className="w-full text-blue-600 hover:text-blue-800 text-sm font-medium"
                onClick={onBack}
              >
                Back to Sign In
              </button>
            )}
          </div>

          <div className="mt-6 text-xs text-gray-500">
            <p>Didn't receive the email? Check your spam folder or try resending.</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmailVerificationForm;