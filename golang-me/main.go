package main

import (
	"fmt"
	"math/rand"
	"time"
	"os"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

type Difficulty struct {
	Name        string
	MaxNumber   int
	MaxAttempts int
}

type GameResult struct {
	Date       string
	Difficulty string
	Outcome    string
	Attempts   int
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var results []GameResult

	for {
		diff := selectDifficulty()
		res := playGame(rng, diff)
		results = append(results, res)

		printResultsJSON(results)
		saveResult(res)

		if !askPlayAgain() {
			fmt.Println("Спасибо за игру! До встречи!")
			break
		}
	}
}

func selectDifficulty() Difficulty {
	fmt.Println("\n=== ВЫБОР СЛОЖНОСТИ ===")
	fmt.Println("1. Easy   (1–50, 15 попыток)")
	fmt.Println("2. Medium (1–100, 10 попыток)")
	fmt.Println("3. Hard   (1–200, 5 попыток)")

	for {
		fmt.Print("Выберите уровень (1-3): ")
		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			clearInputBuffer(err)
			fmt.Println(colorRed + "Ошибка: введите число 1, 2 или 3." + colorReset)
			continue
		}

		switch choice {
		case 1:
			return Difficulty{Name: "Easy", MaxNumber: 50, MaxAttempts: 15}
		case 2:
			return Difficulty{Name: "Medium", MaxNumber: 100, MaxAttempts: 10}
		case 3:
			return Difficulty{Name: "Hard", MaxNumber: 200, MaxAttempts: 5}
		default:
			fmt.Println(colorRed + "Ошибка: выберите вариант от 1 до 3." + colorReset)
		}
	}
}

func generateNumber(rng *rand.Rand, max int) int {
	return rng.Intn(max) + 1
}

func readGuess(max int) int {
	for {
		fmt.Printf("Введите число от 1 до %d: ", max)
		var guess int
		_, err := fmt.Scanln(&guess)
		if err != nil {
			clearInputBuffer(err)
			fmt.Println(colorRed + "Ошибка: введите корректное целое число (попытка не сгорела)." + colorReset)
			continue
		}

		if guess < 1 || guess > max {
			fmt.Printf("%sОшибка: число должно быть в диапазоне от 1 до %d (попытка не сгорела).%s\n", colorRed, max, colorReset)
			continue
		}

		return guess
	}
}

func getHints(guess, target int) (string, string) {
	var direction string
	if guess < target {
		direction = "Секретное число больше 👆"
	} else {
		direction = "Секретное число меньше 👇"
	}

	diff := target - guess
	if diff < 0 {
		diff = -diff
	}

	var distance string
	switch {
	case diff <= 5:
		distance = "🔥 Горячо"
	case diff <= 15:
		distance = "🙂 Тепло"
	default:
		distance = "❄️ Холодно"
	}

	return direction, distance
}

func playGame(rng *rand.Rand, diff Difficulty) GameResult {
	target := generateNumber(rng, diff.MaxNumber)
	var history []int

	fmt.Printf("\nИгра %s'Угадай число'%s -   от 1 до %d! \nУгадайте число за %d попыток!\n",
		colorGreen, colorReset, diff.MaxNumber, diff.MaxAttempts)

	for attempt := 1; attempt <= diff.MaxAttempts; attempt++ {
		fmt.Printf("\n%sПопытка %d из %d%s\n", colorYellow, attempt, diff.MaxAttempts, colorReset)
		if len(history) > 0 {
			fmt.Printf("Ваши прошлые числа: %v\n", history)
		}

		guess := readGuess(diff.MaxNumber)
		history = append(history, guess)

		if guess == target {
			fmt.Printf("%sВы угадали!🙌%s Игра закончена!!", colorGreen, colorReset)
			return GameResult{
				Date:       time.Now().Format("2006-01-02 15:04:05"),
				Difficulty: diff.Name,
				Outcome:    "Победа",
				Attempts:   attempt,
			}
		}

		direction, distance := getHints(guess, target)
		fmt.Printf("%s | %s\n", direction, distance)
	}

	fmt.Printf("%sВы проиграли!😢%s Секретное число было: %d", colorRed, colorReset, target)

	return GameResult{
		Date:       time.Now().Format("2006-01-02 15:04:05"),
		Difficulty: diff.Name,
		Outcome:    "Поражение",
		Attempts:   diff.MaxAttempts,
	}
}

func printResultsJSON(results []GameResult) {
	fmt.Println("\n--- ИСТОРИЯ РЕЗУЛЬТАТОВ (results.json) ---")
	fmt.Println("[")
	for i, r := range results {
		comma := ","
		if i == len(results)-1 {
			comma = ""
		}
		fmt.Printf("  {\"date\": %q, \"difficulty\": %q, \"outcome\": %q, \"attempts\": %d}%s\n",
			r.Date, r.Difficulty, r.Outcome, r.Attempts, comma)
	}
	fmt.Println("]")
}
func saveResult(r GameResult) {
	const filename = "results.json"

	newEntry := fmt.Sprintf("  {\n    \"date\": %q,\n    \"difficulty\": %q,\n    \"outcome\": %q,\n    \"attempts\": %d\n  }",
		r.Date, r.Difficulty, r.Outcome, r.Attempts)

	data, err := os.ReadFile(filename)
	var content string

	if err != nil || len(data) == 0 {
		content = fmt.Sprintf("[\n%s\n]\n", newEntry)
	} else {
		lastBrace := -1
		lastBracket := -1
		for i := len(data) - 1; i >= 0; i-- {
			if lastBracket == -1 && data[i] == ']' {
				lastBracket = i
			}
			if lastBrace == -1 && data[i] == '}' {
				lastBrace = i
				break
			}
		}

		if lastBrace != -1 && lastBracket != -1 {
			content = string(data[:lastBrace+1]) + ",\n" + newEntry + "\n]\n"
		} else {
			content = fmt.Sprintf("[\n%s\n]\n", newEntry)
		}
	}

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Println(colorRed + "Ошибка записи в файл results.json." + colorReset)
		return
	}

	fmt.Println("Результат сохранён в results.json.")
}
func askPlayAgain() bool {
	for {
		fmt.Print("\nСыграть ещё раз? (y/n): ")
		var answer string
		_, err := fmt.Scanln(&answer)
		if err != nil {
			clearInputBuffer(err)
			continue
		}

		switch answer {
		case "y", "Y", "yes", "YES", "д", "Д", "да", "ДА":
			return true
		case "n", "N", "no", "NO", "н", "Н", "нет", "НЕТ":
			return false
		default:
			fmt.Println("Пожалуйста, введите 'y' (да) или 'n' (нет).")
		}
	}
}

func clearInputBuffer(err error) {
	if err != nil && err.Error() != "unexpected newline" {
		var discard string
		fmt.Scanln(&discard)
	}
}