'use client';

import Link from 'next/link';

const adminSections = [
  {
    title: 'Users',
    description: 'Manage user accounts, roles, and permissions',
    href: '/admin/users',
    icon: 'users',
  },
  {
    title: 'Roles',
    description: 'Configure roles and their permissions',
    href: '/admin/roles',
    icon: 'shield',
  },
  {
    title: 'Audit Log',
    description: 'View system activity and changes',
    href: '/admin/audit',
    icon: 'list',
  },
];

export default function AdminPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-6">
        Administration
      </h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {adminSections.map((section) => (
          <Link
            key={section.href}
            href={section.href}
            className="block p-6 bg-white dark:bg-zinc-800 rounded-lg border border-zinc-200 dark:border-zinc-700 hover:border-blue-500 dark:hover:border-blue-500 transition-colors"
          >
            <h2 className="text-lg font-semibold text-zinc-900 dark:text-white mb-2">
              {section.title}
            </h2>
            <p className="text-zinc-600 dark:text-zinc-400 text-sm">
              {section.description}
            </p>
          </Link>
        ))}
      </div>
    </div>
  );
}
