package game

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	baseURL         = "https://games-test.datsteam.dev/api"
	authTokenEnvVar = "AUTH_TOKEN"
	ServerPort      = "8080"
)

type PublicError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DoneTowerResponse struct {
	ID    int     `json:"id"`
	Score float64 `json:"score"`
}

type PlayerBuildRequest struct {
	Done  bool               `json:"done"`
	Words []TowerWordRequest `json:"words"`
}

type PlayerExtendedWordsResponse struct {
	MapSize     [3]int   `json:"mapSize"`
	NextTurnSec int      `json:"nextTurnSec"`
	RoundEndsAt string   `json:"roundEndsAt"`
	ShuffleLeft int      `json:"shuffleLeft"`
	Turn        int      `json:"turn"`
	UsedIndexes []int    `json:"usedIndexes"`
	Words       []string `json:"words"`
}

type PlayerResponse struct {
	DoneTowers []DoneTowerResponse `json:"doneTowers"`
	Score      float64             `json:"score"`
	Tower      PlayerTowerResponse `json:"tower"`
}

type PlayerTowerResponse struct {
	Score float64      `json:"score"`
	Words []PlayerWord `json:"words"`
}

type PlayerWord struct {
	Dir  int    `json:"dir"`
	Pos  [3]int `json:"pos"`
	Text string `json:"text"`
}

type PlayerWordsResponse struct {
	ShuffleLeft int      `json:"shuffleLeft"`
	Words       []string `json:"words"`
}

type TowerWordRequest struct {
	Dir int    `json:"dir"`
	ID  int    `json:"id"`
	Pos [3]int `json:"pos"`
}

type RoundListResponse struct {
	EventID string          `json:"eventId"`
	Now     string          `json:"now"`
	Rounds  []RoundResponse `json:"rounds"`
}

type RoundResponse struct {
	Duration int    `json:"duration"`
	EndAt    string `json:"endAt"`
	Name     string `json:"name"`
	Repeat   int    `json:"repeat"`
	StartAt  string `json:"startAt"`
	Status   string `json:"status"`
}

type Word struct {
	ID   int    `json:"id"`
	Word string `json:"word"`
	Pos  [3]int `json:"pos"`
	Dir  int    `json:"dir"` // 1 = [0, 0, -1], 2 = [1, 0, 0], 3 = [0, 1, 0]
}

type Tower struct {
	Words []Word  `json:"words"`
	Score float64 `json:"score"` //Use float64 internally to match API
	Done  bool    `json:"done"`
}

type Round struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type GameState struct {
	MapSize     [3]int    `json:"mapSize"`
	NextTurnSec int       `json:"nextTurnSec"`
	RoundEndsAt time.Time `json:"roundEndsAt"`
	ShuffleLeft int       `json:"shuffleLeft"`
	Turn        int       `json:"turn"`
	UsedIndexes []int     `json:"usedIndexes"`
	Words       []Word    `json:"words"`
}

func getAuthToken() string {
	token := os.Getenv(authTokenEnvVar)
	if token == "" {
		log.Fatalf("Authentication token not found. Please set the %s environment variable.", authTokenEnvVar)
	}
	return token
}

func makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := baseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("X-Auth-Token", getAuthToken())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	return client.Do(req)
}

func HandleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := makeRequest(http.MethodPost, "/build", r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func HandleShuffle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := makeRequest(http.MethodPost, "/shuffle", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func HandleTowers(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := makeRequest(http.MethodGet, "/towers", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func HandleWords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := makeRequest(http.MethodGet, "/words", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func HandleRounds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := makeRequest(http.MethodGet, "/rounds", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}
