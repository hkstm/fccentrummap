package extraction

import (
	"fmt"
	"strings"
)

type PromptInput struct {
	CleanedArticleText string
	Sentences          []SentenceUnit
}

func BuildDutchPrompt(input PromptInput) (string, error) {
	if strings.TrimSpace(input.CleanedArticleText) == "" {
		return "", fmt.Errorf("cannot build prompt without cleaned article text")
	}
	if len(input.Sentences) == 0 {
		return "", fmt.Errorf("cannot build prompt without sentence-level transcript units")
	}

	var b strings.Builder
	b.WriteString("Je bent een assistent die uit artikel + transcript de 'De spots van' locaties haalt.\n")
	b.WriteString("Doel: extraheer plekken die in Amsterdam of in de directe omgeving van Amsterdam liggen.\n")
	b.WriteString("Belangrijk: gebruik cleaned_article als primaire bron voor plaatsidentificatie.\n")
	b.WriteString("Belangrijk: accepteer alleen plekken met transcriptbewijs en een sentenceStartTimestamp uit transcript_sentences.\n")
	b.WriteString("Vraag: geef idealiter 2 tot 7 kandidaten, maar geef alleen plekken die daadwerkelijk in artikel + transcript worden ondersteund.\n")
	b.WriteString("\n")
	b.WriteString("Selectieregels:\n")
	b.WriteString("- Neem alleen 'echte spots' op die de spreker actief tipt, aanbeveelt of inhoudelijk beschrijft.\n")
	b.WriteString("- Neem géén plekken op die alleen als achtergrond, route of context worden genoemd.\n")
	b.WriteString("- Straat- en gebiedsnamen alleen opnemen als ze expliciet als spot worden gepresenteerd.\n")
	b.WriteString("- Als meerdere namen naar dezelfde spot verwijzen (bijv. Stopera en Nationaal Ballet), geef één canonieke spot terug.\n")
	b.WriteString("- Neem een spot alleen op als transcript_sentences expliciet bewijs levert en kies de best passende sentenceStartTimestamp.\n")
	b.WriteString("- Gebruik schone pleknaamwaarden zonder voorvoegsels zoals 'place:'.\n")
	b.WriteString("\n")
	b.WriteString("[cleaned_article]\n")
	b.WriteString(input.CleanedArticleText)
	b.WriteString("\n\n")
	b.WriteString("[transcript_sentences]\n")
	for i, s := range input.Sentences {
		b.WriteString(fmt.Sprintf("%d. [start=%.3f] %s\n", i+1, s.Start, s.Text))
	}
	b.WriteString("\n")
	b.WriteString("Gebruik de functie-aanroep '")
	b.WriteString(SubmitSpotsFunctionName)
	b.WriteString("' voor je antwoord.\n")
	b.WriteString("Gebruik exact dit arguments-formaat:\n")
	b.WriteString("{\n")
	b.WriteString("  \"presenter_name\": \"<naam van primaire presentator>\",\n")
	b.WriteString("  \"spots\": [\n")
	b.WriteString("    {\"place\": \"<naam van plek>\", \"sentenceStartTimestamp\": <numerieke starttijd>}\n")
	b.WriteString("  ]\n")
	b.WriteString("}\n")
	b.WriteString("Gebruik exact de velden 'presenter_name', 'place' en 'sentenceStartTimestamp'.\n")
	b.WriteString("Als de primaire presentator onbekend is: laat 'presenter_name' weg of gebruik een lege string.\n")

	return b.String(), nil
}
