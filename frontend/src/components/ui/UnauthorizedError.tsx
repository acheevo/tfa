import React from 'react';
import { Link } from 'react-router-dom';
import { Shield, ArrowLeft } from 'lucide-react';
import { Container } from './Container';
import { Button } from './Button';

interface UnauthorizedErrorProps {
  title?: string;
  message?: string;
  actionText?: string;
  actionPath?: string;
  showBackButton?: boolean;
}

/**
 * UnauthorizedError component for displaying access denied messages
 * Used when users don't have sufficient permissions
 */
export const UnauthorizedError: React.FC<UnauthorizedErrorProps> = ({
  title = 'Access Denied',
  message = 'You do not have permission to access this resource.',
  actionText = 'Go Back',
  actionPath = '/dashboard',
  showBackButton = true,
}) => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-red-50 via-white to-orange-50">
      <Container>
        <div className="flex items-center justify-center min-h-screen py-12">
          <div className="text-center max-w-md">
            <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-6">
              <Shield className="w-8 h-8 text-red-600" />
            </div>
            
            <h1 className="text-3xl font-bold text-secondary-900 mb-4">
              {title}
            </h1>
            
            <p className="text-lg text-secondary-600 mb-8">
              {message}
            </p>
            
            <div className="space-y-4">
              {showBackButton && (
                <Link to={actionPath}>
                  <Button className="w-full sm:w-auto">
                    <ArrowLeft className="w-4 h-4 mr-2" />
                    {actionText}
                  </Button>
                </Link>
              )}
              
              <div className="text-sm text-secondary-500">
                If you believe this is an error, please contact your administrator.
              </div>
            </div>
          </div>
        </div>
      </Container>
    </div>
  );
};

export default UnauthorizedError;