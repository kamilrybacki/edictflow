'use client';

import { cn } from '@/lib/utils';

interface User {
  id: string;
  name: string;
  avatar?: string;
}

interface AvatarStackProps {
  users: User[];
  max?: number;
  size?: 'sm' | 'md' | 'lg';
}

const sizeClasses = {
  sm: 'w-6 h-6 text-[10px]',
  md: 'w-8 h-8 text-xs',
  lg: 'w-10 h-10 text-sm',
};

export function AvatarStack({ users, max = 4, size = 'md' }: AvatarStackProps) {
  const displayUsers = users.slice(0, max);
  const remaining = users.length - max;

  return (
    <div className="flex -space-x-2">
      {displayUsers.map((user, index) => (
        <div
          key={user.id}
          className={cn(
            sizeClasses[size],
            'rounded-full bg-gradient-to-br from-primary to-primary/70 flex items-center justify-center font-medium text-primary-foreground ring-2 ring-background transition-transform hover:scale-110 hover:z-10'
          )}
          style={{ zIndex: displayUsers.length - index }}
          title={user.name}
        >
          {user.avatar ? (
            <img src={user.avatar} alt={user.name} className="w-full h-full rounded-full object-cover" />
          ) : (
            user.name.split(' ').map(n => n[0]).join('').slice(0, 2)
          )}
        </div>
      ))}
      {remaining > 0 && (
        <div
          className={cn(
            sizeClasses[size],
            'rounded-full bg-muted flex items-center justify-center font-medium text-muted-foreground ring-2 ring-background'
          )}
          style={{ zIndex: 0 }}
        >
          +{remaining}
        </div>
      )}
    </div>
  );
}
