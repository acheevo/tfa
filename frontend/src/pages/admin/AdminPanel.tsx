import React, { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { api, AdminStatsResponse, UserSummary, ApiError, AdminAuditLogRequest, EnhancedAuditLogEntry } from '../../lib/api';
import { Container } from '../../components/ui/Container';
import { Button } from '../../components/ui/Button';

interface AdminPanelProps {}

export const AdminPanel: React.FC<AdminPanelProps> = () => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'dashboard' | 'users' | 'audit'>('dashboard');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Dashboard data
  const [stats, setStats] = useState<AdminStatsResponse | null>(null);

  // Users data
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [usersLoading, setUsersLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive' | 'suspended'>('all');
  const [roleFilter, setRoleFilter] = useState<'all' | 'user' | 'admin'>('all');

  // Audit logs data
  const [auditLogs, setAuditLogs] = useState<EnhancedAuditLogEntry[]>([]);
  const [auditLoading, setAuditLoading] = useState(false);
  const [auditFilters] = useState<AdminAuditLogRequest>({
    page: 1,
    page_size: 20,
  });
  const [dateRange, setDateRange] = useState({ from: '', to: '' });
  const [selectedAction, setSelectedAction] = useState<string>('all');
  const [selectedLevel, setSelectedLevel] = useState<'all' | 'info' | 'warning' | 'error'>('all');
  const [auditSearchTerm, setAuditSearchTerm] = useState('');

  // Check if user is admin
  if (user?.role !== 'admin') {
    return (
      <Container>
        <div className="text-center py-12">
          <div className="text-red-600 mb-4">Access Denied</div>
          <div className="text-gray-600">You don't have permission to access the admin panel.</div>
        </div>
      </Container>
    );
  }

  const fetchStats = async () => {
    try {
      setLoading(true);
      const data = await api.getAdminStats();
      setStats(data);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Failed to load admin stats');
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchAuditLogs = async () => {
    try {
      setAuditLoading(true);
      const params: AdminAuditLogRequest = {
        ...auditFilters,
        action: selectedAction !== 'all' ? selectedAction : undefined,
        level: selectedLevel !== 'all' ? selectedLevel : undefined,
        date_from: dateRange.from || undefined,
        date_to: dateRange.to || undefined,
      };
      
      const data = await api.getAuditLogs(params);
      setAuditLogs(data.logs);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Failed to load audit logs');
      }
    } finally {
      setAuditLoading(false);
    }
  };

  const fetchUsers = async () => {
    try {
      setUsersLoading(true);
      const params = {
        search: searchTerm || undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined,
        role: roleFilter !== 'all' ? roleFilter : undefined,
        page: 1,
        page_size: 50,
      };
      
      const data = await api.getUsers(params);
      setUsers(data.users);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Failed to load users');
      }
    } finally {
      setUsersLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'dashboard') {
      fetchStats();
    } else if (activeTab === 'users') {
      fetchUsers();
    } else if (activeTab === 'audit') {
      fetchAuditLogs();
    }
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'users') {
      const debounceTimer = setTimeout(() => {
        fetchUsers();
      }, 300);
      return () => clearTimeout(debounceTimer);
    }
  }, [searchTerm, statusFilter, roleFilter]);

  useEffect(() => {
    if (activeTab === 'audit') {
      const debounceTimer = setTimeout(() => {
        fetchAuditLogs();
      }, 300);
      return () => clearTimeout(debounceTimer);
    }
  }, [selectedAction, selectedLevel, dateRange, auditSearchTerm]);

  const handleUserAction = async (userId: number, action: 'activate' | 'deactivate' | 'suspend') => {
    try {
      const reason = prompt(`Please provide a reason for ${action}ing this user:`);
      if (!reason) return;

      const statusMap = {
        activate: 'active' as const,
        deactivate: 'inactive' as const,
        suspend: 'suspended' as const,
      };

      await api.updateUserStatus(userId, {
        status: statusMap[action],
        reason,
      });

      fetchUsers(); // Refresh the user list
    } catch (err) {
      if (err instanceof ApiError) {
        alert(`Error: ${err.message}`);
      } else {
        alert('Failed to update user status');
      }
    }
  };

  const getLevelColor = (level: 'info' | 'warning' | 'error') => {
    switch (level) {
      case 'info':
        return 'bg-blue-100 text-blue-800';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800';
      case 'error':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getActionColor = (action: string) => {
    if (action.includes('login') || action.includes('auth')) {
      return 'bg-green-100 text-green-800';
    } else if (action.includes('delete') || action.includes('suspend')) {
      return 'bg-red-100 text-red-800';
    } else if (action.includes('update') || action.includes('change')) {
      return 'bg-blue-100 text-blue-800';
    } else if (action.includes('create')) {
      return 'bg-purple-100 text-purple-800';
    }
    return 'bg-gray-100 text-gray-800';
  };

  const tabs = [
    { id: 'dashboard' as const, name: 'Dashboard', icon: '📊' },
    { id: 'users' as const, name: 'Users', icon: '👥' },
    { id: 'audit' as const, name: 'Audit Logs', icon: '📋' },
  ];

  return (
    <Container>
      <div className="py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Admin Panel</h1>
          <p className="text-gray-600 mt-2">
            Manage users, view system statistics, and monitor activity.
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 text-red-800 rounded-md">
            {error}
            <button
              onClick={() => setError(null)}
              className="ml-2 text-red-600 hover:text-red-800"
            >
              ×
            </button>
          </div>
        )}

        <div className="bg-white shadow rounded-lg">
          {/* Tabs */}
          <div className="border-b border-gray-200">
            <nav className="-mb-px flex space-x-8 px-6">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <span className="mr-2">{tab.icon}</span>
                  {tab.name}
                </button>
              ))}
            </nav>
          </div>

          {/* Tab Content */}
          <div className="p-6">
            {/* Dashboard Tab */}
            {activeTab === 'dashboard' && (
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-6">System Overview</h3>
                
                {loading ? (
                  <div className="flex items-center justify-center py-12">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                  </div>
                ) : stats ? (
                  <>
                    {/* Stats Grid */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                      <div className="bg-gray-50 rounded-lg p-6">
                        <div className="text-sm font-medium text-gray-500">Total Users</div>
                        <div className="text-3xl font-bold text-gray-900">{stats.total_users}</div>
                      </div>
                      <div className="bg-green-50 rounded-lg p-6">
                        <div className="text-sm font-medium text-green-600">Active Users</div>
                        <div className="text-3xl font-bold text-green-900">{stats.active_users}</div>
                      </div>
                      <div className="bg-yellow-50 rounded-lg p-6">
                        <div className="text-sm font-medium text-yellow-600">Inactive Users</div>
                        <div className="text-3xl font-bold text-yellow-900">{stats.inactive_users}</div>
                      </div>
                      <div className="bg-red-50 rounded-lg p-6">
                        <div className="text-sm font-medium text-red-600">Suspended Users</div>
                        <div className="text-3xl font-bold text-red-900">{stats.suspended_users}</div>
                      </div>
                    </div>

                    {/* Additional Stats */}
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                      <div className="bg-white border rounded-lg p-6">
                        <div className="text-sm font-medium text-gray-500">Admin Users</div>
                        <div className="text-2xl font-bold text-gray-900">{stats.admin_users}</div>
                      </div>
                      <div className="bg-white border rounded-lg p-6">
                        <div className="text-sm font-medium text-gray-500">New Users Today</div>
                        <div className="text-2xl font-bold text-gray-900">{stats.new_users_today}</div>
                      </div>
                      <div className="bg-white border rounded-lg p-6">
                        <div className="text-sm font-medium text-gray-500">New Users This Week</div>
                        <div className="text-2xl font-bold text-gray-900">{stats.new_users_this_week}</div>
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="text-center text-gray-500 py-12">
                    No data available
                  </div>
                )}
              </div>
            )}

            {/* Users Tab */}
            {activeTab === 'users' && (
              <div>
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6">
                  <h3 className="text-lg font-medium text-gray-900">User Management</h3>
                  <Button onClick={fetchUsers} disabled={usersLoading}>
                    {usersLoading ? 'Loading...' : 'Refresh'}
                  </Button>
                </div>

                {/* Filters */}
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Search</label>
                    <input
                      type="text"
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      placeholder="Search users..."
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(e.target.value as any)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="all">All Status</option>
                      <option value="active">Active</option>
                      <option value="inactive">Inactive</option>
                      <option value="suspended">Suspended</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
                    <select
                      value={roleFilter}
                      onChange={(e) => setRoleFilter(e.target.value as any)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="all">All Roles</option>
                      <option value="user">User</option>
                      <option value="admin">Admin</option>
                    </select>
                  </div>
                </div>

                {/* Users Table */}
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          User
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Role
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Status
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Last Login
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {users.map((user) => (
                        <tr key={user.id}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center">
                              <div className="flex-shrink-0 h-10 w-10">
                                {user.avatar ? (
                                  <img 
                                    className="h-10 w-10 rounded-full" 
                                    src={user.avatar} 
                                    alt={`${user.first_name} ${user.last_name}`}
                                  />
                                ) : (
                                  <div className="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                                    <span className="text-sm font-medium text-gray-700">
                                      {user.first_name[0]}{user.last_name[0]}
                                    </span>
                                  </div>
                                )}
                              </div>
                              <div className="ml-4">
                                <div className="text-sm font-medium text-gray-900">
                                  {user.first_name} {user.last_name}
                                </div>
                                <div className="text-sm text-gray-500">
                                  {user.email}
                                </div>
                              </div>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              user.role === 'admin' 
                                ? 'bg-purple-100 text-purple-800' 
                                : 'bg-gray-100 text-gray-800'
                            }`}>
                              {user.role}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              user.status === 'active' 
                                ? 'bg-green-100 text-green-800' 
                                : user.status === 'suspended'
                                ? 'bg-red-100 text-red-800'
                                : 'bg-yellow-100 text-yellow-800'
                            }`}>
                              {user.status}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {user.last_login_at 
                              ? new Date(user.last_login_at).toLocaleDateString()
                              : 'Never'
                            }
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                            {user.status === 'active' ? (
                              <>
                                <button
                                  onClick={() => handleUserAction(user.id, 'deactivate')}
                                  className="text-yellow-600 hover:text-yellow-900"
                                >
                                  Deactivate
                                </button>
                                <button
                                  onClick={() => handleUserAction(user.id, 'suspend')}
                                  className="text-red-600 hover:text-red-900"
                                >
                                  Suspend
                                </button>
                              </>
                            ) : (
                              <button
                                onClick={() => handleUserAction(user.id, 'activate')}
                                className="text-green-600 hover:text-green-900"
                              >
                                Activate
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  
                  {users.length === 0 && !usersLoading && (
                    <div className="text-center text-gray-500 py-8">
                      No users found
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Audit Tab */}
            {activeTab === 'audit' && (
              <div>
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6">
                  <h3 className="text-lg font-medium text-gray-900">Audit Logs</h3>
                  <Button onClick={fetchAuditLogs} disabled={auditLoading}>
                    {auditLoading ? 'Loading...' : 'Refresh'}
                  </Button>
                </div>

                {/* Summary Cards */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                  <div className="bg-blue-50 rounded-lg p-4">
                    <div className="text-sm font-medium text-blue-600">Total Events</div>
                    <div className="text-2xl font-bold text-blue-900">{auditLogs.length}</div>
                  </div>
                  <div className="bg-green-50 rounded-lg p-4">
                    <div className="text-sm font-medium text-green-600">Info Events</div>
                    <div className="text-2xl font-bold text-green-900">
                      {auditLogs.filter(log => log.level === 'info').length}
                    </div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-4">
                    <div className="text-sm font-medium text-yellow-600">Warning Events</div>
                    <div className="text-2xl font-bold text-yellow-900">
                      {auditLogs.filter(log => log.level === 'warning').length}
                    </div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-4">
                    <div className="text-sm font-medium text-red-600">Error Events</div>
                    <div className="text-2xl font-bold text-red-900">
                      {auditLogs.filter(log => log.level === 'error').length}
                    </div>
                  </div>
                </div>

                {/* Filters */}
                <div className="bg-gray-50 rounded-lg p-4 mb-6">
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Date From</label>
                      <input
                        type="date"
                        value={dateRange.from}
                        onChange={(e) => setDateRange({ ...dateRange, from: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Date To</label>
                      <input
                        type="date"
                        value={dateRange.to}
                        onChange={(e) => setDateRange({ ...dateRange, to: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Action</label>
                      <select
                        value={selectedAction}
                        onChange={(e) => setSelectedAction(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="all">All Actions</option>
                        <option value="login">Login</option>
                        <option value="logout">Logout</option>
                        <option value="create">Create</option>
                        <option value="update">Update</option>
                        <option value="delete">Delete</option>
                        <option value="suspend">Suspend</option>
                        <option value="activate">Activate</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Level</label>
                      <select
                        value={selectedLevel}
                        onChange={(e) => setSelectedLevel(e.target.value as any)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="all">All Levels</option>
                        <option value="info">Info</option>
                        <option value="warning">Warning</option>
                        <option value="error">Error</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Search</label>
                      <input
                        type="text"
                        value={auditSearchTerm}
                        onChange={(e) => setAuditSearchTerm(e.target.value)}
                        placeholder="Search logs..."
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  </div>
                </div>

                {/* Audit Logs Table */}
                <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                  {auditLoading ? (
                    <div className="flex items-center justify-center py-12">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                    </div>
                  ) : (
                    <>
                      <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Timestamp
                              </th>
                              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                User
                              </th>
                              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Action
                              </th>
                              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Level
                              </th>
                              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                IP Address
                              </th>
                              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Description
                              </th>
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {auditLogs.map((log) => (
                              <tr key={log.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                  {new Date(log.created_at).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                  {log.user ? (
                                    <div className="flex items-center">
                                      <div className="flex-shrink-0 h-8 w-8">
                                        {log.user.avatar ? (
                                          <img 
                                            className="h-8 w-8 rounded-full" 
                                            src={log.user.avatar} 
                                            alt={`${log.user.first_name} ${log.user.last_name}`}
                                          />
                                        ) : (
                                          <div className="h-8 w-8 rounded-full bg-gray-300 flex items-center justify-center">
                                            <span className="text-xs font-medium text-gray-700">
                                              {log.user.first_name[0]}{log.user.last_name[0]}
                                            </span>
                                          </div>
                                        )}
                                      </div>
                                      <div className="ml-3">
                                        <div className="text-sm font-medium text-gray-900">
                                          {log.user.first_name} {log.user.last_name}
                                        </div>
                                        <div className="text-sm text-gray-500">
                                          {log.user.email}
                                        </div>
                                      </div>
                                    </div>
                                  ) : (
                                    <span className="text-sm text-gray-500">System</span>
                                  )}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                    getActionColor(log.action)
                                  }`}>
                                    {log.action}
                                  </span>
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                    getLevelColor(log.level)
                                  }`}>
                                    {log.level.toUpperCase()}
                                  </span>
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                  {log.ip_address}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-900">
                                  <div className="max-w-xs truncate" title={log.description}>
                                    {log.description}
                                  </div>
                                  {log.metadata && Object.keys(log.metadata).length > 0 && (
                                    <div className="mt-1">
                                      <button 
                                        className="text-xs text-blue-600 hover:text-blue-800"
                                        onClick={() => {
                                          alert(`Metadata: ${JSON.stringify(log.metadata, null, 2)}`);
                                        }}
                                      >
                                        View Details
                                      </button>
                                    </div>
                                  )}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                      
                      {auditLogs.length === 0 && !auditLoading && (
                        <div className="text-center text-gray-500 py-8">
                          No audit logs found
                        </div>
                      )}
                    </>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </Container>
  );
};

export default AdminPanel;