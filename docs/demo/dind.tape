# Records docs/demo.gif against a disposable Docker Engine 29 daemon
# (docker:29-dind, container dockdu-dind-29) after scripts/demo-up.sh ran
# inside it with LOG_MB=100. The hidden setup routes "docker" and "dockdu"
# into that container; the output is the real output of that daemon.
#   vhs docs/demo/dind.tape
Output docs/demo.gif
Set Shell bash
Set FontSize 14
Set Width 1400
Set Height 840
Set Padding 18
Set TypingSpeed 35ms
Set Theme {"name": "dockdu", "background": "#141E27", "foreground": "#EAF1F4", "selection": "#455C6B", "cursor": "#36B6C5", "black": "#202F3B", "red": "#EC8290", "green": "#8AC99D", "yellow": "#DFCA85", "blue": "#8BAED9", "magenta": "#BCA2DB", "cyan": "#36B6C5", "white": "#B4C4CE", "brightBlack": "#627A89", "brightRed": "#F4A2AC", "brightGreen": "#ACDDBA", "brightYellow": "#EBDCAB", "brightBlue": "#AEC9EA", "brightMagenta": "#D3BEEA", "brightCyan": "#81D8E1", "brightWhite": "#F3F7F9"}

Hide
Type `docker() { command docker exec dockdu-dind-29 docker "$@"; }; dockdu() { command docker exec dockdu-dind-29 dockdu "$@"; }; export PS1="$ "; clear`
Enter
Show
Type "docker system df"
Sleep 300ms
Enter
Sleep 2500ms
Type "dockdu --top 3"
Sleep 600ms
Enter
Sleep 9s
