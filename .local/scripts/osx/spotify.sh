#! /bin/sh

pgrep pgrep Spotify &>/dev/null || exit

icon=""

get_song_name() {
	spotify status track
}
song_name=$(get_song_name)

get_artist() {
	spotify status artist
}
artist=$(get_artist)

title="$song_name - $artist"
short_title=$(echo $title | cut -c -30)

should_truncate=$([ ${#title} -gt 30 ] && echo yes || echo no)

[ "$should_truncate" == yes ] && echo "#[bg=red,fg=black] $icon $short_title...#[fg=default]" || echo "#[bg=red,fg=black] $icon $title#[fg=default]"
