package usecases

import (
	"strings"
	"text_analyzer_service/internal/domain/models"
	"unicode"
)

type BasicTextAnalyzer struct{}

func NewBasicTextAnalyzer() *BasicTextAnalyzer {
	return &BasicTextAnalyzer{}
}

func (a *BasicTextAnalyzer) Analyze(text *models.TextContent) (*models.AnalysisResult, error) {
	normalizedText := normalizeText(text.Data)

	result := &models.AnalysisResult{
		IsApproved:   true,
		IsSpam:       isSpam(normalizedText),
		HasSensitive: containsSensitive(text.Data),
		Sentiment:    calculateSentiment(normalizedText),
		Language:     detectLanguage(normalizedText),
	}

	if uppercaseRatio(text.Data) > 0.5 {
		result.IsSpam = true
	}

	if result.IsSpam || result.HasSensitive {
		result.IsApproved = false
	}

	return result, nil
}

func normalizeText(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

func containsSensitive(text string) bool {
	sensitiveWords := []string{
		"bad", "sensitive", "dangerous", "hate", "violence", "attack", "kill",
		"bomb", "terror", "drugs", "weapon", "racist", "nazi", "hitler",
		"suicide", "murder", "rape", "pedo", "scam", "fraud",

		"опасный", "терроризм", "наркотики", "оружие", "насилие", "ненависть",
		"атака", "убийство", "мошенничество", "скам", "расизм", "нацизм",
		"изнасилование", "педофил", "суицид",

		"идиот", "дурак", "кретин", "retard", "fuck", "shit", "asshole",
		"мудак", "сволочь", "ублюдок",
	}

	for _, word := range sensitiveWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func isSpam(text string) bool {
	spamIndicators := []string{
		"click here", "make money", "earn cash", "work from home",
		"limited offer", "special promotion", "buy now", "discount",
		"free gift", "win prize", "congratulations you won",
		"urgent", "only today", "click this link", "unsubscribe",
		"100% free", "risk free", "no cost", "no fees",
		"money back", "guarantee", "increase sales", "double your",
		"extra income", "home based", "be your own boss",

		"бесплатно", "быстрые деньги", "легкий заработок", "гарантированный доход",
		"криптовалюта бесплатно", "выиграй миллион", "деньги сразу", "быстрый кредит",
		"акция только сегодня", "успей до конца дня", "скидка 50%", "распродажа",
		"купи сейчас", "уникальный товар", "ваш аккаунт заблокирован", "срочно обновите данные",
		"нажмите чтобы получить", "кликните по ссылке", "срочное уведомление", "важная информация",
		"добавь в друзья", "подпишись и получи", "репостни чтобы выиграть", "знакомства без регистрации",
	}

	if strings.Contains(text, "http://") || strings.Contains(text, "https://") ||
		strings.Contains(text, ".com") || strings.Contains(text, ".ru") {
		return true
	}

	for _, phrase := range spamIndicators {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	words := strings.Fields(text)
	wordCount := make(map[string]int)
	for _, word := range words {
		if len(word) > 3 {
			wordCount[word]++
			if wordCount[word] > 3 {
				return true
			}
		}
	}

	return false
}

func detectLanguage(text string) string {
	for _, r := range text {
		if unicode.Is(unicode.Cyrillic, r) {
			return "ru"
		}
	}
	return "en"
}

func calculateSentiment(text string) float64 {
	positiveWords := []string{"good", "great", "awesome", "happy", "love", "лучший", "отлично", "прекрасно"}
	negativeWords := []string{"bad", "terrible", "hate", "awful", "angry", "плохо", "ужасно", "ненависть"}

	score := 0.1

	words := strings.Fields(text)
	if len(words) == 0 {
		return score
	}

	positiveCount := 0
	negativeCount := 0

	for _, word := range words {
		for _, pw := range positiveWords {
			if strings.Contains(word, pw) {
				positiveCount++
			}
		}
		for _, nw := range negativeWords {
			if strings.Contains(word, nw) {
				negativeCount++
			}
		}
	}

	total := positiveCount + negativeCount
	if total > 0 {
		score = float64(positiveCount-negativeCount)/float64(total) + 0.1
		if score > 1.0 {
			score = 1.0
		} else if score < -1.0 {
			score = -1.0
		}
	}

	return score
}

func uppercaseRatio(text string) float64 {
	totalLetters := 0
	uppercaseLetters := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalLetters++
			if unicode.IsUpper(r) {
				uppercaseLetters++
			}
		}
	}

	if totalLetters == 0 {
		return 0
	}

	return float64(uppercaseLetters) / float64(totalLetters)
}
