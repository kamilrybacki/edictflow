'use client';

import { useState, useEffect } from 'react';
import {
  Role,
  Permission,
  fetchRoles,
  fetchPermissions,
  createRole,
  addRolePermission,
  removeRolePermission,
} from '@/lib/api';

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newRole, setNewRole] = useState({ name: '', description: '', hierarchy_level: 10 });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [rolesData, permsData] = await Promise.all([fetchRoles(), fetchPermissions()]);
      setRoles(rolesData);
      setPermissions(permsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRole = async () => {
    try {
      await createRole(newRole);
      setShowCreateModal(false);
      setNewRole({ name: '', description: '', hierarchy_level: 10 });
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create role');
    }
  };

  const handleTogglePermission = async (roleId: string, permissionId: string, hasPermission: boolean) => {
    try {
      if (hasPermission) {
        await removeRolePermission(roleId, permissionId);
      } else {
        await addRolePermission(roleId, permissionId);
      }
      await loadData();
      // Refresh selected role
      const updated = roles.find((r) => r.id === roleId);
      if (updated) setSelectedRole(updated);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update permission');
    }
  };

  // Group permissions by category
  const permissionsByCategory = permissions.reduce((acc, perm) => {
    if (!acc[perm.category]) {
      acc[perm.category] = [];
    }
    acc[perm.category].push(perm);
    return acc;
  }, {} as Record<string, Permission[]>);

  if (loading) {
    return <div className="text-zinc-600 dark:text-zinc-400">Loading roles...</div>;
  }

  if (error) {
    return <div className="text-red-600">{error}</div>;
  }

  return (
    <div className="flex gap-6">
      {/* Roles List */}
      <div className="w-80 flex-shrink-0">
        <div className="flex justify-between items-center mb-4">
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">Roles</h1>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-md"
          >
            Add Role
          </button>
        </div>

        <div className="space-y-2">
          {roles.map((role) => (
            <button
              key={role.id}
              onClick={() => setSelectedRole(role)}
              className={`w-full text-left px-4 py-3 rounded-lg border transition-colors ${
                selectedRole?.id === role.id
                  ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                  : 'border-zinc-200 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-800'
              }`}
            >
              <div className="font-medium text-zinc-900 dark:text-white">{role.name}</div>
              <div className="text-sm text-zinc-500">{role.description}</div>
              <div className="text-xs text-zinc-400 mt-1">
                {role.permissions.length} permissions
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Permission Editor */}
      <div className="flex-1">
        {selectedRole ? (
          <div className="bg-white dark:bg-zinc-800 rounded-lg border border-zinc-200 dark:border-zinc-700 p-6">
            <h2 className="text-xl font-semibold text-zinc-900 dark:text-white mb-4">
              {selectedRole.name} Permissions
            </h2>

            <div className="space-y-6">
              {Object.entries(permissionsByCategory).map(([category, perms]) => (
                <div key={category}>
                  <h3 className="text-sm font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider mb-2">
                    {category}
                  </h3>
                  <div className="space-y-2">
                    {perms.map((perm) => {
                      const hasPermission = selectedRole.permissions.includes(perm.name);
                      return (
                        <label
                          key={perm.id}
                          className="flex items-start gap-3 p-3 rounded-md border border-zinc-200 dark:border-zinc-600 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-700"
                        >
                          <input
                            type="checkbox"
                            checked={hasPermission}
                            onChange={() =>
                              handleTogglePermission(selectedRole.id, perm.id, hasPermission)
                            }
                            className="mt-1 rounded border-zinc-300 text-blue-600 focus:ring-blue-500"
                          />
                          <div>
                            <div className="font-medium text-zinc-900 dark:text-white">
                              {perm.name}
                            </div>
                            <div className="text-sm text-zinc-500">{perm.description}</div>
                          </div>
                        </label>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div className="bg-white dark:bg-zinc-800 rounded-lg border border-zinc-200 dark:border-zinc-700 p-6 text-center text-zinc-500">
            Select a role to manage its permissions
          </div>
        )}
      </div>

      {/* Create Role Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-zinc-800 rounded-lg p-6 max-w-md w-full mx-4">
            <h2 className="text-lg font-semibold text-zinc-900 dark:text-white mb-4">
              Create New Role
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Name
                </label>
                <input
                  type="text"
                  value={newRole.name}
                  onChange={(e) => setNewRole({ ...newRole, name: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md dark:bg-zinc-700 dark:text-white"
                  placeholder="Role name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Description
                </label>
                <input
                  type="text"
                  value={newRole.description}
                  onChange={(e) => setNewRole({ ...newRole, description: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md dark:bg-zinc-700 dark:text-white"
                  placeholder="Role description"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Hierarchy Level (1 = highest)
                </label>
                <input
                  type="number"
                  min="1"
                  value={newRole.hierarchy_level}
                  onChange={(e) =>
                    setNewRole({ ...newRole, hierarchy_level: parseInt(e.target.value) || 10 })
                  }
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md dark:bg-zinc-700 dark:text-white"
                />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowCreateModal(false)}
                className="flex-1 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md text-zinc-700 dark:text-zinc-300"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateRole}
                className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
