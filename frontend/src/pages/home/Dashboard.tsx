import React, { useEffect, useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { api, DashboardResponse, ApiError } from '../../lib/api';
import { Button } from '../../components/ui/Button';
import { Card, CardHeader, CardContent } from '../../components/ui/Card';
import { SkeletonStats, SkeletonCard } from '../../components/ui/Skeleton';

export const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [dashboardData, setDashboardData] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDashboard = async () => {
      try {
        setLoading(true);
        const data = await api.getDashboard();
        setDashboardData(data);
      } catch (err) {
        if (err instanceof ApiError) {
          setError(err.message);
        } else {
          setError('Failed to load dashboard');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchDashboard();
  }, []);

  if (loading) {
    return (
      <div className="p-8">
        {/* Welcome Header Skeleton */}
        <div className="mb-8 space-y-3">
          <div className="h-8 bg-secondary-200 rounded-lg w-1/3 animate-pulse"></div>
          <div className="h-4 bg-secondary-200 rounded-lg w-2/3 animate-pulse"></div>
        </div>

        {/* Stats Cards Skeleton */}
        <SkeletonStats className="mb-8" />

        {/* Grid sections skeleton */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          <SkeletonCard />
          <SkeletonCard />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="text-center py-12">
          <div className="text-error-600 mb-4">{error}</div>
          <Button variant="destructive" onClick={() => window.location.reload()}>Try Again</Button>
        </div>
      </div>
    );
  }

  if (!dashboardData) {
    return (
      <div className="p-8">
        <div className="text-center py-12">
          <div className="text-gray-600">No dashboard data available</div>
        </div>
      </div>
    );
  }

  const { stats, recent_logins, notifications } = dashboardData;

  return (
    <div className="p-8">
        {/* Welcome Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-secondary-900">
            Welcome back, {user?.first_name}!
          </h1>
          <p className="text-secondary-600 mt-2">
            Here's an overview of your account activity.
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <Card>
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-10 h-10 bg-primary-100 rounded-xl flex items-center justify-center">
                  <svg className="w-5 h-5 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <div className="text-sm font-medium text-secondary-600">Total Logins</div>
                <div className="text-2xl font-bold text-secondary-900">{stats.total_logins}</div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-10 h-10 bg-success-100 rounded-xl flex items-center justify-center">
                  <svg className="w-5 h-5 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <div className="text-sm font-medium text-secondary-600">Account Age</div>
                <div className="text-2xl font-bold text-secondary-900">{stats.account_age_days} days</div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                  stats.profile_complete ? 'bg-success-100' : 'bg-warning-100'
                }`}>
                  <svg className={`w-5 h-5 ${stats.profile_complete ? 'text-success-600' : 'text-warning-600'}`} 
                       fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <div className="text-sm font-medium text-secondary-600">Profile Status</div>
                <div className="text-2xl font-bold text-secondary-900">
                  {stats.profile_complete ? 'Complete' : 'Incomplete'}
                </div>
              </div>
            </div>
          </Card>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Recent Activity */}
          <Card variant="elevated" padding="none">
            <CardHeader className="px-6 py-4 border-b border-secondary-200">
              <h3 className="text-lg font-semibold text-secondary-900">Recent Login Activity</h3>
            </CardHeader>
            <CardContent className="p-6">
              {recent_logins.length > 0 ? (
                <div className="space-y-4">
                  {recent_logins.slice(0, 5).map((login) => (
                    <div key={login.id} className="flex items-center justify-between py-2">
                      <div className="flex items-center">
                        <div className={`w-2.5 h-2.5 rounded-full mr-3 ${
                          login.success ? 'bg-success-500' : 'bg-error-500'
                        }`} />
                        <div>
                          <div className="text-sm font-medium text-secondary-900">
                            {login.success ? 'Successful Login' : 'Failed Login'}
                          </div>
                          <div className="text-xs text-secondary-500">
                            IP: {login.ip_address}
                          </div>
                        </div>
                      </div>
                      <div className="text-xs text-secondary-500">
                        {new Date(login.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center text-secondary-500 py-4">
                  No recent login activity
                </div>
              )}
            </CardContent>
          </Card>

          {/* Notifications */}
          <Card variant="elevated" padding="none">
            <CardHeader className="px-6 py-4 border-b border-secondary-200">
              <h3 className="text-lg font-semibold text-secondary-900">Notifications</h3>
            </CardHeader>
            <CardContent className="p-6">
              {notifications.length > 0 ? (
                <div className="space-y-4">
                  {notifications.slice(0, 5).map((notification) => (
                    <div key={notification.id} className="flex items-start">
                      <div className={`flex-shrink-0 w-2.5 h-2.5 rounded-full mt-2 mr-3 ${
                        notification.type === 'error' ? 'bg-error-500' :
                        notification.type === 'warning' ? 'bg-warning-500' :
                        notification.type === 'success' ? 'bg-success-500' :
                        'bg-primary-500'
                      }`} />
                      <div className="flex-1">
                        <div className="text-sm font-medium text-secondary-900">
                          {notification.title}
                        </div>
                        <div className="text-sm text-secondary-600 mt-1">
                          {notification.message}
                        </div>
                        <div className="text-xs text-secondary-500 mt-1">
                          {new Date(notification.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center text-secondary-500 py-4">
                  No notifications
                </div>
              )}
            </CardContent>
          </Card>
        </div>
    </div>
  );
};

export default Dashboard;