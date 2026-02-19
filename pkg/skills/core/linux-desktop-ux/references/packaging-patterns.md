# Linux Packaging Patterns

## PKGBUILD: Rust + GTK4 App (Complete Template)

```bash
# Maintainer: Name <email>
pkgname=myapp
pkgver=1.0.0
pkgrel=1
pkgdesc="Short description (max 80 chars, no period)"
arch=('x86_64' 'aarch64')
url="https://github.com/vendor/myapp"
license=('GPL-3.0-or-later')

# Runtime dependencies (GTK4 apps)
depends=(
    'gtk4'
    'libadwaita'
    'dbus'
    'hicolor-icon-theme'   # icon theme base
)

# Build-time only
makedepends=('rust' 'cargo')

# Optional backends
optdepends=(
    'v2ray: V2Ray proxy backend'
    'xray: Xray proxy backend'
)

source=("$pkgname-$pkgver.tar.gz::$url/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')  # Use actual hash in production

prepare() {
    cd "$pkgname-$pkgver"
    export RUSTUP_TOOLCHAIN=stable
    # Pre-fetch dependencies (offline build in subsequent steps)
    cargo fetch --locked --target "$(rustc -vV | sed -n 's/host: //p')"
}

build() {
    cd "$pkgname-$pkgver"
    export RUSTUP_TOOLCHAIN=stable
    export CARGO_TARGET_DIR=target
    # --frozen: use exact Cargo.lock (reproducible)
    cargo build --frozen --release
}

check() {
    cd "$pkgname-$pkgver"
    cargo test --frozen --release
}

package() {
    cd "$pkgname-$pkgver"

    # Binary
    install -Dm755 "target/release/$pkgname" \
        "$pkgdir/usr/bin/$pkgname"

    # Desktop file
    install -Dm644 "assets/$pkgname.desktop" \
        "$pkgdir/usr/share/applications/$pkgname.desktop"

    # Icon (scalable preferred)
    install -Dm644 "assets/$pkgname.svg" \
        "$pkgdir/usr/share/icons/hicolor/scalable/apps/$pkgname.svg"

    # Fallback PNG
    install -Dm644 "assets/$pkgname.png" \
        "$pkgdir/usr/share/icons/hicolor/256x256/apps/$pkgname.png"

    # AppStream metadata
    install -Dm644 "assets/$pkgname.metainfo.xml" \
        "$pkgdir/usr/share/metainfo/$pkgname.metainfo.xml"

    # License
    install -Dm644 LICENSE \
        "$pkgdir/usr/share/licenses/$pkgname/LICENSE"

    # Gettext translations (if any)
    for po_file in po/*.po; do
        locale=$(basename "$po_file" .po)
        msgfmt "$po_file" -o "locale/$locale.mo"
        install -Dm644 "locale/$locale.mo" \
            "$pkgdir/usr/share/locale/$locale/LC_MESSAGES/$pkgname.mo"
    done
}
```

---

## install -D Flag Explanation

| Flag | Meaning |
|---|---|
| `-D` | Create parent directories automatically |
| `-m755` | Permissions: rwxr-xr-x (executable) |
| `-m644` | Permissions: rw-r--r-- (regular file) |
| `-m600` | Permissions: rw------- (private config) |

Always use `-m644` for data files and `-m755` for executables. Never set executables as 644 (won't run) or data as 755 (security noise).

---

## Versioning PKGBUILD with pkgver()

For VCS packages (git), add a `pkgver()` function:

```bash
pkgname=myapp-git
pkgver() {
    cd "$pkgname"
    git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g'
}
```

---

## .SRCINFO Generation

After editing PKGBUILD, regenerate `.SRCINFO` for AUR:

```bash
makepkg --printsrcinfo > .SRCINFO
```

AUR requires `.SRCINFO` to be committed alongside `PKGBUILD`. Never edit `.SRCINFO` manually.

---

## AUR Submission Checklist

- [ ] `namcap PKGBUILD` — no warnings
- [ ] `namcap myapp-1.0.0-1-x86_64.pkg.tar.zst` — no missing deps
- [ ] `desktop-file-validate assets/myapp.desktop`
- [ ] `appstreamcli validate assets/myapp.metainfo.xml`
- [ ] `.SRCINFO` up to date (`makepkg --printsrcinfo > .SRCINFO`)
- [ ] `sha256sums` contain actual checksums (not SKIP) for release packages
- [ ] License file installed to `/usr/share/licenses/`
- [ ] `optdepends` describe what each optional dep enables

---

## Flatpak vs AUR: When to Use Each

| Criterion | AUR (PKGBUILD) | Flatpak |
|---|---|---|
| Target audience | Arch Linux users | All Linux distros |
| Isolation | None (system-integrated) | Sandboxed |
| Update mechanism | Pacman + AUR helper | Flatpak |
| GTK4 themes | Inherits system theme | Needs portal |
| System tray | Works natively | Requires SNI portal |
| Wayland | Native | Native |
| Best for | Power users, Arch | General distribution |

---

## Cargo.lock in Packages

**AUR packages**: commit `Cargo.lock` and use `--frozen --locked` in build to guarantee reproducible builds. Arch policy for security-sensitive packages.

**Flatpak**: use `flatpak-cargo-generator.py` to convert `Cargo.lock` into Flatpak module sources (offline build in Flatpak sandbox).

```bash
# Generate offline sources for Flatpak
python3 flatpak-cargo-generator.py Cargo.lock -o cargo-sources.json
```
