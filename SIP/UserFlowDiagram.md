# User Flow Diagram

## Comprehensive User Flow - Current Implementation (Phases 1-3)

## Key User Journeys

### 1. **New User Journey:**
- Login → Auto-create personal tenant → Upload/record videos → Videos saved as private → Manage in personal workspace

### 2. **Workspace Creation Journey:**
- Any authenticated user can create new organizational workspaces
- Path: Team Page → Create New Workspace → Enter details → Become super admin of new workspace

### 3. **Team Collaboration Journey:**
- Super admin invites users → Users join organizational tenant → Upload videos → Videos shared in workspace → Role-based access control

### 4. **Video Privacy Logic:**
- **Personal Workspace**: Videos always private to user
- **Own Organizational Workspace**: Videos private to user (creator is super admin)
- **Other's Organizational Workspace**: Videos shared within workspace (accessible to all members)

### 5. **Video Management Journey:**
- Select workspace → Upload/record → Privacy determined by workspace type → View tenant-specific videos → Play with comments

## Legend:
- **Blue (Solid)**: Currently implemented features (Phases 1-3)
- **Purple (Dashed)**: Future planned features (Phases 4-10)
- **Green**: Authentication flows
- **Orange**: Tenant/workspace management flows
- **Pink**: Workspace creation flows (available to all users)
- **Yellow**: Personal workspace features
- **Light Green**: Shared workspace features

```mermaid
graph TD
    A[User Opens App] --> B{Authenticated?}
    B -->|No| C[Login Page]
    C --> D[Firebase Auth]
    D --> E[User Dashboard]
    B -->|Yes| E
    
    E --> F[Header with Tenant Switcher]
    F --> G{Current Tenant?}
    G -->|No Personal Tenant| H[Auto-Create Personal Tenant]
    H --> I[Set as Current Tenant]
    G -->|Has Tenant| I
    I --> J[Main Navigation]
    
    J --> K[Videos Page]
    J --> L[Upload Page]
    J --> M[Record Page]
    J --> N[Team Page]
    J --> O[Library Page]
    J --> P[Profile Page]
    J --> Q[Settings Page]
    
    %% Videos Flow
    K --> K1[Show Current Tenant Videos]
    K1 --> K2[Video Cards with Privacy Badges]
    K2 --> K3[Click Video]
    K3 --> K4[Video Player Page]
    K4 --> K5[Comments Section]
    K5 --> K6[Add/View Comments]
    
    %% Upload Flow
    L --> L1[Select Workspace]
    L1 --> L2[Choose Video File]
    L2 --> L3[Enter Title/Description]
    L3 --> L4[Upload to Selected Tenant]
    L4 --> L5{Workspace Type & Ownership?}
    L5 -->|Personal Workspace| L6[Video Saved as Private]
    L5 -->|Own Organizational Workspace| L7[Video Saved as Private]
    L5 -->|Other's Organizational Workspace| L8[Video Saved as Shared in Workspace]
    L6 --> K1
    L7 --> K1
    L8 --> K1
    
    %% Record Flow
    M --> M1[Screen Recording]
    M1 --> M2[Save Recording]
    M2 --> M3[Auto-Upload to Current Tenant]
    M3 --> L5
    
    %% Team Management Flow
    N --> N1{Current Tenant Type?}
    N1 -->|Personal Tenant| N9[View Personal Workspace Info]
    N1 -->|Organizational Tenant| N10{User Role?}
    N10 -->|Super Admin| N2[View Team Members]
    N10 -->|Member| N3[View Workspace Info Only]
    N2 --> N4[Add Users to Current Tenant]
    N2 --> N5[Create New Workspace Button]
    N3 --> N5
    N9 --> N5
    N5 --> N6[Enter Workspace Name & Description]
    N6 --> N7[New Organizational Tenant Created]
    N7 --> N8[User becomes Super Admin of New Tenant]
    N8 --> TS3
    
    %% Future Phases - Channels
    K1 -.->|Phase 4| C1[Channel Organization]
    C1 -.-> C2[Create Channels]
    C2 -.-> C3[Organize Videos in Channels]
    C3 -.-> C4[Channel Permissions]
    
    %% Future Phases - Sharing
    K4 -.->|Phase 5| S1[Share Video]
    S1 -.-> S2[Generate Share Link]
    S2 -.-> S3[Set Sharing Permissions]
    
    %% Tenant Switching Flow
    F --> TS1[Tenant Switcher Dropdown]
    TS1 --> TS2[Select Different Tenant]
    TS2 --> TS3[Switch Context]
    TS3 --> TS4[Reload Tenant-Specific Data]
    TS4 --> I
    
    %% Error Handling
    D -->|Auth Fails| D1[Show Error Message]
    L4 -->|Upload Fails| L9[Show Upload Error]
    N4 -->|Add User Fails| N11[Show Permission Error]
    N7 -->|Creation Fails| N12[Show Creation Error]
    
    %% Styling
    classDef implemented fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef future fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,stroke-dasharray: 5 5
    classDef auth fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef tenant fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef workspace fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    classDef personal fill:#fff8e1,stroke:#f57f17,stroke-width:2px
    classDef shared fill:#e8f5e8,stroke:#388e3c,stroke-width:2px
    
    class A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,K1,K2,K3,K4,K5,K6,L1,L2,L3,L4,L5,L6,L7,L8,L9,M1,M2,M3,N1,N2,N3,N4,N5,N6,N7,N8,N9,N10,N11,N12,TS1,TS2,TS3,TS4,D1 implemented
    class C1,C2,C3,C4,S1,S2,S3 future
    class C,D,D1 auth
    class F,G,H,I,TS1,TS2,TS3,TS4,N,N1,N2,N3,N4,N10 tenant
    class N5,N6,N7,N8,N11,N12 workspace
    class N9,L6,L7 personal
    class L8 shared
```