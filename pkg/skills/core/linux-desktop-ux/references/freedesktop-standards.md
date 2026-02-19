# Freedesktop.org Standards Reference

## XDG Base Directory Variables

| Variable | Default | Purpose |
|---|---|---|
| `XDG_CONFIG_HOME` | `~/.config` | User config files |
| `XDG_DATA_HOME` | `~/.local/share` | User data (persistent) |
| `XDG_CACHE_HOME` | `~/.cache` | Cached data (can be cleared) |
| `XDG_STATE_HOME` | `~/.local/state` | State, history, logs |
| `XDG_RUNTIME_DIR` | `/run/user/<UID>` | Sockets, pipes (ephemeral) |
| `XDG_DATA_DIRS` | `/usr/local/share:/usr/share` | System data search path |
| `XDG_CONFIG_DIRS` | `/etc/xdg` | System config search path |

Always read the variable first, fall back to the default only if unset:

```rust
fn config_home() -> PathBuf {
    std::env::var("XDG_CONFIG_HOME")
        .ok()
        .map(PathBuf::from)
        .unwrap_or_else(|| {
            let home = std::env::var("HOME").unwrap_or_default();
            PathBuf::from(home).join(".config")
        })
}
```

---

## .desktop File Full Specification

```ini
[Desktop Entry]

# Required
Type=Application           # Application | Link | Directory
Name=Display Name          # Shown in launcher
Exec=/usr/bin/app %F       # Command; %F=files, %U=URLs, %u=single URL, %i=icon

# Strongly recommended
Comment=Short description  # Shown as tooltip
Icon=app-icon-name         # Name in icon theme (no path, no extension)
Categories=Network;        # Semicolon-separated, must end with semicolon
StartupNotify=true         # Show "launching..." spinner
StartupWMClass=appname     # Must match WM_CLASS (xprop on running window)
Terminal=false             # Whether to run in terminal

# Optional
Version=1.0
GenericName=VPN Client     # Generic category label
Keywords=proxy;vpn;        # Search keywords (semicolons, lowercase)
MimeType=text/plain;       # File types the app handles
TryExec=/usr/bin/app       # Don't show if binary not found
NoDisplay=false            # Set true to hide from launcher (still usable by other .desktop files)
Hidden=false               # Deprecated — use NoDisplay
Path=/working/dir          # Working directory for Exec

# Desktop Actions (jumplist items)
Actions=NewWindow;

[Desktop Action NewWindow]
Name=Open New Window
Exec=/usr/bin/app --new-window
```

**Validate before packaging:**
```bash
desktop-file-validate myapp.desktop
```

---

## Icon Naming Convention

System icon names (no extension, no path) — look these up in the icon theme:

```
# Actions
list-add-symbolic        # + add button
list-remove-symbolic     # - remove button
edit-delete-symbolic     # delete/trash
document-save-symbolic   # save
go-previous-symbolic     # back arrow
go-next-symbolic         # forward arrow
view-refresh-symbolic    # refresh/reload
preferences-system-symbolic  # settings gear

# Status
network-transmit-receive-symbolic
process-working-symbolic  # loading
dialog-information-symbolic
dialog-warning-symbolic
dialog-error-symbolic

# Application categories
utilities-system-monitor-symbolic
network-workgroup-symbolic
```

For a custom app icon, use the app ID: `com.vendor.myapp` (matches the AppStream ID).

---

## Icon Size Guidelines

| Size | Use case |
|---|---|
| 16×16 | File manager thumbnails (small) |
| 32×32 | Taskbar, file manager (medium) |
| 48×48 | Application grid, minimum required |
| 64×64 | Desktop icons |
| 128×128 | High-DPI displays |
| 256×256 | App store thumbnails |
| scalable (SVG) | Preferred — used at any size |
| symbolic (SVG) | Panel/tray — single color, 16px optical size |

---

## AppStream Field Reference

Required fields for Flathub submission:

| Field | Required | Notes |
|---|---|---|
| `<id>` | ✅ | Must match `.desktop` ID (reverse-DNS) |
| `<name>` | ✅ | App display name |
| `<summary>` | ✅ | Max 60 chars, no period at end |
| `<metadata_license>` | ✅ | License for the metainfo file itself (CC0-1.0) |
| `<project_license>` | ✅ | SPDX expression (GPL-3.0-or-later) |
| `<description>` | ✅ | `<p>` tags, no HTML |
| `<url type="homepage">` | ✅ | Project homepage |
| `<releases>` | ✅ | At least one release with version + date |
| `<content_rating>` | ✅ | OARS rating (use `oars-1.1`) |
| `<screenshots>` | recommended | At least one screenshot for app stores |

Validate:
```bash
appstreamcli validate myapp.metainfo.xml
```

---

## gettext Workflow

### 1. Extract strings

```bash
xgettext --language=Rust --keyword=gettext --keyword=ngettext:1,2 \
    -o po/myapp.pot src/**/*.rs
```

### 2. Create translation

```bash
msginit --input=po/myapp.pot --locale=de_DE --output=po/de.po
# Edit po/de.po in Poedit or similar
```

### 3. Compile .po → .mo

```bash
msgfmt po/de.po -o locale/de/LC_MESSAGES/myapp.mo
```

### 4. Install .mo files

```bash
install -Dm644 locale/de/LC_MESSAGES/myapp.mo \
    "$pkgdir/usr/share/locale/de/LC_MESSAGES/myapp.mo"
```

### 5. Initialize at runtime

```rust
use gettextrs::TextDomain;

TextDomain::new("myapp")
    .locale_dir("/usr/share/locale")
    .init()
    .expect("failed to initialize i18n");
```

---

## Flatpak Manifest (basic)

```yaml
app-id: com.vendor.myapp
runtime: org.gnome.Platform
runtime-version: "47"
sdk: org.gnome.Sdk
command: myapp

finish-args:
  - --share=network
  - --socket=wayland
  - --socket=fallback-x11
  - --share=ipc
  - --device=dri

modules:
  - name: myapp
    buildsystem: simple
    build-commands:
      - cargo build --release
      - install -Dm755 target/release/myapp /app/bin/myapp
    sources:
      - type: git
        url: https://github.com/vendor/myapp.git
        tag: v1.0.0
```
