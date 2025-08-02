import React, { Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Header, Footer, Sidebar } from '@/components';
import { AuthProvider } from './contexts/AuthContext';
import { RBACProvider } from './contexts/RBACContext';
import { ProtectedRoute, PublicRoute, AdminRoute } from './components/routing';
import { LoadingSpinner } from './components/ui/LoadingSpinner';

// Lazy load components for better performance
const Landing = React.lazy(() => import('./pages/Landing'));
const Login = React.lazy(() => import('./pages/Login'));
const Register = React.lazy(() => import('./pages/Register'));
const Dashboard = React.lazy(() => import('./pages/home/Dashboard'));
const Settings = React.lazy(() => import('./pages/settings/Settings'));
const AdminPanel = React.lazy(() => import('./pages/admin/AdminPanel'));

// Loading fallback component
const PageLoader: React.FC = () => (
  <div className="min-h-screen flex items-center justify-center bg-secondary-50">
    <div className="text-center">
      <LoadingSpinner size="lg" />
      <p className="mt-4 text-secondary-600">Loading...</p>
    </div>
  </div>
);

// Layout components
const AuthenticatedLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="min-h-screen flex flex-col">
    <Header />
    <div className="flex flex-1">
      <Sidebar className="w-64 hidden md:block" />
      <main className="flex-1">
        {children}
      </main>
    </div>
  </div>
);

const PublicLayout: React.FC<{ children: React.ReactNode; showFooter?: boolean }> = ({ 
  children, 
  showFooter = true 
}) => (
  <div className="min-h-screen flex flex-col">
    <Header />
    <div className="flex-grow">
      {children}
    </div>
    {showFooter && <Footer />}
  </div>
);

function App() {
  return (
    <AuthProvider>
      <RBACProvider>
        <Suspense fallback={<PageLoader />}>
          <Routes>
            {/* Public Routes - Only accessible when not authenticated */}
            <Route 
              path="/login" 
              element={
                <PublicRoute>
                  <PublicLayout showFooter={false}>
                    <Login />
                  </PublicLayout>
                </PublicRoute>
              } 
            />
            <Route 
              path="/register" 
              element={
                <PublicRoute>
                  <PublicLayout showFooter={false}>
                    <Register />
                  </PublicLayout>
                </PublicRoute>
              } 
            />

            {/* Protected Routes - Require authentication */}
            <Route 
              path="/dashboard" 
              element={
                <ProtectedRoute>
                  <AuthenticatedLayout>
                    <Dashboard />
                  </AuthenticatedLayout>
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/settings" 
              element={
                <ProtectedRoute>
                  <AuthenticatedLayout>
                    <Settings />
                  </AuthenticatedLayout>
                </ProtectedRoute>
              } 
            />

            {/* Admin Routes - Require admin role */}
            <Route 
              path="/admin" 
              element={
                <AdminRoute>
                  <AuthenticatedLayout>
                    <AdminPanel />
                  </AuthenticatedLayout>
                </AdminRoute>
              } 
            />

            {/* Landing/Home Route - Accessible to everyone */}
            <Route 
              path="/" 
              element={
                <PublicLayout>
                  <Landing />
                </PublicLayout>
              } 
            />

            {/* Catch-all route - Redirect to home */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Suspense>
      </RBACProvider>
    </AuthProvider>
  );
}

export default App;