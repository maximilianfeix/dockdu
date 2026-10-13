# Recorded with vhs (https://github.com/charmbracelet/vhs): `vhs demo.tape`
# Runs the container one-liner so the log files inside the Docker Desktop VM are measured.
Output docs/demo.gif
Set Shell bash
Set FontSize 15
Set Width 1180
Set Height 760
Set Padding 18
Set TypingSpeed 35ms
Set Theme {"name": "dockdu", "background": "#141E27", "foreground": "#EAF1F4", "selection": "#455C6B", "cursor": "#36B6C5", "black": "#202F3B", "red": "#EC8290", "green": "#8AC99D", "yellow": "#DFCA85", "blue": "#8BAED9", "magenta": "#BCA2DB", "cyan": "#36B6C5", "white": "#B4C4CE", "brightBlack": "#627A89", "brightRed": "#F4A2AC", "brightGreen": "#ACDDBA", "brightYellow": "#EBDCAB", "brightBlue": "#AEC9EA", "brightMagenta": "#D3BEEA", "brightCyan": "#81D8E1", "brightWhite": "#F3F7F9"}

Hide
Type "clear"
Enter
Show
Type "docker system df"
Sleep 300ms
Enter
Sleep 2500ms
Type "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu --top 4"
Sleep 600ms
Enter
Sleep 9s
