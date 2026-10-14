# The README demo (docs/demo.gif), recorded with vhs (https://github.com/charmbracelet/vhs).
# Real output, no edits: a disposable Docker Engine 29.9 host (docker:29-dind,
# container dockdu-dind-29) after "LOG_MB=500 scripts/demo-up.sh", with the
# dockdu v0.3.0 linux/arm64 release binary (checksum verified) in its PATH.
# The hidden setup opens a shell inside that host, so every command shown
# runs there. Steps to reproduce: docs/demo/walkthrough.md.
Output docs/demo.gif
Set Shell bash
Set FontSize 13
Set Width 1540
Set Height 720
Set Padding 16
Set TypingSpeed 30ms
Set Theme {"name": "dockdu", "background": "#141E27", "foreground": "#EAF1F4", "selection": "#455C6B", "cursor": "#36B6C5", "black": "#202F3B", "red": "#EC8290", "green": "#8AC99D", "yellow": "#DFCA85", "blue": "#8BAED9", "magenta": "#BCA2DB", "cyan": "#36B6C5", "white": "#B4C4CE", "brightBlack": "#627A89", "brightRed": "#F4A2AC", "brightGreen": "#ACDDBA", "brightYellow": "#EBDCAB", "brightBlue": "#AEC9EA", "brightMagenta": "#D3BEEA", "brightCyan": "#81D8E1", "brightWhite": "#F3F7F9"}

Hide
Type "docker exec -it dockdu-dind-29 bash"
Enter
Sleep 1s
Type "export PS1='$ '; clear"
Enter
Sleep 500ms
Show

Type "docker system df"
Sleep 300ms
Enter
Sleep 3.5s

Hide
Type "clear"
Enter
Show
Type "dockdu --top 3"
Sleep 300ms
Enter
Sleep 13s

Hide
Type "clear"
Enter
Show
Type "du -ch /var/lib/docker/containers/*/*-json.log | tail -1"
Sleep 300ms
Enter
Sleep 4s
