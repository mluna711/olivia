package selector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

var BOP = fmt.Sprintf("http://%s:%s", os.Getenv("BOP_HOST"), os.Getenv("PORT"))

type currentQueueMsg struct {
	err         error
	queue       []Song
	mappedQueue map[string]Song
}

func (m model) getCurrentQueue() tea.Msg {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%s/queue", BOP))
	if err != nil {
		return currentQueueMsg{err: err}
	}

	var songs []Song
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&songs)
	if err != nil {
		return currentQueueMsg{err: err}
	}
	defer resp.Body.Close()

	mapped := map[string]Song{}
	for i, s := range songs {
		songs[i].Selected = true
		mapped[s.ID] = songs[i]
	}

	return currentQueueMsg{queue: songs, mappedQueue: mapped}
}

type addedToQueue struct {
	err   error
	empty bool
}

type AddToQueuePayload struct {
	IDS []string `json:"ids"`
}

func (m model) addToQueue() tea.Msg {
	if m.devMode {
		return addedToQueue{}
	}

	songs := []string{}
	for _, s := range m.queue.GetSongs() {
		songs = append(songs, s.ID)
	}

	if len(songs) == 0 && m.screenIndex == songsScreen {
		for _, s := range m.songs.GetSongs() {
			songs = append(songs, s.ID)
		}
	}

	if len(songs) == 0 {
		return addedToQueue{empty: true}
	}

	data := AddToQueuePayload{
		IDS: songs,
	}
	payload, err := json.Marshal(&data)
	if err != nil {
		return addedToQueue{err: err}
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/queue", BOP), bytes.NewBuffer(payload))
	if err != nil {
		return addedToQueue{err: err}
	}
	client := HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return addedToQueue{err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return addedToQueue{err: err}
		}
		return addedToQueue{err: errors.New(string(body))}
	}

	return addedToQueue{}
}

func (m model) addSelectedSongToQueue() tea.Msg {
	song := m.songs.GetSelected().ID
	data := AddToQueuePayload{
		IDS: []string{song},
	}
	payload, err := json.Marshal(&data)
	if err != nil {
		return addedToQueue{err: err}
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/queue", BOP), bytes.NewBuffer(payload))
	if err != nil {
		return addedToQueue{err: err}
	}
	client := HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return addedToQueue{err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return addedToQueue{err: err}
		}
		return addedToQueue{err: errors.New(string(body))}
	}

	return addedToQueue{}
}

type refetchedSongs struct {
	songs []Song
	err   error
}

func (m model) fetchSongs() tea.Msg {
	if m.input.Value() == "" {
		return refetchedSongs{
			songs: []Song{},
		}
	}

	q, err := parseQuery(m.input.Value())
	if err != nil {
		return refetchedSongs{err: err}
	}

	q.Page = m.currentPage + 1
	payload, err := json.Marshal(q)
	if err != nil {
		return refetchedSongs{err: err}
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/advancedsearch", BOP), bytes.NewBuffer(payload))
	if err != nil {
		return refetchedSongs{
			err: errors.New("Could not fetch data..."),
		}
	}

	client := HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return refetchedSongs{
			err: errors.New("Server error..."),
		}
	}
	defer resp.Body.Close()

	var data []Song
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&data)
	if err != nil {
		return refetchedSongs{
			err: errors.New("Could not parse response..."),
		}
	}

	// parsed := []Song{}
	// for i := 0; i < maxCount; i++ {
	// 	if i > len(data)-1 {
	// 		break
	// 	}
	//
	// 	parsed = append(parsed, data[i])
	// }

	return refetchedSongs{
		songs: data,
	}
}

type serverStatusMsg struct {
	err error
}

var serverDownErr = errors.New("server is down!")

func (m model) checkServerStatus() tea.Msg {
	r, err := http.DefaultClient.Get(fmt.Sprintf("%s/health", BOP))
	if err != nil {
		return serverStatusMsg{err: serverDownErr}
	}
	defer r.Body.Close()

	return serverStatusMsg{}
}

type BopQuery struct {
	From  string `json:"from"`
	By    string `json:"by"`
	Query string `json:"query"`
	Page  int    `json:"page"`
}

func parseQuery(query string) (BopQuery, error) {
	bq := BopQuery{}

	// from
	fromIndex := strings.Index(query, "from:")
	fromIndexEnd := -1
	if fromIndex != -1 {
		quotes := false
		q := ""
		index := -1
		for _, char := range query {
			index++
			// +5 'from:' len
			if index < fromIndex+5 {
				continue
			}
			if char == '"' && !quotes {
				quotes = true
				continue
			}

			if char == '"' {
				fromIndexEnd = index
				break
			}

			if char == ' ' && !quotes {
				fromIndexEnd = index
				break
			}

			if !quotes && index+1 == utf8.RuneCount([]byte(query)) {
				fromIndexEnd = index
			}

			q = fmt.Sprintf("%s%c", q, char)
		}
		bq.From = q
	}

	if fromIndexEnd == -1 && fromIndex != -1 {
		return BopQuery{}, errors.New("bad from clause")
	}

	// by
	if fromIndex != -1 && fromIndexEnd != -1 {
		query = string(append([]rune(query)[0:fromIndex], []rune(query)[fromIndexEnd+1:]...))
	}
	byIndex := strings.Index(query, "by:")
	byIndexEnd := -1
	if byIndex != -1 {
		quotes := false
		q := ""
		index := -1
		for _, char := range query {
			index++
			// +5 'by:' len
			if index < byIndex+3 {
				continue
			}
			if char == '"' && !quotes {
				quotes = true
				continue
			}

			if char == '"' {
				byIndexEnd = index
				break
			}

			if char == ' ' && !quotes {
				byIndexEnd = index
				break
			}

			if !quotes && index+1 == utf8.RuneCount([]byte(query)) {
				byIndexEnd = index
			}

			q = fmt.Sprintf("%s%c", q, char)
		}
		bq.By = q
	}

	if byIndexEnd == -1 && byIndex != -1 {
		return BopQuery{}, errors.New("bad by clause")
	}

	if byIndex != -1 && byIndexEnd != -1 {
		query = string(append([]rune(query)[0:byIndex], []rune(query)[byIndexEnd+1:]...))
	}

	bq.Query = query
	return bq, nil
}

func HTTPClient() http.Client {
	c := http.Client{}
	c.Timeout = time.Second * 3

	return c
}
