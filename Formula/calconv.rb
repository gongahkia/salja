class Calconv < Formula
  desc "Universal calendar and task converter CLI"
  homepage "https://github.com/gongahkia/calendar-converter"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/gongahkia/calendar-converter/releases/download/v#{version}/calconv_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    else
      url "https://github.com/gongahkia/calendar-converter/releases/download/v#{version}/calconv_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/gongahkia/calendar-converter/releases/download/v#{version}/calconv_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/gongahkia/calendar-converter/releases/download/v#{version}/calconv_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "calconv"
  end

  test do
    assert_match "calconv version", shell_output("#{bin}/calconv --version")
    assert_match "ics", shell_output("#{bin}/calconv list-formats")
  end
end
