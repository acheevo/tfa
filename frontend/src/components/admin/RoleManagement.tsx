import React, { useState } from 'react';
import { useRBAC, PERMISSIONS } from '../../contexts/RBACContext';
import { api, User, ApiError } from '../../lib/api';
import { AdminGuard, PermissionGuard } from '../auth/RoleGuard';

interface RoleManagementProps {
  user: User;
  onRoleChanged: () => void;
}

interface RoleChangeModalProps {
  user: User;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

// Role change modal component
const RoleChangeModal: React.FC<RoleChangeModalProps> = ({
  user,
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [selectedRole, setSelectedRole] = useState<'user' | 'admin'>(user.role);
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!reason.trim()) {
      setError('Please provide a reason for the role change');
      return;
    }

    if (selectedRole === user.role) {
      setError('Please select a different role');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await api.updateUserRole(user.id, {
        role: selectedRole,
        reason: reason.trim(),
      });

      onSuccess();
      onClose();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Failed to update user role');
      }
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setSelectedRole(user.role);
    setReason('');
    setError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="mt-3">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Change User Role
          </h3>
          
          <div className="mb-4 p-3 bg-gray-50 rounded-md">
            <div className="text-sm text-gray-600">User:</div>
            <div className="font-medium">{user.first_name} {user.last_name}</div>
            <div className="text-sm text-gray-500">{user.email}</div>
            <div className="text-sm">
              Current role: <span className="font-medium capitalize">{user.role}</span>
            </div>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                New Role
              </label>
              <select
                value={selectedRole}
                onChange={(e) => setSelectedRole(e.target.value as 'user' | 'admin')}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                disabled={loading}
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Reason <span className="text-red-500">*</span>
              </label>
              <textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Please provide a reason for this role change..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                rows={3}
                disabled={loading}
                required
              />
            </div>

            {error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                <div className="text-red-600 text-sm">{error}</div>
              </div>
            )}

            <div className="flex space-x-3">
              <button
                type="submit"
                disabled={loading || !reason.trim() || selectedRole === user.role}
                className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Updating...' : 'Update Role'}
              </button>
              <button
                type="button"
                onClick={handleClose}
                disabled={loading}
                className="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

// Role badge component
export const RoleBadge: React.FC<{ role: 'user' | 'admin'; className?: string }> = ({ 
  role, 
  className = '' 
}) => {
  const baseClasses = "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium";
  const roleClasses = {
    user: "bg-green-100 text-green-800",
    admin: "bg-purple-100 text-purple-800",
  };

  return (
    <span className={`${baseClasses} ${roleClasses[role]} ${className}`}>
      {role.charAt(0).toUpperCase() + role.slice(1)}
    </span>
  );
};

// Role management actions component
export const RoleManagementActions: React.FC<RoleManagementProps> = ({
  user,
  onRoleChanged,
}) => {
  const [showRoleModal, setShowRoleModal] = useState(false);
  const { canManageUser } = useRBAC();

  if (!canManageUser(user.id)) {
    return null;
  }

  return (
    <PermissionGuard permission={PERMISSIONS.USER_MANAGE}>
      <div className="flex items-center space-x-2">
        <RoleBadge role={user.role} />
        <button
          onClick={() => setShowRoleModal(true)}
          className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          title="Change user role"
        >
          Change Role
        </button>
      </div>

      <RoleChangeModal
        user={user}
        isOpen={showRoleModal}
        onClose={() => setShowRoleModal(false)}
        onSuccess={onRoleChanged}
      />
    </PermissionGuard>
  );
};

// Bulk role management component
interface BulkRoleManagementProps {
  selectedUsers: User[];
  onRoleChanged: () => void;
  onClearSelection: () => void;
}

export const BulkRoleManagement: React.FC<BulkRoleManagementProps> = ({
  selectedUsers,
  onRoleChanged,
  onClearSelection,
}) => {
  const [selectedRole, setSelectedRole] = useState<'user' | 'admin'>('user');
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (selectedUsers.length === 0) {
    return null;
  }

  const handleBulkRoleChange = async () => {
    if (!reason.trim()) {
      setError('Please provide a reason for the role change');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await api.bulkUpdateUsers({
        user_ids: selectedUsers.map(u => u.id),
        action: 'role_change',
        role: selectedRole,
        reason: reason.trim(),
      });

      onRoleChanged();
      onClearSelection();
      setReason('');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Failed to update user roles');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <AdminGuard>
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-blue-900">
            Bulk Role Management
          </h3>
          <button
            onClick={onClearSelection}
            className="text-blue-600 hover:text-blue-800 text-sm"
          >
            Clear Selection
          </button>
        </div>

        <div className="text-sm text-blue-700 mb-4">
          {selectedUsers.length} user{selectedUsers.length > 1 ? 's' : ''} selected
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-blue-700 mb-1">
              New Role
            </label>
            <select
              value={selectedRole}
              onChange={(e) => setSelectedRole(e.target.value as 'user' | 'admin')}
              className="w-full px-3 py-2 border border-blue-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={loading}
            >
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-blue-700 mb-1">
              Reason
            </label>
            <input
              type="text"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Reason for role change..."
              className="w-full px-3 py-2 border border-blue-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={loading}
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={handleBulkRoleChange}
              disabled={loading || !reason.trim()}
              className="w-full bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Updating...' : 'Update Roles'}
            </button>
          </div>
        </div>

        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
            <div className="text-red-600 text-sm">{error}</div>
          </div>
        )}
      </div>
    </AdminGuard>
  );
};

// Role statistics component
export const RoleStatistics: React.FC<{ stats: any }> = ({ stats }) => {
  return (
    <PermissionGuard permission={PERMISSIONS.ADMIN_READ}>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        <div className="bg-white p-6 rounded-lg shadow border">
          <div className="flex items-center">
            <div className="p-2 bg-green-100 rounded-lg">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
              </svg>
            </div>
            <div className="ml-4">
              <div className="text-sm font-medium text-gray-500">Regular Users</div>
              <div className="text-2xl font-bold text-gray-900">
                {stats?.total_users - stats?.admin_users || 0}
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow border">
          <div className="flex items-center">
            <div className="p-2 bg-purple-100 rounded-lg">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <div className="ml-4">
              <div className="text-sm font-medium text-gray-500">Administrators</div>
              <div className="text-2xl font-bold text-gray-900">
                {stats?.admin_users || 0}
              </div>
            </div>
          </div>
        </div>
      </div>
    </PermissionGuard>
  );
};

export default RoleManagementActions;