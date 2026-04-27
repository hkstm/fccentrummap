package extraction

import (
	"fmt"
	"strings"
)

func BuildDutchPrompt(sentences []SentenceUnit) (string, error) {
	if len(sentences) == 0 {
		return "", fmt.Errorf("cannot build prompt without sentence-level transcript units")
	}

	var b strings.Builder
	b.WriteString("Je bent een assistent die uit een transcript de 'De spots van' locaties haalt.\n")
	b.WriteString("Doel: extraheer plekken die in Amsterdam of in de directe omgeving van Amsterdam liggen.\n")
	b.WriteString("Vraag: geef idealiter 2 tot 7 kandidaten, maar geef alleen plekken die daadwerkelijk in het transcript worden genoemd.\n")
	b.WriteString("Belangrijk: 2-7 is een richtlijn; als er minder of meer zijn, geef de correcte lijst.\n")
	b.WriteString("\n")
	b.WriteString("Selectieregels:\n")
	b.WriteString("- Neem alleen 'echte spots' op die de spreker actief tipt, aanbeveelt of inhoudelijk beschrijft.\n")
	b.WriteString("- Neem géén plekken op die alleen als achtergrond, route of context worden genoemd.\n")
	b.WriteString("- Straat- en gebiedsnamen alleen opnemen als ze expliciet als spot worden gepresenteerd.\n")
	b.WriteString("- Als meerdere namen naar dezelfde spot verwijzen (bijv. Stopera en Nationaal Ballet), geef één canonieke spot terug.\n")
	b.WriteString("- Neem een spot alleen op als die door minimaal 2 verschillende zinnen in het transcript wordt ondersteund.\n")
	b.WriteString("- Gebruik schone pleknaamwaarden zonder voorvoegsels zoals 'place:'.\n")
	b.WriteString("\n")
	b.WriteString("Gebruik uitsluitend de onderstaande zin-eenheden met starttijd als bewijs:\n")
	for i, s := range sentences {
		b.WriteString(fmt.Sprintf("%d. [start=%.3f] %s\n", i+1, s.Start, s.Text))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Gebruik de functie-aanroep '%s' voor je antwoord.\n", SubmitSpotsFunctionName))
	b.WriteString("Gebruik exact dit arguments-formaat:\n")
	b.WriteString("{\n")
	b.WriteString("  \"spots\": [\n")
	b.WriteString("    {\"place\": \"<naam van plek>\", \"sentenceStartTimestamp\": <numerieke starttijd>}\n")
	b.WriteString("  ]\n")
	b.WriteString("}\n")
	b.WriteString("Gebruik exact de velden 'place' en 'sentenceStartTimestamp'.\n")

	return b.String(), nil
}
