# Fonts

This directory contains fonts used by the PentAGI frontend application, including web UI and PDF report generation.

## Inter

- **Files**: `inter-*.woff2` (8 files)
- **License**: SIL Open Font License 1.1
- **License File**: [OFL.txt](./OFL.txt)
- **Source**: https://github.com/rsms/inter
- **Copyright**: Copyright (c) 2016 The Inter Project Authors
- **Usage**: Main UI font for the web application.

## Roboto Mono

- **Files**: `roboto-mono-*.woff2` (8 files)
- **License**: Apache License 2.0
- **License File**: [LICENSE-Apache-2.0.txt](./LICENSE-Apache-2.0.txt)
- **Source**: https://github.com/google/roboto
- **Copyright**: Google Inc.
- **Usage**: Monospace font for code display in the web UI.

## Noto Sans

- **Files**: `NotoSans-Regular.ttf`, `NotoSans-Bold.ttf`, `NotoSans-Italic.ttf`, `NotoSans-BoldItalic.ttf`
- **License**: SIL Open Font License 1.1
- **License File**: [OFL.txt](./OFL.txt)
- **Source**: https://github.com/googlefonts/noto-fonts
- **Copyright**: Copyright 2022 The Noto Project Authors
- **Usage**: Main body font in PDF reports. Covers Latin, Cyrillic, Greek, Devanagari, Vietnamese, and many other scripts.

## Noto Sans Mono

- **Files**: `NotoSansMono-Regular.ttf`, `NotoSansMono-Bold.ttf`
- **License**: SIL Open Font License 1.1
- **License File**: [OFL.txt](./OFL.txt)
- **Source**: https://github.com/googlefonts/noto-fonts
- **Copyright**: Copyright 2022 The Noto Project Authors
- **Usage**: Monospace font for code blocks in PDF reports. Covers Latin and Cyrillic.

## Noto Sans SC (Simplified Chinese)

- **Files**: `NotoSansSC-Regular.otf`, `NotoSansSC-Bold.otf`
- **License**: SIL Open Font License 1.1
- **License File**: [OFL.txt](./OFL.txt)
- **Source**: https://github.com/notofonts/noto-cjk
- **Copyright**: Copyright 2014-2021 Adobe Systems Incorporated
- **Usage**: CJK font for Chinese, Japanese, and Korean characters in PDF reports. Loaded on demand when CJK text is detected.

## Notes

All fonts are open source and freely redistributable under their respective licenses (SIL OFL 1.1 and
Apache 2.0), both compatible with the MIT license of the PentAGI product. Fonts are served from the
same origin as the application and never loaded from external servers.
