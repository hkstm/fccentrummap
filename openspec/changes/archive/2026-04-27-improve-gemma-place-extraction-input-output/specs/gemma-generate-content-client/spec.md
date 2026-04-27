## MODIFIED Requirements

### Requirement: Request payload carries Dutch prompt text as model input
The client SHALL send extraction prompt content in the request body using the API's content/parts text structure, including explicit cleaned-article and transcript sections.

#### Scenario: Prompt included in request
- **WHEN** the extraction command invokes the model client
- **THEN** the request body SHALL include the composed Dutch prompt as text content
- **AND** the prompt SHALL include both cleaned article text and transcript sentence units as separate labeled sections
- **AND** the prompt SHALL explicitly state that article text is primary for place identification and transcript units are required for mention timing/evidence
- **AND** the request SHALL target the configured Gemma model identifier
