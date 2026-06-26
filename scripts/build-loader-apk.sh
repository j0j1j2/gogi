#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SDK="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
if [[ -z "$SDK" ]]; then
  echo "ANDROID_HOME or ANDROID_SDK_ROOT is required" >&2
  exit 1
fi

NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}"
if [[ -z "$NDK" ]]; then
  NDK="$(find "$SDK/ndk" -mindepth 1 -maxdepth 1 -type d | sort -V | tail -1)"
fi

BUILD_TOOLS="$(find "$SDK/build-tools" -mindepth 1 -maxdepth 1 -type d | sort -V | tail -1)"
ANDROID_JAR="$(find "$SDK/platforms" -mindepth 1 -maxdepth 1 -type d | sort -V | tail -1)/android.jar"
CLANG="$(find "$NDK/toolchains/llvm/prebuilt" -type f -path '*/bin/aarch64-linux-android24-clang' | head -1)"
LLVM_NM="$(dirname "$CLANG")/llvm-nm"

if [[ -z "$CLANG" || ! -x "$CLANG" ]]; then
  echo "aarch64-linux-android24-clang not found under $NDK" >&2
  exit 1
fi

OUT="$ROOT/tmp/gogi-loader"
rm -rf "$OUT"
mkdir -p "$OUT/classes" "$OUT/dex" "$OUT/apk/lib/arm64-v8a"

TARGET_SO="$OUT/apk/lib/arm64-v8a/libtarget.so"
"$CLANG" -shared -fPIC \
  "$ROOT/testapps/gogi-loader/native/target.c" \
  -o "$TARGET_SO"

TARGET_RVA="$("$LLVM_NM" -D --defined-only "$TARGET_SO" | awk '$3 == "gogi_target_value" { print "0x"$1; exit }')"
if [[ -z "$TARGET_RVA" ]]; then
  echo "failed to resolve gogi_target_value RVA" >&2
  exit 1
fi

GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC="$CLANG" \
  go build \
  -ldflags "-X github.com/j0j1j2/gogi/payload/runtime.demoTargetValueRVAHex=$TARGET_RVA" \
  -buildmode=c-shared \
  -o "$OUT/apk/lib/arm64-v8a/libgogi.so" "$ROOT/payload"

mapfile -t JAVA_FILES < <(find "$ROOT/testapps/gogi-loader/src" -type f -name '*.java' | sort)
javac -classpath "$ANDROID_JAR" \
  -d "$OUT/classes" \
  "${JAVA_FILES[@]}"

mapfile -t CLASS_FILES < <(find "$OUT/classes" -type f -name '*.class' | sort)
"$BUILD_TOOLS/d8" --min-api 24 --output "$OUT/dex" "${CLASS_FILES[@]}"

"$BUILD_TOOLS/aapt2" link \
  -o "$OUT/base.apk" \
  --manifest "$ROOT/testapps/gogi-loader/AndroidManifest.xml" \
  -I "$ANDROID_JAR" \
  --min-sdk-version 24 \
  --target-sdk-version 35

cp "$OUT/dex/classes.dex" "$OUT/apk/classes.dex"
(
  cd "$OUT/apk"
  zip -qr "$OUT/base.apk" classes.dex lib
)

"$BUILD_TOOLS/zipalign" -f 4 "$OUT/base.apk" "$OUT/aligned.apk"
"$BUILD_TOOLS/apksigner" sign \
  --ks "$HOME/.android/debug.keystore" \
  --ks-pass pass:android \
  --key-pass pass:android \
  --out "$OUT/gogi-loader.apk" \
  "$OUT/aligned.apk"

echo "$OUT/gogi-loader.apk"
