import React from 'react';
import { useNavigate } from 'react-router';
import { useStore } from '@nanostores/react';
import { $currentTenant } from '../stores/tenants';

const ChannelCard = ({ channel, userRole, onManageMembers, onSettings }) => {
  const navigate = useNavigate();
  const currentTenant = useStore($currentTenant);

  const handleViewChannel = () => {
    navigate(`/channels/${channel.id}`);
  };

  const canManageChannel = userRole === 'owner';
  const isPersonalTenant = currentTenant?.tenant?.is_personal;

  return (
    <div className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow duration-300">
      <div className="card-body">
        {/* Channel Header */}
        <div className="flex items-start justify-between">
          <h2 className="card-title text-lg font-bold flex items-center gap-2">
            {/* Default folder icon for all channels */}
            <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-5l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            <span className="truncate">{channel.name}</span>
          </h2>

          {/* Actions Dropdown - Only show for owners */}
          {canManageChannel && (
            <div className="dropdown dropdown-end">
              <div tabIndex={0} role="button" className="btn btn-ghost btn-sm btn-circle">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                        d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zM12 13a1 1 0 110-2 1 1 0 010 2zM12 20a1 1 0 110-2 1 1 0 010 2z" />
                </svg>
              </div>
              <ul tabIndex={0} className="dropdown-content menu bg-base-100 rounded-box z-[1] w-52 p-2 shadow">
                {/* Only show Manage Members for non-personal tenants */}
                {!isPersonalTenant && (
                  <li>
                    <a onClick={() => onManageMembers(channel)}>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                              d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a4 4 0 11-8 0 4 4 0 018 0z" />
                      </svg>
                      Manage Members
                    </a>
                  </li>
                )}
                <li>
                  <a onClick={() => onSettings(channel)}>
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Settings
                  </a>
                </li>
              </ul>
            </div>
          )}
        </div>

        {/* Channel Description */}
        {channel.description && (
          <p className="text-sm text-base-content/70 mt-2 line-clamp-2">
            {channel.description}
          </p>
        )}

        {/* Channel Stats */}
        <div className="flex items-center gap-4 text-sm text-base-content/60 mt-3">
          <span className="flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                    d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            0 videos
          </span>
          {/* Only show member count for organizational tenants and channel owners */}
          {!isPersonalTenant && userRole === 'owner' && (
            <span className="flex items-center gap-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                      d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              {channel.member_count || 0} members
            </span>
          )}
        </div>

        {/* User Role Badge */}
        <div className="flex items-center justify-between mt-4">
          <span className="badge badge-primary">{userRole}</span>
          
          {/* View Channel Button */}
          <button 
            onClick={handleViewChannel}
            className="btn btn-primary btn-sm"
          >
            View Channel
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChannelCard; 