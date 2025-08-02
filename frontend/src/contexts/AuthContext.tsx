import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { api, User, LoginRequest, RegisterRequest, ApiError } from '../lib/api';

// Types
interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isEmailVerified: boolean;
}

interface AuthContextType extends AuthState {
  login: (data: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  checkAuth: () => Promise<void>;
  refreshAuth: () => Promise<void>;
}

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Provider props
interface AuthProviderProps {
  children: ReactNode;
}

// Initial state
const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
  isEmailVerified: false,
};

// AuthProvider component
export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [state, setState] = useState<AuthState>(initialState);

  // Login function
  const login = async (data: LoginRequest): Promise<void> => {
    try {
      setState(prev => ({ ...prev, isLoading: true }));
      
      const response = await api.login(data);
      
      setState({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
        isEmailVerified: response.user.email_verified,
      });
    } catch (error) {
      setState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  // Register function
  const register = async (data: RegisterRequest): Promise<void> => {
    try {
      setState(prev => ({ ...prev, isLoading: true }));
      
      const response = await api.register(data);
      
      setState({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
        isEmailVerified: response.user.email_verified,
      });
    } catch (error) {
      setState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  // Logout function
  const logout = async (): Promise<void> => {
    try {
      await api.logout();
    } catch (error) {
      console.error('Logout error:', error);
      // Continue with local logout even if API call fails
    } finally {
      setState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        isEmailVerified: false,
      });
      
      // Clear any stored tokens or user data
      localStorage.removeItem('user');
    }
  };

  // Logout from all devices
  const logoutAll = async (): Promise<void> => {
    try {
      await api.logoutAll();
    } catch (error) {
      console.error('Logout all error:', error);
      // Continue with local logout even if API call fails
    } finally {
      setState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        isEmailVerified: false,
      });
      
      // Clear any stored tokens or user data
      localStorage.removeItem('user');
    }
  };

  // Check authentication status
  const checkAuth = async (): Promise<void> => {
    try {
      setState(prev => ({ ...prev, isLoading: true }));
      
      const response = await api.checkAuth();
      
      if (response.authenticated && response.user) {
        setState({
          user: response.user,
          isAuthenticated: true,
          isLoading: false,
          isEmailVerified: response.user.email_verified,
        });
      } else {
        setState({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          isEmailVerified: false,
        });
      }
    } catch (error) {
      console.error('Auth check error:', error);
      setState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        isEmailVerified: false,
      });
    }
  };

  // Refresh user data
  const refreshAuth = async (): Promise<void> => {
    if (!state.isAuthenticated) {
      return;
    }

    try {
      const user = await api.getProfile();
      setState(prev => ({
        ...prev,
        user,
        isEmailVerified: user.email_verified,
      }));
    } catch (error) {
      console.error('Refresh auth error:', error);
      // If we can't get the profile, the user might be logged out
      if (error instanceof ApiError && error.status === 401) {
        setState({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          isEmailVerified: false,
        });
      }
    }
  };

  // Check authentication on mount
  useEffect(() => {
    checkAuth();
  }, []);

  // Provide context value
  const contextValue: AuthContextType = {
    ...state,
    login,
    register,
    logout,
    logoutAll,
    checkAuth,
    refreshAuth,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook to use auth context
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  
  return context;
};

// Hook to check if user is authenticated
// Note: This hook is kept for backward compatibility, but ProtectedRoute should be used instead
export const useRequireAuth = (): AuthContextType => {
  const auth = useAuth();
  
  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated) {
      // This is now handled by ProtectedRoute component
      console.warn('useRequireAuth: User is not authenticated. Consider using ProtectedRoute component instead.');
    }
  }, [auth.isLoading, auth.isAuthenticated]);
  
  return auth;
};

// Hook to check if email is verified
// Note: This hook is kept for backward compatibility, but route-level verification should be used instead
export const useRequireEmailVerification = (): AuthContextType => {
  const auth = useAuth();
  
  useEffect(() => {
    if (!auth.isLoading && auth.isAuthenticated && !auth.isEmailVerified) {
      // This is now handled by ProtectedRoute component with requireEmailVerification prop
      console.warn('useRequireEmailVerification: Email not verified. Consider using ProtectedRoute with requireEmailVerification prop.');
    }
  }, [auth.isLoading, auth.isAuthenticated, auth.isEmailVerified]);
  
  return auth;
};

export default AuthContext;