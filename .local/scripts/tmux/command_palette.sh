#! /usr/bin/env bash

RESULTS_FILE="$HOME/.cache/.shift_command_result"
CACHE_PATH="$HOME/.cache/.i_dont_know_how_to_program_and_my_code_should_be_illegal"

SESSIONS_PATH="$HOME/.cache/shift_sessions"
[ ! -d "$SESSIONS_PATH" ] && mkdir "$SESSIONS_PATH"

commands="$(cat - <<EOF
Notes
Clear panes
Terminate processes
Terminate processes and clear
Send command
Kill server
Close session
Open session
Detach client
EOF
)"

input() {
	local title
	local icon
	local input
	local mode

	title="$1"
	icon="$2"
	input="$3"
	mode="${4:-switch}"

	tmux display-popup -w 65 -h 11 -y 15 -E "[ -e \"$RESULTS_FILE\" ] && rm \"$RESULTS_FILE\" ; $HOME/.local/scripts/shift/shift -icon \"$icon\" -title \"$title\" -input \"$input\" -output \"$RESULTS_FILE\" -width 65 -height 9 -mode \"$mode\""
}

read_input() {
	tail -1 "$RESULTS_FILE"
}

input " Command Palette " " 󰘳 " "$commands"

case "$(read_input)" in
	"Open session")
		sessions=$(find "$SESSIONS_PATH" -type f -or -type l | sed "s|^$SESSIONS_PATH/||")
		input " Choose session " "  " "${sessions:-No sessions}"
		session_name=$(read_input)
		session_path="$SESSIONS_PATH/$session_name"

		[  "$session_name" = "No sessions" ] && exit
		[ ! -e "$RESULTS_FILE" ] && exit

		session_path=$(readlink "$session_path")
		session_name_without_extension=$(sed "s/.yml$//" <<< "$session_name")

		if tmux list-sessions | grep -i "$session_name_without_extension" &>/dev/null; then
			tmux switch-client -t "$session_name_without_extension"
		else
			tmuxp load -s "$session_name_without_extension" -d "$session_path" && \
				tmux switch-client -t "$session_name_without_extension"
		fi
		;;

	Notes)
		tmux display-popup -w "65" -h "11" -y 15 -E "$HOME/.local/scripts/notes/notes.sh"
		[ ! -e "$CACHE_PATH" ] && exit # no note selected
		tmux display-popup -b heavy -S fg=yellow -w "80%" -h "80%" -E "$HOME/.local/scripts/tmux/selectedcmd.sh"
		;;

	"Clear panes")
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} C-l
		;;

	"Terminate processes")
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} C-c
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} C-c
		;;

	"Send command")
		input " Command to send " " 󰘳 " " " "rename"
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} "$(read_input)" Enter
		;;

	"Terminate processes and clear")
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} C-c
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} C-c
		tmux list-panes -F "#{pane_index}" | xargs -I{} -n1 tmux send-keys -t {} C-l
		;;

	"Kill server")
		tmux kill-server
		;;

	"Close session")
		tmux kill-session
		;;

	"Detach client")
		tmux detach
		;;
esac
