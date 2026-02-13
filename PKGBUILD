# Maintainer: Gabriel Ong <gongahkia@users.noreply.github.com>
pkgname=salja
pkgver=0.1.0
pkgrel=1
pkgdesc="Universal calendar and task converter"
arch=('x86_64' 'aarch64')
url="https://github.com/gongahkia/salja"
license=('MIT')
makedepends=('go')
source=("$pkgname-$pkgver.tar.gz::https://github.com/gongahkia/salja/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')

build() {
  cd "$pkgname-$pkgver"
  export CGO_ENABLED=0
  go build -o salja ./cmd/salja
}

package() {
  cd "$pkgname-$pkgver"
  install -Dm755 salja "$pkgdir/usr/bin/salja"
  install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
}
