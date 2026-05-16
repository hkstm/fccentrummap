package extraction

import (
	"fmt"
	"strings"
)

type PromptInput struct {
	CleanedArticleText string
	Sentences          []SentenceUnit
}

type RefinementPromptInput struct {
	Sentences   []SentenceUnit
	Pass1Spots  []Candidate
	AudioStarts float64
}

func BuildDutchPrompt(input PromptInput) (string, error) {
	return BuildDutchPass1Prompt(input)
}

func BuildDutchPass1Prompt(input PromptInput) (string, error) {
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
	b.WriteString("- Als artikel en transcript verschillend spellen, volg de spelling uit cleaned_article als canoniek.\n")
	b.WriteString("- Corrigeer transcript-typo's alleen naar een vorm die letterlijk in cleaned_article staat; verzin geen nieuwe varianten.\n")
	b.WriteString("- Neem een spot alleen op als transcript_sentences expliciet bewijs levert en kies de best passende sentenceStartTimestamp.\n")
	b.WriteString("- Gebruik schone pleknaamvalues zonder voorvoegsels zoals 'place:'.\n")
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

func BuildDutchPass2RefinementPrompt(input RefinementPromptInput) (string, error) {
	if len(input.Sentences) == 0 {
		return "", fmt.Errorf("cannot build refinement prompt without sentence-level transcript units")
	}
	if len(input.Pass1Spots) == 0 {
		return "", fmt.Errorf("cannot build refinement prompt without pass-1 spots")
	}

	var b strings.Builder
	b.WriteString("Je bent een assistent die timestamps van bestaande spots verfijnt op basis van transcriptbewijs.\n")
	b.WriteString("Doel: geef voor elke spot de vroegst-logische start van hetzelfde gespreksonderwerp in het transcript.\n")
	b.WriteString("HEEL BELANGRIJK: de output place-values moeten EXACT overeenkomen. Als in pass 1 spots place=Pension Homeland dan in output moeten we ook EXACT 'Pension Homeland' hebben niet bijvoorbeeld 'Pension Homeland Homeland'.\n")
	b.WriteString("Belangrijk: werk in batch voor ALLE spots uit pass 1 in één functie-aanroep.\n")
	b.WriteString("Belangrijk: refinedSentenceStartTimestamp moet <= originalSentenceStartTimestamp zijn.\n")
	b.WriteString("Belangrijk: refinement voor een place moet STRIKT groter zijn dan de original (niet-refined) pass 1 timestamp van de 'vorige' place. Dus: previous original place timestamp < current place refined timestamp <= current place original timestamp.\n")
	b.WriteString("Belangrijk: als geen betere eerdere anchor bestaat, gebruik exact de originele timestamp (no-op).\n")
	b.WriteString("Belangrijk: gebruik uitsluitend transcript_sentences als bewijs; geen article-context nodig in deze pass.\n")
	b.WriteString("Toegestaan: je output mag een subset van pass1_spots bevatten (je hoeft niet elke spot op te nemen).\n")
	b.WriteString("Niet toegestaan: dubbele place-values in spots (elke place maximaal één keer).\n")
	b.WriteString("Niet toegestaan: nieuwe of onbekende place-values die niet in pass1_spots staan.\n")
	b.WriteString("\n")
	b.WriteString("[pass1_spots]\n")
	for i, s := range input.Pass1Spots {
		if s.OriginalSentenceStartTimestamp == nil {
			continue
		}
		b.WriteString(fmt.Sprintf("%d. place=%s originalSentenceStartTimestamp=%.3f\n", i+1, s.Place, *s.OriginalSentenceStartTimestamp))
	}
	b.WriteString("\n")
	b.WriteString("[transcript_sentences]\n")
	for i, s := range input.Sentences {
		b.WriteString(fmt.Sprintf("%d. [start=%.3f] %s\n", i+1, s.Start, s.Text))
	}
	b.WriteString("\n")
	b.WriteString("Gebruik de functie-aanroep '")
	b.WriteString(SubmitRefinedSpotsFunctionName)
	b.WriteString("' voor je antwoord.\n")
	b.WriteString("Gebruik exact dit arguments-formaat:\n")
	b.WriteString("{\n")
	b.WriteString("  \"spots\": [\n")
	b.WriteString("    {\"place\": \"<naam van plek uit pass1>\", \"refinedSentenceStartTimestamp\": <numerieke starttijd>}\n")
	b.WriteString("  ]\n")
	b.WriteString("}\n")
	b.WriteString("Gebruik exact de velden 'place' en 'refinedSentenceStartTimestamp'.\n")

	return b.String(), nil
}
