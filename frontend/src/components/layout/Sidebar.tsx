import { Link, useLocation } from 'react-router-dom';
import { Home, Settings, Shield } from 'lucide-react';
import { useRBAC } from '../../contexts/RBACContext';

interface SidebarItemProps {
  to: string;
  icon: React.ElementType;
  label: string;
  isActive?: boolean;
}

const SidebarItem: React.FC<SidebarItemProps> = ({ to, icon: Icon, label, isActive }) => (
  <Link
    to={to}
    className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
      isActive 
        ? 'bg-primary-100 text-primary-700' 
        : 'text-secondary-600 hover:text-secondary-900 hover:bg-secondary-100'
    }`}
  >
    <Icon className="w-5 h-5 mr-3" />
    {label}
  </Link>
);

interface SidebarProps {
  className?: string;
}

export default function Sidebar({ className = '' }: SidebarProps) {
  const location = useLocation();
  const { isAdmin } = useRBAC();

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className={`bg-white border-r border-secondary-200 ${className}`}>
      <nav className="p-4 space-y-1">
        {/* Main Navigation */}
        <SidebarItem
          to="/dashboard"
          icon={Home}
          label="Dashboard"
          isActive={isActive('/dashboard')}
        />
        
        <SidebarItem
          to="/settings"
          icon={Settings}
          label="Settings"
          isActive={isActive('/settings')}
        />

        {/* Admin Section - At Bottom */}
        {isAdmin() && (
          <div className="pt-4 mt-4 border-t border-secondary-200">
            <SidebarItem
              to="/admin"
              icon={Shield}
              label="Admin"
              isActive={isActive('/admin')}
            />
          </div>
        )}
      </nav>
    </div>
  );
}