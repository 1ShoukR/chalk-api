# Chalk Mobile Frontend PRD

## 1) Document Metadata

- Product: `chalk-mobile`
- Platform: React Native (Expo), TypeScript
- Domain: usechalk.app
- Primary target: iOS
- Secondary target: Android
- Status: Finalized v1 (implementation-ready)
- Team: Solo developer
- Last updated: 2026-02-12
- Backend contract: `docs/openapi.json`

---

## 2) Product Vision

Build a clean, premium-feeling mobile app for coaches and clients to manage training, messaging, sessions, and subscriptions with minimal friction. A single app serving two roles — coach and client — in one experience.

The frontend must feel fast, consistent, and trustworthy while integrating tightly with the existing backend API contract in `docs/openapi.json`.

### Key Differentiators

- Single app, two roles (coach + client in one app)
- Full workout programming on mobile (competitors lack this)
- Built-in scheduling (no third-party calendar dependency)
- $29/mo unlimited vs competitor pricing at $120/mo for 50 clients
- Future: built-in nutrition tracking and progress photos (see Section 24)

### Target Users

1. **Coaches** — Personal trainers managing 5-100+ clients
2. **Clients** — People working with a personal trainer

---

## 3) Goals

- Ship an iOS-first MVP that fully supports the current backend slices:
  - Auth + user profile
  - Coach profile + invite flow
  - Workout templates, assignment visibility, and client execution
  - Messaging
  - Session booking and management
  - Subscription status and feature gating
- Use TypeScript with strict typing and generated API types from OpenAPI.
- Use React Context for global cross-cutting app state only.
- Use FeatureFlags for future feature rollout (nutrition, progress, OAuth, etc.).
- Enable React Compiler and code patterns that maximize compiler benefits.
- Establish a strong design system so UI remains premium as the app grows.

---

## 4) Non-Goals (MVP)

- Building web support as first-class target
- Deep analytics and growth tooling beyond basic event instrumentation
- Complex offline sync conflict resolution
- Full notification center UX (beyond practical MVP push handling)
- Nutrition tracking (backend tables exist but no API endpoints — see Section 24)
- Progress photos and body measurement tracking (deferred — see Section 24)
- OAuth sign-in flows (Google/Facebook/Apple — backend schema exists but flows not finalized)
- Photo messaging (standard text messaging only for MVP)
- Dark mode (architecture supports it, but MVP ships light mode only)

---

## 5) Users and Core Jobs-to-be-Done

### Coach

- Set up profile and credentials
- Invite clients via shareable code/link
- Create workout templates and assign to clients
- Monitor client workout completion and logged sets
- Manage schedule/availability and session types
- Book, complete, cancel, and no-show sessions
- Chat with clients quickly

### Client

- Join a coach from invite link or code
- See assigned workouts and log exercise progress
- Start and complete workouts with set-level logging
- Book sessions in available coach slots
- Message coach
- Understand subscription/feature access state

---

## 6) Technical Requirements

### Framework and Runtime

- Expo: use latest stable SDK at project kickoff
- React Native: version paired with the selected Expo SDK
- TypeScript: strict mode enabled (`strict: true`, `noUncheckedIndexedAccess: true`)
- Version policy for MVP: lock framework versions after initialization and only apply patch-level upgrades during MVP delivery

### React Compiler and Rendering Strategy

- Enable React Compiler in app build setup.
- Prefer compiler-friendly component patterns:
  - pure render functions
  - stable props
  - avoid unnecessary manual memoization unless measured bottleneck
- Use Context for global state domains, not as a replacement for server-state caching.

### State Management

**React Context (required) for cross-cutting app state only:**

- `AuthContext` (tokens, user identity, auth actions)
- `ThemeContext` (theme mode, design tokens)
- `AppConfigContext` (feature flags, env config, app settings)

**TanStack Query (required) for all server state:**

- All API data fetching, caching, retry, and background refetching
- This includes data that's accessed across multiple screens (client lists, workout lists, etc.) — TanStack Query's cache already handles cross-screen sharing
- Invalidation-based refetch on mutations

**Component-level state for ephemeral UI:**

- Form inputs, toggles, modals, animation state
- `WorkoutEditorContext` is the one exception: a local Context scoped to the workout creation/edit flow for complex multi-step form state (mounted only during that flow, not global)

**Why this separation matters:** Putting server-fetched data into React Context creates a parallel cache alongside TanStack Query. This causes stale data bugs (Context doesn't know when to refetch), re-render churn (every Context consumer re-renders on any state change), and dual-responsibility confusion. TanStack Query already solves caching, background refresh, and cross-screen data sharing — let it own that job.

### Navigation

- Expo Router for route organization and deep linking.
- Route groups:
  - `(auth)` — login, register, forgot-password
  - `(onboarding)` — post-auth role selection and profile setup
  - `(coach)` — coach tab navigator and stacks
  - `(client)` — client tab navigator and stacks
  - `(shared)` — settings, profile, subscription, paywall
  - modal stack for overlays (edit profile, filters, confirmations)

### Deep Linking

```
chalk://invite/{code}     → Accept coach invite
chalk://workout/{id}      → Open workout detail
chalk://session/{id}      → Open session detail
chalk://chat/{coachId}    → Open chat with coach
```

### API Integration

- Contract source: `docs/openapi.json`
- Generate typed API models/hooks at build-time (e.g., openapi-typescript + openapi-fetch or similar codegen)
- Keep network layer thin and domain-organized:
  - `src/api/auth`
  - `src/api/users`
  - `src/api/coaches`
  - `src/api/workouts`
  - `src/api/messages`
  - `src/api/sessions`
  - `src/api/subscriptions`

### HTTP Client and Interceptors

```typescript
// src/lib/api.ts — Axios instance with auth interceptors

// Request interceptor: attach Bearer token from SecureStore
// Response interceptor: on 401, attempt token refresh via /auth/refresh
//   - If refresh succeeds: retry original request with new token
//   - If refresh fails: clear tokens, trigger AuthContext logout flow
// Timeout: 10s default
```

---

## 7) Tech Stack

| Category | Technology | Version | Rationale |
|----------|------------|---------|-----------|
| Framework | React Native | 0.73+ | Paired with Expo SDK |
| Platform | Expo | SDK 50+ | Fastest path to iOS+Android with managed workflow |
| Navigation | Expo Router | v3 | File-based routing, deep linking support |
| Language | TypeScript | 5.0+ | Strict mode for type safety |
| UI Library | Tamagui | Latest | Token-first design system, static extraction, premium feel |
| State (global) | React Context | Built-in | Cross-cutting app state only |
| Data Fetching | TanStack Query | v5 | Server state caching, retry, background refetch |
| Forms | React Hook Form | v7 | Performant form state, minimal re-renders |
| Validation | Zod | v3 | Schema validation, pairs with RHF |
| HTTP Client | Axios | v1 | Interceptors for auth token management |
| Storage | expo-secure-store | Latest | Secure token persistence |
| Notifications | expo-notifications | Latest | Push notification handling |
| Payments | react-native-purchases (RevenueCat) | Latest | Subscription management |
| Fonts | Inter + Sora via @expo-google-fonts | Latest | Body/UI + headings |
| Icons | lucide-react-native | Latest | Primary icon set, clean stroke style |
| Icons (fallback) | @expo/vector-icons | Latest | Broad coverage for niche icons |
| SVG | react-native-svg + transformer | Latest | Custom illustrations, empty states |

### Dev Dependencies

| Tool | Purpose |
|------|---------|
| ESLint | Linting |
| Prettier | Formatting |
| Jest | Unit testing |
| Maestro | E2E testing (post-MVP) |
| openapi-typescript | API type generation from OpenAPI |

---

## 8) Project Structure (Locked)

```
src/
├── app/                          # Expo Router routes and layout shells
│   ├── _layout.tsx               # Root layout (providers)
│   ├── index.tsx                 # Entry redirect
│   │
│   ├── (auth)/                   # Unauthenticated flows
│   │   ├── _layout.tsx
│   │   ├── welcome.tsx
│   │   ├── sign-in.tsx
│   │   ├── sign-up.tsx
│   │   └── forgot-password.tsx
│   │
│   ├── (onboarding)/             # Post-auth setup
│   │   ├── _layout.tsx
│   │   ├── role-select.tsx
│   │   ├── coach/
│   │   │   ├── step-1.tsx        # Profile info
│   │   │   ├── step-2.tsx        # Business info
│   │   │   └── step-3.tsx        # Availability + notifications
│   │   └── client/
│   │       ├── step-1.tsx        # Profile info
│   │       └── step-2.tsx        # Connect to coach
│   │
│   ├── (coach)/                  # Coach experience
│   │   ├── _layout.tsx           # Tab navigator
│   │   ├── index.tsx             # Dashboard
│   │   ├── clients/
│   │   │   ├── index.tsx         # Client list
│   │   │   ├── add.tsx           # Add/invite client
│   │   │   └── [id]/
│   │   │       ├── index.tsx     # Client detail
│   │   │       └── workouts.tsx  # Client's workouts
│   │   ├── workout/
│   │   │   ├── create.tsx        # Create workout
│   │   │   └── [id].tsx          # Edit workout
│   │   ├── templates/
│   │   │   ├── index.tsx         # Template list
│   │   │   └── [id].tsx          # Template detail
│   │   ├── schedule/
│   │   │   ├── index.tsx         # Calendar view
│   │   │   ├── availability.tsx  # Set availability
│   │   │   ├── session-types.tsx # Manage session types
│   │   │   └── [id].tsx          # Session detail
│   │   └── messages/
│   │       ├── index.tsx         # Conversation list
│   │       └── [id].tsx          # Chat screen
│   │
│   ├── (client)/                 # Client experience
│   │   ├── _layout.tsx           # Tab navigator
│   │   ├── index.tsx             # Today's workout
│   │   ├── calendar.tsx          # Workout calendar
│   │   ├── workout/
│   │   │   └── [id].tsx          # Workout detail + logging
│   │   ├── schedule/
│   │   │   ├── index.tsx         # My sessions
│   │   │   └── book.tsx          # Book a session
│   │   └── messages.tsx          # Chat with coach
│   │
│   └── (shared)/                 # Both roles
│       ├── settings.tsx
│       ├── profile.tsx
│       ├── subscription.tsx
│       └── paywall.tsx
│
├── features/                     # Domain modules
│   ├── auth/                     # Auth feature components, hooks, utils
│   ├── coaches/                  # Coach-specific features
│   ├── workouts/                 # Workout programming + execution
│   ├── messages/                 # Messaging feature
│   ├── sessions/                 # Scheduling + booking
│   └── subscriptions/            # Subscription + paywall
│
├── components/                   # Shared UI components
│   ├── primitives/               # Box, Text, Stack, Icon (Tamagui-based)
│   ├── composites/               # Card, Input, Button, Avatar, Badge, ListItem
│   └── domain/                   # WorkoutCard, SessionSlotRow, ConversationListItem
│
├── theme/                        # Design tokens + Tamagui config
│   ├── tokens.ts                 # Colors, spacing, radii, shadows
│   ├── typography.ts             # Type scale + font config
│   └── tamagui.config.ts         # Tamagui theme configuration
│
├── contexts/                     # Global app contexts only
│   ├── AuthContext.tsx
│   ├── ThemeContext.tsx
│   └── AppConfigContext.tsx
│
├── api/                          # Generated types + endpoint wrappers
│   ├── generated/                # Auto-generated from openapi.json
│   ├── auth.ts
│   ├── users.ts
│   ├── coaches.ts
│   ├── workouts.ts
│   ├── messages.ts
│   ├── sessions.ts
│   └── subscriptions.ts
│
├── lib/                          # Utilities, constants, formatters
│   ├── api.ts                    # Axios instance + interceptors
│   ├── storage.ts                # SecureStore helpers
│   ├── queryClient.ts            # TanStack Query config
│   ├── purchases.ts              # RevenueCat setup
│   └── notifications.ts          # Push notification helpers
│
└── hooks/                        # Reusable cross-feature hooks
    ├── useDebounce.ts
    └── useFeatureFlag.ts
```

**Why `features/` instead of flat `hooks/` + `types/` + `components/`:** When you have 35+ screens, a flat structure where all hooks live in one folder and all types in another gets hard to navigate. Feature-based colocated modules keep related code together — the workout feature owns its own components, hooks, and utilities. Shared UI lives in `components/`, shared hooks in `hooks/`, and the API layer stays centralized since it's generated from the contract.

---

## 9) Provider Hierarchy

```
<QueryClientProvider>               # TanStack Query (server state)
  <AuthProvider>                    # Auth state (tokens, login/logout)
    <ThemeProvider>                 # Theme tokens + mode
      <AppConfigProvider>           # Feature flags, env config
        <Router>
          (auth)/*      → No additional providers
          (onboarding)/* → No additional providers
          (coach)/*     → No additional providers
          (client)/*    → No additional providers
        </Router>
      </AppConfigProvider>
    </ThemeProvider>
  </AuthProvider>
</QueryClientProvider>
```

**Note:** There are no `CoachProvider` or `ClientProvider` wrappers. Coach-specific data (client lists, templates) and client-specific data (assigned workouts, upcoming sessions) are fetched and cached via TanStack Query hooks within those route groups. The cache handles cross-screen data sharing automatically.

The one exception is `WorkoutEditorContext`, which is mounted only during the workout create/edit flow to manage complex multi-step form state. This is ephemeral client state, not server state, so Context is the right tool.

### Context Definitions

#### AuthContext

```typescript
interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  accessToken: string | null;
  refreshToken: string | null;
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<void>;
}

// Persists tokens to SecureStore
// Auto-refreshes tokens on app load
// Handles token refresh on 401 responses via Axios interceptor
```

#### ThemeContext

```typescript
interface ThemeContextValue {
  mode: 'light';              // MVP: light only, dark planned
  tokens: TamaguiTokens;      // Color roles, spacing, radii, typography
  setMode: (mode: 'light' | 'dark') => void;  // Wired for future
}
```

#### AppConfigContext

```typescript
interface AppConfigContextValue {
  featureFlags: Record<string, boolean>;
  environment: 'development' | 'staging' | 'production';
  apiBaseUrl: string;
  isFeatureEnabled: (flag: string) => boolean;
}

// Feature flags gate unreleased features:
//   'nutrition_tracking' → false for MVP
//   'progress_photos' → false for MVP
//   'oauth_login' → false for MVP
//   'dark_mode' → false for MVP
//   'photo_messages' → false for MVP
```

#### WorkoutEditorContext (Scoped, Not Global)

```typescript
interface WorkoutEditorState {
  workout: WorkoutDraft;
  exercises: WorkoutExerciseDraft[];
  isDirty: boolean;
  isSaving: boolean;
}

interface WorkoutEditorContextValue extends WorkoutEditorState {
  setWorkoutName: (name: string) => void;
  setScheduledDate: (date: Date) => void;
  addExercise: (exercise: Exercise) => void;
  removeExercise: (index: number) => void;
  updateExercise: (index: number, data: Partial<WorkoutExerciseDraft>) => void;
  reorderExercises: (fromIndex: number, toIndex: number) => void;
  saveWorkout: () => Promise<void>;
  discardChanges: () => void;
}

// Mounted only during workout creation/editing
// Handles unsaved changes warnings
// Supports drag-and-drop reordering
```

---

## 10) Navigation Architecture

### Navigation Flow

```
App Launch
    │
    ▼
┌─────────────────┐
│  Check Auth      │
│  (AuthContext)   │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
No Token    Has Token
    │         │
    ▼         ▼
(auth)/    Validate Token
welcome         │
           ┌────┴────┐
           │         │
           ▼         ▼
        Invalid    Valid
           │         │
           ▼         ▼
        (auth)/   Check Onboarding
        welcome        │
                  ┌────┴────┐
                  │         │
                  ▼         ▼
             Not Done    Done
                  │         │
                  ▼         ▼
           (onboarding)  Check Role
                             │
                        ┌────┴────┐
                        │         │
                        ▼         ▼
                     Coach     Client
                        │         │
                        ▼         ▼
                    (coach)/  (client)/
                    index     index
```

### Tab Navigators

#### Coach Tabs

```typescript
const coachTabs = [
  { name: 'index',    title: 'Home',     icon: 'home' },
  { name: 'clients',  title: 'Clients',  icon: 'users' },
  { name: 'schedule', title: 'Schedule', icon: 'calendar' },
  { name: 'messages', title: 'Messages', icon: 'message-circle' },
];
```

#### Client Tabs

```typescript
const clientTabs = [
  { name: 'index',    title: 'Workout',  icon: 'dumbbell' },
  { name: 'schedule', title: 'Sessions', icon: 'calendar' },
  { name: 'messages', title: 'Chat',     icon: 'message-circle' },
];
// Note: Nutrition tab deferred to post-MVP (see Section 24)
```

---

## 11) Design System

### UI Library: Tamagui (Locked)

Tamagui was selected after evaluating Tamagui, Shopify Restyle, NativeWind, React Native Paper, and UI Kitten. The decision was driven by:

- Token-first architecture that matches our design system needs
- Static extraction for performance (compiler-friendly)
- Strong TypeScript DX with Expo compatibility
- Scales to a premium custom visual identity without fighting framework defaults

Fallback policy: Do not switch UI libraries during MVP unless there is a blocking technical issue that prevents delivery.

### Color Tokens

```typescript
export const colors = {
  // Brand
  primary: '#18181B',         // Rich Black
  accent: '#D4AF37',          // Gold

  // Backgrounds
  background: '#FAFAFA',      // App background (light)
  surface: '#FFFFFF',         // Card/sheet background (light)

  // Text
  text: '#18181B',
  textSecondary: '#71717A',
  textTertiary: '#A1A1AA',
  textInverse: '#FAFAFA',

  // Gray scale
  gray50: '#FAFAFA',
  gray100: '#F4F4F5',
  gray200: '#E4E4E7',
  gray300: '#D4D4D8',
  gray400: '#A1A1AA',
  gray500: '#71717A',
  gray600: '#52525B',
  gray700: '#3F3F46',
  gray800: '#27272A',
  gray900: '#18181B',
  gray950: '#09090B',

  // Semantic
  success: '#22C55E',
  successLight: '#DCFCE7',
  warning: '#F59E0B',
  warningLight: '#FEF3C7',
  error: '#EF4444',
  errorLight: '#FEE2E2',
  info: '#3B82F6',
  infoLight: '#DBEAFE',
} as const;
```

### Typography

```typescript
export const typography = {
  fontFamily: {
    body: 'Inter',        // Body text, UI labels
    heading: 'Sora',      // Headings, high-emphasis text
  },

  weights: {
    regular: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
  },

  presets: {
    h1:        { family: 'Sora',  size: 30, weight: '700', lineHeight: 1.25 },
    h2:        { family: 'Sora',  size: 24, weight: '700', lineHeight: 1.25 },
    h3:        { family: 'Sora',  size: 20, weight: '600', lineHeight: 1.25 },
    h4:        { family: 'Sora',  size: 18, weight: '600', lineHeight: 1.25 },
    body:      { family: 'Inter', size: 16, weight: '400', lineHeight: 1.5  },
    bodySmall: { family: 'Inter', size: 14, weight: '400', lineHeight: 1.5  },
    caption:   { family: 'Inter', size: 12, weight: '400', lineHeight: 1.5  },
    label:     { family: 'Inter', size: 14, weight: '500', lineHeight: 1.25 },
    button:    { family: 'Inter', size: 16, weight: '600', lineHeight: 1.25 },
  },
} as const;
```

### Spacing and Sizing

```typescript
export const spacing = {
  0: 0, 1: 4, 2: 8, 3: 12, 4: 16, 5: 20,
  6: 24, 8: 32, 10: 40, 12: 48, 16: 64, 20: 80,
} as const;

export const borderRadius = {
  none: 0, sm: 4, md: 8, lg: 12, xl: 16, '2xl': 24, full: 9999,
} as const;
```

### Component Tiers

**Tier 1 — Primitives** (Tamagui-based building blocks):
`Box`, `Text`, `Stack`, `Icon`

**Tier 2 — Composites** (reusable across domains):

| Component | Variants | Key Props |
|-----------|----------|-----------|
| Button | primary, secondary, outline, ghost, danger | size (sm/md/lg), loading, disabled, fullWidth, leftIcon, rightIcon |
| Input | default, error | label, placeholder, error, secureTextEntry, leftIcon, rightIcon |
| Card | elevated, outlined, filled | padding, onPress |
| Avatar | default | size (sm/md/lg/xl), source, fallbackInitials |
| Badge | default, success, warning, error, info | label |
| ListItem | default, pressable | title, subtitle, leftElement, rightElement, onPress |
| Modal | default | visible, onClose, title |
| BottomSheet | default | snapPoints, onClose |
| Toast | success, error, info, warning | message, duration |
| Skeleton | default | width, height, borderRadius |
| EmptyState | default | icon, title, description, actionLabel, onAction |

**Tier 3 — Domain Components** (feature-specific):
`WorkoutCard`, `ExerciseCard`, `SetLogger`, `RestTimer`, `SessionSlotRow`, `SessionCard`, `ConversationListItem`, `MessageBubble`, `ChatInput`, `ClientCard`, `InviteCodeCard`

---

## 12) UX and Visual Design Principles

- Clean hierarchy: clear information density, whitespace-first layouts
- Strong readability: predictable type scale and spacing rhythm via Sora headings + Inter body
- Motion with intent: subtle transitions, avoid decorative animation clutter
- Fast paths: most-used actions reachable in 1-2 interactions
- Accessible by default: color contrast, text scaling, touch targets, screen reader semantics
- Every screen has four states: loading (skeleton), empty (illustration + CTA), error (message + retry), success (content)

---

## 13) Feature Specifications (MVP)

### F1: Authentication

#### F1.1: Welcome Screen
- Logo + tagline
- "Get Started" button → Sign Up
- "I have an account" link → Sign In

#### F1.2: Sign Up
- Email input
- Password input (8+ chars, 1 uppercase, 1 number)
- Confirm password input
- Terms & Privacy checkbox
- Create Account button with loading state
- OAuth buttons hidden behind `oauth_login` feature flag (see Section 24)

#### F1.3: Sign In
- Email input
- Password input
- "Forgot Password" link
- Sign In button with loading state
- OAuth buttons hidden behind feature flag

#### F1.4: Forgot Password
- Email input
- Send Reset Link button
- Success confirmation with "check your email" message

#### F1.5: Role Selection (post-registration)
- "I'm a Coach" card with icon/description
- "I'm a Client" card with icon/description
- Selection navigates to appropriate onboarding flow

### F2: Coach Onboarding

#### F2.1: Step 1 — Profile
- Profile photo upload (optional, uses expo-image-picker)
- First name (required)
- Last name (required)
- Phone (optional)

#### F2.2: Step 2 — Business
- Business name (optional, defaults to "{First} {Last} Coaching")
- Client count tier (radio: 0-5, 6-20, 21-50, 50+)
- Coaching specialties (multi-select chips: Strength, CrossFit, Weight Loss, Bodybuilding, Sports Performance, General Fitness)

#### F2.3: Step 3 — Setup
- Timezone selector (auto-detect default)
- Push notification permission request
- "Complete Setup" button

### F3: Client Onboarding

#### F3.1: Step 1 — Profile
- Profile photo upload (optional)
- First name (required)
- Last name (required)

#### F3.2: Step 2 — Connect to Coach
- Invite code input (or pre-filled if arrived via deep link `chalk://invite/{code}`)
- Coach preview card (shows coach photo + name + specialties from invite preview endpoint)
- "Connect" button

### F4: Coach Dashboard

#### Layout
- Header: "Dashboard" + profile avatar (tappable → settings)
- Stats row: Active Clients count, Workouts This Week count
- Today's Sessions list (if any)
- Recent Activity feed (workout completions, new messages)

#### Interactions
- Tap stat card → navigate to relevant section
- Tap session → session detail
- Tap activity item → relevant detail screen

### F5: Coach — Client Management

#### F5.1: Client List
- Search bar (filter by name)
- Filter chips: All, Active, Pending
- Client cards: avatar, name, last activity timestamp, status badge
- FAB: "Add Client"

#### F5.2: Add Client / Invite
- Generate invite link button
- Display shareable link + share sheet integration
- Manual invite code display
- List of pending (active) invite codes
- Deactivate invite code action

#### F5.3: Client Detail
- Header: avatar, name, status badge
- Quick action buttons: "Message", "Assign Workout"
- Workouts tab: assigned workout list with status indicators
- Info tab: coach notes (editable), tags

### F6: Coach — Workout Programming

#### F6.1: Create Workout
- Workout name input
- Date picker (assignment date)
- Exercise list (empty initially)
- "Add Exercise" button → Exercise Picker modal
- Each exercise row: name, sets × reps @ weight, drag handle
- Long press → drag to reorder
- Swipe left → delete exercise
- "Save as Template" toggle
- "Assign to Client" button (client selector)

#### F6.2: Exercise Picker
- Search bar
- Exercise list with name + muscle group tags
- Tap → adds to current workout
- Note: Exercise library is template-based from backend. Custom exercise creation uses workout template exercise creation endpoint.

#### F6.3: Exercise Detail Editor (within workout creation)
- Exercise name
- Sets input (number stepper)
- Reps input (text, supports "8-12" range or "AMRAP")
- Weight input (text, supports "135lbs" or "RPE 8")
- Rest input (seconds stepper)
- Notes input (multiline)
- Superset toggle (groups with previous exercise)

#### F6.4: Template Management
- Template list screen (coach's saved templates)
- Template detail: view exercises, duplicate to new workout, edit
- Create from existing workout

### F7: Coach — Scheduling

#### F7.1: Schedule Calendar
- Month/week view toggle
- Days with sessions show dot indicator
- Tap day → show sessions for that day
- Session cards: time, client name, type, status badge
- Tap session → session detail

#### F7.2: Availability Settings
- Day-by-day weekly availability
- Each day: on/off toggle + time slot ranges
- Add multiple slots per day
- Buffer between sessions input
- Default session duration

#### F7.3: Session Types
- List of session types
- Each: name, duration, color indicator
- Add / edit / delete session types

#### F7.4: Session Detail
- Client info (name, avatar)
- Session type + duration
- Date/time
- Status badge: Scheduled | Completed | Cancelled | No-Show
- Actions (permission-gated):
  - Mark Complete
  - Mark No-Show
  - Cancel Session

### F8: Coach — Messaging

#### F8.1: Conversation List
- Search bar
- Conversations sorted by most recent message
- Each row: client avatar, name, last message preview, timestamp
- Unread indicator dot
- Tap → chat screen

#### F8.2: Chat Screen
- Header: client name + avatar
- Message list grouped by date
- Text messages only (photo messages deferred — see Section 24)
- Input bar: text input + send button
- Sending indicator on outbound messages
- Unread state clears on conversation open (mark as read)

### F9: Client — Today's Workout

#### F9.1: Workout View
- Header: workout name + date
- Status indicator: Not Started | In Progress | Completed
- "Start Workout" button (if not started)
- Exercise list:
  - Exercise name + muscle group tags
  - Target: sets × reps @ weight
  - Progress indicator (e.g., "2/4 sets complete")
  - Tap → exercise logging view
- "Complete Workout" button (after all exercises addressed)
- Coach notes section (read-only)

#### F9.2: Exercise Logging
- Exercise name + instructions
- Set logger table:
  - Set # | Target Reps | Actual Reps | Actual Weight | ✓ complete
  - Tap row to edit values
  - Quick input: +/- steppers for reps and weight
- Rest timer (starts automatically after completing a set)
- "Skip Exercise" button
- Notes input
- "Next Exercise" button

#### F9.3: Workout Calendar
- Calendar view with dot indicators on days with assigned workouts
- Tap day → show workout(s) for that date
- Tap workout → workout detail

### F10: Client — Scheduling

#### F10.1: My Sessions
- Upcoming sessions list (sorted by date)
- Past sessions list
- Each: date, time, session type, status badge
- Tap → session detail (read-only with cancel option for scheduled sessions)
- "Book New Session" FAB

#### F10.2: Book Session
- Session type selector
- Calendar showing available dates (computed from coach availability)
- Time slot grid for selected date
- "Confirm Booking" button
- Cancellation policy note

### F11: Client — Messaging

- Single conversation with coach
- Same chat UI as coach side (F8.2)
- Push notification for new messages

### F12: Subscription and Feature Gates

#### F12.1: Subscription Screen (Coach)
- Current plan display
- Usage info (e.g., "3/3 clients on Free tier")
- "Upgrade" button → paywall

#### F12.2: Paywall
- Plan comparison layout
- Free tier: limited client count
- Pro Monthly: $29/mo — unlimited clients
- Pro Annual: $249/yr (save 28%)
- Feature list comparison
- Subscribe button (RevenueCat purchase flow)
- "Restore Purchases" link

#### F12.3: Feature Gate UX
- When a user hits a gated feature, show inline prompt with upgrade CTA
- Backend `/features/{feature}` endpoint determines access
- AppConfigContext surfaces gate decisions to UI components

### F13: Settings

- Profile section → Edit Profile screen (name, photo, phone)
- Notification preferences toggle
- Help & Support link
- Terms of Service link
- Privacy Policy link
- Log Out button
- Delete Account button (with confirmation)

---

## 14) Screens Inventory (MVP)

### Total: 37 screens

| Section | Screen | Priority |
|---------|--------|----------|
| **Auth (5)** | | |
| | Welcome | P0 |
| | Sign Up | P0 |
| | Sign In | P0 |
| | Forgot Password | P1 |
| | Role Select | P0 |
| **Onboarding (5)** | | |
| | Coach Step 1 — Profile | P0 |
| | Coach Step 2 — Business | P0 |
| | Coach Step 3 — Setup | P0 |
| | Client Step 1 — Profile | P0 |
| | Client Step 2 — Connect | P0 |
| **Coach (15)** | | |
| | Dashboard | P0 |
| | Client List | P0 |
| | Client Add/Invite | P0 |
| | Client Detail | P0 |
| | Client Workouts | P0 |
| | Workout Create | P0 |
| | Workout Edit | P0 |
| | Template List | P1 |
| | Template Detail | P1 |
| | Schedule Calendar | P0 |
| | Availability | P0 |
| | Session Types | P0 |
| | Session Detail | P0 |
| | Conversation List | P0 |
| | Chat | P0 |
| **Client (8)** | | |
| | Today's Workout | P0 |
| | Workout Calendar | P0 |
| | Workout Detail + Exercise Log | P0 |
| | Sessions List | P0 |
| | Book Session | P0 |
| | Chat | P0 |
| **Shared (4)** | | |
| | Settings | P0 |
| | Edit Profile | P0 |
| | Subscription | P0 |
| | Paywall | P0 |

---

## 15) Performance and Quality Targets

| Metric | Target | Why |
|--------|--------|-----|
| Cold start (authenticated, modern iPhone) | < 2.5s | Users abandon apps that feel sluggish on launch |
| Screen transitions | < 300ms | Maintains perception of native-quality responsiveness |
| Core list FPS (workout list, client list, messages) | 55-60 FPS sustained | Janky scrolling destroys trust in a fitness app |
| API error handling | No silent failures | Every failed request surfaces user-visible recovery UI |
| Crash-free sessions | 99.5%+ in MVP beta | Table stakes for App Store quality |
| App bundle size | < 50MB | Reduces install friction |

---

## 16) MVP Acceptance Criteria

- **Auth:** User can register, login, refresh token, logout, and remain logged in across app relaunches.
- **Coach profile:** Coach can edit profile and create/list/deactivate invite codes end-to-end.
- **Client invite:** Invite link preview and acceptance works from unauthenticated state through to authenticated + connected.
- **Workouts:** Coach can create workout with exercises, assign to client. Client can view assigned workouts, start workout, complete/skip exercises, log sets, and complete workout.
- **Messaging:** User can list conversations, open chat, send text message, mark conversation read, and see unread count update in tab badge.
- **Sessions:** Coach can set availability and session types. Client can view bookable slots, book session. Both can cancel. Coach can mark complete/no-show.
- **Subscription:** App reflects current subscription status from backend and properly gates premium features with upgrade CTAs.
- **Error UX:** Every API failure path has a user-visible state with retry/recovery action where applicable.
- **States:** Every screen implements loading (skeleton), empty (illustration + CTA), error (message + retry), and success states.

---

## 17) Accessibility Requirements

- Dynamic Type support (text scales with system font size preference)
- Touch target minimum 44×44 pt
- Screen reader labels (`accessibilityLabel`) on all interactive controls and key state indicators
- Color contrast meets WCAG AA
- `prefers-reduced-motion` support for animation-sensitive users
- Semantic heading hierarchy for screen reader navigation

---

## 18) Security and Privacy Requirements

- Store auth tokens in `expo-secure-store` (Keychain on iOS, Keystore on Android)
- No PII in application logs
- Redact sensitive payloads (tokens, passwords) in network logging
- Enforce auth boundary in routing: unauthenticated users cannot access protected routes
- Enforce auth boundary in data fetching: TanStack Query hooks should not fire without valid auth state
- Certificate pinning: evaluate for post-MVP hardening

---

## 19) Analytics and Observability (MVP Lite)

### Key Conversion Events

- `registration.success`
- `onboarding.completed`
- `invite.accepted`
- `workout.assigned` (coach)
- `workout.completed` (client)
- `session.booked`
- `message.first_sent`
- `subscription.upgraded`

### Screen View Tracking

- Every screen navigation logs a screen view event

### Error Reporting

- Sentry (or equivalent) integrated from day 1
- Crash reporting + non-fatal error tracking
- Breadcrumbs for navigation and API call context

---

## 20) Release Strategy

- **Team:** Solo developer
- **Target timeline:** 18 weeks to App Store submission (with buffer)
- **Strategy:** Revenue path first — get coaches paying before the product is feature-complete

1. **Phase 1 gate:** Coach can sign up, onboard, invite clients, and hit paywall → internal alpha
2. **Phase 2 gate:** Coach can create/assign workouts, message clients → TestFlight beta with 3-5 real coaches
3. **Phase 3 gate:** Scheduling works, polish pass complete → iOS App Store submission
4. **Phase 4:** Android parity after iOS stabilizes (post-launch)

---

## 21) Delivery Plan (Solo Developer)

> **Principle:** Build the revenue path first. Each phase produces a testable milestone. Ship when Phase 3 is done — don't wait for perfection.

### Phase 1: Revenue Path (Weeks 1-6)

*Goal: A coach can sign up, set up their profile, invite clients, and hit a paywall. You can validate willingness to pay before the product does anything useful.*

**Week 1-2: Scaffold + Design System**
- [ ] Initialize Expo project with TypeScript strict mode
- [ ] Set up OpenAPI type generation pipeline (`docs/openapi.json` → `src/api/generated/`)
- [ ] Configure Tamagui with color tokens, spacing, typography (Inter + Sora)
- [ ] Set up Expo Router with all route groups (auth, onboarding, coach, client, shared)
- [ ] Install and configure TanStack Query, Axios, SecureStore
- [ ] Set up ESLint + Prettier
- [ ] Build Tier 2 component library: Button, Input, Card, Avatar, Badge, ListItem, EmptyState, Skeleton, Toast
- [ ] Don't build screens yet — validate components in isolation first

**Week 3-4: Auth + Onboarding**
- [ ] Build AuthContext + SecureStore token persistence
- [ ] Build Axios interceptor (attach token, 401 → refresh → retry/logout)
- [ ] Build Welcome, Sign Up, Sign In screens
- [ ] Build Forgot Password screen
- [ ] Build Role Selection screen
- [ ] Build Coach onboarding flow (3 steps: profile, business, setup)
- [ ] Build Client onboarding flow (2 steps: profile, connect)
- [ ] Connect all to auth + user API endpoints
- [ ] Test: register → login → onboard → land on coach/client shell

**Week 5-6: Coach Profile + Invites + Paywall**
- [ ] Build AppConfigContext with feature flags
- [ ] Build Coach profile view/edit screen
- [ ] Build invite code generation, list, deactivation
- [ ] Build invite share (share sheet integration)
- [ ] Build invite preview screen (public, no auth)
- [ ] Build client invite accept flow (deep link `chalk://invite/{code}`)
- [ ] Integrate RevenueCat (react-native-purchases)
- [ ] Build Subscription screen + Paywall
- [ ] Implement feature gate checks (backend `/features/{feature}` endpoint)
- [ ] Build Settings screen (profile, logout, delete account)
- [ ] Test end-to-end: coach signs up → invites client → client accepts → coach hits client limit → paywall

**Phase 1 checkpoint:** Demo to 2-3 coaches. Validate that onboarding flow makes sense and they understand the value proposition enough to consider paying. This is your earliest signal.

---

### Phase 2: Core Value Loop (Weeks 7-12)

*Goal: The app is genuinely useful. Coaches create workouts and message clients. Clients log their training. This is what people open the app for daily.*

**Week 7-8: Workout Programming (Coach Side)**
- [ ] Build WorkoutEditorContext (scoped to create/edit flow)
- [ ] Build Workout Create screen with exercise picker
- [ ] Build exercise detail editor (sets, reps, weight, rest, notes, superset toggle)
- [ ] Build drag-to-reorder (react-native-reanimated or similar)
- [ ] Build swipe-to-delete on exercises
- [ ] Build Template list + detail screens
- [ ] Build "Save as Template" toggle on workout creation
- [ ] Build Coach → Client Detail → Workouts tab (assigned workout list)
- [ ] Connect to workout template + assignment API endpoints
- [ ] Test: coach creates workout → assigns to client → workout appears in client list

**Week 9-10: Workout Execution (Client Side)**
- [ ] Build Client Today's Workout screen (workout view + start button)
- [ ] Build Exercise Logging screen (set logger table, actual reps/weight input)
- [ ] Build Rest Timer component (auto-starts after set completion)
- [ ] Build exercise skip flow
- [ ] Build "Complete Workout" action
- [ ] Build Workout Calendar (client, dot indicators per day)
- [ ] Connect to workout execution + logging API endpoints
- [ ] Test: client opens assigned workout → starts → logs sets → completes → coach sees completion

**Week 11-12: Messaging**
- [ ] Build Conversation List screen (both roles)
- [ ] Build Chat screen (text messages, grouped by date)
- [ ] Build ChatInput component (text input + send button)
- [ ] Build MessageBubble component (sent vs received styling)
- [ ] Implement unread count badge on Messages tab
- [ ] Implement mark-as-read on conversation open
- [ ] Set up push notifications (expo-notifications)
- [ ] Register device token with backend on login
- [ ] Connect to messaging API endpoints
- [ ] Test: coach sends message → client gets push → opens chat → replies → coach sees unread badge

**Phase 2 checkpoint:** Put the app in 3-5 real coaches' hands via TestFlight. Watch them create a workout and assign it. Watch a client log a session. Collect feedback on friction points. This is your most important feedback loop — prioritize fixing what real users struggle with over building new features.

---

### Phase 3: Scheduling + Polish (Weeks 13-17)

*Goal: The app feels complete and professional. Scheduling rounds out the feature set. Polish turns "works" into "feels good."*

**Week 13-14: Scheduling**
- [ ] Build Coach Schedule Calendar (month/week view, session dots)
- [ ] Build Availability Settings screen (day-by-day slots)
- [ ] Build Session Types management (name, duration, color)
- [ ] Build Session Detail screen (status transitions: complete, no-show, cancel)
- [ ] Build Client Sessions List screen (upcoming + past)
- [ ] Build Client Book Session flow (type → date → time slot → confirm)
- [ ] Connect to availability, session types, bookable slots, and session lifecycle API endpoints
- [ ] Test: coach sets availability → client sees slots → books session → coach marks complete

**Week 15-16: Polish + Hardening**
- [ ] Error boundaries on all route groups
- [ ] Loading skeletons on all list screens
- [ ] Empty state illustrations + CTAs on all list screens (clients, workouts, sessions, messages)
- [ ] Retry/recovery actions on all API error states
- [ ] Accessibility pass: touch targets (44×44), screen reader labels, text scaling
- [ ] Performance pass: measure cold start, list FPS, identify and fix jank
- [ ] Android sanity test (install on Android emulator, verify all flows)
- [ ] Edge cases: expired tokens, network loss mid-action, empty coach with no clients, deep link when logged out

**Week 17: Ship Prep**
- [ ] App icons and splash screen
- [ ] App Store screenshots (use simulator, real data)
- [ ] App Store description and metadata
- [ ] Final TestFlight build to beta testers
- [ ] Bug fixes from beta feedback
- [ ] Sentry crash reporting verified
- [ ] Analytics events verified (registration, invite, workout complete, session booked, subscription upgrade)

**Phase 3 checkpoint:** App Store submission. Target 99.5% crash-free rate from TestFlight data before submitting.

---

### Phase 4: Post-Launch (Week 18+)

- [ ] Monitor crash reports and fix critical issues
- [ ] Respond to App Store review feedback if rejected
- [ ] Begin Android parity work
- [ ] Evaluate Phase 2 user feedback for next feature priority (likely nutrition or OAuth — see Section 24)

---

### Solo Dev Survival Rules

1. **Don't polish what isn't validated.** Pixel-perfect animations on the workout editor don't matter if coaches don't understand the invite flow. Ship ugly, get feedback, polish what people actually use.

2. **One screen at a time.** Build, connect to API, test, move on. Don't build 5 screens then try to wire them all up — you'll lose context and introduce integration bugs.

3. **Test on device every day.** The simulator lies. Especially for keyboard handling, touch targets, scroll performance, and push notifications.

4. **Time-box yak shaves.** If a Tamagui config issue or Expo build problem takes more than 2 hours, document it, work around it, and come back later. Don't lose a day to tooling.

5. **Weekly self-demo.** Every Friday, record a 2-minute screen recording of the app's current state. This forces you to see the product through a user's eyes and tracks progress when the work feels invisible.

6. **Feature flags are your friend.** If you start building something and realize it's bigger than expected, flag it and move on. A shipped app with 80% of features beats an unshipped app with 100%.

---

## 22) Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Solo dev burnout / timeline slip | Critical — project dies | Phase checkpoints with real user feedback; each phase is a shippable increment; take weekends off |
| Scope creep from future features leaking into MVP | High — delays launch | Feature flags gate all post-MVP features; if it's not in Phase 1-3, it doesn't exist yet |
| Design inconsistency from moving fast | Medium — erodes premium feel | Build component library in Week 1-2 before any screens; reuse, don't reinvent |
| Context overuse causing re-render churn | High — degrades performance | Contexts limited to 3 global + 1 scoped; all server state in TanStack Query |
| API contract drift between backend and frontend types | High — runtime errors | Regenerate API types from `docs/openapi.json` before each phase; fail build on drift |
| iOS-first assumptions breaking Android | Medium — delays Android launch | Android sanity test in Week 15; don't fix Android bugs during Phases 1-2 |
| Tamagui learning curve slows early weeks | Medium — timeline risk | Week 1-2 is dedicated to component library; absorb the learning curve before feature pressure |
| Tooling yak-shaves (Expo builds, Tamagui config, codegen) | Medium — loses days | 2-hour time-box rule; document, workaround, revisit later |
| No code review or QA partner | Medium — bugs ship | Weekly self-demo recordings; TestFlight beta testers as QA; Sentry from day 1 |
| Premature optimization / polish paralysis | High — delays launch | "Don't polish what isn't validated" rule; ship ugly, fix what real users complain about |

---

## 23) Final Decisions (Locked for MVP)

- **Platform strategy:** iOS-first delivery, then Android parity
- **Language/runtime:** TypeScript strict mode + latest stable Expo/RN at kickoff
- **State architecture:** React Context for cross-cutting app state (3 contexts) + TanStack Query for all server state
- **UI architecture:** Tamagui design system with token-first primitives and composable feature components
- **Typography:** Inter (body/UI) + Sora (headings) via Expo Google Fonts
- **Icons:** Lucide primary, Expo Vector Icons fallback
- **SVG assets:** Curated free assets (unDraw/SVG Repo) via react-native-svg pipeline
- **API contract:** Generated frontend types from `docs/openapi.json` are source of truth
- **Offline behavior:** No full offline mode; graceful failure + retry for network-dependent actions
- **Notification UX:** Basic push handling only; advanced notification center deferred
- **Auth:** Email/password only for MVP; OAuth deferred behind feature flag
- **Messaging:** Text only for MVP; photo messages deferred behind feature flag

---

## 24) Future Features and Monetization Roadmap

These features are intentionally excluded from MVP scope but are planned for post-launch development. Each represents a monetization lever or competitive differentiator. Backend database schemas for nutrition and progress already exist (tables migrated), and the Open Food Facts integration client is ready — these features are waiting on API endpoint implementation.

### Tier 1 — High-Priority Post-MVP (Backend Partially Ready)

#### Nutrition Tracking

**Monetization angle:** Premium feature — free tier gets basic logging, Pro unlocks macro targets, coach comments, and food database search.

**Scope:**
- Client daily nutrition view: circular calorie progress, macro bars (protein/carbs/fat), meal-based food logging
- Add food flow: search Open Food Facts database, recent/favorites, quick-add (manual macros)
- Coach nutrition view: per-client daily log with macro progress, set nutrition targets (cal/P/C/F), add coach comments per meal
- Barcode scanning (expo-camera): scan packaged food → auto-populate from database

**Backend requirements:** Nutrition API endpoints (CRUD for food logs, targets, search proxy to Open Food Facts), meal model

**Feature flag:** `nutrition_tracking`

**Estimated effort:** 3-4 sprints

**Design tokens to add:**
```typescript
// Macro-specific colors for nutrition UI
protein: '#EF4444',    // Red
carbs: '#3B82F6',      // Blue
fat: '#F59E0B',        // Amber
```

#### Progress Tracking

**Monetization angle:** Premium feature — progress photos, body measurements, and data export create high stickiness and justify subscription.

**Scope:**
- Weight chart (line graph over time)
- Log weight entry
- Body measurements section (expandable: chest, waist, hips, arms, etc.)
- Progress photos grid with date stamps
- Add photo flow (expo-image-picker + upload)
- Export data to CSV

**Backend requirements:** Progress API endpoints (weight logs, measurement logs, photo upload/storage), S3 or equivalent for image storage

**Feature flag:** `progress_tracking`

**Estimated effort:** 2-3 sprints

#### OAuth Sign-In (Google / Facebook / Apple)

**Monetization angle:** Reduces registration friction → higher conversion from install to active user.

**Scope:**
- Google OAuth button on Sign Up / Sign In
- Facebook OAuth button
- Apple Sign In (required by App Store if other OAuth present)
- Account linking for existing email/password users

**Backend requirements:** OAuth provider flows finalized (schema exists, flows not implemented yet)

**Feature flag:** `oauth_login`

**Estimated effort:** 1-2 sprints

### Tier 2 — Medium-Priority Post-MVP

#### Photo Messaging

**Monetization angle:** Enriches coach-client communication — form check photos, meal photos, etc.

**Scope:**
- Camera button in chat input bar
- Photo preview before send
- Photo messages in conversation view (thumbnail + tap to expand)
- Image upload to backend storage

**Backend requirements:** Message attachment model, image upload endpoint, storage integration

**Feature flag:** `photo_messages`

**Estimated effort:** 1-2 sprints

#### Dark Mode

**Monetization angle:** Not directly monetizable, but high user demand and retention signal.

**Scope:**
- ThemeContext already supports mode switching
- Tamagui token architecture already supports light/dark token sets
- Dark color palette definition
- System preference detection + manual toggle in Settings

**Backend requirements:** None (purely frontend)

**Feature flag:** `dark_mode`

**Estimated effort:** 1 sprint (token definitions + testing)

#### Exercise Library with Media

**Monetization angle:** Premium exercise library with video demos is a differentiator vs. competitors with text-only instructions.

**Scope:**
- Exercise detail screen with animated GIF or video preview
- Exercise categories: muscle group, equipment type, movement pattern
- Favorite/recent exercise filtering
- Custom exercise creation: name, description, muscle groups, equipment, video URL

**Backend requirements:** Exercise media storage, enhanced exercise model with media URLs

**Feature flag:** `exercise_media`

**Estimated effort:** 2 sprints

### Tier 3 — Longer-Term Growth

#### Coach Dashboard Analytics
- Client engagement metrics, completion rates, revenue tracking
- Requires backend analytics/reporting layer

#### Advanced Notification Center
- In-app notification feed with read/unread state
- Notification preferences per category
- Requires backend notification preference management

#### Invoice and Payments
- Coach can invoice clients for sessions
- Payment processing integration
- Requires backend invoice/payments domain

#### Client Tab: Nutrition (Tab Addition)
When nutrition tracking ships, the client tab bar expands:
```typescript
const clientTabs = [
  { name: 'index',     title: 'Workout',   icon: 'dumbbell' },
  { name: 'nutrition',  title: 'Nutrition',  icon: 'utensils' },
  { name: 'schedule',  title: 'Sessions',  icon: 'calendar' },
  { name: 'messages',  title: 'Chat',       icon: 'message-circle' },
];
```

---

## 25) Definition of Done (MVP)

- **Product:** Feature matches PRD scope (Section 13) and acceptance criteria (Section 16).
- **Engineering:** TypeScript strict mode passes with zero `any`-based bypasses in core flows.
- **Quality:** lint/typecheck green on every commit. Test suite covers auth flow and critical mutations.
- **UX:** Loading/empty/error/success states implemented for all screens in scope.
- **Accessibility:** Touch targets (44×44), screen reader labels, and text scaling validated on key flows.
- **Performance:** Startup < 2.5s, list FPS 55-60, bundle < 50MB on target iOS devices.
- **Observability:** Sentry crash reporting and core product events instrumented for all primary flows.
- **Feature flags:** All post-MVP features gated and confirmed non-leaking in MVP build.
- **Real user validation:** At least 3-5 coaches have used the app via TestFlight and critical feedback has been addressed.

---

## Appendix A: Asset Licensing Policy

All visual assets used in the app must be free and license-compliant for commercial use. Every imported asset (font, icon, illustration, SVG) must be logged in an internal asset registry with source URL, license type, and date added. No paid asset subscriptions in MVP.

## Appendix B: Backend API Reference

The canonical API contract lives in `docs/openapi.json`. Frontend types are generated from this file at build time. The backend PRD (`docs/backend-prd.md`) documents the full system architecture, domain modules, and operational standards. Key backend capabilities that directly inform frontend behavior:

- Auth tokens are JWT with refresh token rotation
- Invite codes have a preview endpoint (public, no auth required) and an accept endpoint (auth required)
- Workout assignment creates a deep copy from template — edits to the template don't affect assigned workouts
- Session booking checks availability + conflict detection server-side
- Subscription state is synced from RevenueCat webhooks — frontend reads, never writes subscription state directly
- Message events fan out to push notifications via transactional outbox — the frontend registers device tokens and receives pushes

## Appendix C: Post-MVP Route Additions

When future features ship, these route groups will be added:

```
app/(client)/nutrition/          # Nutrition tracking screens
  ├── index.tsx                  # Daily nutrition view
  ├── add.tsx                    # Add food entry
  ├── search.tsx                 # Food database search
  └── quick-add.tsx              # Quick macro entry

app/(client)/progress.tsx        # Progress tracking

app/(coach)/clients/[id]/
  ├── nutrition.tsx              # Client nutrition view (coach)
  └── progress.tsx               # Client progress view (coach)
```