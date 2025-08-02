import React, { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { api, ApiError, UpdateProfileRequest, ChangeEmailRequest, UserPreferences, UpdatePreferencesRequest } from '../../lib/api';
import { Button } from '../../components/ui/Button';
import { User, Shield, Bell, Settings as SettingsIcon, Globe, Monitor, Key, Smartphone, LogOut } from 'lucide-react';

export const Settings: React.FC = () => {
  const { user, refreshAuth } = useAuth();
  const [activeTab, setActiveTab] = useState<'account' | 'preferences' | 'notifications' | 'security'>('account');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // Profile form state
  const [profileData, setProfileData] = useState<UpdateProfileRequest>({
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    avatar: user?.avatar || '',
  });

  // Email change form state
  const [emailData, setEmailData] = useState<ChangeEmailRequest>({
    new_email: '',
    password: '',
  });

  // Preferences state
  const [preferences, setPreferences] = useState<UserPreferences>({
    theme: 'light',
    language: 'en',
    timezone: '',
    notifications: {
      email: true,
      sms: false,
      push: true,
    },
    privacy: {
      profile_visible: true,
      show_email: false,
    },
    custom: {},
  });

  useEffect(() => {
    if (user) {
      setProfileData({
        first_name: user.first_name,
        last_name: user.last_name,
        avatar: user.avatar || '',
      });
      
      if (user.preferences) {
        setPreferences({
          theme: user.preferences.theme || 'light',
          language: user.preferences.language || 'en',
          timezone: user.preferences.timezone || '',
          notifications: user.preferences.notifications || {
            email: true,
            sms: false,
            push: true,
          },
          privacy: user.preferences.privacy || {
            profile_visible: true,
            show_email: false,
          },
          custom: user.preferences.custom || {},
        });
      }
    }
  }, [user]);

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);

    try {
      await api.updateProfile(profileData);
      await refreshAuth();
      setMessage({ type: 'success', text: 'Profile updated successfully!' });
    } catch (error) {
      if (error instanceof ApiError) {
        setMessage({ type: 'error', text: error.message });
      } else {
        setMessage({ type: 'error', text: 'Failed to update profile' });
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);

    try {
      await api.changeEmail(emailData);
      setMessage({ type: 'success', text: 'Email change requested. Please check your new email for verification.' });
      setEmailData({ new_email: '', password: '' });
    } catch (error) {
      if (error instanceof ApiError) {
        setMessage({ type: 'error', text: error.message });
      } else {
        setMessage({ type: 'error', text: 'Failed to change email' });
      }
    } finally {
      setLoading(false);
    }
  };

  const handlePreferencesSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);

    try {
      const updateData: UpdatePreferencesRequest = {
        theme: preferences.theme,
        language: preferences.language,
        timezone: preferences.timezone,
        notifications: preferences.notifications || {
          email: true,
          sms: false,
          push: true,
        },
        privacy: preferences.privacy || {
          profile_visible: true,
          show_email: false,
        },
        custom: preferences.custom,
      };

      await api.updatePreferences(updateData);
      await refreshAuth();
      setMessage({ type: 'success', text: 'Preferences updated successfully!' });
    } catch (error) {
      if (error instanceof ApiError) {
        setMessage({ type: 'error', text: error.message });
      } else {
        setMessage({ type: 'error', text: 'Failed to update preferences' });
      }
    } finally {
      setLoading(false);
    }
  };

  const tabs = [
    { id: 'account' as const, name: 'Account', icon: User },
    { id: 'preferences' as const, name: 'Preferences', icon: SettingsIcon },
    { id: 'notifications' as const, name: 'Notifications', icon: Bell },
    { id: 'security' as const, name: 'Security', icon: Shield },
  ];

  return (
    <div className="p-4 sm:p-8">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-secondary-900">Settings</h1>
          <p className="text-secondary-600 mt-2">
            Manage your account settings and preferences.
          </p>
        </div>

        {/* Message */}
        {message && (
          <div className={`mb-6 p-4 rounded-lg ${
            message.type === 'success' 
              ? 'bg-success-50 text-success-800 border border-success-200' 
              : 'bg-error-50 text-error-800 border border-error-200'
          }`}>
            {message.text}
          </div>
        )}

        <div className="bg-white shadow-sm border border-secondary-200 rounded-lg">
          {/* Tabs */}
          <div className="border-b border-secondary-200">
            <nav className="-mb-px flex space-x-8 px-6">
              {tabs.map((tab) => {
                const IconComponent = tab.icon;
                return (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`py-4 px-1 border-b-2 font-medium text-sm flex items-center ${
                      activeTab === tab.id
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
                    }`}
                  >
                    <IconComponent className="w-4 h-4 mr-2" />
                    {tab.name}
                  </button>
                );
              })}
            </nav>
          </div>

          {/* Tab Content */}
          <div className="p-6">
              {/* Account Tab */}
              {activeTab === 'account' && (
                <div className="space-y-6">
                  {/* Profile Information Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <User className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Profile Information</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Update your personal information and profile details.</p>
                    
                    <form onSubmit={handleProfileSubmit} className="space-y-4">
                      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                        <div>
                          <label htmlFor="first_name" className="block text-sm font-medium text-secondary-700 mb-1">
                            First Name
                          </label>
                          <input
                            type="text"
                            id="first_name"
                            value={profileData.first_name}
                            onChange={(e) => setProfileData({ ...profileData, first_name: e.target.value })}
                            className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                            required
                          />
                        </div>
                        <div>
                          <label htmlFor="last_name" className="block text-sm font-medium text-secondary-700 mb-1">
                            Last Name
                          </label>
                          <input
                            type="text"
                            id="last_name"
                            value={profileData.last_name}
                            onChange={(e) => setProfileData({ ...profileData, last_name: e.target.value })}
                            className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                            required
                          />
                        </div>
                      </div>
                      <div>
                        <label htmlFor="avatar" className="block text-sm font-medium text-secondary-700 mb-1">
                          Avatar URL (optional)
                        </label>
                        <input
                          type="url"
                          id="avatar"
                          value={profileData.avatar}
                          onChange={(e) => setProfileData({ ...profileData, avatar: e.target.value })}
                          className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                          placeholder="https://example.com/avatar.jpg"
                        />
                      </div>
                      <div className="flex justify-end">
                        <Button type="submit" disabled={loading}>
                          {loading ? 'Saving...' : 'Save Profile'}
                        </Button>
                      </div>
                    </form>
                  </div>

                  {/* Account Status Card */}
                  <div className="bg-secondary-50 border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Shield className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Account Status</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-4">View your account information and verification status.</p>
                    
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="space-y-3">
                        <div className="flex justify-between">
                          <span className="text-sm font-medium text-secondary-700">Email:</span>
                          <span className="text-sm text-secondary-900">{user?.email}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm font-medium text-secondary-700">Role:</span>
                          <span className="text-sm text-secondary-900 capitalize">{user?.role}</span>
                        </div>
                      </div>
                      <div className="space-y-3">
                        <div className="flex justify-between">
                          <span className="text-sm font-medium text-secondary-700">Status:</span>
                          <span className={`text-sm px-2 py-1 rounded-full text-xs font-medium ${
                            user?.status === 'active' 
                              ? 'bg-success-100 text-success-800' 
                              : 'bg-warning-100 text-warning-800'
                          }`}>
                            {user?.status}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm font-medium text-secondary-700">Email Verified:</span>
                          <span className={`text-sm px-2 py-1 rounded-full text-xs font-medium ${
                            user?.email_verified 
                              ? 'bg-success-100 text-success-800' 
                              : 'bg-error-100 text-error-800'
                          }`}>
                            {user?.email_verified ? 'Verified' : 'Unverified'}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Preferences Tab */}
              {activeTab === 'preferences' && (
                <div className="space-y-6">
                  {/* Theme Settings Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Monitor className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Theme & Display</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Customize how the application looks and feels.</p>
                    
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-secondary-700 mb-3">Theme</label>
                        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                          {[
                            { value: 'light', label: 'Light', icon: '☀️' },
                            { value: 'dark', label: 'Dark', icon: '🌙' },
                            { value: 'system', label: 'System', icon: '💻' }
                          ].map((theme) => (
                            <label key={theme.value} className="flex items-center p-3 border border-secondary-300 rounded-lg cursor-pointer hover:bg-secondary-50 transition-colors">
                              <input
                                type="radio"
                                name="theme"
                                value={theme.value}
                                checked={preferences.theme === theme.value}
                                onChange={(e) => setPreferences({ ...preferences, theme: e.target.value })}
                                className="mr-3"
                              />
                              <span className="mr-2">{theme.icon}</span>
                              <span className="font-medium">{theme.label}</span>
                            </label>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Language & Region Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Globe className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Language & Region</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Set your language and timezone preferences.</p>
                    
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div>
                        <label htmlFor="language" className="block text-sm font-medium text-secondary-700 mb-1">
                          Language
                        </label>
                        <select
                          id="language"
                          value={preferences.language}
                          onChange={(e) => setPreferences({ ...preferences, language: e.target.value })}
                          className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        >
                          <option value="en">English</option>
                          <option value="es">Spanish</option>
                          <option value="fr">French</option>
                          <option value="de">German</option>
                        </select>
                      </div>
                      <div>
                        <label htmlFor="timezone" className="block text-sm font-medium text-secondary-700 mb-1">
                          Timezone
                        </label>
                        <select
                          id="timezone"
                          value={preferences.timezone}
                          onChange={(e) => setPreferences({ ...preferences, timezone: e.target.value })}
                          className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        >
                          <option value="">Select timezone...</option>
                          <option value="UTC">UTC</option>
                          <option value="America/New_York">Eastern Time</option>
                          <option value="America/Chicago">Central Time</option>
                          <option value="America/Denver">Mountain Time</option>
                          <option value="America/Los_Angeles">Pacific Time</option>
                        </select>
                      </div>
                    </div>
                  </div>

                  {/* Privacy Settings Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Shield className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Privacy Settings</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Control how your information is shared with other users.</p>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Profile Visibility</div>
                          <div className="text-sm text-secondary-600">Make your profile visible to other users</div>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            checked={preferences.privacy?.profile_visible || false}
                            onChange={(e) => setPreferences({
                              ...preferences,
                              privacy: {
                                ...preferences.privacy!,
                                profile_visible: e.target.checked
                              }
                            })}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                        </label>
                      </div>
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Show Email Address</div>
                          <div className="text-sm text-secondary-600">Display your email address in your profile</div>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            checked={preferences.privacy?.show_email || false}
                            onChange={(e) => setPreferences({
                              ...preferences,
                              privacy: {
                                ...preferences.privacy!,
                                show_email: e.target.checked
                              }
                            })}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                        </label>
                      </div>
                    </div>
                  </div>

                  <div className="flex justify-end">
                    <Button onClick={handlePreferencesSubmit} disabled={loading}>
                      {loading ? 'Saving...' : 'Save Preferences'}
                    </Button>
                  </div>
                </div>
              )}

              {/* Notifications Tab */}
              {activeTab === 'notifications' && (
                <div className="space-y-6">
                  {/* Email Notifications Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Bell className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Email Notifications</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Choose what email notifications you'd like to receive.</p>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Account Activity</div>
                          <div className="text-sm text-secondary-600">Get notified about login attempts and security events</div>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            checked={preferences.notifications?.email || false}
                            onChange={(e) => setPreferences({
                              ...preferences,
                              notifications: {
                                ...preferences.notifications!,
                                email: e.target.checked
                              }
                            })}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                        </label>
                      </div>
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Product Updates</div>
                          <div className="text-sm text-secondary-600">Receive notifications about new features and updates</div>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            defaultChecked={false}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                        </label>
                      </div>
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Marketing Communications</div>
                          <div className="text-sm text-secondary-600">Receive promotional emails and newsletters</div>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            defaultChecked={false}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                        </label>
                      </div>
                    </div>
                  </div>

                  {/* Push Notifications Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Smartphone className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Push Notifications</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Manage push notification preferences for your devices.</p>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Real-time Alerts</div>
                          <div className="text-sm text-secondary-600">Get instant notifications for important events</div>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            checked={preferences.notifications?.push || false}
                            onChange={(e) => setPreferences({
                              ...preferences,
                              notifications: {
                                ...preferences.notifications!,
                                push: e.target.checked
                              }
                            })}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                        </label>
                      </div>
                    </div>
                  </div>

                  {/* Notification Frequency Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <SettingsIcon className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Notification Frequency</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Control how often you receive notifications.</p>
                    
                    <div>
                      <label htmlFor="frequency" className="block text-sm font-medium text-secondary-700 mb-2">
                        Email Frequency
                      </label>
                      <select
                        id="frequency"
                        className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      >
                        <option value="instant">Instant</option>
                        <option value="daily">Daily Digest</option>
                        <option value="weekly">Weekly Summary</option>
                        <option value="never">Never</option>
                      </select>
                    </div>
                  </div>

                  <div className="flex justify-end">
                    <Button onClick={handlePreferencesSubmit} disabled={loading}>
                      {loading ? 'Saving...' : 'Save Notification Settings'}
                    </Button>
                  </div>
                </div>
              )}

              {/* Security Tab */}
              {activeTab === 'security' && (
                <div className="space-y-6">
                  {/* Change Email Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <User className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Change Email Address</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Update your email address. You'll need to verify the new address.</p>
                    
                    <form onSubmit={handleEmailSubmit} className="space-y-4">
                      <div>
                        <label htmlFor="new_email" className="block text-sm font-medium text-secondary-700 mb-1">
                          New Email Address
                        </label>
                        <input
                          type="email"
                          id="new_email"
                          value={emailData.new_email}
                          onChange={(e) => setEmailData({ ...emailData, new_email: e.target.value })}
                          className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                          required
                        />
                      </div>
                      <div>
                        <label htmlFor="password_confirm" className="block text-sm font-medium text-secondary-700 mb-1">
                          Current Password
                        </label>
                        <input
                          type="password"
                          id="password_confirm"
                          value={emailData.password}
                          onChange={(e) => setEmailData({ ...emailData, password: e.target.value })}
                          className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                          required
                        />
                      </div>
                      <div className="flex justify-end">
                        <Button type="submit" disabled={loading}>
                          {loading ? 'Changing...' : 'Change Email'}
                        </Button>
                      </div>
                    </form>
                  </div>

                  {/* Password Security Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Key className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Password Security</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Manage your password and authentication settings.</p>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-secondary-50 rounded-lg">
                        <div>
                          <div className="font-medium text-secondary-900">Password</div>
                          <div className="text-sm text-secondary-600">Last changed 30 days ago</div>
                        </div>
                        <Button
                          variant="secondary"
                          onClick={() => window.location.href = '/change-password'}
                        >
                          Change Password
                        </Button>
                      </div>
                    </div>
                  </div>

                  {/* Two-Factor Authentication Card (Placeholder) */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Smartphone className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Two-Factor Authentication</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Add an extra layer of security to your account.</p>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-warning-50 border border-warning-200 rounded-lg">
                        <div>
                          <div className="font-medium text-warning-900">Two-Factor Authentication</div>
                          <div className="text-sm text-warning-700">Not enabled - Coming soon</div>
                        </div>
                        <Button variant="secondary" disabled>
                          Enable 2FA
                        </Button>
                      </div>
                    </div>
                  </div>

                  {/* Active Sessions Card (Placeholder) */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <Monitor className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Active Sessions</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Monitor and manage your active login sessions.</p>
                    
                    <div className="space-y-4">
                      <div className="p-4 bg-info-50 border border-info-200 rounded-lg">
                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-medium text-info-900">Current Session</div>
                            <div className="text-sm text-info-700">This device - Active now</div>
                          </div>
                          <div className="px-3 py-1 bg-success-100 text-success-800 rounded-full text-xs font-medium">
                            Current
                          </div>
                        </div>
                      </div>
                      <div className="p-4 bg-secondary-50 rounded-lg">
                        <div className="text-sm text-secondary-600 text-center">
                          Active session management coming soon
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Security Actions Card */}
                  <div className="bg-white border border-secondary-200 rounded-lg p-6">
                    <div className="flex items-center mb-4">
                      <LogOut className="w-5 h-5 text-secondary-600 mr-2" />
                      <h3 className="text-lg font-medium text-secondary-900">Security Actions</h3>
                    </div>
                    <p className="text-sm text-secondary-600 mb-6">Take immediate security actions for your account.</p>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-error-50 border border-error-200 rounded-lg">
                        <div>
                          <div className="font-medium text-error-900">Logout All Devices</div>
                          <div className="text-sm text-error-700">Sign out of all devices except this one</div>
                        </div>
                        <Button
                          variant="secondary"
                          onClick={() => {
                            if (confirm('Are you sure you want to logout from all other devices?')) {
                              api.logoutAll();
                            }
                          }}
                        >
                          Logout All
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;