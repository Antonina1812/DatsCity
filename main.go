package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Constants
const (
	baseURL         = "https://games-test.datsteam.dev/api"
	authTokenEnvVar = "002fdb0f-1651-4450-90aa-1d35ea0e0f8b" //Name of the env var for the auth token
)

// gamesdk package (simplified)
type PublicError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// model package
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
	RoundEndsAt string   `json:"roundEndsAt"` // Store as string, parse if needed
	ShuffleLeft int      `json:"shuffleLeft"`
	Turn        int      `json:"turn"`
	UsedIndexes []int    `json:"usedIndexes"`
	Words       []string `json:"words"` // Just the word, not the full object
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

// TowerWordRequest for incoming build requests
type TowerWordRequest struct {
	Dir int    `json:"dir"`
	ID  int    `json:"id"`
	Pos [3]int `json:"pos"`
}

// game package
type RoundListResponse struct {
	EventID string          `json:"eventId"`
	Now     string          `json:"now"` //Store as string and parse
	Rounds  []RoundResponse `json:"rounds"`
}

type RoundResponse struct {
	Duration int    `json:"duration"`
	EndAt    string `json:"endAt"` //Store as string and parse
	Name     string `json:"name"`
	Repeat   int    `json:"repeat"`
	StartAt  string `json:"startAt"` //Store as string and parse
	Status   string `json:"status"`
}

// Game-specific structs (for internal game logic)
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

// In-memory data (replace with a database in a real application)
var (
	completedTowers []Tower
	currentTower    Tower
	totalScore      float64 // Use float64
	gameState       GameState
	rounds          []Round
	availableWords  []Word // All possible words.
)

var wordList = []string{
	"извратитель",
	"феноменология",
	"необъективность",
	"предплужник",
	"человеконенавистник",
	"подвозчица",
	"насадка",
	"лопушник",
	"обводчик",
	"онанист",
	"копуляция",
	"сизаль",
	"куркума",
	"мостовина",
	"пальба",
	"ясачник",
	"подсек",
	"собранность",
	"иммиграция",
	"придача",
	"электричка",
	"ихтиоз",
	"пчеловод",
	"предвестие",
	"гудронатор",
	"швейцарец",
	"детская",
	"армянин",
	"циркон",
	"чревоугодник",
	"счёска",
	"кутание",
	"погромщик",
	"полеводство",
	"передумывание",
	"успокоение",
	"переборщица",
	"дягиль",
	"космополитка",
	"ныряло",
	"мануфактурсоветник",
	"сардоникс",
	"туземка",
	"некомпетентность",
	"палас",
	"сутяжник",
	"бытовизм",
	"пылкость",
	"многомужие",
	"валидол",
	"парадизка",
	"прозодежда",
	"нечаянность",
	"обдирание",
	"разъяснитель",
	"мерка",
	"защитник",
	"соковыжималка",
	"наказуемость",
	"фикция",
	"подстраивание",
	"остойчивость",
	"опрощенец",
	"подвиливание",
	"взаимопроникновение",
	"сложность",
	"ларчик",
	"сочевичник",
	"перестрагивание",
	"синхронизм",
	"диагностирование",
	"оживка",
	"заслушание",
	"индейководство",
	"дешифратор",
	"лахтак",
	"пруссак",
	"брас",
	"наёмничество",
	"копировщик",
	"отмывание",
	"культпоход",
	"предстоящее",
	"алеут",
	"лампион",
	"замуровывание",
	"гаммаустановка",
	"маслодел",
	"неграмотность",
	"неразличимость",
	"штундизм",
	"причащение",
	"полифония",
	"кувыркание",
	"кан",
	"недогляд",
	"магичность",
	"синеватость",
	"клавикорд",
	"коммивояжёрство",
	"куранта",
	"издольщик",
	"майордом",
	"европеизация",
	"хлёсткость",
	"выверение",
	"флотилия",
	"фитиль",
	"идолопоклонство",
	"возникновение",
	"пасьянс",
	"микрорайон",
	"скальд",
	"подвижник",
	"вендетта",
	"неразвитость",
	"келейник",
	"порез",
	"допашка",
	"зюйд",
	"кольт",
	"атака",
	"прополис",
	"злостность",
	"склёпывание",
	"корд",
	"характеристичность",
	"гнездование",
	"отбой",
	"экстернат",
	"сапонин",
	"биогеоценология",
	"модернистка",
	"федералист",
	"осведомительница",
	"изгнание",
	"продукт",
	"шквара",
	"миноискатель",
	"колодезь",
	"усекание",
	"разгадчица",
	"коринка",
	"подслащивание",
	"ихтиология",
	"рассылание",
	"лимфоцит",
	"торошение",
	"перезарядка",
	"неумолимость",
	"блондин",
	"верховенство",
	"переснащивание",
	"плис",
	"мужественность",
	"агитация",
	"взвывание",
	"закройная",
	"землепашец",
	"недокос",
	"кирзач",
	"трапезарь",
	"крем",
	"нейропатология",
	"невольница",
	"тыквина",
	"классика",
	"скип",
	"гагаузка",
	"налокотник",
	"крахмаление",
	"келейница",
	"доктринёрство",
	"прожировка",
	"югослав",
	"неточность",
	"недоброжелательность",
	"подтаптывание",
	"слушательница",
	"поярок",
	"умирание",
	"рассечение",
	"наличник",
	"хлорофилл",
	"соглядатайство",
	"шествие",
	"регистратура",
	"навозоразбрасыватель",
	"порция",
	"правдолюбие",
	"кабель",
	"примирённость",
	"мраморщик",
	"цежение",
	"моторизация",
	"штабквартира",
	"тигрёнок",
	"замерзание",
	"ворожей",
	"развальца",
	"испытание",
	"кровельщик",
	"невыгодность",
	"страусятина",
	"сберегание",
	"распря",
	"лёсс",
	"тактильность",
	"даурка",
	"рецидивистка",
	"душевность",
	"привой",
	"корнетист",
	"подвязка",
	"дарение",
	"комизм",
	"соединитель",
	"автотягач",
	"кокаин",
	"подсол",
	"синтепон",
	"монарх",
	"жеребьёвка",
	"артишок",
	"клятвопреступление",
	"пристрагивание",
	"тюркизм",
	"разбежка",
	"прогрессивность",
	"опутывание",
}

func init() {
	// Initialize game state and data here.  This is example data.  Replace with actual game logic.
	gameState = GameState{
		MapSize:     [3]int{30, 30, 100},
		NextTurnSec: 60,
		RoundEndsAt: time.Now().Add(time.Minute * 5),
		ShuffleLeft: 3,
		Turn:        1,
		UsedIndexes: []int{},
		Words:       []Word{}, //Initialize to empty, shuffle fills it
	}

	availableWords = make([]Word, len(wordList)) // Initialize availableWords with the word list
	for i, w := range wordList {
		availableWords[i] = Word{ID: i + 1, Word: w} // Assign unique IDs to words
	}

	rounds = []Round{
		{StartTime: time.Now(), EndTime: time.Now().Add(time.Minute * 10)},
		{StartTime: time.Now(), EndTime: time.Now().Add(time.Minute * 15)},
	}

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
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
		req.Header.Set("Content-Type", "application/json") //Set only when you have body
	}

	client := &http.Client{}
	return client.Do(req)
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlayerBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add new words to the current tower.
	for _, towerWordRequest := range req.Words {
		// Convert TowerWordRequest to Word
		word := Word{
			ID:  towerWordRequest.ID,
			Dir: towerWordRequest.Dir,
			Pos: towerWordRequest.Pos,

			// Find the word in availableWords by ID
			Word: func() string {
				for _, aw := range availableWords {
					if aw.ID == towerWordRequest.ID {
						return aw.Word
					}
				}
				return "" // Handle case where word ID isn't found
			}(),
		}

		// Check if word ID is already used
		alreadyUsed := false
		for _, usedID := range gameState.UsedIndexes {
			if word.ID == usedID {
				alreadyUsed = true
				break
			}
		}
		if alreadyUsed {
			//Construct PublicError response
			publicError := PublicError{}
			publicError.Code = 1001 // Example error code
			publicError.Message = fmt.Sprintf("Word ID %d already used", word.ID)

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(publicError) // Use json.NewEncoder
			return
		}

		if word.Word == "" {
			publicError := PublicError{}
			publicError.Code = 1002 //Example error code
			publicError.Message = fmt.Sprintf("Word with ID %d not found", word.ID)

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(publicError) // Use json.NewEncoder
			return
		}

		currentTower.Words = append(currentTower.Words, word)
		gameState.UsedIndexes = append(gameState.UsedIndexes, word.ID)
		// Calculate score based on the added words (replace with actual scoring logic).
		currentTower.Score += float64(len(word.Word)) // Convert to float64
	}

	// Check if the tower is completed.
	if req.Done {
		currentTower.Done = true
		completedTowers = append(completedTowers, currentTower)
		totalScore += currentTower.Score
		currentTower = Tower{} // Start a new tower

		//Reset used indexes for a new tower
		gameState.UsedIndexes = []int{}
	}

	//Return PlayerResponse
	playerResponse := PlayerResponse{}
	playerResponse.Score = totalScore

	//Add done towers to response
	for _, tower := range completedTowers {
		doneTower := DoneTowerResponse{}
		doneTower.ID = 1 // Placeholder. There is no ID field in your tower object

		doneTower.Score = tower.Score

		playerResponse.DoneTowers = append(playerResponse.DoneTowers, doneTower)
	}
	//Create player tower object
	playerTowerResponse := PlayerTowerResponse{}
	playerTowerResponse.Score = currentTower.Score

	for _, word := range currentTower.Words {
		playerWord := PlayerWord{}
		playerWord.Dir = word.Dir
		playerWord.Pos = word.Pos
		playerWord.Text = word.Word

		playerTowerResponse.Words = append(playerTowerResponse.Words, playerWord)
	}

	playerResponse.Tower = playerTowerResponse

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(playerResponse)
}

// handleShuffle handles the /api/shuffle endpoint.
func handleShuffle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if gameState.ShuffleLeft <= 0 {
		//Construct PublicError response
		publicError := PublicError{}
		publicError.Code = 2001
		publicError.Message = "No shuffles left"

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(publicError) // Use json.NewEncoder
		return
	}

	//Get shuffle words
	newWords, err := getShuffleWords()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameState.ShuffleLeft--

	//Construct PlayerWordsResponse
	playerWordsResponse := PlayerWordsResponse{}
	playerWordsResponse.ShuffleLeft = gameState.ShuffleLeft

	//Create the list of words for response, need just the word strings
	words := make([]string, 0)
	for _, word := range newWords {
		words = append(words, word.Word)
	}
	playerWordsResponse.Words = words

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(playerWordsResponse)
}

func getShuffleWords() ([]Word, error) {
	//Use external API to get shuffle words
	resp, err := makeRequest(http.MethodPost, "/shuffle", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var playerWordsResponse PlayerWordsResponse
	err = json.Unmarshal(body, &playerWordsResponse)

	if err != nil {
		return nil, err
	}

	//Get Word with ID from the words list
	words := make([]Word, 0)

	for _, wordString := range playerWordsResponse.Words {
		word, err := getWord(wordString)

		if err != nil {
			return nil, err
		}

		words = append(words, word)
	}

	return words, nil
}

func getWord(wordString string) (Word, error) {
	for _, word := range availableWords {
		if word.Word == wordString {
			return word, nil
		}
	}

	return Word{}, fmt.Errorf("word not found with string: %s", wordString)
}

// handleTowers handles the /api/towers endpoint.
func handleTowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Return PlayerResponse
	playerResponse := PlayerResponse{}
	playerResponse.Score = totalScore

	//Add done towers to response
	for _, tower := range completedTowers {
		doneTower := DoneTowerResponse{}
		doneTower.ID = 1 // Placeholder. There is no ID field in your tower object

		doneTower.Score = tower.Score

		playerResponse.DoneTowers = append(playerResponse.DoneTowers, doneTower)
	}
	//Create player tower object
	playerTowerResponse := PlayerTowerResponse{}
	playerTowerResponse.Score = currentTower.Score

	for _, word := range currentTower.Words {
		playerWord := PlayerWord{}
		playerWord.Dir = word.Dir
		playerWord.Pos = word.Pos
		playerWord.Text = word.Word

		playerTowerResponse.Words = append(playerTowerResponse.Words, playerWord)
	}

	playerResponse.Tower = playerTowerResponse

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playerResponse)
}

// handleWords handles the /api/words endpoint.
func handleWords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Perform request to external api
	resp, err := makeRequest(http.MethodGet, "/words", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var playerExtendedWordsResponse PlayerExtendedWordsResponse
	err = json.Unmarshal(body, &playerExtendedWordsResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playerExtendedWordsResponse)
}

// handleRounds handles the /api/rounds endpoint.
func handleRounds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Perform request to external api
	resp, err := makeRequest(http.MethodGet, "/rounds", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var roundListResponse RoundListResponse
	err = json.Unmarshal(body, &roundListResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roundListResponse)
}

func main() {
	//The server will now route requests to the test API endpoints
	http.HandleFunc("/api/build", handleBuild)
	http.HandleFunc("/api/shuffle", handleShuffle)
	http.HandleFunc("/api/towers", handleTowers)
	http.HandleFunc("/api/words", handleWords)
	http.HandleFunc("/api/rounds", handleRounds)

	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
