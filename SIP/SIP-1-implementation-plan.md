# SIP-1 Implementation Plan
*Tenant and Channels - Incremental Development*

## ✅ Existing Features (Already Complete)
- Firebase authentication system
- User service with user table (`userservice_users`)
- Video service with video upload/recording/playback
- Comment system
- Frontend with React/TypeScript
- Microservices architecture (userservice, videoservice, commentservice)

## 🎯 Auto-Personal Tenant Approach
**Key Decision**: Every user automatically gets a personal tenant, so they can always:
- Upload videos and organize them in channels
- Create channels without needing to set up "company" tenants
- Optionally create additional organizational tenants later
- Have a consistent experience regardless of tenant type

---

## ✅ Phase 1: Extend UserService with Auto-Personal Tenants (Proto-First) - COMPLETED
**Goal**: Add tenant functionality with automatic personal tenant creation
**Duration**: 2-3 days ✅ **COMPLETED**

### ✅ Completed Tasks:
- [x] Updated `proto/userservice.proto` to add tenant management operations
- [x] Added tenant database migrations in `backend/userservice/db/migrations/2_add_tenants.up.sql`
- [x] Added tenant queries in `backend/userservice/db/scripts/queries.sql`
- [x] **User ran `go generate`** in userservice folder
- [x] Implemented tenant business logic in `backend/userservice/api/api.go`
- [x] **Added auto-creation of personal tenant on user registration/login**
- [x] Created migration for existing users: `3_create_personal_tenants_for_existing_users.up.sql`
- [x] Added comprehensive unit tests for tenant operations
- [x] **SECURITY ENHANCEMENT**: Added role-based authorization for tenant operations
- [x] **SECURITY ENHANCEMENT**: Added `GetUserRoleInTenant` query and API protection
- [x] **IMPROVEMENT**: Enhanced API to include role information in `GetUserTenants` response

**Database Changes** (in userservice):
```sql
-- Migration: 2_add_tenants.up.sql ✅ COMPLETED
CREATE TABLE userservice_tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_personal BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE for auto-created personal tenants
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES userservice_users(id)
);

CREATE TABLE userservice_tenant_users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES userservice_tenants(id),
    user_id TEXT NOT NULL REFERENCES userservice_users(id),
    role TEXT NOT NULL DEFAULT 'member', -- Simple string: 'super_admin', 'admin', 'member'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);

-- Migration: 3_create_personal_tenants_for_existing_users.up.sql ✅ COMPLETED
-- Auto-creates personal tenants for existing users
```

**Auto-Creation Logic**: ✅ **IMPLEMENTED**
- When user registers/first logs in → auto-create personal tenant named "{username}'s Workspace"
- User becomes super_admin of their personal tenant
- Personal tenants have `is_personal = TRUE`
- **ENHANCEMENT**: Existing users get personal tenants via database migration

**✅ Test Criteria PASSED**: 
- Every user automatically gets a personal tenant
- Users can create additional organizational tenants
- Personal tenant creation is seamless and invisible to user
- **SECURITY**: Only super_admin can view/manage tenant members
- **SECURITY**: Only super_admin can add users to tenants

---

## ✅ Phase 2: Frontend Tenant Management - COMPLETED
**Goal**: Add tenant creation UI to existing frontend
**Duration**: 1-2 days ✅ **COMPLETED**

### ✅ Completed Tasks:
- [x] Updated `frontend/webapp/src/proto/userservice.ts` with new tenant operations (auto-generated)
- [x] Created tenant creation modal in React (`CreateWorkspaceModal`)
- [x] Added tenant management to existing user store (`stores/tenants.ts`)
- [x] Updated TeamPage with comprehensive tenant management UI
- [x] Added tenant switching in header (`TenantSwitcher` component)
- [x] Connected to userservice tenant endpoints with proper GRPC interceptors
- [x] Added comprehensive error handling and loading states
- [x] **UI ENHANCEMENT**: Added React Feather icons for modern UI
- [x] **UX ENHANCEMENT**: Added modal organization with separate components
- [x] **UX ENHANCEMENT**: Added in-modal error handling and loading states
- [x] **SECURITY ENHANCEMENT**: Added role-based UI rendering (only super_admin sees member management)
- [x] **IMPROVEMENT**: Enhanced protobuf to include `TenantWithRole` for proper role information

**✅ Test Criteria PASSED**: 
- Users can create and manage tenants through the web interface
- **SECURITY**: Role-based access control in UI (members management only for super_admin)
- **UX**: Clean modal organization and error handling
- **UX**: Modern UI with proper iconography and responsive design

---

## ✅ Phase 3: Update Video Service for Tenant-Based Privacy - COMPLETED
**Goal**: Enhance existing video system with tenant-based privacy controls
**Duration**: 1-2 days ✅ **COMPLETED**

### ✅ Completed Tasks:
- [x] Added `is_private` field to videos table (defaults to TRUE)
- [x] Added `tenant_id` field to videos table for tenant association
- [x] Updated protobuf to include `tenant_id` in Video message and requests
- [x] Created database migration `3_add_privacy_and_tenant.up.sql` with proper indexing
- [x] **MIGRATION COMPLETED**: Updated existing videos to be associated with users' personal tenants
- [x] Updated video queries to be tenant-aware and respect privacy
- [x] Modified video API to handle tenant-based video access
- [x] **SECURITY ENHANCEMENT**: Videos are private by default and only accessible by owner
- [x] Updated frontend video store to support tenant-based video management
- [x] Enhanced upload page with tenant selection
- [x] Updated video listing to show tenant-specific videos
- [x] **UX ENHANCEMENT**: Added workspace context to video cards and improved UI
- [x] **AUTHENTICATION**: Fixed video streaming with cookie-based auth and tenant validation
- [x] **ARCHITECTURE**: Implemented proper microservice communication (videoservice ↔ userservice)

**Database Changes**: ✅ **COMPLETED**
```sql
-- Migration: 3_add_privacy_and_tenant.up.sql ✅ COMPLETED
ALTER TABLE videos ADD COLUMN is_private BOOLEAN DEFAULT TRUE;
ALTER TABLE videos ADD COLUMN tenant_id TEXT;

-- Indexes for performance
CREATE INDEX idx_videos_tenant_id ON videos(tenant_id);
CREATE INDEX idx_videos_tenant_user ON videos(tenant_id, uploaded_user_id);

-- ✅ MIGRATION COMPLETED: Migrate existing videos to users' personal tenants
UPDATE videos SET tenant_id = (
    SELECT t.id FROM userservice_tenants t 
    JOIN userservice_tenant_users tu ON t.id = tu.tenant_id 
    WHERE t.is_personal = TRUE AND tu.user_id = videos.uploaded_user_id
    AND tu.role = 'super_admin' LIMIT 1
) WHERE tenant_id IS NULL;
```

**✅ Test Criteria PASSED**: 
- All videos are private by default and uploaded to current tenant
- Users only see their own videos within selected tenant
- **SECURITY**: Cross-tenant video access is prevented
- **MIGRATION**: Existing videos properly migrated to personal tenants ✅ **CONFIRMED**
- **UX**: Tenant-aware upload and video management interface
- **STREAMING**: Video playback works with proper authentication and tenant validation
- **ARCHITECTURE**: Clean microservice communication without circular dependencies

---

## ✅ Phase 4: Channel Foundation (Security-First) - MOSTLY COMPLETED
**Goal**: Add basic channel creation and organization within tenants
**Duration**: 2-3 days ✅ **MOSTLY COMPLETED**

### ✅ Completed Tasks:
- [x] Add ChannelService to `videoservice.proto` with tenant-scoped operations
- [x] Create channel database schema in videoservice database (tenant-scoped)
- [x] **Database Migration**: Added `4_add_channels.up.sql` and `5_add_channel_support.up.sql`
- [x] **API Implementation**: Implemented channel CRUD operations with tenant validation
- [x] **Permission System**: Added channel-user permissions (owner, uploader, viewer)
- [x] **Frontend Implementation**: Added channel dashboard, creation, and management UI
- [x] **Member Management**: Added channel member management (owner can add/remove members)
- [x] **Security**: All channel operations are tenant-scoped and permission-checked

### 🔄 Remaining Tasks:
- [ ] **TENANT-LEVEL VIDEOS DISPLAY**: Add section in channel dashboard for videos not assigned to any channel
- [ ] **UPLOAD DESTINATION CHOICE**: Update upload/record UI to allow choosing between tenant and specific channel
- [ ] **CHANNEL FILTERING**: Extend VideoService ListVideos to support optional channel filtering
- [ ] **CHANNEL ACCESS VALIDATION**: Add channel access validation in VideoService operations

### Key Principles (Security-First):
1. **Videos without channels**: Users can upload videos directly to tenants (current behavior continues)
2. **Channels for organization**: Channels are optional organizational tools within tenants
3. **Private by membership**: Channels are private until owner adds members (no separate is_private field needed)
4. **Simple permissions**: Three roles - `owner` (channel creator), `uploader` (can add videos), `viewer` (can view)
5. **Channel ownership model**: Once video is in channel, channel controls it (no cross-channel movement)
6. **Tenant-scoped security**: All channel operations require tenant validation to prevent cross-tenant data leaks

**Database Schema** (in videoservice database): ✅ **COMPLETED**
```sql
-- Migration: 4_add_channels.up.sql ✅ COMPLETED
CREATE TABLE channels (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL, -- References userservice_tenants(id) but no FK constraint
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL, -- References userservice_users(id) but no FK constraint
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Channel permissions: owner, uploader, viewer
CREATE TABLE channel_members (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL, -- References userservice_users(id) but no FK constraint
    role TEXT NOT NULL DEFAULT 'viewer', -- 'owner', 'uploader', 'viewer'
    added_by TEXT NOT NULL, -- References userservice_users(id) but no FK constraint
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, user_id)
);

-- Migration: 5_add_channel_support.up.sql ✅ COMPLETED
-- Added channel_id field to videos table with foreign key constraint
ALTER TABLE videos ADD COLUMN channel_id TEXT;
CREATE INDEX idx_videos_channel ON videos(channel_id);
CREATE INDEX idx_videos_tenant_channel ON videos(tenant_id, channel_id);
```

**✅ Test Criteria PASSED**: 
- Users can create private channels in any tenant they belong to
- Channel creators become owners automatically
- Only channel members can see the channel
- **SECURITY**: No cross-tenant channel access possible
- **SECURITY**: All channel queries are tenant-scoped
- **Personal tenants**: Channel creation works for personal organization
- **Organizational tenants**: Owners can invite tenant members to channels

---

## 🔄 Phase 5: Video-Channel Association (Ultra-Simplified) - NEXT PHASE
**Goal**: Allow one-way organization of videos into channels
**Duration**: 1-2 days

### Key Principles (Ultra-Simplified):
1. **Optional organization**: Videos can exist without channels (current behavior)
2. **Channel ownership**: Once video moves to channel, channel controls it
3. **One-way movement**: Tenant videos can move to channels, but no cross-channel movement
4. **Clear ownership**: Channel owners control all videos in their channel

### Upload & Movement Logic:
- **Tenant upload** (current): User uploads to tenant, owns video, can later move to channel
- **Channel upload** (new): User uploads directly to channel, channel owns immediately
- **One-time move**: Tenant videos can be moved to channel (user loses control)
- **Channel removal**: Only channel owner can remove videos (back to tenant-level)

### Tasks:
- [x] Add `channel_id` field to videos table (nullable - videos can exist without channels) ✅ **COMPLETED**
- [ ] **UPLOAD DESTINATION UI**: Update frontend upload UI to choose destination (tenant vs channel)
- [ ] **CHANNEL UPLOAD FUNCTIONALITY**: Add "upload to channel" functionality with channel access validation
- [ ] **MOVE VIDEO TO CHANNEL**: Implement "move video to channel" endpoint with tenant+channel validation
- [ ] **CHANNEL VIDEO FILTERING**: Extend existing ListVideos API to support channel filtering (tenant+channel scoped)
- [ ] **REMOVE FROM CHANNEL**: Add "remove from channel" functionality (channel owner only, moves back to tenant)
- [ ] **TENANT-LEVEL VIDEOS SECTION**: Update channel dashboard to show videos not assigned to any channel
- [ ] **CHANNEL VIDEOS DISPLAY**: Update frontend to show proper channel organization
- [ ] Write tests for simplified video-channel operations with security validation

**Security-First Video Operations**:
```sql
-- ✅ SAFE: Get videos in specific channel within tenant
-- name: GetVideosByTenantIDAndChannelID :many
SELECT * FROM videos 
WHERE tenant_id = @tenant_id AND channel_id = @channel_id
ORDER BY created_at DESC;

-- ✅ SAFE: Get tenant-level videos (not assigned to any channel)
-- name: GetTenantVideosByTenantIDWithoutChannel :many
SELECT * FROM videos 
WHERE tenant_id = @tenant_id AND (channel_id IS NULL OR channel_id = '')
ORDER BY created_at DESC;

-- ✅ SAFE: Current query remains unchanged
-- name: GetVideosByTenantID :many  
SELECT * FROM videos 
WHERE tenant_id = @tenant_id 
ORDER BY created_at DESC;

-- ❌ NEVER: Query by channel_id alone (prevents cross-tenant leaks)
```

**Simplified Permission Logic**:
- **Tenant-level videos**: User owns, can move to any channel they have `uploader` access to (one-time)
- **Channel videos**: Channel owns, only channel `owner` can remove back to tenant
- **Upload destination**: Users choose tenant or channel at upload time (with channel access validation)
- **No cross-channel movement**: Videos stay where they end up
- **Security**: All video operations validate tenant_id first, then channel_id within that tenant

**Test Criteria**: 
- Users can upload videos to tenant (current behavior) or directly to channels
- Tenant videos can be moved to channels (one-time, user loses control)
- Channel owners can remove videos from their channels (back to tenant-level)
- **TENANT-LEVEL VIDEOS DISPLAY**: Channel dashboard shows separate section for tenant videos not in any channel
- **No complexity**: No cross-channel movement, clear ownership model

---

## Phase 6: Basic OPA Integration
**Goal**: Replace hardcoded permissions with OPA policies (including role definitions)
**Duration**: 3-4 days

**NOTE**: We've implemented basic RBAC with hardcoded role checks. OPA can enhance this but is not critical for basic functionality.

### Tasks:
- [ ] Set up OPA as embedded Go library
- [ ] Write Rego policies defining what each role can do:
  - `super_admin`: Full tenant control, user management, all operations
  - `admin`: Channel management, user invites, content moderation
  - `member`: Basic video upload, comment, view permissions
- [ ] Replace hardcoded permission checks with OPA calls
- [ ] Add policy testing framework
- [ ] Migrate existing permissions one by one

**Example Rego Policy**:
```rego
package tenant_permissions

# Super admins can do anything in their tenant
allow {
    input.user_role == "super_admin"
    input.tenant_id == input.user_tenant_id
}

# Admins can manage channels and invite users
allow {
    input.user_role == "admin"
    input.action in ["create_channel", "invite_user", "manage_content"]
}

# Members can upload videos and comment
allow {
    input.user_role == "member" 
    input.action in ["upload_video", "comment", "view_content"]
}
```

**Test Criteria**: All role-based permissions work through OPA policies, no hardcoded role logic

---

## Phase 7: Tenant User Management
**Goal**: Super admins can invite/manage users
**Duration**: 2-3 days

### Tasks:
- [ ] Add user invitation system
- [ ] Implement user role management within tenants
- [ ] Add user removal from tenants
- [ ] Update OPA policies for user management
- [ ] Write tests for user management operations

**Test Criteria**: Super admins can invite/remove users from their tenants

---

## Phase 8: Advanced Channel Permissions
**Goal**: Refined channel access controls
**Duration**: 2-3 days

### Tasks:
- [ ] Add granular channel permissions (view, upload, admin)
- [ ] Implement channel member management
- [ ] Update OPA policies for channel permissions
- [ ] Add channel permission inheritance
- [ ] Write comprehensive permission tests

**Test Criteria**: Different users have different access levels to channels

---

## Phase 9: Video Access Controls
**Goal**: Implement video viewing permissions
**Duration**: 2-3 days

### Tasks:
- [ ] Add video viewing permissions through channels
- [ ] Implement public/private video access
- [ ] Add video sharing capabilities
- [ ] Update OPA policies for video access
- [ ] Write tests for video access scenarios

**Test Criteria**: Videos are accessible based on channel membership and permissions

---

## Phase 10: Polish & Optimization
**Goal**: Clean up and optimize
**Duration**: 2-3 days

### Tasks:
- [ ] Add comprehensive error handling
- [ ] Optimize database queries
- [ ] Add proper logging and monitoring
- [ ] Performance testing
- [ ] Security audit of OPA policies
- [ ] Documentation updates

**Test Criteria**: System is production-ready with good performance

---

## **🚨 CRITICAL: Frontend Architecture Pattern**

### **DO NOT use direct REST API calls in frontend components!**

The frontend uses **GRPC services with TypeScript clients**, not REST APIs.

**❌ WRONG:**
```javascript
// Never do this!
const response = await fetch('/api/videoservice/channels', {
  headers: { 'authorization': token }
});
```

**✅ CORRECT:**
```javascript
// Always use stores and GRPC services!
import { fetchChannels } from '../stores/channels';
await fetchChannels();
```

### **Frontend Architecture:**
1. **Stores** (`src/stores/`) - Manage state and contain GRPC calls
2. **Components** - Use stores via `useStore()` hook from nanostores
3. **GRPC Services** - Generated TypeScript clients in `src/proto/`

### **Pattern to Follow:**
- **Videos**: `src/stores/videos.ts` → `VideoServiceClient`
- **Tenants**: `src/stores/tenants.ts` → `UserServiceClient` + `TenantServiceClient`  
- **Channels**: `src/stores/channels.ts` → `ChannelServiceClient`

### **Example Implementation:**
```typescript
// Store pattern (stores/channels.ts)
export const channelService = new ChannelServiceClient(apiUrl, {}, {
  unaryInterceptors: [authInterceptor]
});

export const fetchChannels = async () => {
  const request = new GetChannelsRequest();
  const response = await channelService.GetChannels(request, {});
  $channels.set(response.channels);
};

// Component pattern
const Component = () => {
  const channels = useStore($channels);
  
  useEffect(() => {
    fetchChannels(); // Call store function, not direct API
  }, []);
};
```

### **Headers/Auth:**
- ✅ **Interceptors** handle auth automatically in GRPC services
- ❌ **Manual headers** should never be needed in components

**Always check existing patterns in channel dashboard and `stores/videos.ts` before implementing new features!**

### **🗃️ Database Queries (Simplified):**
- **`GetVideosByTenantIDAndChannelID`**: Get videos for a specific channel
- **`GetAllAccessibleVideosByTenantID`**: Get all videos user can access (private + channel videos)
  - Includes user's private videos (no channel_id)
  - Includes videos in channels where user is a member
- **Frontend filtering**: Filter `GetAllAccessibleVideosByTenantID` results for "My Videos" section

### **🧹 Code Cleanup Completed:**
- ✅ Removed `VideosPage.jsx` - Consolidated into channel dashboard
- ✅ Removed `ListOfVideos.jsx` - Replaced with dashboard video sections
- ✅ Removed `Library.jsx` - Unused component
- ✅ Updated navigation after upload to go to `/channels` instead of `/videos`
- ✅ Simplified video fetching logic with only 2 database queries

**Always check existing patterns in channel dashboard and `stores/videos.ts` before implementing new features!**

## 🎉 Current Status Summary:
**✅ COMPLETED PHASES**: 1, 2, 3, 4 (mostly complete)
**🔄 CURRENT PHASE**: 4 (completing remaining tasks) → 5 (Video-Channel Association)
**📊 PROGRESS**: 4/10 phases complete (40%)

### ✅ Phase 4 Recent Achievements:
- **Complete channel management system**: Create, update, list channels with proper tenant scoping
- **Channel membership system**: Owner/uploader/viewer roles with proper permissions
- **Frontend channel dashboard**: Full UI for channel management with role-based access
- **Member management**: Add/remove channel members (non-personal tenants only)
- **Database schema**: Proper channel and channel_members tables with foreign key constraints

### 🔄 Phase 4 Remaining Tasks:
1. **MY VIDEOS DISPLAY**: Add section in channel dashboard for user's private videos not assigned to any channel ✅ **COMPLETED**
2. **UPLOAD TO CHANNEL**: Add option during video upload/record to select target channel
3. **MOVE VIDEOS TO CHANNELS**: Add functionality to move existing videos into channels

### **🔐 Important Security Model:**
- **Private by Default**: When users upload videos to a tenant, they are **private to that user only**
- **Channel Sharing**: Videos only become visible to other tenant members when moved to a channel
- **No Cross-User Access**: Tenant-level videos are never shared between users automatically
- **Explicit Sharing**: Users must explicitly move videos to channels to share them

### **📋 Video Privacy Levels:**
1. **User Private**: Videos uploaded but not in any channel (visible only to uploader)
2. **Channel Shared**: Videos moved to channels (visible to channel members based on channel permissions)
3. **Tenant Shared**: Only through channel membership, never directly at tenant level

### **🗃️ Database Queries (Simplified):**
- **`GetVideosByTenantIDAndChannelID`**: Get videos for a specific channel
- **`GetAllAccessibleVideosByTenantID`**: Get all videos user can access (private + channel videos)
  - Includes user's private videos (no channel_id)
  - Includes videos in channels where user is a member
- **Frontend filtering**: Filter `GetAllAccessibleVideosByTenantID` results for "My Videos" section

### **🧹 Code Cleanup Completed:**
- ✅ Removed `VideosPage.jsx` - Consolidated into channel dashboard
- ✅ Removed `ListOfVideos.jsx` - Replaced with dashboard video sections
- ✅ Removed `Library.jsx` - Unused component
- ✅ Updated navigation after upload to go to `/channels` instead of `/videos`
- ✅ Simplified video fetching logic with only 2 database queries

**Always check existing patterns in channel dashboard and `stores/videos.ts` before implementing new features!**

### 🎯 Next Priority (Phase 5):
- **Video-Channel Association**: Allow users to upload directly to channels or move existing videos to channels
- **Upload Destination UI**: Provide channel selection during upload/recording
- **Channel Video Organization**: Proper display of channel vs tenant-level videos

## 🔒 Security Enhancements Added:
- Role-based authorization for all tenant operations
- API-level permission checks (only super_admin can manage members)
- Frontend role-based UI rendering
- Comprehensive error handling for permission denied scenarios
- Proper role information in API responses
- **VIDEO PRIVACY**: All videos private by default with tenant-based access control
- **CROSS-TENANT SECURITY**: Videos isolated by tenant boundaries
- **CHANNEL SECURITY**: All channel operations are tenant-scoped with proper permission checks

## 🎨 UX Improvements Added:
- Modern React Feather icons throughout the UI
- Clean modal organization with separate components
- In-modal error handling and loading states
- Responsive design with proper mobile support
- Comprehensive error dismissal functionality
- **TENANT-AWARE VIDEO MANAGEMENT**: Workspace selection for uploads and viewing
- **ENHANCED VIDEO CARDS**: Privacy badges, workspace context, and improved layout
- **CHANNEL MANAGEMENT UI**: Full channel dashboard with creation, settings, and member management

## ✅ Migration Status:
- **Personal Tenants**: All existing users have personal tenants ✅ **CONFIRMED**
- **Video Migration**: All existing videos moved to users' personal tenants ✅ **CONFIRMED**
- **Channel Support**: Videos table updated with channel_id field ✅ **CONFIRMED**

## Testing Strategy for Each Phase:
1. **Unit Tests**: Test individual functions/methods
2. **Integration Tests**: Test API endpoints end-to-end
3. **Manual Testing**: Use Postman/curl to verify functionality
4. **Regression Testing**: Ensure previous phases still work

## Rollback Strategy:
- Each phase should be in a separate branch
- Merge to main only after thorough testing
- Keep database migrations reversible
- Document any breaking changes

## Success Metrics:
- All tests pass for current and previous phases
- No performance degradation
- Clean, maintainable code
- Proper error handling and logging

**Total Estimated Duration**: 20-30 days (4-6 weeks) 
**Completed**: ~12 days (Phases 1, 2, 3, 4 mostly complete)
**Remaining**: ~8-18 days 