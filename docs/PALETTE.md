# dockdu palette

| Role | Hex |
| --- | --- |
| Accent · exposed space | `#36B6C5` |
| Dark background / badge label | `#141E27` |
| Light background | `#F3F7F9` |
| Primary text on dark | `#EAF1F4` |
| Secondary text on dark | `#B4C4CE` |
| Primary text on light | `#141E27` |
| Secondary text on light / storage fill | `#455C6B` |
| Storage track | `#202F3B` |
| Secondary storage fill | `#627A89` |
| Exposed segment background | `#193D48` |

Cyan evokes the cool precision of storage instrumentation and makes previously unreported space visible. It gives dockdu a distinct identity beside the author's coral and amber projects while keeping the same restrained terminal aesthetic.

Use cyan text on dark surfaces; the light banner preserves a dark storage inset for readable accent text. The storage proportions are illustrative, not measured data. The mark uses the same 256-unit geometry in every asset, scaled uniformly.

## Shields.io

```text
style=flat-square&color=36B6C5&labelColor=141E27
```

## VHS demo theme

Paste this single-line custom theme into the demo `.tape` file. Syntax follows the [VHS custom theme documentation](https://github.com/charmbracelet/vhs#theme).

```tape
Set Theme {"name": "dockdu", "background": "#141E27", "foreground": "#EAF1F4", "selection": "#455C6B", "cursor": "#36B6C5", "black": "#202F3B", "red": "#EC8290", "green": "#8AC99D", "yellow": "#DFCA85", "blue": "#8BAED9", "magenta": "#BCA2DB", "cyan": "#36B6C5", "white": "#B4C4CE", "brightBlack": "#627A89", "brightRed": "#F4A2AC", "brightGreen": "#ACDDBA", "brightYellow": "#EBDCAB", "brightBlue": "#AEC9EA", "brightMagenta": "#D3BEEA", "brightCyan": "#81D8E1", "brightWhite": "#F3F7F9"}
```

Use cyan for highlighted disk usage and the same monospace stack as the artwork where available: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace`. The wordmark and tagline use `ui-sans-serif, system-ui, -apple-system, "Segoe UI", Inter, sans-serif`; no font files are required.
