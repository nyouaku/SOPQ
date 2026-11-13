# Theme System - Stock Opname Dashboard

## Overview

Dashboard menyediakan **3 layout berbeda** dan **3 tema warna** yang bisa dikombinasikan sesuai preferensi Anda.

**Total Kombinasi:** 3 layouts × 3 colors = **9 tema berbeda!**

---

## 🎨 Color Themes

### 1. Rack Room (Dark Teal)
- **Background:** Dark gray (#14171C)
- **Primary:** Teal cyan (#5EEAD4)
- **Text:** Light gray (#C7CDD6)
- **Best for:** NOC environment, low-light rooms
- **Mood:** Professional, modern, tech-focused

### 2. NOC Console (Terminal Green)
- **Background:** Very dark (#0B0E0C)
- **Primary:** Bright green (#00FF9C)
- **Text:** Green (#00FF9C)
- **Best for:** Terminal enthusiasts, retro feel
- **Mood:** Classic NOC, command-line, hacker aesthetic

### 3. Daylight Ops (Light Blue)
- **Background:** Light gray (#F7F8FA)
- **Primary:** Blue (#1D4ED8)
- **Text:** Dark gray (#1C222B)
- **Best for:** Bright environments, daylight work
- **Mood:** Clean, professional, corporate

---

## 📐 Layout Styles

### 1. Classic Sidebar Layout 📋
**Best for:** Traditional desktop users, multi-tasking

**Features:**
- Fixed left sidebar navigation (64px wide)
- Vertical menu with icons
- User info at bottom
- Content area on right
- Always-visible navigation

**Structure:**
```
┌──────────┬────────────────────────┐
│          │                        │
│  LOGO    │     CONTENT AREA       │
│          │                        │
│  Menu 1  │                        │
│  Menu 2  │                        │
│  Menu 3  │                        │
│  Menu 4  │                        │
│  Menu 5  │                        │
│          │                        │
│  [User]  │                        │
│  Logout  │                        │
└──────────┴────────────────────────┘
```

**Pros:**
- Easy navigation access
- Familiar pattern
- Good for frequent menu switching

**Cons:**
- Takes horizontal space
- Less screen for content

---

### 2. Modern Top Nav Layout 🎯
**Best for:** Wide screens, content-focused work

**Features:**
- Horizontal navigation bar at top
- Full-width content area
- Center-aligned menu items
- Sticky header (always visible on scroll)
- Mobile-friendly dropdown

**Structure:**
```
┌────────────────────────────────────┐
│ LOGO  [Menu1][Menu2][Menu3]  USER  │
├────────────────────────────────────┤
│                                    │
│                                    │
│         FULL-WIDTH CONTENT         │
│                                    │
│                                    │
└────────────────────────────────────┘
```

**Pros:**
- Maximum content width
- Modern, clean look
- Better for reading/viewing data

**Cons:**
- Menu hidden on mobile (dropdown)
- More scrolling for long pages

---

### 3. Command Center Layout 🖥️
**Best for:** Ops teams, monitoring, power users

**Features:**
- 3-column layout (nav | content | status)
- Status bar at top with live time
- Left panel: navigation + system status
- Center panel: main content
- Right panel: user info, quick stats, activity log
- Monospace font (terminal style)
- Keyboard shortcuts displayed

**Structure:**
```
┌─────────────────────────────────────────┐
│ ●ONLINE | CONTROL | USER | TIME         │
├────┬────────────────────────┬───────────┤
│    │                        │           │
│NAV │   MAIN TERMINAL        │  STATUS   │
│    │                        │           │
│ D  │   [CONTENT AREA]       │  [USER]   │
│ S  │                        │           │
│ E  │                        │  [STATS]  │
│ P  │                        │           │
│ R  │                        │  [LOG]    │
│ T  │                        │           │
│    │                        │ [HOTKEYS] │
│SYS │                        │           │
│ ●  │                        │           │
└────┴────────────────────────┴───────────┘
```

**Pros:**
- Maximum information density
- Terminal/ops aesthetic
- Status panels always visible
- Keyboard shortcuts

**Cons:**
- Complex layout
- Requires wider screen
- May feel cluttered on small screens

---

## 🎨 How to Change Themes

### Via Settings Page

1. **Login** ke dashboard
2. Klik **Menu (☰)** → **Settings**
3. Scroll ke **"Theme Configuration"**

**Change Layout:**
- Klik salah satu dari 3 layout cards:
  - 📋 Classic Sidebar
  - 🎯 Modern Top Nav
  - 🖥️ Command Center

**Change Color:**
- Klik salah satu dari 3 color cards:
  - Rack Room (Dark Teal)
  - NOC Console (Terminal Green)
  - Daylight Ops (Light Blue)

**Changes apply instantly!** Navigate to Dashboard to see the new look.

---

## 💡 Recommended Combinations

### For NOC Operations
```
Layout: Command Center
Color: NOC Console
→ Full terminal experience
```

### For Management/Reports
```
Layout: Modern Top Nav
Color: Daylight Ops
→ Clean, professional look
```

### For Daily Warehouse Work
```
Layout: Classic Sidebar
Color: Rack Room
→ Easy navigation, modern dark theme
```

### For Mobile/Tablet
```
Layout: Modern Top Nav
Color: Any (based on environment lighting)
→ Best responsive experience
```

---

## 🔧 Technical Details

### Storage
- Themes saved to **localStorage**
- Persists across sessions
- Per-browser/device setting

### Implementation
- **Svelte stores** for state management
- **TailwindCSS + DaisyUI** for styling
- **Dynamic component loading** based on layout
- **CSS custom properties** for colors

### Files
```
dashboard/src/lib/stores/theme.ts          - Theme store & types
dashboard/src/lib/components/ThemeSelector.svelte - UI selector
dashboard/src/lib/components/layouts/
  ├── SidebarLayout.svelte    - Classic layout
  ├── TopNavLayout.svelte     - Top nav layout
  ├── CommandLayout.svelte    - Command center layout
  └── LayoutWrapper.svelte    - Layout switcher
```

---

## 🎯 Usage in Pages

**Wrap your page content with LayoutWrapper:**

```svelte
<script>
  import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
</script>

<LayoutWrapper>
  <!-- Your page content here -->
  <h1>Dashboard</h1>
  <div>Content...</div>
</LayoutWrapper>
```

**LayoutWrapper automatically:**
- Detects current theme from store
- Renders correct layout component
- Includes navigation
- Handles user info/logout
- Applies color theme

---

## 📱 Responsive Behavior

### Classic Sidebar
- Desktop: Full sidebar visible
- Tablet: Sidebar visible
- Mobile: Sidebar becomes dropdown menu

### Modern Top Nav
- Desktop: Horizontal menu center-aligned
- Tablet: Horizontal menu compact
- Mobile: Dropdown menu

### Command Center
- Desktop: 3-column layout
- Tablet: 2-column (nav + content, status hidden)
- Mobile: Single column, collapsible panels

---

## ⌨️ Keyboard Shortcuts (Command Center Only)

- `D` → Dashboard
- `S` → Scan
- `E` → Sessions
- `P` → Products
- `R` → Reports
- `T` → Settings

*(Shortcuts shown in left panel)*

---

## 🐛 Troubleshooting

### Theme not changing?
1. Hard refresh: `Ctrl + F5`
2. Clear localStorage: `F12` → Console → `localStorage.clear()`
3. Check browser console for errors

### Layout broken?
- Ensure LayoutWrapper is used in all pages
- Check browser compatibility (modern browsers only)
- Clear cache and reload

### Colors not applying?
- Check `tailwind.config.js` has all 3 themes defined
- Verify `data-theme` attribute on `<html>` tag
- Inspect element to see which theme is active

---

## 🎨 Customization

### Add New Color Theme

1. Edit `dashboard/tailwind.config.js`:
```javascript
daisyui: {
  themes: [
    {
      'your-theme-name': {
        primary: '#COLOR',
        secondary: '#COLOR',
        // ... other colors
      }
    }
  ]
}
```

2. Edit `dashboard/src/lib/stores/theme.ts`:
```typescript
export type ThemeColor = 'rack-room' | 'noc-console' | 'daylight-ops' | 'your-theme-name';

export const THEME_COLORS = [
  // ... existing themes
  {
    id: 'your-theme-name',
    name: 'Your Theme',
    description: 'Description',
    preview: { bg: '#...', primary: '#...', text: '#...' }
  }
];
```

3. Rebuild dashboard

### Add New Layout

1. Create `dashboard/src/lib/components/layouts/YourLayout.svelte`
2. Add to theme.ts types and THEME_LAYOUTS
3. Import in LayoutWrapper.svelte
4. Add conditional rendering
5. Rebuild dashboard

---

## 📊 Theme Usage Analytics

*(To be implemented)*
- Track which themes are most popular
- A/B test different layouts
- Gather user feedback

---

## Support

**Documentation:**
- Theme Guide: `docs/THEME_GUIDE.md` (this file)
- Scanner Guide: `docs/SCANNER_GUIDE.md`
- Troubleshooting: `docs/TROUBLESHOOTING.md`

**Access:**
- Settings: http://localhost/settings
- Dashboard: http://localhost/

**Default Theme:**
- Layout: Classic Sidebar
- Color: Rack Room
